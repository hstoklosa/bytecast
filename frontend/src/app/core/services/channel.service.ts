import { HttpClient } from "@angular/common/http";
import { Injectable, inject, signal, EventEmitter, Output } from "@angular/core";
import { Observable, catchError, map, throwError, tap } from "rxjs";
import { toast } from "ngx-sonner";
import { AddChannelDTO, Channel } from "../models";

@Injectable({
  providedIn: "root",
})
export class ChannelService {
  private http = inject(HttpClient);
  private apiUrl = "/api/v1";

  // State management using signals
  private _searchResults = signal<Channel[]>([]);
  private _channels = signal<Channel[]>([]);

  // Event emitter for channel changes
  @Output() channelAdded = new EventEmitter<number>();

  // Public readonly signals
  readonly searchResults = this._searchResults.asReadonly();
  readonly channels = this._channels.asReadonly();

  /**
   * Search for YouTube channels by query
   * @param query Search query string
   */
  searchChannels(query: string): Observable<Channel[]> {
    return this.http
      .get<Channel[]>(`${this.apiUrl}/search?q=${encodeURIComponent(query)}`)
      .pipe(
        map((channels) => {
          this._searchResults.set(channels);
          return channels;
        }),
        catchError((error) => {
          toast.error("Error searching channels");
          return throwError(() => error);
        })
      );
  }

  /**
   * Add a channel to a watchlist
   * @param watchlistId ID of the watchlist
   * @param channelId YouTube channel ID or URL
   */
  addChannelToWatchlist(watchlistId: number, channelId: string): Observable<void> {
    const data: AddChannelDTO = { channel_id: channelId };

    return this.http
      .post<void>(`${this.apiUrl}/watchlists/${watchlistId}/channels`, data)
      .pipe(
        tap(() => {
          // After successfully adding a channel, refresh the channels list
          this.getChannelsInWatchlist(watchlistId).subscribe();
          // Emit an event to notify components that a channel was added
          this.channelAdded.emit(watchlistId);
        }),
        map(() => {
          toast.success("Channel added to watchlist");
        }),
        catchError((error) => {
          toast.error("Failed to add channel");
          return throwError(() => error);
        })
      );
  }

  /**
   * Remove a channel from a watchlist
   * @param channelId Database ID of the channel
   * @param watchlistId ID of the watchlist
   * @param youtubeId YouTube ID of the channel
   */
  removeChannelFromWatchlist(
    channelId: number,
    watchlistId: string,
    youtubeId: string
  ): Observable<void> {
    return this.http
      .delete<void>(
        `${this.apiUrl}/watchlists/${watchlistId}/channels/${youtubeId}`
      )
      .pipe(
        tap(() => {
          // After successfully removing a channel, refresh the channels list
          this.getChannelsInWatchlist(Number(watchlistId)).subscribe();
        }),
        map(() => {
          toast.success("Channel removed from watchlist");
        }),
        catchError((error) => {
          toast.error("Failed to remove channel");
          return throwError(() => error);
        })
      );
  }

  /**
   * Get all channels in a watchlist
   * @param watchlistId ID of the watchlist
   */
  getChannelsInWatchlist(watchlistId: number): Observable<Channel[]> {
    return this.http
      .get<{ channels: Channel[] }>(
        `${this.apiUrl}/watchlists/${watchlistId}/channels`
      )
      .pipe(
        map((response) => {
          const channels = response.channels;
          this._channels.set(channels);
          return channels;
        }),
        catchError((error) => {
          toast.error("Failed to fetch channels");
          return throwError(() => error);
        })
      );
  }
}
