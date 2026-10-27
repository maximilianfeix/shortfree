# Accuracy

What shortfree's data proves, what it does not, and how to report a wrong result.

## Method

1. **Scope.** Every ASCII LDH label directly below a TLD: a–z, 0–9 and inner hyphens, no `--` at positions 3–4, length 1–4. Internationalized names (Unicode or `xn--` A-labels) are not enumerated, and `shortfree check` rejects them. Labels shorter than a registry's minimum length ([tlds.md](tlds.md)) are not published.
2. **DNS sweep.** An `NS` query for each label goes to the TLD's own authoritative servers. `NXDOMAIN` = not in the zone; a delegation = registered. A name without an answer (timeout, SERVFAIL) is recorded as `unknown`: not checked, never counted as taken.
3. **RDAP verification (1–3 characters).** Names not in the zone are looked up in the registry's RDAP service. `404` = no registration. `redemption period` / `pending delete` = dropping. Before a TLD is verified, two controls run: a known registered name must answer "registered", and a random 28 character name must answer `404`. If either fails, the TLD's RDAP answers that day are not used.
4. **Evidence age.** An RDAP answer is published as verified for 14 days (`max_evidence_age_days`). Older answers are published as `unverified` and listed in `stale`. The scan rechecks after 10 days and reserves a quarter of each registry's budget for overdue rechecks.
5. **4 characters.** DNS only, one of seven shards per day. Which shards were swept and when is in `coverage.json`; names in a shard that was never swept are `unscanned`, not taken. A shard sweep that runs out of time (slow name servers) keeps the answers it got but is not recorded as swept: the shard keeps its previous coverage, and names it did not reach keep their previous state.
6. **Publishing safeguards.** A TLD whose daily job fails keeps its previous data (coverage `stale`). A TLD whose count of not-delegated names, or of free 4 character names in a shard, changes by more than 50 % and 200 names in one run keeps its previous data (coverage `held`, with the reason). A TLD that never had data is left out (`missing`) instead of showing 0. The one-off bootstrap publishes nothing unless every TLD and every shard succeeded.

Each daily run is archived as an immutable `snapshot-YYYY-MM-DD.tar.gz` with a `manifest.json` (checksums, counts, coverage, source commit and workflow run) in the [`data` release](https://github.com/maximilianfeix/shortfree/releases/tag/data).

## What "available" means

**Not registered**, as answered by the registry at most 14 days before the scan. It does **not** mean a registrar will sell the name, or at a normal price. `shortfree check` reports this as `registration: not_registered` and `registrable: unconfirmed`.

## Known false-positive classes

Names that appear available or free but cannot be registered normally:

| Class | Why | Where |
|---|---|---|
| Registry-reserved names | Registries answer `404` for names they withhold (e.g. Identity Digital's reserved list, .io's controlled short names) | mostly 1–2 characters outside .com/.net/.de; flagged `premium_likely` |
| Premium-priced names | Unregistered, but sold at premium prices; RDAP has no price | short names on most new gTLDs |
| Restricted or eligibility rules | Some registries require local presence or other conditions | country-code TLDs, see [tlds.md](tlds.md) |
| Registered without name servers (4 characters) | 4 character names are DNS-only: a registered name without delegation looks free | `dns_only` / `free` 4 character counts |
| Registered after the scan | Data is a daily snapshot | everything; `shortfree check` asks live |
| Registry or resolver faults | A broken endpoint can answer `404` or "registered" for everything | mitigated by the canaries and the jump check above |

And false negatives (a free name that is not listed): names without a DNS answer (`unknown`), names in never-swept shards (`unscanned`), TLDs whose latest data was held or is stale, and names whose RDAP check is still queued (`unverified`).

## Measuring it

`coverage.json` (also `api/v1/coverage.json`) publishes per TLD and length: the last sweep, DNS unknowns, unscanned shards, verified / pending / stale counts, the oldest and median RDAP evidence age, and held or missing TLDs. A registrar sample (checking a random set of listed names at registrars) is not automated yet; results reported through the issue template below are the current accuracy signal.

## Reporting a wrong result

Open a [wrong result issue](https://github.com/maximilianfeix/shortfree/issues/new?template=wrong-result.yml) with the domain, what shortfree showed (status, snapshot date) and what the registrar said (price, "reserved", "unavailable"). Please include the output of `shortfree check -json <domain>`. Systematic errors (a whole TLD or class) lead to a registry policy entry in [tlds.md](tlds.md) or a code fix.
