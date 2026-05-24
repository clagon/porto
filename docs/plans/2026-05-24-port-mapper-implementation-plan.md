# UPnP Port Mapper Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a local UPnP port mapping tool with a Go backend and a Svelte + Vite frontend, packaged as a single desktop-friendly binary that serves its own UI and stores `config.json` beside the binary.

**Architecture:** The Go backend owns discovery (SSDP), control (SOAP), config persistence, HTTP API, and static asset serving. The frontend is a Svelte + Vite app that talks only to the local API. The backend must be developed with TDD and table-driven tests, with testable interfaces for discovery, SOAP, config, and browser-opening logic.

**Tech Stack:** Go, `net/http`, `encoding/xml`, `embed`, `testing`, table-driven tests, Svelte, Vite, TypeScript, JSON config.

---

## Implementation Notes

- Repository root: `/home/clagon/ghq/github.com/clagon/port-mapper`
- `config.json` lives in the **same directory as the binary**.
- Local web server binds to `127.0.0.1` only.
- Discovery runs on startup, plus a manual re-scan button in the UI.
- Backend tests must be table-driven wherever behavior has multiple cases.
- Prefer small interfaces to keep tests isolated:
  - `DiscoveryClient`
  - `SOAPClient`
  - `ConfigStore`
  - `Clock`
  - `BrowserOpener`

---

## Task 1: Initialize the Go backend skeleton and repo layout

