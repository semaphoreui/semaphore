#!/usr/bin/env python3
"""Checks the AGENTS/ documentation. Two checks — link and path integrity, and file size."""

import argparse
import os
import re
import sys
from urllib.parse import unquote

# [text](target) — target has no spaces, angle brackets are stripped
MD_LINK = re.compile(r'\[[^\]]*\]\(\s*<?([^)\s>]+)>?\s*\)')

# `path` — candidate for a path mentioned in prose
BACKTICK = re.compile(r'`([^`\n]+)`')

# [[name]] — memory note link, the alias after | is dropped
WIKI_LINK = re.compile(r'\[\[([^\]|#]+)(?:\|[^\]]*)?\]\]')

# line-number suffix: file.md:17, file.md:17-29, file.md:17,31
LINE_SUFFIX = re.compile(r':\d+(?:[-,]\d+)*$')

# dir/{a.md,b.md} — a set of files in one folder
BRACE_SET = re.compile(r'^(.*)\{([^{}]+)\}(.*)$')

EXTERNAL_PREFIXES = ('http://', 'https://', 'mailto:', 'tel:', 'ftp://')

# File read thresholds in bytes: above WARNING the channel trims the preview, above ERROR it
# trims the read itself, above HARD the file is not read at all
SIZE_WARNING = 30000
SIZE_ERROR = 61000
SIZE_HARD = 262144

# Things that are not a path: a glob, a cut-off, a placeholder, a variable
PROSE_STOP_CHARS = ('*', '?', '<', '>', '|', '$', '..."', '...')


def repo_root(scan_path):
    """Repository root — the parent of the scanned folder"""
    return os.path.dirname(os.path.abspath(scan_path.rstrip(os.sep))) or os.sep


def alias_dirs(scan_path, root):
    """Folders a file is read from as if it lived there: `agent-primary.md` is read through the
    root symlinks `CLAUDE.md` / `AGENTS.md`, so its links are written from the root — whether the
    symlinks exist yet or not. Any other root symlink into the scanned folder counts the same way"""
    aliases = {}
    primary = os.path.realpath(os.path.join(root, 'AGENTS', 'agent-primary.md'))
    if os.path.isfile(primary):
        aliases[primary] = [root]

    scan_real = os.path.realpath(scan_path)
    for name in os.listdir(root):
        link = os.path.join(root, name)
        if not os.path.islink(link):
            continue
        target = os.path.realpath(link)
        if not target.startswith(scan_real + os.sep):
            continue
        bases = aliases.setdefault(target, [])
        if os.path.dirname(link) not in bases:
            bases.append(os.path.dirname(link))

    return aliases


def load_ignored(ignore_path, root):
    """List of skipped folders: one line per path from the repo root, `#` starts a comment"""
    ignored = set()
    if not os.path.isfile(ignore_path):
        return ignored

    with open(ignore_path, encoding='utf-8') as handle:
        for line in handle:
            entry = line.split('#')[0].strip()
            if not entry:
                continue
            if not os.path.isabs(entry):
                entry = os.path.join(root, entry)
            ignored.add(os.path.normpath(entry))

    return ignored


def collect_files(scan_path, excluded, ignored):
    files = []
    for base, dirs, names in os.walk(scan_path):
        kept = []
        for name in sorted(dirs):
            if name in excluded:
                continue
            if os.path.abspath(os.path.join(base, name)) in ignored:
                continue
            kept.append(name)
        dirs[:] = kept
        for name in sorted(names):
            if name.endswith('.md'):
                files.append(os.path.join(base, name))

    return files


def link_targets(line):
    """Markdown link targets on the line that are worth checking"""
    targets = []
    for match in MD_LINK.finditer(line):
        raw = match.group(1)
        target = unquote(raw.split('#')[0]).strip()
        if not target:
            continue
        if target.startswith(EXTERNAL_PREFIXES):
            continue
        targets.append((raw, target))

    return targets


def memory_dir(scan_path, root):
    """Notes folder: `[[name]]` only points into it"""
    inside = os.path.join(scan_path, 'memory')
    if os.path.isdir(inside):
        return inside

    outside = os.path.join(root, 'AGENTS', 'memory')
    if os.path.isdir(outside):
        return outside

    return None


def wiki_targets(line):
    targets = []
    for match in WIKI_LINK.finditer(line):
        name = match.group(1).strip()
        if not name:
            continue
        targets.append((match.group(0), name.removesuffix('.md')))

    return targets


def expand_braces(token):
    match = BRACE_SET.match(token)
    if not match:
        return [token]

    head, body, tail = match.group(1), match.group(2), match.group(3)
    variants = []
    for part in body.split(','):
        variants.append(head + part.strip() + tail)

    return variants


def prose_targets(line, root):
    """Paths in backticks: only in-repo addresses, globs and cut-offs are dropped"""
    targets = []
    for match in BACKTICK.finditer(line):
        token = match.group(1).strip()
        if any(stop in token for stop in PROSE_STOP_CHARS):
            continue
        if ' ' in token or '/' not in token:
            continue
        # A leading slash means an app route or a URL, not a repo path
        if token.startswith('/'):
            continue

        for variant in expand_braces(token):
            candidate = LINE_SUFFIX.sub('', variant).rstrip('.,;')
            if not candidate.endswith('.md') and not candidate.endswith('/'):
                continue
            # The first segment must be a repo folder, otherwise it's a path from a different root
            head = candidate.split('/')[0]
            if not os.path.isdir(os.path.join(root, head)):
                continue
            targets.append((token, candidate))

    return targets


