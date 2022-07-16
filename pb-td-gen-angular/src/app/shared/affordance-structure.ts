export interface Affordances {
  Props: Property[],
  Actions: Affs[],
  Events: Affs[],
}

export interface Property {
  Name: string
  GetProp: Affs
  SetProp: Affs
  Category: number
}

export interface Affs {
  Name: string
  Req: DataSchema
  Res: DataSchema
}

export interface DataSchema {
  Type: string
  Properties: Properties[] | null
}

export interface Properties {
  Key: string
  Value: DataSchema
}
