# Security Notes

RayFlow Client should be boringly explicit about sensitive actions.

## Defaults

- Backend binds to `127.0.0.1`.
- API token can be enabled with `RAYFLOW_AUTH_TOKEN`.
- External core binaries are not bundled in the MVP.
- Core paths must be absolute.

## Secret Masking

The logs manager masks:

- full UUID values
- proxy URLs such as `vless://...`
- subscription URLs in UI contexts

## Process Execution

ProcessManager uses:

```go
exec.CommandContext(ctx, binaryPath, args...)
```

It does not invoke a shell.

Core validation uses fixed version/file checks only. It does not accept arbitrary command strings from the UI.

## Windows System Proxy

System proxy changes are explicit. RayFlow stores previous user proxy settings before enabling its proxy and restores them on disconnect or `/system-proxy/disable`.

## TUN Mode

TUN is opt-in, generated for sing-box, and requires administrator privileges. It is not enabled silently.
