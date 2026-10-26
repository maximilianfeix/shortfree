<p align="right"><a href="README.md">English</a> · <b>Deutsch</b> · <a href="README.zh-CN.md">简体中文</a></p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="brand/banner-dark.svg">
      <img src="brand/banner-light.svg" alt="shortfree: jede freie Domain mit 1–4 Zeichen, täglich geprüft" width="800">
    </picture>
  </a>
</p>

<p align="center">
  Jede nicht registrierte Domain mit 1–4 Zeichen auf <b>.com .net .de .ai .io .dev .app .tools</b>,<br>
  jeden Tag direkt bei den Nameservern und RDAP-Diensten der Registries geprüft.
</p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/"><img src="https://img.shields.io/badge/Liste_durchsuchen-1FC16B?style=for-the-badge" alt="Liste durchsuchen"></a>
  <a href="#lists"><img src="https://img.shields.io/badge/Listen_laden-0B0C0E?style=for-the-badge" alt="Listen herunterladen"></a>
  <a href="#try-it"><img src="https://img.shields.io/badge/Installieren-0B0C0E?style=for-the-badge" alt="Installieren"></a>
  <a href="#mcp"><img src="https://img.shields.io/badge/MCP-0B0C0E?style=for-the-badge" alt="MCP-Server"></a>
  <a href="#api"><img src="https://img.shields.io/badge/API-0B0C0E?style=for-the-badge" alt="JSON- und CSV-API"></a>
</p>

<p align="center">
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml/badge.svg" alt="täglicher Scan"></a>
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/maximilianfeix/shortfree/releases"><img src="https://img.shields.io/github/v/release/maximilianfeix/shortfree?color=1FC16B" alt="aktuelles Release"></a>
  <a href="https://pkg.go.dev/github.com/maximilianfeix/shortfree"><img src="https://pkg.go.dev/badge/github.com/maximilianfeix/shortfree.svg" alt="Go reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/maximilianfeix/shortfree?color=0B0C0E" alt="MIT-Lizenz"></a>
  <a href="https://maximilianfeix.github.io/shortfree/api/v1/badge.json"><img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fmaximilianfeix.github.io%2Fshortfree%2Fapi%2Fv1%2Fbadge.json" alt="Anzahl der per RDAP bestätigten freien Kurz-Domains"></a>
</p>

<!-- demo -->

<p align="center">
  <a href="#try-it">Ausprobieren</a> ·
  <a href="#numbers">Zahlen</a> ·
  <a href="#lists">Listen</a> ·
  <a href="#limits">Grenzen</a> ·
  <a href="#how-it-works">Funktionsweise</a> ·
  <a href="#mcp">MCP</a> ·
  <a href="#api">API</a> ·
  <a href="#cli">CLI</a> ·
  <a href="#faq">FAQ</a> ·
  <a href="#contributing">Mitmachen</a>
</p>

<a id="try-it"></a>
## Ausprobieren

