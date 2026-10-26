# Supported TLDs

shortfree scans every TLD in [`internal/tlds`](../internal/tlds/tlds.go)
(`shortfree tlds` prints the list; both workflows build their job matrix from
`shortfree tlds -json`). Each one needs:

- an RDAP base URL in `rdap.Endpoints` (IANA bootstrap
  <https://data.iana.org/rdap/dns.json> first, the registry's own service
  otherwise) and a rate for its host in `rdap.DefaultRate`
  (1-2 requests/s for hosts we have not measured),
- a check that the endpoint answers `404` for a random long label and `200`
  for a registered name, plus `shortfree check <random>.<tld> nic.<tld>` for
  the DNS side,
- a `Policy` in `internal/tlds` if the registry rejects some short labels
  (`MinLen`) or if `nic.<tld>` is not registered (`Canary`, see below).

**Canary.** Before verifying a TLD, the daily scan looks up `nic.<tld>`. If the
registry does not report it as registered (404, errors, rate-limit ban), the
TLD's RDAP step is skipped for that run, so a broken endpoint can never mark
names as available. `nic.<tld>` was delegated on every TLD below on
2026-10-26.

## Scanned

Verified 2026-10-26: RDAP `404` for a random 15 character label, `200` for
`google.<tld>`; DNS `free` for `q-z7.<tld>`, `taken` for `google.<tld>` /
`nic.<tld>`; `shortfree check` end to end.

| TLD | Registry / backend | RDAP base URL | Source | Min length | Notes |
|---|---|---|---|--:|---|
| com | Verisign | `https://rdap.verisign.com/com/v1/` | IANA | 1 | all 1-3 char registered or reserved in practice |
| net | Verisign | `https://rdap.verisign.com/net/v1/` | IANA | 1 | |
| de | DENIC | `https://rdap.denic.de/` | DENIC | 1 | not in IANA bootstrap |
| ai | Government of Anguilla / Identity Digital | `https://rdap.identitydigital.services/rdap/` | IANA | 1 | shared host, see rate note |
| io | ICB / Identity Digital | `https://rdap.identitydigital.services/rdap/` | registry | 1 | not in IANA bootstrap |
| dev | Google Registry | `https://pubapi.registry.google/rdap/` | IANA | 1 | 429 at 10 req/s |
| app | Google Registry | `https://pubapi.registry.google/rdap/` | IANA | 1 | |
| tools | Identity Digital | `https://rdap.identitydigital.services/rdap/` | IANA | 1 | |
| co | .CO Internet / GoDaddy Registry | `https://rdap.registry.co/co/` | registry | 1 | not in IANA bootstrap; 1-2 char mostly premium |
| org | Public Interest Registry | `https://rdap.publicinterestregistry.org/rdap/` | IANA | 1 | 1-2 char mostly reserved/premium |
| xyz | XYZ.COM / CentralNic | `https://rdap.centralnic.com/xyz/` | IANA | 1 | premium pricing for short names |
| so | Somali NIC (SONIC) | `https://rdap.nic.so/` | registry | **3** | not in IANA bootstrap. 3-63 chars per GoDaddy/EuroDNS; local-presence rule unclear |
| to | Tonic | `https://rdap.tonicregistry.to/rdap/` | IANA | **3** | Tonic's FAQ states no minimum; GoDaddy and Easyname say 3, 101domain says 2. 3 is the conservative choice |
| cc | eNIC / Verisign | `https://tld-rdap.verisign.com/cc/v1/` | IANA | 1 | |
| tv | .tv Corporation / GoDaddy Registry | `https://rdap.nic.tv/` | IANA | 1 | premium pricing for short names |
| fm | dotFM / CentralNic | `https://rdap.centralnic.com/fm/` | IANA | 1 | |
| is | ISNIC | `https://rdap.isnic.is/rdap/` | IANA | 1 | [ISNIC rules](https://www.isnic.is/en/domain/rules) §6 set no minimum; all 1 char names were delegated on 2026-10-26 |
| ly | Libya NIC | `https://rdap.nic.ly/` | IANA | **4** | 4-63 chars per Openprovider, WebNIC, EuroDNS, NiceNIC; shorter names (e.g. bit.ly) are grandfathered |
| run | Identity Digital | `https://rdap.identitydigital.services/rdap/` | IANA | 1 | RDAP check pending, see below |
| inc | Intercap / CentralNic | `https://rdap.centralnic.com/inc/` | IANA | 1 | high registration price |

Rates (`rdap.DefaultRate`) are the **total** per RDAP host name, however
many runners scan in parallel: Verisign 4/s (.com + .net), Identity Digital
2/s (.ai + .io + .tools + .run), DENIC 2/s, PIR 2/s, `.co` 2/s, CentralNic
2/s (.xyz + .fm + .inc), Verisign `.cc` 2/s, Google 1.5/s (.dev + .app),
`.tv` `.to` `.is` `.ly` `.so` 1/s. Each scan takes the share of the TLDs it
scans (`rdap.SharedRate`): the per-TLD matrix job for `.ai` gets 2/4 =
0.5/s, a local run of all TLDs gets the whole total. Adding a TLD to a
shared host lowers the others' share instead of raising the load.

**Identity Digital ban (2026-10-26).** Later on the day its 10 req/s
measurement, `rdap.identitydigital.services` answered every request from the
measuring IP with Cloudflare error 1015 and `Retry-After: 83359` (~23 h).
Its total is now 2/s, and a `Retry-After` above 10 minutes makes the client
stop asking that host until it expires (`rdap.ErrBlocked`) instead of
retrying every queued name. Because of the ban, the `.run` RDAP answers
above could not be checked from here; `.run` is in the IANA bootstrap
(same host as `.ai`/`.tools`), DNS was verified, and the canary keeps it
safe until the first scan shows real answers.

## Not added (yet)

| TLD | Reason | Next step |
|---|---|---|
| me | No RDAP in IANA bootstrap or delegation record (WHOIS `whois.nic.me`). Probably served by Identity Digital, but unverifiable on 2026-10-26 because of the ban above. | Check `https://rdap.identitydigital.services/rdap/domain/nic.me` (200) and a random name (404), then add |
| sh | Same: ICB like `.io` (WHOIS `whois.nic.sh`), likely Identity Digital, unverifiable on 2026-10-26. | As for `.me` |
| vc | Same: WHOIS is `whois.identitydigital.services`; `rdap.afilias.net/rdap/vc/` (listed by aggregators) only answered via IPv6 here, and GitHub runners have no IPv6. | As for `.me` |
| gg | No RDAP service: none in IANA bootstrap or delegation record; `rdap.gg` is a web page answering 200 for everything. | DNS-only support would need a "no RDAP" mode; not worth it while 1-3 char results cannot be verified |

DNS (NS queries to the authoritative servers) worked for all of them on
2026-10-26.
