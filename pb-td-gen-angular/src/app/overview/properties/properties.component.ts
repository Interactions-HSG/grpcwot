import {Component, Input, OnInit} from '@angular/core';
import {Property} from "../../shared/affordance-structure";
import {AffordanceType} from "../overview.service";

@Component({
  selector: 'app-properties',
  templateUrl: './properties.component.html',
  styleUrls: ['./properties.component.css']
})
export class PropertiesComponent implements OnInit {
  @Input()
  property!: {id: number, property: Property};
  affordanceType: typeof AffordanceType = AffordanceType;

  constructor() {
  }

  ngOnInit(): void {
  }

  getProperty(): Property {
    return this.property.property;
  }
}
