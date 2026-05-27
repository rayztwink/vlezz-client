# Local API

Base URL:

```text
http://127.0.0.1:8787
```

If `RAYFLOW_AUTH_TOKEN` is set, requests must include:

```text
Authorization: Bearer <token>
```

## Health

```text
GET /
GET /health
```

## Nodes

```text
GET    /nodes
POST   /nodes/import
DELETE /nodes/:id
POST   /nodes/:id/check
POST   /nodes/:id/connect
POST   /nodes/disconnect
```

## Subscriptions

```text
GET    /subscriptions
POST   /subscriptions
POST   /subscriptions/:id/update
DELETE /subscriptions/:id
```

`POST /subscriptions/:id/update` fetches the subscription by explicit user action, imports new VLESS links, skips duplicates, and returns import statistics.

## Zapret

```text
GET  /zapret/presets
POST /zapret/presets/update
POST /zapret/presets/:id/start
POST /zapret/stop
GET  /zapret/logs
```

## Routing

```text
GET    /routing/rules
POST   /routing/rules
DELETE /routing/rules/:id
```

## Diagnostics

```text
POST /diagnostics/check
GET  /diagnostics/history
```

## Settings

```text
GET   /settings
PATCH /settings
```

## Cores

```text
GET  /cores/status
POST /cores/validate
```

`/cores/validate` accepts:

```json
{
  "core": "sing-box",
  "path": "C:\\Tools\\sing-box\\sing-box.exe"
}
```

## Logs

```text
GET /logs?source=&limit=200
```

## Connection

```text
GET  /connection/status
POST /connection/disconnect
GET  /connection/report
```

`POST /nodes/:id/connect` accepts optional connection choices:

```json
{
  "core": "sing-box",
  "networkMode": "local_proxy"
}
```

Supported network modes:

- `local_proxy`
- `system_proxy`
- `tun`

## System Proxy

```text
GET  /system-proxy/status
POST /system-proxy/enable
POST /system-proxy/disable
```

System proxy control is implemented for Windows through registry APIs and WinInet notification, not shell commands.
