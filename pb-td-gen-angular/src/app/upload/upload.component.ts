import {Component, OnInit} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {DataStorageService} from "../shared/data-storage-service";
import {OverviewService} from "../overview/overview.service";
import {Affordances} from "../shared/affordance-structure";
import {Router} from "@angular/router";

@Component({
  selector: 'app-upload',
  templateUrl: './upload.component.html',
  styleUrls: ['./upload.component.css']
})
export class UploadComponent implements OnInit {
  fileName = '';

  constructor(private http: HttpClient,
              private dataStorageService: DataStorageService,
              private overviewService: OverviewService,
              private router: Router) {
  }

  ngOnInit(): void {
  }

  onFileSelected(event: any) {

    const file: File = event.target.files[0];

    if (file) {
      this.fileName = file.name;

      this.dataStorageService.fetchProtoData(file)
        .subscribe(value => {
          //var test : Affordances = JSON.parse(value)
          console.log(value);
          this.overviewService.setAffordances(value);
          this.router.navigate(["/overview"]);
        });
    }
  }
}
