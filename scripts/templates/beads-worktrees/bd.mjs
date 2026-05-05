#!/usr/bin/env node
/**
 * Beads worktree wrapper (repo-local).
 *
 * Adds:
 * - explicit claiming (`claimed_by` / `claimed_at`)
 * - "unblocked before start" checks (via `dependencies[].type === "blocks"`)
 * - per-issue git worktrees under `.trees/<id>`
 *
 * Expects upstream Beads JSONL at `.beads/issues.jsonl`.
 */

import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

// Allow piping to `head`/`rg` without noisy stack traces.
process.stdout.on('error', (err) => {
  if (err?.code === 'EPIPE') process.exit(0);
});

const VALID_PRIORITIES = ['P0', 'P1', 'P2', 'P3', '0', '1', '2', '3', 0, 1, 2, 3];
const PRIORITY_ORDER = { P0: 0, P1: 1, P2: 2, P3: 3, 0: 0, 1: 1, 2: 2, 3: 3, undefined: 4 };

function normalizePriority(p) {
  if (p === null || p === undefined) return undefined;
  if (typeof p === 'number') return `P${p}`;
  if (typeof p === 'string' && /^\d$/.test(p.trim())) return `P${p.trim()}`;
  if (typeof p === 'string' && /^P[0-3]$/.test(p.trim())) return p.trim();
  return undefined;
}

function usage() {
  const lines = [
    'bd (worktrees + claiming) — repo-local wrapper',
    '',
    'Usage:',
    '  bd ready [--priority <P0|P1|P2|P3>] [--claimed|--unclaimed|--claimed-by "<owner>"] [--json]',
    '  bd list  [--open] [--priority <P0|P1|P2|P3>] [--claimed|--unclaimed|--claimed-by "<owner>"] [--json]',
    '  bd show  <id> [--json]',
    '',
    '  bd issue claim   <id> --by "<owner>" [--force] [--json]',
    '  bd issue unclaim <id> [--by "<owner>"] [--force] [--json]',
    '  bd issue start   <id> --by "<owner>" [--base <branch>] [--json]',
    '  bd issue finish  <id> [--no-merge] [--base <branch>] [--json]',
    '',
    'Notes:',
    '  - Issues are stored in `.beads/issues.jsonl` (JSON Lines).',
    '  - Worktrees are created under `.trees/<id>` from the base branch (default: main).',
    '  - Blockers are inferred from `dependencies[]` entries where `type == "blocks"`.',
  ];
  process.stdout.write(lines.join('\n') + '\n');
}

function fail(message) {
  console.error(message);
  process.exit(1);
}

function nowIsoUtc() {
  return new Date().toISOString();
}

function spawnGit(args, cwd) {
  const r = spawnSync('git', args, { cwd, encoding: 'utf8' });
  return { status: r.status ?? 1, stdout: r.stdout ?? '', stderr: r.stderr ?? '' };
}

function repoRoot() {
  return process.cwd();
}

function beadsDir(root) {
  return path.join(root, '.beads');
}

function issuesPath(root) {
  return path.join(beadsDir(root), 'issues.jsonl');
}

function ensureTreesDir(root) {
  fs.mkdirSync(path.join(root, '.trees'), { recursive: true });
}

function readIssues(root) {
  const filePath = issuesPath(root);
  if (!fs.existsSync(filePath)) return [];
  const raw = fs.readFileSync(filePath, 'utf8');
  const lines = raw.split('\n').filter((l) => l.trim().length > 0);

  const issues = [];
  for (const [idx, line] of lines.entries()) {
    try {
      issues.push(JSON.parse(line));
    } catch {
      throw new Error(`Invalid JSON in ${path.relative(root, filePath)} at line ${idx + 1}.`);
    }
  }
  return issues;
}

function writeIssues(root, issues) {
  const filePath = issuesPath(root);
  const lines = issues.map((i) => JSON.stringify(i));
  fs.writeFileSync(filePath, lines.join('\n') + '\n', 'utf8');
}