**Ohne Installation:** [maximilianfeix.github.io/shortfree](https://maximilianfeix.github.io/shortfree/) öffnen, 1–4 Zeichen eintippen und den Status auf allen acht TLDs sehen. Die komplette Liste lässt sich nach TLD, Länge und Zeichenart filtern.

**Kommandozeile** (Go 1.27 oder neuer):

```sh
go install github.com/maximilianfeix/shortfree/cmd/shortfree@latest
shortfree check x7.dev acme.io
```

**Docker:**

```sh
docker run -i --rm ghcr.io/maximilianfeix/shortfree check x7.dev
```

**Claude Code** (MCP):

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

Fertige Binaries für Linux, macOS und Windows hängen an jedem [Release](https://github.com/maximilianfeix/shortfree/releases).

<a id="numbers"></a>
## Aktuelle Zahlen

Die Tabelle mit den Zahlen pro TLD und Länge wird nach jedem Scan automatisch aktualisiert und steht in der [englischen README](README.md#numbers). Die Rohdaten liefert [`stats.json`](https://maximilianfeix.github.io/shortfree/api/v1/stats.json).

<a id="lists"></a>
## Nur die Listen?

Das Begleit-Repo [free-short-domains](https://github.com/maximilianfeix/free-short-domains) enthält schlichte CSV- und TXT-Dateien und wird jeden Tag nach dem Scan aktualisiert. Direktlinks:

| Datei | Inhalt |
|---|---|
| [`csv/all.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/all.csv) | jede per RDAP bestätigte freie Domain mit 1–3 Zeichen auf allen acht TLDs |
| `csv/<tld>.csv`, z. B. [`csv/dev.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/dev.csv) | Namen mit 1–3 Zeichen und Status `available`, `dropping` oder `unverified` |
| `txt/<tld>.txt`, z. B. [`txt/io.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/io.txt) | bestätigte freie Domains mit 1–3 Zeichen, eine pro Zeile |
| `four-letter/<tld>.txt`, z. B. [`four-letter/com.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/four-letter/com.txt) | Namen aus 4 Buchstaben (a–z), die nicht in der Zone stehen |

Alle Namen mit 4 Zeichen, auch mit Ziffern und Bindestrich, liegen als `four-<tld>.txt.gz` im [`data`-Release](https://github.com/maximilianfeix/shortfree/releases/tag/data). Dieselben Daten gibt es als [JSON/CSV-API](#api). Alles steht unter CC0: kein Namensnennungszwang, kein API-Schlüssel.

```sh
# alle freien .dev-Domains mit 2 Zeichen
curl -s https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/dev.txt | awk 'length($0) == 6'
```

<a id="limits"></a>
## Was es kann und was nicht

shortfree

- prüft jedes gültige Label (a–z, 0–9, Bindestrich innen) bis 4 Zeichen: 49.284 Namen pro TLD bis 3 Zeichen und 1.774.224 weitere mit 4 Zeichen;
- fragt direkt die autoritativen Nameserver der TLD, ohne Resolver-Cache;
- bestätigt jeden Kandidaten mit 1–3 Zeichen beim RDAP-Dienst der Registry (`404` = nicht registriert);
- führt registrierte Namen in Redemption oder Pending Delete getrennt auf, mit Ablaufdatum;
- veröffentlicht alles als einfache Dateien unter CC0, ohne Konto, Schlüssel oder Tracking.

shortfree kann nicht

- **Preise oder Reservierungen sehen.** Registries reservieren manche Namen und verkaufen andere zu Premium-Preisen. RDAP meldet beides als nicht registriert. Das betrifft vor allem Namen mit 1–2 Zeichen auf .ai .io .dev .app .tools. Den echten Preis zeigt der Registrar.
- **Namen mit 4 Zeichen per RDAP prüfen.** Sie sind nur per DNS geprüft. Ein registrierter Name ohne Nameserver kann deshalb in den 4-Zeichen-Listen auftauchen. Vor dem Kauf einen Namen mit `shortfree check` prüfen.
- **aktueller sein als der letzte Scan.** Die Listen sind ein täglicher Stand. Ein Name kann Minuten nach dem Scan registriert werden. `shortfree check` und das MCP-Tool `check_domains` fragen die Registry live.

<a id="how-it-works"></a>
## Funktionsweise

1. **Durchlauf.** Für jedes Label schickt shortfree eine `NS`-Abfrage an die autoritativen Nameserver der TLD. `NXDOMAIN` heißt: Der Name steht nicht in der Zone. Alle Namen mit 1–3 Zeichen auf acht TLDs sind rund 400.000 Abfragen und dauern Minuten.
2. **Bestätigung.** Nicht in der Zone heißt noch nicht frei: Namen mit Hold-Status, ohne Nameserver oder in Redemption fehlen dort auch. Darum wird jeder Kandidat mit 1–3 Zeichen beim RDAP-Dienst der Registry nachgeschlagen (Verisign, DENIC, Identity Digital, Google Registry). `404` = nicht registriert. `pending delete` oder `redemption period` = läuft bald aus.
3. **Wiederholung.** [`scan.yml`](.github/workflows/scan.yml) läuft täglich um 03:17 UTC. Es prüft alle Namen mit 1–3 Zeichen und einen von sieben Teilen des 4-Zeichen-Raums, jeder Name mit 4 Zeichen kommt also einmal pro Woche dran. RDAP-Ergebnisse werden spätestens nach 10 Tagen erneuert. Noch nie geprüfte Namen kommen zuerst, dann auslaufende, dann die ältesten Prüfungen.

Jeder RDAP-Host hat ein eigenes Ratenlimit (1,5–6 Anfragen pro Sekunde, siehe [`internal/rdap/rdap.go`](internal/rdap/rdap.go)), damit die Last im Rahmen bleibt. Der Scanner-Zustand liegt im `data`-Release, die Git-Historie enthält nur die kleinen öffentlichen Listen.

<a id="mcp"></a>
## MCP-Server

`shortfree mcp` ist ein [MCP](https://modelcontextprotocol.io)-Server für Claude, Cursor, VS Code und andere Clients. Er bietet vier Tools, die nur lesen:

| Tool | Funktion |
|---|---|
| `check_domains` | Live-Prüfung per DNS + RDAP für bis zu 50 Domains, beliebige Länge und TLD (RDAP für die acht TLDs oben) |
| `find_free_domains` | durchsucht den täglichen Stand nach TLD, Länge, Zeichenart, Teilstring oder Anfang |
| `suggest_domains` | macht aus einem Wort oder Projektnamen kurze Kandidaten und liefert die freien |
| `domain_stats` | Zahlen pro TLD und Länge |

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

Einrichtung für Claude Desktop, Cursor und VS Code und die Bewertungsregeln von `suggest_domains`: [docs/mcp.md](docs/mcp.md) (Englisch).

<a id="api"></a>
## API

Eine statische JSON/CSV-API auf GitHub Pages, nur lesend, nach jedem Scan neu erzeugt. Kein Schlüssel, CORS ist offen.

```
https://maximilianfeix.github.io/shortfree/api/v1/
```

| Pfad | Inhalt |
|---|---|
| `stats.json` | Zahlen pro TLD und Länge |
| `{tld}.json`, `{tld}/{n}.json` | Namen mit 1–3 Zeichen auf einer TLD, optional nur einer Länge |
| `dropping.json` | Namen in Redemption oder Pending Delete, die nächsten zuerst |
| `changes.json` | in den letzten 30 Tagen frei gewordene oder registrierte Namen |
| `free.csv` | jede per RDAP bestätigte freie Domain mit 1–3 Zeichen |
| `badge.json` | [shields.io-Endpoint](https://shields.io/badges/endpoint-badge) mit der aktuellen Zahl |

```sh
curl -s https://maximilianfeix.github.io/shortfree/api/v1/io/3.json | jq -r '.available[] + ".io"'
```

Alle Endpunkte, Felder und Beispiele in Python und JavaScript: [docs/api.md](docs/api.md) (Englisch). OpenAPI-Spezifikation: [docs/openapi.yaml](docs/openapi.yaml).

<a id="cli"></a>
## CLI

```sh
shortfree check q7.tools x4k.app abc.de
```

```
DOMAIN    DNS    RDAP        VERDICT
q7.tools  free   available   FREE
x4k.app   taken  registered  taken
abc.de    taken  registered  taken
```

`check` funktioniert für Domains jeder Länge und TLD. Bei TLDs ohne RDAP-Unterstützung in shortfree gibt es nur ein DNS-Ergebnis (`probably free`). `-json` gibt ein Objekt pro Domain aus:

```sh
shortfree check -json q7.tools | jq '.[0].verdict'
```

Die übrigen Befehle führen den Scan selbst aus: `scan` (1–3 Zeichen, DNS + RDAP), `scan4` (ein Teil der 4-Zeichen-Namen), `export` (schreibt `data/` und die Zahlentabelle), `merge` und `mcp`. `shortfree <befehl> -h` zeigt die Optionen.

### In GitHub Actions

Das Repo ist auch eine GitHub Action. Sie prüft eine Liste von Domains und kann den Job fehlschlagen lassen oder ein Issue öffnen, sobald eine frei wird:

```yaml
- uses: maximilianfeix/shortfree@main
  id: watch
  with:
    domains: acme.dev acme.io
    fail-when: never   # free | taken | never
- run: echo '${{ steps.watch.outputs.free }}'   # ["acme.dev"] oder []
```

Eingaben, Ausgaben und ein täglicher Watch-Workflow: [docs/action.md](docs/action.md) (Englisch).

<a id="faq"></a>
## FAQ

**Ein Name steht als frei in der Liste, aber mein Registrar sagt „nicht verfügbar“ oder „Premium“. Warum?**
Die Registry hat ihn reserviert oder verlangt einen Premium-Preis. RDAP sagt nur, ob ein Name registriert ist, nicht, was er kostet. Das betrifft vor allem Namen mit 1–2 Zeichen außerhalb von .com, .net und .de. Die Website und die MCP-Tools markieren solche Namen als wahrscheinlich Premium.

**Wie aktuell sind die Daten?**
Der Scan startet jeden Tag um 03:17 UTC. Jede JSON-Datei hat ein Feld `updated` mit der Scan-Zeit. Namen mit 1–3 Zeichen werden täglich geprüft, Namen mit 4 Zeichen einmal pro Woche.

**Warum werden Namen mit 4 Zeichen nicht per RDAP bestätigt?**
Es gibt 1.774.224 davon pro TLD. Bei den Ratenlimits von shortfree (1,5–6 Anfragen pro Sekunde) würde ein Durchlauf für eine einzige TLD zwischen 3 und 14 Tagen dauern. DNS beantwortet dieselbe Frage für die meisten Namen in Minuten.

**Warum nur diese acht TLDs?**
Jede TLD braucht einen RDAP-Dienst der Registry, der bei maßvoller Rate zuverlässig antwortet, und jede kostet rund 300.000 DNS-Abfragen am Tag (alle Namen mit 1–3 Zeichen plus ein Siebtel der 4-Zeichen-Namen). Wie man eine hinzufügt, steht in [CONTRIBUTING.md](CONTRIBUTING.md).

**Verkauft shortfree Domains oder trackt es mich?**
Nein. Es gibt keine Affiliate-Links, keine Konten und keine Analytics. Website und API sind statische Dateien auf GitHub Pages. Der Code steht unter MIT, die Daten unter CC0.

<a id="contributing"></a>
## Mitmachen

Issues und Pull Requests sind willkommen. [CONTRIBUTING.md](CONTRIBUTING.md) erklärt, wie man eine TLD hinzufügt. Vor einem PR:

```sh
go test ./...
go vet ./...
gofmt -l .
```

Einen kleinen Scan lokal ausführen:

```sh
go run ./cmd/shortfree scan -tlds tools -rdap-time 1m   # schreibt state/state.json.gz
go run ./cmd/shortfree export -tlds tools -readme ''      # schreibt data/
```

Sicherheitslücken: siehe [SECURITY.md](SECURITY.md).

## Lizenz

Code: [MIT](LICENSE). Daten: [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). Die Registry-Daten stammen aus öffentlichen DNS- und RDAP-Diensten.
