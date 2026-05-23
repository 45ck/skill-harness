import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const root = process.cwd();
const args = process.argv.slice(2);
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = fs.existsSync(configPath) ? JSON.parse(fs.readFileSync(configPath, 'utf8')) : {};
const manifestPath = config.capabilities?.developerArtifacts?.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json';
const manifest = JSON.parse(fs.readFileSync(path.join(root, manifestPath), 'utf8'));

function normalize(value) {
  return String(value ?? '').replaceAll('\\', '/').replace(/^.\//, '');
}

function changedFiles() {
  const explicitIndex = args.indexOf('--files');
  if (explicitIndex >= 0) return args.slice(explicitIndex + 1).filter((item) => !item.startsWith('--')).map(normalize);
  if (args.includes('--changed')) {
    const tracked = execFileSync('git', ['diff', '--name-only', 'HEAD'], { cwd: root, encoding: 'utf8' });
    const untracked = execFileSync('git', ['ls-files', '--others', '--exclude-standard'], { cwd: root, encoding: 'utf8' });
    return [...new Set((tracked + '\n' + untracked).split(/\r?\n/).map(normalize).filter(Boolean))];
  }
  return [];
}

function touches(pattern, filePath) {
  const normalized = normalize(pattern);
  if (normalized.endsWith('/')) return filePath.startsWith(normalized);
  if (normalized.includes('*')) {
    const escaped = normalized.replace(/[.+?^${}()|[\]\\]/g, '\\$&').replaceAll('\\*', '[^/]*');
    return new RegExp('^' + escaped + '$').test(filePath);
  }
  return filePath === normalized || filePath.startsWith(normalized + '/');
}

const files = changedFiles();
const modelArtifacts = (manifest.artifacts ?? []).filter((artifact) => artifact.type === 'model-view' || artifact.type === 'model-diff');
const impacted = [];

for (const artifact of modelArtifacts) {
  const touchpoints = [...(artifact.implementationTouchpoints ?? []), ...(artifact.docTouchpoints ?? []), artifact.source].filter(Boolean);
  const matchedFiles = files.filter((filePath) => touchpoints.some((touchpoint) => touches(touchpoint, filePath)));
  if (matchedFiles.length > 0) {
    const sourceChanged = matchedFiles.includes(normalize(artifact.source));
    impacted.push({
      modelId: artifact.modelId,
      artifactId: artifact.id,
      matchedFiles,
      verdict: sourceChanged ? 'canonical-model-update-required' : 'evidence-only',
    });
  }
}

if (files.length === 0) {
  console.log(JSON.stringify({ changedFiles: [], impactedModels: [], verdict: 'none', note: 'pass --changed or --files <paths...>' }, null, 2));
} else {
  console.log(JSON.stringify({ changedFiles: files, impactedModels: impacted, verdict: impacted.length === 0 ? 'none' : 'model-impact-detected' }, null, 2));
}
