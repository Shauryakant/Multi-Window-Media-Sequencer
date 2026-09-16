# Multi-Window Media Sequencer with Sync Playback

A full-stack application where multiple independent display windows (**A**, **B**, **C**) continuously loop through their configured media playlists (images and videos), support dynamic playlist additions, and support a global **Sync** action that temporarily forces all windows to display a single selected media item simultaneously before seamlessly resuming their individual playback sequence exactly where they were interrupted.

---

## 🛠 Tech Stack

- **Backend**: Golang (`net/http`, custom router, `sync.Mutex`)
- **Frontend**: React (TypeScript + Vite)
- **Database**: MongoDB (Official `mongo-driver/v2` with fallback `MemoryStore` for local development/testing)
- **Real-Time Engine**: HTTP Polling (1s client interval)

---

## 🚀 Quick Start (Local Run)

### Prerequisites
- Go 1.22+
- Node.js 18+ / npm

### 1. Run Backend
```bash
cd backend
go run main.go
```
The Go server starts on `http://localhost:8080`.
*On startup, if no MongoDB URI is specified, the backend automatically uses an in-memory store and seeds Windows A, B, and C with sample image and video playlists.*

To connect to a live MongoDB instance, set the `MONGO_URI` environment variable:
```bash
export MONGO_URI="mongodb+srv://<user>:<password>@cluster.mongodb.net"
export PORT=8080
export SYNC_DURATION_SECONDS=5
go run main.go
```

### 2. Run Frontend
In a separate terminal:
```bash
cd frontend
npm install
npm run dev
```
The React frontend starts on `http://localhost:5173`.

---

## ⚙️ Environment Variables

| Variable | Scope | Description | Default |
|---|---|---|---|
| `PORT` | Backend | HTTP server listening port | `8080` |
| `MONGO_URI` | Backend | MongoDB connection connection string | *(Uses in-memory fallback if empty)* |
| `MONGO_DB_NAME` | Backend | MongoDB database name | `media_sequencer` |
| `SYNC_DURATION_SECONDS` | Backend | Default duration in seconds for global sync override | `5` |
| `VITE_API_BASE_URL` | Frontend | Base URL pointing to the Go backend API | `http://localhost:8080` |

---

## 💡 How Sync Works (Plain Language & Worked Example)

Playback position for any window is **never stored as a mutable index**. Instead, it is computed purely as a function of elapsed time:
$$\text{elapsed} = \text{now} - \text{cycle\_started\_at}$$
$$\text{cycle\_offset} = \text{elapsed} \pmod{\text{total\_playlist\_duration}}$$

### Worked Example:
1. **Normal Playback**:
   - Window A has playlist: `[Image1 (10s), Video2 (15s), Image3 (5s)]`. Total duration = `30s`.
   - `cycle_started_at` = `10:00:00`.
   - At `10:00:22` (elapsed `22s`), cycle offset is `22s`. Window A is playing **Video2** (index 1) at **offset 12s** (`22s - 10s`).

2. **Triggering Sync**:
   - At `10:00:22`, user sends `POST /api/sync` for `MediaX` (duration 5 seconds).
   - Backend immediately **snapshots** Window A's current position: `{item_index: 1, offset_sec: 12.0s}`.
   - `SyncState.Active` becomes `true`.
   - `GET /api/state` returns `sync_active: true` and `sync_item: MediaX`.
   - All 3 windows immediately display `MediaX`.

3. **Auto-Resuming on Sync End**:
   - At `10:00:27` (5 seconds later), `time.AfterFunc` fires `EndSync()`.
   - To resume Window A from `{item_index: 1, offset_sec: 12.0s}` at time `now = 10:00:27`:
     - Cumulative duration up to index 1 = `10s`.
     - Target offset in loop = `10s + 12s = 22s`.
     - Backend calculates new `cycle_started_at = 10:00:27 - 22s = 10:00:05`.
   - Backend updates Window A's `cycle_started_at` to `10:00:05` in MongoDB.
   - Next poll at `10:00:28`: elapsed = `10:00:28 - 10:00:05 = 23s`. Cycle offset `23s` lands back on **Video2 at offset 13s** (`23s - 10s`)!
   - Window A resumes smoothly as if sync paused time for the window.

