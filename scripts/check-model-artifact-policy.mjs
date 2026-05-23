import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const config = JSON.parse(fs.readFileSync(path.join(root, '.skill-harness', 'project.json'), 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const modelPolicy = developerArtifacts.modelPolicy ?? {};
const modeling = modelPolicy.uml ?? developerArtifacts.modeling ?? {};
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const reviewRoot = path.resolve(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');
const modelReviewRoot = path.resolve(root, modeling.reviewDir ?? 'generated/review/models');
const allowedMethods = new Set(modeling.methods ?? []);
const allowedSourceExtensions = new Set(modeling.allowedSourceExtensions ?? ['.md', '.toon', '.mmd', '.puml', '.plantuml', '.dsl', '.json', '.yaml', '.yml']);
const allowedModelKinds = new Set(modelPolicy.allowedModelKinds ?? []);
const allowedNotations = new Set(modelPolicy.allowedNotations ?? []);
const methodModelKinds = modeling.methodModelKinds ?? {};
const allowedFacets = modeling.allowedFacets ?? {};
const allowedDriftVerdicts = new Set(['aligned', 'source-missing', 'mapping-missing', 'evidence-stale', 'review-stale', 'unsafe']);

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function resolveInsideRoot(relativePath, fieldName, failures) {
  if (typeof relativePath !== 'string' || relativePath.trim() === '') {
    failures.push(fieldName + ' must be a non-empty repo-relative path');
    return null;
  }
  if (path.isAbsolute(relativePath)) {
    failures.push(fieldName + ' must be repo-relative: ' + relativePath);
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

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function isModelArtifact(artifact) {
  return artifact?.type === 'model-view' || artifact?.type === 'model-diff' || typeof artifact?.modelKind === 'string';
}

const failures = [];
if (!developerArtifacts.enabled) failures.push('developerArtifacts must be enabled');
if (!modeling.enabled) failures.push('modelPolicy.uml.enabled must be true for this checker');
if (!fs.existsSync(manifestPath)) failures.push('missing artifact manifest: ' + repoPath(manifestPath));

if (fs.existsSync(manifestPath)) {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  const artifacts = Array.isArray(manifest.artifacts) ? manifest.artifacts : [];
  const byId = new Map();
  for (const artifact of artifacts) if (artifact?.id) byId.set(artifact.id, artifact);

  for (const [index, artifact] of artifacts.entries()) {
    if (!isModelArtifact(artifact)) continue;
    const label = artifact?.id ? 'artifact ' + artifact.id : 'artifact #' + index;

    for (const field of ['id', 'type', 'source', 'status', 'modelId', 'modelKind', 'notation', 'method', 'abstractionLevel', 'owner', 'driftVerdict']) {
      if (typeof artifact[field] !== 'string' || artifact[field].trim() === '') failures.push(label + ' missing required model field: ' + field);
    }
    for (const field of ['implementationTouchpoints', 'docTouchpoints', 'evidenceLinks', 'updateTriggers']) {
      if (!Array.isArray(artifact[field]) || artifact[field].length === 0) failures.push(label + ' missing required non-empty array: ' + field);
    }

    if (artifact.type && !['model-view', 'model-diff'].includes(artifact.type)) failures.push(label + ' model artifacts must use type model-view or model-diff');
    if (artifact.modelKind && allowedModelKinds.size > 0 && !allowedModelKinds.has(artifact.modelKind)) failures.push(label + ' has unsupported modelKind: ' + artifact.modelKind);
    if (artifact.notation && allowedNotations.size > 0 && !allowedNotations.has(artifact.notation)) failures.push(label + ' has unsupported notation: ' + artifact.notation);
    if (artifact.method && allowedMethods.size > 0 && !allowedMethods.has(artifact.method)) failures.push(label + ' has unsupported method: ' + artifact.method);
    if (artifact.method && artifact.modelKind && Array.isArray(methodModelKinds[artifact.method]) && !methodModelKinds[artifact.method].includes(artifact.modelKind)) {
      failures.push(label + ' method ' + artifact.method + ' does not allow modelKind ' + artifact.modelKind);
    }
    if (artifact.driftVerdict && !allowedDriftVerdicts.has(artifact.driftVerdict)) {
      failures.push(label + ' has unsupported driftVerdict: ' + artifact.driftVerdict);
    }

    const facets = asArray(artifact.facets);
    if (artifact.method === 'uwe') {
      if (facets.length === 0) failures.push(label + ' method uwe requires facets');
      const allowed = new Set(allowedFacets.uwe ?? []);
      for (const facet of facets) if (!allowed.has(facet)) failures.push(label + ' has unsupported UWE facet: ' + facet);
    }

    const sourcePath = resolveInsideRoot(artifact.source, label + '.source', failures);
    if (sourcePath) {
      if (!fs.existsSync(sourcePath)) {
        failures.push(label + ' source does not exist: ' + artifact.source);
      } else {
        const ext = path.extname(sourcePath).toLowerCase();
        if (!allowedSourceExtensions.has(ext)) failures.push(label + ' source extension is not allowed for canonical model source: ' + ext);
        if (artifact.sourceHash && artifact.sourceHash !== hashFile(sourcePath)) failures.push(label + ' sourceHash is stale for ' + artifact.source);
      }
    }

    if (artifact.reviewSurface) {
      const reviewPath = resolveInsideRoot(artifact.reviewSurface, label + '.reviewSurface', failures);
      if (reviewPath) {
        if (path.extname(reviewPath) !== '.html') failures.push(label + ' reviewSurface must be an HTML review artifact');
        if (!reviewPath.startsWith(reviewRoot + path.sep) && reviewPath !== reviewRoot) failures.push(label + ' reviewSurface must be under ' + repoPath(reviewRoot));
        if (!reviewPath.startsWith(modelReviewRoot + path.sep) && reviewPath !== modelReviewRoot) failures.push(label + ' modeling reviewSurface should be under ' + repoPath(modelReviewRoot));
        if (artifact.status === 'ready' && !fs.existsSync(reviewPath)) failures.push(label + ' ready model review surface does not exist: ' + artifact.reviewSurface);
      }
    } else if (artifact.status === 'ready') {
      failures.push(label + ' needs a generated HTML reviewSurface');
    }

    if (artifact.type === 'model-diff') {
      const diff = artifact.diff ?? {};
      for (const field of ['beforeArtifactId', 'afterArtifactId', 'method', 'reviewSurface']) {
        if (typeof diff[field] !== 'string' || diff[field].trim() === '') failures.push(label + ' missing diff.' + field);
      }
      if (diff.method && !['source', 'semantic'].includes(diff.method)) failures.push(label + ' diff.method must be source or semantic');
      for (const field of ['beforeArtifactId', 'afterArtifactId']) if (diff[field] && !byId.has(diff[field])) failures.push(label + ' diff.' + field + ' references unknown artifact: ' + diff[field]);
      if (diff.reviewSurface && artifact.reviewSurface && diff.reviewSurface !== artifact.reviewSurface) failures.push(label + ' diff.reviewSurface must match artifact.reviewSurface');
    }
  }
}

if (failures.length > 0) {
  console.error('Model artifact policy failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Model artifact policy passed');
