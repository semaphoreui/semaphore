#!/usr/bin/env python3
"""Generates the keyboard cheat-sheet infographic for the workflow editor docs."""
import html

W = 1120
PAD = 28
COL_GAP = 28
COLS = 3
COL_W = (W - PAD * 2 - COL_GAP * (COLS - 1)) // COLS
ROW_H = 46
HEAD_H = 54
KEY_H = 30
KEY_PAD_X = 10
FONT = "-apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"

# (column title, accent colour, rows) — a row is (keys, label); keys is a list of
# key caps, with "+" between them, or a list of alternatives separated by "/".
COLUMNS = [
    ("Navigate", "#1976d2", [
        ([["←"], ["→"], ["↑"], ["↓"]], "Pan the canvas"),
        ([["Shift", "←→↑↓"]], "Pan in larger steps"),
        ([["Tab"]], "Focus the next node"),
        ([["Enter"]], "Open the focused node"),
        ([["Esc"]], "Deselect / close panel"),
    ]),
    ("Zoom", "#00897b", [
        ([["+"], ["−"]], "Zoom in / out"),
        ([["0"]], "Fit the whole graph"),
        ([["1"]], "Reset zoom to 100 %"),
        ([["Ctrl", "wheel"]], "Zoom towards the pointer"),
        ([["wheel"]], "Pan with the wheel / trackpad"),
    ]),
    ("Edit", "#ab47bc", [
        ([["Delete"]], "Delete selected node or edge"),
        ([["Ctrl", "Z"]], "Undo (up to 50 steps)"),
        ([["Ctrl", "Shift", "Z"]], "Redo"),
        ([["Ctrl", "Y"]], "Redo (alternative)"),
        ([["Right-click"]], "Add a node at the pointer"),
    ]),
]

FOOT = "On macOS use Cmd instead of Ctrl. Shortcuts work while the canvas has focus: click empty canvas first."


def text_width(s, size=13, bold=False):
    # Rough average glyph width; wide enough for key caps.
    per = size * (0.62 if bold else 0.56)
    wide = sum(1 for ch in s if ch in "←→↑↓−+")
    return int(len(s) * per + wide * 3)


def keycap(x, y, label):
    w = max(KEY_H, text_width(label, 13, True) + KEY_PAD_X * 2)
    parts = [
        f'<rect x="{x}" y="{y + 3}" width="{w}" height="{KEY_H}" rx="6" fill="#d5d8dd"/>',
        f'<rect x="{x}" y="{y}" width="{w}" height="{KEY_H}" rx="6" fill="#ffffff" stroke="#c3c7cd"/>',
        f'<text x="{x + w / 2}" y="{y + KEY_H / 2 + 4.5}" text-anchor="middle" font-family="{FONT}" '
        f'font-size="13" font-weight="600" fill="#1f2937">{html.escape(label)}</text>',
    ]
    return w, "".join(parts)


def combo(x, y, alternatives):
    """Draws key alternatives separated by '/', each alternative a chord joined by '+'."""
    out = []
    cx = x
    for ai, chord in enumerate(alternatives):
        if ai > 0:
            out.append(f'<text x="{cx + 6}" y="{y + KEY_H / 2 + 4.5}" font-family="{FONT}" font-size="13" fill="#6b7280">/</text>')
            cx += 18
        for ki, key in enumerate(chord):
            if ki > 0:
                out.append(f'<text x="{cx + 4}" y="{y + KEY_H / 2 + 4.5}" font-family="{FONT}" font-size="13" fill="#6b7280">+</text>')
                cx += 16
            w, svg = keycap(cx, y, key)
            out.append(svg)
            cx += w + 4
    return cx - x, "".join(out)


def build():
    rows = max(len(c[2]) for c in COLUMNS)
    H = PAD + HEAD_H + rows * ROW_H + 18 + 30 + PAD
    parts = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" '
        f'aria-label="Workflow editor keyboard shortcuts">',
        f'<rect x="0.5" y="0.5" width="{W - 1}" height="{H - 1}" rx="14" fill="#fafafa" stroke="#e3e5e8"/>',
    ]
    for ci, (title, accent, items) in enumerate(COLUMNS):
        x0 = PAD + ci * (COL_W + COL_GAP)
        y0 = PAD
        parts.append(f'<rect x="{x0}" y="{y0}" width="{COL_W}" height="{HEAD_H + rows * ROW_H + 8}" rx="12" fill="#ffffff" stroke="#e3e5e8"/>')
        parts.append(f'<rect x="{x0}" y="{y0}" width="{COL_W}" height="4" rx="2" fill="{accent}"/>')
        parts.append(f'<text x="{x0 + 18}" y="{y0 + 34}" font-family="{FONT}" font-size="15" font-weight="700" fill="#111827" letter-spacing="0.04em">{html.escape(title.upper())}</text>')
        for ri, (keys, label) in enumerate(items):
            y = y0 + HEAD_H + ri * ROW_H
            kw, svg = combo(x0 + 18, y, keys)
            parts.append(svg)
            parts.append(f'<text x="{x0 + COL_W - 18}" y="{y + KEY_H / 2 + 4.5}" text-anchor="end" font-family="{FONT}" font-size="13" fill="#374151">{html.escape(label)}</text>')
    parts.append(f'<text x="{PAD}" y="{H - PAD - 4}" font-family="{FONT}" font-size="12.5" fill="#6b7280">{html.escape(FOOT)}</text>')
    parts.append("</svg>")
    return "\n".join(parts)


if __name__ == "__main__":
    import sys
    open(sys.argv[1], "w").write(build())
    print("written", sys.argv[1])
