import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AffordanceDetailComponent } from './affordance-detail.component';

describe('AffordanceDetailComponent', () => {
  let component: AffordanceDetailComponent;
  let fixture: ComponentFixture<AffordanceDetailComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AffordanceDetailComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AffordanceDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
