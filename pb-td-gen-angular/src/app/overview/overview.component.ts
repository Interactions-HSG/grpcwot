import {Component, OnInit} from '@angular/core';
import {Affordances, Affs, Property} from "../shared/affordance-structure";
import {OverviewService, AffordanceType} from "./overview.service";

@Component({
  selector: 'app-overview',
  templateUrl: './overview.component.html',
  styleUrls: ['./overview.component.css']
})
export class OverviewComponent implements OnInit {
  properties: Property[];
  actions: Affs[];
  events: Affs[];
  affordanceType: typeof AffordanceType = AffordanceType;

  constructor(private overviewService: OverviewService) {
    this.properties = overviewService.getProperties();
    this.actions = overviewService.getActions();
    this.events = overviewService.getEvents();
  }

  ngOnInit(): void {
  }

}
