// Code generated - EDITING IS FUTILE. DO NOT EDIT.
//
// Generated by:
//     public/app/plugins/gen.go
// Using jennies:
//     TSTypesJenny
//     PluginTsTypesJenny
//
// Run 'make gen-cue' from repository root to regenerate.

import * as common from '@grafana/schema';

export const pluginVersion = "11.6.0-pre";

export enum PointShape {
  Circle = 'circle',
  Square = 'square',
}

export enum SeriesMapping {
  Auto = 'auto',
  Manual = 'manual',
}

export enum XYShowMode {
  Lines = 'lines',
  Points = 'points',
  PointsAndLines = 'points+lines',
}

/**
 * NOTE: (copied from dashboard_kind.cue, since not exported)
 * Matcher is a predicate configuration. Based on the config a set of field(s) or values is filtered in order to apply override / transformation.
 * It comes with in id ( to resolve implementation from registry) and a configuration that’s specific to a particular matcher type.
 */
export interface MatcherConfig {
  /**
   * The matcher id. This is used to find the matcher implementation from registry.
   */
  id: string;
  /**
   * The matcher options. This is specific to the matcher implementation.
   */
  options?: unknown;
}

export const defaultMatcherConfig: Partial<MatcherConfig> = {
  id: '',
};

export interface FieldConfig extends common.HideableFieldConfig, common.AxisConfig {
  fillOpacity?: number;
  lineStyle?: common.LineStyle;
  lineWidth?: number;
  pointShape?: PointShape;
  pointSize?: {
    fixed?: number;
    min?: number;
    max?: number;
  };
  pointStrokeWidth?: number;
  show?: XYShowMode;
}

export const defaultFieldConfig: Partial<FieldConfig> = {
  fillOpacity: 50,
  show: XYShowMode.Points,
};

export interface XYSeriesConfig {
  color?: {
    matcher: MatcherConfig;
  };
  frame?: {
    matcher: MatcherConfig;
  };
  name?: {
    fixed?: string;
  };
  size?: {
    matcher: MatcherConfig;
  };
  x?: {
    matcher: MatcherConfig;
  };
  y?: {
    matcher: MatcherConfig;
  };
}

export interface Options extends common.OptionsWithLegend, common.OptionsWithTooltip {
  mapping: SeriesMapping;
  series: Array<XYSeriesConfig>;
}

export const defaultOptions: Partial<Options> = {
  series: [],
};
