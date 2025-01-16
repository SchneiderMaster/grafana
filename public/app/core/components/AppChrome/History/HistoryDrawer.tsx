import { css, cx } from '@emotion/css';

import { GrafanaTheme2 } from '@grafana/data';
import { IconButton, useStyles2, Text, Box } from '@grafana/ui';
import { useGrafana } from 'app/core/context/GrafanaContext';
import { t } from 'app/core/internationalization';

import { TOP_BAR_LEVEL_HEIGHT } from '../types';

import { HistoryWrapper } from './HistoryWrapper';

export const HISTORY_DRAWER_WIDTH = '300px';
const DOCK_HISTORY_BUTTON_ID = 'dock-history-button';
const HISTORY_HEADER_TOGGLE_ID = 'history-header-toggle';

export const HistoryDrawer = () => {
  const { chrome } = useGrafana();
  const state = chrome.useState();
  const styles = useStyles2(getStyles);
  return (
    <div className={cx(styles.drawer, state.historyDocked && styles.dockedDrawer)}>
      <HistoryHeader />
      <HistoryWrapper />
    </div>
  );
};

export const HistoryHeader = () => {
  const { chrome } = useGrafana();
  const state = chrome.useState();
  const styles = useStyles2(getStyles);
  const handleMegaMenu = () => {
    chrome.setHistoryOpen(!state.historyOpen);
  };

  const handleDockedMenu = () => {
    chrome.setHistoryDocked(!state.historyDocked);
  };
  return (
    <div className={styles.header}>
      <Text element="h6">{t('nav.history-container.drawer-tittle', 'History')}</Text>
      <Box justifyContent={'center'} display={'flex'}>
        <IconButton
          id={DOCK_HISTORY_BUTTON_ID}
          tooltip={
            state.historyDocked ? t('nav.history-drawer.undock', 'Undock') : t('nav.history-drawer.dock', 'Dock')
          }
          name="web-section-alt"
          onClick={handleDockedMenu}
          variant="secondary"
          className={styles.dockedHistoryButton}
        />
        <IconButton
          id={HISTORY_HEADER_TOGGLE_ID}
          tooltip={t('nav.history-drawer.close', 'Close')}
          name="times"
          onClick={handleMegaMenu}
          size="xl"
          variant="secondary"
        />
      </Box>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => {
  return {
    drawer: css({
      maxWidth: HISTORY_DRAWER_WIDTH,
      width: '100%',
      boxSizing: 'border-box',
      height: `calc(100vh - ${TOP_BAR_LEVEL_HEIGHT}px)`,
      overflowY: 'auto',
      overflowX: 'hidden',
      background: theme.colors.background.secondary,
      borderLeft: `1px solid ${theme.colors.border.medium}`,
      boxShadow: theme.shadows.z2,
      zIndex: theme.zIndex.modal,
      position: 'fixed',
      top: TOP_BAR_LEVEL_HEIGHT,
      right: 0,
      display: 'flex',
      flexDirection: 'column',
      padding: theme.spacing(2),
    }),
    dockedDrawer: css({
      background: theme.colors.background.primary,
      borderLeft: `1px solid ${theme.colors.border.weak}`,
      display: 'none',
      zIndex: theme.zIndex.sidemenu,
      boxShadow: 'none',
      [theme.breakpoints.up('xl')]: {
        display: 'block',
      },
    }),
    header: css({
      alignItems: 'center',
      display: 'flex',
      gap: theme.spacing(1),
      justifyContent: 'space-between',
      paddingBottom: theme.spacing(2),
      height: TOP_BAR_LEVEL_HEIGHT,
      minHeight: TOP_BAR_LEVEL_HEIGHT,
    }),
    dockedHistoryButton: css({
      display: 'none',
      [theme.breakpoints.up('xl')]: {
        display: 'inline-flex',
      },
    }),
  };
};
