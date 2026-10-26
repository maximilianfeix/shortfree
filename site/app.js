// shortfree front end: reads the daily data files and answers everything
// client-side. 4 character results come from per-TLD bitmaps (1 bit per name,
// base-37 index over a-z 0-9 -), loaded only when needed.
"use strict";

const ALPHA = "abcdefghijklmnopqrstuvwxyz0123456789-";
const SPACE4 = 37 ** 4;
const PAGE = matchMedia("(max-width: 767px)").matches ? 60 : 360;
const REPO = "https://github.com/maximilianfeix/shortfree";
const LIVE_DOCS = `${REPO}#check-a-domain-yourself`;
const MCP_DOCS = `${REPO}/blob/main/docs/mcp.md`;
const STALE_HOURS = 36;
const SHORTLIST_KEY = "shortfree.shortlist.v1";
const SHORTLIST_MAX = 200;
const $ = (s) => document.querySelector(s);
const el = (tag, props = {}, ...kids) => {
  const n = Object.assign(document.createElement(tag), props);
  n.append(...kids.filter((k) => k != null && k !== false));
  return n;
};
const ext = (href, text, label) => {
  const a = el("a", { href, target: "_blank", rel: "noopener noreferrer", textContent: `${text} ↗` });
  a.setAttribute("aria-label", `${label || text} (opens in a new tab)`);
  return a;
};
const fmt = (n) => n.toLocaleString("en-US");
const toDate = (v) => (typeof v === "number" ? new Date(v < 1e12 ? v * 1000 : v) : new Date(v));
const scanTime = (v) => {
  const d = toDate(v);
  return isNaN(d) ? "unknown" : d.toISOString().slice(0, 16).replace("T", " ") + " UTC";
};
const hoursSince = (v) => Math.max(0, (Date.now() - toDate(v).getTime()) / 3.6e6);
function ageText(v) {
  const h = hoursSince(v);
  if (!isFinite(h)) return "";
  if (h < 1) return "under 1 h ago";
  if (h < 48) return `${Math.round(h)} h ago`;
  return `${Math.round(h / 24)} days ago`;
}
// Registries of new gTLDs usually hold back 1-2 character names (reserved or
// premium) even though RDAP reports them as unregistered.
const premiumLikely = (label, tld) => label.length <= 2 && !["com", "net", "de"].includes(tld);

// Local storage can be missing or throw (private mode, blocked site data).
const store = {
  get(k, fallback) {
    try { const v = localStorage.getItem(k); return v == null ? fallback : JSON.parse(v); } catch { return fallback; }
  },
  set(k, v) {
    try { localStorage.setItem(k, JSON.stringify(v)); return true; } catch { return false; }
  },
};

const DEFAULTS = { len: 3, kind: "letters", view: "az", ver: "verified", has: "" };
const state = {
  stats: null, avail: null, cov: null, sets: {}, byLen: {}, checked: {}, bitmaps: {}, four: {},
  tld: null, ...DEFAULTS, kindSet: false, touched: false, shown: PAGE, res: null,
};

/* ---------- labels ---------- */

function valid(s) {
  return /^[a-z0-9-]{1,4}$/.test(s) && !s.startsWith("-") && !s.endsWith("-") && !s.includes("--");
}
function index(s) {
  let i = 0;
  for (const c of s) i = i * 37 + ALPHA.indexOf(c);
  return i;
}
const decode4 = (i) => ALPHA[(i / 50653) | 0] + ALPHA[((i / 1369) | 0) % 37] + ALPHA[((i / 37) | 0) % 37] + ALPHA[i % 37];

// Pattern flags, packed into one small integer per label:
//   bits 0-1 kind (0 letters, 1 digits, 2 mixed), 4 pronounceable,
//   8 repeats, 16 hyphen, 32 look-alike characters.
// The definitions are shown on the page (VIEW_NOTE); keep them in sync.
const KIND_N = { letters: 0, digits: 1, mixed: 2 };
const VOWEL = new Uint8Array(37);
for (const c of "aeiou") VOWEL[ALPHA.indexOf(c)] = 1;
const PRON = new Set();
for (const p of ["C", "V", "CV", "VC", "CVC", "VCV", "CVCV", "VCVC", "CVVC"]) {
  let bits = 0;
  for (let k = 0; k < p.length; k++) if (p[k] === "V") bits |= 1 << k;
  PRON.add(p.length * 16 + bits);
}
const cs4 = new Uint8Array(4);
function metaOf(cs, n) {
  let letters = 0, digits = 0, hyphen = 0, vbits = 0, look = 0;
  for (let k = 0; k < n; k++) {
    const c = cs[k];
    if (c < 26) { letters++; if (VOWEL[c]) vbits |= 1 << k; } else if (c < 36) { digits++; if (c === 26 || c === 27) look = 1; } else hyphen = 1;
  }
  const kind = letters === n ? 0 : digits === n ? 1 : 2;
  let p = kind;
  if (kind === 0 && PRON.has(n * 16 + vbits)) p |= 4;
  if (n >= 2) {
    let a = cs[0], b = -1, distinct = 1;
    for (let k = 1; k < n; k++) {
      const c = cs[k];
      if (c === a || c === b) continue;
      if (b < 0) { b = c; distinct = 2; } else { distinct = 3; break; }
    }
    if (distinct === 1 || (distinct === 2 && n >= 3)) p |= 8;
  }
  if (hyphen) p |= 16;
  if (kind === 2 && look) p |= 32; // 0 next to letters reads as o, 1 as l
  return p;
}
const metaOfLabel = (s) => metaOf([...s].map((c) => ALPHA.indexOf(c)), s.length);
// Ranking score from the pattern flags: pronounceable letters first, then
// other letters, all-digit, mixed; repeats get a small boost; hyphens last.
const SCORE = new Uint8Array(64);
for (let p = 0; p < 64; p++) {
  let s = [8, 5, 2, 0][p & 3];
  if (p & 4) s += 6;
  if (p & 8) s += 3;
  if (p & 32) s -= 1;
  if (p & 16) s = p & 8 ? 1 : 0;
  SCORE[p] = Math.max(0, s);
}
let META4 = null; // pattern flags for all 37^4 labels, built on first use
function meta4() {
  if (META4) return META4;
  const t0 = performance.now();
  META4 = new Uint8Array(SPACE4);
  let i = 0;
  for (let a = 0; a < 37; a++) for (let b = 0; b < 37; b++) for (let c = 0; c < 37; c++) for (let d = 0; d < 37; d++) {
    cs4[0] = a; cs4[1] = b; cs4[2] = c; cs4[3] = d;
    META4[i++] = metaOf(cs4, 4);
  }
  perf("meta4", t0);
  return META4;
}
const perfLog = [];
function perf(name, t0) {
  perfLog.push([name, Math.round(performance.now() - t0)]);
  if (perfLog.length > 50) perfLog.shift();
}
window.__shortfreePerf = perfLog;

