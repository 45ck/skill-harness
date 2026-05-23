import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const reviewRoot = path.resolve(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');
const allowedTypes = new Set(developerArtifacts.artifactTypes ?? []);
const allowedModelKinds = new Set(developerArtifacts.modelPolicy?.allowedModelKinds ?? []);
const allowedNotations = new Set(developerArtifacts.modelPolicy?.allowedNotations ?? []);

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function resolveInsideRoot(relativePath, fieldName, failures) {
  if (typeof relativePath !== 'string' || relativePath.trim() === '') return null;
  if (path.isAbsolute(relativePath)) {
    failures.push(fieldName + ' must be a repo-relative path: ' + relativePath);
    return null;
  }
  const resolved = path.resolve(root, relativePath);
  if (!resolved.startsWith(root + path.sep) && resolved !== root) {
    failures.push(fieldName + ' escapes the repo root: ' + relativePath);
    return null;
  }
  return resolved;
}

function hashFile(filePath) {
  return crypto.createHash('sha256').update(fs.readFileSync(filePath)).digest('hex');
}

const failures = [];
if (!fs.existsSync(manifestPath)) {
  failures.push('missing artifact manifest: ' + repoPath(manifestPath));
} else {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  if (manifest.version !== 1) failures.push('manifest.version must be 1');
  if (!Array.isArray(manifest.artifacts)) failures.push('manifest.artifacts must be an array');

  for (const [index, artifact] of (manifest.artifacts ?? []).entries()) {
    const label = artifact?.id ? 'artifact ' + artifact.id : 'artifact #' + index;
    if (!artifact || typeof artifact !== 'object') {
      failures.push(label + ' must be an object');
      continue;
    }
    for (const field of ['id', 'type', 'source', 'status']) {
      if (typeof artifact[field] !== 'string' || artifact[field].trim() === '') failures.push(label + ' missing required string field: ' + field);
    }
    if (artifact.type && allowedTypes.size > 0 && !allowedTypes.has(artifact.type)) failures.push(label + ' has unsupported type: ' + artifact.type);
    if (artifact.modelKind && allowedModelKinds.size > 0 && !allowedModelKinds.has(artifact.modelKind)) failures.push(label + ' has unsupported modelKind: ' + artifact.modelKind);
    if (artifact.notation && allowedNotations.size > 0 && !allowedNotations.has(artifact.notation)) failures.push(label + ' has unsupported notation: ' + artifact.notation);

    const sourcePath = resolveInsideRoot(artifact.source, label + '.source', failures);
    if (sourcePath && !fs.existsSync(sourcePath)) failures.push(label + ' source does not exist: ' + artifact.source);
    if (sourcePath && fs.existsSync(sourcePath) && artifact.sourceHash && artifact.sourceHash !== hashFile(sourcePath)) {
      failures.push(label + ' sourceHash is stale for ' + artifact.source);
    }

    if (artifact.reviewSurface) {
      const reviewPath = resolveInsideRoot(artifact.reviewSurface, label + '.reviewSurface', failures);
      if (reviewPath && path.extname(reviewPath) === '.html' && !reviewPath.startsWith(reviewRoot + path.sep)) {
        failures.push(label + ' HTML review surface must be under ' + repoPath(reviewRoot));
      }
      if (artifact.status === 'ready' && reviewPath && !fs.existsSync(reviewPath)) failures.push(label + ' ready review surface does not exist: ' + artifact.reviewSurface);
    }
    if (artifact.status === 'ready' && (!Array.isArray(artifact.evidenceLinks) || artifact.evidenceLinks.length === 0)) {
      failures.push(label + ' ready artifact needs evidenceLinks');
    }
  }
}

if (failures.length > 0) {
  console.error('Artifact manifest policy failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Artifact manifest policy passed');
