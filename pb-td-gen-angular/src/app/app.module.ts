import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';

import {AppComponent} from './app.component';
import {OverviewComponent} from './overview/overview.component';
import {HttpClientModule} from "@angular/common/http";
import {UploadComponent} from './upload/upload.component';
import {AppRoutingModule} from "./app-routing.module";
import {PropertiesComponent} from './overview/properties/properties.component';
import {AffordanceComponent} from './overview/affordance/affordance.component';
import {DataSchemaComponent} from './overview/affordance/data-schema/data-schema.component';
import {DataSchemaDetailComponent} from './overview/affordance/data-schema/data-schema-detail/data-schema-detail.component';
import {AffordanceDetailComponent} from './overview/affordance-detail/affordance-detail.component';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {HeaderComponent} from './header/header.component';
import {FinalComponent} from './final/final.component';

@NgModule({
  declarations: [
    AppComponent,
    OverviewComponent,
    UploadComponent,
    PropertiesComponent,
    AffordanceComponent,
    DataSchemaComponent,
    DataSchemaDetailComponent,
    AffordanceDetailComponent,
    HeaderComponent,
    FinalComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    ReactiveFormsModule,
    AppRoutingModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule {
}
