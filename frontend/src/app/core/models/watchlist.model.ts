export interface Channel {
  id: number;
  youtube_id: string;
  title: string;
  description: string;
  thumbnail_url: string;
  custom_name?: string;
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

export interface AddChannelDTO {
  channel_id: string;
}
