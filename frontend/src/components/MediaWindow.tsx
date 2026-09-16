import React, { useRef, useEffect } from 'react';
import { usePlayback } from '../context/PlaybackContext';
import type { MediaItem } from '../types/media';

interface MediaWindowProps {
  windowId: string;
}

export const MediaWindow: React.FC<MediaWindowProps> = ({ windowId }) => {
  const { state, loading, error } = usePlayback();
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const isSyncActive = state?.sync_active || false;
  const syncItem = state?.sync_item;
  const windowState = state?.windows?.[windowId];

  const currentItem: MediaItem | undefined = isSyncActive
    ? syncItem
    : windowState?.current_item;

  const offsetSec = isSyncActive ? 0 : windowState?.offset_sec || 0;

  // Handle video element seeking/sync when switching items or offset changes significantly
  useEffect(() => {
    if (currentItem?.type === 'video' && videoRef.current) {
      // If natural duration playback, autoplay muted
      videoRef.current.play().catch(() => {
        // Handle autoplay policy restriction if unmuted
      });
    }
  }, [currentItem?.id, isSyncActive]);

  return (
    <div className="media-window-card">
      <div className="window-header">
        <span className="window-title">Window {windowId}</span>
        {isSyncActive ? (
          <span className="badge sync-badge">SYNC MODE</span>
        ) : (
          <span className="badge live-badge">LOOPING</span>
        )}
      </div>

      <div className="media-viewport">
        {loading && !state ? (
          <div className="media-placeholder">Loading state...</div>
        ) : error ? (
          <div className="media-placeholder error">Error: {error}</div>
        ) : !currentItem || !currentItem.url ? (
          <div className="media-placeholder blank">Blank Screen</div>
        ) : currentItem.type === 'video' ? (
          <video
            ref={videoRef}
            key={currentItem.id + (isSyncActive ? '_sync' : '_norm')}
            src={currentItem.url}
            className="media-content"
            autoPlay
            muted
            playsInline
            controls={false}
          />
        ) : (
          <img
            key={currentItem.id + (isSyncActive ? '_sync' : '_norm')}
            src={currentItem.url}
            alt={currentItem.id}
            className="media-content"
          />
        )}
      </div>

      <div className="window-footer">
        <div className="item-info">
          <strong>Item:</strong> {currentItem?.id || 'N/A'} ({currentItem?.type || 'blank'})
        </div>
        {!isSyncActive && (
          <div className="offset-info">
            <strong>Offset:</strong> {offsetSec.toFixed(1)}s / {currentItem?.duration_sec ? `${currentItem.duration_sec}s` : 'auto'}
          </div>
        )}
      </div>
    </div>
  );
};