/* ---------- data ---------- */

async function json(path) {
  const r = await fetch(path, { cache: "no-cache" });
  if (!r.ok) throw new Error(`${path}: ${r.status}`);
  return r.json();
}

// bitmap resolves to {bm} or {error: "missing" | "failed"}; failures are not
// cached so a retry can fetch again.
function bitmap(tld) {
  if (!state.bitmaps[tld]) {
    state.bitmaps[tld] = fetch(`data/4/${tld}.bin`)
      .then(async (r) => {
        if (r.status === 404) return { error: "missing" };
        if (!r.ok) throw new Error(String(r.status));
        return { bm: new Uint8Array(await r.arrayBuffer()) };
      })
      .catch(() => ({ error: "failed" }))
      .then((x) => { if (x.error) delete state.bitmaps[tld]; return x; });
  }
  return state.bitmaps[tld];
}
const bit = (bm, i) => (bm[i >> 3] >> (i & 7)) & 1;

// four(tld): the free 4 character indices of a TLD as one compact array,
// built once per TLD instead of on every filter change.
const POP = new Uint8Array(256);
for (let i = 1; i < 256; i++) POP[i] = POP[i >> 1] + (i & 1);
async function four(tld) {
  if (state.four[tld]) return state.four[tld];
  const r = await bitmap(tld);
  if (!r.bm) return r;
  const t0 = performance.now();
  const bm = r.bm;
  let n = 0;
  for (let j = 0; j < bm.length; j++) n += POP[bm[j]];
  const idx = new Uint32Array(n);
  let k = 0;
  for (let j = 0; j < bm.length; j++) {
    let b = bm[j];
    while (b) {
      const low = b & -b;
      const i = j * 8 + 31 - Math.clz32(low);
      if (i < SPACE4) idx[k++] = i;
      b ^= low;
    }
  }
  state.four[tld] = { idx: idx.subarray(0, k) };
  perf(`index ${tld}`, t0);
  return state.four[tld];
}

// Entries may be plain labels or, in newer data, objects such as
// {name, checked}; a TLD may also carry a {label: time} "checked" map.
function normalizeList(arr, tld) {
  const out = [];
  for (const x of Array.isArray(arr) ? arr : []) {
    if (typeof x === "string") { out.push(x); continue; }
    const n = x && (x.name || x.label);
    if (typeof n !== "string") continue;
    out.push(n);
    const ck = x.checked ?? x.verified ?? x.c;
    if (ck) (state.checked[tld] ||= new Map()).set(n, ck);
  }
  return out;
}
function prepare(avail, stats) {
  for (const tld of stats.order) {
    const a = avail.tlds?.[tld];
    if (!a) continue;
    const free = normalizeList(a.available, tld);
    const unv = normalizeList(a.unverified, tld);
    const drops = (Array.isArray(a.dropping) ? a.dropping : []).filter((d) => d && typeof d.name === "string");
    if (a.checked && typeof a.checked === "object" && !Array.isArray(a.checked)) {
      for (const [k, v] of Object.entries(a.checked)) (state.checked[tld] ||= new Map()).set(k, v);
    }
    const unk = new Set(normalizeList(a.unknown, tld)); // no DNS answer: not checked
    state.sets[tld] = { free: new Set(free), unv: new Set(unv), drop: new Set(drops.map((d) => d.name)), unk, drops, lists: [free, unv] };
  }
}
// byLen(tld, n) -> [verified, unverified], each a list of {s, i, p} (label,
// base-37 index, pattern flags) in index order; built once per TLD on use.
function byLen(tld, n) {
  if (!state.byLen[tld]) {
    const t0 = performance.now();
    const by = { 1: [[], []], 2: [[], []], 3: [[], []] };
    const lists = state.sets[tld]?.lists || [[], []];
    const cs = new Uint8Array(4);
    lists.forEach((arr, u) => {
      for (const s of arr) {
        if (!by[s.length]) continue;
        let i = 0;
        for (let k = 0; k < s.length; k++) { cs[k] = ALPHA.indexOf(s[k]); i = i * 37 + cs[k]; }
        by[s.length][u].push({ s, i, p: metaOf(cs, s.length) });
      }
    });
    for (const k of [1, 2, 3]) for (const l of by[k]) l.sort((x, y) => x.i - y.i);
    state.byLen[tld] = by;
    perf(`byLen ${tld}`, t0);
  }
  return state.byLen[tld][n] || [[], []];
}

// coverageFor reads data/coverage.json defensively; it is optional and its
// shape may vary. Returns {pct (0-1), at} or null.
function coverageFor(tld, len) {
  const c = state.cov;
  if (!c || typeof c !== "object") return null;
  const t = (c.tlds && c.tlds[tld]) || c[tld];
  if (!t || typeof t !== "object") return null;
  // shortfree's coverage.json: tlds.<tld>.lengths.<n>.{swept, dns_unknown, unscanned}.
  // Coverage = names with a DNS answer / registrable names (swept + never swept).
  if (t.lengths && typeof t.lengths === "object") {
    let done = 0, total = 0;
    for (const x of len != null ? [t.lengths[len]] : Object.values(t.lengths)) {
      if (!x || typeof x.swept !== "number") continue;
      done += x.swept - (x.dns_unknown || 0);
      total += x.swept + (x.unscanned || 0);
    }
    return { pct: total ? Math.max(0, Math.min(1, done / total)) : null, at: t.scanned_at ?? null };
  }
  const x = len != null && t[len] && typeof t[len] === "object" ? t[len] : t;
  let pct = [x.pct, x.percent, x.coverage, x.ratio].find((v) => typeof v === "number");
  if (pct == null) {
    const done = [x.answered, x.scanned, x.done, x.checked].find((v) => typeof v === "number");
    const total = [x.total, x.valid, x.expected].find((v) => typeof v === "number" && v > 0);
    if (done != null && total) pct = done / total;
  }
  const at = [x.updated, x.scanned_at, x.last_scan, x.at, t.updated, t.scanned_at, t.last_scan].find((v) => typeof v === "string" || typeof v === "number");
  if (pct == null && at == null) return null;
  if (pct != null && pct > 1) pct /= 100;
  return { pct: pct == null || !isFinite(pct) ? null : Math.max(0, Math.min(1, pct)), at: at ?? null };
}

/* ---------- status ---------- */

