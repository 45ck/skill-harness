import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const config = JSON.parse(fs.readFileSync(path.join(root, '.skill-harness', 'project.json'), 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const requiredCsp = developerArtifacts.htmlPolicy?.requiredCSP ?? '';
const reviewRoot = path.join(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');

const blockedTagPatterns = [
  /<script\b/i,
  /<iframe\b/i,
  /<object\b/i,
  /<embed\b/i,
  /<form\b/i,
  /<meta\b[^>]*http-equiv=["']?refresh/i,
  /<link\b[^>]*rel=["']?(?:preload|prefetch|preconnect)/i
];
const blockedApiPatterns = [
  /\bfetch\s*\(/,
  /\bXMLHttpRequest\b/,
  /\bWebSocket\b/,
  /\bEventSource\b/,
  /\bsendBeacon\s*\(/,
  /\bserviceWorker\b/,
  /\bdocument\.cookie\b/,
  /\blocalStorage\b/,
  /\bsessionStorage\b/
];
const externalReferencePattern = /\b(?:src|href|action)=["'](?:https?:|\/\/)/i;

function walk(dir) {
  if (!fs.existsSync(dir)) return [];
  const files = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) files.push(...walk(fullPath));
    if (entry.isFile() && entry.name.endsWith('.html')) files.push(fullPath);
  }
  return files;
}

function checkFile(filePath) {
  const html = fs.readFileSync(filePath, 'utf8');
  const failures = [];
  if (!html.includes('Content-Security-Policy') || !html.includes(requiredCsp)) failures.push('missing required CSP meta tag');
  for (const pattern of blockedTagPatterns) if (pattern.test(html)) failures.push('blocked tag or preload pattern: ' + pattern);
  for (const pattern of blockedApiPatterns) if (pattern.test(html)) failures.push('blocked browser API: ' + pattern);
  if (externalReferencePattern.test(html)) failures.push('external src/href/action reference');
  return failures;
}

const failures = [];
for (const filePath of walk(reviewRoot)) {
  for (const failure of checkFile(filePath)) failures.push(path.relative(root, filePath).replaceAll(path.sep, '/') + ': ' + failure);
}

if (failures.length > 0) {
  console.error('Artifact HTML policy failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Artifact HTML policy passed');
