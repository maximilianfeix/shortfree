# MCP

`shortfree mcp` is an [MCP](https://modelcontextprotocol.io) server. It lets Claude, Cursor, VS Code or any other MCP client check domains live and search every free 1–4 character domain on **.com .net .org .io .ai .dev** and the other TLDs in [tlds.md](tlds.md).

Ask things like:

- *"Is acme.dev free? And acme.io, acme.app?"*
- *"Find free 3 letter .io domains that contain 'ai'."*
- *"Suggest short domains for my project 'Deep Mind Labs', .com or .dev."*
- *"How many 4 character .de domains are still free?"*
- *"Which 2 character domains are about to drop?"*

No API key, no account. Searches use the public data files from the daily scan. `check_domains` goes straight to the registries.

## Install

You need either Docker or Go. Docker is the easiest.

### Claude Code

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

or without Docker, with Go 1.27+:

```sh
claude mcp add shortfree -- go run github.com/maximilianfeix/shortfree/cmd/shortfree@latest mcp
```

### Claude Desktop

Settings → Developer → Edit Config, then add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "shortfree": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "ghcr.io/maximilianfeix/shortfree"]
    }
  }
}
```

### Cursor

[Add to Cursor](cursor://anysphere.cursor-deeplink/mcp/install?name=shortfree&config=eyJjb21tYW5kIjoiZG9ja2VyIiwiYXJncyI6WyJydW4iLCItaSIsIi0tcm0iLCJnaGNyLmlvL21heGltaWxpYW5mZWl4L3Nob3J0ZnJlZSJdfQ==), or put the same `mcpServers` block as above in `~/.cursor/mcp.json`.

### VS Code

```sh
code --add-mcp '{"name":"shortfree","command":"docker","args":["run","-i","--rm","ghcr.io/maximilianfeix/shortfree"]}'
```

or in `.vscode/mcp.json`:

```json
{
  "servers": {
    "shortfree": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "-i", "--rm", "ghcr.io/maximilianfeix/shortfree"]
    }
  }
}
```

### Binary

Download a binary from the [releases](https://github.com/maximilianfeix/shortfree/releases) (or `go install github.com/maximilianfeix/shortfree/cmd/shortfree@latest`) and use `shortfree mcp` as the command in any of the configs above.

## Tools

All tools are read-only.

### `check_domains`

Live check of up to 50 domains of any length on any TLD. For each domain shortfree asks the TLD's authoritative name servers (no resolver cache) and the registry's RDAP service.

| field | meaning |
|---|---|
| `dns` | `free` (NXDOMAIN at the registry's servers), `taken`, `unknown` |
| `rdap` | `available` (404), `registered`, `dropping` (redemption / pending delete), `error`, `-` (TLD without RDAP in shortfree) |
| `verdict` | `FREE`, `taken`, `dropping soon`, `probably free (…)`, `unknown (…)`, `invalid (…)` |
| `free` | `true` only for `FREE` |
| `premium_likely` | 1–2 characters outside .com/.net/.de: usually reserved or premium-priced, even though RDAP reports them as unregistered |
| `expires`, `epp_status` | from RDAP, when the registry returns them |

RDAP is supported for every scanned TLD ([tlds.md](tlds.md)); other TLDs get a DNS-only verdict. Names below a registry's minimum length get `not registrable (...)`. Each registry is rate-limited, so expect about a second per domain.

### `find_free_domains`

Searches the daily snapshot. Parameters (all optional):

| parameter | |
|---|---|
| `tld` | `io`, `.io` or `io,dev`; default all scanned TLDs |
| `length` | 1–4; default all, shortest first |
| `kind` | `letters`, `digits`, `mixed`, `any` |
| `contains`, `starts_with` | substring filters |
| `include_dropping` | also list registered names in redemption / pending delete |
| `limit` | default 100, max 1000; `total_matches` always has the full count |

Each result has a `status`:

- `rdap_verified`: 1–3 characters, the registry's RDAP said "not registered" at the last scan
- `unverified`: 1–3 characters, not in DNS, RDAP check still pending
- `dns_only`: 4 characters, not in the zone. Never RDAP-checked, so a name can be registered without name servers.
- `dropping`: registered, but in redemption or pending delete (with `expires`)

### `suggest_domains`

Turns a word or brand idea into short (≤ 4 character) names and returns the ones that are free, best first. The heuristic is deterministic. Each candidate gets a score:

| score | pattern | "Deep Mind Labs" → |
|--:|---|---|
| 100 | the word itself, if ≤ 4 characters | |
| 95 | initials | `dml` |
| 90 | a short word of the idea | `deep`, `mind`, `labs` |
| 85 | initials of the first two words | `dm` |
| 80 / 75 / 70 | first 4 / 3 / 2 letters | `deep`, `dee`, `de` |
| 78 / 73 | consonant skeleton (Flicker → `flck`) | `dpmn`, `dpm` |
| 72 / 68 / 66 | word blends 2+2, 1+3, 3+1, 2+1+1 | `demi`, `dmin`, `deem`, `deml` |
| 55–47 | base + digit 1–9 | `dml1`, `dm2` |
| 50 / 48 / 45 | base + `x`, + `hq` / `ly`, `go` / `my` + base | `dmlx`, `dmhq`, `godm` |

Stop words (the, of, and, …) are dropped, accents transliterated and camelCase split. 1–2 character names outside .com/.net/.de lose 15 points because registries usually price them as premium. Ties go to your TLD order, then RDAP-verified before unverified/DNS-only.

### `domain_stats`

Counts per TLD and length from the latest scan: `rdap_verified`, `unverified`, `dropping`, `dns_only_4char`, `valid_names` and `free_total`.

## Notes

- **The snapshot is up to a day old.** Search results come from the daily scan, so confirm a pick with `check_domains`. The assistant is told to do this.
- **Free ≠ cheap.** RDAP can't see premium prices. The registrar shows the real price.
- Data files are cached in memory for one hour. Set `SHORTFREE_DATA_URL` to use another copy of the data (for tests or a mirror). It's the base URL of the `data/` directory and must contain `available.json`, `stats.json` and `4/<tld>.bin`.
- The server sends nothing anywhere except DNS queries to the TLD's name servers, RDAP requests to the registries and requests for the data files on GitHub Pages.
