import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const checkOnly = process.argv.slice(2).includes('--check');
const config = JSON.parse(fs.readFileSync(path.join(root, '.skill-harness', 'project.json'), 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const modeling = developerArtifacts.modeling ?? developerArtifacts.modelPolicy?.uml ?? {};
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const modelReviewDir = path.join(root, modeling.reviewDir ?? 'generated/review/models');
const requiredCsp = developerArtifacts.htmlPolicy?.requiredCSP ?? "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'";

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function safeName(value) {
  return String(value || 'model').toLowerCase().replace(/[^a-z0-9._-]+/g, '-').replace(/^-+|-+$/g, '') || 'model';
}

function escapeHtml(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function escapeAttribute(value) {
  return String(value ?? '').replaceAll('&', '&amp;').replaceAll('"', '&quot;');
}

function hrefBetween(fromFile, targetPath) {
  if (typeof targetPath !== 'string' || targetPath.trim() === '') return '';
  const resolved = path.resolve(root, targetPath);
  if (!resolved.startsWith(root + path.sep) && resolved !== root) return '';
  return encodeURI(path.relative(path.dirname(fromFile), resolved).replaceAll(path.sep, '/'));
}

function linkFor(fromFile, targetPath, label) {
  const href = hrefBetween(fromFile, targetPath);
  if (!href) return escapeHtml(label ?? targetPath ?? '');
  return '<a href="' + escapeAttribute(href) + '">' + escapeHtml(label ?? targetPath) + '</a>';
}

function readSource(artifact) {
  const fullPath = path.resolve(root, artifact.source ?? '');
  if ((!fullPath.startsWith(root + path.sep) && fullPath !== root) || !fs.existsSync(fullPath) || !fs.statSync(fullPath).isFile()) return '';
  return fs.readFileSync(fullPath, 'utf8');
}

function firstParagraph(markdown) {
  const withoutFrontmatter = String(markdown || '').replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '');
  const withoutFences = withoutFrontmatter.replace(/```[\s\S]*?```/g, '');
  for (const line of withoutFences.split(/\r?\n/).map((item) => item.trim())) {
    if (line && !line.startsWith('#') && !line.startsWith('|') && !line.startsWith('- ')) return line;
  }
  return '';
}

function fencedBlocks(markdown) {
  const blocks = [];
  const pattern = /```([a-zA-Z0-9_-]*)\r?\n([\s\S]*?)```/g;
  let match;
  while ((match = pattern.exec(markdown)) !== null) blocks.push({ language: match[1] || 'text', body: match[2].trim() });
  return blocks;
}

function compactDiagramMarkup(source) {
  const lines = String(source || '').split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  const edges = [];
  for (const line of lines) {
    const arrow = line.match(/^"?([^"\[\](){}:;-][^":;-]*?)"?\s*(?:-->|->>|--|-\)|-\])\s*"?([^":;]+?)"?(?::.*)?$/);
    if (arrow) edges.push([arrow[1].replace(/\[.*$/, '').trim(), arrow[2].replace(/\[.*$/, '').trim()]);
  }
  if (edges.length === 0) return '<pre>' + escapeHtml(source || 'No diagram source found.') + '</pre>';
  const nodes = [...new Set(edges.flat())].filter(Boolean).slice(0, 12);
  const width = Math.max(480, nodes.length * 150);
  const height = 190;
  const nodeByName = new Map(nodes.map((node, index) => [node, { x: 34 + index * 145, y: 72 }]));
  let svg = '<svg role="img" aria-label="Static model diagram preview" viewBox="0 0 ' + width + ' ' + height + '" xmlns="http://www.w3.org/2000/svg">';
  svg += '<defs><marker id="arrow" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto"><path d="M0,0 L0,6 L8,3 z" fill="#0f766e"/></marker></defs>';
  for (const [from, to] of edges) {
    const a = nodeByName.get(from);
    const b = nodeByName.get(to);
    if (!a || !b) continue;
    svg += '<line x1="' + (a.x + 112) + '" y1="' + (a.y + 25) + '" x2="' + b.x + '" y2="' + (b.y + 25) + '" stroke="#0f766e" stroke-width="2" marker-end="url(#arrow)"/>';
  }
  for (const [name, point] of nodeByName.entries()) {
    svg += '<rect x="' + point.x + '" y="' + point.y + '" width="112" height="50" rx="7" fill="#ffffff" stroke="#9fb3c8"/>';
    svg += '<text x="' + (point.x + 56) + '" y="' + (point.y + 30) + '" text-anchor="middle" font-size="12" fill="#17202a">' + escapeHtml(name.slice(0, 22)) + '</text>';
  }
  return svg + '</svg>';
}

function diagramSection(source, artifact) {
  const blocks = fencedBlocks(source);
  const preferred = blocks.find((block) => ['mermaid', 'plantuml', 'puml', 'structurizr'].includes(block.language.toLowerCase())) ?? blocks[0];
  return '<div class="diagram-card"><div class="diagram-header"><strong>' + escapeHtml(artifact.notation || 'model') + ' ' + escapeHtml(artifact.modelKind || 'view') + '</strong><span class="muted">static preview, source-backed</span></div><div class="diagram-body">' + compactDiagramMarkup(preferred?.body || source) + '</div></div>';
}

function imageMime(filePath) {
  switch (path.extname(filePath).toLowerCase()) {
    case '.png': return 'image/png';
    case '.jpg':
    case '.jpeg': return 'image/jpeg';
    case '.gif': return 'image/gif';
    case '.webp': return 'image/webp';
    case '.svg': return 'image/svg+xml';
    default: return '';
  }
}

function imageDataUrl(relativePath) {
  if (typeof relativePath !== 'string' || relativePath.trim() === '') return null;
  const fullPath = path.resolve(root, relativePath);
  if ((!fullPath.startsWith(root + path.sep) && fullPath !== root) || !fs.existsSync(fullPath) || !fs.statSync(fullPath).isFile()) return null;
  const mime = imageMime(fullPath);
  if (!mime) return null;
  const maxBytes = 2 * 1024 * 1024;
  if (fs.statSync(fullPath).size > maxBytes) return null;
  return 'data:' + mime + ';base64,' + fs.readFileSync(fullPath).toString('base64');
}

function artifactImages(artifact) {
  const values = [];
  for (const key of ['screenshots', 'images', 'visualEvidence']) {
    if (Array.isArray(artifact[key])) values.push(...artifact[key]);
  }
  return values.map((entry) => typeof entry === 'string' ? { path: entry, alt: entry } : entry).filter((entry) => entry && typeof entry.path === 'string');
}

function gallerySection(artifact) {
  const figures = [];
  for (const image of artifactImages(artifact)) {
    const dataUrl = imageDataUrl(image.path);
    if (!dataUrl) continue;
    figures.push('<figure><img src="' + escapeAttribute(dataUrl) + '" alt="' + escapeAttribute(image.alt || image.caption || image.path) + '"><figcaption>' + escapeHtml(image.caption || image.path) + '</figcaption></figure>');
  }
  if (figures.length === 0) return '<p class="muted">No screenshot or image evidence is listed for this artifact.</p>';
  return '<div class="gallery">' + figures.join('\n') + '</div>';
}

function listItems(values, emptyText, currentFile) {
  if (!Array.isArray(values) || values.length === 0) return '<p class="muted">' + escapeHtml(emptyText) + '</p>';
  return '<ul>' + values.map((value) => {
    const text = typeof value === 'string' ? value : JSON.stringify(value);
    if (currentFile && typeof value === 'string') return '<li>' + linkFor(currentFile, value, value) + '</li>';
    return '<li>' + escapeHtml(text) + '</li>';
  }).join('') + '</ul>';
}

function htmlPage(title, body) {
  return '<!doctype html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n<meta name="viewport" content="width=device-width, initial-scale=1">\n<meta http-equiv="Content-Security-Policy" content="' + escapeAttribute(requiredCsp) + '">\n<title>' + escapeHtml(title) + '</title>\n<style>\n:root{color-scheme:light dark;font-family:Inter,Segoe UI,Arial,sans-serif;line-height:1.5;--bg:#f7f8fa;--panel:#fff;--text:#17202a;--muted:#52616b;--line:#d9e2ec;--accent:#0f766e;--code:#102a43;--codeText:#f0f4f8}body{margin:0;color:var(--text);background:var(--bg)}main{max-width:1180px;margin:0 auto;padding:28px 18px 44px}.hero{display:grid;grid-template-columns:minmax(0,1.2fr) minmax(240px,.8fr);gap:18px;align-items:start}.panel,section{margin:18px 0;padding:18px;background:var(--panel);border:1px solid var(--line);border-radius:8px}h1,h2,h3{line-height:1.2;margin:0 0 10px;letter-spacing:0}p{margin:0 0 12px}.muted{color:var(--muted)}table{width:100%;border-collapse:collapse;background:var(--panel)}th,td{text-align:left;vertical-align:top;border-bottom:1px solid var(--line);padding:10px}code,pre{font-family:ui-monospace,SFMono-Regular,Consolas,monospace}pre{white-space:pre-wrap;overflow:auto;background:var(--code);color:var(--codeText);padding:16px;border-radius:6px}.meta{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:10px}.meta div{padding:9px 10px;background:#eef6f6;border:1px solid #c8e7e3;border-radius:6px}.tabs{margin-top:18px}.tabs>input{position:absolute;inline-size:1px;block-size:1px;overflow:hidden;clip:rect(0 0 0 0)}.tab-labels{display:flex;flex-wrap:wrap;gap:8px;border-bottom:1px solid var(--line);padding-bottom:10px}.tab-labels label{cursor:pointer;padding:8px 11px;border:1px solid var(--line);border-radius:6px;background:var(--panel);font-weight:600}.tab-panel{display:none}.tabs input:nth-of-type(1):checked~.tab-panels .tab-panel:nth-of-type(1),.tabs input:nth-of-type(2):checked~.tab-panels .tab-panel:nth-of-type(2),.tabs input:nth-of-type(3):checked~.tab-panels .tab-panel:nth-of-type(3),.tabs input:nth-of-type(4):checked~.tab-panels .tab-panel:nth-of-type(4),.tabs input:nth-of-type(5):checked~.tab-panels .tab-panel:nth-of-type(5){display:block}.diagram-card{border:1px solid var(--line);border-radius:8px;overflow:hidden;background:#fbfdff}.diagram-header{display:flex;justify-content:space-between;gap:10px;padding:10px 12px;background:#eef2f7;border-bottom:1px solid var(--line)}.diagram-body{padding:14px}.flow{display:flex;flex-wrap:wrap;gap:10px;align-items:center}.node{padding:10px 12px;border:1px solid #9fb3c8;border-radius:6px;background:#fff;min-width:88px;text-align:center}.arrow{color:var(--accent);font-weight:700}.compare{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:14px}@media (max-width:760px){.hero{grid-template-columns:1fr}.tab-labels label{flex:1 1 auto;text-align:center}}@media (prefers-color-scheme:dark){:root{--bg:#102a43;--panel:#1f2933;--text:#d9e2ec;--muted:#bcccdc;--line:#334e68;--accent:#5eead4}.meta div{background:#243b53;border-color:#486581}.diagram-header{background:#243b53}.diagram-card,.node{background:#1f2933}}\n</style>\n</head>\n<body>\n<main>\n' + body + '\n</main>\n</body>\n</html>\n';
}

function artifactPath(artifact) {
  const fileName = safeName(artifact.reviewSurface ? path.basename(artifact.reviewSurface, '.html') : artifact.modelId || artifact.id) + '.html';
  return path.join(modelReviewDir, fileName);
}

function renderArtifact(artifact, byId, outPath) {
  const source = readSource(artifact);
  const summary = artifact.summary || artifact.purpose || firstParagraph(source) || 'Source-backed model review artifact.';
  const rows = [
    ['ID', artifact.id],
    ['Status', artifact.status],
    ['Model ID', artifact.modelId],
    ['Kind', artifact.modelKind],
    ['Method', artifact.method],
    ['Notation', artifact.notation],
    ['Source', artifact.source],
    ['Owner', artifact.owner]
  ];
  let body = '<div class="hero"><section><h1>' + escapeHtml(artifact.title || artifact.modelId || artifact.id) + '</h1><p>' + escapeHtml(summary) + '</p></section><section><h2>Review Focus</h2><div class="meta"><div><strong>Status</strong><br>' + escapeHtml(artifact.status) + '</div><div><strong>Kind</strong><br>' + escapeHtml(artifact.modelKind) + '</div><div><strong>Drift</strong><br>' + escapeHtml(artifact.driftVerdict || '') + '</div><div><strong>Owner</strong><br>' + escapeHtml(artifact.owner) + '</div></div></section></div><section class="meta">';
  for (const [label, value] of rows) body += '<div><strong>' + escapeHtml(label) + '</strong><br>' + escapeHtml(value) + '</div>';
  body += '</section><div class="tabs"><input id="tab-overview" name="tabs" type="radio" checked><input id="tab-visual" name="tabs" type="radio"><input id="tab-source" name="tabs" type="radio"><input id="tab-evidence" name="tabs" type="radio"><input id="tab-diff" name="tabs" type="radio"><div class="tab-labels"><label for="tab-overview">Overview</label><label for="tab-visual">Visuals</label><label for="tab-source">Source</label><label for="tab-evidence">Evidence</label><label for="tab-diff">Diff</label></div><div class="tab-panels">';
  body += '<section class="tab-panel"><h2>Overview</h2><p>' + escapeHtml(summary) + '</p><div class="meta"><div><strong>Abstraction</strong><br>' + escapeHtml(artifact.abstractionLevel) + '</div><div><strong>Review Surface</strong><br>' + linkFor(outPath, artifact.reviewSurface, artifact.reviewSurface || '') + '</div><div><strong>Implementation Touchpoints</strong><br>' + (artifact.implementationTouchpoints || []).map((item) => linkFor(outPath, item, item)).join(', ') + '</div><div><strong>Doc Touchpoints</strong><br>' + (artifact.docTouchpoints || []).map((item) => linkFor(outPath, item, item)).join(', ') + '</div></div></section>';
  body += '<section class="tab-panel"><h2>Visuals</h2>' + diagramSection(source, artifact) + '</section>';
  body += '<section class="tab-panel"><h2>Canonical Source</h2><p>' + linkFor(outPath, artifact.source, artifact.source || 'source') + '</p><pre>' + escapeHtml(source || 'Source not found or not readable.') + '</pre></section>';
  body += '<section class="tab-panel"><h2>Evidence</h2>' + listItems(artifact.evidenceLinks, 'No evidence links are listed yet.', outPath) + '<h3>Update Triggers</h3>' + listItems(artifact.updateTriggers, 'No update triggers are listed yet.') + '<h3>Freshness</h3><div class="meta"><div><strong>Source Hash</strong><br>' + escapeHtml(artifact.sourceHash || '') + '</div><div><strong>Renderer</strong><br>' + escapeHtml(artifact.renderer || 'skill-harness model review generator') + '</div><div><strong>Generated</strong><br>' + escapeHtml(artifact.generatedAt || artifact.freshness?.generatedAt || 'not-recorded') + '</div></div></section>';
  if (artifact.type === 'model-diff') {
    const diff = artifact.diff ?? {};
    const before = byId.get(diff.beforeArtifactId);
    const after = byId.get(diff.afterArtifactId);
    body += '<section class="tab-panel"><h2>Before And After</h2><div class="compare"><div class="panel"><h3>Before</h3><p><strong>' + escapeHtml(diff.beforeArtifactId) + '</strong></p><pre>' + escapeHtml(before ? readSource(before) : 'Missing before artifact.') + '</pre></div><div class="panel"><h3>After</h3><p><strong>' + escapeHtml(diff.afterArtifactId) + '</strong></p><pre>' + escapeHtml(after ? readSource(after) : 'Missing after artifact.') + '</pre></div></div></section>';
  } else {
    body += '<section class="tab-panel"><h2>Diff</h2><p class="muted">This is a model-view artifact. Create a model-diff artifact to compare model revisions.</p></section>';
  }
  body += '</div></div>';
  return htmlPage('Model Review: ' + (artifact.modelId || artifact.id), body);
}

function isModelArtifact(artifact) {
  return artifact?.type === 'model-view' || artifact?.type === 'model-diff' || typeof artifact?.modelKind === 'string';
}

if (!fs.existsSync(manifestPath)) {
  console.error('Missing artifact manifest: ' + repoPath(manifestPath));
  process.exit(1);
}

fs.mkdirSync(modelReviewDir, { recursive: true });
const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
const artifacts = Array.isArray(manifest.artifacts) ? manifest.artifacts.filter(isModelArtifact) : [];
const byId = new Map();
for (const artifact of artifacts) if (artifact?.id) byId.set(artifact.id, artifact);

const expectedFiles = new Map();
const indexRows = [];
for (const artifact of artifacts) {
  const outPath = artifactPath(artifact);
  const reviewSurface = repoPath(outPath);
  artifact.reviewSurface = reviewSurface;
  if (artifact.type === 'model-diff') {
    artifact.diff = artifact.diff ?? {};
    artifact.diff.reviewSurface = reviewSurface;
  }
  expectedFiles.set(outPath, renderArtifact(artifact, byId, outPath));
  const indexPath = path.join(modelReviewDir, 'index.html');
  indexRows.push('<tr><td>' + escapeHtml(artifact.id) + '</td><td>' + escapeHtml(artifact.modelKind) + '</td><td>' + escapeHtml(artifact.method) + '</td><td>' + escapeHtml(artifact.status) + '</td><td>' + linkFor(indexPath, artifact.source, artifact.source) + '</td><td>' + linkFor(indexPath, reviewSurface, reviewSurface) + '</td></tr>');
}

const indexBody = '<section><h1>Model Review Index</h1><p>Static human review surfaces generated from canonical model sources. Edit source first, then regenerate these pages.</p></section><section><table><thead><tr><th>ID</th><th>Kind</th><th>Method</th><th>Status</th><th>Source</th><th>HTML Review</th></tr></thead><tbody>' + indexRows.join('\n') + '</tbody></table></section>';
expectedFiles.set(path.join(modelReviewDir, 'index.html'), htmlPage('Model Review Index', indexBody));

if (checkOnly) {
  const failures = [];
  for (const [filePath, expected] of expectedFiles) {
    if (!fs.existsSync(filePath)) {
      failures.push('missing generated model review: ' + repoPath(filePath));
      continue;
    }
    if (fs.readFileSync(filePath, 'utf8') !== expected) failures.push('stale generated model review: ' + repoPath(filePath));
  }
  const currentManifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  if (JSON.stringify(currentManifest, null, 2) + '\n' !== JSON.stringify(manifest, null, 2) + '\n') failures.push('manifest reviewSurface entries are stale; run node scripts/generate-model-review.mjs');
  if (failures.length > 0) {
    console.error('Model review drift check failed:');
    for (const failure of failures) console.error('- ' + failure);
    process.exit(1);
  }
  console.log('Model review drift check passed');
} else {
  for (const [filePath, html] of expectedFiles) fs.writeFileSync(filePath, html);
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n');
  console.log('Generated ' + artifacts.length + ' model review artifact(s) in ' + repoPath(modelReviewDir));
}
