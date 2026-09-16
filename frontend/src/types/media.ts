export interface MediaItem {
  id: string;
  type: 'image' | 'video' | 'blank';
  url: string;
  duration_sec: number;
}

export interface WindowPlaybackState {
  current_item: MediaItem;
  offset_sec: number;
}

export interface SyncStateResponse {
  sync_active: boolean;
  sync_item?: MediaItem;
  sync_ends_at?: string;
  windows: Record<string, WindowPlaybackState>;
}