def link_resolves(target, file_path, bases):
    """A link must resolve relative to its file's folder — this is a style rule, not a convenience"""
    if os.path.isabs(target):
        return os.path.exists(target)

    for base in bases:
        if os.path.exists(os.path.normpath(os.path.join(base, target))):
            return True

    return False


def prose_resolves(target, file_path, root):
    """A prose mention is written from the repo root, less often from its own folder"""
    if os.path.exists(os.path.normpath(os.path.join(root, target))):
        return True

    return os.path.exists(os.path.normpath(os.path.join(os.path.dirname(file_path), target)))


def spaced(number):
    return format(number, ',').replace(',', ' ')


def size_problem(file_path):
    """File size against the read thresholds: NORMAL does not go into the report"""
    size = os.path.getsize(file_path)
    if size > SIZE_HARD:
        status, limit = 'HARD', SIZE_HARD
    elif size > SIZE_ERROR:
        status, limit = 'ERROR', SIZE_ERROR
    elif size > SIZE_WARNING:
        status, limit = 'WARNING', SIZE_WARNING
    else:
        return None

    return (file_path, 'size', status, '%s B, limit %s' % (spaced(size), spaced(limit)))


def check_file(file_path, root, aliases, notes, with_prose):
    bases = [os.path.dirname(file_path)] + aliases.get(os.path.realpath(file_path), [])

    problems = []
    oversize = size_problem(file_path)
    if oversize is not None:
        problems.append(oversize)

    fenced = False
    with open(file_path, encoding='utf-8') as handle:
        for number, line in enumerate(handle, 1):
            text = line.rstrip('\n')

            if text.lstrip().startswith(('```', '~~~')):
                fenced = not fenced

                continue
            # A code block and backticks hold a markup sample, not a link
            if fenced:
                continue

            code_free = BACKTICK.sub(' ', text)

            seen = set()
            for raw, target in link_targets(code_free):
                seen.add(raw)
                if link_resolves(target, file_path, bases):
                    continue
                problems.append((file_path, 'line', number, text))

            for raw, name in wiki_targets(code_free):
                if notes is None or name in notes:
                    continue
                problems.append((file_path, 'line', number, text))

            if not with_prose:
                continue

            for raw, target in prose_targets(text, root):
                # Link text in backticks was already checked as a link
                if raw in seen:
                    continue
                if prose_resolves(target, file_path, root):
                    continue
                problems.append((file_path, 'line', number, text))

    return problems


def unique_lines(problems):
    """A line with two dead links is one report entry: there's no pointer inside it, and a
    duplicate is indistinguishable"""
    seen = set()
    unique = []
    for problem in problems:
        key = problem[:3]
        if key in seen:
            continue
        seen.add(key)
        unique.append(problem)

    return unique


def report(problems, root):
    """Report grouped by file, walk order is preserved"""
    lines = []
    current = None
    for problem in problems:
        file_path, kind, key, text = problem
        if file_path != current:
            if current is not None:
                lines.append('')
            lines.append(os.path.relpath(file_path, root))
            current = file_path
        if kind == 'size':
            lines.append('SIZE #%s' % key)
        else:
            lines.append('LINE #%d' % key)
        lines.append(text.strip())

    return '\n'.join(lines)


def script_file(name):
    return os.path.join(os.path.dirname(os.path.abspath(__file__)), name)


def main():
    parser = argparse.ArgumentParser(description='Checks the AGENTS/ documentation: link and path integrity, file size.')
    parser.add_argument('path', nargs='?', default='AGENTS', help='directory to scan, default AGENTS')
    parser.add_argument('--prose', action='store_true', help='also check paths written in backticks')
    parser.add_argument('--exclude', action='append', default=[], help='directory name to skip during the walk')
    parser.add_argument('--report', default=script_file('report.txt'), help='report file, default report.txt next to the script')
    parser.add_argument('--ignore', default=script_file('ignore.txt'), help='list of skipped folders, default ignore.txt next to the script')
    args = parser.parse_args()

    if not os.path.isdir(args.path):
        print('Directory not found: %s' % args.path, file=sys.stderr)

        return 2

    # The report file is cleared at the start: a past run must not survive the new one
    try:
        report_file = open(args.report, 'w', encoding='utf-8')
    except OSError as error:
        print('Could not open report: %s' % error, file=sys.stderr)

        return 2

    root = repo_root(args.path)
    excluded = set(args.exclude) | {'tmp', '.git', 'node_modules', 'vendor'}
    aliases = alias_dirs(args.path, root)

    notes = None
    notes_dir = memory_dir(args.path, root)
    if notes_dir is not None:
        notes = set()
        for name in os.listdir(notes_dir):
            if name.endswith('.md'):
                notes.add(name[:-3])

    ignored = load_ignored(args.ignore, root)

    problems = []
    for file_path in collect_files(args.path, excluded, ignored):
        problems.extend(check_file(file_path, root, aliases, notes, args.prose))

    problems = unique_lines(problems)
    with report_file:
        if not problems:
            report_file.write('No problems found.\n')
            print('No problems found. Report: %s' % args.report)

            return 0

        report_file.write(report(problems, root))
        report_file.write('\n\nProblems found: %d\n' % len(problems))

    print('Problems found: %d. Report: %s' % (len(problems), args.report))

    return 1


if __name__ == '__main__':
    sys.exit(main())
