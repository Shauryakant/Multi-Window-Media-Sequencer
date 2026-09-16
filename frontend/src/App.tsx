import React, { useState } from 'react';
import { PlaybackProvider } from './context/PlaybackContext';
import { MediaWindow } from './components/MediaWindow';
import { ControlPanel } from './components/ControlPanel';
import { SyncPanel } from './components/SyncPanel';
import './App.css';

export const AppContent: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'control' | 'sync'>('control');

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Multi-Window Media Sequencer</h1>
        <p className="app-subtitle">
          Continuous Playback & Dynamic Sync Engine (Go + React + MongoDB)
        </p>
      </header>

      <main className="app-main">
        {/* Flexible wrapping horizontal row for Windows A, B, C */}
        <section className="windows-flex-container">
          <MediaWindow windowId="A" />
          <MediaWindow windowId="B" />
          <MediaWindow windowId="C" />
        </section>
      </main>

      {/* Downside fixed tabbed control dock */}
      <div className="bottom-control-dock">
        <div className="tab-bar">
          <button
            type="button"
            className={`tab-button ${activeTab === 'control' ? 'active' : ''}`}
            onClick={() => setActiveTab('control')}
          >
            🛠️ Control Panel (Add Media)
          </button>
          <button
            type="button"
            className={`tab-button ${activeTab === 'sync' ? 'active' : ''}`}
            onClick={() => setActiveTab('sync')}
          >
            ⚡ Global Sync Panel
          </button>
        </div>

        <div className="tab-content">
          {activeTab === 'control' ? <ControlPanel /> : <SyncPanel />}
        </div>
      </div>
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