function isNonEmptyString(value) {
  return typeof value === 'string' && value.trim().length > 0;
}

function readOption(argv, name) {
  const idx = argv.indexOf(name);
  if (idx < 0) return null;
  const value = argv[idx + 1];
  if (!isNonEmptyString(value)) throw new Error(`Missing value for ${name}.`);
  return value;
}

function parseJsonFlag(argv) {
  return argv.includes('--json');
}

function getIssueStatus(issue) {
  const status = issue?.status;
  if (!isNonEmptyString(status)) return 'open';
  return status.trim();
}

function isClosed(issue) {
  return getIssueStatus(issue) === 'closed';
}

function closedIdsSet(issues) {
  return new Set(issues.filter((i) => isClosed(i)).map((i) => i.id));
}

function readClaim(issue) {
  const claimedBy = issue?.claimed_by;
  const claimedAt = issue?.claimed_at;
  return {
    claimedBy: isNonEmptyString(claimedBy) ? claimedBy.trim() : null,
    claimedAt: isNonEmptyString(claimedAt) ? claimedAt.trim() : null,
  };
}

function hasClaim(issue) {
  return readClaim(issue).claimedBy !== null;
}

function readClaimFilters(argv) {
  const claimedOnly = argv.includes('--claimed');
  const unclaimedOnly = argv.includes('--unclaimed');
  if (claimedOnly && unclaimedOnly) throw new Error('Cannot combine --claimed and --unclaimed.');

  const claimedByRaw = readOption(argv, '--claimed-by');
  const claimedBy = claimedByRaw ? claimedByRaw.trim() : null;

  return { claimedOnly, unclaimedOnly, claimedBy };
}

function applyClaimFilters(issues, claimFilters) {
  let result = issues;
  if (claimFilters.claimedOnly) result = result.filter((issue) => hasClaim(issue));
  if (claimFilters.unclaimedOnly) result = result.filter((issue) => !hasClaim(issue));
  if (claimFilters.claimedBy) {
    result = result.filter((issue) => readClaim(issue).claimedBy === claimFilters.claimedBy);
  }
  return result;
}

function blockersForIssue(issue) {
  const deps = Array.isArray(issue?.dependencies) ? issue.dependencies : [];
  return deps
    .filter((d) => d?.type === 'blocks' && isNonEmptyString(d?.depends_on_id))
    .map((d) => d.depends_on_id);
}

function isUnblocked(issue, closedIds) {
  const blockers = blockersForIssue(issue);
  if (blockers.length === 0) return true;
  return blockers.every((depId) => closedIds.has(depId));
}

function prioritySortKey(issue) {
  return PRIORITY_ORDER[normalizePriority(issue?.priority)] ?? PRIORITY_ORDER['undefined'];
}

function sortByPriority(issues) {
  return [...issues].sort((a, b) => {
    const pd = prioritySortKey(a) - prioritySortKey(b);
    if (pd !== 0) return pd;
    return (a.created_at ?? '').localeCompare(b.created_at ?? '');
  });
}

function formatPriority(p) {
  const n = normalizePriority(p);
  if (!n) return '  --';
  return n;
}

function formatClaimSuffix(issue) {
  const { claimedBy } = readClaim(issue);
  if (!claimedBy) return '';
  return ` [claimed:${claimedBy}]`;
}

function print(value, asJson) {
  if (asJson) {
    process.stdout.write(JSON.stringify(value, null, 2) + '\n');
  } else {
    process.stdout.write(String(value) + '\n');
  }
}

function printIssueList(issues, asJson) {
  if (asJson) {
    print(issues, true);
    return;
  }
  for (const issue of issues) {
    const status = getIssueStatus(issue);
    const priority = formatPriority(issue?.priority);
    process.stdout.write(`${priority} ${status.padEnd(11)} ${issue.id}  ${issue.title}${formatClaimSuffix(issue)}\n`);
  }
}

function findIssueIndex(issues, id) {
  const idx = issues.findIndex((i) => i?.id === id);
  if (idx < 0) throw new Error(`Issue not found: ${id}`);
  return idx;
}

