import {Component, OnDestroy, OnInit} from '@angular/core';
import {AffordanceType, OverviewService} from "../overview.service";
import {Subscription} from "rxjs";
import {Affs, Properties, Property} from "../../shared/affordance-structure";
import {FormControl, FormGroup} from "@angular/forms";

@Component({
  selector: 'app-affordance-detail',
  templateUrl: './affordance-detail.component.html',
  styleUrls: ['./affordance-detail.component.css']
})
export class AffordanceDetailComponent implements OnInit, OnDestroy {
  private subscription: Subscription | undefined;
  actAffordance: { id: number; affordance: Affs; type: AffordanceType } | null;
  affordanceType: typeof AffordanceType = AffordanceType;
  selectProperty: boolean = false;
  radioSelected: any = 'get';

  constructor(private overviewService: OverviewService) {
    this.actAffordance = null;
  }

  ngOnInit(): void {
    this.subscription = this.overviewService.actAffordanceChanged
      .subscribe((affordance: { id: number; affordance: Affs; type: AffordanceType }) => {
        this.actAffordance = affordance;
        this.selectProperty = false;
      });
  }

  ngOnDestroy(): void {
    if (this.subscription != undefined)
      this.subscription.unsubscribe();
  }

  getProperties(): Property[] {
    return this.overviewService.getProperties();
  }

  moveToAction() {
    this.overviewService.moveActAffordanceToAction();
  }

  getActAffordance(): Affs {
    return this.actAffordance!.affordance
  }

  moveToEvent() {
    this.overviewService.moveActAffordanceToEvent();
  }

  moveToNewProperty() {
    this.selectProperty = false;
    this.overviewService.moveActAffordanceToProperty(this.radioSelected === 'get');
  }

  moveToExistingProperty(property: { property: Property; id: number }) {
    this.selectProperty = false;
    this.overviewService.addActAffordanceToExistingProperty(this.radioSelected === 'get', property);
  }
}
