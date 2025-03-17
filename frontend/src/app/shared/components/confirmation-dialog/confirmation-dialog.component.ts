import { Component, EventEmitter, Input, Output } from "@angular/core";
import { CommonModule } from "@angular/common";
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

@Component({
  selector: "app-confirmation-dialog",
  standalone: true,
  imports: [
    CommonModule,
    BrnAlertDialogTriggerDirective,
    BrnAlertDialogContentDirective,
    HlmAlertDialogComponent,
    HlmAlertDialogContentComponent,
    HlmAlertDialogHeaderComponent,
    HlmAlertDialogFooterComponent,
    HlmAlertDialogTitleDirective,
    HlmAlertDialogDescriptionDirective,
    HlmAlertDialogCancelButtonDirective,
    HlmButtonDirective,
  ],
  templateUrl: "./confirmation-dialog.component.html",
})
export class ConfirmationDialogComponent {
  /**
   * Text for the trigger button
   */
  @Input() triggerText = "Open Dialog";

  /**
   * CSS variant for the trigger button (outline, ghost, etc.)
   */
  @Input() triggerVariant:
    | "default"
    | "destructive"
    | "outline"
    | "secondary"
    | "ghost"
    | "link"
    | null
    | undefined = "outline";

  /**
   * CSS variant for the confirm button
   */
  @Input() confirmVariant:
    | "default"
    | "destructive"
    | "outline"
    | "secondary"
    | "ghost"
    | "link"
    | null
    | undefined = "default";

  /**
   * Title of the confirmation dialog
   */
  @Input() title = "Confirmation";

  /**
   * Description text for the confirmation dialog
   */
  @Input() description = "Are you sure you want to perform this action?";

  /**
   * Text for the cancel button
   */
  @Input() cancelText = "Cancel";

  /**
   * Text for the confirm/action button
   */
  @Input() confirmText = "Confirm";

  /**
   * CSS class for the trigger button
   */
  @Input() triggerClass = "";

  /**
   * Whether the dialog is disabled
   */
  @Input() disabled = false;

  /**
   * Event emitted when the confirm button is clicked
   */
  @Output() confirmed = new EventEmitter<void>();

  /**
   * Event emitted when the cancel button is clicked
   */
  @Output() cancelled = new EventEmitter<void>();

  /**
   * Handle the confirm action
   */
  onConfirm(close: () => void): void {
    this.confirmed.emit();
    close();
  }

  /**
   * Handle the cancel action
   */
  onCancel(close: () => void): void {
    this.cancelled.emit();
    close();
  }
}
