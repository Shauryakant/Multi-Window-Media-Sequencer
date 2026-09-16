import React, { useRef, useEffect, useState } from 'react';
import { usePlayback } from '../context/PlaybackContext';
import type { MediaItem } from '../types/media';

interface MediaWindowProps {
  windowId: string;
}

export const MediaWindow: React.FC<MediaWindowProps> = ({ windowId }) => {
  const { state, loading, error } = usePlayback();
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const [isMediaLoading, setIsMediaLoading] = useState<boolean>(false);
  const [mediaError, setMediaError] = useState<boolean>(false);
  const prevUrlRef = useRef<string | null>(null);

  const isSyncActive = state?.sync_active || false;
  const syncItem = state?.sync_item;
  const windowState = state?.windows?.[windowId];

  const currentItem: MediaItem | undefined = isSyncActive
    ? syncItem
    : windowState?.current_item;

  const offsetSec = isSyncActive ? 0 : windowState?.offset_sec || 0;

  const currentUrl = currentItem?.url || '';

  // Reset media loading indicator when URL changes
  useEffect(() => {
    if (currentUrl !== prevUrlRef.current) {
      prevUrlRef.current = currentUrl;
      setMediaError(false);
      if (videoRef.current && videoRef.current.readyState >= 2) {
        setIsMediaLoading(false);
      } else if (currentUrl) {
        setIsMediaLoading(true);
      }
    }
  }, [currentUrl]);

  // Handle video element playback & force reset to start (currentTime = 0) on sync trigger
  useEffect(() => {
    if (currentItem?.type === 'video' && videoRef.current) {
      const vid = videoRef.current;
      
      // When global sync starts, force video playback to start from beginning (0s) on all windows
      if (isSyncActive) {
        try {
          vid.currentTime = 0;
        } catch {
          // ignore seek restriction before metadata load
        }
      }

      if (vid.readyState >= 2) {
        setIsMediaLoading(false);
      }
      
      vid.play().catch(() => {
        setIsMediaLoading(false);
      });
    }
  }, [currentItem?.id, currentUrl, isSyncActive]);

  return (
    <div className="media-window-card">
      <div className="window-header">
        <span className="window-title">Window {windowId}</span>
        {isSyncActive ? (
          <span className="badge sync-badge">⚡ SYNC MODE</span>
        ) : (
          <span className="badge live-badge">▶ LOOPING</span>
        )}
      </div>

      <div className="media-viewport">
        {loading && !state ? (
          <div className="media-placeholder">Loading playback state...</div>
        ) : error ? (
          <div className="media-placeholder error">API Error: {error}</div>
        ) : !currentItem || !currentUrl ? (
          <div className="media-placeholder blank">Blank Screen (No Media Configured)</div>
        ) : (
          <>
            {isMediaLoading && !mediaError && (
              <div className="media-loading-overlay">
                <div className="loading-spinner"></div>
                <span>Buffering media...</span>
              </div>
            )}

            {mediaError ? (
              <div className="media-placeholder error">
                ⚠️ Media URL unavailable: {currentUrl}
              </div>
            ) : currentItem.type === 'video' ? (
              <video
                ref={videoRef}
                src={currentUrl}
                className="media-content visible"
                autoPlay
                muted
                playsInline
                loop
                preload="auto"
                controls={false}
                onLoadedData={() => setIsMediaLoading(false)}
                onCanPlay={() => setIsMediaLoading(false)}
                onPlay={() => setIsMediaLoading(false)}
                onError={(e) => {
                  const target = e.currentTarget;
                  if (target.error && (target.error.code === 3 || target.error.code === 4)) {
                    setIsMediaLoading(false);
                    setMediaError(true);
                  } else {
                    setIsMediaLoading(false);
                  }
                }}
              />
            ) : (
              <img
                src={currentUrl}
                alt={currentItem.id}
                className="media-content visible"
                onLoad={() => setIsMediaLoading(false)}
                onError={() => {
                  setIsMediaLoading(false);
                  setMediaError(true);
                }}
              />
            )}
          </>
        )}
      </div>

      <div className="window-footer">
        <div className="item-info">
          <strong>Item:</strong> {currentItem?.id || 'N/A'} <span className="type-tag">{currentItem?.type || 'blank'}</span>
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
