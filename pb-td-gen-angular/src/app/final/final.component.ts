import { Component, OnInit } from '@angular/core';
import {OverviewService} from "../overview/overview.service";

@Component({
  selector: 'app-final',
  templateUrl: './final.component.html',
  styleUrls: ['./final.component.css']
})
export class FinalComponent implements OnInit {

  constructor(private overviewService: OverviewService) { }

  ngOnInit(): void {
  }

  getTD() {
    return this.overviewService.tdString;
  }
}
