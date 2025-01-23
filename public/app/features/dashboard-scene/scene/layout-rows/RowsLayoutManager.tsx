import { css } from '@emotion/css';
import { useEffect, useState } from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import {
  LocalValueVariable,
  MultiValueVariable,
  SceneComponentProps,
  sceneGraph,
  SceneGridItemLike,
  SceneGridRow,
  SceneObjectBase,
  SceneObjectState,
  SceneVariableSet,
  VariableValueSingle,
  VizPanel,
} from '@grafana/scenes';
import { useStyles2 } from '@grafana/ui';

import { getMultiVariableValues } from '../../utils/utils';
import { DashboardScene } from '../DashboardScene';
import { DashboardGridItem } from '../layout-default/DashboardGridItem';
import { DefaultGridLayoutManager } from '../layout-default/DefaultGridLayoutManager';
import { RowRepeaterBehavior } from '../layout-default/RowRepeaterBehavior';
import { ResponsiveGridLayoutManager } from '../layout-responsive-grid/ResponsiveGridLayoutManager';
import { getRepeatKeyForSceneObject, isRepeatedSceneObject } from '../layouts-shared/repeatUtils';
import { DashboardLayoutManager, LayoutRegistryItem } from '../types';

import { RowItem } from './RowItem';
import { RowItemRepeaterBehavior } from './RowItemRepeaterBehavior';

interface RowsLayoutManagerState extends SceneObjectState {
  rows: RowItem[];
}

export class RowsLayoutManager extends SceneObjectBase<RowsLayoutManagerState> implements DashboardLayoutManager {
  public isDashboardLayoutManager: true = true;

  public editModeChanged(isEditing: boolean): void {}

  public addPanel(vizPanel: VizPanel): void {
    // Try to add new panels to the selected row
    const selectedObject = this.getSelectedObject();
    if (selectedObject instanceof RowItem) {
      return selectedObject.onAddPanel(vizPanel);
    }

    // If we don't have selected row add it to the first row
    if (this.state.rows.length > 0) {
      return this.state.rows[0].onAddPanel(vizPanel);
    }

    // Otherwise fallback to adding a new row and a panel
    this.addNewRow();
    this.state.rows[this.state.rows.length - 1].onAddPanel(vizPanel);
  }

  public addNewRow(): void {
    this.setState({
      rows: [
        ...this.state.rows,
        new RowItem({
          title: 'New row',
          layout: ResponsiveGridLayoutManager.createEmpty(),
        }),
      ],
    });
  }

  public getNextPanelId(): number {
    return 0;
  }

  public removePanel(panel: VizPanel) {}

  public removeRow(row: RowItem) {
    this.setState({
      rows: this.state.rows.filter((r) => r !== row),
    });
  }

  public duplicatePanel(panel: VizPanel): void {
    throw new Error('Method not implemented.');
  }

  public getVizPanels(): VizPanel[] {
    const panels: VizPanel[] = [];

    for (const row of this.state.rows) {
      const innerPanels = row.state.layout.getVizPanels();
      panels.push(...innerPanels);
    }

    return panels;
  }

  public getOptions() {
    return [];
  }

  public activateRepeaters() {
    this.state.rows.forEach((row) => {
      if (row.state.$behaviors) {
        for (const behavior of row.state.$behaviors) {
          if (behavior instanceof RowItemRepeaterBehavior && !row.isActive) {
            row.activate();
            break;
          }
        }

        if (!row.getLayout().isActive) {
          row.getLayout().activate();
        }
      }
    });
  }

  public getDescriptor(): LayoutRegistryItem {
    return RowsLayoutManager.getDescriptor();
  }

  public getSelectedObject() {
    return sceneGraph.getAncestor(this, DashboardScene).state.editPane.state.selectedObject?.resolve();
  }

  public static getDescriptor(): LayoutRegistryItem {
    return {
      name: 'Rows',
      description: 'Rows layout',
      id: 'rows-layout',
      createFromLayout: RowsLayoutManager.createFromLayout,
    };
  }

  public static createEmpty() {
    return new RowsLayoutManager({ rows: [] });
  }

