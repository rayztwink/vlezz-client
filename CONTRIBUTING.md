# Contributing

Thanks for helping build RayFlow Client.

## Development

- Keep frontend and backend separated.
- Do not start system processes from the UI.
- Use the backend ProcessManager for all core lifecycle actions.
- Keep security-sensitive changes small and reviewable.
- Add tests for parser, config generation, diagnostics, and process lifecycle logic.

## Code Style

Backend:

```bash
cd apps/backend
gofmt -w ./cmd ./internal
go test ./...
```

Frontend:

```bash
cd apps/desktop
npm run build
```

## Pull Requests

Please include:

- Problem statement
- Summary of changes
- Security considerations
- Manual test notes
- Screenshots for UI changes

