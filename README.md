<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="brand/banner-dark.svg">
      <img src="brand/banner-light.svg" alt="shortfree – free 1–4 character domains, rescanned daily" width="800">
    </picture>
  </a>
</p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/"><b>Search the list →</b></a> ·
  <a href="https://github.com/maximilianfeix/shortfree/releases/tag/data">Download all lists</a> ·
  <a href="#how-it-works">How it works</a>
</p>

<p align="center">
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml/badge.svg" alt="daily scan"></a>
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/code-MIT-315BEE" alt="MIT"></a>
  <img src="https://img.shields.io/badge/data-CC0-69E0AA" alt="data CC0">
</p>

**shortfree** lists every unregistered 1–4 character domain on **.com .net .de .ai .io .dev .app .tools**. A GitHub Action sweeps all ~14 million candidates against the registries' own name servers and confirms the short ones via RDAP every day, so the list never goes stale.

- **Verified, not guessed.** 1–3 character names are confirmed by the registry's RDAP service (`404` = not registered), not by a third-party WHOIS mirror.
- **Complete.** Every valid label: a–z, 0–9 and inner hyphens. 49,284 names per TLD up to 3 characters, 1.8 million more at 4.
- **Dropping soon.** Registered names in redemption or pending delete are listed separately.
- **Plain data.** JSON and text files in this repo and in the [data release](https://github.com/maximilianfeix/shortfree/releases/tag/data), CC0. No API key, no tracking.

## Current numbers

<!-- shortfree:start -->
_The first scan has not run yet._
<!-- shortfree:end -->

## Use the data

| File | Content |
|---|---|
| [`data/available.json`](data/available.json) | per TLD: `available` (RDAP-verified), `dropping` (with expiry), `unverified` (not in DNS, RDAP pending) for 1–3 characters |
| [`data/txt/<tld>.txt`](data/txt) | verified free 1–3 character domains, one per line |
| [`data/stats.json`](data/stats.json) | counts per TLD and length |
| [`data/changes.json`](data/changes.json) | names freed or taken in the last 30 days |
| [release `data`](https://github.com/maximilianfeix/shortfree/releases/tag/data) | `four-<tld>.txt.gz` (all 4 character names not in the zone), bitmaps, scanner state |

```sh
# every free 2 character .dev domain
curl -s https://raw.githubusercontent.com/maximilianfeix/shortfree/main/data/available.json \
  | jq -r '.tlds.dev.available[] | select(length==2) + ".dev"'

# all free 4 letter .io names
curl -sL https://github.com/maximilianfeix/shortfree/releases/download/data/four-io.txt.gz \
  | gunzip | grep -E '^[a-z]{4}\.io$'
```

## Check a domain yourself

```sh
go install github.com/maximilianfeix/shortfree/cmd/shortfree@latest
shortfree check q7.tools x4k.app abc.de
```

```
DOMAIN    DNS    RDAP        VERDICT
q7.tools  free   available   FREE
x4k.app   taken  registered  taken
abc.de    taken  registered  taken
```

## How it works

1. **Sweep.** For every label, `shortfree` sends an `NS` query straight to the TLD's authoritative name servers (no recursive resolver, no cache). `NXDOMAIN` means the name is not in the zone. This takes minutes for 400,000 queries.
2. **Verify.** Not being in the zone is not the same as being free: registered names on hold, without name servers or in redemption are missing too. Every 1–3 character candidate is therefore looked up in the registry's RDAP service (Verisign, DENIC, Identity Digital, Google Registry). `404` = not registered; `pending delete` / `redemption period` = dropping.
3. **Repeat.** [`scan.yml`](.github/workflows/scan.yml) runs daily at 03:17 UTC: all 1–3 character names, plus one of seven shards of the 4 character space per TLD, so each 4 character name is rechecked weekly. Names whose DNS status changes are re-verified the same day; the rest are re-verified at least every 10 days.

Each registry gets its own rate limit (1.5–6 requests/s for RDAP) so the scan stays well within fair use.

### Caveats

- **Free ≠ cheap.** Registries reserve some names or sell them at premium prices (common for 1–2 characters on new TLDs). RDAP does not expose that; your registrar shows the final price.
- **4 characters are DNS-only.** Registered names without name servers can appear in the 4 character lists. Check one with `shortfree check` before you get attached.
- **Snapshots age.** A name can be registered minutes after the scan.

## Development

```sh
go test ./...
go run ./cmd/shortfree scan -tlds tools -rdap-time 1m   # writes state/state.json.gz
go run ./cmd/shortfree export -tlds tools -readme ''      # writes data/
```

The site in [`site/`](site) is plain HTML/CSS/JS and reads the same data files. See [CONTRIBUTING.md](CONTRIBUTING.md) to add a TLD.

## License

Code: [MIT](LICENSE). Data: [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). Registry data comes from public DNS and RDAP services.
