import { ChangeDetectionStrategy, Component, inject, Input } from "@angular/core";
import { FormControl, ReactiveFormsModule } from "@angular/forms";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { NgIf } from "@angular/common";
import { HlmSpinnerComponent } from "@spartan-ng/ui-spinner-helm";
import { LucideAngularModule, Search, Plus } from "lucide-angular";
import { toast } from "ngx-sonner";

import { ChannelService } from "../../../core/services/channel.service";
import { Channel } from "../../../core/models";

@Component({
  selector: "app-add-channel",
  standalone: true,
  imports: [
    ReactiveFormsModule,
    HlmButtonDirective,
    HlmInputDirective,
    HlmSpinnerComponent,
    NgIf,
    LucideAngularModule,
  ],
  templateUrl: "./add-channel.component.html",
  styleUrls: ["./add-channel.component.css"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AddChannelComponent {
  @Input({ required: true }) watchlistId!: number;

  private channelService = inject(ChannelService);

  searchQuery = new FormControl("");
  isLoading = false;
  searchResults: Channel[] = [];
  searchIcon = Search;
  plusIcon = Plus;

  handleSearch(): void {
    if (!this.searchQuery.value?.trim()) return;

    this.isLoading = true;
    this.channelService
      .addChannelToWatchlist(this.watchlistId, this.searchQuery.value)
      .subscribe({
        next: () => {
          toast.success("Channel added to watchlist");
          this.searchQuery.reset();
          this.isLoading = false;
        },
        error: (error) => {
          toast.error(
            "Failed to add channel: " + (error.error?.message || "Unknown error")
          );
          this.isLoading = false;
        },
      });
  }
}
