import { PanelPlugin } from '@grafana/data';
import { commonOptionsBuilder } from '@grafana/ui';
import { optsWithHideZeros } from '@grafana/ui/src/options/builder/tooltip';

import { TimeSeriesPanel } from './TimeSeriesPanel';
import { TimezonesEditor } from './TimezonesEditor';
import { defaultTimeseriesGraphConfig, getGraphFieldConfig } from './config';
import { graphPanelChangedHandler } from './migrations';
import { TimeSeriesFieldConfig, Options } from './panelcfg.gen';
import { TimeSeriesSuggestionsSupplier } from './suggestions';

export const plugin = new PanelPlugin<Options, TimeSeriesFieldConfig>(TimeSeriesPanel)
  .setPanelChangeHandler(graphPanelChangedHandler)
  .useFieldConfig(getGraphFieldConfig(defaultTimeseriesGraphConfig))
  .setPanelOptions((builder) => {
    commonOptionsBuilder.addTooltipOptions(builder, false, true, optsWithHideZeros);
    commonOptionsBuilder.addLegendOptions(builder);

    builder.addCustomEditor({
      id: 'timezone',
      name: 'Time zone',
      path: 'timezone',
      category: ['Axis'],
      editor: TimezonesEditor,
      defaultValue: undefined,
    });
  })
  .setSuggestionsSupplier(new TimeSeriesSuggestionsSupplier())
  .setDataSupport({ annotations: true, alertStates: true });
