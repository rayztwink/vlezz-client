# Security Policy

RayFlow Client is local-first networking software. Security boundaries matter.

## Supported Versions

The project is pre-1.0. Security fixes target the main branch until stable releases begin.

## Reporting a Vulnerability

Please do not open public issues for sensitive vulnerabilities. Use private disclosure channels once the project repository is published.

## Security Rules

- UI must never execute processes directly.
- No raw shell command execution.
- ProcessManager must use explicit executable and args.
- Core binary paths must be absolute and validated.
- Do not log full UUIDs, proxy links, or subscription URLs.
- Do not add hidden background traffic.
- All user-impacting network actions must be explicit.

## Legal Notice

Пользователь сам несет ответственность за соблюдение законодательства своей страны.

