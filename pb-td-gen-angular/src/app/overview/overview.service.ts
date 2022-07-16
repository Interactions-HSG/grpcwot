import {Injectable} from "@angular/core";
import {DataStorageService} from "../shared/data-storage-service";
import {Affordances, Affs, DataSchema, Property} from "../shared/affordance-structure";
import {Subject} from "rxjs";
import {
  ActionAffordance,
  ThingDescription,
  DataSchema as WoTDataSchema, EventAffordance, PropertyAffordance
} from "../shared/thing-description-structure";
import {Router} from "@angular/router";

@Injectable({providedIn: "root"})
export class OverviewService {
  private affordances: Affordances | undefined
  actAffordanceChanged = new Subject<{ id: number; affordance: Affs; type: AffordanceType }>()
  private actAffordance: { id: number; affordance: Affs; type: AffordanceType } | null
  private td: ThingDescription | undefined;
  tdString: string = '';

  constructor(private dataStorageService: DataStorageService,
              private router: Router) {
    this.affordances = {
      "Props": [
        {
          "Name": "Mode",
          "GetProp": {
            "Name": "GetMode",
            "Req": {
              "Type": "object",
              "Properties": null
            },
            "Res": {
              "Type": "object",
              "Properties": [
                {
                  "Key": "status_code",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                },
                {
                  "Key": "mode",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                }
              ]
            }
          },
          "SetProp": {
            "Name": "SetMode",
            "Req": {
              "Type": "object",
              "Properties": [
                {
                  "Key": "status_code",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                },
                {
                  "Key": "mode",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                }
              ]
            },
            "Res": {
              "Type": "object",
              "Properties": [
                {
                  "Key": "mode",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                },
                {
                  "Key": "status_code",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                }
              ]
            }
          },
          "Category": 2
        },
        {
          "Name": "Position",
          "GetProp": {
            "Name": "GetPosition",
            "Req": {
              "Type": "object",
              "Properties": null
            },
            "Res": {
              "Type": "object",
              "Properties": [
                {
                  "Key": "status_code",
                  "Value": {
                    "Type": "integer",
                    "Properties": null
                  }
                },
                {
                  "Key": "x",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                },
                {
                  "Key": "y",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                },
                {
                  "Key": "z",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                },
                {
                  "Key": "roll",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                },
                {
                  "Key": "pitch",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                },
                {
                  "Key": "yaw",
                  "Value": {
                    "Type": "number",
                    "Properties": null
                  }
                }
              ]
            }
          },
          "SetProp": {
            "Name": "",
            "Req": {
              "Type": "",
              "Properties": null
            },
            "Res": {
              "Type": "",
              "Properties": null
            }
          },
          "Category": 0
        }
      ],
      "Actions": [
        {
          "Name": "SetPosition",
          "Req": {
            "Type": "object",
            "Properties": [
              {
                "Key": "timeout",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "status_code",
                "Value": {
                  "Type": "integer",
                  "Properties": null
                }
              },
              {
                "Key": "pose",
                "Value": {
                  "Type": "object",
                  "Properties": [
                    {
                      "Key": "status_code",
                      "Value": {
                        "Type": "integer",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "x",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "y",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "z",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "roll",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "pitch",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "yaw",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    }
                  ]
                }
              },
              {
                "Key": "radius",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "speed",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "acc",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "mvtime",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "wait",
                "Value": {
                  "Type": "boolean",
                  "Properties": null
                }
              }
            ]
          },
          "Res": {
            "Type": "object",
            "Properties": [
              {
                "Key": "pose",
                "Value": {
                  "Type": "object",
                  "Properties": [
                    {
                      "Key": "x",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "y",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "z",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "roll",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "pitch",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "yaw",
                      "Value": {
                        "Type": "number",
                        "Properties": null
                      }
                    },
                    {
                      "Key": "status_code",
                      "Value": {
                        "Type": "integer",
                        "Properties": null
                      }
                    }
                  ]
                }
              },
              {
                "Key": "radius",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "speed",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "acc",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "mvtime",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "wait",
                "Value": {
                  "Type": "boolean",
                  "Properties": null
                }
              },
              {
                "Key": "timeout",
                "Value": {
                  "Type": "number",
                  "Properties": null
                }
              },
              {
                "Key": "status_code",
                "Value": {
                  "Type": "integer",
                  "Properties": null
                }
              }
            ]
          }
        }
      ],
      "Events": []
    }
    this.actAffordance = null
  }

  changeActualAffordance(affordance: { id: number; affordance: Affs; type: AffordanceType }): void {
    this.actAffordance = affordance;
    this.actAffordanceChanged.next(this.actAffordance)
  }

  setAffordances(affordances: Affordances): void {
    this.affordances = affordances;
  }

  getProperties(): Property[] {
    if (this.affordances != undefined) {
      return this.affordances.Props;
    } else {
      return [];
    }
  }

  getActions(): Affs[] {
    if (this.affordances != undefined) {
      return this.affordances.Actions;
    } else {
      return [];
    }
  }

  getEvents(): Affs[] {
    if (this.affordances != undefined) {
      return this.affordances.Events;
    } else {
      return [];
    }
  }

  moveActAffordanceToProperty(isGet: boolean) {
    let elem = this.getActualAffAndRemoveInPriorClass();
    let i = this.getProperties().push(
      {
        Name: elem.Name,
        GetProp: (isGet ? elem : this.emptyAff),
        SetProp: (isGet ? this.emptyAff : elem),
        Category: (isGet ? 0 : 1),
      }
    );
    this.actAffordance = {
      id: i - 1,
      affordance: elem,
      type: (isGet ? AffordanceType.PROPERTY_GET : AffordanceType.PROPERTY_SET)
    };
    this.actAffordanceChanged.next(this.actAffordance);
  }

  addActAffordanceToExistingProperty(isGet: boolean, property: { property: Property; id: number }): void {
    let elem = this.getActualAffAndRemoveInPriorClass()
    let inject = this.getProperties()[property.id]
    if (isGet)
      inject.GetProp = elem;
    else
      inject.SetProp = elem;
    inject.Category = 2;
    this.actAffordance = {
      id: property.id,
      affordance: elem,
      type: (isGet ? AffordanceType.PROPERTY_GET : AffordanceType.PROPERTY_SET)
    };
    this.actAffordanceChanged.next(this.actAffordance);
  }

  moveActAffordanceToAction() {
    let elem = this.getActualAffAndRemoveInPriorClass()
    let i = this.getActions().push(elem);
    this.actAffordance = {id: i - 1, affordance: elem, type: AffordanceType.ACTION};
    this.actAffordanceChanged.next(this.actAffordance);
  }

  moveActAffordanceToEvent() {
    let elem = this.getActualAffAndRemoveInPriorClass()
    let i = this.getEvents().push(elem);
    this.actAffordance = {id: i - 1, affordance: elem, type: AffordanceType.EVENT};
    this.actAffordanceChanged.next(this.actAffordance);
  }

  private emptyAff: Affs = {
    Name: '',
    Req: {
      Type: '',
      Properties: null,
    },
    Res: {
      Type: '',
      Properties: null,
    },
  }

  private getActualAffAndRemoveInPriorClass(): Affs {
    let elem: Affs;
    switch (this.actAffordance!.type) {
      case AffordanceType.ACTION:
        elem = this.getActions()[this.actAffordance!.id];
        this.getActions().splice(this.actAffordance!.id, 1);
        break;
      case AffordanceType.PROPERTY_GET:
        elem = this.affordances!.Props[this.actAffordance!.id].GetProp;
        if (this.getProperties()[this.actAffordance!.id].Category == 2) {
          this.getProperties()[this.actAffordance!.id].GetProp = this.emptyAff;
          this.getProperties()[this.actAffordance!.id].Category = 1;
        } else
          this.getProperties().splice(this.actAffordance!.id, 1);
        break;
      case AffordanceType.PROPERTY_SET:
        elem = this.getProperties()[this.actAffordance!.id].SetProp;
        if (this.getProperties()[this.actAffordance!.id].Category == 2) {
          this.getProperties()[this.actAffordance!.id].SetProp = this.emptyAff;
          this.getProperties()[this.actAffordance!.id].Category = 0;
        } else
          this.getProperties().splice(this.actAffordance!.id, 1);
        break;
      case AffordanceType.EVENT:
        elem = this.getEvents()[this.actAffordance!.id];
        this.getEvents().splice(this.actAffordance!.id, 1);
    }
    return elem;
  }

  produceTD(title: string): void {
    this.td = {
      title: title,
      security: "no-sec",
      properties: this.produceProperties(),
      actions: this.produceActions(),
      events: this.produceEvents(),
    }
    this.router.navigate(["/final"]);
    this.tdString = JSON.stringify(this.td, this.replacer);
  }

  replacer(key: any, value: any) {
    if (value instanceof Map) {
      return Object.fromEntries(value);
    } else {
      return value;
    }
  }

  produceActions(): Map<string, ActionAffordance> {
    const res = new Map<string, ActionAffordance>();
    for (let action of this.affordances!.Actions) {
      res.set(action.Name, {
        input: this.produceWotDataSchema(action.Req),
        output: this.produceWotDataSchema(action.Res),
      });
    }
    return res;
  }

  produceEvents(): Map<string, EventAffordance> {
    const res = new Map<string, EventAffordance>();
    for (let event of this.affordances!.Events) {
      res.set(event.Name, {
        data: this.produceWotDataSchema(event.Res),
      });
    }
    return res;
  }

  produceProperties(): Map<string, PropertyAffordance> {
    const res = new Map<string, PropertyAffordance>();
    for (let prop of this.affordances!.Props) {
      let ds = this.produceWotDataSchema(prop.Category === 1 ? prop.SetProp.Req : prop.GetProp.Res);
      res.set(prop.Name, {
        type: ds.type,
        properties: ds.properties,
      });
    }
    return res;
  }

  produceWotDataSchema(ds: DataSchema): WoTDataSchema {
    const m = new Map<string, WoTDataSchema>();
    if (ds.Properties !== null && ds.Properties.length !== 0) {
      for (let p of ds.Properties) {
        m.set(p.Key, this.produceWotDataSchema(p.Value))
      }
      return {
        type: ds.Type,
        properties: m,
      }
    } else {
      return {
        type: ds.Type
      }
    }
  }
}

export enum AffordanceType {
  PROPERTY_GET,
  PROPERTY_SET,
  ACTION,
  EVENT,
}
