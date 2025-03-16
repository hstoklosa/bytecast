import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  signal,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormBuilder, ReactiveFormsModule, Validators } from "@angular/forms";
import { WatchlistService } from "../../../core/services/watchlist.service";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { BrnSelectImports } from "@spartan-ng/brain/select";
import { HlmSelectImports } from "@spartan-ng/ui-select-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import {
  HlmCardContentDirective,
  HlmCardDirective,
} from "@spartan-ng/ui-card-helm";
import { HlmBadgeDirective } from "@spartan-ng/ui-badge-helm";
import { AddChannelComponent } from "../add-channel/add-channel.component";
import { toast } from "ngx-sonner";
import { LucideAngularModule, Edit, Trash2, Plus, Check, X } from "lucide-angular";
import { ColorOption } from "./watchlist-manager.interface";
import { ConfirmationDialogComponent } from "../../../shared/components";

@Component({
  selector: "app-watchlist-manager",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    BrnSelectImports,
    HlmSelectImports,
    HlmInputDirective,
    HlmButtonDirective,
    HlmCardDirective,
    HlmCardContentDirective,
    HlmBadgeDirective,
    AddChannelComponent,
    LucideAngularModule,
    ConfirmationDialogComponent,
  ],
  templateUrl: "./watchlist-manager.component.html",
  styleUrls: ["./watchlist-manager.component.css"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class WatchlistManagerComponent {
  readonly editIcon = Edit;
  readonly trashIcon = Trash2;
  readonly plusIcon = Plus;
  readonly checkIcon = Check;
  readonly xIcon = X;
  private fb = inject(FormBuilder);
  private watchlistService = inject(WatchlistService);

  // State signals
  readonly watchlists = this.watchlistService.watchlists;
  readonly selectedWatchlistId = signal<number | null>(null);
  readonly isCreating = signal(false);
  readonly isEditing = signal(false);

  // Computed signal for the active watchlist
  readonly activeWatchlist = computed(() => {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return null;
    return this.watchlists()?.find((w) => w.id === selectedId) || null;
  });

  // Form controls
  watchlistForm = this.fb.group({
    selectedWatchlist: [null as number | null],
  });

  createForm = this.fb.group({
    name: ["", [Validators.required, Validators.minLength(1)]],
    description: [""],
    color: ["#3b82f6", [Validators.required]],
  });

  editForm = this.fb.group({
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

  constructor() {
    // Restore active watchlist from localStorage
    const storedWatchlistId = localStorage.getItem("activeWatchlist");
    if (storedWatchlistId) {
      const watchlistId = parseInt(storedWatchlistId, 10);
      this.selectedWatchlistId.set(watchlistId);
      this.watchlistForm.patchValue({ selectedWatchlist: watchlistId });
    }

    // Initialize by fetching watchlists
    this.watchlistService
      .refreshWatchlists()
      .pipe(takeUntilDestroyed())
      .subscribe({
        next: (watchlists) => {
          if (watchlists.length > 0 && !this.selectedWatchlistId()) {
            this.selectedWatchlistId.set(watchlists[0].id);
            this.watchlistForm.patchValue({ selectedWatchlist: watchlists[0].id });
          }
        },
        error: () => {
          toast.error("Failed to load watchlists");
        },
      });

    // React to watchlist selection changes
    this.watchlistForm.get("selectedWatchlist")?.valueChanges.subscribe((id) => {
      if (id !== null) {
        this.onWatchlistSelect(id);
        const watchlist = this.watchlists()?.find((w) => w.id === id);
        if (watchlist) {
          this.watchlistService.setActiveWatchlist(watchlist);
        }
      }
    });
  }

  // UI Helpers
  getSelectedWatchlistName(): string {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return "Select a watchlist";
    const watchlist = this.watchlists()?.find((w) => w.id === selectedId);
    return watchlist ? watchlist.name : "Select a watchlist";
  }

  getColorLabel(value: string): string {
    return this.colorOptions.find((c) => c.value === value)?.label || "Color";
  }

  // Event Handlers
  onWatchlistSelect(id: number | null): void {
    this.selectedWatchlistId.set(id);
    if (id) {
      localStorage.setItem("activeWatchlist", id.toString());
    } else {
      localStorage.removeItem("activeWatchlist");
      this.watchlistService.setActiveWatchlist(null);
    }
  }

  // Create Operations
  startCreating(): void {
    this.isCreating.set(true);
    this.createForm.reset({ name: "", description: "", color: "#3b82f6" });
  }

  cancelCreate(): void {
    this.isCreating.set(false);
    this.createForm.reset();
  }

  createWatchlist(): void {
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
            this.selectedWatchlistId.set(newWatchlist.id);
            this.watchlistService.setActiveWatchlist(newWatchlist);
            this.isCreating.set(false);
            this.createForm.reset();
          },
          error: () => {
            toast.error("Failed to create watchlist");
          },
        });
    }
  }

  // Edit Operations
  startEditing(): void {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return;

    const watchlist = this.watchlists()?.find((w) => w.id === selectedId);
    if (watchlist) {
      this.editForm.patchValue({
        name: watchlist.name,
        description: watchlist.description || "",
        color: watchlist.color,
      });
      this.isEditing.set(true);
    }
  }

  cancelEdit(): void {
    this.isEditing.set(false);
    this.editForm.reset();
  }

  saveEdit(): void {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId || !this.editForm.valid) return;

    const { name, description, color } = this.editForm.value;
    this.watchlistService
      .updateWatchlist(selectedId, {
        name: name!,
        description: description || "",
        color: color!,
      })
      .subscribe({
        next: () => {
          this.isEditing.set(false);
          this.editForm.reset();
        },
        error: () => {
          toast.error("Failed to update watchlist");
        },
      });
  }

  // Delete Operation
  deleteWatchlist(): void {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return;

    // Don't allow deleting the last watchlist
    if (this.watchlists()?.length <= 1) {
      toast.error("You must have at least one watchlist");
      return;
    }

    const watchlist = this.watchlists()?.find((w) => w.id === selectedId);
    if (!watchlist) return;
  }

  /**
   * Get the confirmation message for deleting a watchlist
   */
  getDeleteConfirmationMessage(): string {
    const watchlist = this.activeWatchlist();
    if (!watchlist) return "Are you sure you want to delete this watchlist?";

    return `Are you sure you want to delete "${watchlist.name}" watchlist? This action cannot be undone.`;
  }

  // Handle confirmation from the dialog
  onDeleteConfirmed(): void {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return;

    // Don't allow deleting the last watchlist
    if (this.watchlists()?.length <= 1) {
      toast.error("You must have at least one watchlist");
      return;
    }

    this.watchlistService.deleteWatchlist(selectedId).subscribe({
      next: () => {
        // Switch to the first available watchlist
        const remainingWatchlist = this.watchlists()?.find(
          (w) => w.id !== selectedId
        );
        if (remainingWatchlist) {
          this.selectedWatchlistId.set(remainingWatchlist.id);
        } else {
          this.selectedWatchlistId.set(null);
        }
      },
      error: () => {
        toast.error("Failed to delete watchlist");
      },
    });
  }
}
