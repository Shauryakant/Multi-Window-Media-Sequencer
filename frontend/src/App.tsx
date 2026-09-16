import React from 'react';
import { PlaybackProvider } from './context/PlaybackContext';
import { MediaWindow } from './components/MediaWindow';
import { ControlPanel } from './components/ControlPanel';
import { SyncPanel } from './components/SyncPanel';
import './App.css';

export const AppContent: React.FC = () => {
  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Multi-Window Media Sequencer</h1>
        <p className="app-subtitle">
          Continuous Playback & Dynamic Sync Engine (Go + React + MongoDB)
        </p>
      </header>

      <main className="app-main">
        <section className="windows-grid">
          <MediaWindow windowId="A" />
          <MediaWindow windowId="B" />
          <MediaWindow windowId="C" />
        </section>

        <section className="controls-grid">
          <ControlPanel />
          <SyncPanel />
        </section>
      </main>

      <footer className="app-footer">
        <p>1s Polling Real-time Engine | Pure Elapsed-Time Cycle Math | Resettable cycle_started_at Sync State</p>
      </footer>
    </div>
  );
};

export function App() {
  return (
    <PlaybackProvider>
      <AppContent />
    </PlaybackProvider>
  );
}

export default App;
