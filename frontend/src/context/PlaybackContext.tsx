import React, { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import type { SyncStateResponse } from '../types/media';

interface PlaybackContextType {
  state: SyncStateResponse | null;
  loading: boolean;
  error: string | null;
  refreshState: () => Promise<void>;
  apiBaseUrl: string;
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

const PlaybackContext = createContext<PlaybackContextType | undefined>(undefined);

export const PlaybackProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [state, setState] = useState<SyncStateResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchState = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/state`);
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }
      const data: SyncStateResponse = await res.json();
      setState(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch playback state');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchState();
    const interval = setInterval(fetchState, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <PlaybackContext.Provider
      value={{
        state,
        loading,
        error,
        refreshState: fetchState,
        apiBaseUrl: API_BASE_URL,
      }}
    >
      {children}
    </PlaybackContext.Provider>
  );
};

export const usePlayback = (): PlaybackContextType => {
  const context = useContext(PlaybackContext);
  if (!context) {
    throw new Error('usePlayback must be used within a PlaybackProvider');
  }
  return context;
};
