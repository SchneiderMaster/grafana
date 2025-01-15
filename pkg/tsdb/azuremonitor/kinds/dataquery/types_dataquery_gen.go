// Code generated - EDITING IS FUTILE. DO NOT EDIT.
//
// Generated by:
//     public/app/plugins/gen.go
// Using jennies:
//     PluginGoTypesJenny
//
// Run 'make gen-cue' from repository root to regenerate.

package dataquery

// Defines values for AppInsightsGroupByQueryKind.
const (
	AppInsightsGroupByQueryKindAppInsightsGroupByQuery AppInsightsGroupByQueryKind = "AppInsightsGroupByQuery"
)

// Defines values for AppInsightsMetricNameQueryKind.
const (
	AppInsightsMetricNameQueryKindAppInsightsMetricNameQuery AppInsightsMetricNameQueryKind = "AppInsightsMetricNameQuery"
)

// Defines values for AzureQueryType.
const (
	AzureQueryTypeAzureLogAnalytics               AzureQueryType = "Azure Log Analytics"
	AzureQueryTypeAzureMetricNames                AzureQueryType = "Azure Metric Names"
	AzureQueryTypeAzureMonitor                    AzureQueryType = "Azure Monitor"
	AzureQueryTypeAzureNamespaces                 AzureQueryType = "Azure Namespaces"
	AzureQueryTypeAzureRegions                    AzureQueryType = "Azure Regions"
	AzureQueryTypeAzureResourceGraph              AzureQueryType = "Azure Resource Graph"
	AzureQueryTypeAzureResourceGroups             AzureQueryType = "Azure Resource Groups"
	AzureQueryTypeAzureResourceNames              AzureQueryType = "Azure Resource Names"
	AzureQueryTypeAzureSubscriptions              AzureQueryType = "Azure Subscriptions"
	AzureQueryTypeAzureTraces                     AzureQueryType = "Azure Traces"
	AzureQueryTypeAzureWorkspaces                 AzureQueryType = "Azure Workspaces"
	AzureQueryTypeGrafanaTemplateVariableFunction AzureQueryType = "Grafana Template Variable Function"
	AzureQueryTypeTraceql                         AzureQueryType = "traceql"
)

// Defines values for GrafanaTemplateVariableQueryType.
const (
	GrafanaTemplateVariableQueryTypeAppInsightsGroupByQuery    GrafanaTemplateVariableQueryType = "AppInsightsGroupByQuery"
	GrafanaTemplateVariableQueryTypeAppInsightsMetricNameQuery GrafanaTemplateVariableQueryType = "AppInsightsMetricNameQuery"
	GrafanaTemplateVariableQueryTypeMetricNamesQuery           GrafanaTemplateVariableQueryType = "MetricNamesQuery"
	GrafanaTemplateVariableQueryTypeMetricNamespaceQuery       GrafanaTemplateVariableQueryType = "MetricNamespaceQuery"
	GrafanaTemplateVariableQueryTypeResourceGroupsQuery        GrafanaTemplateVariableQueryType = "ResourceGroupsQuery"
	GrafanaTemplateVariableQueryTypeResourceNamesQuery         GrafanaTemplateVariableQueryType = "ResourceNamesQuery"
	GrafanaTemplateVariableQueryTypeSubscriptionsQuery         GrafanaTemplateVariableQueryType = "SubscriptionsQuery"
	GrafanaTemplateVariableQueryTypeUnknownQuery               GrafanaTemplateVariableQueryType = "UnknownQuery"
	GrafanaTemplateVariableQueryTypeWorkspacesQuery            GrafanaTemplateVariableQueryType = "WorkspacesQuery"
)

// Defines values for MetricDefinitionsQueryKind.
const (
	MetricDefinitionsQueryKindMetricDefinitionsQuery MetricDefinitionsQueryKind = "MetricDefinitionsQuery"
)

// Defines values for MetricNamesQueryKind.
const (
	MetricNamesQueryKindMetricNamesQuery MetricNamesQueryKind = "MetricNamesQuery"
)

