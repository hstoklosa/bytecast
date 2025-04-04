import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  inject,
  signal,
  ViewChild,
  QueryList,
  ViewChildren,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormBuilder,
  FormControl,
  ReactiveFormsModule,
  Validators,
} from "@angular/forms";
import { WatchlistService, ChannelService } from "../../../core/services";
import { takeUntilDestroyed, toSignal } from "@angular/core/rxjs-interop";
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
import {
  LucideAngularModule,
  Trash2,
  Plus,
  Search,
  Edit,
  Bell,
  ExternalLink,
  MoreHorizontal,
} from "lucide-angular";
import { ColorOption } from "./watchlist-manager.interface";
import { ConfirmationDialogComponent } from "../../../shared/components";
import { CreateWatchlistDialogComponent } from "./create-watchlist-dialog/create-watchlist-dialog.component";
import { EditWatchlistDialogComponent } from "./edit-watchlist-dialog/edit-watchlist-dialog.component";
import { ChannelCardComponent } from "./channel-card";
import {
  HlmTabsComponent,
  HlmTabsContentDirective,
  HlmTabsListComponent,
  HlmTabsTriggerDirective,
} from "@spartan-ng/ui-tabs-helm";
import { NgIconComponent, provideIcons } from "@ng-icons/core";
import { HlmIconDirective } from "@spartan-ng/ui-icon-helm";
import {
  lucideLayoutDashboard,
  lucideLayoutGrid,
  lucideList,
} from "@ng-icons/lucide";
import { Watchlist } from "../../../core/models";
import { BrnMenuTriggerDirective } from "@spartan-ng/brain/menu";
import {
  HlmMenuComponent,
  HlmMenuItemDirective,
  HlmMenuGroupComponent,
  HlmMenuLabelComponent,
  HlmMenuSeparatorComponent,
} from "@spartan-ng/ui-menu-helm";

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
    CreateWatchlistDialogComponent,
    EditWatchlistDialogComponent,
    HlmTabsComponent,
    HlmTabsListComponent,
    HlmTabsTriggerDirective,
    HlmTabsContentDirective,
    NgIconComponent,
    HlmIconDirective,
    BrnMenuTriggerDirective,
    HlmMenuComponent,
    HlmMenuItemDirective,
  ],
  providers: [
    provideIcons({
      lucideLayoutDashboard,
      lucideLayoutGrid,
      lucideList,
    }),
  ],
  templateUrl: "./watchlist-manager.component.html",
  styleUrls: ["./watchlist-manager.component.css"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class WatchlistManagerComponent {
  // Icons
  readonly trashIcon = Trash2;
  readonly plusIcon = Plus;
  readonly searchIcon = Search;
  readonly editIcon = Edit;
  readonly bellIcon = Bell;
  readonly externalLinkIcon = ExternalLink;
  readonly moreIcon = MoreHorizontal;

  private fb = inject(FormBuilder);
  private watchlistService = inject(WatchlistService);
  private channelService = inject(ChannelService);

  // State signals
  readonly watchlists = this.watchlistService.watchlists;
  readonly channels = this.watchlistService.channels;
  readonly selectedWatchlistId = signal<number | null>(null);
  readonly isEditing = signal(false);
  readonly watchlistChannelCounts = signal<Map<number, number>>(new Map());

  // View mode state
  readonly viewMode = signal<"default" | "grid" | "list">("default");

  // Search controls with toSignal for reactive updates
  readonly watchlistSearchControl = new FormControl("");
  readonly channelSearchControl = new FormControl("");

  // Convert form control values to signals
  readonly watchlistSearchTerm = toSignal(
    this.watchlistSearchControl.valueChanges,
    { initialValue: "" }
  );
  readonly channelSearchTerm = toSignal(this.channelSearchControl.valueChanges, {
    initialValue: "",
  });

  // Hardcoded notifications state
  private notificationsEnabled = new Set<number>();

  // Computed signal for the active watchlist
  readonly activeWatchlist = computed(() => {
    const selectedId = this.selectedWatchlistId();
    if (!selectedId) return null;
    return this.watchlists()?.find((w) => w.id === selectedId) || null;
  });

  // Filtered watchlists based on search
  readonly filteredWatchlists = computed(() => {
    const searchTerm = this.watchlistSearchTerm()?.toLowerCase() || "";
    if (!searchTerm) return this.watchlists() || [];

    return (
      this.watchlists()?.filter((watchlist) =>
        watchlist.name.toLowerCase().includes(searchTerm)
      ) || []
    );
  });

  // Filtered channels based on search
  readonly filteredChannels = computed(() => {
    const searchTerm = this.channelSearchTerm()?.toLowerCase() || "";
    if (!searchTerm) return this.channels() || [];

    return (
      this.channels()?.filter((channel) =>
        channel.title.toLowerCase().includes(searchTerm)
      ) || []
    );
  });

  // Form controls
  watchlistForm = this.fb.group({
    selectedWatchlist: [null as number | null],
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

  @ViewChildren("editWatchlistDialog")
  editWatchlistDialogs!: QueryList<EditWatchlistDialogComponent>;

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
          // Load channel counts for all watchlists
          this.loadAllWatchlistChannelCounts(watchlists);

          // Set active watchlist after watchlists are loaded
          if (watchlists.length > 0) {
            let activeWatchlistId = this.selectedWatchlistId();

            // If no active watchlist is set, use the first one
            if (!activeWatchlistId) {
              activeWatchlistId = watchlists[0].id;
              this.selectedWatchlistId.set(activeWatchlistId);
              this.watchlistForm.patchValue({
                selectedWatchlist: activeWatchlistId,
              });
            } else {
              // Verify the stored watchlist ID exists in the loaded watchlists
              const watchlistExists = watchlists.some(
                (w) => w.id === activeWatchlistId
              );
              if (!watchlistExists) {
                activeWatchlistId = watchlists[0].id;
                this.selectedWatchlistId.set(activeWatchlistId);
                this.watchlistForm.patchValue({
                  selectedWatchlist: activeWatchlistId,
                });
              }
            }

            // Set the active watchlist in the service and load its channels
            const activeWatchlist = watchlists.find(
              (w) => w.id === activeWatchlistId
            );
            if (activeWatchlist) {
              this.watchlistService.setActiveWatchlist(activeWatchlist);
              // Explicitly load channels for the active watchlist
              this.refreshChannels(activeWatchlistId);
            }
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
      }
    });

    // Initialize search terms with current values
    this.watchlistSearchControl.setValue("", { emitEvent: true });
    this.channelSearchControl.setValue("", { emitEvent: true });

    // Subscribe to channel service channels signal to keep UI updated
    effect(() => {
      // This effect will run whenever the channels signal changes
      const updatedChannels = this.channelService.channels();
      if (updatedChannels) {
        // Update the watchlist service channels
        this.watchlistService.updateChannels(updatedChannels);
      }
    });

    // Effect to update the form when the active watchlist changes
    effect(() => {
      const activeWatchlist = this.watchlistService.activeWatchlist();
      if (activeWatchlist) {
        this.selectedWatchlistId.set(activeWatchlist.id);
        this.watchlistForm.patchValue(
          { selectedWatchlist: activeWatchlist.id },
          { emitEvent: false }
        );
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

  // Get channel count for a watchlist
  getChannelCount(watchlistId: number): number {
    // Get the count from the watchlistChannelCounts map
    return this.watchlistChannelCounts()?.get(watchlistId) || 0;
  }

  // Load channel counts for all watchlists
  private loadAllWatchlistChannelCounts(watchlists: Watchlist[]): void {
    const counts = new Map<number, number>();

    // Initialize all counts to 0
    watchlists.forEach((watchlist) => {
      counts.set(watchlist.id, 0);
    });

    // Set initial counts
    this.watchlistChannelCounts.set(counts);

    // Fetch actual counts for each watchlist
    watchlists.forEach((watchlist) => {
      this.channelService.getChannelsInWatchlist(watchlist.id).subscribe({
        next: (channels) => {
          // Update the count for this watchlist
          const updatedCounts = new Map(this.watchlistChannelCounts());
          updatedCounts.set(watchlist.id, channels.length);
          this.watchlistChannelCounts.set(updatedCounts);
        },
        error: () => {
          // Keep the count at 0 if there's an error
        },
      });
    });
  }

  // Check if a channel has notifications enabled
  hasNotifications(channelId: number): boolean {
    return this.notificationsEnabled.has(channelId);
  }

  // Toggle channel notifications
  toggleNotifications(channelId: number): void {
    // Hardcoded implementation for now
    if (this.notificationsEnabled.has(channelId)) {
      this.notificationsEnabled.delete(channelId);
      toast.success("Notifications disabled");
    } else {
      this.notificationsEnabled.add(channelId);
      toast.success("Notifications enabled");
    }
  }

  // Remove channel from watchlist
  removeChannel(channelId: number): void {
    const watchlistId = this.selectedWatchlistId();
    if (!watchlistId) return;

    // Find the channel to get its YouTube ID
    const channel = this.channels()?.find((c) => c.id === channelId);
    if (!channel) {
      toast.error("Channel not found");
      return;
    }

    this.channelService
      .removeChannelFromWatchlist(
        channelId,
        String(watchlistId),
        channel.youtube_id
      )
      .subscribe({
        next: () => {
          // The channel list will be refreshed by the channelService
        },
        error: () => {
          toast.error("Failed to remove channel from watchlist");
        },
      });
  }

  // Event Handlers
  onWatchlistSelect(id: number | null): void {
    this.selectedWatchlistId.set(id);
    if (id) {
      localStorage.setItem("activeWatchlist", id.toString());

      // Find the watchlist object and set it as active in the service
      const watchlist = this.watchlists()?.find((w) => w.id === id);
      if (watchlist) {
        // First set the active watchlist in the service
        this.watchlistService.setActiveWatchlist(watchlist);

        // Then explicitly load channels for this watchlist to ensure correct data
        this.channelService.getChannelsInWatchlist(id).subscribe({
          next: (channels) => {
            // Update the channels in the watchlistService
            this.watchlistService.updateChannels(channels);

            // Update the channel count for this watchlist
            const updatedCounts = new Map(this.watchlistChannelCounts());
            updatedCounts.set(id, channels.length);
            this.watchlistChannelCounts.set(updatedCounts);
          },
          error: () => {
            toast.error("Failed to load channels for selected watchlist");
          },
        });
      }
    } else {
      localStorage.removeItem("activeWatchlist");
      this.watchlistService.setActiveWatchlist(null);
    }
  }

  // Handle watchlist created/updated events
  onWatchlistCreated(): void {
    // Refresh watchlists
    this.watchlistService.refreshWatchlists().subscribe({
      next: (watchlists) => {
        // Get the active watchlist from the service and set it as selected
        const activeWatchlist = this.watchlistService.activeWatchlist();
        if (activeWatchlist) {
          this.selectedWatchlistId.set(activeWatchlist.id);
          this.watchlistForm.patchValue({ selectedWatchlist: activeWatchlist.id });

          // Explicitly load channels for the newly selected watchlist
          this.channelService.getChannelsInWatchlist(activeWatchlist.id).subscribe({
            next: (channels) => {
              // Update the channels in the watchlistService
              this.watchlistService.updateChannels(channels);

              // Update the channel count for this watchlist
              const updatedCounts = new Map(this.watchlistChannelCounts());
              updatedCounts.set(activeWatchlist.id, channels.length);
              this.watchlistChannelCounts.set(updatedCounts);
            },
          });
        }

        // Update channel counts for all watchlists
        this.loadAllWatchlistChannelCounts(watchlists);
      },
    });
  }

  onWatchlistUpdated(): void {
    // Refresh watchlists
    this.watchlistService.refreshWatchlists().subscribe({
      next: (watchlists) => {
        // Update channel counts for all watchlists
        this.loadAllWatchlistChannelCounts(watchlists);

        // Refresh channels for the currently selected watchlist
        const activeWatchlistId = this.selectedWatchlistId();
        if (activeWatchlistId) {
          this.channelService.getChannelsInWatchlist(activeWatchlistId).subscribe({
            next: (channels) => {
              // Update the channels in the watchlistService
              this.watchlistService.updateChannels(channels);
            },
          });
        }
      },
    });
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

  // Handle channel removed event
  onChannelRemoved(): void {
    const watchlistId = this.selectedWatchlistId();
    if (watchlistId) {
      this.refreshChannels(watchlistId);

      // Update the channel count for this watchlist
      this.channelService.getChannelsInWatchlist(watchlistId).subscribe({
        next: (channels) => {
          const updatedCounts = new Map(this.watchlistChannelCounts());
          updatedCounts.set(watchlistId, channels.length);
          this.watchlistChannelCounts.set(updatedCounts);
        },
      });
    }
  }

  // Handle channel added event
  onChannelAdded(): void {
    const watchlistId = this.selectedWatchlistId();
    if (watchlistId) {
      this.refreshChannels(watchlistId);

      // Update the channel count for this watchlist
      this.channelService.getChannelsInWatchlist(watchlistId).subscribe({
        next: (channels) => {
          const updatedCounts = new Map(this.watchlistChannelCounts());
          updatedCounts.set(watchlistId, channels.length);
          this.watchlistChannelCounts.set(updatedCounts);
        },
      });
    }
  }

  private refreshChannels(watchlistId: number): void {
    // This method is used to explicitly refresh channels for a watchlist
    // and update both the watchlistService channels and the channel counts
    this.channelService.getChannelsInWatchlist(watchlistId).subscribe({
      next: (channels) => {
        // Update the channels in the watchlistService
        this.watchlistService.updateChannels(channels);

        // Update the channel count for this watchlist
        const updatedCounts = new Map(this.watchlistChannelCounts());
        updatedCounts.set(watchlistId, channels.length);
        this.watchlistChannelCounts.set(updatedCounts);
      },
      error: () => {
        toast.error("Failed to refresh channels");
      },
    });
  }

  // Set the current view mode
  setViewMode(mode: "default" | "grid" | "list"): void {
    this.viewMode.set(mode);
  }

  // Get YouTube channel URL
  getYoutubeChannelUrl(channelId: string): string {
    return `https://www.youtube.com/channel/${channelId}`;
  }

  // Open edit dialog programmatically
  openEditDialog(watchlist: Watchlist, event: MouseEvent): void {
    event.stopPropagation();

    // Find the dialog for this watchlist
    const dialogs = this.editWatchlistDialogs.toArray();
    const dialog = dialogs.find((d) => d.watchlist?.id === watchlist.id);

    if (dialog) {
      dialog.openDialog();
    }
  }
}
