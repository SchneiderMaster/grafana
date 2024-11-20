package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/net/context"
	"k8s.io/client-go/rest"

	"github.com/grafana/authlib/claims"
	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/services/apiserver"
	"github.com/grafana/grafana/pkg/services/authn"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/serviceaccounts"
	"github.com/grafana/grafana/pkg/services/user"
)

type BackgroundIdentityService interface {
	WorkerIdentity(ctx context.Context, namespace string) (identity.Requester, error)

	// This will return a rest.Config with a worker identity attached in the context
	RestConfigForNamespace(ctx context.Context, namespace string) (*rest.Config, error)
}

func ProvideProvisioningIdentityService(
	serviceAccounts serviceaccounts.Service,
	authn authn.Service,
	clientConfigProvider apiserver.DirectRestConfigProvider,
	// HACK, for now we will use a global grafana admin user
	// service accounts need permissions granted (as far as i can tell)
	users user.Service,
	orgs org.Service,
) BackgroundIdentityService {
	prefix := "provisioning-background-worker"

	return &backgroundIdentities{
		serviceAccountNamePrefix: prefix,
		role:                     org.RoleAdmin,
		clientConfigProvider:     clientConfigProvider,

		log:             slog.Default().With("logger", "background-identities", "prefix", prefix),
		accounts:        make(map[int64]string),
		serviceAccounts: serviceAccounts,
		authn:           authn,
		users:           users,
		orgs:            orgs,
	}
}

type backgroundIdentities struct {
	log *slog.Logger

	serviceAccountNamePrefix string
	role                     org.RoleType
	clientConfigProvider     apiserver.DirectRestConfigProvider

	// typed ids
	accounts map[int64]string
	mutex    sync.Mutex

	users           user.Service
	serviceAccounts serviceaccounts.Service
	authn           authn.Service
	orgs            org.Service
}

func (o *backgroundIdentities) WorkerIdentity(ctx context.Context, namespace string) (identity.Requester, error) {
	info, err := claims.ParseNamespace(namespace)
	if err != nil {
		return nil, err
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	id, ok := o.accounts[info.OrgID]
	if !ok {
		if false { // <<<<< HACK
			id, err = o.makeAdminUser(ctx, info.OrgID)
		} else {
			id, err = o.verifyServiceAccount(ctx, info.OrgID)
		}

		if err != nil {
			return nil, err
		}

		o.accounts[info.OrgID] = id
	}

	return o.authn.ResolveIdentity(ctx, info.OrgID, id)
}

func (o *backgroundIdentities) makeAdminUser(ctx context.Context, orgId int64) (string, error) {
	found, err := o.users.GetByLogin(ctx, &user.GetUserByLoginQuery{
		LoginOrEmail: o.serviceAccountNamePrefix,
	})
	if found == nil {
		// Same user across all orgs??
		found, err = o.users.Create(ctx, &user.CreateUserCommand{
			Login:          o.serviceAccountNamePrefix,
			Name:           o.serviceAccountNamePrefix,
			Email:          o.serviceAccountNamePrefix,
			OrgID:          orgId,
			IsAdmin:        true,
			DefaultOrgRole: string(org.RoleAdmin),
		})
	}
	if err != nil {
		return "", err
	}

	// Make sure the user exists in the org
	err = o.orgs.UpdateOrgUser(ctx, &org.UpdateOrgUserCommand{
		Role:   identity.RoleAdmin,
		OrgID:  orgId,
		UserID: found.ID,
	})

	return claims.NewTypeID(claims.TypeUser, found.UID), err
}

func (o *backgroundIdentities) verifyServiceAccount(ctx context.Context, orgId int64) (string, error) {
	serviceAccountName := fmt.Sprintf("%s-org-%d", o.serviceAccountNamePrefix, orgId)
	saForm := serviceaccounts.CreateServiceAccountForm{
		Name: serviceAccountName,
		Role: &o.role,
	}

	serviceAccount, err := o.serviceAccounts.CreateServiceAccount(ctx, orgId, &saForm)
	if serviceAccount == nil {
		accountAlreadyExists := errors.Is(err, serviceaccounts.ErrServiceAccountAlreadyExists)
		if accountAlreadyExists {
			accountId, err := o.serviceAccounts.RetrieveServiceAccountIdByName(ctx, orgId, serviceAccountName)
			if err != nil {
				o.log.Error("Failed to retrieve service account", "err", err, "accountName", serviceAccountName)
				return "", err
			}
			// update org_role to make sure everything works properly if someone has changed the role since SA's original creation
			dto, err := o.serviceAccounts.UpdateServiceAccount(ctx, orgId, accountId, &serviceaccounts.UpdateServiceAccountForm{
				Name: &serviceAccountName,
				Role: &o.role,
			})
			if err != nil {
				o.log.Error("Failed to update service account", "err", err, "accountName", serviceAccountName)
				return "", err
			}

			serviceAccount = &serviceaccounts.ServiceAccountDTO{
				Id:  dto.Id,
				UID: dto.UID,
			}
		}
	}
	if serviceAccount == nil {
		o.log.Error("Failed to retrieve service account", "err", err, "accountName", serviceAccountName)
		return "", err
	}

	return claims.NewTypeID(claims.TypeServiceAccount, fmt.Sprintf("%d", serviceAccount.Id)), nil
}

func (o *backgroundIdentities) RestConfigForNamespace(ctx context.Context, namespace string) (*rest.Config, error) {
	id, err := o.WorkerIdentity(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return o.clientConfigProvider.GetRestConfigForBackgroundWorker(id), nil
}
