import { noop } from 'lodash';
import { useCallback, useMemo } from 'react';
import { useAsync } from 'react-use';

import { CoreApp, DataSourcePluginContextProvider, LoadingState } from '@grafana/data';
import { PromQuery } from '@grafana/prometheus';
import { getDataSourceSrv } from '@grafana/runtime';
import { DataQuery } from '@grafana/schema';
import { Alert, Button, Stack } from '@grafana/ui';
import { LokiQuery } from 'app/plugins/datasource/loki/types';

import { CloudAlertPreview } from './CloudAlertPreview';
import { usePreview } from './PreviewRule';

export interface ExpressionEditorProps {
  value?: string;
  onChange: (value: string) => void;
  dataSourceName: string; // will be a prometheus or loki datasource
  showPreview: boolean;
}

export const ExpressionEditor = ({ value, onChange, dataSourceName, showPreview = true }: ExpressionEditorProps) => {
  const { mapToValue, mapToQuery } = useQueryMappers(dataSourceName);
  const dataQuery = mapToQuery({ refId: 'A', hide: false }, value);

  const {
    error,
    loading,
    value: dataSource,
  } = useAsync(() => {
    return getDataSourceSrv().get(dataSourceName);
  }, [dataSourceName]);

  const onChangeQuery = useCallback(
    (query: DataQuery) => {
      onChange(mapToValue(query));
    },
    [onChange, mapToValue]
  );

  const [alertPreview, onPreview] = usePreview();

  const onRunQueriesClick = async () => {
    onPreview();
  };

  if (loading || dataSource?.name !== dataSourceName) {
    return null;
  }

  const dsi = getDataSourceSrv().getInstanceSettings(dataSourceName);

  if (error || !dataSource || !dataSource?.components?.QueryEditor || !dsi) {
    const errorMessage = error?.message || 'Data source plugin does not export any Query Editor component';
    return <div>Could not load query editor due to: {errorMessage}</div>;
  }

  const previewLoaded = alertPreview?.data.state === LoadingState.Done;

  const QueryEditor = dataSource?.components?.QueryEditor;

  // The Preview endpoint returns the preview as a single-element array of data frames
  const previewDataFrame = alertPreview?.data?.series?.find((s) => s.name === 'evaluation results');
  // The preview API returns arrays with empty elements when there are no firing alerts
  const previewHasAlerts = previewDataFrame && previewDataFrame.fields.some((field) => field.values.length > 0);

  return (
    <Stack direction="column" gap={1}>
      <DataSourcePluginContextProvider instanceSettings={dsi}>
        <QueryEditor
          query={dataQuery}
          queries={[dataQuery]}
          app={CoreApp.CloudAlerting}
          onChange={onChangeQuery}
          onRunQuery={noop}
          datasource={dataSource}
        />
      </DataSourcePluginContextProvider>
      {showPreview && (
        <Stack direction="column" gap={1} alignItems="flex-start">
          <Button
            type="button"
            onClick={onRunQueriesClick}
            disabled={alertPreview?.data.state === LoadingState.Loading || !value}
          >
            Preview alerts
          </Button>
          {previewLoaded && !previewHasAlerts && (
            <Alert title="Alerts preview" severity="info">
              There are no firing alerts for your query.
            </Alert>
          )}
          {previewHasAlerts && <CloudAlertPreview preview={previewDataFrame} />}
        </Stack>
      )}
    </Stack>
  );
};

type QueryMappers<T extends DataQuery = DataQuery> = {
  mapToValue: (query: T) => string;
  mapToQuery: (existing: T, value: string | undefined) => T;
};

export function useQueryMappers(dataSourceName: string): QueryMappers {
  return useMemo(() => {
    const settings = getDataSourceSrv().getInstanceSettings(dataSourceName);

    switch (settings?.type) {
      case 'loki':
      case 'prometheus':
        return {
          mapToValue: (query: DataQuery) => (query as PromQuery | LokiQuery).expr,
          mapToQuery: (existing: DataQuery, value: string | undefined) => ({ ...existing, expr: value }),
        };
      default:
        throw new Error(`${dataSourceName} is not supported as an expression editor`);
    }
  }, [dataSourceName]);
}
