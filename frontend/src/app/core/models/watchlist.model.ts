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
  color: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWatchlistDTO {
  name: string;
  description?: string;
  color: string;
}

export interface UpdateWatchlistDTO {
  name: string;
  description?: string;
  color: string;
}
