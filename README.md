<!-- mcp-name: io.github.maximilianfeix/shortfree -->
<p align="right"><b>English</b> · <a href="README.de.md">Deutsch</a> · <a href="README.zh-CN.md">简体中文</a></p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="brand/banner-dark.svg">
      <img src="brand/banner-light.svg" alt="shortfree: every free 1–4 character domain, verified daily" width="800">
    </picture>
  </a>
</p>

<p align="center">
  Every unregistered 1–4 character domain on <b>.com .net .org .io .ai .dev</b> and <a href="docs/tlds.md">more TLDs</a>,<br>
  checked at the registries' own name servers and RDAP services every day.
</p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/"><img src="https://img.shields.io/badge/Search_the_list-1FC16B?style=for-the-badge" alt="Search the list"></a>
  <a href="#lists"><img src="https://img.shields.io/badge/Download_lists-0B0C0E?style=for-the-badge" alt="Download lists"></a>
  <a href="#try-it"><img src="https://img.shields.io/badge/Install-0B0C0E?style=for-the-badge" alt="Install"></a>
  <a href="#mcp"><img src="https://img.shields.io/badge/MCP-0B0C0E?style=for-the-badge" alt="MCP server"></a>
  <a href="#api"><img src="https://img.shields.io/badge/API-0B0C0E?style=for-the-badge" alt="JSON and CSV API"></a>
</p>

<p align="center">
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml/badge.svg" alt="daily scan"></a>
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml/badge.svg" alt="ci"></a>
  <a href="https://github.com/maximilianfeix/shortfree/releases"><img src="https://img.shields.io/github/v/release/maximilianfeix/shortfree?color=1FC16B" alt="latest release"></a>
  <a href="https://pkg.go.dev/github.com/maximilianfeix/shortfree"><img src="https://pkg.go.dev/badge/github.com/maximilianfeix/shortfree.svg" alt="Go reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/maximilianfeix/shortfree?color=0B0C0E" alt="MIT license"></a>
  <a href="https://maximilianfeix.github.io/shortfree/api/v1/badge.json"><img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fmaximilianfeix.github.io%2Fshortfree%2Fapi%2Fv1%2Fbadge.json" alt="number of RDAP-verified free short domains"></a>
</p>

<!-- demo -->

<p align="center">
  <a href="#try-it">Try it</a> ·
  <a href="#numbers">Numbers</a> ·
  <a href="#lists">Lists</a> ·
  <a href="#limits">Limits</a> ·
  <a href="#how-it-works">How it works</a> ·
  <a href="#mcp">MCP</a> ·
  <a href="#api">API</a> ·
  <a href="#cli">CLI</a> ·
  <a href="#faq">FAQ</a> ·
  <a href="#contributing">Contributing</a>
</p>

<a id="try-it"></a>
## Try it