// Defines values for MetricNamespaceQueryKind.
const (
	MetricNamespaceQueryKindMetricNamespaceQuery MetricNamespaceQueryKind = "MetricNamespaceQuery"
)

// Defines values for ResourceGroupsQueryKind.
const (
	ResourceGroupsQueryKindResourceGroupsQuery ResourceGroupsQueryKind = "ResourceGroupsQuery"
)

// Defines values for ResourceNamesQueryKind.
const (
	ResourceNamesQueryKindResourceNamesQuery ResourceNamesQueryKind = "ResourceNamesQuery"
)

// Defines values for ResultFormat.
const (
	ResultFormatLogs       ResultFormat = "logs"
	ResultFormatTable      ResultFormat = "table"
	ResultFormatTimeSeries ResultFormat = "time_series"
	ResultFormatTrace      ResultFormat = "trace"
)

// Defines values for SubscriptionsQueryKind.
const (
	SubscriptionsQueryKindSubscriptionsQuery SubscriptionsQueryKind = "SubscriptionsQuery"
)

// Defines values for UnknownQueryKind.
const (
	UnknownQueryKindUnknownQuery UnknownQueryKind = "UnknownQuery"
)

// Defines values for WorkspacesQueryKind.
const (
	WorkspacesQueryKindWorkspacesQuery WorkspacesQueryKind = "WorkspacesQuery"
)

// AppInsightsGroupByQuery defines model for AppInsightsGroupByQuery.
type AppInsightsGroupByQuery struct {
	Kind       *AppInsightsGroupByQueryKind `json:"kind,omitempty"`
	MetricName *string                      `json:"metricName,omitempty"`
	RawQuery   *string                      `json:"rawQuery,omitempty"`
}

// AppInsightsGroupByQueryKind defines model for AppInsightsGroupByQuery.Kind.
type AppInsightsGroupByQueryKind string

// AppInsightsMetricNameQuery defines model for AppInsightsMetricNameQuery.
type AppInsightsMetricNameQuery struct {
	Kind     *AppInsightsMetricNameQueryKind `json:"kind,omitempty"`
	RawQuery *string                         `json:"rawQuery,omitempty"`
}

// AppInsightsMetricNameQueryKind defines model for AppInsightsMetricNameQuery.Kind.
type AppInsightsMetricNameQueryKind string

// Azure Monitor Logs sub-query properties
type AzureLogsQuery struct {
	// If set to true the query will be run as a basic logs query
	BasicLogsQuery *bool `json:"basicLogsQuery,omitempty"`

	// Denotes if logs query editor is in builder mode
	BuilderMode *bool `json:"builderMode,omitempty"`

	// If set to true the dashboard time range will be used as a filter for the query. Otherwise the query time ranges will be used. Defaults to false.
	DashboardTime *bool `json:"dashboardTime,omitempty"`

	// @deprecated Use dashboardTime instead
	IntersectTime *bool `json:"intersectTime,omitempty"`

	// KQL query to be executed.
	Query *string `json:"query,omitempty"`

	// @deprecated Use resources instead
	Resource *string `json:"resource,omitempty"`

	// Array of resource URIs to be queried.
	Resources    []string      `json:"resources,omitempty"`
	ResultFormat *ResultFormat `json:"resultFormat,omitempty"`

	// If dashboardTime is set to true this value dictates which column the time filter will be applied to. Defaults to the first tables timeSpan column, the first datetime column found, or TimeGenerated
	TimeColumn *string `json:"timeColumn,omitempty"`

	// Workspace ID. This was removed in Grafana 8, but remains for backwards compat.
	Workspace *string `json:"workspace,omitempty"`
}

// AzureMetricDimension defines model for AzureMetricDimension.
type AzureMetricDimension struct {
	// Name of Dimension to be filtered on.
	Dimension *string `json:"dimension,omitempty"`

	// @deprecated filter is deprecated in favour of filters to support multiselect.
	Filter *string `json:"filter,omitempty"`

	// Values to match with the filter.
	Filters []string `json:"filters,omitempty"`

	// String denoting the filter operation. Supports 'eq' - equals,'ne' - not equals, 'sw' - starts with. Note that some dimensions may not support all operators.
	Operator *string `json:"operator,omitempty"`
}

