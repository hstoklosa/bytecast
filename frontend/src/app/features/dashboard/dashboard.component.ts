import { ChangeDetectionStrategy, Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  HlmTabsComponent,
  HlmTabsContentDirective,
  HlmTabsListComponent,
  HlmTabsTriggerDirective,
} from "@spartan-ng/ui-tabs-helm";
import { WatchlistManagerComponent } from "./watchlist-manager/watchlist-manager.component";
import { WatchlistService } from "../../core/services/watchlist.service";
import { DashboardLayoutComponent } from "../../layout/dashboard-layout/dashboard-layout.component";

@Component({
  selector: "app-dashboard",
  standalone: true,
  imports: [
    CommonModule,
    HlmTabsComponent,
    HlmTabsListComponent,
    HlmTabsTriggerDirective,
    HlmTabsContentDirective,
    WatchlistManagerComponent,
    DashboardLayoutComponent,
  ],
  template: `
    <app-dashboard-layout>
      <app-watchlist-manager />

      <hlm-tabs tab="feed">
        <hlm-tabs-list
          class="w-full grid grid-cols-3 mt-8 mb-6"
          aria-label="Manage your watchlist"
        >
          <button hlmTabsTrigger="feed">Video Feed</button>
          <button hlmTabsTrigger="channels">Channels</button>
          <button hlmTabsTrigger="settings">Settings</button>
        </hlm-tabs-list>

        <div hlmTabsContent="feed">
          <h2 class="text-xl font-semibold mb-4">Feed</h2>
          <p>Your personalized video feed will appear here.</p>
        </div>

        <div hlmTabsContent="channels">
          <h2 class="text-xl font-semibold mb-4">Channels</h2>
          @if (!watchlistService.activeWatchlist()) {
          <p class="text-muted-foreground">
            Please select a watchlist to view its channels.
          </p>
          } @else if (watchlistService.channels().length) {
          <div class="grid gap-4">
            @for (channel of watchlistService.channels(); track channel.id) {
            <div
              class="flex items-center justify-between p-4 rounded-lg border-2 transition-all duration-200"
              [ngStyle]="{
                borderColor: watchlistService.activeWatchlist()?.color || '#e5e7eb',
                backgroundColor: watchlistService.activeWatchlist()?.color + '10'
              }"
            >
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
            </div>
            }
          </div>
          } @else {
          <p class="text-muted-foreground">
            No channels in "{{ watchlistService.activeWatchlist()?.name }}" yet. Use
            the search above to add channels.
          </p>
          }
        </div>

        <div hlmTabsContent="settings">
          <h2 class="text-xl font-semibold mb-4">Settings</h2>
          <div class="space-y-4">
            <p>Manage your notification preferences and account settings.</p>
          </div>
        </div>
      </hlm-tabs>
    </app-dashboard-layout>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DashboardComponent {
  protected watchlistService = inject(WatchlistService);
}
