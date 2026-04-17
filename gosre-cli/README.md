# gosre-cli

SRE platform CLI — HTTP, TCP, DNS and TLS checks from the terminal.

## Installation

```bash
go install github.com/gosre/gosre-cli@latest
```

Pre-built binaries for Linux, macOS and Windows are available at
[github.com/viniciushammett/gosre/releases](https://github.com/viniciushammett/gosre/releases).

## Usage

### Check commands

**HTTP**
```bash
$ gosre check http --address https://example.com
TIMESTAMP  TARGET               STATUS  DURATION  ERROR
15:04:05   https://example.com  ok      212ms
```

**TCP**
```bash
$ gosre check tcp --address example.com:80
TIMESTAMP  TARGET          STATUS  DURATION  ERROR
15:04:05   example.com:80  ok      18ms
```

**DNS**
```bash
$ gosre check dns --address example.com --record-type A
TIMESTAMP  TARGET       STATUS  DURATION  ERROR
15:04:05   example.com  ok      9ms
```

**TLS**
```bash
$ gosre check tls --address example.com:443 --expiry-days 30
TIMESTAMP  TARGET              STATUS  DURATION  ERROR
15:04:05   example.com:443     ok      94ms
```

**JSON output**
```bash
$ gosre check http --address https://api.example.com --output json
[{"id":"...","check_id":"cli","target_id":"https://api.example.com","status":"ok","duration_ms":214000000,...}]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--quiet` | `false` | Suppress output; use exit code only |
| `--timeout` | `10s` | Check timeout (e.g. `5s`, `30s`) |

These flags are global and apply to all check subcommands.

### Config file (`~/.gosre.yaml`)

```yaml
defaults:
  timeout: 10s
  output: table

targets:
  - name: my-api
    type: http
    address: https://api.mycompany.com/healthz
    tags: [production, api]

  - name: db-primary
    type: tcp
    address: db.mycompany.com:5432
    tags: [production, database]

  - name: main-cert
    type: tls
    address: mycompany.com:443
    tags: [production, tls]
```

Use `--target-name` to run a check against a named target from the config file:

```bash
gosre check http --target-name my-api
gosre check tcp --target-name db-primary
gosre check tls --target-name main-cert
```

### Targets

List all targets configured in `~/.gosre.yaml`:

```bash
$ gosre targets list
NAME        TYPE  ADDRESS                              TAGS
my-api      http  https://api.mycompany.com/healthz   production,api
db-primary  tcp   db.mycompany.com:5432               production,database
main-cert   tls   mycompany.com:443                   production,tls
```

## Output formats

**table** (default) — human-readable aligned columns via `text/tabwriter`:
```
TIMESTAMP  TARGET  STATUS  DURATION  ERROR
```

**json** — newline-delimited JSON array, suitable for piping to `jq`:
```bash
gosre check http --address https://example.com --output json | jq '.[0].status'
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Check passed (`ok`) |
| `1` | Check failed, timed out, or an error occurred |

## License

Apache 2.0 — see [LICENSE](../LICENSE).