// AzureMetricQuery defines model for AzureMetricQuery.
type AzureMetricQuery struct {
	// The aggregation to be used within the query. Defaults to the primaryAggregationType defined by the metric.
	Aggregation *string `json:"aggregation,omitempty"`

	// Aliases can be set to modify the legend labels. e.g. {{ resourceGroup }}. See docs for more detail.
	Alias *string `json:"alias,omitempty"`

	// Time grains that are supported by the metric.
	AllowedTimeGrainsMs []int64 `json:"allowedTimeGrainsMs,omitempty"`

	// Used as the value for the metricNamespace property when it's different from the resource namespace.
	CustomNamespace *string `json:"customNamespace,omitempty"`

	// @deprecated This property was migrated to dimensionFilters and should only be accessed in the migration
	Dimension *string `json:"dimension,omitempty"`

	// @deprecated This property was migrated to dimensionFilters and should only be accessed in the migration
	DimensionFilter *string `json:"dimensionFilter,omitempty"`

	// Filters to reduce the set of data returned. Dimensions that can be filtered on are defined by the metric.
	DimensionFilters []AzureMetricDimension `json:"dimensionFilters,omitempty"`

	// @deprecated Use metricNamespace instead
	MetricDefinition *string `json:"metricDefinition,omitempty"`

	// The metric to query data for within the specified metricNamespace. e.g. UsedCapacity
	MetricName *string `json:"metricName,omitempty"`

	// metricNamespace is used as the resource type (or resource namespace).
	// It's usually equal to the target metric namespace. e.g. microsoft.storage/storageaccounts
	// Kept the name of the variable as metricNamespace to avoid backward incompatibility issues.
	MetricNamespace *string `json:"metricNamespace,omitempty"`

	// The Azure region containing the resource(s).
	Region *string `json:"region,omitempty"`

	// @deprecated Use resources instead
	ResourceGroup *string `json:"resourceGroup,omitempty"`

	// @deprecated Use resources instead
	ResourceName *string `json:"resourceName,omitempty"`

	// @deprecated Use resourceGroup, resourceName and metricNamespace instead
	ResourceUri *string `json:"resourceUri,omitempty"`

	// Array of resource URIs to be queried.
	Resources []AzureMonitorResource `json:"resources,omitempty"`

	// The granularity of data points to be queried. Defaults to auto.
	TimeGrain *string `json:"timeGrain,omitempty"`

	// TimeGrainUnit @deprecated
	TimeGrainUnit *string `json:"timeGrainUnit,omitempty"`

	// Maximum number of records to return. Defaults to 10.
	Top *string `json:"top,omitempty"`
}

// AzureMonitorDataQuery defines model for AzureMonitorDataQuery.
type AzureMonitorDataQuery = map[string]any

// AzureMonitorQuery defines model for AzureMonitorQuery.
type AzureMonitorQuery struct {
	// Azure Monitor Logs sub-query properties
	AzureLogAnalytics  *AzureLogsQuery          `json:"azureLogAnalytics,omitempty"`
	AzureMonitor       *AzureMetricQuery        `json:"azureMonitor,omitempty"`
	AzureResourceGraph *AzureResourceGraphQuery `json:"azureResourceGraph,omitempty"`

	// Application Insights Traces sub-query properties
	AzureTraces *AzureTracesQuery `json:"azureTraces,omitempty"`

	// For mixed data sources the selected datasource is on the query level.
	// For non mixed scenarios this is undefined.
	// TODO find a better way to do this ^ that's friendly to schema
	// TODO this shouldn't be unknown but DataSourceRef | null
	Datasource                *any `json:"datasource,omitempty"`
	GrafanaTemplateVariableFn *any `json:"grafanaTemplateVariableFn,omitempty"`

	// If hide is set to true, Grafana will filter out the response(s) associated with this query before returning it to the panel.
	Hide      *bool   `json:"hide,omitempty"`
	Namespace *string `json:"namespace,omitempty"`

	// Used only for exemplar queries from Prometheus
	Query *string `json:"query,omitempty"`

	// Specify the query flavor
	// TODO make this required and give it a default
	QueryType *string `json:"queryType,omitempty"`

	// A unique identifier for the query within the list of targets.
	// In server side expressions, the refId is used as a variable name to identify results.
	// By default, the UI will assign A->Z; however setting meaningful names may be useful.
	RefId    *string `json:"refId,omitempty"`
	Region   *string `json:"region,omitempty"`
	Resource *string `json:"resource,omitempty"`

	// Template variables params. These exist for backwards compatiblity with legacy template variables.
	ResourceGroup *string `json:"resourceGroup,omitempty"`

	// Azure subscription containing the resource(s) to be queried.
	Subscription *string `json:"subscription,omitempty"`

	// Subscriptions to be queried via Azure Resource Graph.
	Subscriptions []string `json:"subscriptions,omitempty"`
}

