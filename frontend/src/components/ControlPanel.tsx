import React, { useState, useEffect } from 'react';
import { usePlayback } from '../context/PlaybackContext';
import type { MediaItem } from '../types/media';

export const ControlPanel: React.FC = () => {
  const { apiBaseUrl, refreshState } = usePlayback();
  const [selectedWindow, setSelectedWindow] = useState<string>('A');
  const [type, setType] = useState<'image' | 'video'>('image');
  const [url, setUrl] = useState<string>('');
  const [durationSec, setDurationSec] = useState<number>(8);
  const [playlist, setPlaylist] = useState<MediaItem[]>([]);
  const [statusMsg, setStatusMsg] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const fetchPlaylist = async (winId: string) => {
    try {
      const res = await fetch(`${apiBaseUrl}/api/windows/${winId}/playlist`);
      if (res.ok) {
        const data = await res.json();
        setPlaylist(data.playlist || []);
      }
    } catch (err) {
      console.error('Failed to fetch playlist:', err);
    }
  };

  useEffect(() => {
    fetchPlaylist(selectedWindow);
  }, [selectedWindow, apiBaseUrl]);

  const handleAddMedia = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) {
      setStatusMsg('Please enter a media URL');
      return;
    }

    setIsSubmitting(true);
    setStatusMsg('');
    try {
      const res = await fetch(`${apiBaseUrl}/api/windows/${selectedWindow}/media`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type,
          url: url.trim(),
          duration_sec: Number(durationSec) || 5,
        }),
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Failed to add media');
      }

      setStatusMsg(`Successfully appended media to Window ${selectedWindow}!`);
      setUrl('');
      await fetchPlaylist(selectedWindow);
      await refreshState();
    } catch (err: any) {
      setStatusMsg(`Error: ${err.message}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="panel-card">
      <h2>Control Panel: Add Media</h2>
      <form onSubmit={handleAddMedia} className="control-form">
        <div className="form-group">
          <label>Target Window:</label>
          <select
            value={selectedWindow}
            onChange={(e) => setSelectedWindow(e.target.value)}
          >
            <option value="A">Window A</option>
            <option value="B">Window B</option>
            <option value="C">Window C</option>
          </select>
        </div>

        <div className="form-group">
          <label>Media Type:</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as 'image' | 'video')}
          >
            <option value="image">Image</option>
            <option value="video">Video</option>
          </select>
        </div>

        <div className="form-group">
          <label>Media URL:</label>
          <input
            type="url"
            placeholder="https://picsum.photos/id/100/800/600"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label>Duration (sec):</label>
          <input
            type="number"
            min="1"
            max="300"
            value={durationSec}
            onChange={(e) => setDurationSec(Number(e.target.value))}
            required
          />
        </div>

        <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
          {isSubmitting ? 'Adding...' : 'Add to Playlist'}
        </button>
      </form>

      {statusMsg && <div className="status-banner">{statusMsg}</div>}

      <div className="current-playlist">
        <h3>Window {selectedWindow} Current Playlist ({playlist.length} items)</h3>
        <ul className="playlist-list">
          {playlist.map((item, idx) => (
            <li key={item.id || idx}>
              <span className="item-badge">{item.type}</span>
              <span className="item-id">{item.id}</span>
              <span className="item-dur">{item.duration_sec}s</span>
              <span className="item-url">{item.url}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};
