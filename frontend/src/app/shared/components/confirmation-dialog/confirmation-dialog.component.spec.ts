import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ConfirmationDialogComponent } from "./confirmation-dialog.component";
import {
  BrnAlertDialogContentDirective,
  BrnAlertDialogTriggerDirective,
} from "@spartan-ng/brain/alert-dialog";
import {
  HlmAlertDialogActionButtonDirective,
  HlmAlertDialogCancelButtonDirective,
  HlmAlertDialogComponent,
  HlmAlertDialogContentComponent,
  HlmAlertDialogDescriptionDirective,
  HlmAlertDialogFooterComponent,
  HlmAlertDialogHeaderComponent,
  HlmAlertDialogOverlayDirective,
  HlmAlertDialogTitleDirective,
} from "@spartan-ng/ui-alertdialog-helm";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { By } from "@angular/platform-browser";

describe("ConfirmationDialogComponent", () => {
  let component: ConfirmationDialogComponent;
  let fixture: ComponentFixture<ConfirmationDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        ConfirmationDialogComponent,
        BrnAlertDialogTriggerDirective,
        BrnAlertDialogContentDirective,
        HlmAlertDialogComponent,
        HlmAlertDialogContentComponent,
        HlmAlertDialogHeaderComponent,
        HlmAlertDialogFooterComponent,
        HlmAlertDialogTitleDirective,
        HlmAlertDialogDescriptionDirective,
        HlmAlertDialogCancelButtonDirective,
        HlmAlertDialogActionButtonDirective,
        HlmAlertDialogOverlayDirective,
        HlmButtonDirective,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmationDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should display the trigger text", () => {
    component.triggerText = "Test Trigger";
    fixture.detectChanges();

    const buttonElement = fixture.debugElement.query(
      By.css("button")
    ).nativeElement;
    expect(buttonElement.textContent.trim()).toBe("Test Trigger");
  });

  it("should emit confirmed event when confirm button is clicked", () => {
    spyOn(component.confirmed, "emit");
    const closeSpy = jasmine.createSpy("close");

    component.onConfirm(closeSpy);

    expect(component.confirmed.emit).toHaveBeenCalled();
    expect(closeSpy).toHaveBeenCalled();
  });

  it("should emit cancelled event when cancel button is clicked", () => {
    spyOn(component.cancelled, "emit");
    const closeSpy = jasmine.createSpy("close");

    component.onCancel(closeSpy);

    expect(component.cancelled.emit).toHaveBeenCalled();
    expect(closeSpy).toHaveBeenCalled();
  });

  it("should disable the trigger button when disabled is true", () => {
    component.disabled = true;
    fixture.detectChanges();

    const buttonElement = fixture.debugElement.query(
      By.css("button")
    ).nativeElement;
    expect(buttonElement.disabled).toBeTrue();
  });
});
