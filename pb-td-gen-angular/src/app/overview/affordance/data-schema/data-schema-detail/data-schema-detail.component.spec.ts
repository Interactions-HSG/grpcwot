import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DataSchemaDetailComponent } from './data-schema-detail.component';

describe('DataSchemaDetailComponent', () => {
  let component: DataSchemaDetailComponent;
  let fixture: ComponentFixture<DataSchemaDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ DataSchemaDetailComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DataSchemaDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
