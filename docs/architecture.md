# Architecture

RayFlow Client is split into a desktop UI and a local Go sidecar.

```text
apps/desktop
  Tauri + Vue 3 + Pinia + TailwindCSS

apps/backend
  rayflowd Go daemon
```

The UI communicates with the backend through a local HTTP API. The backend owns storage, diagnostics, config generation, subscriptions, logs, and process lifecycle.

## Process Boundary

Only `internal/process.Manager` may start or stop external processes.

Allowed:

```text
UI -> HTTP API -> backend service -> ProcessManager -> core binary
```

Disallowed:

```text
UI -> shell
UI -> core binary
backend -> raw shell string
```

## Modes

- Direct: no proxy or zapret process.
- Proxy: sing-box or xray-core with selected VLESS node.
- Zapret: zapret preset without proxy.
- Hybrid: routing rules combine direct, proxy, and zapret.
- Smart: diagnostics choose the minimal working route.

