// shortfree front end: reads the daily data files and answers everything
// client-side. 4 character results come from per-TLD bitmaps (1 bit per name,
// base-37 index over a-z 0-9 -), loaded only when needed.
"use strict";

const ALPHA = "abcdefghijklmnopqrstuvwxyz0123456789-";
const SPACE4 = 37 ** 4;
const PAGE = 600;
const $ = (s) => document.querySelector(s);
const el = (tag, props = {}, ...kids) => {
  const n = Object.assign(document.createElement(tag), props);
  n.append(...kids);
  return n;
};
const fmt = (n) => n.toLocaleString("en-US");
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

async function bitmap(tld) {
  if (!state.bitmaps[tld]) {
    state.bitmaps[tld] = fetch(`data/4/${tld}.bin`).then((r) => (r.ok ? r.arrayBuffer() : null)).then((b) => (b ? new Uint8Array(b) : null));
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
const LABEL = { free: "free", free4: "free (DNS)", unv: "likely free", drop: "dropping", taken: "taken", unknown: "no data" };
const CLS = { free: "free", free4: "free", unv: "unv", drop: "drop", taken: "taken", unknown: "unv" };

let searchSeq = 0;
async function search() {
  const q = $("#q").value.trim().toLowerCase().replace(/^\.+|\..*$/g, "");
  const out = $("#chips");
  const seq = ++searchSeq;
  history.replaceState(null, "", q ? `#q=${q}` : location.pathname);
  if (!q) return out.replaceChildren();
  if (!valid(q)) {
    out.replaceChildren(el("li", { className: "chip" }, el("span", { className: "d", textContent: `${q} is not a valid 1–4 character label` })));
    return;
  }
  const results = await Promise.all(state.stats.order.map(async (tld) => [tld, await status(q, tld)]));
  if (seq !== searchSeq) return;
  out.replaceChildren(...results.map(([tld, st]) => {
    const d = `${q}.${tld}`;
    const free = st === "free" || st === "free4" || st === "unv";
    const tag = el("span", { className: `s ${CLS[st]}`, textContent: LABEL[st] });
    if (free && premiumLikely(q, tld)) tag.append(el("span", { className: "p", textContent: " · premium?", title: "1–2 character names on new TLDs are often reserved or premium-priced" }));
    const inner = [el("span", { className: "d", textContent: d }), tag];
    const li = el("li", { className: `chip ${free ? "free" : ""}` });
    if (free) li.append(el("a", { href: registrar(d), target: "_blank", rel: "noopener", title: `register ${d}`, style: "display:contents" }, ...inner));
    else li.append(...inner);
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
      const btn = el("button", { type: "button", title: `browse ${n} character .${tld}` });
      btn.append(total ? fmt(count) : el("span", { className: "zero", textContent: "0" }));
      if (unv) btn.append(el("span", { className: "q", textContent: ` +${fmt(unv)}?` }));
      const pct = Math.min(100, (100 * total) / s.valid);
      btn.append(el("span", { className: "bar" }, el("i", { style: `width:${pct.toFixed(2)}%` })));
      btn.onclick = () => { state.tld = tld; setFilter("len", n); $("#browse").scrollIntoView({ behavior: "smooth" }); };
      return el("td", {}, btn);
    });
    const drop = state.avail.tlds[tld].dropping.length;
    return el("tr", {}, el("td", { className: "tld", textContent: `.${tld}` }), ...cells,
      el("td", {}, el("span", { className: drop ? "" : "zero", textContent: fmt(drop), style: "display:block;padding:12px 16px" })));
  }));
}

function seg(id, key, options) {
  $(id).replaceChildren(...options.map(([value, text]) => {
    const b = el("button", { type: "button", role: "radio", textContent: text });
    b.dataset.value = value;
    b.onclick = () => setFilter(key, typeof state[key] === "number" ? Number(value) : value);
    return b;
  }));
}
function setFilter(key, value) {
  state[key] = value;
  for (const [id, k] of [["#f-tld", "tld"], ["#f-len", "len"], ["#f-kind", "kind"]]) {
    for (const b of $(id).children) b.setAttribute("aria-checked", String(b.dataset.value === String(state[k])));
  }
  buildList();
}

