export interface ThingDescription {
  title: string,
  description?: string,
  securityDefinitions?: SecurityDefinitions,
  security: string,
  properties?: Map<string, PropertyAffordance>,
  actions?: Map<string, ActionAffordance>,
  events?: Map<string, EventAffordance>,
}

export interface PropertyAffordance extends DataSchema {
  title?: string,
  description?: string,
}

export interface ActionAffordance {
  title?: string,
  description?: string,
  input?: DataSchema,
  output?: DataSchema,
}

export interface EventAffordance {
  title?: string,
  description?: string,
  data?: DataSchema,
}

export interface SecurityDefinitions {

}

export interface DataSchema {
  type: string,
  properties?: Map<string, DataSchema>,
}