function applyUpdate(issues, id, updates) {
  const idx = findIssueIndex(issues, id);
  const now = nowIsoUtc();
  const existing = issues[idx];
  issues[idx] = { ...existing, ...updates, updated_at: now };
  return issues[idx];
}

function requireRepoRootHasBeads(root) {
  if (!fs.existsSync(beadsDir(root))) {
    fail(`.beads/ not found in ${root}. Run from repo root.`);
  }
  if (!fs.existsSync(issuesPath(root))) {
    fail(`.beads/issues.jsonl not found. Initialize Beads or create the file.`);
  }
}

function defaultBaseBranch(rawArgv) {
  const baseRaw = readOption(rawArgv, '--base');
  return (baseRaw ? baseRaw.trim() : 'main') || 'main';
}

function main() {
  const rawArgv = process.argv.slice(2);
  if (rawArgv.length === 0 || rawArgv.includes('--help') || rawArgv.includes('-h')) {
    usage();
    return;
  }

  const root = repoRoot();
  requireRepoRootHasBeads(root);

  const asJson = parseJsonFlag(rawArgv);
  const cmd = rawArgv[0];
  const rest = rawArgv.slice(1);

  const issues = readIssues(root);

  if (cmd === 'ready') {
    const priorityFilter = readOption(rawArgv, '--priority');
    if (priorityFilter && !VALID_PRIORITIES.includes(priorityFilter)) {
      fail(`Invalid priority "${priorityFilter}". Expected P0..P3 or 0..3.`);
    }
    const claimFilters = readClaimFilters(rawArgv);

    const closedIds = closedIdsSet(issues);
    let result = issues.filter((i) => !isClosed(i) && isUnblocked(i, closedIds));
    if (priorityFilter) {
      const want = normalizePriority(priorityFilter);
      result = result.filter((i) => normalizePriority(i?.priority) === want);
    }
    result = applyClaimFilters(result, claimFilters);
    printIssueList(sortByPriority(result), asJson);
    return;
  }

  if (cmd === 'list') {
    const openOnly = rawArgv.includes('--open');

    const priorityFilter = readOption(rawArgv, '--priority');
    if (priorityFilter && !VALID_PRIORITIES.includes(priorityFilter)) {
      fail(`Invalid priority "${priorityFilter}". Expected P0..P3 or 0..3.`);
    }
    const claimFilters = readClaimFilters(rawArgv);

    let result = issues;
    if (openOnly) result = result.filter((i) => !isClosed(i));
    if (priorityFilter) {
      const want = normalizePriority(priorityFilter);
      result = result.filter((i) => normalizePriority(i?.priority) === want);
    }
    result = applyClaimFilters(result, claimFilters);
    printIssueList(sortByPriority(result), asJson);
    return;
  }

  if (cmd === 'show') {
    const id = rest[0];
    if (!isNonEmptyString(id)) fail('Missing issue id for show.');
    const idx = findIssueIndex(issues, id);
    print(issues[idx], true);
    return;
  }

  if (cmd !== 'issue') {
    fail(`Unknown command: ${cmd}\n\nRun: bd --help`);
  }

  const verb = rest[0];
  const sub = rest.slice(1);

  if (verb === 'claim') {
    const id = sub[0];
    if (!isNonEmptyString(id)) fail('Missing issue id for issue claim.');

    const byRaw = readOption(rawArgv, '--by');
    if (!byRaw) fail('Missing --by for issue claim.');
    const by = byRaw.trim();
    const force = rawArgv.includes('--force');

    const idx = findIssueIndex(issues, id);
    const issue = issues[idx];
    const existing = readClaim(issue).claimedBy;
    if (existing && existing !== by && !force) {
      fail(`Issue ${id} is claimed by "${existing}", not "${by}". Use --force to override.`);
    }

    const now = nowIsoUtc();
    issues[idx] = { ...issue, claimed_by: by, claimed_at: now, updated_at: now };
    writeIssues(root, issues);
    print(issues[idx], asJson);
    return;
  }

  if (verb === 'unclaim') {
    const id = sub[0];
    if (!isNonEmptyString(id)) fail('Missing issue id for issue unclaim.');

    const byRaw = readOption(rawArgv, '--by');
    const by = byRaw ? byRaw.trim() : null;
    const force = rawArgv.includes('--force');

    const idx = findIssueIndex(issues, id);
    const issue = issues[idx];
    const existing = readClaim(issue).claimedBy;
    if (!existing) fail(`Issue ${id} is not claimed.`);

    if (by && existing !== by && !force) {
      fail(`Issue ${id} is claimed by "${existing}", not "${by}". Use --force to clear.`);
    }

    const updated = applyUpdate(issues, id, { claimed_by: undefined, claimed_at: undefined });
    writeIssues(root, issues);
    print(updated, asJson);
    return;
  }

  if (verb === 'start') {
    const id = sub[0];
    if (!isNonEmptyString(id)) fail('Missing issue id for issue start.');

    const byRaw = readOption(rawArgv, '--by');
    if (!byRaw) fail('Missing --by for issue start.');
    const by = byRaw.trim();
    const baseBranch = defaultBaseBranch(rawArgv);

    const idx = findIssueIndex(issues, id);
    const issue = issues[idx];
    if (isClosed(issue)) fail(`Cannot start closed issue: ${id}.`);

    const closedIds = closedIdsSet(issues);
    if (!isUnblocked(issue, closedIds)) {
      const openBlockers = blockersForIssue(issue).filter((dep) => !closedIds.has(dep));
      fail(`Issue ${id} is blocked by: ${openBlockers.join(', ')}`);
    }

    ensureTreesDir(root);
    const worktreePath = path.join(root, '.trees', id);
    if (fs.existsSync(worktreePath)) {
      fail(`Worktree already exists at .trees/${id}. Remove it or run issue finish.`);
    }

    const gitResult = spawnGit(
      ['worktree', 'add', '-b', id, path.join('.trees', id), baseBranch],
      root,
    );
    if (gitResult.status !== 0) {
      fail(`git worktree add failed:\n${gitResult.stderr}`);
    }

    const now = nowIsoUtc();
    issues[idx] = { ...issue, claimed_by: by, claimed_at: now, updated_at: now, status: 'in_progress' };
    writeIssues(root, issues);
    print(issues[idx], asJson);
    return;
  }

  if (verb === 'finish') {
    const id = sub[0];
    if (!isNonEmptyString(id)) fail('Missing issue id for issue finish.');

    const noMerge = rawArgv.includes('--no-merge');

    const idx = findIssueIndex(issues, id);
    const issue = issues[idx];
    if (isClosed(issue)) fail(`Issue ${id} is already closed.`);

    const worktreePath = path.join(root, '.trees', id);
    if (!fs.existsSync(worktreePath)) {
      fail(`Worktree .trees/${id} not found. Did you run issue start?`);
    }

    if (!noMerge) {
      const mergeResult = spawnGit(['merge', '--no-ff', id, '-m', `merge: ${id}`], root);
      if (mergeResult.status !== 0) fail(`Merge failed:\n${mergeResult.stderr}`);
    }

    const removeResult = spawnGit(['worktree', 'remove', path.join('.trees', id)], root);
    if (removeResult.status !== 0) spawnGit(['worktree', 'remove', '--force', path.join('.trees', id)], root);
    spawnGit(['branch', '-d', id], root);

    const now = nowIsoUtc();
    issues[idx] = { ...issue, status: 'closed', claimed_by: undefined, claimed_at: undefined, closed_at: now, updated_at: now };
    writeIssues(root, issues);
    print(issues[idx], asJson);
    return;
  }

  fail(`Unknown issue verb: ${verb}\n\nRun: bd --help`);
}

try {
  main();
} catch (err) {
  fail(err instanceof Error ? err.message : String(err));
}
