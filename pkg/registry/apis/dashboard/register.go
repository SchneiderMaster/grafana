package dashboard

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"

	claims "github.com/grafana/authlib/types"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	dashboardinternal "github.com/grafana/grafana/pkg/apis/dashboard"
	dashboardv0alpha1 "github.com/grafana/grafana/pkg/apis/dashboard/v0alpha1"
	dashboardv1alpha1 "github.com/grafana/grafana/pkg/apis/dashboard/v1alpha1"
	dashboardv2alpha1 "github.com/grafana/grafana/pkg/apis/dashboard/v2alpha1"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/registry/apis/dashboard/legacy"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/apiserver/builder"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/provisioning"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/storage/legacysql"
	"github.com/grafana/grafana/pkg/storage/unified/apistore"
	"github.com/grafana/grafana/pkg/storage/unified/resource"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ builder.APIGroupBuilder      = (*DashboardsAPIBuilder)(nil)
	_ builder.OpenAPIPostProcessor = (*DashboardsAPIBuilder)(nil)
)

// This is used just so wire has something unique to return
type DashboardsAPIBuilder struct {
	dashboardService dashboards.DashboardService
	features         featuremgmt.FeatureToggles

	accessControl                accesscontrol.AccessControl
	legacy                       *DashboardStorage
	unified                      resource.ResourceClient
	dashboardProvisioningService dashboards.DashboardProvisioningService
	scheme                       *runtime.Scheme

	log log.Logger
	reg prometheus.Registerer
}

func RegisterAPIService(cfg *setting.Cfg, features featuremgmt.FeatureToggles,
	apiregistration builder.APIRegistrar,
	dashboardService dashboards.DashboardService,
	provisioningDashboardService dashboards.DashboardProvisioningService,
	accessControl accesscontrol.AccessControl,
	provisioning provisioning.ProvisioningService,
	dashStore dashboards.Store,
	reg prometheus.Registerer,
	sql db.DB,
	tracing *tracing.TracingService,
	unified resource.ResourceClient,
) *DashboardsAPIBuilder {
	softDelete := features.IsEnabledGlobally(featuremgmt.FlagDashboardRestore)
	dbp := legacysql.NewDatabaseProvider(sql)
	namespacer := request.GetNamespaceMapper(cfg)
	builder := &DashboardsAPIBuilder{
		log: log.New("grafana-apiserver.dashboards"),

		dashboardService:             dashboardService,
		features:                     features,
		accessControl:                accessControl,
		unified:                      unified,
		dashboardProvisioningService: provisioningDashboardService,

		legacy: &DashboardStorage{
			Resource:       dashboardv0alpha1.DashboardResourceInfo,
			Access:         legacy.NewDashboardAccess(dbp, namespacer, dashStore, provisioning, softDelete),
			TableConverter: dashboardv0alpha1.DashboardResourceInfo.TableConverter(),
			Features:       features,
		},
		reg: reg,
	}
	apiregistration.RegisterAPI(builder)
	return builder
}

func (b *DashboardsAPIBuilder) GetGroupVersions() []schema.GroupVersion {
	return []schema.GroupVersion{
		dashboardv0alpha1.DashboardResourceInfo.GroupVersion(),
		dashboardv1alpha1.DashboardResourceInfo.GroupVersion(),
		dashboardv2alpha1.DashboardResourceInfo.GroupVersion(),
	}
}

func (b *DashboardsAPIBuilder) InstallSchema(scheme *runtime.Scheme) error {
	b.scheme = scheme
	if err := dashboardinternal.AddToScheme(scheme); err != nil {
		return err
	}
	if err := dashboardv0alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := dashboardv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := dashboardv2alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return scheme.SetVersionPriority(b.GetGroupVersions()...)
}