// AzureMonitorResource defines model for AzureMonitorResource.
type AzureMonitorResource struct {
	MetricNamespace *string `json:"metricNamespace,omitempty"`
	Region          *string `json:"region,omitempty"`
	ResourceGroup   *string `json:"resourceGroup,omitempty"`
	ResourceName    *string `json:"resourceName,omitempty"`
	Subscription    *string `json:"subscription,omitempty"`
}

// Defines the supported queryTypes. GrafanaTemplateVariableFn is deprecated
type AzureQueryType string

// AzureResourceGraphQuery defines model for AzureResourceGraphQuery.
type AzureResourceGraphQuery struct {
	// Azure Resource Graph KQL query to be executed.
	Query *string `json:"query,omitempty"`

	// Specifies the format results should be returned as. Defaults to table.
	ResultFormat *string `json:"resultFormat,omitempty"`
}

// AzureTracesFilter defines model for AzureTracesFilter.
type AzureTracesFilter struct {
	// Values to filter by.
	Filters []string `json:"filters"`

	// Comparison operator to use. Either equals or not equals.
	Operation string `json:"operation"`

	// Property name, auto-populated based on available traces.
	Property string `json:"property"`
}

// Application Insights Traces sub-query properties
type AzureTracesQuery struct {
	// Filters for property values.
	Filters []AzureTracesFilter `json:"filters,omitempty"`

	// Operation ID. Used only for Traces queries.
	OperationId *string `json:"operationId,omitempty"`

	// KQL query to be executed.
	Query *string `json:"query,omitempty"`

	// Array of resource URIs to be queried.
	Resources    []string      `json:"resources,omitempty"`
	ResultFormat *ResultFormat `json:"resultFormat,omitempty"`

	// Types of events to filter by.
	TraceTypes []string `json:"traceTypes,omitempty"`
}

// BaseGrafanaTemplateVariableQuery defines model for BaseGrafanaTemplateVariableQuery.
type BaseGrafanaTemplateVariableQuery struct {
	RawQuery *string `json:"rawQuery,omitempty"`
}

// These are the common properties available to all queries in all datasources.
// Specific implementations will *extend* this interface, adding the required
// properties for the given context.
type DataQuery struct {
	// For mixed data sources the selected datasource is on the query level.
	// For non mixed scenarios this is undefined.
	// TODO find a better way to do this ^ that's friendly to schema
	// TODO this shouldn't be unknown but DataSourceRef | null
	Datasource *any `json:"datasource,omitempty"`

	// If hide is set to true, Grafana will filter out the response(s) associated with this query before returning it to the panel.
	Hide *bool `json:"hide,omitempty"`

	// Specify the query flavor
	// TODO make this required and give it a default
	QueryType *string `json:"queryType,omitempty"`

	// A unique identifier for the query within the list of targets.
	// In server side expressions, the refId is used as a variable name to identify results.
	// By default, the UI will assign A->Z; however setting meaningful names may be useful.
	RefId string `json:"refId"`
}

// GrafanaTemplateVariableQueryType defines model for GrafanaTemplateVariableQueryType.
type GrafanaTemplateVariableQueryType string

