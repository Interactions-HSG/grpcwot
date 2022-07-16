import {Component, Input, OnInit} from '@angular/core';
import {DataSchema} from "../../../shared/affordance-structure";

@Component({
  selector: 'app-data-schema',
  templateUrl: './data-schema.component.html',
  styleUrls: ['./data-schema.component.css']
})
export class DataSchemaComponent implements OnInit {
  @Input()
  ds!: DataSchema

  constructor() { }

  ngOnInit(): void {
  }

}