// Validate will prevent deletion of provisioned dashboards, unless the grace period is set to 0, indicating a force deletion
func (b *DashboardsAPIBuilder) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	op := a.GetOperation()
	if op == admission.Delete {
		obj := a.GetOperationOptions()
		deleteOptions, ok := obj.(*metav1.DeleteOptions)
		if !ok {
			return fmt.Errorf("expected v1.DeleteOptions")
		}

		if deleteOptions.GracePeriodSeconds == nil || *deleteOptions.GracePeriodSeconds != 0 {
			nsInfo, err := claims.ParseNamespace(a.GetNamespace())
			if err != nil {
				return fmt.Errorf("%v: %w", "failed to parse namespace", err)
			}

			provisioningData, err := b.dashboardProvisioningService.GetProvisionedDashboardDataByDashboardUID(ctx, nsInfo.OrgID, a.GetName())
			if err != nil {
				if errors.Is(err, dashboards.ErrProvisionedDashboardNotFound) {
					return nil
				}

				return fmt.Errorf("%v: %w", "failed to check if dashboard is provisioned", err)
			}

			if provisioningData != nil {
				return dashboards.ErrDashboardCannotDeleteProvisionedDashboard
			}
		}
	}

	return nil
}

// Mutate removes any internal ID set in the spec & adds it as a label
func (b *DashboardsAPIBuilder) Mutate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	op := a.GetOperation()
	if op != admission.Create && op != admission.Update {
		return nil
	}
	obj := a.GetObject()
	dash, ok := obj.(*dashboardinternal.Dashboard)
	if !ok {
		return fmt.Errorf("mutation error: expected *dashboardinternal.Dashboard, got %T", obj)
	}

	if id, ok := dash.Spec.Object["id"].(float64); ok {
		delete(dash.Spec.Object, "id")
		if id != 0 {
			meta, err := utils.MetaAccessor(obj)
			if err != nil {
				return err
			}
			meta.SetDeprecatedInternalID(int64(id)) // nolint:staticcheck
		}
	}

	return nil
}

func (b *DashboardsAPIBuilder) UpdateAPIGroupInfo(apiGroupInfo *genericapiserver.APIGroupInfo, opts builder.APIGroupOptions) error {
	internalDashResourceInfo := dashboardinternal.DashboardResourceInfo

	// wrap the optsGetter to add encode versioner
	opts.OptsGetter = &dashboardOptsGetter{
		optsGetter: opts.OptsGetter,
	}

	legacyStore, err := b.legacy.NewStore(opts.Scheme, opts.OptsGetter, b.reg)
	if err != nil {
		return err
	}

	storageOpts := apistore.StorageOptions{
		RequireDeprecatedInternalID: true,
	}

	// Split dashboards when they are large
	var largeObjects apistore.LargeObjectSupport
	if b.legacy.Features.IsEnabledGlobally(featuremgmt.FlagUnifiedStorageBigObjectsSupport) {
		largeObjects = NewDashboardLargeObjectSupport(opts.Scheme)
		storageOpts.LargeObjectSupport = largeObjects
	}
	opts.StorageOptions(internalDashResourceInfo.GroupResource(), storageOpts)

	// v0alpha1
	storage, err := b.storageForVersion(opts, legacyStore, largeObjects, func() runtime.Object {
		return &dashboardv0alpha1.DashboardWithAccessInfo{}
	})
	if err != nil {
		return err
	}
	apiGroupInfo.VersionedResourcesStorageMap[dashboardv0alpha1.VERSION] = storage

	// v1alpha1
	storage, err = b.storageForVersion(opts, legacyStore, largeObjects, func() runtime.Object {
		return &dashboardv1alpha1.DashboardWithAccessInfo{}
	})
	if err != nil {
		return err
	}
	apiGroupInfo.VersionedResourcesStorageMap[dashboardv1alpha1.VERSION] = storage

	// v2alpha1
	storage, err = b.storageForVersion(opts, legacyStore, largeObjects, func() runtime.Object {
		return &dashboardv2alpha1.DashboardWithAccessInfo{}
	})
	if err != nil {
		return err
	}
	apiGroupInfo.VersionedResourcesStorageMap[dashboardv2alpha1.VERSION] = storage

	return nil
}

