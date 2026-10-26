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
const $ = (s) => document.querySelector(s);
const el = (tag, props = {}, ...kids) => {
  const n = Object.assign(document.createElement(tag), props);
  n.append(...kids.filter((k) => k != null && k !== false));
  return n;
};
const fmt = (n) => n.toLocaleString("en-US");
const scanTime = (iso) => new Date(iso).toISOString().slice(0, 16).replace("T", " ") + " UTC";
// Registries of new gTLDs usually hold back 1-2 character names (reserved or
// premium) even though RDAP reports them as unregistered.
const premiumLikely = (label, tld) => label.length <= 2 && !["com", "net", "de"].includes(tld);
const registrar = (d) => `https://porkbun.com/checkout/search?q=${encodeURIComponent(d)}`;

const state = { stats: null, avail: null, sets: {}, bitmaps: {}, tld: null, len: 3, kind: "all", contains: "", shown: PAGE, list: [] };

function valid(s) {
  return /^[a-z0-9-]{1,4}$/.test(s) && !s.startsWith("-") && !s.endsWith("-") && !s.includes("--");
}
function index(s) {
  let i = 0;
  for (const c of s) i = i * 37 + ALPHA.indexOf(c);
  return i;
}
function kindOf(s) {
  if (/^[a-z]+$/.test(s)) return "letters";
  if (/^[0-9]+$/.test(s)) return "digits";
  return "mixed";
}

async function json(path) {
  const r = await fetch(path, { cache: "no-cache" });
  if (!r.ok) throw new Error(`${path}: ${r.status}`);
  return r.json();
}

function bitmap(tld) {
  if (!state.bitmaps[tld]) {
    state.bitmaps[tld] = fetch(`data/4/${tld}.bin`)
      .then((r) => (r.ok ? r.arrayBuffer() : null))
      .then((b) => (b ? new Uint8Array(b) : null))
      .catch(() => { delete state.bitmaps[tld]; return null; });
  }
  return state.bitmaps[tld];
}
const bit = (bm, i) => (bm[i >> 3] >> (i & 7)) & 1;

// status of label.tld: free | free4 | unv (DNS-free, RDAP pending) | drop | taken | unknown
async function status(label, tld) {
  if (label.length === 4) {
    const bm = await bitmap(tld);
    if (!bm) return "unknown";
    return bit(bm, index(label)) ? "free4" : "taken";
  }
  const s = state.sets[tld];
  if (s.free.has(label)) return "free";
  if (s.drop.has(label)) return "drop";
  if (s.unv.has(label)) return "unv";
  return "taken";
}
const WORD = { free: "free", free4: "free", unv: "likely free", drop: "dropping", taken: "taken", unknown: "no data" };
const STATE = { free: "free", free4: "free", unv: "unverified", drop: "dropping", taken: "taken", unknown: "unknown" };
const ASIDE = { free4: "DNS check only", unv: "RDAP check queued" };
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

async function copy(text, btn) {
  try {
    await navigator.clipboard.writeText(text);
    btn.textContent = "Copied";
  } catch {
    btn.textContent = "Press ⌘C / Ctrl+C";
    const r = document.createRange();
    r.selectNodeContents(btn.closest(".row").querySelector(".domain"));
    getSelection().removeAllRanges();
    getSelection().addRange(r);
  }
  setTimeout(() => { btn.textContent = "Copy"; }, 2000);
}

