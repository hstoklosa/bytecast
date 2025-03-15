import { ChangeDetectionStrategy, Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import {
  HlmTabsComponent,
  HlmTabsContentDirective,
  HlmTabsListComponent,
  HlmTabsTriggerDirective,
} from "@spartan-ng/ui-tabs-helm";
import { WatchlistManagerComponent } from "./watchlist-manager/watchlist-manager.component";
import { AuthService } from "../../core/auth/auth.service";
import { WatchlistService } from "../../core/services/watchlist.service";

@Component({
  selector: "app-dashboard",
  standalone: true,
  imports: [
    CommonModule,
    HlmButtonDirective,
    HlmTabsComponent,
    HlmTabsListComponent,
    HlmTabsTriggerDirective,
    HlmTabsContentDirective,
    WatchlistManagerComponent,
  ],
  template: `
    <div class="container mx-auto py-6 space-y-8">
      <header class="space-y-2">
        <h1 class="text-3xl font-bold tracking-tight">YouTube Watchlist</h1>
        <p class="text-muted-foreground">
          Keep track of your favorite YouTube channels and get notified when new
          videos are uploaded
        </p>
      </header>

      <app-watchlist-manager />

      <hlm-tabs tab="feed">
        <hlm-tabs-list
          class="w-full grid grid-cols-3"
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
            <button
              hlmBtn
              variant="destructive"
              (click)="onLogout()"
            >
              Logout
            </button>
          </div>
        </div>
      </hlm-tabs>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DashboardComponent {
  private authService = inject(AuthService);
  protected watchlistService = inject(WatchlistService);

  onLogout(): void {
    this.authService.logout().subscribe();
  }
}
