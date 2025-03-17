import {
  ChangeDetectionStrategy,
  Component,
  EventEmitter,
  Output,
  inject,
  signal,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormBuilder, ReactiveFormsModule, Validators } from "@angular/forms";
import {
  BrnDialogContentDirective,
  BrnDialogTriggerDirective,
} from "@spartan-ng/brain/dialog";
import {
  HlmDialogComponent,
  HlmDialogContentComponent,
  HlmDialogDescriptionDirective,
  HlmDialogFooterComponent,
  HlmDialogHeaderComponent,
  HlmDialogTitleDirective,
} from "@spartan-ng/ui-dialog-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { BrnSelectImports } from "@spartan-ng/brain/select";
import { HlmSelectImports } from "@spartan-ng/ui-select-helm";
import { LucideAngularModule, Plus } from "lucide-angular";
import { WatchlistService } from "../../../../core/services";
import { CreateWatchlistDTO } from "../../../../core/models";
import { toast } from "ngx-sonner";
import { ColorOption } from "../watchlist-manager.interface";

@Component({
  selector: "app-create-watchlist-dialog",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    BrnDialogTriggerDirective,
    BrnDialogContentDirective,
    HlmDialogComponent,
    HlmDialogContentComponent,
    HlmDialogHeaderComponent,
    HlmDialogFooterComponent,
    HlmDialogTitleDirective,
    HlmDialogDescriptionDirective,
    HlmInputDirective,
    HlmButtonDirective,
    BrnSelectImports,
    HlmSelectImports,
    LucideAngularModule,
  ],
  templateUrl: "./create-watchlist-dialog.component.html",
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CreateWatchlistDialogComponent {
  @Output() watchlistCreated = new EventEmitter<void>();

  readonly plusIcon = Plus;
  private fb = inject(FormBuilder);
  private watchlistService = inject(WatchlistService);

  // Form control
  createForm = this.fb.group({
    name: ["", [Validators.required, Validators.minLength(1)]],
    description: [""],
    color: ["#3b82f6", [Validators.required]],
  });

  // Color options with hex values
  readonly colorOptions: ColorOption[] = [
    { value: "#64748b", label: "Slate" },
    { value: "#ef4444", label: "Red" },
    { value: "#f97316", label: "Orange" },
    { value: "#f59e0b", label: "Amber" },
    { value: "#eab308", label: "Yellow" },
    { value: "#84cc16", label: "Lime" },
    { value: "#22c55e", label: "Green" },
    { value: "#10b981", label: "Emerald" },
    { value: "#14b8a6", label: "Teal" },
    { value: "#06b6d4", label: "Cyan" },
    { value: "#3b82f6", label: "Blue" },
    { value: "#6366f1", label: "Indigo" },
    { value: "#8b5cf6", label: "Violet" },
    { value: "#a855f7", label: "Purple" },
    { value: "#d946ef", label: "Fuchsia" },
    { value: "#ec4899", label: "Pink" },
    { value: "#f43f5e", label: "Rose" },
  ];

  getColorLabel(value: string): string {
    return this.colorOptions.find((c) => c.value === value)?.label || "Color";
  }

  createWatchlist(dialogRef: any): void {
    if (this.createForm.valid) {
      const { name, description, color } = this.createForm.value;
      this.watchlistService
        .createWatchlist({
          name: name!,
          description: description || "",
          color: color!,
        })
        .subscribe({
          next: (newWatchlist) => {
            // The watchlist is already set as active in the service
            // Reset the form
            this.createForm.reset({ name: "", description: "", color: "#3b82f6" });
            // Close the dialog
            dialogRef.close();
            // Emit the event after everything is done
            this.watchlistCreated.emit();
          },
          error: () => {
            // Toast is already shown in the service
          },
        });
    }
  }
}
