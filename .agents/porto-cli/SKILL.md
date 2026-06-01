---
name: porto-cli
description: Use Porto's local CLI to open, close, or inspect UPnP router port mappings for this machine.
---

# Porto CLI

Use this skill when an AI agent needs to expose a local game server, dev server, or other local TCP/UDP service through the user's UPnP-capable router using Porto.

## Preconditions

- Run commands from the `backend` directory in this repository, or use an installed `porto` binary if one is available.
- Porto only talks to the local network router through UPnP. It does not configure cloud firewalls, OS firewalls, or the target local service.
- Opening a port exposes a service on the user's network to the internet. Ask for explicit user intent before opening a new public port unless the user already requested it in the current task.

## Commands

During development, prefer:

```bash
cd backend
go run ./cmd/porto open 25565
```

With an installed binary, prefer:

```bash
porto open 25565
```

Open a TCP port:

```bash
go run ./cmd/porto open 25565
```

Open a UDP port:

```bash
go run ./cmd/porto open --protocol udp 19132
```

Map a different internal port:

```bash
go run ./cmd/porto open --internal-port 3000 --description "Local web app" 443
```

Close a port:

```bash
go run ./cmd/porto close 25565
```

Close a UDP port:

```bash
go run ./cmd/porto close --protocol udp 19132
```

Show router and mapping status:

```bash
go run ./cmd/porto status
```

## Options

- `--protocol tcp|udp`: selects the mapping protocol. Defaults to `tcp`.
- `--internal-ip IP`: uses a specific LAN IP. Omit this unless the user asked for a specific host; Porto auto-detects this machine's LAN IP.
- `--internal-port PORT`: maps the public port to a different local port. Defaults to the public port.
- `--description TEXT`: labels the mapping on routers that show descriptions. Defaults to `Porto CLI`.
- `--lease SECONDS`: requests a limited lease. `0` means permanent until closed or router cleanup.
- `--config PATH`: use a specific Porto config file.

## Agent Workflow

1. Identify the local service port and protocol. For Minecraft Java use TCP `25565`; for Minecraft Bedrock use UDP `19132`.
2. Confirm the service is running locally when possible.
3. Open the mapping with `porto open` or `go run ./cmd/porto open`.
4. Report the opened external port and protocol to the user.
5. When the task ends or the user asks to stop sharing, close the mapping with `porto close`.

## Failure Handling

- If discovery reports no router or no gateway, tell the user their router may not support UPnP, UPnP may be disabled, or the machine may be on a network segment where the router is unreachable.
- If opening fails after discovery, surface the exact CLI error. Common causes are router policy, conflicting existing mappings, invalid ports, or local firewall rules.
- Do not retry with a different public port unless the user approves or the requested workflow explicitly allows choosing an available alternative.
