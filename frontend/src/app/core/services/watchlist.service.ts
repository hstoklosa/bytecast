import { HttpClient } from "@angular/common/http";
import { Injectable, inject, signal } from "@angular/core";
import { catchError, map, tap } from "rxjs/operators";
import { Observable, throwError } from "rxjs";
import { toast } from "ngx-sonner";
import {
  Channel,
  CreateWatchlistDTO,
  UpdateWatchlistDTO,
  Watchlist,
} from "../models";

@Injectable({
  providedIn: "root",
})
export class WatchlistService {
  private http = inject(HttpClient);
  private apiUrl = "/api/v1";

  // State management using signals
  private _channels = signal<Channel[]>([]);
  private _watchlists = signal<Watchlist[]>([]);
  private _activeWatchlist = signal<Watchlist | null>(null);

  // Public readonly signals
  readonly channels = this._channels.asReadonly();
  readonly watchlists = this._watchlists.asReadonly();
  readonly activeWatchlist = this._activeWatchlist.asReadonly();

  // Initialize by loading watchlists
  constructor() {
    this.refreshWatchlists().subscribe();

    // Try to restore active watchlist from localStorage
    const storedWatchlistId = localStorage.getItem("activeWatchlist");
    if (storedWatchlistId) {
      this.getWatchlist(parseInt(storedWatchlistId, 10)).subscribe((watchlist) =>
        this.setActiveWatchlist(watchlist)
      );
    }
  }

  // Watchlist CRUD operations
  createWatchlist(data: CreateWatchlistDTO): Observable<Watchlist> {
    return this.http.post<Watchlist>(`${this.apiUrl}/watchlists`, data).pipe(
      tap((newWatchlist) => {
        this._watchlists.update((current) => [...current, newWatchlist]);
        toast.success("Watchlist created successfully");
      }),
      catchError((error) => {
        toast.error("Failed to create watchlist");
        return throwError(() => error);
      })
    );
  }

  getWatchlist(id: number): Observable<Watchlist> {
    return this.http.get<Watchlist>(`${this.apiUrl}/watchlists/${id}`).pipe(
      catchError((error) => {
        toast.error("Failed to fetch watchlist");
        return throwError(() => error);
      })
    );
  }

  refreshWatchlists(): Observable<Watchlist[]> {
    return this.http
      .get<{ watchlists: Watchlist[] }>(`${this.apiUrl}/watchlists`)
      .pipe(
        map((response) => response.watchlists),
        tap((watchlists) => {
          this._watchlists.set(watchlists);
        }),
        catchError((error) => {
          toast.error("Failed to fetch watchlists");
          return throwError(() => error);
        })
      );
  }

  updateWatchlist(id: number, data: UpdateWatchlistDTO): Observable<Watchlist> {
    return this.http.put<Watchlist>(`${this.apiUrl}/watchlists/${id}`, data).pipe(
      tap((updatedWatchlist) => {
        this._watchlists.update((current) =>
          current.map((w) => (w.id === id ? updatedWatchlist : w))
        );
        if (this._activeWatchlist()?.id === id) {
          this._activeWatchlist.set(updatedWatchlist);
        }
        toast.success("Watchlist updated successfully");
      }),
      catchError((error) => {
        toast.error("Failed to update watchlist");
        return throwError(() => error);
      })
    );
  }

  deleteWatchlist(id: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/watchlists/${id}`).pipe(
      tap(() => {
        this._watchlists.update((current) => current.filter((w) => w.id !== id));
        if (this._activeWatchlist()?.id === id) {
          this._activeWatchlist.set(null);
          localStorage.removeItem("activeWatchlist");
        }
        toast.success("Watchlist deleted successfully");
      }),
      catchError((error) => {
        toast.error("Failed to delete watchlist");
        return throwError(() => error);
      })
    );
  }

  // Active watchlist management
  setActiveWatchlist(watchlist: Watchlist | null): void {
    this._activeWatchlist.set(watchlist);
    if (watchlist) {
      localStorage.setItem("activeWatchlist", watchlist.id.toString());
      this.refreshWatchlistChannels(watchlist.id);
    } else {
      localStorage.removeItem("activeWatchlist");
      this._channels.set([]);
    }
  }

  // Channel operations
  searchChannels(query: string): Observable<Channel[]> {
    return this.http.get<Channel[]>(`${this.apiUrl}/search?q=${query}`).pipe(
      catchError((error) => {
        toast.error("Error searching channels");
        return throwError(() => error);
      })
    );
  }

  addChannelToWatchlist(watchlistId: number, channelId: string): Observable<void> {
    return this.http
      .post<void>(`${this.apiUrl}/watchlists/${watchlistId}/channels`, {
        channel_id: channelId,
      })
      .pipe(
        tap(() => {
          toast.success("Channel added to watchlist");
          this.refreshWatchlistChannels(watchlistId);
        }),
        catchError((error) => {
          toast.error("Failed to add channel");
          return throwError(() => error);
        })
      );
  }

  removeChannelFromWatchlist(
    watchlistId: number,
    channelId: string
  ): Observable<void> {
    return this.http
      .delete<void>(
        `${this.apiUrl}/watchlists/${watchlistId}/channels/${channelId}`
      )
      .pipe(
        tap(() => {
          toast.success("Channel removed from watchlist");
          this.refreshWatchlistChannels(watchlistId);
        }),
        catchError((error) => {
          toast.error("Failed to remove channel");
          return throwError(() => error);
        })
      );
  }

  private refreshWatchlistChannels(watchlistId: number): void {
    this.http
      .get<{ channels: Channel[] }>(
        `${this.apiUrl}/watchlists/${watchlistId}/channels`
      )
      .pipe(
        map((response) => response.channels),
        tap((channels) => this._channels.set(channels)),
        catchError((error) => {
          toast.error("Failed to fetch channels");
          return throwError(() => error);
        })
      )
      .subscribe();
  }
}
