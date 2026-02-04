# KirpichMessanger Desktop

Lightweight desktop client built with Tauri 1.5+ and SvelteKit 2+.

## Requirements
- Node.js 18+
- Rust 1.70+
- Tauri CLI (`cargo install tauri-cli`)

## Development
```bash
npm install
npm run dev
```

In another terminal start Tauri:
```bash
npm run tauri dev
```

## Build
```bash
npm run build
npm run tauri build
```

## Architecture
- `src-tauri/` Rust backend with IPC commands and system integration
- `src/` SvelteKit UI with chat, messaging, profile, and settings modules
- `static/` assets (icons, media placeholders)

## Backend Commands
- `login` authentication
- `send_message` send chat messages
- `upload_media` upload attachments
- `get_chats` fetch chat list
- `handle_notifications` native notifications

## Security
- CSP defined in `tauri.conf.json`
- IPC allowlist enabled
- File and network access scoped to API endpoints

## Troubleshooting
- Replace placeholder icons in `src-tauri/icons/` with real assets before packaging.
- Ensure the API base URL is reachable for HTTP and WebSocket connections.
