import { css, cx } from '@emotion/css';
import { Draggable } from '@hello-pangea/dnd';

import { Action, DataFrame, GrafanaTheme2 } from '@grafana/data';
import { Badge } from '@grafana/ui';
import { Icon } from '@grafana/ui/src/components/Icon/Icon';
import { IconButton } from '@grafana/ui/src/components/IconButton/IconButton';
import { useStyles2 } from '@grafana/ui/src/themes';
import { t } from '@grafana/ui/src/utils/i18n';

export interface ActionsListItemProps {
  index: number;
  action: Action;
  data: DataFrame[];
  onChange: (index: number, action: Action) => void;
  onEdit: () => void;
  onRemove: () => void;
  isEditing?: boolean;
  itemKey: string;
}

export const ActionListItem = ({ action, onEdit, onRemove, index, itemKey }: ActionsListItemProps) => {
  const styles = useStyles2(getActionListItemStyles);
  const { title = '', oneClick = false } = action;

  const hasTitle = title.trim() !== '';

  return (
    <Draggable key={itemKey} draggableId={itemKey} index={index}>
      {(provided) => (
        <div
          className={cx(styles.wrapper, styles.dragRow)}
          ref={provided.innerRef}
          {...provided.draggableProps}
          key={index}
        >
          <div className={styles.linkDetails}>
            <div className={cx(styles.url, !hasTitle && styles.notConfigured)}>
              {hasTitle ? title : 'Title not provided'}
            </div>
          </div>
          <div className={styles.icons}>
            {oneClick && (
              <Badge
                color="blue"
                text={t('grafana-ui.data-links-inline-editor.one-click', 'One click')}
                tooltip={t('grafana-ui.data-links-inline-editor.one-click-enabled', 'One click enabled')}
              />
            )}
            <IconButton name="pen" onClick={onEdit} className={styles.icon} tooltip="Edit" />
            <IconButton name="trash-alt" onClick={onRemove} className={styles.icon} tooltip="Remove" />
            <div className={styles.dragIcon} {...provided.dragHandleProps}>
              <Icon name="draggabledots" size="lg" />
            </div>
          </div>
        </div>
      )}
    </Draggable>
  );
};

const getActionListItemStyles = (theme: GrafanaTheme2) => {
  return {
    wrapper: css({
      display: 'flex',
      flexGrow: 1,
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '5px 0 5px 10px',
      borderRadius: theme.shape.radius.default,
      background: theme.colors.background.secondary,
      gap: 8,
    }),
    linkDetails: css({
      display: 'flex',
      flexDirection: 'column',
      flexGrow: 1,
      maxWidth: `calc(100% - 100px)`,
    }),
    errored: css({
      color: theme.colors.error.text,
      fontStyle: 'italic',
    }),
    notConfigured: css({
      fontStyle: 'italic',
    }),
    title: css({
      color: theme.colors.text.primary,
      fontSize: theme.typography.size.sm,
      fontWeight: theme.typography.fontWeightMedium,
    }),
    url: css({
      color: theme.colors.text.secondary,
      fontSize: theme.typography.size.sm,
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
    }),
    dragRow: css({
      position: 'relative',
      margin: '8px',
    }),
    icons: css({
      display: 'flex',
      padding: 6,
      alignItems: 'center',
      gap: 8,
    }),
    dragIcon: css({
      cursor: 'grab',
      color: theme.colors.text.secondary,
      margin: theme.spacing(0, 0.5),
    }),
    icon: css({
      color: theme.colors.text.secondary,
    }),
  };
};
