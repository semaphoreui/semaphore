"""Compare two runs: pixels (BlazeDiff), accessibility tree (.aria.yml) and text inventory (.json).

Output: runtime/visual/compare/<a>__vs__<b>/ — report.html, report.md, summary.json, diff/<shot>.png
(changed pixels red on the faded page, regions boxed blue, the strip only one image has magenta),
aria/<shot>.diff and inventory/<shot>.diff. A pixel share ranks the pairs; it is not a verdict.
"""

import difflib
import html
import json
import os
import shutil
import tempfile
from collections import deque
from pathlib import Path

from .paths import VISUAL, label_dir

DIFF_RED = (255, 0, 0, 255)  # blazediff mask: red — changed, yellow — anti-aliasing (not counted)
PAD = (255, 0, 255, 255)
BOX = (0, 110, 255, 255)
CELL = 16


def _padded(img, size):
    from PIL import Image
    if img.size == size:
        return img
    canvas = Image.new("RGBA", size, PAD)
    canvas.paste(img, (0, 0))
    return canvas


def _regions(mask):
    w, h = mask.size
    gw, gh = (w + CELL - 1) // CELL, (h + CELL - 1) // CELL
    px = mask.load()
    on = {(x // CELL, y // CELL) for y in range(h) for x in range(w) if px[x, y] == DIFF_RED}
    grown = {(cx + dx, cy + dy) for cx, cy in on for dx in (-1, 0, 1) for dy in (-1, 0, 1)
             if 0 <= cx + dx < gw and 0 <= cy + dy < gh}
    seen, boxes = set(), []
    for start in sorted(grown):
        if start in seen:
            continue
        seen.add(start)
        queue, cells = deque([start]), []
        while queue:
            c = queue.popleft()
            cells.append(c)
            for d in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                nb = (c[0] + d[0], c[1] + d[1])
                if nb in grown and nb not in seen:
                    seen.add(nb)
                    queue.append(nb)
        hit = [c for c in cells if c in on]
        x0, y0 = min(c[0] for c in hit) * CELL, min(c[1] for c in hit) * CELL
        x1 = min((max(c[0] for c in hit) + 1) * CELL, w)
        y1 = min((max(c[1] for c in hit) + 1) * CELL, h)
        boxes.append((x0, y0, x1 - x0, y1 - y0))
    return sorted(boxes, key=lambda b: (b[1], b[0]))


def pixel_diff(a_path, b_path, out_path, threshold):
    """(differing pixels, total pixels, size a, size b, regions); writes out_path when they differ."""
    from blazediff import compare
    from PIL import Image, ImageDraw

    a = Image.open(a_path).convert("RGBA")
    b = Image.open(b_path).convert("RGBA")
    size = (max(a.width, b.width), max(a.height, b.height))
    total = size[0] * size[1]
    with tempfile.TemporaryDirectory() as tmp:
        pa, pb = Path(tmp, "a.png"), Path(tmp, "b.png")
        _padded(a, size).save(pa)
        _padded(b, size).save(pb)
        res = compare(str(pa), str(pb), str(Path(tmp, "d.png")), threshold=threshold, antialiasing=True)
        if res.match_result:
            return 0, total, a.size, b.size, []
        count = res.diff_count or 0
        mask = Path(tmp, "mask.png")
        compare(str(pa), str(pb), str(mask), threshold=threshold, antialiasing=True, diff_mask=True)
        boxes = _regions(Image.open(mask).convert("RGBA"))
        img = Image.open(Path(tmp, "d.png")).convert("RGBA")
        draw = ImageDraw.Draw(img)
        for x, y, w, h in boxes:
            draw.rectangle((max(x - 3, 0), max(y - 3, 0), x + w + 2, y + h + 2), outline=BOX, width=3)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        img.save(out_path)
        if count == 0 and a.size != b.size:  # blazediff gives no count on a layout difference
            count = total - min(a.width, b.width) * min(a.height, b.height)
    return count, total, a.size, b.size, boxes


def text_diff(a_path, b_path, out_path, name_a, name_b):
    a = a_path.read_text().splitlines(keepends=True) if a_path.is_file() else []
    b = b_path.read_text().splitlines(keepends=True) if b_path.is_file() else []
    if a == b:
        return 0
    lines = list(difflib.unified_diff(a, b, fromfile=name_a, tofile=name_b))
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("".join(lines))
    return sum(1 for ln in lines if ln[:1] in "+-" and not ln.startswith(("+++", "---")))


def run(label_a, label_b, threshold=0.1):
    dir_a, dir_b = label_dir(label_a), label_dir(label_b)
    for d, lab in ((dir_a, label_a), (dir_b, label_b)):
        if not d.is_dir():
            raise SystemExit(f"vu: no run '{lab}' in {VISUAL} — ./vu.sh run {lab}")
    out = VISUAL / "compare" / f"{label_a}__vs__{label_b}"
    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)

    stems = sorted({p.stem for p in dir_a.glob("*.png")} | {p.stem for p in dir_b.glob("*.png")})
    rows = []
    for stem in stems:
        pa, pb = dir_a / f"{stem}.png", dir_b / f"{stem}.png"
        row = {"shot": stem, "a": pa.is_file(), "b": pb.is_file(), "pixels": 0, "total": 0, "ratio": 0.0,
               "size_a": None, "size_b": None, "regions": [], "aria": 0, "inventory": 0}
        if row["a"] and row["b"]:
            count, total, sa, sb, boxes = pixel_diff(pa, pb, out / "diff" / f"{stem}.png", threshold)
            row.update(pixels=count, total=total, ratio=count / total if total else 0.0,
                       size_a=list(sa), size_b=list(sb), regions=boxes)
        row["aria"] = text_diff(dir_a / f"{stem}.aria.yml", dir_b / f"{stem}.aria.yml",
                                out / "aria" / f"{stem}.diff", f"{label_a}/{stem}.aria.yml",
                                f"{label_b}/{stem}.aria.yml")
        row["inventory"] = text_diff(dir_a / f"{stem}.json", dir_b / f"{stem}.json",
                                     out / "inventory" / f"{stem}.diff", f"{label_a}/{stem}.json",
                                     f"{label_b}/{stem}.json")
        row["status"] = ("only " + (label_a if row["a"] else label_b)) if not (row["a"] and row["b"]) else (
            "differs" if row["pixels"] or row["aria"] or row["inventory"] or row["size_a"] != row["size_b"]
            else "same")
        rows.append(row)

    rows.sort(key=lambda r: (r["status"] == "same", not r["status"].startswith("only"), -r["ratio"],
                             -(r["aria"] + r["inventory"]), r["shot"]))
    summary = {
        "a": label_a, "b": label_b, "threshold": threshold, "pairs": len(rows),
        "same": sum(r["status"] == "same" for r in rows),
        "pixel_diffs": sum(r["pixels"] > 0 for r in rows),
        "size_diffs": sum(r["a"] and r["b"] and r["size_a"] != r["size_b"] for r in rows),
        "aria_diffs": sum(r["aria"] > 0 for r in rows),
        "inventory_diffs": sum(r["inventory"] > 0 for r in rows),
        "missing": sum(r["status"].startswith("only") for r in rows),
    }
    (out / "summary.json").write_text(json.dumps({"summary": summary, "rows": rows}, indent=2) + "\n")
    _markdown(out, rows, summary, dir_a, dir_b)
    _html(out, rows, summary, dir_a, dir_b)
    print(f"pairs {summary['pairs']}: same {summary['same']}, pixel diffs {summary['pixel_diffs']}, "
          f"size diffs {summary['size_diffs']}, aria diffs {summary['aria_diffs']}, "
          f"inventory diffs {summary['inventory_diffs']}, missing {summary['missing']}")
    print(f"report: {out / 'report.html'}")
    return 0 if summary["same"] == summary["pairs"] else 1


def _rel(path, out):
    return os.path.relpath(path, out)


def _size(s):
    return f"{s[0]}×{s[1]}" if s else "—"


def _markdown(out, rows, s, dir_a, dir_b):
    lines = [f"# {s['a']} vs {s['b']}", "",
             f"Pairs {s['pairs']}: same {s['same']}, pixel diffs {s['pixel_diffs']}, size diffs "
             f"{s['size_diffs']}, aria diffs {s['aria_diffs']}, inventory diffs {s['inventory_diffs']}, "
             f"missing {s['missing']}. Threshold {s['threshold']}.", "",
             "| Shot | Status | Pixels | Size a → b | Aria lines | Inventory lines | Links |",
             "| --- | --- | --- | --- | --- | --- | --- |"]
    for r in rows:
        links = []
        if r["a"]:
            links.append(f"[a]({_rel(dir_a / (r['shot'] + '.png'), out)})")
        if r["b"]:
            links.append(f"[b]({_rel(dir_b / (r['shot'] + '.png'), out)})")
        if r["pixels"]:
            links.append(f"[diff](diff/{r['shot']}.png)")
        if r["aria"]:
            links.append(f"[aria](aria/{r['shot']}.diff)")
        if r["inventory"]:
            links.append(f"[inventory](inventory/{r['shot']}.diff)")
        lines.append(f"| `{r['shot']}` | {r['status']} | {100 * r['ratio']:.4f}% | "
                     f"{_size(r['size_a'])} → {_size(r['size_b'])} | {r['aria']} | {r['inventory']} | "
                     f"{' '.join(links)} |")
    (out / "report.md").write_text("\n".join(lines) + "\n")


def _html(out, rows, s, dir_a, dir_b):
    def img(path):
        rel = html.escape(_rel(path, out))
        return f'<a href="{rel}"><img src="{rel}" loading="lazy"></a>'

    body = []
    for r in rows:
        if r["status"] == "same":
            continue
        cells = [img(dir_a / f"{r['shot']}.png") if r["a"] else "—",
                 img(dir_b / f"{r['shot']}.png") if r["b"] else "—",
                 img(out / "diff" / f"{r['shot']}.png") if r["pixels"] else "—"]
        notes = []
        for kind in ("aria", "inventory"):
            if r[kind]:
                p = out / kind / f"{r['shot']}.diff"
                notes.append(f'<details><summary>{kind}: {r[kind]} lines</summary><pre>'
                             f'{html.escape(p.read_text())}</pre></details>')
        body.append(
            f"<tr><td><code>{html.escape(r['shot'])}</code><br>{r['status']}<br>"
            f"{100 * r['ratio']:.4f}%<br>{_size(r['size_a'])} → {_size(r['size_b'])}{''.join(notes)}</td>"
            + "".join(f"<td>{c}</td>" for c in cells) + "</tr>")
    same = [r["shot"] for r in rows if r["status"] == "same"]
    doc = f"""<!doctype html><meta charset="utf-8"><title>{s['a']} vs {s['b']}</title>
<style>body{{font:14px sans-serif;margin:16px}}table{{border-collapse:collapse}}td{{border:1px solid #ccc;
vertical-align:top;padding:6px}}img{{max-width:420px;max-height:600px}}pre{{max-width:420px;overflow:auto;
font-size:12px}}</style>
<h1>{html.escape(s['a'])} vs {html.escape(s['b'])}</h1>
<p>Pairs {s['pairs']}: same {s['same']}, pixel diffs {s['pixel_diffs']}, size diffs {s['size_diffs']},
aria diffs {s['aria_diffs']}, inventory diffs {s['inventory_diffs']}, missing {s['missing']}.
Threshold {s['threshold']}. Sorted by status, then share of differing pixels.</p>
<table><tr><th>Shot</th><th>{html.escape(s['a'])}</th><th>{html.escape(s['b'])}</th><th>diff</th></tr>
{''.join(body) or '<tr><td colspan="4">no differences</td></tr>'}</table>
<details><summary>{len(same)} identical</summary><pre>{html.escape(chr(10).join(same))}</pre></details>
"""
    (out / "report.html").write_text(doc)
