import moment from 'moment';
import { ComponentProps, useMemo, useState } from 'react';

import { dateTimeFormatTimeAgo } from '@grafana/data';
import { config } from '@grafana/runtime';
import {
  Alert,
  Badge,
  Box,
  Button,
  Checkbox,
  Column,
  ConfirmModal,
  Drawer,
  InteractiveTable,
  LoadingPlaceholder,
  Stack,
  Text,
  Tooltip,
} from '@grafana/ui';
import { RevisionModel, VersionHistoryComparison } from 'app/core/components/VersionHistory/VersionHistoryComparison';
import { Trans, t } from 'app/core/internationalization';
import { DiffGroup } from 'app/features/dashboard-scene/settings/version-history/DiffGroup';
import { Diffs, jsonDiff } from 'app/features/dashboard-scene/settings/version-history/utils';
import { GrafanaRuleDefinition, RulerGrafanaRuleDTO } from 'app/types/unified-alerting-dto';

import { trackRuleVersionsComparisonClick } from '../../../Analytics';
import { alertRuleApi } from '../../../api/alertRuleApi';
import { stringifyErrorLike } from '../../../utils/misc';

const { useGetAlertVersionHistoryQuery } = alertRuleApi;

export interface AlertVersionHistoryProps {
  ruleUid: string;
}

const VERSIONS_PAGE_SIZE = 20;

const grafanaAlertPropertiesToIgnore: Array<keyof GrafanaRuleDefinition> = [
  'id',
  'uid',
  'updated',
  'updated_by',
  'version',
];

function preprocessRuleForDiffDisplay(rulerRule: RulerGrafanaRuleDTO<GrafanaRuleDefinition>) {
  const { grafana_alert, ...rest } = rulerRule;

  // translations for properties not in `grafana_alert`
  const translationMap: Record<string, string> = {
    for: t('alerting.alertVersionHistory.pendingPeriod', 'Pending period'),
    annotations: t('alerting.alertVersionHistory.annotations', 'Annotations'),
    labels: t('alerting.alertVersionHistory.labels', 'Labels'),
  };

  // translations for properties in `grafana_alert`
  const grafanaAlertTranslationMap: Partial<Record<keyof GrafanaRuleDefinition, string>> = {
    title: t('alerting.alertVersionHistory.name', 'Name'),
    namespace_uid: t('alerting.alertVersionHistory.namespace_uid', 'Folder UID'),
    data: t('alerting.alertVersionHistory.queryAndAlertCondition', 'Query and alert condition'),
    notification_settings: t('alerting.alertVersionHistory.contactPointRouting', 'Contact point routing'),
    no_data_state: t('alerting.alertVersionHistory.noDataState', 'Alert state when no data'),
    exec_err_state: t('alerting.alertVersionHistory.execErrorState', 'Alert state when execution error'),
    is_paused: t('alerting.alertVersionHistory.paused', 'Paused state'),
    rule_group: t('alerting.alertVersionHistory.rule_group', 'Rule group'),
    condition: t('alerting.alertVersionHistory.condition', 'Alert condition'),
  };

  const processedTopLevel = Object.entries(rest).reduce((acc, [key, value]) => {
    const translation = translationMap[key] || key;
    return {
      ...acc,
      [translation]: value,
    };
  }, {});

  const processedGrafanaAlert = Object.entries(grafana_alert).reduce((acc, [key, value]) => {
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const grafanaRuleKey = key as keyof GrafanaRuleDefinition;

    if (grafanaAlertPropertiesToIgnore.includes(grafanaRuleKey)) {
      return acc;
    }

    const potentiallyTranslatedKey = grafanaAlertTranslationMap[grafanaRuleKey] || key;
    return {
      ...acc,
      [potentiallyTranslatedKey]: value,
    };
  }, {});

  return {
    ...processedTopLevel,
    ...processedGrafanaAlert,
  };
}

/**
 * Render the version history of a given Grafana managed alert rule, showing different edits
 * and allowing to restore to a previous version.
 */