**Objective:** Create the backend module, directory structure, and the smallest runnable local server with a health endpoint.

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/port-mapper/main.go`
- Create: `backend/internal/server/server.go`
- Create: `backend/internal/server/routes.go`
- Create: `backend/internal/server/types.go`
- Create: `backend/internal/app/app.go`
- Create: `backend/internal/app/app_test.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/config_test.go`
- Create: `backend/assets/embed.go`
- Create: `README.md`
- Create: `.gitignore`

**Step 1: Write failing test**

Create `backend/internal/app/app_test.go` with a simple table test for server startup behavior:

```go
func TestNewApp(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "ok", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(AppOptions{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./... -v
```

Expected: fail because the package and types do not exist yet.

**Step 3: Write minimal implementation**

Implement the app container, a `GET /api/health` route, and a server listener bound to `127.0.0.1`.

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./... -v
```

Expected: tests pass, and the health handler returns `{ok:true}`.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend README.md .gitignore
git commit -m "feat: initialize backend skeleton"
```

---

## Task 2: Add config loading and persistence beside the binary

**Objective:** Implement `config.json` loading/saving next to the executable, with defaults for first run.

**Files:**
- Create: `backend/internal/config/store.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Create: `backend/internal/config/testdata/` if needed
- Create: `config.json` (example only; runtime file is created beside the binary)

**Step 1: Write failing test**

Add a table test for config resolution:
- no file present → defaults are returned and file can be created
- invalid JSON → error
- valid JSON → settings loaded

Example:

```go
func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "missing file", content: "", wantErr: false},
		{name: "invalid json", content: "{", wantErr: true},
		{name: "valid json", content: `{"auto_discover":true}`, wantErr: false},
	}
	// ...
}
```

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/config -v
```

Expected: fail until the loader exists.

**Step 3: Write minimal implementation**

Implement:
- `Load(path string) (Config, error)`
- `Save(path string, cfg Config) error`
- `DefaultConfig()`
- path resolution based on the binary directory or explicit path injection for tests

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/config -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend/internal/config
git commit -m "feat: add config loading"
```

---

## Task 3: Add validation-first domain models with table-driven tests

**Objective:** Define the core port mapping model and validate user input before any SOAP request is sent.

**Files:**
- Create: `backend/internal/upnp/models.go`
- Create: `backend/internal/upnp/validate.go`
- Create: `backend/internal/upnp/validate_test.go`
- Create: `backend/internal/upnp/models_test.go`

**Step 1: Write failing test**

Create a table-driven validation test with cases like:
- valid TCP mapping
- valid UDP mapping
- invalid protocol
- external port out of range
- missing internal IP
- invalid internal IP
- empty description allowed
- negative / huge lease duration

Example:

```go
func TestValidatePortMapping(t *testing.T) {
	tests := []struct {
		name    string
		in      PortMapping
		wantErr bool
	}{
		{name: "valid tcp", in: PortMapping{Protocol: "TCP", ExternalPort: 8080, InternalIP: "192.168.1.20", InternalPort: 8080}, wantErr: false},
		{name: "invalid protocol", in: PortMapping{Protocol: "ICMP", ExternalPort: 8080, InternalIP: "192.168.1.20", InternalPort: 8080}, wantErr: true},
	}
	// ...
}
```

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestValidatePortMapping -v
```

Expected: fail until validation exists.

**Step 3: Write minimal implementation**

Implement `ValidatePortMapping` plus helper functions for protocol and IP validation.

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestValidatePortMapping -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend/internal/upnp
git commit -m "feat: add port mapping validation"
```

---

## Task 4: Implement SSDP discovery with testable parsing

**Objective:** Discover UPnP IGD candidates on the LAN and resolve control URLs from device descriptions.

**Files:**
- Create: `backend/internal/upnp/discover.go`
- Create: `backend/internal/upnp/discover_test.go`
- Create: `backend/internal/upnp/xml.go`
- Create: `backend/internal/upnp/testdata/rootdesc-*.xml`

**Step 1: Write failing test**

Create table tests for XML parsing and service selection:
- WANIPConnection:1 present
- WANIPConnection:2 present
- WANPPPConnection:1 fallback
- malformed XML
- no matching service

Example:

```go
func TestParseRootDevice(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{name: "wanipconn1", xml: string(testdata.Read("rootdesc-wanipconn1.xml")), wantErr: false},
		{name: "malformed", xml: "<xml", wantErr: true},
	}
	// ...
}
```

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestParseRootDevice -v
```

Expected: fail until parser exists.

**Step 3: Write minimal implementation**

Implement:
- root device XML parsing
- service selection priority
- control URL resolution
- discovery result model

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestParseRootDevice -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend/internal/upnp
git commit -m "feat: add upnp discovery parsing"
```

---

## Task 5: Implement SOAP client and table-driven request/response tests

**Objective:** Build SOAP envelope generation and response parsing for `GetExternalIPAddress`, `AddPortMapping`, and `DeletePortMapping`.

**Files:**
- Create: `backend/internal/upnp/soap.go`
- Create: `backend/internal/upnp/soap_test.go`
- Create: `backend/internal/upnp/testdata/soap-*.xml`

**Step 1: Write failing test**

Create table tests for:
- envelope contains correct action and namespace
- AddPortMapping request includes all fields
- DeletePortMapping request includes protocol and external port
- response parser extracts external IP
- SOAP fault becomes error

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestSOAP -v
```

Expected: fail until SOAP client exists.

**Step 3: Write minimal implementation**

Implement:
- envelope builder
- HTTP POST helper
- response decoding
- SOAP fault parsing

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/upnp -run TestSOAP -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend/internal/upnp
git commit -m "feat: add soap client"
```

---

## Task 6: Build the local HTTP API with handler tests

**Objective:** Expose discovery, status, ports, and settings endpoints for the frontend.

**Files:**
- Create: `backend/internal/server/handlers.go`
- Create: `backend/internal/server/handlers_test.go`
- Modify: `backend/internal/server/routes.go`
- Modify: `backend/internal/server/server.go`
- Create: `backend/internal/auth/token.go`
- Create: `backend/internal/auth/token_test.go`

**Step 1: Write failing test**

Create table-driven handler tests for:
- `GET /api/health`
- `GET /api/status`
- `POST /api/discover`
- `POST /api/ports/open`
- `POST /api/ports/close`
- `GET /api/settings`
- `POST /api/settings`

Each test should assert status code and JSON shape.

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/server -v
```

Expected: fail until handlers exist.

**Step 3: Write minimal implementation**

Implement request decoding, validation errors, and JSON response helpers.

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/server -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend/internal/server backend/internal/auth
git commit -m "feat: add local api server"
```

---

## Task 7: Create the Svelte + Vite frontend skeleton

**Objective:** Build the UI shell and connect it to the local API.

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/app.css`
- Create: `frontend/src/lib/api.ts`
- Create: `frontend/src/lib/types.ts`
- Create: `frontend/src/lib/stores.ts`
- Create: `frontend/src/lib/validate.ts`
- Create: `frontend/src/routes/+layout.svelte`
- Create: `frontend/src/routes/+page.svelte`
- Create: `frontend/src/routes/settings/+page.svelte`
- Create: `frontend/src/components/Header.svelte`
- Create: `frontend/src/components/RouterCard.svelte`
- Create: `frontend/src/components/PortForm.svelte`
- Create: `frontend/src/components/PortList.svelte`
- Create: `frontend/src/components/StatusBadge.svelte`
- Create: `frontend/src/components/Toast.svelte`

**Step 1: Write failing test or smoke check**

For frontend, use a smoke-oriented check:
- build should fail before scaffolding exists
- app should reference API types and endpoint wrappers

**Step 2: Run check to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/frontend
npm run build
```

Expected: fail until scaffold exists.

**Step 3: Write minimal implementation**

Implement a basic dashboard that reads `/api/status` and `/api/ports`, plus a form that calls `/api/ports/open`.

**Step 4: Run check to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/frontend
npm run build
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add frontend
git commit -m "feat: add svelte frontend scaffold"
```

---

## Task 8: Embed the frontend and serve it from the Go binary

**Objective:** Produce a single binary that serves the built frontend assets locally.

**Files:**
- Modify: `backend/assets/embed.go`
- Modify: `backend/internal/server/server.go`
- Modify: `backend/cmd/port-mapper/main.go`
- Create: `backend/internal/server/static.go`
- Create: `backend/internal/server/static_test.go`

**Step 1: Write failing test**

Add a server test that asserts the root path serves the frontend index page and SPA fallback behaves correctly.

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/server -run TestStatic -v
```

Expected: fail until embedding exists.

**Step 3: Write minimal implementation**

Use `go:embed` to bundle frontend build output.

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./internal/server -run TestStatic -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add backend
git commit -m "feat: embed frontend into backend"
```

---

## Task 9: Add startup UX, safe defaults, and docs

**Objective:** Make the app launch into a usable local desktop tool with sane defaults and clear documentation.

**Files:**
- Modify: `README.md`
- Create: `docs/usage.md`
- Create: `docs/security.md`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/server/server.go`

**Step 1: Write failing test**

Add tests for startup behaviors:
- browser opener called only when enabled
- config path resolution uses the binary directory
- localhost bind enforced

**Step 2: Run test to verify failure**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./... -v
```

Expected: fail until startup UX is wired.

**Step 3: Write minimal implementation**

Implement browser opening, startup logging, and safe defaults.

**Step 4: Run test to verify pass**

Run:

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper/backend
go test ./... -v
```

Expected: pass.

**Step 5: Commit**

```bash
cd /home/clagon/ghq/github.com/clagon/port-mapper
git add .
git commit -m "docs: finalize port mapper workflow"
```

---

## Verification Checklist

- [ ] Backend tests are table-driven for validation, discovery, SOAP, and handlers
- [ ] Config is stored beside the binary as `config.json`
- [ ] Local server binds only to `127.0.0.1`
- [ ] Frontend is Svelte + Vite
- [ ] Frontend build output is embedded into the Go binary
- [ ] Discovery works on startup and can be re-run manually
- [ ] Add/DeletePortMapping works end to end
- [ ] UI can display and modify mappings
- [ ] Safe defaults are in place (token, warnings, lease duration)

---

## Recommended Execution Order

1. Task 1: backend skeleton
2. Task 2: config
3. Task 3: validation
4. Task 4: discovery
5. Task 5: SOAP
6. Task 6: API server
7. Task 7: Svelte frontend
8. Task 8: embed and serve frontend
9. Task 9: startup UX and docs

---

## Notes for Implementers

- Keep every table test small and explicit.
- Prefer pure functions for validation and XML parsing.
- Inject dependencies into handlers and services so tests can mock them.
- Do not mix UI logic into the backend packages.
- Do not let the app bind to non-local addresses by default.
- Keep `config.json` format stable and human-editable.