  public static createFromLayout(layout: DashboardLayoutManager): RowsLayoutManager {
    let rows: RowItem[];

    if (layout instanceof DefaultGridLayoutManager) {
      const config: Array<{
        title?: string;
        isCollapsed?: boolean;
        isDraggable?: boolean;
        isResizable?: boolean;
        children: SceneGridItemLike[];
        repeat?: string;
      }> = [];
      let children: SceneGridItemLike[] | undefined;

      layout.state.grid.forEachChild((child) => {
        if (!(child instanceof DashboardGridItem) && !(child instanceof SceneGridRow)) {
          throw new Error('Child is not a DashboardGridItem or SceneGridRow, invalid scene');
        }

        if (child instanceof SceneGridRow) {
          if (!isRepeatedSceneObject(child)) {
            const behaviour = child.state.$behaviors?.find((b) => b instanceof RowRepeaterBehavior);

            config.push({
              title: child.state.title,
              isCollapsed: !!child.state.isCollapsed,
              isDraggable: child.state.isDraggable,
              isResizable: child.state.isResizable,
              children: child.state.children,
              repeat: behaviour?.state.variableName,
            });

            // Since we encountered a row item, any subsequent panels should be added to a new row
            children = undefined;
          }
        } else {
          if (!children) {
            children = [];
            config.push({ children });
          }

          children.push(child);
        }
      });

      rows = config.map(
        (rowConfig) =>
          new RowItem({
            title: rowConfig.title ?? 'Row title',
            isCollapsed: !!rowConfig.isCollapsed,
            repeatByVariable: rowConfig.repeat,
            layout: DefaultGridLayoutManager.fromGridItems(
              rowConfig.children,
              rowConfig.isDraggable,
              rowConfig.isResizable
            ),
          })
      );
    } else {
      rows = [new RowItem({ layout: layout.clone(), title: 'Row title' })];
    }

    return new RowsLayoutManager({ rows });
  }

  public static Component = ({ model }: SceneComponentProps<RowsLayoutManager>) => {
    const { rows } = model.useState();
    const styles = useStyles2(getStyles);

    return (
      <div className={styles.wrapper}>
        {rows.map((row) => (
          <RenderRow row={row} key={row.state.key} />
        ))}
      </div>
    );
  };
}

function RenderRow({ row }: { row: RowItem }) {
  const { repeatByVariable } = row.useState();
  if (!repeatByVariable) {
    return <RowItem.Component model={row} key={row.state.key!} />;
  }

  const variable = sceneGraph.lookupVariable(repeatByVariable, row.parent!);

  if (!(variable instanceof MultiValueVariable)) {
    return <RowItem.Component model={row} key={row.state.key!} />;
  }

  const { values, texts } = getMultiVariableValues(variable);

  variable.useState();

  return (
    <>
      {values.map((value, index) => (
        <RenderRowClone row={row} variable={variable} value={value} text={texts[index]} key={index} />
      ))}
    </>
  );
}

function RenderRowClone({
  row,
  variable,
  value,
  text,
}: {
  row: RowItem;
  variable: MultiValueVariable;
  value: VariableValueSingle;
  text: VariableValueSingle;
}) {
  const rowState = row.useState();
  const cloneKey = getRepeatKeyForSceneObject(row, value);

  useEffect(() => {
    const $variables = new SceneVariableSet({
      variables: [
        new LocalValueVariable({
          name: variable.state.name,
          value,
          text: String(text),
          isMulti: variable.state.isMulti,
          includeAll: variable.state.includeAll,
        }),
      ],
    });

    const clone = row.clone({ key: cloneKey, $variables });

    row.setState({ repeats: [clone, ...(row.state.repeats ?? [])] });

    return () => {
      row.setState({ repeats: row.state.repeats?.filter((r) => r.state.key !== cloneKey) });
    };
  }, [row, value, text, variable, rowState, cloneKey]);

  const clone = rowState.repeats?.find((r) => r.state.key === cloneKey);

  if (!clone) {
    return null;
  }

  return <RowItem.Component model={clone} key={clone.state.key!} />;
}

function getStyles(theme: GrafanaTheme2) {
  return {
    wrapper: css({
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(1),
      flexGrow: 1,
      width: '100%',
    }),
  };
}
