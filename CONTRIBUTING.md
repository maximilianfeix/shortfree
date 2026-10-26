# Contributing

Thanks for helping. Issues and pull requests are welcome.

## Add a TLD

1. Find the registry's RDAP base URL in the [IANA bootstrap file](https://data.iana.org/rdap/dns.json) (or the registry's docs if it is missing there).
2. Add it to `Endpoints` in [`internal/rdap/rdap.go`](internal/rdap/rdap.go) and, if the host is new, a conservative rate (keyed by host name) to `DefaultRate`.
3. Add the TLD to `Default` in [`internal/tlds/tlds.go`](internal/tlds/tlds.go), plus a `Policy` if the registry has a minimum length or `nic.<tld>` is not registered. Both workflows pick the list up from there; a new TLD bootstraps itself in the daily scan.
4. Check that free and taken names are told apart:
   ```sh
   go run ./cmd/shortfree check q-z7.<tld> nic.<tld>
   ```
5. Add a row to [`docs/tlds.md`](docs/tlds.md) and open a pull request with the output.

A new TLD adds ~49k DNS queries and up to a few thousand RDAP lookups per day, so it should have a real audience.

## Ground rules

- Stay polite to registries: no rate increases without a measurement in the PR.
- `gofmt`, `go vet ./...` and `go test ./...` must pass.
- Do not commit changes to `data/` by hand; the scan workflow owns it.
