import { ComponentProps, useMemo, useState } from 'react';

import { config } from '@grafana/runtime';
import { Alert, Box, Button, Drawer, EmptyState, LoadingPlaceholder, Stack, Text, Tooltip } from '@grafana/ui';
import { RevisionModel, VersionHistoryComparison } from 'app/core/components/VersionHistory/VersionHistoryComparison';
import { Trans, t } from 'app/core/internationalization';
import { GrafanaRuleDefinition, RulerGrafanaRuleDTO } from 'app/types/unified-alerting-dto';

import { LogMessages, logInfo, trackRuleVersionsComparisonClick } from '../../../Analytics';
import { alertRuleApi } from '../../../api/alertRuleApi';
import { stringifyErrorLike } from '../../../utils/misc';

import { VersionHistoryTable } from './components/VersionHistoryTable';
import { getSpecialUidMap, preprocessRuleForDiffDisplay } from './versions-utils';

const { useGetAlertVersionHistoryQuery } = alertRuleApi;

interface AlertVersionHistoryProps {
  ruleUid: string;
}

/** List of (top level) properties to exclude from being shown in human readable summary of version changes */
export const grafanaAlertPropertiesToIgnore: Array<keyof GrafanaRuleDefinition> = [
  'id',
  'uid',
  'updated',
  'updated_by',
  'version',
];

/**
 * Render the version history of a given Grafana managed alert rule, showing different edits
 * and allowing to restore to a previous version.
 */
export function AlertVersionHistory({ ruleUid }: AlertVersionHistoryProps) {
  const SPECIAL_UID_MAP = getSpecialUidMap();
  const { isLoading, currentData: ruleVersions = [], error } = useGetAlertVersionHistoryQuery({ uid: ruleUid });

  const [oldVersion, setOldVersion] = useState<RulerGrafanaRuleDTO<GrafanaRuleDefinition>>();
  const [newVersion, setNewVersion] = useState<RulerGrafanaRuleDTO<GrafanaRuleDefinition>>();
  const [showDrawer, setShowDrawer] = useState(false);
  // checked versions for comparison. key is the version number, value is whether it's checked
  const [checkedVersions, setCheckedVersions] = useState<Map<string, boolean>>(new Map());
  const canCompare = useMemo(
    () => Array.from(checkedVersions.values()).filter((value) => value).length > 1,
    [checkedVersions]
  );

  if (error) {
    return (
      <Alert title={t('alerting.alertVersionHistory.errorloading', 'Failed to load alert rule versions')}>
        {stringifyErrorLike(error)}
      </Alert>
    );
  }

  if (isLoading) {
    return <LoadingPlaceholder text={t('alerting.common.loading', 'Loading...')} />;
  }

  if (!ruleVersions.length) {
    // We don't expect this to happen - all alert rules _should_ have at least one version
    logInfo(LogMessages.noAlertRuleVersionsFound, { ruleUid });
    return (
      <EmptyState
        variant="not-found"
        message={t('alerting.alertVersionHistory.noVersionsFound', 'No versions found for this rule')}
      />
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

  /**
   * Turns a version of a Grafana rule definition into data structure
   * used to display the version summary when comparing versions
   */
  const parseVersionInfo = (version: RulerGrafanaRuleDTO<GrafanaRuleDefinition>): RevisionModel => {
    const unknown = t('alerting.alertVersionHistory.unknown', 'Unknown');
    const createdBy = (() => {
      const updatedBy = version?.grafana_alert.updated_by;
      const uid = updatedBy?.uid;
      const name = updatedBy?.name;

      if (!updatedBy) {
        return unknown;
      }
      if (uid && SPECIAL_UID_MAP[uid]) {
        return SPECIAL_UID_MAP[uid].name;
      }
      if (name) {
        return name;
      }
      return uid ? t('alerting.alertVersionHistory.user-id', 'User ID {{uid}}', { uid }) : unknown;
    })();

    return {
      created: version.grafana_alert.updated || unknown,
      createdBy,
      version: version.grafana_alert.version || unknown,
    };
  };

  return (
    <Stack direction="column" gap={2}>
      <Text variant="body">
        <Trans i18nKey="alerting.alertVersionHistory.description">
          Each time you edit the alert rule a new version is created. You can restore an older version. Select two
          versions below and compare their differences.
        </Trans>
      </Text>
      <Stack>
        <Tooltip
          content={t('core.versionHistory.comparison.select', 'Select two versions to start comparing')}
          placement="bottom"
        >
          <Button type="button" disabled={!canCompare} onClick={compareVersions} icon="code-branch">
            <Trans i18nKey="alerting.alertVersionHistory.compareVersions">Compare versions</Trans>
          </Button>
        </Tooltip>
      </Stack>
      {showDrawer && oldVersion && newVersion && (
        <Drawer
          onClose={() => setShowDrawer(false)}
          title={t('alerting.alertVersionHistory.comparing-versions', 'Comparing versions')}
        >
          <VersionHistoryComparison
            oldSummary={parseVersionInfo(oldVersion)}
            oldVersion={oldVersion}
            newSummary={parseVersionInfo(newVersion)}
            newVersion={newVersion}
            preprocessVersion={preprocessRuleForDiffDisplay}
          />
          {config.featureToggles.alertingRuleVersionHistoryRestore && (
            <Box paddingTop={2}>
              <Stack justifyContent="flex-end">
                <Button variant="destructive" onClick={() => {}}>
                  <Trans i18nKey="alerting.alertVersionHistory.reset">
                    Reset to version {{ version: oldVersion.grafana_alert.version }}
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
