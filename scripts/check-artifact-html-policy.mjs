import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const config = JSON.parse(fs.readFileSync(path.join(root, '.skill-harness', 'project.json'), 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const requiredCsp = developerArtifacts.htmlPolicy?.requiredCSP ?? '';
const reviewedSvgPanZoomCsp = "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'";
const reviewRoot = path.join(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const manifest = fs.existsSync(manifestPath) ? JSON.parse(fs.readFileSync(manifestPath, 'utf8')) : {};
const reviewedSvgPanZoomSurfaces = new Set((manifest.artifacts ?? [])
  .filter((artifact) => artifact.htmlInteractionLane === 'reviewed-svg-pan-zoom' && typeof artifact.reviewSurface === 'string')
  .map((artifact) => path.normalize(path.resolve(root, artifact.reviewSurface))));

const blockedTagPatterns = [
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
const blockedAttributePatterns = [
  /\son[a-z]+\s*=/i,
  /\b(?:href|src|action)\s*=\s*["']\s*javascript:/i
];
const externalReferencePattern = /\b(?:src|href|action)=["'](?:https?:|\/\/)/i;
const allowedSvgPanZoomRuntime = fs.existsSync(path.join(root, 'node_modules', 'svg-pan-zoom', 'dist', 'svg-pan-zoom.min.js'))
  ? fs.readFileSync(path.join(root, 'node_modules', 'svg-pan-zoom', 'dist', 'svg-pan-zoom.min.js'), 'utf8')
  : '';
const allowedSvgPanZoomInitializer = 'document.querySelectorAll("[data-svg-pan-zoom=true] svg").forEach(function(svg){svgPanZoom(svg,{controlIconsEnabled:true,fit:true,center:true,minZoom:.1,maxZoom:20,zoomScaleSensitivity:.25});});';

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
  const reviewedSvgPanZoom = reviewedSvgPanZoomSurfaces.has(path.normalize(path.resolve(filePath)));
  const expectedCsp = reviewedSvgPanZoom ? reviewedSvgPanZoomCsp : requiredCsp;
  if (!html.includes('Content-Security-Policy') || !html.includes(expectedCsp)) failures.push('missing required CSP meta tag');
  const scriptMatches = [...html.matchAll(/<script>([\s\S]*?)<\/script>/gi)].map((match) => match[1]);
  if (!reviewedSvgPanZoom && scriptMatches.length > 0) failures.push('blocked script tag');
  if (reviewedSvgPanZoom) {
    if (scriptMatches.length !== 2) failures.push('reviewed-svg-pan-zoom requires exactly two inline scripts');
    if (scriptMatches[0] !== allowedSvgPanZoomRuntime) failures.push('reviewed-svg-pan-zoom runtime does not match bundled svg-pan-zoom');
    if (scriptMatches[1] !== allowedSvgPanZoomInitializer) failures.push('reviewed-svg-pan-zoom initializer is not the approved static initializer');
  }
  for (const pattern of blockedTagPatterns) if (pattern.test(html)) failures.push('blocked tag or preload pattern: ' + pattern);
  for (const pattern of blockedApiPatterns) if (pattern.test(html)) failures.push('blocked browser API: ' + pattern);
  for (const pattern of blockedAttributePatterns) if (pattern.test(html)) failures.push('blocked inline event or javascript URL pattern: ' + pattern);
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
