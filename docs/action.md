# GitHub Action: shortfree domain watch

Checks a list of domains live, the same way `shortfree check` does: an `NS` query at the TLD's authoritative name servers, then the registry's RDAP service. Use it to watch domains you want and get told when one becomes free, or to make sure the domains you own stay registered.

The action builds shortfree from its own source at the ref you use (`@main`, a tag or a commit SHA). It needs Go, which it installs with [actions/setup-go](https://github.com/actions/setup-go) (module and build cache included), plus `bash` and `jq`, which every GitHub-hosted runner has.

## Inputs

| Input | Default | |
|---|---|---|
| `domains` | (required) | Domains separated by newlines, spaces or commas. Any length, any TLD. |
| `fail-when` | `never` | `free`: fail the step if any domain is free. `taken`: fail if any domain is registered (including dropping). `never`: only report. |

## Outputs

| Output | Example | |
|---|---|---|
| `free` | `["acme.dev"]` | JSON array of the domains whose verdict is `FREE`, i.e. the registry's RDAP service returned `404`. `[]` if none. |
| `taken` | `["acme.io"]` | JSON array of the registered domains (verdict `taken` or `dropping soon`). |
| `results` | `[{"domain":"acme.dev","dns":"free","rdap":"available","verdict":"FREE","free":true,"premium_likely":false}]` | One object per domain, the output of `shortfree check -json`. |

The step summary shows a table with the verdict, DNS and RDAP status and, for dropping names, the expiry date. Invalid domains and failed lookups are reported as warnings and are in neither `free` nor `taken`.

`free` only contains names the registry confirmed. On TLDs without RDAP support in shortfree (anything not listed in [tlds.md](tlds.md)) the verdict is at best `probably free (no RDAP for this TLD)`, which you find in `results`.

A free name can still be reserved or premium-priced; `premium_likely` is set for 1–2 character names outside .com, .net and .de. Names shorter than a registry's minimum length (e.g. 1–3 characters on .ly) get the verdict `not registrable (...)`. Check the price at your registrar.

## Example: open an issue when a watched domain becomes free

`.github/workflows/domain-watch.yml`:

```yaml
name: domain watch

on:
  schedule:
    - cron: "0 7 * * *"   # daily at 07:00 UTC
  workflow_dispatch:

permissions:
  issues: write

jobs:
  watch:
    runs-on: ubuntu-latest
    steps:
      - id: watch
        uses: maximilianfeix/shortfree@main   # or pin a tag / commit SHA
        with:
          domains: |
            acme.dev
            acme.io
            acme.app

      - name: Open an issue per free domain
        if: steps.watch.outputs.free != '[]'
        env:
          GH_TOKEN: ${{ github.token }}
          FREE: ${{ steps.watch.outputs.free }}
        run: |
          for domain in $(jq -r '.[]' <<< "$FREE"); do
            title="$domain is free"
            open=$(gh issue list -R "$GITHUB_REPOSITORY" --state open --search "\"$title\" in:title" --json number --jq length)
            if [ "$open" = 0 ]; then
              gh issue create -R "$GITHUB_REPOSITORY" --title "$title" \
                --body "shortfree found \`$domain\` unregistered at $(date -u '+%Y-%m-%d %H:%M UTC'). The registry's RDAP service returned 404. Check the price at your registrar before you buy."
            fi
          done
```

The job needs no checkout. `GITHUB_TOKEN` with `issues: write` is enough. The search keeps it from opening a second issue for the same domain while the first is still open.

## Example: fail when a domain you own lapses

```yaml
- uses: maximilianfeix/shortfree@main
  with:
    domains: example.com example.dev
    fail-when: free
```

A failed scheduled run sends the usual GitHub notification email.

## Rate limits

Each registry's RDAP service is rate-limited inside shortfree (a few requests per second per registry), so expect about a second per domain. Keep watch lists short and run them once a day.