// @deprecated Use MetricNamespaceQuery instead
type MetricDefinitionsQuery struct {
	Kind            *MetricDefinitionsQueryKind `json:"kind,omitempty"`
	MetricNamespace *string                     `json:"metricNamespace,omitempty"`
	RawQuery        *string                     `json:"rawQuery,omitempty"`
	ResourceGroup   *string                     `json:"resourceGroup,omitempty"`
	ResourceName    *string                     `json:"resourceName,omitempty"`
	Subscription    *string                     `json:"subscription,omitempty"`
}

// MetricDefinitionsQueryKind defines model for MetricDefinitionsQuery.Kind.
type MetricDefinitionsQueryKind string

// MetricNamesQuery defines model for MetricNamesQuery.
type MetricNamesQuery struct {
	Kind            *MetricNamesQueryKind `json:"kind,omitempty"`
	MetricNamespace *string               `json:"metricNamespace,omitempty"`
	RawQuery        *string               `json:"rawQuery,omitempty"`
	ResourceGroup   *string               `json:"resourceGroup,omitempty"`
	ResourceName    *string               `json:"resourceName,omitempty"`
	Subscription    *string               `json:"subscription,omitempty"`
}

// MetricNamesQueryKind defines model for MetricNamesQuery.Kind.
type MetricNamesQueryKind string

// MetricNamespaceQuery defines model for MetricNamespaceQuery.
type MetricNamespaceQuery struct {
	Kind            *MetricNamespaceQueryKind `json:"kind,omitempty"`
	MetricNamespace *string                   `json:"metricNamespace,omitempty"`
	RawQuery        *string                   `json:"rawQuery,omitempty"`
	ResourceGroup   *string                   `json:"resourceGroup,omitempty"`
	ResourceName    *string                   `json:"resourceName,omitempty"`
	Subscription    *string                   `json:"subscription,omitempty"`
}

// MetricNamespaceQueryKind defines model for MetricNamespaceQuery.Kind.
type MetricNamespaceQueryKind string

// ResourceGroupsQuery defines model for ResourceGroupsQuery.
type ResourceGroupsQuery struct {
	Kind         *ResourceGroupsQueryKind `json:"kind,omitempty"`
	RawQuery     *string                  `json:"rawQuery,omitempty"`
	Subscription *string                  `json:"subscription,omitempty"`
}

// ResourceGroupsQueryKind defines model for ResourceGroupsQuery.Kind.
type ResourceGroupsQueryKind string

// ResourceNamesQuery defines model for ResourceNamesQuery.
type ResourceNamesQuery struct {
	Kind            *ResourceNamesQueryKind `json:"kind,omitempty"`
	MetricNamespace *string                 `json:"metricNamespace,omitempty"`
	RawQuery        *string                 `json:"rawQuery,omitempty"`
	ResourceGroup   *string                 `json:"resourceGroup,omitempty"`
	Subscription    *string                 `json:"subscription,omitempty"`
}

// ResourceNamesQueryKind defines model for ResourceNamesQuery.Kind.
type ResourceNamesQueryKind string

// ResultFormat defines model for ResultFormat.
type ResultFormat string

// SubscriptionsQuery defines model for SubscriptionsQuery.
type SubscriptionsQuery struct {
	Kind     *SubscriptionsQueryKind `json:"kind,omitempty"`
	RawQuery *string                 `json:"rawQuery,omitempty"`
}

// SubscriptionsQueryKind defines model for SubscriptionsQuery.Kind.
type SubscriptionsQueryKind string

// UnknownQuery defines model for UnknownQuery.
type UnknownQuery struct {
	Kind     *UnknownQueryKind `json:"kind,omitempty"`
	RawQuery *string           `json:"rawQuery,omitempty"`
}

// UnknownQueryKind defines model for UnknownQuery.Kind.
type UnknownQueryKind string

// WorkspacesQuery defines model for WorkspacesQuery.
type WorkspacesQuery struct {
	Kind         *WorkspacesQueryKind `json:"kind,omitempty"`
	RawQuery     *string              `json:"rawQuery,omitempty"`
	Subscription *string              `json:"subscription,omitempty"`
}

// WorkspacesQueryKind defines model for WorkspacesQuery.Kind.
type WorkspacesQueryKind string
