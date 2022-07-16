import {Component, NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UploadComponent} from "./upload/upload.component";
import {OverviewComponent} from "./overview/overview.component";
import {FinalComponent} from "./final/final.component";

const appRoutes: Routes = [
  {path: '', redirectTo: '/upload', pathMatch: 'full'},
  {path: 'upload', component: UploadComponent},
  {path: 'overview', component: OverviewComponent},
  {path: 'final', component: FinalComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(appRoutes)],
  exports: [RouterModule]
})
export class AppRoutingModule {
}