let listSeq = 0;
async function buildList() {
  const seq = ++listSeq;
  const { tld, len, kind, contains } = state;
  const keep = (s) => (kind === "all" || kindOf(s) === kind) && (!contains || s.includes(contains));
  let items = [];
  if (len < 4) {
    const a = state.avail.tlds[tld];
    items = a.available.filter((s) => s.length === len && keep(s)).map((s) => [s, false]);
    items.push(...a.unverified.filter((s) => s.length === len && keep(s)).map((s) => [s, true]));
    items.sort((x, y) => index(x[0]) - index(y[0])); // a-z, then 0-9, then hyphens
  } else {
    $("#count").textContent = "loading 4 character data…";
    const bm = await bitmap(tld);
    if (seq !== listSeq) return;
    if (bm) {
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
  renderList();
}
function renderList() {
  const { list, shown, tld, len } = state;
  const unv = list.filter((x) => x[1]).length;
  $("#count").textContent = `${fmt(list.length)} free ${len} character .${tld} names` + (unv ? ` (${fmt(unv)} awaiting RDAP, marked ?)` : "") + (len === 4 ? " · DNS check only" : "") +
    (premiumLikely("x".repeat(len), tld) && list.length ? " · often reserved or premium at this length" : "");
  $("#list").replaceChildren(...list.slice(0, shown).map(([s, u]) => {
    const d = `${s}.${tld}`;
    return el("li", { className: u ? "unv" : "" }, el("a", { href: registrar(d), target: "_blank", rel: "noopener", textContent: d }));
  }));
  $("#more").hidden = shown >= list.length;
  $("#more").textContent = `show ${fmt(Math.min(PAGE, list.length - shown))} more`;
}

function renderFeeds(changes) {
  const freed = changes.filter((c) => c.event === "freed").slice(0, 200);
  $("#freed").replaceChildren(...(freed.length
    ? freed.map((c) => el("li", {}, el("a", { href: registrar(c.domain), target: "_blank", rel: "noopener", textContent: c.domain }), el("span", { className: "when", textContent: c.date })))
    : [el("li", { className: "empty", textContent: "Nothing yet – the change log starts with the second daily scan." })]));
  const drops = [];
  for (const tld of state.stats.order) for (const d of state.avail.tlds[tld].dropping) drops.push([`${d.name}.${tld}`, d.expires]);
  drops.sort((a, b) => a[0].length - b[0].length || (a[0] < b[0] ? -1 : 1));
  $("#dropping").replaceChildren(...(drops.length
    ? drops.slice(0, 300).map(([d, exp]) => el("li", {}, el("span", { textContent: d }), el("span", { className: "when", textContent: exp && !exp.startsWith("0001") ? `exp ${exp.slice(0, 10)}` : "" })))
    : [el("li", { className: "empty", textContent: "No short names in redemption right now." })]));
}

async function main() {
  try {
    const [stats, avail, changes] = await Promise.all([json("data/stats.json"), json("data/available.json"), json("data/changes.json").catch(() => [])]);
    state.stats = stats;
    state.avail = avail;
    for (const tld of stats.order) {
      const a = avail.tlds[tld];
      state.sets[tld] = { free: new Set(a.available), unv: new Set(a.unverified), drop: new Set(a.dropping.map((d) => d.name)) };
    }
    let total = 0;
    for (const tld of stats.order) for (const n of [1, 2, 3]) total += stats.tlds[tld][n].available || 0;
    $("#scan").replaceChildren("last scan ", el("b", { textContent: new Date(stats.updated).toISOString().slice(0, 16).replace("T", " ") + " UTC" }), ` · ${fmt(total)} verified free`);

    renderMatrix();
    seg("#f-tld", "tld", stats.order.map((t) => [t, `.${t}`]));
    seg("#f-len", "len", [[1, "1"], [2, "2"], [3, "3"], [4, "4"]]);
    seg("#f-kind", "kind", [["all", "all"], ["letters", "abc"], ["digits", "123"], ["mixed", "a1-"]]);
    const free3 = (t) => (stats.tlds[t][3].available || 0) + (stats.tlds[t][3].unverified || 0);
    state.tld = stats.order.reduce((best, t) => (free3(t) > free3(best) ? t : best), stats.order[0]);
    setFilter("len", 3);
    renderFeeds(changes);

    $("#q").addEventListener("input", search);
    $("#f-contains").addEventListener("input", (e) => { state.contains = e.target.value.trim().toLowerCase(); buildList(); });
    $("#more").onclick = () => { state.shown += PAGE; renderList(); };
    $("#dl").onclick = (e) => {
      const blob = new Blob([state.list.map(([s]) => `${s}.${state.tld}\n`).join("")], { type: "text/plain" });
      e.currentTarget.href = URL.createObjectURL(blob);
      e.currentTarget.download = `shortfree-${state.len}-${state.tld}.txt`;
    };
    const m = location.hash.match(/q=([a-z0-9-]{1,4})/);
    if (m) { $("#q").value = m[1]; search(); }
  } catch (err) {
    $("#scan").textContent = "data not available yet";
    console.error(err);
  }
}
main();