export function AlertVersionHistory({ ruleUid }: AlertVersionHistoryProps) {
  const { isLoading, currentData: ruleVersions = [], error } = useGetAlertVersionHistoryQuery({ uid: ruleUid });

  const [oldVersion, setOldVersion] = useState<RulerGrafanaRuleDTO<GrafanaRuleDefinition>>();
  const [newVersion, setNewVersion] = useState<RulerGrafanaRuleDTO<GrafanaRuleDefinition>>();
  const [showDrawer, setShowDrawer] = useState(false);
  const [checkedVersions, setCheckedVersions] = useState<Map<string, boolean>>(new Map());
  const canCompare = useMemo(
    () => Array.from(checkedVersions.values()).filter((value) => value).length > 1,
    [checkedVersions]
  );

  if (error) {
    return (
      <Alert title={t('alerting.alertVersionHistory.errorloading', 'Failed to load alert rule history')}>
        {stringifyErrorLike(error)}
      </Alert>
    );
  }

  if (isLoading) {
    return <LoadingPlaceholder text={t('alerting.common.loading', 'Loading...')} />;
  }

  if (!ruleVersions.length) {
    return (
      <Trans i18nKey="alerting.alertVersionHistory.noVersionsFound">
        {/* I think this is not possible? */}
        No versions found for this rule
      </Trans>
    );
  }

  const compareVersions = () => {
    const [older, newer] = ruleVersions
      .filter((rule) => {
        const version = rule.grafana_alert.version;
        if (!version && version !== 0) {
          return;
        }
        return checkedVersions.get(String(rule.grafana_alert.version));
      })
      .sort((a, b) => {
        const aVersion = a.grafana_alert.version;
        const bVersion = b.grafana_alert.version;
        if (aVersion === undefined || bVersion === undefined) {
          return 0;
        }
        return aVersion - bVersion;
      });

    trackRuleVersionsComparisonClick({
      latest: newer === ruleVersions[0],
      oldVersion: older?.grafana_alert.version || 0,
      newVersion: newer?.grafana_alert.version || 0,
    });

    setOldVersion(older);
    setNewVersion(newer);
    setShowDrawer(true);
  };

  const handleCheckedVersionChange: ComponentProps<typeof VersionHistoryTable>['onVersionsChecked'] = (versions) => {
    setCheckedVersions(versions);
    setOldVersion(undefined);
    setNewVersion(undefined);
  };

  // const resetSelection = () => {
  //   setCheckedVersions(new Map());
  //   setOldVersion(undefined);
  //   setNewVersion(undefined);
  // };

  return (
    <Stack direction="column" gap={2}>
      <Text variant="body">
        <Trans i18nKey="alerting.alertVersionHistory.description">
          Each time you edit the alert rule a new version is created. You can restore an older version. Select two
          versions below and compare their differences.
        </Trans>
      </Text>
      <Stack>
        <Tooltip content="Select two versions to start comparing" placement="bottom">
          <Button type="button" disabled={!canCompare} onClick={compareVersions} icon="code-branch">
            <Trans i18nKey="alerting.alertVersionHistory.compareVersions">Compare versions</Trans>
          </Button>
        </Tooltip>
        {/* {checkedVersions.size > 0 && (
          <Button type="button" onClick={resetSelection} icon="times" variant="secondary">
            <Trans i18nKey="alerting.alertVersionHistory.clearSelection">Clear selected versions</Trans>
          </Button>
        )} */}
      </Stack>
      {showDrawer && oldVersion && newVersion && (
        <Drawer
          onClose={() => setShowDrawer(false)}
          title={t('alerting.alertVersionHistory.comparing-versions', 'Comparing versions')}
        >
          <VersionHistoryComparison
            oldInfo={{
              created: oldVersion.grafana_alert.updated || 'unknown',
              createdBy: oldVersion.grafana_alert.updated_by?.name || 'unknown',
              version: oldVersion.grafana_alert.version || 'unknown',
            }}
            oldVersion={oldVersion}
            newInfo={{
              created: newVersion.grafana_alert.updated || 'unknown',
              createdBy: newVersion.grafana_alert.updated_by?.name || 'unknown',
              version: newVersion.grafana_alert.version || 'unknown',
            }}
            newVersion={newVersion}
            preprocessVersion={preprocessRuleForDiffDisplay}
          />
          {config.featureToggles.alertingRuleVersionHistory && (
            <Box paddingTop={2}>
              <Stack justifyContent="flex-end">
                <Button variant="destructive" onClick={() => {}}>
                  <Trans i18nKey="alerting.alertVersionHistory.reset">
                    Reset to version {oldVersion.grafana_alert.version}
                  </Trans>
                </Button>
              </Stack>
            </Box>
          )}
        </Drawer>
      )}

      <VersionHistoryTable
        onVersionsChecked={handleCheckedVersionChange}
        ruleVersions={ruleVersions}
        disableSelection={canCompare}
      />
    </Stack>
  );
}