// status of label.tld: free | free4 (DNS-only candidate) | unv (DNS-free, RDAP
// pending) | drop | taken | unknown
// A name without a DNS answer, or in a 4 character shard never swept, is
// "unknown", not "taken".
async function status(label, tld) {
  if (label.length === 4) {
    const r = await bitmap(tld);
    if (!r.bm) return "unknown";
    if (bit(r.bm, index(label))) return "free4";
    return swept4(label, tld) ? "taken" : "unknown";
  }
  const s = state.sets[tld];
  if (!s) return "unknown";
  if (s.free.has(label)) return "free";
  if (s.drop.has(label)) return "drop";
  if (s.unv.has(label)) return "unv";
  if (s.unk.has(label)) return "unknown";
  return "taken";
}
// swept4 reports whether label's 4 character shard (first character index
// mod shard count) was ever swept, per coverage.json. No data: assume yes.
function swept4(label, tld) {
  const shards = state.cov?.tlds?.[tld]?.lengths?.["4"]?.shards;
  if (!Array.isArray(shards) || !shards.length) return true;
  return Boolean(shards[ALPHA.indexOf(label[0]) % shards.length]?.scanned_at);
}
const WORD = { free: "free", free4: "likely free", unv: "likely free", drop: "dropping", taken: "taken", unknown: "no data" };
const STATE = { free: "free", free4: "unverified", unv: "unverified", drop: "dropping", taken: "taken", unknown: "unknown" };
const ASIDE = { free4: "DNS check only", unv: "RDAP check queued" };
const EVIDENCE = {
  free: "Not registered: the registry's RDAP service had no record of it (checked within 14 days). Reserved or premium names look the same; the registrar has the final word.",
  free4: "Candidate: not in the zone (DNS check only). 4 character names get no RDAP check, so some are registered but not delegated, reserved or premium.",
  unv: "Unverified: not in the zone; the RDAP check is still queued.",
  drop: "Registered, but in redemption or pending delete.",
  taken: "Registered at the last scan.",
  unknown: "Not checked: no scan data, no DNS answer at the last sweep, or its 4 character shard has not been swept yet.",
};
const isFree = (st) => st === "free" || st === "free4" || st === "unv";
const statusTag = (st) => {
  const s = el("span", { className: "status", textContent: WORD[st] });
  s.dataset.state = STATE[st];
  return s;
};

