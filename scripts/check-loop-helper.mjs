#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const repoRoot = path.resolve(import.meta.dirname, '..');
const finder = path.join(repoRoot, 'plugins', 'skill-harness-helpers', 'scripts', 'find-loop.mjs');

function runFinder(args) {
  return spawnSync(process.execPath, [finder, ...args], {
    cwd: repoRoot,
    encoding: 'utf8'
  });
}

function fail(message, detail = '') {
  console.error(message);
  if (detail) console.error(detail.trim());
  process.exit(1);
}

function assert(condition, message, detail = '') {
  if (!condition) fail(message, detail);
}

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'skill-harness-loop-'));
try {
  const catalogPath = path.join(tempDir, 'catalog.json');
  fs.writeFileSync(catalogPath, JSON.stringify({
    updated: '2099-01-01',
    loops: [
      {
        number: '001',
        title: 'The UI Polish Inventory Loop',
        url: 'https://example.test/loops/ui-polish-inventory/',
        category: { label: 'Design' },
        description: 'Inventory every screen, state, input, route, and workflow before polishing UI.',
        useWhen: 'Use this when UI and UX polish require production-like local data and complete state coverage.',
        verification: { title: 'Every inventoried critical flow passes the fixed rubric.' },
        keywords: ['ui', 'ux', 'polish', 'inventory']
      },
      {
        number: '002',
        title: 'The Production Data Cleanup Loop',
        url: 'https://example.test/loops/production-data-cleanup/',
        category: { label: 'Operations' },
        description: 'Clean up production data after policy drift.',
        useWhen: 'Use this when a production dataset contains disallowed records.',
        verification: { title: 'Every remaining record meets policy.' },
        keywords: ['production', 'data']
      }
    ]
  }, null, 2));

  const empty = runFinder([]);
  assert(empty.status === 2, 'empty query should exit with status 2', empty.stderr);
  assert(empty.stderr.includes('Usage:'), 'empty query should print usage', empty.stderr);

  const jsonRun = runFinder(['--file', catalogPath, '--json', 'ui ux polish inventory']);
  assert(jsonRun.status === 0, 'json local catalog run should succeed', jsonRun.stderr);
  const parsed = JSON.parse(jsonRun.stdout);
  assert(parsed.authorization === 'reference-only; does not run or authorize loops', 'json output should preserve authorization boundary');
  assert(parsed.selectionMode === 'lexical shortlist; manual fit review still required', 'json output should identify shortlist mode');
  assert(parsed.catalogUpdated === '2099-01-01', 'json output should include local catalog freshness');
  assert(parsed.results[0]?.title === 'The UI Polish Inventory Loop', 'ui loop should be the top-ranked local result', jsonRun.stdout);
  assert(parsed.results[0]?.matchedTerms.includes('ui'), 'json output should include matched terms', jsonRun.stdout);

  const markdownRun = runFinder(['--file', catalogPath, '--limit', '1', 'production data cleanup']);
  assert(markdownRun.status === 0, 'markdown local catalog run should succeed', markdownRun.stderr);
  assert(markdownRun.stdout.includes('manual fit review is still required'), 'markdown output should state shortlist limitation');
  assert(markdownRun.stdout.includes('does not authorize or run any loop'), 'markdown output should state reference-only authority');
  assert(markdownRun.stdout.includes('The Production Data Cleanup Loop'), 'markdown output should include the top match', markdownRun.stdout);
  assert(!markdownRun.stdout.includes('The UI Polish Inventory Loop'), '--limit 1 should return only one match', markdownRun.stdout);

  const noMatch = runFinder(['--file', catalogPath, 'billing invoice payment']);
  assert(noMatch.status === 0, 'no-match search should finish cleanly', noMatch.stderr);
  assert(noMatch.stdout.includes('No scored matches.'), 'no-match search should give drafting guidance', noMatch.stdout);

  const badLimit = runFinder(['--limit', '0', 'ui']);
  assert(badLimit.status === 1, 'invalid limit should fail', badLimit.stderr);
  assert(badLimit.stderr.includes('--limit requires a positive integer'), 'invalid limit should explain failure', badLimit.stderr);

  const tooMany = runFinder(['--limit', '4', 'ui']);
  assert(tooMany.status === 1, 'limit above three should fail', tooMany.stderr);
  assert(tooMany.stderr.includes('--limit cannot exceed 3'), 'limit above three should explain failure', tooMany.stderr);

  const weakQuery = runFinder(['--file', catalogPath, 'the and of']);
  assert(weakQuery.status === 1, 'stop-word-only query should fail', weakQuery.stderr);
  assert(weakQuery.stderr.includes('at least one searchable term'), 'stop-word-only query should explain failure', weakQuery.stderr);

  const malformedPath = path.join(tempDir, 'malformed.json');
  fs.writeFileSync(malformedPath, JSON.stringify({ updated: '2099-01-01' }));
  const malformed = runFinder(['--file', malformedPath, 'ui']);
  assert(malformed.status === 1, 'malformed catalog should fail closed', malformed.stderr);
  assert(malformed.stderr.includes('catalog must be an object with a loops array'), 'malformed catalog should explain schema failure', malformed.stderr);

  const missingFieldPath = path.join(tempDir, 'missing-field.json');
  fs.writeFileSync(missingFieldPath, JSON.stringify({ loops: [{ title: 'Bad Loop', url: 'https://example.test/bad' }] }));
  const missingField = runFinder(['--file', missingFieldPath, 'bad']);
  assert(missingField.status === 1, 'catalog entries missing required fields should fail closed', missingField.stderr);
  assert(missingField.stderr.includes('missing required string field'), 'missing loop field should explain schema failure', missingField.stderr);

  console.log('Loop helper check passed');
} finally {
  fs.rmSync(tempDir, { recursive: true, force: true });
}