function resultRow(label, tld, st, focus) {
  const d = `${label}.${tld}`;
  const free = isFree(st);
  const meta = el("div", { className: "meta" }, statusTag(st));
  const notes = [ASIDE[st], free && premiumLikely(label, tld) ? "often reserved or premium" : ""].filter(Boolean);
  if (notes.length) meta.append(el("span", { className: "aside", textContent: notes.join(" · ") }));
  if (free) {
    const c = el("button", { type: "button", className: focus ? "btn btn-free" : "btn", textContent: "Copy" });
    c.setAttribute("aria-label", `Copy ${d}`);
    c.onclick = () => copy(d, c);
    const reg = el("a", { className: "btn", href: registrar(d), target: "_blank", rel: "noopener", textContent: "Register ↗" });
    reg.setAttribute("aria-label", `Register ${d} at Porkbun (opens in a new tab)`);
    meta.append(el("span", { className: "actions" }, c, reg));
  }
  const li = el("li", { className: focus ? "row focus" : "row" }, el("span", { className: "domain", textContent: d }), meta);
  if (focus) {
    li.append(el("p", { className: "row-explain" },
      `From the scan of ${scanTime(state.stats.updated)}. This site is a static snapshot; `,
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
  history.replaceState(null, "", raw.trim() ? `#q=${encodeURIComponent(raw.trim().toLowerCase())}` : location.pathname + location.search);
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
  const kids = [];
  if (tld && !known) {
    kids.push(el("p", { className: "msg" }, `.${tld} is not part of the daily scan (${order.map((t) => "." + t).join(" ")}). `,
      el("a", { href: LIVE_DOCS, textContent: "Check it live" }), ` with `, el("code", { textContent: `shortfree check ${label}.${tld}` }), "."));
  }
  const focusRow = known && tld ? results.find(([t]) => t === tld) : null;
  const rest = focusRow ? results.filter(([t]) => t !== tld) : results;
  if (focusRow) kids.push(el("ul", { className: "rows", style: "border-top:0" }, resultRow(label, tld, focusRow[1], true)));
  kids.push(el("div", { className: "result-head" },
    el("h2", { className: "sr-only", textContent: `Results for ${label}` }),
    el("p", { textContent: `${focusRow ? "Same name on the other TLDs" : `“${label}” on ${order.length} TLDs`} · ${free} of ${order.length} free` })));
  kids.push(el("ul", { className: "rows" }, ...rest.map(([t, st]) => resultRow(label, t, st, false))));
  out.replaceChildren(...kids);
}

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
    const L = state.stats.tlds[tld];
    const cells = [1, 2, 3, 4].map((n) => {
      const s = L[n];
      const count = n === 4 ? s.free || 0 : s.available || 0;
      const unv = n === 4 ? 0 : s.unverified || 0;
      const total = count + unv;
      // Accessible name = visible text plus hidden context (WCAG 2.5.3).
      const btn = el("button", { type: "button", className: "cell" }, el("span", { className: "sr-only", textContent: "Browse " }));
      btn.append(total ? fmt(count) : el("span", { className: "zero", textContent: "0" }));
      if (unv) btn.append(el("span", { className: "q", textContent: ` +${fmt(unv)}?` }));
      btn.append(el("span", { className: "sr-only", textContent: ` free ${n} character .${tld} names` + (unv ? ` (+${fmt(unv)} awaiting RDAP)` : "") }));
      const pct = Math.min(100, (100 * total) / s.valid);
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
    const drop = state.avail.tlds[tld].dropping.length;
    const dtd = el("td", {}, el("span", { className: drop ? "drop-n" : "drop-n zero", textContent: fmt(drop) }));
    dtd.dataset.label = "Dropping";
    return el("tr", {}, el("th", { scope: "row" }, el("a", { href: `${tld}/`, textContent: `.${tld}` })), ...cells, dtd);
  }));
}

function seg(id, key, options) {
  const fs = $(id);
  const box = el("div", { className: "seg-opts" }, ...options.map(([value, text, label]) => {
    const input = el("input", { type: "radio", name: key, value: String(value) });
    input.onchange = () => setFilter({ [key]: typeof state[key] === "number" ? Number(value) : value });
    return el("label", { className: "seg-opt" }, input, el("span", {}, text, label ? el("span", { className: "sr-only", textContent: ` (${label})` }) : null));
  }));
  fs.querySelector(".seg-opts")?.remove();
  fs.append(box);
}
function setFilter(changes) {
  Object.assign(state, changes);
  for (const k of ["tld", "len", "kind"]) {
    for (const r of document.querySelectorAll(`input[name="${k}"]`)) r.checked = r.value === String(state[k]);
  }
  buildList();
}

let listSeq = 0;
async function buildList() {
  const seq = ++listSeq;
  const { tld, len, kind, contains } = state;
  const keep = (s) => (kind === "all" || kindOf(s) === kind) && (!contains || s.includes(contains));
  let items = [];
  let missing = false;
  if (len < 4) {
    const a = state.avail.tlds[tld];
    items = a.available.filter((s) => s.length === len && keep(s)).map((s) => [s, false]);
    items.push(...a.unverified.filter((s) => s.length === len && keep(s)).map((s) => [s, true]));
    items.sort((x, y) => index(x[0]) - index(y[0])); // a-z, then 0-9, then hyphens
  } else {
    $("#count").textContent = `Loading 4 character data for .${tld}…`;
    $("#list").setAttribute("aria-busy", "true");
    const bm = await bitmap(tld);
    if (seq !== listSeq) return;
    $("#list").removeAttribute("aria-busy");
    if (!bm) missing = true;
    else {
      for (let i = 0; i < SPACE4; i++) {
        if (!bit(bm, i)) continue;
        const s = ALPHA[(i / 50653) | 0] + ALPHA[((i / 1369) | 0) % 37] + ALPHA[((i / 37) | 0) % 37] + ALPHA[i % 37];
        if (keep(s)) items.push([s, false]);
      }
    }
  }
  if (seq !== listSeq) return;
  state.list = items;
  state.shown = PAGE;
  renderList(missing);
}
function renderList(missing) {
  const { list, shown, tld, len, kind, contains } = state;
  const unv = list.filter((x) => x[1]).length;
  let text;
  if (missing) text = `The 4 character data for .${tld} is not published yet.`;
  else if (!list.length) text = `No free ${len} character .${tld} names` + (kind !== "all" || contains ? " match these filters. Try “all” characters or clear “Contains”." : " right now.");
  else {
    text = `${fmt(list.length)} free ${len} character .${tld} names` + (unv ? ` (${fmt(unv)} awaiting RDAP, marked ?)` : "") + (len === 4 ? " · DNS check only" : "") +
      (premiumLikely("x".repeat(len), tld) ? " · often reserved or premium at this length" : "");
  }
  $("#count").textContent = text;
  $("#list").replaceChildren(...list.slice(0, shown).map(([s, u]) => {
    const d = `${s}.${tld}`;
    const a = el("a", { href: registrar(d), target: "_blank", rel: "noopener", textContent: d });
    if (u) a.setAttribute("aria-label", `${d}, awaiting RDAP verification`);
    return el("li", { className: u ? "unv" : "" }, a);
  }));
  $("#more").hidden = shown >= list.length;
  $("#more").textContent = `Show ${fmt(Math.min(PAGE, list.length - shown))} more`;
  $("#dl").hidden = !list.length;
}

function renderFeeds(changes) {
  const freed = changes.filter((c) => c.event === "freed").slice(0, 200);
  $("#freed").replaceChildren(...(freed.length
    ? freed.map((c) => el("li", {}, el("a", { href: registrar(c.domain), target: "_blank", rel: "noopener", textContent: c.domain }), el("span", { className: "when", textContent: c.date })))
    : [el("li", { className: "empty", textContent: "Nothing in the last 30 days. The log fills from the second daily scan on." })]));
  const drops = [];
  for (const tld of state.stats.order) for (const d of state.avail.tlds[tld].dropping) drops.push([`${d.name}.${tld}`, d.expires]);
  drops.sort((a, b) => a[0].length - b[0].length || (a[0] < b[0] ? -1 : 1));
  $("#dropping").replaceChildren(...(drops.length
    ? drops.slice(0, 300).map(([d, exp]) => el("li", {}, el("span", { textContent: d }),
      el("span", { className: "when" }, statusTag("drop"), exp && !exp.startsWith("0001") ? ` exp ${exp.slice(0, 10)}` : "")))
    : [el("li", { className: "empty", textContent: "No short names in redemption right now." })]));
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

let wired = false;
async function main() {
  $("#error").hidden = true;
  try {
    await afterPaint();
    const [stats, avail, changes] = await Promise.all([json("data/stats.json"), json("data/available.json"), json("data/changes.json").catch(() => [])]);
    state.stats = stats;
    state.avail = avail;
    for (const tld of stats.order) {
      const a = avail.tlds[tld];
      state.sets[tld] = { free: new Set(a.available), unv: new Set(a.unverified), drop: new Set(a.dropping.map((d) => d.name)) };
    }
    let total = 0;
    for (const tld of stats.order) for (const n of [1, 2, 3]) total += stats.tlds[tld][n].available || 0;
    const t = el("time", { dateTime: stats.updated, textContent: scanTime(stats.updated) });
    $("#f-scan").replaceChildren(t);
    $("#f-total").textContent = fmt(total);
    $("#f-tlds").textContent = String(stats.order.length);
    $("#facts").removeAttribute("aria-busy");

    renderMatrix();
    renderTLDIndex();
    seg("#f-tld", "tld", stats.order.map((t) => [t, `.${t}`]));
    seg("#f-len", "len", [[1, "1"], [2, "2"], [3, "3"], [4, "4"]]);
    seg("#f-kind", "kind", [["all", "all"], ["letters", "abc", "letters only"], ["digits", "123", "digits only"], ["mixed", "a1-", "mixed"]]);
    const free3 = (t) => (stats.tlds[t][3].available || 0) + (stats.tlds[t][3].unverified || 0);
    const best = stats.order.reduce((b, t) => (free3(t) > free3(b) ? t : b), stats.order[0]);
    setFilter({ tld: state.tld && stats.order.includes(state.tld) ? state.tld : best, len: state.len });
    renderFeeds(changes);

    if (!wired) {
      wired = true;
      $("#q").addEventListener("input", search);
      $("#search").addEventListener("submit", (e) => { e.preventDefault(); search(); });
      let t;
      $("#f-contains").addEventListener("input", (e) => {
        clearTimeout(t);
        t = setTimeout(() => { state.contains = e.target.value.trim().toLowerCase(); buildList(); }, 120);
      });
      $("#more").onclick = () => { state.shown += PAGE; renderList(); };
      $("#dl").onclick = (e) => {
        const blob = new Blob([state.list.map(([s]) => `${s}.${state.tld}\n`).join("")], { type: "text/plain" });
        e.currentTarget.href = URL.createObjectURL(blob);
        e.currentTarget.download = `shortfree-${state.len}-${state.tld}.txt`;
      };
    }
    const m = location.hash.match(/q=([^&]+)/);
    if (m) { $("#q").value = decodeURIComponent(m[1]); search(); }
  } catch (err) {
    showError(err);
  }
}
$("#retry").onclick = main;
$("#browse-h").tabIndex = -1;
main();