**No install:** open [maximilianfeix.github.io/shortfree](https://maximilianfeix.github.io/shortfree/), type 1–4 characters and see the status on every scanned TLD. Filter the full list by TLD, length and characters.

**Command line** (Go 1.27+):

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

Binaries for Linux, macOS and Windows are attached to each [release](https://github.com/maximilianfeix/shortfree/releases).

<a id="numbers"></a>
## Current numbers

<!-- shortfree:start -->

_Last scan: **2026-10-29 03:25 UTC**. Updated automatically by GitHub Actions._

| TLD | 1 char | 2 chars | 3 chars | 4 chars¹ | dropping² | list |
|---|--:|--:|--:|--:|--:|---|
| `.com` | 0 | 0 | 0 | 664,153 | 2 | [txt](data/txt/com.txt) |
| `.net` | 0 | 13 | 4,533 | 1,467,126 | 219 | [txt](data/txt/net.txt) |
| `.de` | 0 | 0 | 0 | 1,468,444 | 61 | [txt](data/txt/de.txt) |
| `.ai` | 0 | 0 | 7,167 (+16,388?) | 519,636 (partial) | 25 | [txt](data/txt/ai.txt) |
| `.io` | 0 | 35 | 7,108 (+14,556?) | 541,499 (partial) | 34 | [txt](data/txt/io.txt) |
| `.dev` | 25 | 680 | 10,065 (+24,488?) | 1,719,356 | 18 | [txt](data/txt/dev.txt) |
| `.app` | 2 | 349 | 10,286 (+10,762?) | 1,681,992 | 100 | [txt](data/txt/app.txt) |
| `.tools` | 11 | 913 | 6,269 (+38,466?) | 656,101 (partial) | 5 | [txt](data/txt/tools.txt) |
| `.co` | 0 | 47 | 22,999 | 1,713,480 | 364 | [txt](data/txt/co.txt) |
| `.org` | 0 | 78 | 784 (+10,126?) | 646,541 (partial) | 13 | [txt](data/txt/org.txt) |
| `.xyz` | 9 | 577 | 1,697 (+4,624?) | 1,318,838 | 6944 | [txt](data/txt/xyz.txt) |
| `.so` | 0 | 0 | 14,306 (+30,973?) | 1,769,171 | 11 | [txt](data/txt/so.txt) |
| `.to` | 0 | 0 | 13,415 (+27,126?) | 1,763,093 | 12 | [txt](data/txt/to.txt) |
| `.cc` | 0 | 16 | 427 | 1,443,569 | 487 | [txt](data/txt/cc.txt) |
| `.tv` | 2 | 96 (+727?) | 396 (+40,653?) | 1,700,885 (partial) | 0 | [txt](data/txt/tv.txt) |
| `.fm` | 0 | 1,066 | 8,464 (+37,357?) | 1,770,009 | 0 | [txt](data/txt/fm.txt) |
| `.is` | 0 | 62 (+26?) | 476 (+41,416?) | 1,765,885 | 0 | [txt](data/txt/is.txt) |
| `.ly` | 0 | 0 | 0 | 1,768,838 | 0 | [txt](data/txt/ly.txt) |
| `.run` | 0 | 688 | 6,484 (+36,999?) | 655,545 (partial) | 21 | [txt](data/txt/run.txt) |
| `.inc` | 8 | 1,147 | 8,346 (+38,809?) | 1,772,087 | 4 | [txt](data/txt/inc.txt) |

¹ not in the zone (DNS check only, not RDAP-verified). Full lists: [latest data release](../../releases/tag/data).  
² registered, but in redemption or pending delete. The RDAP expiration date is not a drop date.  
`?` = not delegated in DNS, but no current RDAP answer (check queued, or the last one is older than 14 days).  
Counted names are not registered (RDAP 404). That does not prove a registrar will sell them: registries reserve or premium-price some names. See [accuracy](docs/accuracy.md).

### Recently freed

Names the registry reported as registered or dropping that are now not registered (RDAP 404).

`10p.co` · `11k.co` · `17b.co` · `23t.co` · `29k.co` · `2ds.co` · `2mc.co` · `2pk.co` · `3cl.co` · `3tc.co` · `4be.co` · `4mc.co` · `4tv.co` · `50x.co` · `77m.co` · `b89.co` · `l-n.cc` · `mf8.co` · `y27.net` · `27j.net` · `2q8.net` · `4f2.net` · `76d.net` · `8rw.net` · `94n.net` · `b8g.net` · `b9q.net` · `g5c.net` · `g7d.net` · `j5n.net`

### Free 1–2 character domains

**.net** – `0v.net` `2d.net` `4d.net` `68.net` `9y.net` `c0.net` `jv.net` `k4.net` `o4.net` `ol.net` `tj.net` `ue.net` `ye.net`

**.io** – `03.io` `1q.io` `1s.io` `3u.io` `4o.io` `5b.io` `5d.io` `5j.io` `5n.io` `6b.io` `6o.io` `6v.io` `6x.io` `7o.io` `8i.io` `8j.io` `8u.io` `96.io` `9n.io` `b0.io` `em.io` `f7.io` `g8.io` `j8.io` `kx.io` `lp.io` `m4.io` `nw.io` `p7.io` `r1.io` `vl.io` `x2.io` `y7.io` `ya.io` `yh.io`

<details><summary><b>.dev</b> – 705 names</summary>

`a.dev` `b.dev` `c.dev` `d.dev` `e.dev` `f.dev` `h.dev` `i.dev` `j.dev` `k.dev` `l.dev` `m.dev` `n.dev` `o.dev` `p.dev` `q.dev` `r.dev` `s.dev` `t.dev` `u.dev` `v.dev` `w.dev` `x.dev` `y.dev` `z.dev` `00.dev` `02.dev` `03.dev` `04.dev` `05.dev` `06.dev` `07.dev` `08.dev` `09.dev` `0a.dev` `0b.dev` `0c.dev` `0d.dev` `0e.dev` `0f.dev` `0g.dev` `0h.dev` `0i.dev` `0j.dev` `0k.dev` `0l.dev` `0n.dev` `0o.dev` `0q.dev` `0r.dev` `0s.dev` `0t.dev` `0u.dev` `0v.dev` `0w.dev` `0y.dev` `0z.dev` `11.dev` `12.dev` `13.dev` `14.dev` `16.dev` `17.dev` `18.dev` `19.dev` `1a.dev` `1b.dev` `1d.dev` `1e.dev` `1f.dev` `1g.dev` `1h.dev` `1i.dev` `1j.dev` `1k.dev` `1l.dev` `1m.dev` `1n.dev` `1o.dev` `1q.dev` `1r.dev` `1s.dev` `1t.dev` `1u.dev` `1v.dev` `1w.dev` `1y.dev` `1z.dev` `20.dev` `23.dev` `26.dev` `27.dev` `28.dev` `29.dev` `2a.dev` `2b.dev` `2e.dev` `2g.dev` `2h.dev` `2i.dev` `2j.dev` `2l.dev` `2m.dev` `2n.dev` `2o.dev` `2p.dev` `2q.dev` `2r.dev` `2s.dev` `2t.dev` `2u.dev` `2v.dev` `2w.dev` `2x.dev` `2y.dev` `2z.dev` `30.dev` `31.dev` `32.dev` `33.dev` `34.dev` `35.dev` `37.dev` `38.dev` `39.dev` `3c.dev` `3e.dev` `3f.dev` `3g.dev` `3h.dev` `3i.dev` `3j.dev` `3k.dev` `3l.dev` `3n.dev` `3o.dev` `3p.dev` `3q.dev` `3r.dev` `3s.dev` `3t.dev` `3u.dev` `3v.dev` `3w.dev` `3x.dev` `3y.dev` `40.dev` `41.dev` `42.dev` `43.dev` `44.dev` `46.dev` `47.dev` `48.dev` `4a.dev` `4b.dev` `4c.dev` `4d.dev` `4e.dev` `4f.dev` `4g.dev` `4h.dev` `4i.dev` `4j.dev` `4l.dev` `4m.dev` `4n.dev` `4o.dev` `4p.dev` `4q.dev` `4r.dev` `4t.dev` `4u.dev` `4v.dev` `4w.dev` `4x.dev` `4y.dev` `4z.dev` `50.dev` `51.dev` `52.dev` `54.dev` `56.dev` `57.dev` `58.dev` `59.dev` `5a.dev` `5b.dev` `5c.dev` `5d.dev` `5e.dev` `5g.dev` `5h.dev` `5i.dev` `5j.dev` `5k.dev` `5l.dev` `5m.dev` `5n.dev` `5o.dev` `5p.dev` `5q.dev` `5r.dev` `5s.dev` `5t.dev` `5u.dev` `5v.dev` `5w.dev` `5x.dev` `5y.dev` `5z.dev` `60.dev` `61.dev` `62.dev` `63.dev` `64.dev` `65.dev` `66.dev` `68.dev` `69.dev` `6a.dev` `6b.dev` `6c.dev` `6d.dev` `6e.dev` `6f.dev` `6g.dev` `6h.dev` `6i.dev` `6j.dev` `6k.dev` `6l.dev` `6m.dev` `6n.dev` `6o.dev` `6p.dev` `6q.dev` `6r.dev` `6s.dev` `6t.dev` `6u.dev` `6v.dev` `6w.dev` `6x.dev` `6y.dev` `6z.dev` `70.dev` `71.dev` `72.dev` `73.dev` `75.dev` `76.dev` `77.dev` `78.dev` `79.dev` `7a.dev` `7b.dev` `7c.dev` `7d.dev` `7e.dev` `7f.dev` `7g.dev` `7h.dev` `7i.dev` `7j.dev` `7k.dev` `7l.dev` `7m.dev` `7n.dev` `7o.dev` `7p.dev` `7q.dev` `7r.dev` `7s.dev` `7t.dev` `7u.dev` `7v.dev` `7w.dev` `7x.dev` `7y.dev` `7z.dev` `81.dev` `82.dev` `83.dev` `84.dev` `85.dev` `86.dev` `87.dev` `88.dev` `89.dev` `8a.dev` `8b.dev` `8c.dev` `8d.dev` `8e.dev` `8f.dev` `8g.dev` `8h.dev` `8i.dev` `8j.dev` `8k.dev` `8l.dev` `8n.dev` `8o.dev` `8p.dev` `8q.dev` `8r.dev` `8s.dev` `8t.dev` `8u.dev` `8v.dev` `8w.dev` `8x.dev` `8y.dev` `8z.dev` `90.dev` `91.dev` `92.dev` `93.dev` `94.dev` `95.dev` `96.dev` `97.dev` `98.dev` `9a.dev` `9b.dev` `9c.dev` `9d.dev` `9e.dev` `9f.dev` `9g.dev` `9h.dev` `9i.dev` `9j.dev` `9k.dev` `9l.dev` `9m.dev` `9n.dev` `9o.dev` `9p.dev` `9q.dev` `9r.dev` `9s.dev` `9t.dev` `9u.dev` `9v.dev` `9w.dev` `9y.dev` `9z.dev` `a4.dev` `a6.dev` `a7.dev` `a8.dev` `au.dev` `ay.dev` `b3.dev` `b4.dev` `b5.dev` `b6.dev` `b7.dev` `bj.dev` `bn.dev` `bq.dev` `bu.dev` `bv.dev` `c4.dev` `c5.dev` `c7.dev` `c8.dev` `c9.dev` `cq.dev` `ct.dev` `d2.dev` `d4.dev` `d5.dev` `d6.dev` `d8.dev` `d9.dev` `dc.dev` `dl.dev` `dp.dev` `dq.dev` `e1.dev` `e3.dev` `e4.dev` `e5.dev` `e6.dev` `e8.dev` `e9.dev` `ea.dev` `ec.dev` `ej.dev` `ek.dev` `ep.dev` `er.dev` `eu.dev` `ew.dev` `ey.dev` `f1.dev` `f2.dev` `f5.dev` `f7.dev` `f8.dev` `f9.dev` `fq.dev` `fu.dev` `fy.dev` `fz.dev` `g0.dev` `g1.dev` `g3.dev` `g4.dev` `g5.dev` `g6.dev` `g7.dev` `g9.dev` `gd.dev` `gn.dev` `gp.dev` `gq.dev` `gu.dev` `gw.dev` `gy.dev` `gz.dev` `h4.dev` `h5.dev` `h6.dev` `h7.dev` `h8.dev` `h9.dev` `hc.dev` `he.dev` `hj.dev` `hl.dev` `hm.dev` `ho.dev` `hs.dev` `hv.dev` `hy.dev` `i1.dev` `i2.dev` `i3.dev` `i4.dev` `i5.dev` `i7.dev` `i8.dev` `i9.dev` `ic.dev` `ig.dev` `ih.dev` `ij.dev` `ik.dev` `iu.dev` `iy.dev` `iz.dev` `j0.dev` `j1.dev` `j3.dev` `j4.dev` `j5.dev` `j6.dev` `j7.dev` `j8.dev` `j9.dev` `jc.dev` `jn.dev` `jo.dev` `jq.dev` `ju.dev` `jv.dev` `jy.dev` `jz.dev` `k1.dev` `k3.dev` `k4.dev` `k5.dev` `k6.dev` `k8.dev` `kq.dev` `ku.dev` `kx.dev` `l0.dev` `l2.dev` `l6.dev` `l7.dev` `l9.dev` `lb.dev` `ld.dev` `lh.dev` `lo.dev` `lq.dev` `lr.dev` `ly.dev` `m4.dev` `m5.dev` `m6.dev` `m7.dev` `m9.dev` `mq.dev` `n2.dev` `n3.dev` `n5.dev` `n6.dev` `n9.dev` `nc.dev` `nf.dev` `nh.dev` `nq.dev` `nr.dev` `nw.dev` `ny.dev` `o1.dev` `o4.dev` `o5.dev` `o6.dev` `o7.dev` `o8.dev` `o9.dev` `od.dev` `oh.dev` `oj.dev` `ol.dev` `or.dev` `ot.dev` `ou.dev` `ow.dev` `oy.dev` `p3.dev` `p4.dev` `p5.dev` `p7.dev` `p8.dev` `p9.dev` `pf.dev` `pn.dev` `po.dev` `pq.dev` `q1.dev` `q2.dev` `q3.dev` `q4.dev` `q5.dev` `q6.dev` `q7.dev` `q8.dev` `q9.dev` `qc.dev` `qd.dev` `qe.dev` `qf.dev` `qg.dev` `qh.dev` `qj.dev` `qk.dev` `ql.dev` `qn.dev` `qo.dev` `qv.dev` `qw.dev` `qy.dev` `qz.dev` `r5.dev` `r6.dev` `r9.dev` `rf.dev` `rj.dev` `rk.dev` `rn.dev` `rq.dev` `rt.dev` `ry.dev` `rz.dev` `s3.dev` `s6.dev` `s7.dev` `s8.dev` `s9.dev` `sx.dev` `sz.dev` `t4.dev` `t5.dev` `t7.dev` `t9.dev` `te.dev` `tf.dev` `tl.dev` `tp.dev` `tq.dev` `u1.dev` `u2.dev` `u3.dev` `u4.dev` `u5.dev` `u6.dev` `u7.dev` `u8.dev` `u9.dev` `ub.dev` `uf.dev` `uh.dev` `uj.dev` `ul.dev` `un.dev` `uo.dev` `ut.dev` `uu.dev` `uw.dev` `v3.dev` `v5.dev` `v6.dev` `v7.dev` `v9.dev` `va.dev` `vd.dev` `vf.dev` `vg.dev` `vj.dev` `vl.dev` `vq.dev` `vu.dev` `vy.dev` `vz.dev` `w0.dev` `w1.dev` `w2.dev` `w4.dev` `w5.dev` `w6.dev` `w7.dev` `w8.dev` `w9.dev` `wb.dev` `wd.dev` `wg.dev` `wh.dev` `wi.dev` `wj.dev` `wk.dev` `wl.dev` `wn.dev` `wq.dev` `wr.dev` `wt.dev` `wu.dev` `wy.dev` `wz.dev` `x0.dev` `x4.dev` `x5.dev` `x6.dev` `x8.dev` `x9.dev` `xa.dev` `xc.dev` `xf.dev` `xg.dev` `xh.dev` `xj.dev` `xk.dev` `xn.dev` `xq.dev` `xu.dev` `xw.dev` `xz.dev` `y1.dev` `y3.dev` `y4.dev` `y5.dev` `y6.dev` `y7.dev` `y8.dev` `y9.dev` `yd.dev` `ye.dev` `yf.dev` `yg.dev` `yh.dev` `yj.dev` `yk.dev` `yl.dev` `ym.dev` `yn.dev` `yq.dev` `ys.dev` `yt.dev` `yv.dev` `yw.dev` `yx.dev` `yz.dev` `z2.dev` `z5.dev` `z6.dev` `z7.dev` `z8.dev` `zc.dev` `zd.dev` `zh.dev` `zj.dev` `zl.dev` `zn.dev` `zq.dev` `zr.dev` `zs.dev` `zu.dev` `zv.dev`

</details>

<details><summary><b>.app</b> – 351 names</summary>

`g.app` `y.app` `0a.app` `0b.app` `0c.app` `0d.app` `0e.app` `0f.app` `0h.app` `0i.app` `0j.app` `0k.app` `0l.app` `0n.app` `0p.app` `0q.app` `0r.app` `0s.app` `0t.app` `0u.app` `0v.app` `0w.app` `0y.app` `0z.app` `1d.app` `1e.app` `1g.app` `1j.app` `1l.app` `1n.app` `1o.app` `1p.app` `1q.app` `1r.app` `1t.app` `1v.app` `1y.app` `1z.app` `2c.app` `2d.app` `2e.app` `2f.app` `2h.app` `2i.app` `2l.app` `2m.app` `2n.app` `2o.app` `2t.app` `2y.app` `2z.app` `3c.app` `3f.app` `3h.app` `3i.app` `3j.app` `3n.app` `3o.app` `3w.app` `3z.app` `4a.app` `4b.app` `4e.app` `4f.app` `4h.app` `4j.app` `4k.app` `4l.app` `4m.app` `4n.app` `4o.app` `4p.app` `4q.app` `4r.app` `4t.app` `4u.app` `4v.app` `4w.app` `4x.app` `4y.app` `4z.app` `5i.app` `5j.app` `5k.app` `6a.app` `6c.app` `6e.app` `6f.app` `6h.app` `6i.app` `6j.app` `6l.app` `6m.app` `6n.app` `6o.app` `6q.app` `6t.app` `6v.app` `6w.app` `6x.app` `6y.app` `6z.app` `7a.app` `7b.app` `7d.app` `7f.app` `7h.app` `7i.app` `7k.app` `7l.app` `7n.app` `7o.app` `7q.app` `7r.app` `7u.app` `7y.app` `7z.app` `8a.app` `8c.app` `8d.app` `8e.app` `8h.app` `8i.app` `8l.app` `8n.app` `8o.app` `8q.app` `8s.app` `8v.app` `8z.app` `9c.app` `9i.app` `9j.app` `9l.app` `9m.app` `9n.app` `9o.app` `9p.app` `9q.app` `9t.app` `9v.app` `9w.app` `9x.app` `a3.app` `a4.app` `ap.app` `au.app` `b2.app` `bj.app` `bq.app` `c5.app` `c8.app` `cj.app` `cz.app` `d2.app` `dk.app` `dz.app` `e0.app` `e5.app` `e6.app` `e7.app` `ec.app` `ef.app` `eg.app` `eh.app` `ep.app` `eu.app` `ew.app` `ey.app` `f0.app` `f2.app` `f6.app` `f7.app` `f9.app` `fe.app` `fk.app` `fs.app` `fu.app` `fv.app` `fw.app` `fy.app` `fz.app` `g3.app` `g4.app` `gb.app` `gi.app` `gk.app` `gq.app` `gu.app` `h0.app` `h7.app` `ha.app` `hc.app` `hf.app` `hm.app` `hn.app` `ho.app` `hx.app` `i1.app` `i2.app` `i4.app` `i5.app` `i6.app` `ib.app` `ic.app` `ie.app` `ih.app` `ij.app` `ik.app` `io.app` `is.app` `iu.app` `iv.app` `iy.app` `j0.app` `j4.app` `j6.app` `j9.app` `ja.app` `je.app` `jm.app` `jn.app` `jr.app` `jv.app` `k0.app` `k4.app` `kh.app` `kj.app` `ku.app` `kv.app` `kz.app` `l3.app` `l6.app` `l7.app` `l8.app` `l9.app` `lf.app` `lh.app` `lp.app` `lq.app` `lv.app` `m6.app` `m7.app` `m9.app` `mb.app` `ml.app` `n4.app` `n5.app` `nj.app` `nk.app` `nm.app` `no.app` `np.app` `nq.app` `o4.app` `o5.app` `o6.app` `o9.app` `ot.app` `ow.app` `oy.app` `p4.app` `p7.app` `pl.app` `pu.app` `pz.app` `q4.app` `q5.app` `q6.app` `qe.app` `qf.app` `qg.app` `qj.app` `qk.app` `qn.app` `qu.app` `qw.app` `qz.app` `r0.app` `r3.app` `r4.app` `r5.app` `r8.app` `rg.app` `rl.app` `rm.app` `rz.app` `s7.app` `sl.app` `sr.app` `t5.app` `t7.app` `te.app` `tf.app` `tl.app` `u2.app` `u3.app` `u4.app` `u6.app` `ue.app` `ug.app` `uj.app` `ul.app` `un.app` `uo.app` `uq.app` `uy.app` `uz.app` `v3.app` `v5.app` `vd.app` `ve.app` `vg.app` `w0.app` `w4.app` `w8.app` `wn.app` `wo.app` `wq.app` `wt.app` `wu.app` `wv.app` `x4.app` `xv.app` `xz.app` `y0.app` `y4.app` `y7.app` `ya.app` `yk.app` `ym.app` `yp.app` `yr.app` `yt.app` `yv.app` `z2.app` `z4.app` `z5.app` `z9.app` `za.app` `zc.app` `zd.app` `ze.app` `zf.app` `zi.app` `zj.app` `zr.app` `zu.app` `zv.app` `zw.app` `zy.app`

</details>

<details><summary><b>.tools</b> – 924 names</summary>

`2.tools` `3.tools` `4.tools` `5.tools` `6.tools` `8.tools` `d.tools` `k.tools` `l.tools` `n.tools` `t.tools` `00.tools` `02.tools` `03.tools` `04.tools` `05.tools` `06.tools` `07.tools` `08.tools` `09.tools` `0a.tools` `0b.tools` `0c.tools` `0d.tools` `0e.tools` `0f.tools` `0g.tools` `0h.tools` `0i.tools` `0j.tools` `0k.tools` `0l.tools` `0m.tools` `0n.tools` `0o.tools` `0p.tools` `0q.tools` `0r.tools` `0s.tools` `0t.tools` `0u.tools` `0v.tools` `0w.tools` `0y.tools` `0z.tools` `12.tools` `14.tools` `15.tools` `16.tools` `17.tools` `18.tools` `1b.tools` `1c.tools` `1d.tools` `1f.tools` `1g.tools` `1h.tools` `1i.tools` `1j.tools` `1k.tools` `1l.tools` `1m.tools` `1n.tools` `1o.tools` `1p.tools` `1q.tools` `1r.tools` `1s.tools` `1t.tools` `1u.tools` `1v.tools` `1w.tools` `1x.tools` `1y.tools` `1z.tools` `20.tools` `22.tools` `24.tools` `26.tools` `27.tools` `29.tools` `2a.tools` `2b.tools` `2c.tools` `2d.tools` `2e.tools` `2f.tools` `2g.tools` `2h.tools` `2i.tools` `2m.tools` `2n.tools` `2o.tools` `2p.tools` `2q.tools` `2r.tools` `2s.tools` `2t.tools` `2u.tools` `2v.tools` `2w.tools` `2y.tools` `2z.tools` `30.tools` `31.tools` `32.tools` `33.tools` `34.tools` `35.tools` `36.tools` `37.tools` `38.tools` `39.tools` `3a.tools` `3b.tools` `3c.tools` `3d.tools` `3e.tools` `3f.tools` `3g.tools` `3h.tools` `3i.tools` `3j.tools` `3k.tools` `3l.tools` `3m.tools` `3n.tools` `3o.tools` `3p.tools` `3q.tools` `3r.tools` `3t.tools` `3v.tools` `3w.tools` `3x.tools` `3y.tools` `3z.tools` `40.tools` `41.tools` `43.tools` `44.tools` `45.tools` `46.tools` `47.tools` `48.tools` `49.tools` `4a.tools` `4b.tools` `4c.tools` `4d.tools` `4e.tools` `4f.tools` `4g.tools` `4h.tools` `4i.tools` `4j.tools` `4k.tools` `4l.tools` `4m.tools` `4n.tools` `4o.tools` `4p.tools` `4q.tools` `4r.tools` `4s.tools` `4t.tools` `4v.tools` `4w.tools` `4x.tools` `4y.tools` `4z.tools` `50.tools` `51.tools` `52.tools` `53.tools` `54.tools` `55.tools` `56.tools` `57.tools` `58.tools` `59.tools` `5a.tools` `5b.tools` `5c.tools` `5d.tools` `5f.tools` `5g.tools` `5h.tools` `5i.tools` `5j.tools` `5k.tools` `5l.tools` `5m.tools` `5n.tools` `5o.tools` `5p.tools` `5q.tools` `5r.tools` `5s.tools` `5t.tools` `5u.tools` `5v.tools` `5w.tools` `5x.tools` `5y.tools` `5z.tools` `61.tools` `62.tools` `63.tools` `64.tools` `65.tools` `66.tools` `67.tools` `68.tools` `69.tools` `6a.tools` `6b.tools` `6c.tools` `6d.tools` `6e.tools` `6f.tools` `6g.tools` `6h.tools` `6i.tools` `6j.tools` `6k.tools` `6l.tools` `6m.tools` `6n.tools` `6o.tools` `6p.tools` `6q.tools` `6r.tools` `6s.tools` `6t.tools` `6u.tools` `6v.tools` `6w.tools` `6x.tools` `6y.tools` `6z.tools` `70.tools` `71.tools` `72.tools` `73.tools` `74.tools` `75.tools` `76.tools` `78.tools` `7a.tools` `7b.tools` `7c.tools` `7d.tools` `7e.tools` `7f.tools` `7g.tools` `7h.tools` `7i.tools` `7j.tools` `7k.tools` `7l.tools` `7m.tools` `7n.tools` `7o.tools` `7p.tools` `7q.tools` `7r.tools` `7s.tools` `7t.tools` `7u.tools` `7v.tools` `7w.tools` `7x.tools` `7y.tools` `7z.tools` `80.tools` `81.tools` `82.tools` `83.tools` `84.tools` `85.tools` `86.tools` `87.tools` `89.tools` `8a.tools` `8b.tools` `8c.tools` `8d.tools` `8e.tools` `8f.tools` `8g.tools` `8h.tools` `8i.tools` `8j.tools` `8k.tools` `8l.tools` `8m.tools` `8n.tools` `8o.tools` `8p.tools` `8q.tools` `8r.tools` `8s.tools` `8t.tools` `8u.tools` `8v.tools` `8w.tools` `8x.tools` `8y.tools` `8z.tools` `90.tools` `92.tools` `93.tools` `94.tools` `95.tools` `96.tools` `97.tools` `9a.tools` `9b.tools` `9c.tools` `9d.tools` `9e.tools` `9f.tools` `9g.tools` `9h.tools` `9i.tools` `9j.tools` `9k.tools` `9l.tools` `9m.tools` `9n.tools` `9o.tools` `9p.tools` `9q.tools` `9r.tools` `9s.tools` `9t.tools` `9u.tools` `9v.tools` `9w.tools` `9x.tools` `9y.tools` `9z.tools` `a1.tools` `a2.tools` `a3.tools` `a4.tools` `a5.tools` `a7.tools` `a9.tools` `aa.tools` `ah.tools` `aj.tools` `al.tools` `ap.tools` `aq.tools` `as.tools` `au.tools` `ay.tools` `b0.tools` `b1.tools` `b2.tools` `b3.tools` `b4.tools` `b5.tools` `b6.tools` `b7.tools` `b8.tools` `bd.tools` `bf.tools` `bj.tools` `bl.tools` `bo.tools` `bq.tools` `bs.tools` `bt.tools` `bv.tools` `by.tools` `bz.tools` `c0.tools` `c1.tools` `c2.tools` `c3.tools` `c4.tools` `c5.tools` `c6.tools` `c7.tools` `c8.tools` `c9.tools` `cc.tools` `cd.tools` `cg.tools` `cj.tools` `ck.tools` `cl.tools` `cm.tools` `cn.tools` `cq.tools` `cu.tools` `cw.tools` `cz.tools` `d0.tools` `d2.tools` `d4.tools` `d6.tools` `d7.tools` `d8.tools` `d9.tools` `du.tools` `dv.tools` `dy.tools` `dz.tools` `e0.tools` `e1.tools` `e2.tools` `e3.tools` `e4.tools` `e5.tools` `e6.tools` `e7.tools` `e8.tools` `e9.tools` `eb.tools` `ec.tools` `ef.tools` `eg.tools` `eh.tools` `ei.tools` `en.tools` `eo.tools` `ep.tools` `et.tools` `eu.tools` `ey.tools` `f0.tools` `f1.tools` `f2.tools` `f3.tools` `f4.tools` `f5.tools` `f6.tools` `f7.tools` `f8.tools` `f9.tools` `fa.tools` `fg.tools` `fh.tools` `fj.tools` `fk.tools` `fo.tools` `fq.tools` `fr.tools` `fu.tools` `fv.tools` `fz.tools` `g0.tools` `g1.tools` `g2.tools` `g3.tools` `g5.tools` `g6.tools` `g7.tools` `g8.tools` `g9.tools` `ga.tools` `gd.tools` `ge.tools` `gf.tools` `gh.tools` `gi.tools` `gj.tools` `gn.tools` `go.tools` `gp.tools` `gq.tools` `gs.tools` `gv.tools` `gw.tools` `gy.tools` `gz.tools` `h0.tools` `h1.tools` `h2.tools` `h3.tools` `h4.tools` `h5.tools` `h6.tools` `h7.tools` `h8.tools` `h9.tools` `he.tools` `hg.tools` `hh.tools` `hk.tools` `hl.tools` `hm.tools` `hn.tools` `ho.tools` `hp.tools` `ht.tools` `hw.tools` `hx.tools` `i0.tools` `i1.tools` `i2.tools` `i3.tools` `i4.tools` `i5.tools` `i6.tools` `i7.tools` `i8.tools` `i9.tools` `ie.tools` `if.tools` `ih.tools` `ii.tools` `ij.tools` `ik.tools` `il.tools` `in.tools` `it.tools` `iu.tools` `iv.tools` `iw.tools` `ix.tools` `iy.tools` `iz.tools` `j0.tools` `j1.tools` `j2.tools` `j3.tools` `j4.tools` `j5.tools` `j6.tools` `j7.tools` `j8.tools` `j9.tools` `jb.tools` `jc.tools` `je.tools` `jf.tools` `jg.tools` `jh.tools` `ji.tools` `jl.tools` `jm.tools` `jn.tools` `jo.tools` `jq.tools` `jt.tools` `ju.tools` `jv.tools` `jw.tools` `jx.tools` `jy.tools` `jz.tools` `k0.tools` `k1.tools` `k3.tools` `k4.tools` `k5.tools` `k6.tools` `k7.tools` `k9.tools` `ka.tools` `kb.tools` `kd.tools` `ke.tools` `kf.tools` `kg.tools` `kh.tools` `kj.tools` `kk.tools` `kl.tools` `kn.tools` `kp.tools` `kq.tools` `kr.tools` `kt.tools` `ku.tools` `kv.tools` `kw.tools` `kx.tools` `kz.tools` `l0.tools` `l3.tools` `l4.tools` `l6.tools` `l7.tools` `l8.tools` `l9.tools` `la.tools` `le.tools` `lf.tools` `lg.tools` `lk.tools` `ll.tools` `lq.tools` `lr.tools` `lt.tools` `lu.tools` `lv.tools` `lw.tools` `ly.tools` `lz.tools` `m0.tools` `m4.tools` `m5.tools` `m6.tools` `m7.tools` `m8.tools` `m9.tools` `mk.tools` `mn.tools` `mq.tools` `mt.tools` `n0.tools` `n2.tools` `n3.tools` `n4.tools` `n5.tools` `n6.tools` `n7.tools` `n8.tools` `n9.tools` `na.tools` `nb.tools` `nc.tools` `nh.tools` `ni.tools` `nj.tools` `nm.tools` `nn.tools` `np.tools` `nr.tools` `nt.tools` `nu.tools` `ny.tools` `nz.tools` `o0.tools` `o1.tools` `o2.tools` `o3.tools` `o4.tools` `o5.tools` `o6.tools` `o7.tools` `o8.tools` `o9.tools` `oa.tools` `od.tools` `oe.tools` `oh.tools` `oj.tools` `ok.tools` `ol.tools` `om.tools` `oq.tools` `ou.tools` `ov.tools` `ow.tools` `oy.tools` `p0.tools` `p1.tools` `p2.tools` `p4.tools` `p5.tools` `p6.tools` `p7.tools` `p8.tools` `p9.tools` `pb.tools` `pe.tools` `ph.tools` `pl.tools` `pn.tools` `po.tools` `pu.tools` `pv.tools` `q0.tools` `q1.tools` `q2.tools` `q3.tools` `q4.tools` `q5.tools` `q6.tools` `q7.tools` `q8.tools` `q9.tools` `qc.tools` `qd.tools` `qe.tools` `qf.tools` `qg.tools` `qh.tools` `qj.tools` `qk.tools` `ql.tools` `qm.tools` `qn.tools` `qo.tools` `qp.tools` `qs.tools` `qt.tools` `qu.tools` `qv.tools` `qx.tools` `qy.tools` `qz.tools` `r1.tools` `r2.tools` `r3.tools` `r4.tools` `r5.tools` `r6.tools` `r7.tools` `r8.tools` `r9.tools` `ra.tools` `rg.tools` `rh.tools` `ri.tools` `rk.tools` `rq.tools` `rt.tools` `rv.tools` `ry.tools` `rz.tools` `s0.tools` `s1.tools` `s2.tools` `s4.tools` `s5.tools` `s6.tools` `s7.tools` `s9.tools` `sa.tools` `sb.tools` `se.tools` `sg.tools` `sj.tools` `sl.tools` `sm.tools` `so.tools` `sq.tools` `ss.tools` `su.tools` `sv.tools` `sy.tools` `sz.tools` `t0.tools` `t2.tools` `t6.tools` `t7.tools` `t8.tools` `t9.tools` `te.tools` `tf.tools` `ti.tools` `tj.tools` `tl.tools` `tm.tools` `tq.tools` `tr.tools` `tt.tools` `tv.tools` `tx.tools` `ty.tools` `tz.tools` `u1.tools` `u2.tools` `u3.tools` `u4.tools` `u5.tools` `u6.tools` `u7.tools` `u8.tools` `u9.tools` `ub.tools` `uf.tools` `ug.tools` `uh.tools` `uj.tools` `ul.tools` `un.tools` `uo.tools` `uq.tools` `ut.tools` `uu.tools` `uv.tools` `uy.tools` `uz.tools` `v0.tools` `v1.tools` `v3.tools` `v5.tools` `v7.tools` `v8.tools` `v9.tools` `va.tools` `vb.tools` `vd.tools` `vf.tools` `vh.tools` `vl.tools` `vn.tools` `vo.tools` `vq.tools` `vs.tools` `vt.tools` `vw.tools` `vz.tools` `w0.tools` `w1.tools` `w2.tools` `w4.tools` `w5.tools` `w6.tools` `w7.tools` `w8.tools` `w9.tools` `wc.tools` `wd.tools` `wg.tools` `wi.tools` `wj.tools` `wk.tools` `wo.tools` `wq.tools` `wr.tools` `wt.tools` `wu.tools` `wv.tools` `wy.tools` `wz.tools` `x0.tools` `x2.tools` `x4.tools` `x5.tools` `x6.tools` `x7.tools` `x8.tools` `x9.tools` `xa.tools` `xb.tools` `xd.tools` `xe.tools` `xf.tools` `xg.tools` `xh.tools` `xi.tools` `xj.tools` `xk.tools` `xm.tools` `xn.tools` `xo.tools` `xq.tools` `xs.tools` `xt.tools` `xu.tools` `xv.tools` `xw.tools` `xy.tools` `y0.tools` `y1.tools` `y2.tools` `y3.tools` `y4.tools` `y5.tools` `y6.tools` `y7.tools` `y8.tools` `y9.tools` `yb.tools` `yc.tools` `yd.tools` `ye.tools` `yf.tools` `yg.tools` `yh.tools` `yi.tools` `yj.tools` `yk.tools` `yl.tools` `ym.tools` `yn.tools` `yp.tools` `yq.tools` `yr.tools` `ys.tools` `yu.tools` `yv.tools` `yw.tools` `yx.tools` `yy.tools` `yz.tools` `z0.tools` `z1.tools` `z2.tools` `z3.tools` `z4.tools` `z5.tools` `z6.tools` `z7.tools` `z8.tools` `z9.tools` `za.tools` `zb.tools` `zc.tools` `zd.tools` `zf.tools` `zh.tools` `zl.tools` `zm.tools` `zo.tools` `zp.tools` `zr.tools` `zt.tools` `zu.tools` `zv.tools` `zw.tools` `zx.tools` `zy.tools` `zz.tools`

</details>

**.co** – `13.co` `1m.co` `51.co` `ae.co` `ao.co` `at.co` `az.co` `bi.co` `ci.co` `ck.co` `co.co` `cr.co` `da.co` `eu.co` `g1.co` `gg.co` `gl.co` `gy.co` `hu.co` `id.co` `il.co` `im.co` `in.co` `je.co` `jp.co` `k7.co` `ke.co` `l3.co` `ln.co` `lu.co` `mw.co` `n2.co` `ni.co` `nl.co` `nz.co` `p3.co` `rs.co` `sz.co` `th.co` `tq.co` `tt.co` `ug.co` `uk.co` `v2.co` `v8.co` `z3.co` `za.co`

<details><summary><b>.org</b> – 78 names</summary>

`0b.org` `0c.org` `0u.org` `2i.org` `2t.org` `3f.org` `4p.org` `5d.org` `5p.org` `5s.org` `5y.org` `6c.org` `6e.org` `6i.org` `6k.org` `6p.org` `6q.org` `6u.org` `6v.org` `6y.org` `6z.org` `7g.org` `7h.org` `7i.org` `7j.org` `7k.org` `7m.org` `7p.org` `7q.org` `7r.org` `7x.org` `7y.org` `8a.org` `8b.org` `8c.org` `8e.org` `8f.org` `8h.org` `8j.org` `8r.org` `8s.org` `8u.org` `8v.org` `8x.org` `9g.org` `9j.org` `9k.org` `9l.org` `9m.org` `9n.org` `9p.org` `9s.org` `a2.org` `ar.org` `b5.org` `b7.org` `bs.org` `bu.org` `c8.org` `do.org` `e7.org` `i0.org` `iq.org` `j8.org` `j9.org` `l4.org` `l8.org` `mw.org` `pt.org` `q2.org` `q6.org` `q7.org` `r5.org` `sk.org` `tv.org` `vc.org` `xp.org` `zn.org`

</details>

<details><summary><b>.xyz</b> – 586 names</summary>

`4.xyz` `5.xyz` `7.xyz` `j.xyz` `l.xyz` `u.xyz` `x.xyz` `y.xyz` `z.xyz` `00.xyz` `02.xyz` `03.xyz` `04.xyz` `05.xyz` `06.xyz` `07.xyz` `08.xyz` `09.xyz` `13.xyz` `14.xyz` `17.xyz` `18.xyz` `19.xyz` `1p.xyz` `1y.xyz` `20.xyz` `23.xyz` `24.xyz` `25.xyz` `27.xyz` `28.xyz` `2c.xyz` `30.xyz` `31.xyz` `32.xyz` `33.xyz` `34.xyz` `37.xyz` `38.xyz` `39.xyz` `3c.xyz` `3d.xyz` `3i.xyz` `3m.xyz` `3v.xyz` `40.xyz` `41.xyz` `43.xyz` `44.xyz` `45.xyz` `46.xyz` `47.xyz` `48.xyz` `4d.xyz` `4s.xyz` `50.xyz` `53.xyz` `54.xyz` `56.xyz` `57.xyz` `58.xyz` `59.xyz` `5v.xyz` `60.xyz` `61.xyz` `62.xyz` `63.xyz` `64.xyz` `65.xyz` `68.xyz` `69.xyz` `6c.xyz` `6f.xyz` `6m.xyz` `6w.xyz` `70.xyz` `71.xyz` `72.xyz` `73.xyz` `74.xyz` `75.xyz` `76.xyz` `77.xyz` `78.xyz` `79.xyz` `81.xyz` `82.xyz` `83.xyz` `84.xyz` `85.xyz` `86.xyz` `87.xyz` `89.xyz` `90.xyz` `91.xyz` `92.xyz` `93.xyz` `94.xyz` `95.xyz` `97.xyz` `9x.xyz` `ab.xyz` `ac.xyz` `aj.xyz` `al.xyz` `aq.xyz` `as.xyz` `au.xyz` `av.xyz` `aw.xyz` `ay.xyz` `b1.xyz` `ba.xyz` `bd.xyz` `bf.xyz` `bh.xyz` `bi.xyz` `bk.xyz` `bn.xyz` `bp.xyz` `bq.xyz` `br.xyz` `bt.xyz` `bv.xyz` `bw.xyz` `bx.xyz` `by.xyz` `bz.xyz` `c7.xyz` `cc.xyz` `ce.xyz` `cg.xyz` `ci.xyz` `cj.xyz` `cl.xyz` `cn.xyz` `cp.xyz` `cq.xyz` `cr.xyz` `ct.xyz` `cu.xyz` `cv.xyz` `cw.xyz` `cx.xyz` `cy.xyz` `d8.xyz` `da.xyz` `df.xyz` `dg.xyz` `dh.xyz` `dj.xyz` `dk.xyz` `dl.xyz` `dm.xyz` `dn.xyz` `dq.xyz` `dr.xyz` `dt.xyz` `du.xyz` `dw.xyz` `dx.xyz` `dy.xyz` `dz.xyz` `ea.xyz` `eb.xyz` `ec.xyz` `ed.xyz` `ee.xyz` `ef.xyz` `eg.xyz` `eh.xyz` `ej.xyz` `ek.xyz` `en.xyz` `ep.xyz` `er.xyz` `et.xyz` `eu.xyz` `ew.xyz` `ey.xyz` `fa.xyz` `fb.xyz` `fe.xyz` `fg.xyz` `fh.xyz` `fj.xyz` `fk.xyz` `fl.xyz` `fm.xyz` `fq.xyz` `fr.xyz` `ft.xyz` `fu.xyz` `fv.xyz` `fz.xyz` `g2.xyz` `g8.xyz` `gc.xyz` `gd.xyz` `ge.xyz` `gf.xyz` `gh.xyz` `gj.xyz` `gk.xyz` `gl.xyz` `gm.xyz` `gn.xyz` `gp.xyz` `gq.xyz` `gt.xyz` `gu.xyz` `gv.xyz` `gw.xyz` `gy.xyz` `gz.xyz` `h4.xyz` `h5.xyz` `ha.xyz` `hb.xyz` `hc.xyz` `hd.xyz` `he.xyz` `hf.xyz` `hg.xyz` `hh.xyz` `hj.xyz` `hm.xyz` `ho.xyz` `hp.xyz` `hr.xyz` `hs.xyz` `hw.xyz` `hy.xyz` `i7.xyz` `ib.xyz` `ie.xyz` `if.xyz` `ig.xyz` `ih.xyz` `ii.xyz` `ij.xyz` `ik.xyz` `il.xyz` `in.xyz` `io.xyz` `iq.xyz` `ir.xyz` `is.xyz` `it.xyz` `iu.xyz` `iv.xyz` `iw.xyz` `ix.xyz` `iy.xyz` `iz.xyz` `ja.xyz` `jb.xyz` `jc.xyz` `jd.xyz` `je.xyz` `jf.xyz` `jg.xyz` `jh.xyz` `ji.xyz` `jk.xyz` `jl.xyz` `jm.xyz` `jn.xyz` `jp.xyz` `jr.xyz` `ju.xyz` `jv.xyz` `jx.xyz` `jy.xyz` `jz.xyz` `ka.xyz` `kb.xyz` `kd.xyz` `ke.xyz` `kf.xyz` `kg.xyz` `kh.xyz` `ki.xyz` `kj.xyz` `kl.xyz` `kn.xyz` `kp.xyz` `kq.xyz` `kr.xyz` `ks.xyz` `kt.xyz` `ku.xyz` `kv.xyz` `kw.xyz` `ky.xyz` `kz.xyz` `la.xyz` `lb.xyz` `lc.xyz` `ld.xyz` `le.xyz` `lf.xyz` `lg.xyz` `lh.xyz` `li.xyz` `lj.xyz` `lk.xyz` `ll.xyz` `lm.xyz` `lo.xyz` `lq.xyz` `lr.xyz` `ls.xyz` `lt.xyz` `lu.xyz` `lv.xyz` `lw.xyz` `lx.xyz` `ly.xyz` `lz.xyz` `ma.xyz` `mb.xyz` `md.xyz` `mh.xyz` `mi.xyz` `mj.xyz` `mk.xyz` `mn.xyz` `mp.xyz` `mq.xyz` `mr.xyz` `ms.xyz` `mt.xyz` `mu.xyz` `mv.xyz` `mx.xyz` `na.xyz` `nb.xyz` `nc.xyz` `nd.xyz` `nf.xyz` `nh.xyz` `ni.xyz` `nj.xyz` `nk.xyz` `nl.xyz` `nm.xyz` `np.xyz` `nq.xyz` `nr.xyz` `nv.xyz` `nw.xyz` `nx.xyz` `ny.xyz` `nz.xyz` `ob.xyz` `oc.xyz` `od.xyz` `og.xyz` `oj.xyz` `ol.xyz` `oo.xyz` `op.xyz` `oq.xyz` `ot.xyz` `ou.xyz` `ow.xyz` `ox.xyz` `oy.xyz` `pd.xyz` `pe.xyz` `pf.xyz` `pg.xyz` `ph.xyz` `pj.xyz` `pk.xyz` `pm.xyz` `pn.xyz` `pq.xyz` `pr.xyz` `ps.xyz` `pt.xyz` `pu.xyz` `pv.xyz` `pw.xyz` `q1.xyz` `q8.xyz` `qa.xyz` `qb.xyz` `qc.xyz` `qd.xyz` `qe.xyz` `qf.xyz` `qg.xyz` `qh.xyz` `qj.xyz` `qk.xyz` `ql.xyz` `qm.xyz` `qn.xyz` `qo.xyz` `qp.xyz` `qq.xyz` `qs.xyz` `qt.xyz` `qv.xyz` `qw.xyz` `qx.xyz` `qy.xyz` `qz.xyz` `r8.xyz` `rb.xyz` `rc.xyz` `rd.xyz` `rf.xyz` `rg.xyz` `rh.xyz` `rj.xyz` `rk.xyz` `rl.xyz` `rm.xyz` `rn.xyz` `ro.xyz` `rp.xyz` `rq.xyz` `rr.xyz` `rs.xyz` `rt.xyz` `ru.xyz` `rw.xyz` `rx.xyz` `ry.xyz` `rz.xyz` `sa.xyz` `sc.xyz` `sd.xyz` `se.xyz` `sf.xyz` `sg.xyz` `sh.xyz` `sj.xyz` `sk.xyz` `sn.xyz` `so.xyz` `sr.xyz` `st.xyz` `su.xyz` `sw.xyz` `sy.xyz` `sz.xyz` `t9.xyz` `tb.xyz` `td.xyz` `tg.xyz` `th.xyz` `ti.xyz` `tj.xyz` `tm.xyz` `tn.xyz` `to.xyz` `tr.xyz` `ts.xyz` `tt.xyz` `tu.xyz` `tv.xyz` `tw.xyz` `ty.xyz` `tz.xyz` `ua.xyz` `ub.xyz` `ue.xyz` `uf.xyz` `uh.xyz` `uj.xyz` `uk.xyz` `un.xyz` `uo.xyz` `uq.xyz` `us.xyz` `ut.xyz` `uu.xyz` `uw.xyz` `uy.xyz` `va.xyz` `vb.xyz` `vd.xyz` `ve.xyz` `vf.xyz` `vg.xyz` `vh.xyz` `vi.xyz` `vj.xyz` `vk.xyz` `vl.xyz` `vn.xyz` `vo.xyz` `vp.xyz` `vq.xyz` `vs.xyz` `vw.xyz` `vx.xyz` `vy.xyz` `vz.xyz` `wa.xyz` `wb.xyz` `wc.xyz` `wd.xyz` `wf.xyz` `wg.xyz` `wh.xyz` `wi.xyz` `wj.xyz` `wk.xyz` `wl.xyz` `wm.xyz` `wn.xyz` `wo.xyz` `wp.xyz` `wq.xyz` `wr.xyz` `ws.xyz` `wt.xyz` `wu.xyz` `wv.xyz` `wx.xyz` `wy.xyz` `wz.xyz` `xa.xyz` `xc.xyz` `xd.xyz` `xe.xyz` `xf.xyz` `xg.xyz` `xh.xyz` `xi.xyz` `xj.xyz` `xk.xyz` `xl.xyz` `xm.xyz` `xn.xyz` `xo.xyz` `xq.xyz` `xr.xyz` `xt.xyz` `xu.xyz` `xv.xyz` `xw.xyz` `xx.xyz` `xz.xyz` `ya.xyz` `yd.xyz` `ye.xyz` `yf.xyz` `yg.xyz` `yh.xyz` `yi.xyz` `yj.xyz` `yk.xyz` `yl.xyz` `yp.xyz` `yq.xyz` `yr.xyz` `yt.xyz` `yu.xyz` `yv.xyz` `yw.xyz` `yx.xyz` `zb.xyz` `zc.xyz` `zd.xyz` `zf.xyz` `zg.xyz` `zh.xyz` `zi.xyz` `zj.xyz` `zl.xyz` `zm.xyz` `zn.xyz` `zp.xyz` `zq.xyz` `zr.xyz` `zs.xyz` `zu.xyz` `zv.xyz` `zw.xyz` `zy.xyz`

</details>

**.cc** – `af.cc` `bs.cc` `cu.cc` `dn.cc` `dr.cc` `fl.cc` `fp.cc` `ld.cc` `mx.cc` `ny.cc` `oq.cc` `qu.cc` `sa.cc` `sc.cc` `th.cc` `xu.cc`

<details><summary><b>.tv</b> – 98 names</summary>

`g.tv` `i.tv` `08.tv` `0a.tv` `0b.tv` `0c.tv` `0d.tv` `0e.tv` `0f.tv` `0g.tv` `0h.tv` `0i.tv` `0j.tv` `0k.tv` `0l.tv` `0m.tv` `0n.tv` `0o.tv` `0p.tv` `0q.tv` `0r.tv` `0s.tv` `0t.tv` `0u.tv` `0v.tv` `0w.tv` `0x.tv` `0y.tv` `0z.tv` `11.tv` `12.tv` `13.tv` `15.tv` `16.tv` `19.tv` `1a.tv` `1b.tv` `1c.tv` `1d.tv` `1e.tv` `1f.tv` `1g.tv` `1h.tv` `1i.tv` `1j.tv` `1k.tv` `1l.tv` `1m.tv` `1n.tv` `1o.tv` `1p.tv` `1q.tv` `1r.tv` `1t.tv` `1u.tv` `1v.tv` `1w.tv` `1x.tv` `1y.tv` `1z.tv` `21.tv` `23.tv` `26.tv` `27.tv` `29.tv` `2e.tv` `2f.tv` `2g.tv` `2h.tv` `2i.tv` `2j.tv` `2k.tv` `2l.tv` `2m.tv` `2n.tv` `2o.tv` `2q.tv` `2r.tv` `2v.tv` `2w.tv` `2y.tv` `2z.tv` `31.tv` `32.tv` `33.tv` `34.tv` `35.tv` `36.tv` `37.tv` `38.tv` `39.tv` `3a.tv` `3b.tv` `3f.tv` `3g.tv` `3h.tv` `3i.tv` `3j.tv`

</details>

<details><summary><b>.fm</b> – 1066 names</summary>

`0a.fm` `0b.fm` `0c.fm` `0d.fm` `0e.fm` `0f.fm` `0g.fm` `0h.fm` `0i.fm` `0j.fm` `0k.fm` `0l.fm` `0m.fm` `0n.fm` `0p.fm` `0q.fm` `0r.fm` `0s.fm` `0t.fm` `0u.fm` `0v.fm` `0w.fm` `0y.fm` `0z.fm` `11.fm` `12.fm` `13.fm` `14.fm` `15.fm` `16.fm` `17.fm` `18.fm` `19.fm` `1a.fm` `1b.fm` `1c.fm` `1d.fm` `1e.fm` `1f.fm` `1g.fm` `1h.fm` `1i.fm` `1j.fm` `1k.fm` `1l.fm` `1m.fm` `1n.fm` `1o.fm` `1p.fm` `1q.fm` `1r.fm` `1s.fm` `1t.fm` `1u.fm` `1v.fm` `1w.fm` `1x.fm` `1y.fm` `1z.fm` `20.fm` `22.fm` `23.fm` `24.fm` `25.fm` `26.fm` `27.fm` `28.fm` `29.fm` `2a.fm` `2b.fm` `2c.fm` `2d.fm` `2e.fm` `2f.fm` `2g.fm` `2h.fm` `2i.fm` `2j.fm` `2k.fm` `2l.fm` `2m.fm` `2o.fm` `2q.fm` `2r.fm` `2s.fm` `2t.fm` `2u.fm` `2v.fm` `2w.fm` `2x.fm` `2y.fm` `2z.fm` `30.fm` `31.fm` `32.fm` `33.fm` `34.fm` `35.fm` `36.fm` `37.fm` `38.fm` `39.fm` `3a.fm` `3b.fm` `3c.fm` `3e.fm` `3f.fm` `3g.fm` `3h.fm` `3i.fm` `3j.fm` `3k.fm` `3l.fm` `3n.fm` `3o.fm` `3p.fm` `3q.fm` `3r.fm` `3s.fm` `3t.fm` `3u.fm` `3v.fm` `3w.fm` `3x.fm` `3y.fm` `3z.fm` `40.fm` `41.fm` `42.fm` `43.fm` `44.fm` `45.fm` `46.fm` `47.fm` `48.fm` `49.fm` `4a.fm` `4b.fm` `4c.fm` `4d.fm` `4e.fm` `4f.fm` `4g.fm` `4h.fm` `4i.fm` `4j.fm` `4k.fm` `4l.fm` `4m.fm` `4n.fm` `4o.fm` `4p.fm` `4q.fm` `4r.fm` `4s.fm` `4t.fm` `4u.fm` `4v.fm` `4w.fm` `4x.fm` `4y.fm` `4z.fm` `50.fm` `51.fm` `52.fm` `53.fm` `54.fm` `55.fm` `56.fm` `57.fm` `58.fm` `59.fm` `5a.fm` `5b.fm` `5c.fm` `5d.fm` `5e.fm` `5f.fm` `5g.fm` `5h.fm` `5i.fm` `5j.fm` `5k.fm` `5l.fm` `5m.fm` `5n.fm` `5o.fm` `5p.fm` `5q.fm` `5r.fm` `5s.fm` `5t.fm` `5u.fm` `5v.fm` `5w.fm` `5x.fm` `5y.fm` `5z.fm` `60.fm` `61.fm` `62.fm` `63.fm` `64.fm` `65.fm` `66.fm` `67.fm` `68.fm` `69.fm` `6a.fm` `6b.fm` `6c.fm` `6d.fm` `6e.fm` `6f.fm` `6g.fm` `6h.fm` `6i.fm` `6j.fm` `6k.fm` `6l.fm` `6m.fm` `6n.fm` `6o.fm` `6p.fm` `6q.fm` `6r.fm` `6s.fm` `6t.fm` `6u.fm` `6v.fm` `6w.fm` `6x.fm` `6y.fm` `6z.fm` `70.fm` `71.fm` `72.fm` `73.fm` `74.fm` `75.fm` `76.fm` `77.fm` `79.fm` `7a.fm` `7b.fm` `7c.fm` `7d.fm` `7e.fm` `7f.fm` `7g.fm` `7h.fm` `7i.fm` `7j.fm` `7k.fm` `7l.fm` `7m.fm` `7n.fm` `7o.fm` `7p.fm` `7q.fm` `7r.fm` `7s.fm` `7t.fm` `7u.fm` `7v.fm` `7w.fm` `7x.fm` `7y.fm` `7z.fm` `81.fm` `82.fm` `83.fm` `84.fm` `85.fm` `86.fm` `87.fm` `88.fm` `89.fm` `8a.fm` `8b.fm` `8c.fm` `8d.fm` `8e.fm` `8f.fm` `8g.fm` `8h.fm` `8i.fm` `8j.fm` `8k.fm` `8l.fm` `8m.fm` `8n.fm` `8o.fm` `8p.fm` `8q.fm` `8r.fm` `8s.fm` `8t.fm` `8u.fm` `8v.fm` `8w.fm` `8x.fm` `8y.fm` `8z.fm` `90.fm` `91.fm` `92.fm` `93.fm` `94.fm` `95.fm` `96.fm` `97.fm` `98.fm` `9a.fm` `9b.fm` `9c.fm` `9d.fm` `9e.fm` `9f.fm` `9g.fm` `9h.fm` `9i.fm` `9j.fm` `9k.fm` `9l.fm` `9m.fm` `9n.fm` `9o.fm` `9p.fm` `9q.fm` `9r.fm` `9s.fm` `9t.fm` `9u.fm` `9v.fm` `9w.fm` `9x.fm` `9y.fm` `9z.fm` `a0.fm` `a1.fm` `a3.fm` `a4.fm` `a5.fm` `a6.fm` `a7.fm` `a8.fm` `aa.fm` `ab.fm` `ac.fm` `ae.fm` `af.fm` `ag.fm` `aj.fm` `ak.fm` `al.fm` `an.fm` `ao.fm` `aq.fm` `ar.fm` `as.fm` `au.fm` `av.fm` `aw.fm` `ax.fm` `ay.fm` `az.fm` `b0.fm` `b1.fm` `b2.fm` `b3.fm` `b4.fm` `b5.fm` `b6.fm` `b7.fm` `b9.fm` `ba.fm` `bc.fm` `bd.fm` `be.fm` `bf.fm` `bg.fm` `bh.fm` `bi.fm` `bj.fm` `bk.fm` `bl.fm` `bm.fm` `bn.fm` `bo.fm` `bq.fm` `br.fm` `bt.fm` `bu.fm` `bv.fm` `bw.fm` `bx.fm` `c0.fm` `c1.fm` `c3.fm` `c4.fm` `c5.fm` `c6.fm` `c7.fm` `c8.fm` `c9.fm` `ca.fm` `cb.fm` `cc.fm` `ce.fm` `cf.fm` `cg.fm` `ci.fm` `cj.fm` `ck.fm` `cl.fm` `cm.fm` `cn.fm` `cp.fm` `cq.fm` `cr.fm` `cs.fm` `cv.fm` `cw.fm` `cx.fm` `cy.fm` `cz.fm` `d0.fm` `d2.fm` `d3.fm` `d4.fm` `d5.fm` `d6.fm` `d7.fm` `d8.fm` `d9.fm` `da.fm` `db.fm` `dc.fm` `dd.fm` `de.fm` `df.fm` `dg.fm` `dh.fm` `dk.fm` `dl.fm` `dm.fm` `dn.fm` `do.fm` `dp.fm` `dq.fm` `dr.fm` `ds.fm` `dt.fm` `du.fm` `dv.fm` `dw.fm` `dx.fm` `dy.fm` `dz.fm` `e0.fm` `e1.fm` `e2.fm` `e4.fm` `e5.fm` `e6.fm` `e7.fm` `e8.fm` `e9.fm` `ea.fm` `eb.fm` `ec.fm` `ee.fm` `ef.fm` `eg.fm` `eh.fm` `ei.fm` `ej.fm` `el.fm` `en.fm` `eo.fm` `ep.fm` `er.fm` `es.fm` `eu.fm` `ev.fm` `ew.fm` `ex.fm` `ey.fm` `f0.fm` `f1.fm` `f2.fm` `f3.fm` `f4.fm` `f5.fm` `f6.fm` `f7.fm` `f8.fm` `f9.fm` `fa.fm` `fc.fm` `fd.fm` `fe.fm` `fg.fm` `fh.fm` `fi.fm` `fj.fm` `fk.fm` `fl.fm` `fn.fm` `fo.fm` `fq.fm` `fr.fm` `fs.fm` `ft.fm` `fu.fm` `fv.fm` `fw.fm` `fy.fm` `fz.fm` `g0.fm` `g1.fm` `g2.fm` `g3.fm` `g4.fm` `g5.fm` `g6.fm` `g7.fm` `g8.fm` `g9.fm` `ga.fm` `gb.fm` `ge.fm` `gf.fm` `gi.fm` `gj.fm` `gk.fm` `gl.fm` `gn.fm` `gp.fm` `gq.fm` `gr.fm` `gu.fm` `gv.fm` `gw.fm` `gx.fm` `gy.fm` `gz.fm` `h0.fm` `h1.fm` `h2.fm` `h3.fm` `h4.fm` `h5.fm` `h6.fm` `h7.fm` `h8.fm` `h9.fm` `ha.fm` `he.fm` `hf.fm` `hh.fm` `hi.fm` `hj.fm` `hm.fm` `hn.fm` `ho.fm` `hs.fm` `hu.fm` `hv.fm` `hw.fm` `hx.fm` `hy.fm` `i0.fm` `i1.fm` `i2.fm` `i3.fm` `i4.fm` `i5.fm` `i6.fm` `i7.fm` `i8.fm` `i9.fm` `ia.fm` `ib.fm` `ic.fm` `ig.fm` `ih.fm` `ii.fm` `ij.fm` `ik.fm` `il.fm` `ir.fm` `iu.fm` `iv.fm` `iw.fm` `ix.fm` `iy.fm` `j0.fm` `j1.fm` `j2.fm` `j3.fm` `j4.fm` `j5.fm` `j6.fm` `j7.fm` `j8.fm` `j9.fm` `ja.fm` `je.fm` `jf.fm` `jg.fm` `jh.fm` `ji.fm` `jj.fm` `jk.fm` `jl.fm` `jm.fm` `jn.fm` `jo.fm` `jp.fm` `jq.fm` `jr.fm` `js.fm` `ju.fm` `jv.fm` `jy.fm` `jz.fm` `k0.fm` `k1.fm` `k2.fm` `k3.fm` `k4.fm` `k5.fm` `k6.fm` `k7.fm` `k8.fm` `k9.fm` `ka.fm` `kd.fm` `kf.fm` `kg.fm` `kh.fm` `ki.fm` `kj.fm` `kl.fm` `kn.fm` `kp.fm` `kq.fm` `kr.fm` `ks.fm` `ku.fm` `kv.fm` `ky.fm` `kz.fm` `l0.fm` `l1.fm` `l2.fm` `l3.fm` `l4.fm` `l5.fm` `l6.fm` `l7.fm` `l9.fm` `la.fm` `lc.fm` `ld.fm` `le.fm` `lh.fm` `lj.fm` `lk.fm` `ll.fm` `ln.fm` `lq.fm` `lr.fm` `ls.fm` `lv.fm` `lw.fm` `lx.fm` `ly.fm` `lz.fm` `m0.fm` `m2.fm` `m3.fm` `m4.fm` `m5.fm` `m6.fm` `m7.fm` `m8.fm` `m9.fm` `mb.fm` `md.fm` `me.fm` `mi.fm` `mn.fm` `mo.fm` `mq.fm` `ms.fm` `mt.fm` `mv.fm` `my.fm` `mz.fm` `n0.fm` `n1.fm` `n2.fm` `n3.fm` `n4.fm` `n5.fm` `n6.fm` `n7.fm` `n8.fm` `n9.fm` `nc.fm` `nd.fm` `ne.fm` `nf.fm` `ng.fm` `nh.fm` `ni.fm` `nj.fm` `nk.fm` `nm.fm` `nn.fm` `nq.fm` `nu.fm` `nw.fm` `nx.fm` `nz.fm` `o0.fm` `o1.fm` `o2.fm` `o3.fm` `o4.fm` `o5.fm` `o6.fm` `o7.fm` `o8.fm` `o9.fm` `oa.fm` `ob.fm` `oc.fm` `oe.fm` `of.fm` `og.fm` `oj.fm` `ok.fm` `ol.fm` `op.fm` `oq.fm` `ot.fm` `ou.fm` `ov.fm` `oz.fm` `p0.fm` `p1.fm` `p4.fm` `p5.fm` `p6.fm` `p7.fm` `p8.fm` `p9.fm` `pa.fm` `pc.fm` `pf.fm` `ph.fm` `pl.fm` `pn.fm` `po.fm` `ps.fm` `pt.fm` `pu.fm` `pv.fm` `pw.fm` `px.fm` `py.fm` `pz.fm` `q0.fm` `q1.fm` `q2.fm` `q3.fm` `q4.fm` `q5.fm` `q6.fm` `q7.fm` `q8.fm` `q9.fm` `qa.fm` `qb.fm` `qd.fm` `qe.fm` `qf.fm` `qg.fm` `qh.fm` `qi.fm` `qj.fm` `qk.fm` `ql.fm` `qn.fm` `qo.fm` `qp.fm` `qu.fm` `qv.fm` `qw.fm` `qx.fm` `qy.fm` `qz.fm` `r0.fm` `r1.fm` `r3.fm` `r4.fm` `r5.fm` `r6.fm` `r7.fm` `r8.fm` `r9.fm` `rb.fm` `rd.fm` `rf.fm` `rg.fm` `rh.fm` `ri.fm` `rj.fm` `rk.fm` `rl.fm` `rn.fm` `rp.fm` `rq.fm` `rv.fm` `rw.fm` `ry.fm` `s0.fm` `s1.fm` `s2.fm` `s4.fm` `s5.fm` `s6.fm` `s8.fm` `s9.fm` `sa.fm` `sb.fm` `sc.fm` `se.fm` `sg.fm` `si.fm` `sj.fm` `sk.fm` `sl.fm` `sm.fm` `sn.fm` `so.fm` `sp.fm` `sq.fm` `sr.fm` `su.fm` `sv.fm` `sw.fm` `sy.fm` `sz.fm` `t0.fm` `t1.fm` `t2.fm` `t3.fm` `t4.fm` `t5.fm` `t6.fm` `t7.fm` `t8.fm` `t9.fm` `tc.fm` `te.fm` `tf.fm` `tg.fm` `th.fm` `ti.fm` `tj.fm` `tk.fm` `tl.fm` `tm.fm` `tn.fm` `to.fm` `tp.fm` `tq.fm` `tr.fm` `tu.fm` `ty.fm` `tz.fm` `u0.fm` `u1.fm` `u2.fm` `u3.fm` `u4.fm` `u5.fm` `u6.fm` `u7.fm` `u8.fm` `u9.fm` `ub.fm` `ud.fm` `ue.fm` `uf.fm` `ug.fm` `uh.fm` `ui.fm` `ul.fm` `um.fm` `un.fm` `uo.fm` `uq.fm` `ut.fm` `uu.fm` `uv.fm` `uw.fm` `ux.fm` `uy.fm` `uz.fm` `v0.fm` `v1.fm` `v2.fm` `v3.fm` `v4.fm` `v5.fm` `v6.fm` `v7.fm` `v8.fm` `v9.fm` `va.fm` `vb.fm` `ve.fm` `vf.fm` `vg.fm` `vh.fm` `vi.fm` `vl.fm` `vm.fm` `vn.fm` `vp.fm` `vq.fm` `vr.fm` `vs.fm` `vt.fm` `vu.fm` `vx.fm` `vy.fm` `vz.fm` `w0.fm` `w1.fm` `w2.fm` `w3.fm` `w4.fm` `w5.fm` `w6.fm` `w7.fm` `w8.fm` `w9.fm` `wa.fm` `wc.fm` `wd.fm` `wf.fm` `wg.fm` `wh.fm` `wi.fm` `wj.fm` `wl.fm` `wm.fm` `wn.fm` `wo.fm` `wq.fm` `wr.fm` `ws.fm` `wt.fm` `wu.fm` `wv.fm` `wx.fm` `wy.fm` `wz.fm` `x0.fm` `x1.fm` `x2.fm` `x3.fm` `x4.fm` `x5.fm` `x6.fm` `x7.fm` `x8.fm` `x9.fm` `xa.fm` `xb.fm` `xc.fm` `xd.fm` `xe.fm` `xf.fm` `xg.fm` `xh.fm` `xi.fm` `xj.fm` `xk.fm` `xl.fm` `xm.fm` `xn.fm` `xp.fm` `xq.fm` `xs.fm` `xt.fm` `xu.fm` `xv.fm` `xw.fm` `xz.fm` `y0.fm` `y1.fm` `y2.fm` `y4.fm` `y5.fm` `y6.fm` `y7.fm` `y9.fm` `yc.fm` `yd.fm` `ye.fm` `yf.fm` `yg.fm` `yh.fm` `yi.fm` `yj.fm` `yk.fm` `yl.fm` `ym.fm` `yn.fm` `yp.fm` `yq.fm` `yr.fm` `ys.fm` `yt.fm` `yu.fm` `yv.fm` `yw.fm` `yx.fm` `yz.fm` `z0.fm` `z4.fm` `z5.fm` `z6.fm` `z7.fm` `z8.fm` `z9.fm` `za.fm` `zb.fm` `zc.fm` `zd.fm` `ze.fm` `zg.fm` `zh.fm` `zi.fm` `zj.fm` `zl.fm` `zm.fm` `zn.fm` `zo.fm` `zp.fm` `zq.fm` `zr.fm` `zs.fm` `zt.fm` `zw.fm` `zx.fm` `zy.fm` `zz.fm`

</details>

<details><summary><b>.is</b> – 62 names</summary>

`0g.is` `1n.is` `2j.is` `2m.is` `2o.is` `2v.is` `3a.is` `3e.is` `3i.is` `3j.is` `3k.is` `4p.is` `5a.is` `5b.is` `5e.is` `5f.is` `5h.is` `5j.is` `5k.is` `5n.is` `5o.is` `5r.is` `5y.is` `5z.is` `6a.is` `6c.is` `6e.is` `6f.is` `6h.is` `6i.is` `6j.is` `6k.is` `6l.is` `6m.is` `6o.is` `6q.is` `6r.is` `6s.is` `6t.is` `6u.is` `6w.is` `6y.is` `6z.is` `7a.is` `7b.is` `7d.is` `7e.is` `9y.is` `e6.is` `f4.is` `i6.is` `l3.is` `l6.is` `s9.is` `t7.is` `u3.is` `u6.is` `w6.is` `y3.is` `y5.is` `y6.is` `z5.is`

</details>

<details><summary><b>.run</b> – 688 names</summary>

`0a.run` `0b.run` `0c.run` `0d.run` `0e.run` `0f.run` `0g.run` `0h.run` `0i.run` `0j.run` `0k.run` `0l.run` `0m.run` `0n.run` `0o.run` `0p.run` `0q.run` `0r.run` `0s.run` `0t.run` `0u.run` `0v.run` `0w.run` `0y.run` `0z.run` `10.run` `12.run` `13.run` `15.run` `16.run` `19.run` `1a.run` `1b.run` `1c.run` `1d.run` `1e.run` `1f.run` `1g.run` `1h.run` `1i.run` `1j.run` `1k.run` `1l.run` `1m.run` `1n.run` `1o.run` `1p.run` `1q.run` `1r.run` `1u.run` `1v.run` `1y.run` `1z.run` `23.run` `25.run` `26.run` `27.run` `29.run` `2a.run` `2c.run` `2d.run` `2e.run` `2f.run` `2g.run` `2h.run` `2i.run` `2j.run` `2k.run` `2l.run` `2m.run` `2n.run` `2o.run` `2p.run` `2q.run` `2r.run` `2s.run` `2t.run` `2u.run` `2v.run` `2w.run` `2x.run` `2y.run` `2z.run` `30.run` `31.run` `32.run` `34.run` `35.run` `36.run` `37.run` `38.run` `39.run` `3a.run` `3c.run` `3f.run` `3g.run` `3h.run` `3i.run` `3j.run` `3k.run` `3m.run` `3n.run` `3o.run` `3p.run` `3q.run` `3r.run` `3s.run` `3t.run` `3u.run` `3v.run` `3w.run` `3x.run` `3y.run` `3z.run` `40.run` `41.run` `43.run` `44.run` `45.run` `46.run` `48.run` `49.run` `4b.run` `4c.run` `4d.run` `4e.run` `4f.run` `4g.run` `4h.run` `4i.run` `4j.run` `4l.run` `4m.run` `4n.run` `4o.run` `4p.run` `4q.run` `4r.run` `4s.run` `4t.run` `4u.run` `4v.run` `4w.run` `4x.run` `4y.run` `4z.run` `51.run` `52.run` `53.run` `54.run` `59.run` `5a.run` `5b.run` `5c.run` `5d.run` `5e.run` `5g.run` `5h.run` `5i.run` `5j.run` `5l.run` `5m.run` `5n.run` `5o.run` `5p.run` `5r.run` `5s.run` `5t.run` `5u.run` `5v.run` `5w.run` `5x.run` `5y.run` `5z.run` `60.run` `63.run` `64.run` `65.run` `68.run` `6a.run` `6b.run` `6c.run` `6d.run` `6e.run` `6f.run` `6g.run` `6h.run` `6i.run` `6j.run` `6k.run` `6l.run` `6m.run` `6n.run` `6o.run` `6p.run` `6q.run` `6r.run` `6s.run` `6t.run` `6u.run` `6v.run` `6w.run` `6x.run` `6y.run` `6z.run` `70.run` `72.run` `73.run` `74.run` `75.run` `76.run` `78.run` `79.run` `7a.run` `7b.run` `7c.run` `7d.run` `7e.run` `7f.run` `7g.run` `7h.run` `7i.run` `7j.run` `7k.run` `7l.run` `7n.run` `7o.run` `7p.run` `7q.run` `7r.run` `7s.run` `7t.run` `7u.run` `7v.run` `7w.run` `7x.run` `7y.run` `7z.run` `83.run` `84.run` `86.run` `87.run` `89.run` `8a.run` `8b.run` `8c.run` `8d.run` `8e.run` `8f.run` `8g.run` `8h.run` `8i.run` `8j.run` `8k.run` `8l.run` `8m.run` `8n.run` `8o.run` `8p.run` `8q.run` `8r.run` `8s.run` `8t.run` `8u.run` `8v.run` `8w.run` `8y.run` `8z.run` `90.run` `93.run` `94.run` `95.run` `9a.run` `9b.run` `9c.run` `9e.run` `9f.run` `9g.run` `9i.run` `9j.run` `9k.run` `9l.run` `9n.run` `9o.run` `9p.run` `9q.run` `9r.run` `9s.run` `9t.run` `9u.run` `9v.run` `9w.run` `9x.run` `9y.run` `9z.run` `a0.run` `a3.run` `a4.run` `a5.run` `a6.run` `a9.run` `ah.run` `an.run` `aq.run` `au.run` `av.run` `ay.run` `b0.run` `b2.run` `b3.run` `b5.run` `b6.run` `b7.run` `b8.run` `b9.run` `bh.run` `bj.run` `bm.run` `bp.run` `bv.run` `c0.run` `c2.run` `c5.run` `c7.run` `c8.run` `c9.run` `cn.run` `cq.run` `cu.run` `cw.run` `cz.run` `d0.run` `d2.run` `d4.run` `d5.run` `d6.run` `d7.run` `d8.run` `d9.run` `df.run` `dn.run` `dq.run` `dv.run` `dy.run` `e0.run` `e1.run` `e4.run` `e5.run` `e7.run` `e8.run` `e9.run` `eb.run` `ec.run` `eg.run` `ej.run` `ek.run` `eo.run` `ep.run` `eu.run` `ew.run` `ey.run` `f0.run` `f4.run` `f5.run` `f7.run` `f8.run` `f9.run` `fd.run` `fh.run` `fk.run` `ft.run` `fv.run` `fz.run` `g0.run` `g1.run` `g2.run` `g3.run` `g6.run` `g7.run` `g8.run` `g9.run` `gb.run` `gf.run` `gi.run` `gj.run` `gk.run` `gl.run` `gn.run` `go.run` `gq.run` `gz.run` `h0.run` `h1.run` `h2.run` `h4.run` `h7.run` `h8.run` `h9.run` `hc.run` `hn.run` `ht.run` `hv.run` `hw.run` `hz.run` `i0.run` `i2.run` `i3.run` `i4.run` `i5.run` `i6.run` `i7.run` `i8.run` `i9.run` `ih.run` `ij.run` `ik.run` `iv.run` `iy.run` `iz.run` `j0.run` `j1.run` `j2.run` `j3.run` `j4.run` `j5.run` `j6.run` `j7.run` `j8.run` `j9.run` `ja.run` `jb.run` `je.run` `jf.run` `jk.run` `jm.run` `jn.run` `jr.run` `jt.run` `jz.run` `k0.run` `k1.run` `k2.run` `k3.run` `k4.run` `k5.run` `k6.run` `k7.run` `k8.run` `kg.run` `kh.run` `ko.run` `kq.run` `kr.run` `kt.run` `kw.run` `ky.run` `kz.run` `l0.run` `l1.run` `l2.run` `l3.run` `l4.run` `l5.run` `l6.run` `l8.run` `l9.run` `ld.run` `lf.run` `lh.run` `lj.run` `lq.run` `lt.run` `lw.run` `m3.run` `m4.run` `m7.run` `mb.run` `mn.run` `ms.run` `mz.run` `n0.run` `n2.run` `n3.run` `n4.run` `n5.run` `n6.run` `na.run` `nf.run` `nq.run` `nz.run` `o1.run` `o2.run` `o3.run` `o4.run` `o5.run` `o6.run` `o7.run` `o9.run` `ob.run` `od.run` `og.run` `ol.run` `oq.run` `ou.run` `ow.run` `p2.run` `p3.run` `p4.run` `p5.run` `p6.run` `p7.run` `p8.run` `ph.run` `pj.run` `pq.run` `pu.run` `pv.run` `q0.run` `q2.run` `q3.run` `q4.run` `q5.run` `q6.run` `q7.run` `q8.run` `q9.run` `qb.run` `qc.run` `qe.run` `qf.run` `qg.run` `qh.run` `qj.run` `qm.run` `qn.run` `qo.run` `qp.run` `qs.run` `qv.run` `qw.run` `qy.run` `r0.run` `r3.run` `r5.run` `r6.run` `r7.run` `r9.run` `rg.run` `rh.run` `rk.run` `rn.run` `rp.run` `rs.run` `ru.run` `ry.run` `rz.run` `s1.run` `s3.run` `s4.run` `s5.run` `s6.run` `s7.run` `s9.run` `sx.run` `sz.run` `t1.run` `t2.run` `t3.run` `t4.run` `t5.run` `t6.run` `t7.run` `t9.run` `tb.run` `td.run` `tj.run` `tn.run` `tq.run` `u0.run` `u1.run` `u2.run` `u3.run` `u4.run` `u5.run` `u6.run` `u7.run` `u8.run` `u9.run` `ub.run` `ud.run` `ue.run` `uf.run` `ug.run` `uh.run` `uj.run` `un.run` `uo.run` `uq.run` `ur.run` `ut.run` `uw.run` `uy.run` `uz.run` `v5.run` `v7.run` `v9.run` `vb.run` `vd.run` `vf.run` `vg.run` `vh.run` `vj.run` `vl.run` `vp.run` `vq.run` `vt.run` `vw.run` `vz.run` `w0.run` `w1.run` `w2.run` `w4.run` `w6.run` `w7.run` `w8.run` `w9.run` `wa.run` `wc.run` `wg.run` `wk.run` `wr.run` `wt.run` `wv.run` `x0.run` `x2.run` `x5.run` `x6.run` `x7.run` `x9.run` `xf.run` `xh.run` `xi.run` `xk.run` `xq.run` `xu.run` `xv.run` `xw.run` `y0.run` `y1.run` `y2.run` `y3.run` `y4.run` `y5.run` `y6.run` `y7.run` `y8.run` `y9.run` `yd.run` `yf.run` `yg.run` `yj.run` `yk.run` `yn.run` `yp.run` `ys.run` `yv.run` `yw.run` `yz.run` `z0.run` `z1.run` `z4.run` `z5.run` `z6.run` `z7.run` `z9.run` `zb.run` `zc.run` `zd.run` `zf.run` `zj.run` `zl.run` `zn.run` `zp.run` `zq.run` `zr.run` `zv.run` `zw.run` `zx.run`

</details>

<details><summary><b>.inc</b> – 1155 names</summary>

`2.inc` `3.inc` `4.inc` `5.inc` `6.inc` `8.inc` `9.inc` `k.inc` `02.inc` `03.inc` `04.inc` `05.inc` `06.inc` `07.inc` `08.inc` `09.inc` `0a.inc` `0b.inc` `0c.inc` `0d.inc` `0e.inc` `0f.inc` `0g.inc` `0h.inc` `0i.inc` `0j.inc` `0k.inc` `0l.inc` `0m.inc` `0n.inc` `0o.inc` `0p.inc` `0q.inc` `0r.inc` `0s.inc` `0t.inc` `0u.inc` `0v.inc` `0w.inc` `0x.inc` `0y.inc` `0z.inc` `10.inc` `11.inc` `12.inc` `13.inc` `14.inc` `15.inc` `16.inc` `17.inc` `18.inc` `19.inc` `1a.inc` `1b.inc` `1c.inc` `1d.inc` `1e.inc` `1f.inc` `1g.inc` `1h.inc` `1i.inc` `1j.inc` `1k.inc` `1l.inc` `1m.inc` `1n.inc` `1o.inc` `1p.inc` `1q.inc` `1r.inc` `1s.inc` `1t.inc` `1v.inc` `1w.inc` `1y.inc` `1z.inc` `20.inc` `21.inc` `22.inc` `23.inc` `24.inc` `25.inc` `26.inc` `27.inc` `28.inc` `29.inc` `2a.inc` `2b.inc` `2c.inc` `2d.inc` `2e.inc` `2f.inc` `2g.inc` `2h.inc` `2i.inc` `2j.inc` `2k.inc` `2l.inc` `2m.inc` `2n.inc` `2o.inc` `2p.inc` `2q.inc` `2r.inc` `2s.inc` `2t.inc` `2u.inc` `2v.inc` `2w.inc` `2y.inc` `2z.inc` `30.inc` `31.inc` `32.inc` `33.inc` `34.inc` `35.inc` `36.inc` `37.inc` `38.inc` `39.inc` `3a.inc` `3c.inc` `3e.inc` `3f.inc` `3g.inc` `3h.inc` `3i.inc` `3j.inc` `3k.inc` `3l.inc` `3n.inc` `3o.inc` `3p.inc` `3q.inc` `3r.inc` `3s.inc` `3t.inc` `3u.inc` `3v.inc` `3w.inc` `3x.inc` `3y.inc` `3z.inc` `40.inc` `41.inc` `43.inc` `44.inc` `45.inc` `46.inc` `47.inc` `48.inc` `49.inc` `4a.inc` `4b.inc` `4c.inc` `4d.inc` `4e.inc` `4f.inc` `4g.inc` `4h.inc` `4i.inc` `4j.inc` `4k.inc` `4l.inc` `4m.inc` `4n.inc` `4o.inc` `4p.inc` `4q.inc` `4r.inc` `4s.inc` `4t.inc` `4u.inc` `4v.inc` `4w.inc` `4x.inc` `4y.inc` `4z.inc` `50.inc` `51.inc` `52.inc` `53.inc` `54.inc` `55.inc` `56.inc` `57.inc` `58.inc` `59.inc` `5a.inc` `5b.inc` `5c.inc` `5d.inc` `5e.inc` `5f.inc` `5g.inc` `5h.inc` `5i.inc` `5j.inc` `5k.inc` `5l.inc` `5m.inc` `5n.inc` `5o.inc` `5p.inc` `5q.inc` `5r.inc` `5s.inc` `5t.inc` `5u.inc` `5v.inc` `5w.inc` `5x.inc` `5y.inc` `5z.inc` `60.inc` `61.inc` `62.inc` `63.inc` `64.inc` `65.inc` `66.inc` `67.inc` `68.inc` `69.inc` `6a.inc` `6b.inc` `6c.inc` `6d.inc` `6e.inc` `6f.inc` `6g.inc` `6h.inc` `6i.inc` `6j.inc` `6k.inc` `6l.inc` `6m.inc` `6n.inc` `6o.inc` `6p.inc` `6q.inc` `6r.inc` `6s.inc` `6t.inc` `6u.inc` `6v.inc` `6w.inc` `6x.inc` `6y.inc` `6z.inc` `70.inc` `71.inc` `72.inc` `73.inc` `75.inc` `76.inc` `77.inc` `78.inc` `79.inc` `7a.inc` `7b.inc` `7c.inc` `7d.inc` `7e.inc` `7f.inc` `7g.inc` `7h.inc` `7i.inc` `7j.inc` `7k.inc` `7l.inc` `7m.inc` `7n.inc` `7o.inc` `7p.inc` `7q.inc` `7r.inc` `7s.inc` `7t.inc` `7u.inc` `7v.inc` `7w.inc` `7x.inc` `7y.inc` `7z.inc` `80.inc` `81.inc` `82.inc` `83.inc` `84.inc` `85.inc` `86.inc` `87.inc` `88.inc` `89.inc` `8a.inc` `8b.inc` `8c.inc` `8d.inc` `8e.inc` `8f.inc` `8g.inc` `8h.inc` `8i.inc` `8j.inc` `8k.inc` `8l.inc` `8m.inc` `8n.inc` `8o.inc` `8p.inc` `8q.inc` `8r.inc` `8s.inc` `8t.inc` `8u.inc` `8v.inc` `8w.inc` `8x.inc` `8y.inc` `8z.inc` `90.inc` `91.inc` `92.inc` `93.inc` `94.inc` `95.inc` `96.inc` `97.inc` `98.inc` `99.inc` `9a.inc` `9b.inc` `9c.inc` `9d.inc` `9e.inc` `9f.inc` `9g.inc` `9h.inc` `9i.inc` `9j.inc` `9k.inc` `9l.inc` `9m.inc` `9n.inc` `9o.inc` `9p.inc` `9q.inc` `9r.inc` `9s.inc` `9t.inc` `9u.inc` `9v.inc` `9w.inc` `9x.inc` `9y.inc` `9z.inc` `a0.inc` `a1.inc` `a2.inc` `a3.inc` `a4.inc` `a5.inc` `a7.inc` `a8.inc` `a9.inc` `aa.inc` `ac.inc` `ae.inc` `af.inc` `ah.inc` `aj.inc` `ak.inc` `am.inc` `an.inc` `ao.inc` `ap.inc` `aq.inc` `as.inc` `au.inc` `av.inc` `aw.inc` `ax.inc` `ay.inc` `az.inc` `b0.inc` `b1.inc` `b2.inc` `b3.inc` `b4.inc` `b5.inc` `b6.inc` `b7.inc` `b8.inc` `ba.inc` `bb.inc` `bc.inc` `bg.inc` `bi.inc` `bj.inc` `bk.inc` `bm.inc` `bn.inc` `bo.inc` `bp.inc` `bq.inc` `br.inc` `bs.inc` `bt.inc` `bu.inc` `bv.inc` `bw.inc` `bx.inc` `by.inc` `bz.inc` `c0.inc` `c1.inc` `c2.inc` `c4.inc` `c5.inc` `c6.inc` `c7.inc` `c8.inc` `c9.inc` `cb.inc` `cc.inc` `cd.inc` `ce.inc` `cf.inc` `cg.inc` `cj.inc` `cl.inc` `cm.inc` `cn.inc` `cp.inc` `cq.inc` `cr.inc` `cu.inc` `cx.inc` `cy.inc` `cz.inc` `d0.inc` `d1.inc` `d2.inc` `d4.inc` `d5.inc` `d6.inc` `d7.inc` `d8.inc` `d9.inc` `da.inc` `dd.inc` `de.inc` `df.inc` `dg.inc` `dh.inc` `di.inc` `dk.inc` `dl.inc` `dm.inc` `dn.inc` `do.inc` `dp.inc` `dq.inc` `ds.inc` `du.inc` `dv.inc` `dw.inc` `dx.inc` `dy.inc` `dz.inc` `e0.inc` `e1.inc` `e2.inc` `e3.inc` `e4.inc` `e5.inc` `e6.inc` `e7.inc` `e8.inc` `e9.inc` `ea.inc` `eb.inc` `ed.inc` `ee.inc` `ef.inc` `eh.inc` `ej.inc` `ek.inc` `el.inc` `em.inc` `eo.inc` `ep.inc` `eq.inc` `er.inc` `es.inc` `et.inc` `ev.inc` `ew.inc` `ex.inc` `ey.inc` `ez.inc` `f0.inc` `f2.inc` `f3.inc` `f5.inc` `f6.inc` `f7.inc` `f8.inc` `f9.inc` `fb.inc` `fc.inc` `fd.inc` `fe.inc` `ff.inc` `fg.inc` `fh.inc` `fj.inc` `fk.inc` `fl.inc` `fm.inc` `fn.inc` `fo.inc` `fq.inc` `ft.inc` `fu.inc` `fv.inc` `fw.inc` `fx.inc` `fy.inc` `fz.inc` `g0.inc` `g1.inc` `g2.inc` `g3.inc` `g4.inc` `g5.inc` `g6.inc` `g7.inc` `g8.inc` `g9.inc` `gb.inc` `gd.inc` `ge.inc` `gf.inc` `gg.inc` `gh.inc` `gj.inc` `gk.inc` `gl.inc` `gm.inc` `gn.inc` `gs.inc` `gt.inc` `gu.inc` `gv.inc` `gx.inc` `gy.inc` `gz.inc` `h0.inc` `h1.inc` `h3.inc` `h4.inc` `h5.inc` `h6.inc` `h7.inc` `h8.inc` `h9.inc` `ha.inc` `hb.inc` `hc.inc` `hd.inc` `he.inc` `hf.inc` `hg.inc` `hj.inc` `hk.inc` `hl.inc` `hm.inc` `hn.inc` `ho.inc` `hr.inc` `hs.inc` `ht.inc` `hu.inc` `hv.inc` `hx.inc` `hy.inc` `hz.inc` `i0.inc` `i1.inc` `i3.inc` `i4.inc` `i5.inc` `i6.inc` `i7.inc` `i8.inc` `i9.inc` `ia.inc` `ib.inc` `ic.inc` `id.inc` `ie.inc` `if.inc` `ig.inc` `ih.inc` `ij.inc` `ik.inc` `il.inc` `ip.inc` `is.inc` `iu.inc` `iv.inc` `iw.inc` `ix.inc` `iy.inc` `iz.inc` `j0.inc` `j1.inc` `j2.inc` `j3.inc` `j4.inc` `j5.inc` `j6.inc` `j7.inc` `j8.inc` `j9.inc` `ja.inc` `jb.inc` `jc.inc` `jd.inc` `je.inc` `jf.inc` `jg.inc` `jh.inc` `ji.inc` `jj.inc` `jk.inc` `jl.inc` `jm.inc` `jn.inc` `jo.inc` `jp.inc` `jq.inc` `jr.inc` `js.inc` `jt.inc` `ju.inc` `jv.inc` `jw.inc` `jx.inc` `jy.inc` `jz.inc` `k0.inc` `k1.inc` `k2.inc` `k3.inc` `k4.inc` `k5.inc` `k6.inc` `k7.inc` `k8.inc` `k9.inc` `ka.inc` `kb.inc` `kc.inc` `kd.inc` `ke.inc` `kf.inc` `kg.inc` `kh.inc` `ki.inc` `kj.inc` `kk.inc` `kl.inc` `km.inc` `kn.inc` `ko.inc` `kp.inc` `kq.inc` `kr.inc` `ks.inc` `kt.inc` `ku.inc` `kv.inc` `kw.inc` `kx.inc` `ky.inc` `kz.inc` `l0.inc` `l2.inc` `l3.inc` `l4.inc` `l5.inc` `l6.inc` `l7.inc` `l8.inc` `l9.inc` `la.inc` `lb.inc` `lc.inc` `ld.inc` `lf.inc` `lg.inc` `lh.inc` `li.inc` `lj.inc` `lk.inc` `ll.inc` `ln.inc` `lq.inc` `ls.inc` `lt.inc` `lu.inc` `lv.inc` `lw.inc` `lx.inc` `ly.inc` `m0.inc` `m1.inc` `m3.inc` `m4.inc` `m5.inc` `m6.inc` `m7.inc` `m9.inc` `ma.inc` `mb.inc` `mc.inc` `md.inc` `mf.inc` `mg.inc` `mh.inc` `mj.inc` `mk.inc` `mn.inc` `mq.inc` `ms.inc` `mt.inc` `mu.inc` `mv.inc` `mw.inc` `mx.inc` `mz.inc` `n0.inc` `n2.inc` `n3.inc` `n4.inc` `n5.inc` `n6.inc` `n7.inc` `n8.inc` `n9.inc` `na.inc` `nb.inc` `nc.inc` `nd.inc` `ne.inc` `nf.inc` `ng.inc` `nh.inc` `ni.inc` `nj.inc` `nk.inc` `nl.inc` `nm.inc` `nn.inc` `nq.inc` `nr.inc` `ns.inc` `nt.inc` `nu.inc` `nv.inc` `nw.inc` `nx.inc` `ny.inc` `nz.inc` `o0.inc` `o2.inc` `o3.inc` `o4.inc` `o5.inc` `o6.inc` `o7.inc` `o8.inc` `o9.inc` `oa.inc` `oc.inc` `od.inc` `oe.inc` `of.inc` `og.inc` `oj.inc` `ok.inc` `ol.inc` `om.inc` `on.inc` `oo.inc` `oq.inc` `or.inc` `os.inc` `ou.inc` `ov.inc` `ow.inc` `oy.inc` `p1.inc` `p2.inc` `p3.inc` `p4.inc` `p5.inc` `p6.inc` `p7.inc` `p8.inc` `p9.inc` `pb.inc` `pc.inc` `pd.inc` `pf.inc` `pg.inc` `ph.inc` `pj.inc` `pk.inc` `pl.inc` `pm.inc` `pn.inc` `po.inc` `pp.inc` `pq.inc` `pt.inc` `pu.inc` `pv.inc` `pw.inc` `px.inc` `py.inc` `pz.inc` `q0.inc` `q1.inc` `q2.inc` `q3.inc` `q5.inc` `q6.inc` `q7.inc` `q8.inc` `q9.inc` `qa.inc` `qb.inc` `qc.inc` `qd.inc` `qe.inc` `qf.inc` `qg.inc` `qh.inc` `qj.inc` `qk.inc` `ql.inc` `qm.inc` `qn.inc` `qo.inc` `qp.inc` `qr.inc` `qs.inc` `qt.inc` `qu.inc` `qv.inc` `qw.inc` `qx.inc` `qy.inc` `qz.inc` `r0.inc` `r1.inc` `r2.inc` `r3.inc` `r4.inc` `r5.inc` `r6.inc` `r7.inc` `r8.inc` `r9.inc` `ra.inc` `rb.inc` `rc.inc` `rd.inc` `re.inc` `rf.inc` `rg.inc` `rh.inc` `ri.inc` `rj.inc` `rk.inc` `rm.inc` `rn.inc` `ro.inc` `rp.inc` `rq.inc` `rr.inc` `rs.inc` `rt.inc` `ru.inc` `rv.inc` `rx.inc` `ry.inc` `rz.inc` `s0.inc` `s1.inc` `s3.inc` `s4.inc` `s5.inc` `s6.inc` `s7.inc` `s8.inc` `s9.inc` `sa.inc` `sc.inc` `se.inc` `sg.inc` `sj.inc` `sk.inc` `sl.inc` `sm.inc` `sn.inc` `so.inc` `sr.inc` `ss.inc` `su.inc` `sv.inc` `sw.inc` `sx.inc` `sy.inc` `sz.inc` `t0.inc` `t1.inc` `t3.inc` `t4.inc` `t6.inc` `t7.inc` `t8.inc` `t9.inc` `ta.inc` `tb.inc` `tc.inc` `td.inc` `te.inc` `tg.inc` `th.inc` `ti.inc` `tj.inc` `tl.inc` `tm.inc` `tn.inc` `to.inc` `tp.inc` `tq.inc` `tv.inc` `tw.inc` `tx.inc` `ty.inc` `tz.inc` `u0.inc` `u1.inc` `u2.inc` `u3.inc` `u4.inc` `u5.inc` `u6.inc` `u7.inc` `u8.inc` `u9.inc` `ua.inc` `ub.inc` `uc.inc` `ud.inc` `ue.inc` `uf.inc` `ug.inc` `uh.inc` `uj.inc` `uk.inc` `ul.inc` `um.inc` `un.inc` `uo.inc` `uq.inc` `ur.inc` `ut.inc` `uu.inc` `uv.inc` `uw.inc` `uy.inc` `uz.inc` `v0.inc` `v1.inc` `v2.inc` `v3.inc` `v4.inc` `v5.inc` `v6.inc` `v7.inc` `v8.inc` `v9.inc` `va.inc` `vc.inc` `vd.inc` `ve.inc` `vf.inc` `vg.inc` `vh.inc` `vj.inc` `vk.inc` `vl.inc` `vm.inc` `vn.inc` `vo.inc` `vq.inc` `vr.inc` `vs.inc` `vu.inc` `vv.inc` `vw.inc` `vx.inc` `vy.inc` `vz.inc` `w0.inc` `w1.inc` `w2.inc` `w4.inc` `w5.inc` `w6.inc` `w7.inc` `w8.inc` `w9.inc` `wa.inc` `wb.inc` `wf.inc` `wg.inc` `wh.inc` `wi.inc` `wj.inc` `wl.inc` `wn.inc` `wo.inc` `wq.inc` `wr.inc` `ws.inc` `wt.inc` `wu.inc` `wv.inc` `ww.inc` `wx.inc` `wy.inc` `wz.inc` `x0.inc` `x2.inc` `x3.inc` `x4.inc` `x5.inc` `x6.inc` `x7.inc` `x8.inc` `x9.inc` `xa.inc` `xb.inc` `xc.inc` `xd.inc` `xe.inc` `xf.inc` `xg.inc` `xh.inc` `xi.inc` `xj.inc` `xl.inc` `xm.inc` `xn.inc` `xo.inc` `xq.inc` `xr.inc` `xs.inc` `xt.inc` `xu.inc` `xv.inc` `xw.inc` `xx.inc` `xz.inc` `y0.inc` `y1.inc` `y2.inc` `y3.inc` `y4.inc` `y5.inc` `y6.inc` `y7.inc` `y8.inc` `y9.inc` `ya.inc` `yb.inc` `yc.inc` `yd.inc` `ye.inc` `yf.inc` `yg.inc` `yh.inc` `yi.inc` `yj.inc` `yk.inc` `yl.inc` `ym.inc` `yn.inc` `yo.inc` `yp.inc` `yq.inc` `yr.inc` `ys.inc` `yt.inc` `yu.inc` `yv.inc` `yw.inc` `yx.inc` `yy.inc` `yz.inc` `z0.inc` `z1.inc` `z2.inc` `z3.inc` `z4.inc` `z5.inc` `z6.inc` `z7.inc` `z8.inc` `z9.inc` `za.inc` `zb.inc` `zc.inc` `zd.inc` `ze.inc` `zf.inc` `zg.inc` `zh.inc` `zi.inc` `zj.inc` `zk.inc` `zl.inc` `zm.inc` `zn.inc` `zo.inc` `zp.inc` `zq.inc` `zr.inc` `zs.inc` `zt.inc` `zu.inc` `zv.inc` `zw.inc` `zx.inc` `zy.inc` `zz.inc`

</details>

<!-- shortfree:end -->

<a id="lists"></a>
## Only want the lists?

The companion repo [free-short-domains](https://github.com/maximilianfeix/free-short-domains) holds plain CSV and TXT files and is updated every day after the scan. Raw links:

| File | Content |
|---|---|
| [`csv/all.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/all.csv) | every RDAP-verified free 1–3 character domain on all scanned TLDs |
| `csv/<tld>.csv`, e.g. [`csv/dev.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/dev.csv) | 1–3 character names with status `available`, `dropping` or `unverified` |
| `txt/<tld>.txt`, e.g. [`txt/io.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/io.txt) | verified free 1–3 character domains, one per line |
| `four-letter/<tld>.txt`, e.g. [`four-letter/com.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/four-letter/com.txt) | 4-letter names (a–z) that are not in the zone |

All 4 character names, digits and hyphens included, are in the [`data` release](https://github.com/maximilianfeix/shortfree/releases/tag/data) as `four-<tld>.txt.gz`. The same data is available as a [JSON/CSV API](#api). Everything is CC0: no attribution, no key.

```sh
# every free 2 character .dev domain
curl -s https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/dev.txt | awk 'length($0) == 6'
```

<a id="limits"></a>
## What it does, and what it can't do

shortfree does:

- check every valid ASCII label (a–z, 0–9, inner hyphens) up to 4 characters: 49,284 names per TLD up to 3 characters and 1,774,224 more at 4 characters;
- ask the TLD's authoritative name servers directly, without a resolver cache;
- confirm each 1–3 character candidate in the registry's RDAP service (`404` = not registered);
- list registered names in redemption or pending delete separately, with their phase and RDAP expiration date (which is not a drop date);
- publish only RDAP answers younger than 14 days as verified, and say per TLD, length and shard what was covered ([coverage.json](docs/api.md#coverage), [accuracy](docs/accuracy.md));
- publish everything as plain files under CC0, with no account, key or tracking.

shortfree can't:

- **see prices or reservations.** "Not registered" (RDAP `404`) is not "registrable": registries reserve some names and sell others at premium prices, and RDAP reports both as unregistered. This is common for 1–2 character names on most TLDs other than .com, .net and .de. Your registrar shows the real price.
- **cover internationalized names.** Only ASCII labels are enumerated and checked; IDNs (Unicode or `xn--` labels) are rejected by `check`.
- **verify 4 character names via RDAP.** They are DNS-only. A registered name without name servers can appear in the 4 character lists. Run `shortfree check` on a name before you rely on it.
- **be newer than the last scan.** The lists are a daily snapshot. A name can be registered minutes after the scan. `shortfree check` and the MCP tool `check_domains` ask the registry live.

<a id="how-it-works"></a>
## How it works

1. **Sweep.** For every label, shortfree sends an `NS` query to the TLD's authoritative name servers. `NXDOMAIN` means the name is not in the zone. All 1–3 character names of one TLD are about 50,000 queries and take minutes.
2. **Verify.** Not being in the zone is not the same as being free: names on hold, without name servers or in redemption are missing too. Each 1–3 character candidate is looked up in the registry's RDAP service (Verisign, PIR, DENIC, Identity Digital, Google Registry, CentralNic and others, see [docs/tlds.md](docs/tlds.md)). Names shorter than a registry's minimum length are skipped. `404` = not registered. `pending delete` or `redemption period` = dropping.
3. **Repeat.** [`scan.yml`](.github/workflows/scan.yml) runs every day at 03:17 UTC, one job per TLD. It sweeps all 1–3 character names and one of seven shards of the 4 character space, so each 4 character name is rechecked once a week. RDAP results are rechecked at least every 10 days. Never-checked names go first, then dropping names, then the oldest checks.
4. **Publish.** Each TLD goes live as soon as its job is done, first its 1–3 character results, then its 4 character shard; nothing waits for the slowest TLD. Every job uploads only its own TLD, and [`pages.yml`](.github/workflows/pages.yml) merges whatever is done into the published data and deploys the site, one run at a time. When all jobs are done, the last run also writes the dated snapshot and commits `data/`.

Each RDAP host gets one total rate limit of a few requests per second, shared by all TLDs on that host and all parallel jobs (see [`internal/rdap/rdap.go`](internal/rdap/rdap.go)), to stay within fair use. Scanner state lives in the `data` release, so the git history only carries the small public lists.

<a id="mcp"></a>
## MCP server

`shortfree mcp` is an [MCP](https://modelcontextprotocol.io) server for Claude, Cursor, VS Code and other clients. It has four read-only tools:

| Tool | What it does |
|---|---|
| `check_domains` | live DNS + RDAP check of up to 50 domains, any length, any TLD (RDAP for every scanned TLD) |
| `find_free_domains` | search the daily snapshot by TLD, length, character kind, substring or prefix |
| `suggest_domains` | turn a word or project name into short candidates and return the free ones |
| `domain_stats` | counts per TLD and length |

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

Setup for Claude Desktop, Cursor and VS Code, and the scoring rules of `suggest_domains`: [docs/mcp.md](docs/mcp.md).

<a id="api"></a>
## API

A static, read-only JSON/CSV API on GitHub Pages, rebuilt after every scan. No key, CORS enabled.

```
https://maximilianfeix.github.io/shortfree/api/v1/
```

| Path | Content |
|---|---|
| `stats.json` | counts per TLD and length |
| `{tld}.json`, `{tld}/{n}.json` | 1–3 character names on one TLD, optionally of one length |
| `dropping.json` | names in redemption or pending delete, soonest first |
| `changes.json` | names freed or taken in the last 30 days |
| `free.csv` | every RDAP-verified free 1–3 character domain |
| `badge.json` | [shields.io endpoint](https://shields.io/badges/endpoint-badge) with the current count |

```sh
curl -s https://maximilianfeix.github.io/shortfree/api/v1/io/3.json | jq -r '.available[] + ".io"'
```

All endpoints, field meanings and examples in Python and JavaScript: [docs/api.md](docs/api.md). OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml).

<a id="cli"></a>
## CLI

```sh
shortfree check q7.tools x4k.app abc.de
```

```
DOMAIN    DNS    RDAP        VERDICT
q7.tools  free   available   not registered (RDAP 404)
x4k.app   taken  registered  taken
abc.de    taken  registered  taken
```

`check` works for ASCII domains of any length and TLD. For TLDs without RDAP support in shortfree it gives a DNS-only verdict (`not in DNS`). `-json` prints one object per domain, with `registration` (`not_registered`, `registered`, `dropping`, `not_delegated`, `unknown`) and `registrable` (`no` or `unconfirmed`: shortfree never asks a registrar) next to the verdict:

```sh
shortfree check -json q7.tools | jq '.[0].verdict'
```

The other commands run the scan itself: `scan` (1–3 characters, DNS + RDAP), `scan4` (one shard of 4 characters), `export` (writes `data/` and the table above), `merge` and `mcp`. `shortfree <command> -h` lists the flags.

### In GitHub Actions

The repo is also a GitHub Action. It checks a list of domains and can fail the job, or open an issue, when one becomes free:

```yaml
- uses: maximilianfeix/shortfree@main
  id: watch
  with:
    domains: acme.dev acme.io
    fail-when: never   # free | taken | never
- run: echo '${{ steps.watch.outputs.free }}'   # ["acme.dev"] or []
```

Inputs, outputs and a daily watch workflow: [docs/action.md](docs/action.md).

<a id="faq"></a>
## FAQ

**A name is listed as free, but my registrar says it is unavailable or premium. Why?**
The registry has reserved it or prices it as premium. RDAP only says whether a name is registered, not what it costs. This happens mostly with 1–2 character names outside .com, .net and .de. The site and the MCP tools mark these as likely premium.

**How fresh is the data?**
The scan starts at 03:17 UTC every day. Every JSON file has an `updated` field with the scan time. 1–3 character names are swept daily, 4 character names once a week.

**Why are 4 character names not RDAP-verified?**
There are 1,774,224 of them per TLD. At the rate limits shortfree uses (a few requests per second per registry, shared between its TLDs), one pass over a single TLD would take weeks. DNS answers the same question for most names in minutes.

**Why only these TLDs?**
Each TLD needs a registry RDAP service that answers reliably at a polite rate, and each one adds about 300,000 DNS queries a day (all 1–3 character names plus a seventh of the 4 character space). The current list with registries and sources is in [docs/tlds.md](docs/tlds.md); to add one, see [CONTRIBUTING.md](CONTRIBUTING.md).

**Does shortfree sell domains or track me?**
No. There are no affiliate links, no accounts and no analytics. The site and the API are static files on GitHub Pages. The code is MIT, the data CC0.

<a id="contributing"></a>
## Contributing

Issues and pull requests are welcome. [CONTRIBUTING.md](CONTRIBUTING.md) explains how to add a TLD. Before you open a PR:

```sh
go test ./...
go vet ./...
gofmt -l .
```

To run a small scan locally:

```sh
go run ./cmd/shortfree scan -tlds tools -rdap-time 1m   # writes state/state.json.gz
go run ./cmd/shortfree export -tlds tools -readme ''      # writes data/
```

Security issues: see [SECURITY.md](SECURITY.md).

## License

Code: [MIT](LICENSE). Data: [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). Registry data comes from public DNS and RDAP services.
