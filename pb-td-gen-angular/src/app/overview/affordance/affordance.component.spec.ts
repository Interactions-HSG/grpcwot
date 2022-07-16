import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AffordanceComponent } from './affordance.component';

describe('AffordanceComponent', () => {
  let component: AffordanceComponent;
  let fixture: ComponentFixture<AffordanceComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AffordanceComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AffordanceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
