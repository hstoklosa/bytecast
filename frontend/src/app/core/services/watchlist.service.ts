import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, map, tap } from 'rxjs/operators';
import { Observable, throwError } from 'rxjs';
import { toast } from 'ngx-sonner';

export interface Channel {
  id: string;
  title: string;
  description: string;
  thumbnailUrl: string;
  subscriberCount: string;
}

export interface Watchlist {
  id: number;
  name: string;
  description: string;
  channels: Channel[];
}

@Injectable({
  providedIn: 'root'
})
export class WatchlistService {
  private http = inject(HttpClient);
  private apiUrl = '/api/watchlist';

  private _channels = signal<Channel[]>([]);
  channels = this._channels.asReadonly();

  searchChannels(query: string): Observable<Channel[]> {
    // In a real app, this would call the YouTube API
    // For now, returning mock data matching the shape we expect
    return this.http.get<Channel[]>(`${this.apiUrl}/search?q=${query}`).pipe(
      catchError(error => {
        console.error('Error searching channels:', error);
        return throwError(() => error);
      })
    );
  }

  addToWatchlist(channelId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}`, { channelId }).pipe(
      tap(() => {
        toast.success('Channel added to your watchlist');
        // Optimistically update the UI
        this.refreshWatchlist().subscribe();
      }),
      catchError(error => {
        console.error('Error adding channel to watchlist:', error);
        return throwError(() => error);
      })
    );
  }

  refreshWatchlist(): Observable<Channel[]> {
    return this.http.get<Channel[]>(`${this.apiUrl}/channels`).pipe(
      tap(channels => {
        this._channels.set(channels);
      }),
      catchError(error => {
        console.error('Error fetching watchlist:', error);
        return throwError(() => error);
      })
    );
  }
}