// parseQuery turns "x7", ".x7", "x7.dev", "https://www.x7.dev/" into {label, tld}.
function parseQuery(raw) {
  let q = raw.trim().toLowerCase().replace(/^[a-z]+:\/\//, "").replace(/[/?#].*$/, "").replace(/^\.+|\.+$/g, "");
  if (q.startsWith("www.") && q.split(".").length > 2) q = q.slice(4);
  const dot = q.indexOf(".");
  if (dot < 0) return { label: q, tld: "" };
  return { label: q.slice(0, dot), tld: q.slice(dot + 1) };
}

async function copyText(text, btn, idle = "Copy") {
  try {
    await navigator.clipboard.writeText(text);
    btn.textContent = "Copied";
  } catch {
    btn.textContent = "Copy failed";
  }
  setTimeout(() => { btn.textContent = idle; }, 2000);
}

/* ---------- shortlist ---------- */

let shortlist = [];
let shortlistSaved = true;
function loadShortlist() {
  const v = store.get(SHORTLIST_KEY, []);
  shortlist = Array.isArray(v)
    ? v.filter((x) => x && typeof x.d === "string" && /^[a-z0-9-]{1,4}\.[a-z0-9.-]+$/.test(x.d)).slice(0, SHORTLIST_MAX)
    : [];
}
const isSaved = (d) => shortlist.some((x) => x.d === d);
function toggleSave(d, st) {
  if (isSaved(d)) shortlist = shortlist.filter((x) => x.d !== d);
  else if (shortlist.length < SHORTLIST_MAX) shortlist.push({ d, s: st || "", at: state.stats ? state.stats.updated : "" });
  shortlistSaved = store.set(SHORTLIST_KEY, shortlist);
  syncSaveButtons();
  renderShortlist();
}
function saveButton(d, st) {
  const b = el("button", { type: "button", className: "btn" });
  b.dataset.save = d;
  b.dataset.st = st;
  b.onclick = () => toggleSave(d, st);
  paintSave(b);
  return b;
}
function paintSave(b) {
  const on = isSaved(b.dataset.save);
  b.textContent = on ? "Saved" : "Save";
  b.setAttribute("aria-pressed", String(on));
  b.setAttribute("aria-label", `${on ? "Saved to" : "Save to"} shortlist: ${b.dataset.save}`);
}
function syncSaveButtons() {
  for (const b of document.querySelectorAll("[data-save]")) paintSave(b);
  for (const li of document.querySelectorAll("#list li[data-d]")) li.classList.toggle("saved", isSaved(li.dataset.d));
}

function shareURL(list) {
  // domains are [a-z0-9.-] only, so the list needs no escaping
  return `${location.origin}${location.pathname}?list=${list.join(",")}`;
}
function sharedFromURL() {
  const raw = new URLSearchParams(location.search).get("list");
  if (!raw) return [];
  const out = [];
  for (const d of raw.toLowerCase().split(",")) {
    const { label, tld } = parseQuery(d);
    if (valid(label) && /^[a-z0-9-]+(\.[a-z0-9-]+)*$/.test(tld) && !out.includes(`${label}.${tld}`)) out.push(`${label}.${tld}`);
    if (out.length >= SHORTLIST_MAX) break;
  }
  return out;
}

async function shortlistRow(d, mode) {
  const { label, tld } = parseQuery(d);
  const known = state.stats && state.stats.order.includes(tld);
  const st = known ? await status(label, tld) : "unknown";
  const meta = el("div", { className: "meta" }, statusTag(st));
  const saved = shortlist.find((x) => x.d === d);
  if (saved && saved.s && isFree(saved.s) && !isFree(st) && st !== "unknown") meta.append(el("span", { className: "aside warn-text", textContent: "was free when saved" }));
  else if (!known) meta.append(el("span", { className: "aside", textContent: "TLD not in the scan" }));
  const det = el("button", { type: "button", className: "btn", textContent: "Details" });
  det.setAttribute("aria-label", `Details for ${d}`);
  det.onclick = () => openDetails(label, tld, st, det);
  const actions = el("span", { className: "actions" }, det);
  if (mode === "mine") {
    const rm = el("button", { type: "button", className: "btn", textContent: "Remove" });
    rm.setAttribute("aria-label", `Remove ${d} from shortlist`);
    rm.onclick = () => {
      const next = rm.closest("li").nextElementSibling || rm.closest("li").previousElementSibling;
      toggleSave(d);
      (next?.querySelector("button") || $("#sl-h")).focus();
    };
    actions.append(rm);
  } else actions.append(saveButton(d, st));
  meta.append(actions);
  return el("li", { className: "row" }, el("span", { className: "domain", textContent: d }), meta);
}

let slSeq = 0;
async function renderShortlist() {
  const seq = ++slSeq;
  const shared = sharedFromURL();
  const mine = shortlist.map((x) => x.d);
  const [mineRows, sharedRows] = await Promise.all([Promise.all(mine.map((d) => shortlistRow(d, "mine"))), Promise.all(shared.map((d) => shortlistRow(d, "shared")))]);
  if (seq !== slSeq) return;
  $("#sl-count").textContent = mine.length ? ` (${mine.length})` : "";
  $("#sl-list").replaceChildren(...mineRows);
  $("#sl-empty").hidden = mine.length > 0;
  $("#sl-actions").hidden = mine.length === 0;
  $("#sl-storage").hidden = shortlistSaved;
  $("#sl-shared").hidden = shared.length === 0;
  if (shared.length) {
    $("#sl-shared-n").textContent = `${shared.length} ${shared.length === 1 ? "name" : "names"}`;
    $("#sl-shared-list").replaceChildren(...sharedRows);
  }
}

/* ---------- details: decision step + due diligence ---------- */

const REGISTRARS = [
  ["Porkbun", (d) => `https://porkbun.com/checkout/search?q=${encodeURIComponent(d)}`],
  ["Namecheap", (d) => `https://www.namecheap.com/domains/registration/results/?domain=${encodeURIComponent(d)}`],
  ["Cloudflare", (d) => `https://domains.cloudflare.com/?domainToCheck=${encodeURIComponent(d)}`],
];
function diligence(label) {
  const q = encodeURIComponent(label);
  const tm = [
    ["EUIPO TMview", `https://www.tmdn.org/tmview/#/tmview/results?page=1&pageSize=30&criteria=C&basicSearch=${q}`, "trademarks in the EU and member states"],
    ["USPTO", `https://tmsearch.uspto.gov/search/search-results?query=${q}`, "US trademarks"],
  ];
  const names = [
    ["GitHub", `https://github.com/${q}`, "user or organisation"],
    ["npm", `https://www.npmjs.com/package/${q}`, "package"],
    ["PyPI", `https://pypi.org/project/${q}/`, "package"],
  ];
  if (!label.includes("-")) names.push(["X", `https://x.com/${q}`, "handle"], ["Instagram", `https://www.instagram.com/${q}/`, "handle"]);
  return { tm, names };
}

let detailsOpener = null;
function openDetails(label, tld, st, opener) {
  const d = `${label}.${tld}`;
  const dlg = $("#details");
  detailsOpener = opener || document.activeElement;
  $("#d-h").textContent = d;
  const body = [];
  const head = el("p", { className: "d-status" }, statusTag(st), " ", EVIDENCE[st]);
  body.push(head);

  const facts = [];
  const ck = state.checked[tld]?.get(label);
  if (ck && st === "free") facts.push(`RDAP check ${scanTime(ck)} (${ageText(ck)})`);
  if (state.stats) facts.push(`Scan ${scanTime(state.stats.updated)} (${ageText(state.stats.updated)})`);
  const cov = coverageFor(tld, label.length);
  if (cov && cov.pct != null) facts.push(`.${tld} ${label.length} character coverage ${(cov.pct * 100).toFixed(1)}%`);
  if (isFree(st) && premiumLikely(label, tld)) facts.push("often reserved or premium at this length");
  if (facts.length) body.push(el("p", { className: "note" }, facts.join(" · ")));
  if (state.stats && hoursSince(state.stats.updated) > STALE_HOURS) body.push(el("p", { className: "note warn-text", textContent: "This snapshot is older than usual. Check live before you decide." }));

  const copyBtn = el("button", { type: "button", className: "btn", textContent: "Copy" });
  copyBtn.setAttribute("aria-label", `Copy ${d}`);
  copyBtn.onclick = () => copyText(d, copyBtn);
  body.push(el("div", { className: "d-actions" }, copyBtn, saveButton(d, st)));

  if (st !== "taken") {
    body.push(el("h3", { textContent: "Check price and register" }));
    body.push(el("p", { className: "note", textContent: "The registrar shows whether the name can really be registered, if it is premium, and the first-year and renewal price. Plain search links, no referral or affiliate codes." }));
    body.push(el("ul", { className: "d-links" }, ...REGISTRARS.map(([n, f]) => el("li", {}, ext(f(d), n, `Check ${d} at ${n}`)))));
  }

  const { tm, names } = diligence(label);
  body.push(el("h3", { textContent: "Before you commit" }));
  body.push(el("p", { className: "note", textContent: "Searches on other sites. shortfree does not check them; a free domain says nothing about trademarks or handles." }));
  body.push(el("ul", { className: "d-links" }, ...tm.map(([n, href, what]) => el("li", {}, ext(href, n, `Search ${what} for ${label} on ${n}`)))));
  body.push(el("ul", { className: "d-links" }, ...names.map(([n, href, what]) => el("li", {}, ext(href, n, `${n} ${what} ${label}`)))));
  const p = metaOfLabel(label);
  if (p & 32) body.push(el("p", { className: "note", textContent: "Contains 0 or 1 next to letters, which read like o and l. Easy to mistype when spoken." }));
  else if (p & 16) body.push(el("p", { className: "note", textContent: "Contains a hyphen, which people tend to drop when typing or saying the name." }));

  body.push(el("p", { className: "note" }, "Live answer from the registry: ", el("code", { textContent: `shortfree check ${d}` }), " or the ", el("a", { href: MCP_DOCS, textContent: "MCP server" }), "."));
  $("#d-body").replaceChildren(...body);
  if (!dlg.open) dlg.showModal();
  $("#d-close").focus();
}
function closeDetails() {
  $("#details").close();
}

/* ---------- search ---------- */

function resultRow(label, tld, st, focus) {
  const d = `${label}.${tld}`;
  const free = isFree(st);
  const meta = el("div", { className: "meta" }, statusTag(st));
  const notes = [ASIDE[st], free && premiumLikely(label, tld) ? "often reserved or premium" : ""].filter(Boolean);
  if (notes.length) meta.append(el("span", { className: "aside", textContent: notes.join(" · ") }));
  if (free) {
    const actions = el("span", { className: "actions" });
    if (focus) {
      const c = el("button", { type: "button", className: "btn btn-free", textContent: "Copy" });
      c.setAttribute("aria-label", `Copy ${d}`);
      c.onclick = () => copyText(d, c);
      actions.append(c);
    }
    const det = el("button", { type: "button", className: "btn", textContent: "Details" });
    det.setAttribute("aria-label", `Details and price check for ${d}`);
    det.onclick = () => openDetails(label, tld, st, det);
    actions.append(saveButton(d, st), det);
    meta.append(actions);
  }
  const li = el("li", { className: focus ? "row focus" : "row" }, el("span", { className: "domain", textContent: d }), meta);
  if (focus) {
    li.append(el("p", { className: "row-explain" },
      `From the scan of ${scanTime(state.stats.updated)} (${ageText(state.stats.updated)}). This site is a static snapshot; `,
      el("a", { href: LIVE_DOCS, textContent: "check live" }),
      ` with `, el("code", { textContent: `shortfree check ${d}` }), ` or the `,
      el("a", { href: MCP_DOCS, textContent: "MCP server" }), `, which ask the registry directly.`));
  }
  return li;
}

let searchSeq = 0;
async function search() {
  const raw = $("#q").value;
  const out = $("#results");
  const seq = ++searchSeq;
  const { label, tld } = parseQuery(raw);
  writeURL(false);
  if (!label) return out.replaceChildren();
  if (!valid(label)) {
    out.replaceChildren(el("p", { className: "msg", textContent: `“${label}” is not a 1–4 character label. Use a–z, 0–9 and hyphens that are not at the start or end.` }));
    return;
  }
  if (!state.stats) return;
  const order = state.stats.order;
  const known = !tld || order.includes(tld);
  if (label.length === 4) out.setAttribute("aria-busy", "true");
  const results = await Promise.all(order.map(async (t) => [t, await status(label, t)]));
  out.removeAttribute("aria-busy");
  if (seq !== searchSeq) return;
  const free = results.filter(([, st]) => isFree(st)).length;
  const verified = results.filter(([, st]) => st === "free").length;
  const kids = [];
  if (tld && !known) {
    kids.push(el("p", { className: "msg" }, `.${tld} is not part of the daily scan (${order.map((t) => "." + t).join(" ")}). `,
      el("a", { href: LIVE_DOCS, textContent: "Check it live" }), ` with `, el("code", { textContent: `shortfree check ${label}.${tld}` }), "."));
  }
  const focusRow = known && tld ? results.find(([t]) => t === tld) : null;
  const rest = focusRow ? results.filter(([t]) => t !== tld) : results;
  if (focusRow) kids.push(el("ul", { className: "rows", style: "border-top:0" }, resultRow(label, tld, focusRow[1], true)));
  const tally = label.length === 4 ? `${free} of ${order.length} likely free (DNS check only)` : `${verified} of ${order.length} verified free` + (free > verified ? `, ${free - verified} unverified` : "");
  kids.push(el("div", { className: "result-head" },
    el("h2", { className: "sr-only", textContent: `Results for ${label}` }),
    el("p", { textContent: `${focusRow ? "Same name on the other TLDs" : `“${label}” on ${order.length} TLDs`} · ${tally}` })));
  kids.push(el("ul", { className: "rows" }, ...rest.map(([t, st]) => resultRow(label, t, st, false))));
  out.replaceChildren(...kids);
}

/* ---------- URL state ---------- */

const KINDS = ["all", "letters", "digits", "mixed"];
const VIEWS = ["az", "rank", "pron", "rep"];
function readURL() {
  const u = new URLSearchParams(location.search);
  const legacy = location.hash.match(/^#q=([^&]+)/); // links from before ?q=
  const r = { q: u.get("q") ?? (legacy ? decodeURIComponent(legacy[1]) : null) };
  for (const k of ["tld", "len", "kind", "view", "ver", "has"]) r[k] = u.get(k);
  return r;
}
function applyURL() {
  const u = readURL();
  const order = state.stats.order;
  const browse = ["tld", "len", "kind", "view", "ver", "has"].some((k) => u[k] != null);
  state.touched = browse;
  const ch = {};
  if (u.tld && order.includes(u.tld)) ch.tld = u.tld;
  if (["1", "2", "3", "4"].includes(u.len)) ch.len = Number(u.len);
  ch.kind = KINDS.includes(u.kind) ? u.kind : null;
  state.kindSet = ch.kind != null;
  if (ch.kind == null) ch.kind = (ch.len ?? state.len) >= 3 ? "letters" : "all";
  ch.view = VIEWS.includes(u.view) ? u.view : "az";
  ch.ver = u.ver === "all" ? "all" : "verified";
  ch.has = u.has && /^[a-z0-9-]{1,4}$/.test(u.has.toLowerCase()) ? u.has.toLowerCase() : "";
  Object.assign(state, ch);
  $("#f-contains").value = state.has;
  if (u.q != null) $("#q").value = u.q;
}
function writeURL(push) {
  const u = new URLSearchParams();
  const q = $("#q").value.trim().toLowerCase();
  if (q) u.set("q", q);
  if (state.touched && state.tld) {
    u.set("tld", state.tld);
    u.set("len", String(state.len));
    u.set("kind", state.kind);
    if (state.view !== "az") u.set("view", state.view);
    if (state.ver !== "verified" && state.len < 4) u.set("ver", state.ver);
    if (state.has) u.set("has", state.has);
  }
  const list = new URLSearchParams(location.search).get("list");
  if (list) u.set("list", list);
  const s = u.toString().replace(/%2C/g, ",");
  const hash = location.hash && !location.hash.includes("=") ? location.hash : "";
  const url = location.pathname + (s ? `?${s}` : "") + hash;
  if (url !== location.pathname + location.search + location.hash) history[push ? "pushState" : "replaceState"](null, "", url);
}

/* ---------- matrix + index ---------- */

// renderTLDIndex lists the static per-TLD pages (written by "shortfree
// pages") for every TLD in stats.json; the HTML holds a no-JS fallback.
function renderTLDIndex() {
  $("#tld-index").replaceChildren(...state.stats.order.map((tld) => {
    const li = el("li", {}, el("a", { href: `${tld}/`, textContent: `.${tld}` }));
    for (const n of [1, 2, 3, 4]) li.append(" ", el("a", { href: `${tld}/${n}-letter/`, textContent: String(n) }));
    return li;
  }));
}

function renderMatrix() {
  const tb = $("#matrix tbody");
  tb.replaceChildren(...state.stats.order.map((tld) => {
    const L = state.stats.tlds[tld] || {};
    const cells = [1, 2, 3, 4].map((n) => {
      const s = L[n] || {};
      const count = n === 4 ? s.free || 0 : s.available || 0;
      const unv = n === 4 ? 0 : s.unverified || 0;
      const total = count + unv;
      // Accessible name = visible text plus hidden context (WCAG 2.5.3).
      const btn = el("button", { type: "button", className: "cell" }, el("span", { className: "sr-only", textContent: "Browse " }));
      btn.append(total ? fmt(count) : el("span", { className: "zero", textContent: "0" }));
      if (unv) btn.append(el("span", { className: "q", textContent: ` +${fmt(unv)}?` }));
      btn.append(el("span", { className: "sr-only", textContent: ` ${n === 4 ? "likely free (DNS only)" : "verified free"} ${n} character .${tld} names` + (unv ? ` (+${fmt(unv)} awaiting RDAP)` : "") }));
      const pct = s.valid ? Math.min(100, (100 * total) / s.valid) : 0;
      btn.append(el("span", { className: "bar", ariaHidden: "true" }, el("i", { style: `width:${pct.toFixed(2)}%` })));
      btn.onclick = () => {
        setFilter({ tld, len: n });
        $("#browse").scrollIntoView();
        $("#browse-h").focus({ preventScroll: true });
      };
      const td = el("td", {}, btn);
      td.dataset.label = n === 1 ? "1 char" : `${n} chars`;
      return td;
    });
    const drop = state.sets[tld]?.drops.length || 0;
    const dtd = el("td", {}, el("span", { className: drop ? "drop-n" : "drop-n zero", textContent: fmt(drop) }));
    dtd.dataset.label = "Dropping";
    return el("tr", {}, el("th", { scope: "row" }, el("a", { href: `${tld}/`, textContent: `.${tld}` })), ...cells, dtd);
  }));
}

/* ---------- browse ---------- */

function seg(id, key, options) {
  const fs = $(id);
  const box = el("div", { className: "seg-opts" }, ...options.map(([value, text, label]) => {
    const input = el("input", { type: "radio", name: key, value: String(value) });
    input.onchange = () => setFilter({ [key]: key === "len" ? Number(value) : value });
    return el("label", { className: "seg-opt" }, input, el("span", {}, text, label ? el("span", { className: "sr-only", textContent: ` (${label})` }) : null));
  }));
  fs.querySelector(".seg-opts")?.remove();
  fs.append(box);
}
function syncControls() {
  for (const k of ["tld", "len", "kind", "view", "ver"]) {
    for (const r of document.querySelectorAll(`input[name="${k}"]`)) r.checked = r.value === String(state[k]);
  }
  $("#f-ver").hidden = state.len === 4;
  renderViews();
}
// setFilter applies user filter changes. 3-4 character views default to
// letters only until the user picks a character class themselves.
function setFilter(changes, { push = true, fromUser = true } = {}) {
  if ("kind" in changes && fromUser) state.kindSet = true;
  if ("len" in changes && !("kind" in changes) && !state.kindSet) changes.kind = changes.len >= 3 ? "letters" : "all";
  Object.assign(state, changes);
  syncControls();
  if (fromUser) { state.touched = true; writeURL(push); }
  buildList();
}

const VIEW_NOTE = {
  az: "a–z, then 0–9, then hyphens.",
  rank: "Ranked by a simple pattern score: pronounceable letter names first, then other letters-only names, all-digit names and mixed names; names built from one or two characters get a small boost, hyphens and 0/1 next to letters rank last. It says nothing about value or demand.",
  pron: "Pronounceable: letters only, alternating consonants and vowels (a e i o u): CV, VC, CVC, VCV, CVCV, VCVC or CVVC. A rule, not a dictionary.",
  rep: "Repeats: one character repeated (aa, 777) or, for 3–4 characters, only two different characters (aba, abab, 7007).",
};
const KIND_TEXT = { letters: "letters only", digits: "digits only", mixed: "mixed", all: "" };

// Curated starting points. Each is a plain, shareable link.
function renderViews() {
  const tld = state.tld;
  if (!tld) return;
  const views = [
    ["CVC / VCV 3-letter", { len: 3, kind: "letters", view: "pron" }],
    ["4-letter pronounceable", { len: 4, kind: "letters", view: "pron" }],
    ["3-letter ranked", { len: 3, kind: "letters", view: "rank" }],
    ["All-digit 3", { len: 3, kind: "digits", view: "az" }],
    ["4-char repeats", { len: 4, kind: "all", view: "rep" }],
  ];
  $("#views").replaceChildren(...views.map(([text, v]) => {
    const href = `?tld=${tld}&len=${v.len}&kind=${v.kind}` + (v.view !== "az" ? `&view=${v.view}` : "") + "#browse";
    const a = el("a", { href, textContent: text });
    if (state.len === v.len && state.kind === v.kind && state.view === v.view && !state.has) a.setAttribute("aria-current", "true");
    a.onclick = (e) => {
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.button) return;
      e.preventDefault();
      state.kindSet = true;
      $("#f-contains").value = "";
      setFilter({ ...v, has: "" });
    };
    return el("li", {}, a);
  }));
}

let listSeq = 0;
async function buildList() {
  const seq = ++listSeq;
  const t0 = performance.now();
  const { tld, len, kind, view, ver, has } = state;
  const kn = KIND_N[kind];
  const want = (p) => (kind === "all" || (p & 3) === kn) && (view !== "pron" || p & 4) && (view !== "rep" || p & 8);
  const res = { len, tld, n: 0, unvHidden: 0, error: null };
  if (len < 4) {
    const by = byLen(tld, len);
    const items = [];
    const take = (arr, u) => {
      for (const x of arr) {
        if (has && !x.s.includes(has)) continue;
        if (want(x.p)) items.push([x.s, u, SCORE[x.p], x.i]);
      }
    };
    take(by[0], false);
    const before = items.length;
    take(by[1], true);
    if (ver === "verified") { res.unvHidden = items.length - before; items.length = before; }
    if (view === "rank") items.sort((x, y) => y[2] - x[2] || x[3] - y[3]);
    else if (ver === "all") items.sort((x, y) => x[3] - y[3]);
    res.n = items.length;
    res.at = (j) => items[j];
    res.lines = (f) => items.map(f);
  } else {
    $("#count").textContent = `Loading 4 character data for .${tld}…`;
    $("#list").setAttribute("aria-busy", "true");
    const F = await four(tld);
    if (seq !== listSeq) return;
    $("#list").removeAttribute("aria-busy");
    if (!F.idx) res.error = F.error;
    else {
      const M = meta4();
      const idx = F.idx;
      const out = new Uint32Array(idx.length);
      const hc = [...has].map((c) => ALPHA.indexOf(c));
      const m = hc.length;
      const d = new Uint8Array(4);
      let k = 0;
      for (let j = 0; j < idx.length; j++) {
        const i = idx[j];
        if (!want(M[i])) continue;
        if (m) {
          d[0] = (i / 50653) | 0; d[1] = ((i / 1369) | 0) % 37; d[2] = ((i / 37) | 0) % 37; d[3] = i % 37;
          let hit = false;
          for (let s = 0; s + m <= 4 && !hit; s++) {
            let ok = true;
            for (let t = 0; t < m; t++) if (d[s + t] !== hc[t]) { ok = false; break; }
            hit = ok;
          }
          if (!hit) continue;
        }
        out[k++] = i;
      }
      let list = out.subarray(0, k);
      if (view === "rank") {
        // counting sort by score, highest first, a-z within a score
        const cnt = new Uint32Array(32);
        for (let j = 0; j < k; j++) cnt[SCORE[M[list[j]]]]++;
        const start = new Uint32Array(32);
        for (let s = 30, acc = 0; s >= 0; s--) { start[s] = acc; acc += cnt[s]; }
        const sorted = new Uint32Array(k);
        for (let j = 0; j < k; j++) sorted[start[SCORE[M[list[j]]]]++] = list[j];
        list = sorted;
      }
      res.n = k;
      res.at = (j) => [decode4(list[j]), false];
      res.lines = (f) => { const a = new Array(k); for (let j = 0; j < k; j++) a[j] = f([decode4(list[j]), false]); return a; };
    }
  }
  if (seq !== listSeq) return;
  perf(`list ${tld} ${len} ${kind} ${view}${has ? " has" : ""}`, t0);
  state.res = res;
  state.shown = PAGE;
  renderList();
}

function describe() {
  const { tld, len, kind, view, has, ver } = state;
  const what = len === 4 ? "likely free" : ver === "verified" ? "verified free" : "free or unverified";
  const bits = [KIND_TEXT[kind], view === "pron" ? "pronounceable" : view === "rep" ? "repeats" : "", has ? `containing “${has}”` : ""].filter(Boolean);
  return { what, label: `${len} character .${tld} names`, bits };
}
function renderList(append) {
  const res = state.res;
  if (!res) return;
  const { tld, len, shown } = state;
  const count = $("#count");
  const retry = $("#retry4");
  retry.hidden = true;
  $("#view-note").textContent = VIEW_NOTE[state.view];
  if (res.error) {
    count.textContent = res.error === "missing"
      ? `The 4 character data for .${tld} is not published yet.`
      : `The 4 character data for .${tld} did not load.`;
    retry.hidden = res.error === "missing";
    $("#list").replaceChildren();
    $("#more").hidden = true;
    $("#dl").hidden = $("#dl-csv").hidden = true;
    return;
  }
  const { what, label, bits } = describe();
  const filters = bits.length ? ` · ${bits.join(" · ")}` : "";
  let text;
  if (!res.n) {
    text = `No ${what} ${label}${filters}.`;
    if (res.unvHidden) text += ` ${fmt(res.unvHidden)} unverified ${res.unvHidden === 1 ? "name matches" : "names match"}; choose “+ unverified” to see ${res.unvHidden === 1 ? "it" : "them"}.`;
    else if (bits.length) text += " Try “all” characters, another view or clear “Contains”.";
  } else {
    text = `${fmt(res.n)} ${what} ${label}${filters}`;
    if (res.unvHidden) text += ` · ${fmt(res.unvHidden)} unverified hidden`;
    if (len < 4 && state.ver === "all") text += " · unverified marked ?";
    if (len === 4) text += " · DNS check only, not RDAP-verified";
    if (premiumLikely("x".repeat(len), tld)) text += " · often reserved or premium at this length";
  }
  const cov = coverageFor(tld, len);
  if (cov && cov.pct != null && cov.pct < 0.995) text += ` · scan coverage ${(cov.pct * 100).toFixed(1)}%, names not answered are missing`;
  if (cov && cov.at && hoursSince(cov.at) > STALE_HOURS) text += ` · .${tld} last scanned ${ageText(cov.at)}`;
  count.textContent = text;

  const from = append ? $("#list").children.length : 0;
  const to = Math.min(shown, res.n);
  const frag = document.createDocumentFragment();
  for (let j = from; j < to; j++) {
    const [s, u] = res.at(j);
    const d = `${s}.${tld}`;
    const b = el("button", { type: "button", className: "pick", textContent: d });
    if (u) b.setAttribute("aria-label", `${d}, awaiting RDAP verification`);
    const li = el("li", { className: (u ? "unv" : "") + (isSaved(d) ? " saved" : "") }, b);
    li.dataset.d = d;
    li.dataset.st = len === 4 ? "free4" : u ? "unv" : "free";
    frag.append(li);
  }
  if (append) $("#list").append(frag);
  else $("#list").replaceChildren(frag);
  $("#more").hidden = shown >= res.n;
  $("#more").textContent = `Show ${fmt(Math.min(PAGE, res.n - shown))} more`;
  $("#dl").hidden = $("#dl-csv").hidden = !res.n;
}

let objectURL = null;
function download(e, csv) {
  const res = state.res;
  if (!res || !res.n) return e.preventDefault();
  const t0 = performance.now();
  const { tld, len } = state;
  const scan = state.stats.updated;
  const evidence = (u) => (len === 4 ? "dns-only" : u ? "unverified" : "verified");
  const body = csv
    ? "domain,status,evidence,scan\n" + res.lines(([s, u]) => `${s}.${tld},${len === 4 || u ? "likely-free" : "free"},${evidence(u)},${scan}\n`).join("")
    : res.lines(([s]) => `${s}.${tld}\n`).join("");
  if (objectURL) URL.revokeObjectURL(objectURL);
  objectURL = URL.createObjectURL(new Blob([body], { type: csv ? "text/csv" : "text/plain" }));
  const a = e.currentTarget;
  a.href = objectURL;
  const parts = [state.len, state.tld, state.kind !== "all" ? state.kind : "", state.view !== "az" ? state.view : "", state.has, len < 4 && state.ver === "all" ? "with-unverified" : len < 4 ? "verified" : "dns-only"].filter(Boolean);
  a.download = `shortfree-${parts.join("-")}.${csv ? "csv" : "txt"}`;
  perf("download", t0);
}

/* ---------- feeds + freshness ---------- */

const FREED = new Set(["freed", "free", "available", "became_available", "released"]);
function renderFeeds(changes) {
  const freed = [];
  for (const c of Array.isArray(changes) ? changes : []) {
    if (!c || !FREED.has(String(c.event || c.type || "").toLowerCase())) continue;
    const domain = c.domain || (c.name && c.tld ? `${c.name}.${c.tld}` : null);
    if (typeof domain !== "string") continue;
    freed.push([domain, String(c.date || c.at || "").slice(0, 10)]);
    if (freed.length >= 200) break;
  }
  $("#freed").replaceChildren(...(freed.length
    ? freed.map(([domain, date]) => {
      const { label, tld } = parseQuery(domain);
      const b = el("button", { type: "button", className: "pick", textContent: domain });
      b.onclick = async () => openDetails(label, tld, state.stats.order.includes(tld) ? await status(label, tld) : "unknown", b);
      return el("li", {}, b, el("span", { className: "when", textContent: date }));
    })
    : [el("li", { className: "empty", textContent: "Nothing in the last 30 days. The log fills from the second daily scan on." })]));
  const drops = [];
  for (const tld of state.stats.order) for (const d of state.sets[tld]?.drops || []) drops.push([`${d.name}.${tld}`, d.expires]);
  drops.sort((a, b) => a[0].length - b[0].length || (a[0] < b[0] ? -1 : 1));
  $("#dropping").replaceChildren(...(drops.length
    ? drops.slice(0, 300).map(([d, exp]) => el("li", {}, el("span", { textContent: d }),
      el("span", { className: "when" }, statusTag("drop"), typeof exp === "string" && exp && !exp.startsWith("0001") ? ` exp ${exp.slice(0, 10)}` : "")))
    : [el("li", { className: "empty", textContent: "No short names in redemption right now." })]));
}

function renderFreshness() {
  const up = state.stats.updated;
  const stale = hoursSince(up) > STALE_HOURS;
  const age = el("span", { className: "age", textContent: ageText(up) });
  const kids = [el("time", { dateTime: up, textContent: scanTime(up) }), " ", age];
  if (stale) {
    const chip = el("span", { className: "status", textContent: "stale" });
    chip.dataset.state = "dropping";
    kids.push(" ", chip);
  }
  $("#f-scan").replaceChildren(...kids);
  const stl = $("#stale");
  stl.hidden = !stale;
  if (stale) stl.textContent = `The last finished scan is ${ageText(up).replace(" ago", "")} old; it normally runs daily. Names may have changed since, so check live before you buy.`;
  // per-TLD scan times or coverage, when data/coverage.json provides them
  const rows = [];
  for (const tld of state.stats.order) {
    const c = coverageFor(tld);
    if (!c) continue;
    const bits = [];
    if (c.pct != null) bits.push(`${(c.pct * 100).toFixed(c.pct < 1 ? 1 : 0)}%`);
    if (c.at) bits.push(ageText(c.at));
    const warn = (c.pct != null && c.pct < 0.95) || (c.at && hoursSince(c.at) > STALE_HOURS);
    rows.push(el("li", { className: warn ? "warn-text" : "" }, el("span", { className: "mono", textContent: `.${tld}` }), ` ${bits.join(" · ")}`));
  }
  $("#coverage").hidden = !rows.length;
  $("#coverage-list").replaceChildren(...rows);
}

function showError(err) {
  console.error(err);
  $("#error").hidden = false;
  $("#facts").removeAttribute("aria-busy");
  $("#f-scan").textContent = "unavailable";
  $("#count").textContent = "The list is unavailable until the data loads.";
  for (const id of ["#freed", "#dropping"]) $(id).replaceChildren(el("li", { className: "empty", textContent: "Unavailable." }));
}

// Fetch the data after the first paint so the static page renders first.
const afterPaint = () => new Promise((r) => requestAnimationFrame(() => setTimeout(r, 0)));
const afterLoad = () => new Promise((r) => (document.readyState === "complete" ? r() : addEventListener("load", () => r(), { once: true }))).then(afterPaint);

let wired = false;
function wire() {
  if (wired) return;
  wired = true;
  let qt;
  $("#q").addEventListener("input", () => { clearTimeout(qt); qt = setTimeout(search, 60); });
  $("#search").addEventListener("submit", (e) => { e.preventDefault(); clearTimeout(qt); search(); });
  let t;
  $("#f-contains").addEventListener("input", (e) => {
    clearTimeout(t);
    t = setTimeout(() => {
      const v = e.target.value.trim().toLowerCase();
      if (v && !/^[a-z0-9-]{1,4}$/.test(v)) return;
      setFilter({ has: v }, { push: false });
    }, 150);
  });
  $("#more").onclick = () => { state.shown += PAGE; renderList(true); };
  $("#retry4").onclick = () => buildList();
  $("#dl").onclick = (e) => download(e, false);
  $("#dl-csv").onclick = (e) => download(e, true);
  $("#list").addEventListener("click", (e) => {
    const b = e.target.closest("button.pick");
    if (!b) return;
    const li = b.closest("li");
    const { label, tld } = parseQuery(li.dataset.d);
    openDetails(label, tld, li.dataset.st, b);
  });
  const dlg = $("#details");
  $("#d-close").onclick = closeDetails;
  dlg.addEventListener("click", (e) => { if (e.target === dlg) closeDetails(); });
  dlg.addEventListener("close", () => {
    if (detailsOpener && detailsOpener.isConnected) detailsOpener.focus();
    detailsOpener = null;
  });
  $("#sl-copy").onclick = (e) => copyText(shortlist.map((x) => x.d).join("\n"), e.currentTarget, "Copy all");
  $("#sl-share").onclick = (e) => copyText(shareURL(shortlist.map((x) => x.d)), e.currentTarget, "Copy share link");
  $("#sl-dl").onclick = (e) => {
    if (objectURL) URL.revokeObjectURL(objectURL);
    objectURL = URL.createObjectURL(new Blob([shortlist.map((x) => `${x.d}\n`).join("")], { type: "text/plain" }));
    e.currentTarget.href = objectURL;
  };
  $("#sl-clear").onclick = () => {
    if (!confirm(`Remove all ${shortlist.length} names from this shortlist?`)) return;
    shortlist = [];
    shortlistSaved = store.set(SHORTLIST_KEY, shortlist);
    syncSaveButtons();
    renderShortlist();
    $("#sl-h").focus();
  };
  $("#sl-keep").onclick = () => {
    for (const d of sharedFromURL()) if (!isSaved(d) && shortlist.length < SHORTLIST_MAX) shortlist.push({ d, s: "", at: state.stats.updated });
    shortlistSaved = store.set(SHORTLIST_KEY, shortlist);
    syncSaveButtons();
    renderShortlist();
  };
  addEventListener("popstate", () => {
    if (!state.stats) return;
    applyURL();
    syncControls();
    buildList();
    search();
    renderShortlist();
  });
  addEventListener("storage", (e) => {
    if (e.key !== SHORTLIST_KEY) return;
    loadShortlist();
    syncSaveButtons();
    renderShortlist();
  });
}

async function main() {
  $("#error").hidden = true;
  loadShortlist();
  try {
    await afterLoad();
    const [stats, avail, changes, cov] = await Promise.all([
      json("data/stats.json"), json("data/available.json"),
      json("data/changes.json").catch(() => []), json("data/coverage.json").catch(() => null),
    ]);
    state.stats = stats;
    state.avail = avail;
    state.cov = cov;
    await afterPaint();
    prepare(avail, stats);
    let total = 0;
    for (const tld of stats.order) for (const n of [1, 2, 3]) total += stats.tlds[tld]?.[n]?.available || 0;
    renderFreshness();
    $("#f-total").textContent = fmt(total);
    $("#f-tlds").textContent = String(stats.order.length);
    $("#facts").removeAttribute("aria-busy");

    renderMatrix();
    renderTLDIndex();
    seg("#f-tld", "tld", stats.order.map((t) => [t, `.${t}`]));
    seg("#f-len", "len", [[1, "1"], [2, "2"], [3, "3"], [4, "4"]]);
    seg("#f-kind", "kind", [["all", "all"], ["letters", "abc", "letters only"], ["digits", "123", "digits only"], ["mixed", "a1-", "mixed"]]);
    seg("#f-view", "view", [["az", "a–z"], ["rank", "ranked"], ["pron", "pronounceable"], ["rep", "repeats"]]);
    seg("#f-ver", "ver", [["verified", "verified"], ["all", "+ unverified"]]);
    const free3 = (t) => (stats.tlds[t]?.[3]?.available || 0) + (stats.tlds[t]?.[3]?.unverified || 0);
    const best = stats.order.reduce((b, t) => (free3(t) > free3(b) ? t : b), stats.order[0]);
    applyURL();
    if (!state.tld || !stats.order.includes(state.tld)) state.tld = best;
    syncControls();
    wire();
    // yield between steps so no single task blocks input for long
    await afterPaint();
    await buildList();
    await afterPaint();
    renderFeeds(changes);
    renderShortlist();
    if ($("#q").value.trim()) search();
  } catch (err) {
    showError(err);
  }
}
$("#retry").onclick = main;
$("#browse-h").tabIndex = -1;
$("#sl-h").tabIndex = -1;
main();
