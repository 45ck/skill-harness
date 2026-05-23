import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { pathToFileURL } from 'node:url';

const root = process.cwd();
const args = process.argv.slice(2);
const jsonMode = args.includes('--json');
const printOnly = args.includes('--print') || args.includes('--dry-run') || process.env.CI === 'true';
const explicitTarget = args.find((arg) => !arg.startsWith('--'));

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function resolveReviewPath(value) {
  if (typeof value !== 'string' || value.trim() === '') return null;
  const resolved = path.resolve(root, value);
  if (!resolved.startsWith(root + path.sep) && resolved !== root) return null;
  return resolved;
}

function readJSON(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch {
    return null;
  }
}

function firstExisting(paths) {
  for (const candidate of paths) if (candidate && fs.existsSync(candidate) && fs.statSync(candidate).isFile()) return candidate;
  return null;
}

function discoverTarget() {
  if (explicitTarget) {
    const resolved = resolveReviewPath(explicitTarget);
    if (resolved && fs.existsSync(resolved)) return resolved;
    throw new Error('review artifact not found or outside repo: ' + explicitTarget);
  }
  const config = readJSON(path.join(root, '.skill-harness', 'project.json')) ?? {};
  const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
  const reviewDir = developerArtifacts.reviewSurface?.outDir ?? 'generated/review';
  const modelReviewDir = developerArtifacts.modeling?.reviewDir ?? developerArtifacts.modelPolicy?.uml?.reviewDir ?? path.join(reviewDir, 'models');
  const manifest = readJSON(path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json'));
  const manifestTargets = [];
  for (const artifact of Array.isArray(manifest?.artifacts) ? manifest.artifacts : []) {
    if (typeof artifact?.reviewSurface === 'string' && artifact.reviewSurface.endsWith('.html')) {
      const resolved = resolveReviewPath(artifact.reviewSurface);
      if (resolved) manifestTargets.push(resolved);
    }
  }
  const discovered = firstExisting([path.join(root, modelReviewDir, 'index.html'), path.join(root, reviewDir, 'index.html'), ...manifestTargets]);
  if (discovered) return discovered;
  throw new Error('no generated HTML review artifact found; generate one first');
}

function hostHint() {
  const originator = process.env.CODEX_INTERNAL_ORIGINATOR_OVERRIDE ?? '';
  if (process.env.CODEX_THREAD_ID || /codex/i.test(originator)) return 'Codex app detected: prefer opening this file with the Browser plugin when the agent has it.';
  if (process.env.CLAUDE_DESKTOP || /claude/i.test(originator)) return 'Claude desktop context detected: prefer the built-in browser or preview tool when available.';
  return '';
}

function hostAction() {
  const originator = process.env.CODEX_INTERNAL_ORIGINATOR_OVERRIDE ?? '';
  if (process.env.CODEX_THREAD_ID || /codex/i.test(originator)) return 'codex-browser-plugin';
  if (process.env.CLAUDE_DESKTOP || /claude/i.test(originator)) return 'claude-desktop-preview';
  if (printOnly) return 'print-file-url';
  return 'system-default-browser';
}

function openSystemDefault(filePath) {
  const url = pathToFileURL(filePath).href;
  let command;
  let commandArgs;
  if (process.platform === 'win32') {
    command = 'cmd';
    commandArgs = ['/c', 'start', '', url];
  } else if (process.platform === 'darwin') {
    command = 'open';
    commandArgs = [url];
  } else {
    command = 'xdg-open';
    commandArgs = [url];
  }
  const child = spawn(command, commandArgs, { detached: true, stdio: 'ignore' });
  child.unref();
  return url;
}

try {
  const target = discoverTarget();
  const url = pathToFileURL(target).href;
  const hint = hostHint();
  if (jsonMode) {
    console.log(JSON.stringify({
      path: target,
      repoPath: repoPath(target),
      url,
      hostAction: hostAction(),
      openMode: printOnly ? 'print' : 'open',
      hint
    }, null, 2));
    process.exit(0);
  }
  if (hint) console.log(hint);
  if (printOnly) {
    console.log(url);
  } else {
    console.log('Opening ' + repoPath(target));
    console.log(openSystemDefault(target));
  }
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