---

## 📌 Design Assumptions & Decisions

1. **HTTP Polling (1s) vs WebSockets**:
   HTTP polling at a 1-second interval was selected over WebSockets for architecture simplicity and stateless connection reliability across container hosts. It introduces a maximum visible sync latency of $\le 1\text{s}$, which is an acceptable design trade-off for multi-window digital signage.

2. **Ephemeral Sync State**:
   Sync state (`SyncState` struct) is intentionally held in an in-memory, mutex-protected data structure. Sync overrides are temporary transient events and do not need to persist across server restarts, unlike window playlists and cycle start times which are persisted in MongoDB.

3. **Append-Only Playlist Additions**:
   New media items appended via `POST /api/windows/:id/media` are pushed to the end of the `playlist` array.

4. **Video Playback Resume**:
   Upon sync termination, videos resume from their stored item index in the playlist sequence. Video elements begin playback from 0 rather than performing sub-second frame seeking over network polling.

5. **5-Hour Cycle Interpretation**:
   The 5-hour cycle requirement is modeled as a fixed-length looping window timeline starting from `cycle_started_at`, allowing continuous sequence repeating without arbitrary blank cutoffs.

---

## 📡 REST API Documentation

### 1. Health Check
`GET /health`
```json
{
  "status": "ok"
}
```

### 2. Poll State
`GET /api/state`
Returns active sync status or individual window playback positions.
```json
{
  "sync_active": false,
  "sync_item": null,
  "sync_ends_at": null,
  "windows": {
    "A": {
      "current_item": {
        "id": "m1_a",
        "type": "image",
        "url": "https://picsum.photos/id/10/800/600",
        "duration_sec": 8
      },
      "offset_sec": 3.2
    },
    "B": {
      "current_item": {
        "id": "m1_b",
        "type": "video",
        "url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4",
        "duration_sec": 15
      },
      "offset_sec": 1.1
    },
    "C": {
      "current_item": {
        "id": "m1_c",
        "type": "image",
        "url": "https://picsum.photos/id/50/800/600",
        "duration_sec": 5
      },
      "offset_sec": 0.5
    }
  }
}
```

### 3. Get Window Playlist
`GET /api/windows/:id/playlist`
```json
{
  "_id": "A",
  "playlist": [
    { "id": "m1_a", "type": "image", "url": "https://picsum.photos/id/10/800/600", "duration_sec": 8 },
    { "id": "m2_a", "type": "video", "url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", "duration_sec": 15 }
  ],
  "cycle_started_at": "2026-09-17T10:00:00Z"
}
```

### 4. Add Media to Window
`POST /api/windows/:id/media`
**Request Body**:
```json
{
  "type": "image",
  "url": "https://picsum.photos/id/100/800/600",
  "duration_sec": 10
}
```

### 5. Trigger Sync
`POST /api/sync`
**Request Body (By Existing Item ID)**:
```json
{
  "id": "m2_a",
  "duration_sec": 5
}
```
**Request Body (By Custom Media URL)**:
```json
{
  "type": "image",
  "url": "https://picsum.photos/id/200/800/600",
  "duration_sec": 5
}
```

---

## 🚢 Deployment Instructions

### Docker Deployment (Backend)
The backend includes a multi-stage `Dockerfile` generating a container image under `<50MB`.

Build and run container:
```bash
docker build -t media-sequencer-backend .
docker run -p 8080:8080 -e PORT=8080 -e SYNC_DURATION_SECONDS=5 media-sequencer-backend
```

### Deployment Services
- **Backend (Render / Fly.io)**: Connect repository to Render Web Service, select Docker runtime, and configure `PORT` and `MONGO_URI`.
- **Frontend (Vercel / Netlify)**: Deploy `frontend/` directory to Vercel, setting `VITE_API_BASE_URL` to your deployed backend URL.

---

## 🧪 Running Unit Tests

To run the backend test suite:
```bash
cd backend
go test -v ./...
```
