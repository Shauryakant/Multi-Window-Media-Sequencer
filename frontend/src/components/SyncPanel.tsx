import React, { useState, useEffect } from 'react';
import { usePlayback } from '../context/PlaybackContext';

export const SyncPanel: React.FC = () => {
  const { state, apiBaseUrl, refreshState } = usePlayback();
  const [syncMode, setSyncMode] = useState<'existing' | 'custom'>('existing');
  const [selectedMediaId, setSelectedMediaId] = useState<string>('');
  const [customType, setCustomType] = useState<'image' | 'video'>('image');
  const [customUrl, setCustomUrl] = useState<string>('');
  const [syncDuration, setSyncDuration] = useState<number>(5);
  const [statusMsg, setStatusMsg] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [availableMediaIds, setAvailableMediaIds] = useState<{ id: string; windowId: string; type: string; url: string }[]>([]);

  // Gather available media IDs across windows A, B, C for quick sync dropdown selection
  useEffect(() => {
    const fetchAllPlaylists = async () => {
      const windows = ['A', 'B', 'C'];
      const items: { id: string; windowId: string; type: string; url: string }[] = [];
      for (const w of windows) {
        try {
          const res = await fetch(`${apiBaseUrl}/api/windows/${w}/playlist`);
          if (res.ok) {
            const data = await res.json();
            if (data.playlist) {
              for (const m of data.playlist) {
                items.push({ id: m.id, windowId: w, type: m.type, url: m.url });
              }
            }
          }
        } catch (err) {
          console.error(`Failed to fetch playlist for ${w}:`, err);
        }
      }
      setAvailableMediaIds(items);
      if (items.length > 0 && !selectedMediaId) {
        setSelectedMediaId(items[0].id);
      }
    };

    fetchAllPlaylists();
  }, [apiBaseUrl, state?.windows]);

  const handleTriggerSync = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setStatusMsg('');

    try {
      let body: any = { duration_sec: Number(syncDuration) || 5 };
      if (syncMode === 'existing') {
        if (!selectedMediaId) {
          throw new Error('Please select a media item to sync');
        }
        body.id = selectedMediaId;
      } else {
        if (!customUrl.trim()) {
          throw new Error('Please enter a custom media URL');
        }
        body.type = customType;
        body.url = customUrl.trim();
      }

      const res = await fetch(`${apiBaseUrl}/api/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Failed to trigger sync');
      }

      setStatusMsg('Sync triggered successfully! All windows are now playing synced item.');
      await refreshState();
    } catch (err: any) {
      setStatusMsg(`Error: ${err.message}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const isSyncActive = state?.sync_active || false;
  const syncEndsAt = state?.sync_ends_at;

  return (
    <div className="panel-card sync-panel">
      <h2>Sync Panel: Global Media Override</h2>

      {isSyncActive && (
        <div className="sync-active-banner">
          ⚠️ <strong>GLOBAL SYNC IN PROGRESS</strong>
          <div>Synced Item: {state?.sync_item?.id || state?.sync_item?.url}</div>
          {syncEndsAt && <div>Resuming normal loop at: {new Date(syncEndsAt).toLocaleTimeString()}</div>}
        </div>
      )}

      <form onSubmit={handleTriggerSync} className="control-form">
        <div className="form-group">
          <label>Sync Source:</label>
          <div className="radio-group">
            <label>
              <input
                type="radio"
                name="syncMode"
                value="existing"
                checked={syncMode === 'existing'}
                onChange={() => setSyncMode('existing')}
              />
              Existing Playlist Item
            </label>
            <label>
              <input
                type="radio"
                name="syncMode"
                value="custom"
                checked={syncMode === 'custom'}
                onChange={() => setSyncMode('custom')}
              />
              Custom Media URL
            </label>
          </div>
        </div>

        {syncMode === 'existing' ? (
          <div className="form-group">
            <label>Select Media Item ID:</label>
            <select
              value={selectedMediaId}
              onChange={(e) => setSelectedMediaId(e.target.value)}
            >
              {availableMediaIds.map((item) => (
                <option key={item.id + '_' + item.windowId} value={item.id}>
                  [{item.windowId}] {item.id} ({item.type})
                </option>
              ))}
            </select>
          </div>
        ) : (
          <>
            <div className="form-group">
              <label>Custom Media Type:</label>
              <select
                value={customType}
                onChange={(e) => setCustomType(e.target.value as 'image' | 'video')}
              >
                <option value="image">Image</option>
                <option value="video">Video</option>
              </select>
            </div>
            <div className="form-group">
              <label>Custom Media URL:</label>
              <input
                type="url"
                placeholder="https://picsum.photos/id/200/800/600"
                value={customUrl}
                onChange={(e) => setCustomUrl(e.target.value)}
                required
              />
            </div>
          </>
        )}

        <div className="form-group">
          <label>Sync Duration (seconds):</label>
          <input
            type="number"
            min="1"
            max="60"
            value={syncDuration}
            onChange={(e) => setSyncDuration(Number(e.target.value))}
            required
          />
        </div>

        <button type="submit" className="btn btn-warning" disabled={isSubmitting}>
          {isSubmitting ? 'Triggering...' : 'Sync Now across Windows (A, B, C)'}
        </button>
      </form>

      {statusMsg && <div className="status-banner">{statusMsg}</div>}
    </div>
  );
};
