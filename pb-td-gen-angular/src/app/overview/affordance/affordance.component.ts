import {Component, Input, OnInit} from '@angular/core';
import {Affs} from "../../shared/affordance-structure";
import {AffordanceType, OverviewService} from "../overview.service";

@Component({
  selector: 'app-affordance',
  templateUrl: './affordance.component.html',
  styleUrls: ['./affordance.component.css']
})
export class AffordanceComponent implements OnInit {
  @Input()
  affordance!: {id: number, affordance: Affs, type: AffordanceType};

  constructor(private overviewService: OverviewService) {
  }

  ngOnInit(): void {
  }

  selectAffordance() {
    this.overviewService.changeActualAffordance(this.affordance);
  }
}
