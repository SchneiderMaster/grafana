import { css } from '@emotion/css';
import { useEffect } from 'react';
import { useToggle } from 'react-use';

import { GrafanaTheme2, store } from '@grafana/data';
import { ToolbarButton, useStyles2 } from '@grafana/ui';
import { appEvents } from 'app/core/app_events';
import { t } from 'app/core/internationalization';
import { RecordHistoryEntryEvent } from 'app/types/events';

import { HISTORY_LOCAL_STORAGE_KEY } from '../AppChromeService';
import { NavToolbarSeparator } from '../NavToolbar/NavToolbarSeparator';
import { HistoryEntry } from '../types';

import { HistoryDrawer } from './HistoryDrawer';

export function HistoryContainer() {
  const [showHistoryDrawer, onToggleShowHistoryDrawer] = useToggle(false);
  const styles = useStyles2(getStyles);

  useEffect(() => {
    const sub = appEvents.subscribe(RecordHistoryEntryEvent, (ev) => {
      const clickedHistory = store.getObject<boolean>('CLICKING_HISTORY');
      if (clickedHistory) {
        store.setObject('CLICKING_HISTORY', false);
        return;
      }
      const history = store.getObject<HistoryEntry[]>(HISTORY_LOCAL_STORAGE_KEY, []);
      let lastEntry = history[0];
      const newUrl = ev.payload.url;
      const lastUrl = lastEntry.views[0]?.url;
      if (lastUrl !== newUrl) {
        lastEntry.views = [
          {
            name: ev.payload.name,
            description: ev.payload.description,
            url: newUrl,
            time: Date.now(),
          },
          ...lastEntry.views,
        ];
        store.setObject(HISTORY_LOCAL_STORAGE_KEY, [...history]);
      }
      return () => {
        sub.unsubscribe();
      };
    });
  }, []);

  return (
    <>
      <ToolbarButton
        onClick={onToggleShowHistoryDrawer}
        iconOnly
        icon="history"
        aria-label={t('nav.history-container.drawer-tittle', 'History')}
      />
      <NavToolbarSeparator className={styles.separator} />
      {showHistoryDrawer && <HistoryDrawer onClose={() => onToggleShowHistoryDrawer(false)} />}
    </>
  );
}

const getStyles = (theme: GrafanaTheme2) => {
  return {
    separator: css({
      [theme.breakpoints.down('sm')]: {
        display: 'none',
      },
    }),
  };
};
