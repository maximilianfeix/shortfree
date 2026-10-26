# Security

## Reporting

Please report vulnerabilities privately via [GitHub security advisories](https://github.com/maximilianfeix/shortfree/security/advisories/new), not in public issues. You get a reply within a few days.

## Scope

- the `shortfree` CLI and MCP server (input handling, network requests it makes)
- the GitHub Actions workflows (injection, token permissions)
- the website in `site/` (XSS via data files)

Wrong availability results are not security issues; open a "Wrong result" issue instead.

## What shortfree sends where

`shortfree` only talks to the TLD's authoritative name servers (DNS `NS` queries) and the registries' RDAP services listed in `internal/rdap/rdap.go`. It sends no telemetry.