func (b *DashboardsAPIBuilder) storageForVersion(
	opts builder.APIGroupOptions,
	legacyStore grafanarest.LegacyStorage,
	largeObjects apistore.LargeObjectSupport,
	newDTOFunc func() runtime.Object,
) (map[string]rest.Storage, error) {
	var (
		err                      error
		scheme                   = opts.Scheme
		dualWriteBuilder         = opts.DualWriteBuilder
		internalDashResourceInfo = dashboardinternal.DashboardResourceInfo
		libraryPanelResourceInfo = dashboardinternal.LibraryPanelResourceInfo
	)

	storage := map[string]rest.Storage{}
	storage[internalDashResourceInfo.StoragePath()] = legacyStore

	// Dual writes if a RESTOptionsGetter is provided
	if dualWriteBuilder != nil {
		store, err := grafanaregistry.NewRegistryStore(scheme, internalDashResourceInfo, opts.OptsGetter)
		if err != nil {
			return nil, err
		}
		storage[internalDashResourceInfo.StoragePath()], err = dualWriteBuilder(internalDashResourceInfo.GroupResource(), legacyStore, store)
		if err != nil {
			return nil, err
		}
	}

	if b.features.IsEnabledGlobally(featuremgmt.FlagKubernetesRestore) {
		storage[internalDashResourceInfo.StoragePath("restore")] = NewRestoreConnector(
			b.unified,
			internalDashResourceInfo.GroupResource(),
		)

		storage[internalDashResourceInfo.StoragePath("latest")] = NewLatestConnector(
			b.unified,
			internalDashResourceInfo.GroupResource(),
		)
	}

	// Register the DTO endpoint that will consolidate all dashboard bits
	storage[internalDashResourceInfo.StoragePath("dto")], err = NewDTOConnector(
		storage[internalDashResourceInfo.StoragePath()],
		largeObjects,
		b.legacy.Access,
		b.unified,
		b.accessControl,
		scheme,
		newDTOFunc,
	)
	if err != nil {
		return nil, err
	}

	// Expose read only library panels
	storage[libraryPanelResourceInfo.StoragePath()] = &LibraryPanelStore{
		Access:       b.legacy.Access,
		ResourceInfo: dashboardinternal.LibraryPanelResourceInfo,
	}

	return storage, nil
}

func (b *DashboardsAPIBuilder) GetOpenAPIDefinitions() common.GetOpenAPIDefinitions {
	return func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
		defs := dashboardv0alpha1.GetOpenAPIDefinitions(ref)
		maps.Copy(defs, dashboardv1alpha1.GetOpenAPIDefinitions(ref))
		maps.Copy(defs, dashboardv2alpha1.GetOpenAPIDefinitions(ref))
		return defs
	}
}

func (b *DashboardsAPIBuilder) PostProcessOpenAPI(oas *spec3.OpenAPI) (*spec3.OpenAPI, error) {
	// The plugin description
	oas.Info.Description = "Grafana dashboards as resources"

	// Hide cluster-scoped resources
	for i := 0; i < 3; i++ {
		version := fmt.Sprintf("v%dalpha1", i)
		p := path.Join("/apis/", dashboardinternal.GROUP, version, "dashboards")
		delete(oas.Paths.Paths, p)
	}

	return oas, nil
}

var _ genericregistry.RESTOptionsGetter = &dashboardOptsGetter{}

type dashboardOptsGetter struct {
	optsGetter genericregistry.RESTOptionsGetter
}

func (d *dashboardOptsGetter) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (genericregistry.RESTOptions, error) {
	opts, err := d.optsGetter.GetRESTOptions(resource, example)
	if err != nil {
		return genericregistry.RESTOptions{}, err
	}
	// this ensures that the dashboard resource is encoded at v0alpha1
	opts.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(
		dashboardv0alpha1.DashboardResourceInfo.GroupVersion(),
		dashboardv0alpha1.DashboardResourceInfo.GroupVersionKind().GroupKind(),
		dashboardv0alpha1.LibraryPanelResourceInfo.GroupVersionKind().GroupKind(),
	)
	return opts, nil
}
