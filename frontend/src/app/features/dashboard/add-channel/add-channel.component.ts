import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { HlmButtonDirective } from '@spartan-ng/ui-button-helm';
import {
  HlmCardContentDirective,
  HlmCardDescriptionDirective,
  HlmCardDirective,
  HlmCardHeaderDirective,
  HlmCardTitleDirective
} from '@spartan-ng/ui-card-helm';
import { HlmInputDirective } from '@spartan-ng/ui-input-helm';
import { WatchlistService } from '../../../core/services/watchlist.service';
import { NgIf } from '@angular/common';
import { HlmSpinnerComponent } from '@spartan-ng/ui-spinner-helm';
import { Channel } from '../../../core/services/watchlist.service';
import { LucideAngularModule, Search, Plus } from 'lucide-angular';

@Component({
  selector: 'app-add-channel',
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
    LucideAngularModule
  ],
  template: `
    <section hlmCard>
      <div hlmCardContent class="pt-6">
        <form (ngSubmit)="handleSearch()" class="flex w-full items-center space-x-2">
          <input
            [formControl]="searchQuery"
            type="text"
            placeholder="Search for YouTube channels"
            hlmInput
            class="flex-1"
          />
          <button hlmBtn type="submit" [disabled]="isLoading">
            <hlm-spinner *ngIf="isLoading" size="sm" class="mr-2" />
            <lucide-angular
              *ngIf="!isLoading"
              [img]="searchIcon"
              class="mr-2 h-4 w-4"
            ></lucide-angular>
            Search
          </button>
        </form>

        @if (searchResults.length > 0) {
          <div class="mt-4 space-y-4">
            <h3 class="font-medium">Search Results</h3>
            <div class="space-y-3">
              @for (channel of searchResults; track channel.id) {
                <div class="flex items-center justify-between">
                  <div class="flex items-center space-x-3">
                    <img
                      [src]="channel.thumbnailUrl"
                      [alt]="channel.title"
                      class="h-11 w-11 rounded-full object-cover"
                    />
                    <div>
                      <h4 class="font-medium">{{ channel.title }}</h4>
                      <p class="text-sm text-muted-foreground">
                        {{ channel.subscriberCount }} subscribers
                      </p>
                    </div>
                  </div>
                  <button hlmBtn size="sm" (click)="addToWatchlist(channel)">
                    <lucide-angular
                      [img]="plusIcon"
                      class="mr-2 h-4 w-4"
                    ></lucide-angular>
                    Add
                  </button>
                </div>
              }
            </div>
          </div>
        }
      </div>
    </section>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AddChannelComponent {
  private watchlistService = inject(WatchlistService);

  searchQuery = new FormControl('');
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
      }
    });
  }

  addToWatchlist(channel: Channel): void {
    this.watchlistService.addToWatchlist(channel.id).subscribe({
      next: () => {
        this.searchResults = [];
        this.searchQuery.reset();
      }
    });
  }
}
