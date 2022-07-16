import { Component, OnInit } from '@angular/core';
import {OverviewService} from "../overview/overview.service";
import {NgForm} from "@angular/forms";

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit {
  fileName = "tsconfig.json";
  sourcePath = ""

  constructor(private overviewService: OverviewService) {
  }

  ngOnInit(): void {
  }

  onSubmit(tdForm: NgForm) {
    if (tdForm.invalid) {
      return;
    }
    const title = tdForm.value.title;
    this.overviewService.produceTD(title);
    tdForm.reset();
  }
}
