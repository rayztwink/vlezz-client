# RayFlow Client

RayFlow Client is an open-source desktop proxy client and anti-DPI routing toolkit.

It is designed for users who need a modern local-first desktop client for VLESS/VLESS Reality, subscriptions, sing-box, xray-core, zapret/Flowseal presets, routing rules, diagnostics, logs, and latency monitoring.

RayFlow is closer in role to v2rayN, v2box, Hiddify, Happ, and Nekoray than to a generic admin panel: the user imports proxy links or subscriptions, selects a node, clicks Connect, and the app generates core configs, starts the selected engine safely, and shows status, logs, checks, and route decisions.

> Пользователь сам несет ответственность за соблюдение законодательства своей страны.

## Features

- Cross-platform desktop shell with Tauri and Vue 3
- Go sidecar backend with local HTTP API
- VLESS and VLESS Reality link import
- HTTP/HTTPS subscription import for plain or base64 VLESS links
- sing-box and xray-core config generation
- Local proxy, Windows system proxy, and sing-box TUN connection modes
- Safe ProcessManager for core lifecycle
- SQLite storage
- Zapret / Flowseal preset management
- Direct, Proxy, Zapret, Hybrid, and Smart modes
- DNS, TCP, HTTP, and latency diagnostics
- Logs viewer with secret masking
- Light/dark themes
- Russian and English i18n foundation

## Screenshots

Screenshots will be added as the UI stabilizes.

| Client | Zapret | Diagnostics |
| --- | --- | --- |
| _placeholder_ | _placeholder_ | _placeholder_ |

## Architecture

```text
Tauri + Vue UI
  |
  | HTTP API on 127.0.0.1
  v
Go sidecar: rayflowd
  |
  | only ProcessManager can start processes
  v
sing-box / xray-core / zapret
```

The frontend never starts processes directly. All sensitive system actions go through `rayflowd`, which validates inputs, generates configs, masks secrets in logs, and starts external cores only through the ProcessManager.

## Repository Layout

```text
apps/
  backend/      Go sidecar daemon
  desktop/      Tauri + Vue desktop app
docs/           architecture and security notes
scripts/        helper scripts
.github/        CI workflows
```

## Development Setup

### Backend

```bash
cd apps/backend
go mod tidy
go run ./cmd/rayflowd
```

The backend listens on:

```text
http://127.0.0.1:8787
```

### Desktop

```bash
cd apps/desktop
npm install
npm run dev
```

Set `VITE_RAYFLOW_API_URL` if your backend is not using the default local URL.

## Environment

See `.env.example`.

Core binaries are not bundled in the MVP. Configure absolute paths:

- `RAYFLOW_SING_BOX_PATH`
- `RAYFLOW_XRAY_PATH`
- `RAYFLOW_ZAPRET_PATH`

## API

```text
GET    /
GET    /health

GET    /nodes
POST   /nodes/import
DELETE /nodes/:id
POST   /nodes/:id/check
POST   /nodes/:id/connect
POST   /nodes/disconnect

GET    /subscriptions
POST   /subscriptions
POST   /subscriptions/:id/update
DELETE /subscriptions/:id

GET    /zapret/presets
POST   /zapret/presets/update
POST   /zapret/presets/:id/start
POST   /zapret/stop
GET    /zapret/logs

GET    /routing/rules
POST   /routing/rules
DELETE /routing/rules/:id

POST   /diagnostics/check
POST   /diagnostics/ip-check
GET    /diagnostics/history

GET    /settings
PATCH  /settings

GET    /cores/status
POST   /cores/validate

GET    /logs

GET    /connection/status
POST   /connection/disconnect
GET    /connection/report

GET    /runtime/capabilities

GET    /system-proxy/status
POST   /system-proxy/enable
POST   /system-proxy/disable
```

## MVP Roadmap

### Phase 1

- Backend skeleton
- SQLite
- VLESS parser
- Nodes API
- ProcessManager
- Vue/Tauri shell

### Phase 2

- sing-box integration
- Connect/disconnect
- Logs manager
- Latency checks

### Phase 3

- xray-core support
- Subscriptions
- Server selection

### Phase 4

- zapret integration
- Flowseal presets
- Start/stop presets

### Phase 5

- Hybrid routing
- Smart Mode
- Core updater
- Preset updater

## Security Principles

- No arbitrary shell execution
- Frontend never starts processes
- Backend binds to localhost by default
- External processes are started only through ProcessManager
- UUIDs and proxy links are masked in logs
- Subscription URLs are masked in UI/logs
- Core binary paths must be explicit and absolute
- Core validation runs only explicit version/file checks
- Windows system proxy changes are explicit and restore previous settings on disconnect
- TUN mode is opt-in and requires administrator privileges
- No hidden background traffic

## License

Apache-2.0.
