import { ChangeDetectionStrategy, Component, inject, Input } from "@angular/core";
import { FormControl, ReactiveFormsModule } from "@angular/forms";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import {
  HlmCardContentDirective,
  HlmCardDescriptionDirective,
  HlmCardDirective,
  HlmCardHeaderDirective,
  HlmCardTitleDirective,
} from "@spartan-ng/ui-card-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { NgIf } from "@angular/common";
import { HlmSpinnerComponent } from "@spartan-ng/ui-spinner-helm";
import { LucideAngularModule, Search, Plus } from "lucide-angular";

import { WatchlistService } from "../../../core/services/watchlist.service";
import { Channel } from "../../../core/models";

@Component({
  selector: "app-add-channel",
  standalone: true,
  imports: [
    ReactiveFormsModule,
    HlmButtonDirective,
    HlmCardDirective,
    HlmCardHeaderDirective,
    HlmCardTitleDirective,
    HlmCardDescriptionDirective,
    HlmCardContentDirective,
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

  private watchlistService = inject(WatchlistService);

  searchQuery = new FormControl("");
  isLoading = false;
  searchResults: Channel[] = [];
  searchIcon = Search;
  plusIcon = Plus;

  handleSearch(): void {
    if (!this.searchQuery.value?.trim()) return;

    this.isLoading = true;
    this.watchlistService.searchChannels(this.searchQuery.value).subscribe({
      next: (results) => {
        this.searchResults = results;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      },
    });
  }

  addToWatchlist(channel: Channel): void {
    this.watchlistService
      .addChannelToWatchlist(this.watchlistId, channel.id)
      .subscribe({
        next: () => {
          this.searchResults = [];
          this.searchQuery.reset();
        },
      });
  }
}
