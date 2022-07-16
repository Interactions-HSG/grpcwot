import {Component, Input, OnInit} from '@angular/core';
import {Properties} from "../../../../shared/affordance-structure";

@Component({
  selector: 'app-data-schema-detail',
  templateUrl: './data-schema-detail.component.html',
  styleUrls: ['./data-schema-detail.component.css']
})
export class DataSchemaDetailComponent implements OnInit {
  @Input()
  property!: Properties

  constructor() { }

  ngOnInit(): void {
  }

}