function VersionHistoryTable({
  onVersionsChecked,
  ruleVersions,
  disableSelection,
}: {
  onVersionsChecked: (value: Map<string, boolean>) => void;
  ruleVersions: Array<RulerGrafanaRuleDTO<GrafanaRuleDefinition>>;
  disableSelection: boolean;
}) {
  const [restoreDiff, setRestoreDiff] = useState<Diffs | undefined>();
  const [checkedVersions, setCheckedVersions] = useState<Map<string, boolean>>(new Map());

  const [confirmRestore, setConfirmRestore] = useState(false);

  const showConfirmation = (id: string) => {
    const currentVersion = ruleVersions[0];
    const restoreVersion = ruleVersions.find((rule) => String(rule.grafana_alert.version) === id);
    if (!restoreVersion) {
      return;
    }

    setConfirmRestore(true);
    setRestoreDiff(jsonDiff(currentVersion, restoreVersion));
  };

  const hideConfirmation = () => {
    setConfirmRestore(false);
  };

  const rows: RevisionModel[] = ruleVersions.map((rule, index) => ({
    id: String(rule.grafana_alert.version),
    version: rule.grafana_alert.version || `unknown-rule-${index}`,
    created: rule.grafana_alert.updated || 'unknown',
    createdBy: rule.grafana_alert.updated_by?.name || 'unknown',
  }));

  const columns: Array<Column<RevisionModel>> = [
    {
      disableGrow: true,
      id: 'id',
      header: 'Version',
      cell: ({ value }) => {
        const thisValue = checkedVersions.get(String(value ?? false)) ?? false;
        return (
          <Stack direction="row">
            <Checkbox
              label={value}
              checked={thisValue}
              disabled={disableSelection && !thisValue}
              onChange={() => {
                setCheckedVersions((prevState) => {
                  const newState = new Map(prevState);
                  newState.set(String(value), !prevState.get(String(value)));
                  onVersionsChecked(newState);
                  return newState;
                });
              }}
            />
          </Stack>
        );
      },
    },
    {
      id: 'createdBy',
      header: 'Updated By',
      disableGrow: true,
      cell: ({ value }) => value,
    },
    {
      id: 'created',
      header: 'Date',
      disableGrow: true,
      cell: ({ value }) => moment(value).toLocaleString(),
    },
    {
      id: 'timeSince',
      disableGrow: true,
      cell: ({ row }) => dateTimeFormatTimeAgo(row.values.created),
    },
    {
      id: 'actions',
      disableGrow: true,
      cell: ({ row }) => {
        const isFirstItem = row.index === 0;

        return (
          <Stack direction="row" alignItems="center" justifyContent="flex-end">
            {isFirstItem ? (
              <Badge text={t('alerting.alertVersionHistory.latest', 'Latest')} color="blue" />
            ) : config.featureToggles.alertingRuleVersionHistory ? (
              <Button
                variant="secondary"
                size="sm"
                icon="history"
                onClick={() => {
                  showConfirmation(row.values.id);
                }}
              >
                <Trans i18nKey="alerting.alertVersionHistory.restore">Restore</Trans>
              </Button>
            ) : null}
          </Stack>
        );
      },
    },
  ];

  return (
    <>
      <InteractiveTable
        pageSize={VERSIONS_PAGE_SIZE}
        columns={columns}
        data={rows}
        getRowId={(row) => `${row.version}`}
      />
      <ConfirmModal
        isOpen={confirmRestore}
        title={t('alerting.alertVersionHistory.restore-modal.title', 'Restore Version')}
        body={
          <Stack direction="column" gap={2}>
            <Trans i18nKey="alerting.alertVersionHistory.restore-modal.body">
              Are you sure you want to restore the alert rule definition to this version? All unsaved changes will be
              lost.
            </Trans>
            <Text variant="h6">
              <Trans i18nKey="alerting.alertVersionHistory.restore-modal.summary">
                Summary of changes to be applied:
              </Trans>
            </Text>
            <div>
              {restoreDiff && (
                <>
                  {Object.entries(restoreDiff).map(([key, diffs]) => (
                    <DiffGroup diffs={diffs} key={key} title={key} />
                  ))}
                </>
              )}
            </div>
          </Stack>
        }
        confirmText={'Yes, restore configuration'}
        onConfirm={() => {
          // if (activeRestoreVersion) {
          //   restoreVersion(activeRestoreVersion);
          // }

          hideConfirmation();
        }}
        onDismiss={() => hideConfirmation()}
      />
    </>
  );
}
