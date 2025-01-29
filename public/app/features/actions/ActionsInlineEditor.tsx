import { css } from '@emotion/css';
import { DragDropContext, Droppable, DropResult } from '@hello-pangea/dnd';
import { cloneDeep } from 'lodash';
import { useEffect, useState } from 'react';

import { Action, DataFrame, GrafanaTheme2, defaultActionConfig, VariableSuggestion } from '@grafana/data';
import { Button } from '@grafana/ui/src/components/Button';
import { Modal } from '@grafana/ui/src/components/Modal/Modal';
import { useStyles2 } from '@grafana/ui/src/themes';
import { Trans } from 'app/core/internationalization';

import { ActionEditorModalContent } from './ActionEditorModalContent';
import { ActionListItem } from './ActionsListItem';

interface ActionsInlineEditorProps {
  actions?: Action[];
  onChange: (actions: Action[]) => void;
  data: DataFrame[];
  getSuggestions: () => VariableSuggestion[];
}

export const ActionsInlineEditor = ({
  actions,
  onChange,
  data,
  getSuggestions,
}: ActionsInlineEditorProps) => {
  const [editIndex, setEditIndex] = useState<number | null>(null);
  const [isNew, setIsNew] = useState(false);

  const [actionsSafe, setActionsSafe] = useState<Action[]>([]);

  useEffect(() => {
    setActionsSafe(actions ?? []);
  }, [actions]);

  const styles = useStyles2(getActionsInlineEditorStyle);
  const isEditing = editIndex !== null;

  const onActionChange = (index: number, action: Action) => {
    if (isNew) {
      if (action.title.trim() === '') {
        setIsNew(false);
        setEditIndex(null);
        return;
      } else {
        setEditIndex(null);
        setIsNew(false);
      }
    }

    if (action.oneClick === true) {
      actionsSafe.forEach((action) => {
        action.oneClick = false;
      });
    }

    const update = cloneDeep(actionsSafe);
    update[index] = action;
    onChange(update);

    setEditIndex(null);
  };

  const onActionAdd = () => {
    let update = cloneDeep(actionsSafe);
    setEditIndex(update.length);
    setIsNew(true);
  };

  const onActionCancel = (index: number) => {
    if (isNew) {
      setIsNew(false);
    }
    setEditIndex(null);
  };

  const onActionRemove = (index: number) => {
    const update = cloneDeep(actionsSafe);
    update.splice(index, 1);
    onChange(update);
  };

  const onDragEnd = (result: DropResult) => {
    if (!actions || !result.destination) {
      return;
    }

    const update = cloneDeep(actionsSafe);
    const action = update[result.source.index];

    update.splice(result.source.index, 1);
    update.splice(result.destination.index, 0, action);

    setActionsSafe(update);
    onChange(update);
  };

  return (
    <div className={styles.container}>
      <DragDropContext onDragEnd={onDragEnd}>
        <Droppable droppableId="sortable-actions" direction="vertical">
          {(provided) => (
            <div className={styles.wrapper} ref={provided.innerRef} {...provided.droppableProps}>
              {actionsSafe.map((action, idx) => {
                const key = `${action.title}/${idx}`;

                return (
                  <ActionListItem
                    key={key}
                    index={idx}
                    action={action}
                    onChange={onActionChange}
                    onEdit={() => setEditIndex(idx)}
                    onRemove={() => onActionRemove(idx)}
                    data={data}
                    itemKey={key}
                  />
                );
              })}
              {provided.placeholder}
            </div>
          )}
        </Droppable>
      </DragDropContext>

      {isEditing && editIndex !== null && (
        <Modal
          title="Edit action"
          isOpen={true}
          closeOnBackdropClick={false}
          onDismiss={() => {
            onActionCancel(editIndex);
          }}
        >
          <ActionEditorModalContent
            index={editIndex}
            action={isNew ? defaultActionConfig : actionsSafe[editIndex]}
            data={data}
            onSave={onActionChange}
            onCancel={onActionCancel}
            getSuggestions={getSuggestions}
            showOneClick={!!dataLinks}
          />
        </Modal>
      )}

      <Button size="sm" icon="plus" onClick={onActionAdd} variant="secondary" className={styles.button}>
        <Trans i18nKey="actions-editor.inline.add-button">Add action</Trans>
      </Button>
    </div>
  );
};

const getActionsInlineEditorStyle = (theme: GrafanaTheme2) => ({
  container: css({
    position: 'relative',
  }),
  wrapper: css({
    marginBottom: theme.spacing(2),
    display: 'flex',
    flexDirection: 'column',
  }),
  button: css({
    marginLeft: theme.spacing(1),
  }),
});
