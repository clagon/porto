# port-mapper

A small local UPnP port-mapping tool.

## Backend

The Go backend lives in `backend/` and serves a local HTTP API on `127.0.0.1`.

### Health endpoint

- `GET /api/health` → `{"ok":true}`

### Run

```bash
cd backend
go run ./cmd/port-mapper
```
