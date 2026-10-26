# Contributing

Thanks for helping. Issues and pull requests are welcome.

## Add a TLD

1. Find the registry's RDAP base URL in the [IANA bootstrap file](https://data.iana.org/rdap/dns.json) (or the registry's docs if it is missing there).
2. Add it to `Endpoints` in [`internal/rdap/rdap.go`](internal/rdap/rdap.go) and, if the host is new, a conservative rate to `DefaultRate`.
3. Add the TLD to `DefaultTLDs` in [`cmd/shortfree/main.go`](cmd/shortfree/main.go).
4. Check that free and taken names are told apart:
   ```sh
   go run ./cmd/shortfree check some-random-name-xyz.<tld> google.<tld>
   ```
5. Open a pull request with the output.

A new TLD adds ~49k DNS queries and up to a few thousand RDAP lookups per day, so it should have a real audience.

## Ground rules

- Stay polite to registries: no rate increases without a measurement in the PR.
- `gofmt`, `go vet ./...` and `go test ./...` must pass.
- Do not commit changes to `data/` by hand; the scan workflow owns it.
