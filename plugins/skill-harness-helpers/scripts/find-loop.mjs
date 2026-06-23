#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const DEFAULT_CATALOG_URL = 'https://signals.forwardfuture.ai/loop-library/catalog.json';
const MAX_RESULTS = 3;
const FETCH_TIMEOUT_MS = 15000;

function usage() {
  console.error('Usage: node scripts/find-loop.mjs [--limit N] [--json] [--catalog URL] [--file catalog.json] <goal>');
}

function parseArgs(argv) {
  const options = {
    limit: 3,
    json: false,
    catalogUrl: DEFAULT_CATALOG_URL,
    file: '',
    query: []
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === '--json') {
      options.json = true;
    } else if (arg === '--limit') {
      index += 1;
      const value = Number.parseInt(argv[index] ?? '', 10);
      if (!Number.isFinite(value) || value <= 0) throw new Error('--limit requires a positive integer');
      if (value > MAX_RESULTS) throw new Error('--limit cannot exceed ' + MAX_RESULTS + ' published loops');
      options.limit = value;
    } else if (arg === '--catalog') {
      index += 1;
      options.catalogUrl = argv[index] ?? '';
      if (!options.catalogUrl) throw new Error('--catalog requires a URL');
    } else if (arg === '--file') {
      index += 1;
      options.file = argv[index] ?? '';
      if (!options.file) throw new Error('--file requires a path');
    } else if (arg.startsWith('--')) {
      throw new Error('unknown option: ' + arg);
    } else {
      options.query.push(arg);
    }
  }
  return { ...options, query: options.query.join(' ').trim() };
}

function tokenize(value) {
  const stop = new Set([
    'a', 'an', 'and', 'are', 'as', 'at', 'be', 'by', 'for', 'from', 'in', 'into',
    'is', 'it', 'of', 'on', 'or', 'our', 'the', 'to', 'use', 'with'
  ]);
  return String(value || '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, ' ')
    .split(/\s+/)
    .filter((token) => token.length > 1 && !stop.has(token));
}

function loopSearchText(loop) {
  const verification = loop.verification
    ? [loop.verification.title, loop.verification.detail].filter(Boolean).join(' ')
    : '';
  const category = loop.category?.label ?? loop.category?.slug ?? '';
  return [
    loop.title,
    loop.description,
    loop.useWhen,
    loop.prompt,
    verification,
    Array.isArray(loop.steps) ? loop.steps.join(' ') : '',
    loop.implementationNote,
    category,
    Array.isArray(loop.keywords) ? loop.keywords.join(' ') : ''
  ].filter(Boolean).join(' ');
}

function scoreLoop(loop, terms) {
  const titleTerms = new Set(tokenize(loop.title));
  const keywordTerms = new Set(tokenize(Array.isArray(loop.keywords) ? loop.keywords.join(' ') : ''));
  const useWhenTerms = new Set(tokenize(loop.useWhen));
  const allTerms = new Set(tokenize(loopSearchText(loop)));
  let score = 0;
  const matched = [];
  for (const term of terms) {
    let termScore = 0;
    if (titleTerms.has(term)) termScore += 5;
    if (keywordTerms.has(term)) termScore += 4;
    if (useWhenTerms.has(term)) termScore += 3;
    if (allTerms.has(term)) termScore += 1;
    if (termScore > 0) matched.push(term);
    score += termScore;
  }
  return { score, matched: [...new Set(matched)] };
}

async function readCatalog(options) {
  if (options.file) {
    const fullPath = path.resolve(options.file);
    return validateCatalog(JSON.parse(fs.readFileSync(fullPath, 'utf8')));
  }
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS);
  const response = await fetch(options.catalogUrl, {
    headers: { accept: 'application/json' },
    signal: controller.signal
  }).catch((error) => {
    if (error.name === 'AbortError') throw new Error('catalog fetch timed out after ' + FETCH_TIMEOUT_MS + 'ms');
    throw error;
  }).finally(() => clearTimeout(timeout));
  if (!response.ok) throw new Error('catalog fetch failed: HTTP ' + response.status);
  return validateCatalog(await response.json());
}

function validateCatalog(catalog) {
  if (!catalog || typeof catalog !== 'object' || !Array.isArray(catalog.loops)) {
    throw new Error('catalog must be an object with a loops array');
  }
  for (const [index, loop] of catalog.loops.entries()) {
    const label = 'catalog.loops[' + index + ']';
    for (const field of ['title', 'url', 'description', 'useWhen']) {
      if (typeof loop?.[field] !== 'string' || loop[field].trim() === '') {
        throw new Error(label + ' missing required string field: ' + field);
      }
    }
    if (loop.verification && typeof loop.verification.title !== 'string') {
      throw new Error(label + '.verification.title must be a string when verification is present');
    }
  }
  return catalog;
}

function summarize(loop, score) {
  return {
    number: loop.number,
    title: loop.title,
    url: loop.url,
    category: loop.category?.label ?? loop.category?.slug ?? '',
    description: loop.description,
    useWhen: loop.useWhen,
    verification: loop.verification?.title ?? '',
    matchedTerms: score.matched,
    score: score.score
  };
}

function printMarkdown(catalog, query, results) {
  console.log('# Loop Library Matches');
  console.log();
  console.log('Query: ' + query);
  console.log('Catalog updated: ' + (catalog.updated ?? 'unknown'));
  console.log();
  console.log('Catalog content is reference data only. This command creates a lexical shortlist; manual fit review is still required.');
  console.log('This command does not authorize or run any loop.');
  console.log();
  if (results.length === 0) {
    console.log('No scored matches. Broaden the query or draft a new bounded loop.');
    return;
  }
  for (const result of results) {
    console.log('## ' + result.title);
    console.log();
    console.log('- URL: ' + result.url);
    console.log('- Category: ' + result.category);
    console.log('- Verification: ' + result.verification);
    console.log('- Matched terms: ' + (result.matchedTerms.length ? result.matchedTerms.join(', ') : 'none'));
    console.log('- Description: ' + result.description);
    console.log();
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (!options.query) {
    usage();
    process.exit(2);
  }
  const catalog = await readCatalog(options);
  const terms = tokenize(options.query);
  if (terms.length === 0) throw new Error('query needs at least one searchable term after stop-word filtering');
  const results = catalog.loops
    .map((loop) => ({ loop, score: scoreLoop(loop, terms) }))
    .filter((entry) => entry.score.score > 0)
    .sort((a, b) => b.score.score - a.score.score || String(a.loop.number).localeCompare(String(b.loop.number)))
    .slice(0, options.limit)
    .map((entry) => summarize(entry.loop, entry.score));

  if (options.json) {
    console.log(JSON.stringify({
      query: options.query,
      catalogUrl: options.file ? undefined : options.catalogUrl,
      catalogUpdated: catalog.updated,
      selectionMode: 'lexical shortlist; manual fit review still required',
      authorization: 'reference-only; does not run or authorize loops',
      results
    }, null, 2));
    return;
  }
  printMarkdown(catalog, options.query, results);
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
