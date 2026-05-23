import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const checkOnly = process.argv.slice(2).includes('--check');
const rendererName = 'skill-harness artifact review generator';
const config = JSON.parse(fs.readFileSync(path.join(root, '.skill-harness', 'project.json'), 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const reviewRoot = path.resolve(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');
const requiredCsp = developerArtifacts.htmlPolicy?.requiredCSP ?? "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'";
const families = new Set(['product', 'business', 'data', 'research', 'ux']);
const defaultInfographicTools = [
  { id: 'mermaid', label: 'Mermaid', role: 'architecture, workflow, sequence, and model diagrams', output: 'pre-rendered inline SVG or static markup' },
  { id: 'vega-lite', label: 'Vega-Lite', role: 'default declarative charts for metrics, comparisons, and evidence dashboards', output: 'static SVG generated from source specs' },
  { id: 'observable-plot', label: 'Observable Plot', role: 'compact exploratory charts and statistical views', output: 'static SVG generated from source specs' },
  { id: 'd3', label: 'D3', role: 'custom infographic layouts when canned charts are not expressive enough', output: 'static SVG generated during artifact generation' },
  { id: 'graphviz', label: 'Graphviz', role: 'node-edge dependency, lineage, and relationship maps', output: 'static SVG generated from DOT or structured edges' },
  { id: 'echarts', label: 'Apache ECharts', role: 'dashboard-style chart families when a richer chart grammar is useful', output: 'static SVG or PNG generated outside the browser runtime' },
  { id: 'rawgraphs', label: 'RAWGraphs', role: 'design-led or unusual infographic forms using tabular data', output: 'exported SVG copied into the generated review surface' },
  { id: 'chartjs', label: 'Chart.js', role: 'simple familiar charts when existing source data already matches Chart.js conventions', output: 'server-rendered image or static SVG equivalent' }
];
const svgPanZoomRuntime = (() => {
  try {
    return fs.readFileSync(path.join(root, 'node_modules', 'svg-pan-zoom', 'dist', 'svg-pan-zoom.min.js'), 'utf8');
  } catch {
    return '';
  }
})();
const cytoscapeRuntime = (() => {
  try {
    return fs.readFileSync(path.join(root, 'node_modules', 'cytoscape', 'dist', 'cytoscape.min.js'), 'utf8');
  } catch {
    return '';
  }
})();
const dagreRuntime = (() => {
  try {
    return fs.readFileSync(path.join(root, 'node_modules', 'dagre', 'dist', 'dagre.min.js'), 'utf8');
  } catch {
    return '';
  }
})();
const cytoscapeDagreRuntime = (() => {
  try {
    return fs.readFileSync(path.join(root, 'node_modules', 'cytoscape-dagre', 'cytoscape-dagre.js'), 'utf8');
  } catch {
    return '';
  }
})();
const uweWorkspaceRuntime = (() => {
  try {
    return fs.readFileSync(path.join(root, 'scripts', 'uwe-workspace-runtime.js'), 'utf8');
  } catch {
    return '';
  }
})();
const configuredInfographicTools = Array.isArray(developerArtifacts.infographicPolicy?.tools)
  ? developerArtifacts.infographicPolicy.tools
  : defaultInfographicTools;
const infographicTools = configuredInfographicTools.map((tool) => typeof tool === 'string'
  ? (defaultInfographicTools.find((candidate) => candidate.id === normalizeToolId(tool)) ?? { id: normalizeToolId(tool), label: tool, role: 'source-declared infographic renderer', output: 'static review output' })
  : {
      id: normalizeToolId(tool.id ?? tool.name ?? tool.label),
      label: tool.label ?? tool.name ?? tool.id,
      role: tool.role ?? 'source-declared infographic renderer',
      output: tool.output ?? 'static review output'
    });
const infographicToolIds = new Set(infographicTools.map((tool) => tool.id));
let graphvizInstancePromise;

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function isInside(parent, child) {
  const relative = path.relative(parent, child);
  return relative === '' || (!!relative && !relative.startsWith('..') && !path.isAbsolute(relative));
}

function safeName(value) {
  return String(value || 'artifact').toLowerCase().replace(/[^a-z0-9._-]+/g, '-').replace(/^-+|-+$/g, '') || 'artifact';
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

function dotQuote(value) {
  return '"' + String(value ?? '').replaceAll('\\', '\\\\').replaceAll('"', '\\"') + '"';
}

function dotHtmlText(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
}

function normalizeToolId(value) {
  const raw = String(value || '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
  if (['vega', 'vega-lite', 'vegalite'].includes(raw)) return 'vega-lite';
  if (['plot', 'observable', 'observable-plot', 'observablehq-plot'].includes(raw)) return 'observable-plot';
  if (['apache-echarts', 'echarts'].includes(raw)) return 'echarts';
  if (['chart-js', 'chartjs'].includes(raw)) return 'chartjs';
  if (['raw-graphs', 'rawgraphs'].includes(raw)) return 'rawgraphs';
  return raw;
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

function hashSource(artifact) {
  const fullPath = path.resolve(root, artifact.source ?? '');
  if ((!fullPath.startsWith(root + path.sep) && fullPath !== root) || !fs.existsSync(fullPath) || !fs.statSync(fullPath).isFile()) return '';
  return crypto.createHash('sha256').update(fs.readFileSync(fullPath)).digest('hex');
}

function sourceGeneratedAt(artifact) {
  const match = readSource(artifact).match(/^freshness:\s*\r?\n(?:\s+.+\r?\n)*?\s+generatedAt:\s*([0-9]{4}-[0-9]{2}-[0-9]{2})/m);
  return match ? match[1] : '';
}

function firstParagraph(markdown) {
  const withoutFrontmatter = String(markdown || '').replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '');
  const fence = String.fromCharCode(96).repeat(3);
  const withoutFences = withoutFrontmatter.replace(new RegExp(fence + '[\\s\\S]*?' + fence, 'g'), '');
  for (const line of withoutFences.split(/\r?\n/).map((item) => item.trim())) {
    if (line && !line.startsWith('#') && !line.startsWith('|') && !line.startsWith('- ') && !line.match(/^\d+\./)) return line;
  }
  return '';
}

function sourceTitle(markdown) {
  const withoutFrontmatter = String(markdown || '').replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '');
  const match = withoutFrontmatter.match(/^#\s+(.+)$/m);
  return match ? match[1].trim() : '';
}

function headings(markdown) {
  return String(markdown || '')
    .split(/\r?\n/)
    .map((line) => line.match(/^(#{2,3})\s+(.+)$/))
    .filter(Boolean)
    .map((match) => ({ level: match[1].length, text: match[2].trim() }))
    .slice(0, 8);
}

function familyFor(artifact) {
  if (families.has(artifact.family)) return artifact.family;
  const source = String(artifact.source ?? '').replaceAll('\\', '/');
  const match = source.match(/docs\/artifacts\/source\/([^/]+)\//);
  if (match && families.has(match[1])) return match[1];
  if (['research-synthesis', 'claim-evidence-matrix'].includes(artifact.type)) return 'research';
  if (['product-brief', 'opportunity-brief', 'planning-artifact', 'e2e-product-system-atlas'].includes(artifact.type)) return 'product';
  if (['business-case', 'stakeholder-map'].includes(artifact.type)) return 'business';
  if (['data-dictionary', 'metric-definition', 'lineage-map'].includes(artifact.type)) return 'data';
  if (['high-fidelity-prototype', 'interaction-state-board', 'journey-map', 'visual-review'].includes(artifact.type)) return 'ux';
  return 'review';
}

function defaultReviewSurface(artifact) {
  const family = familyFor(artifact);
  if (families.has(family)) return 'generated/review/' + family + '/' + safeName(artifact.id) + '.html';
  return 'generated/review/' + safeName(artifact.id) + '.html';
}

function isModelArtifact(artifact) {
  return artifact?.type === 'model-view' || artifact?.type === 'model-diff' || typeof artifact?.modelKind === 'string';
}

function isManagedArtifact(artifact) {
  if (!artifact || isModelArtifact(artifact)) return false;
  if (artifact.renderer === rendererName) return true;
  return artifact.reviewRequired === true && !artifact.reviewSurface;
}

function resolveReviewSurface(artifact) {
  const outPath = path.resolve(root, artifact.reviewSurface ?? '');
  if (!isInside(reviewRoot, outPath) || path.extname(outPath).toLowerCase() !== '.html') {
    throw new Error('artifact ' + (artifact.id ?? '<unknown>') + ' reviewSurface must be an HTML file under ' + repoPath(reviewRoot));
  }
  return outPath;
}

function listItems(values, emptyText, currentFile) {
  if (!Array.isArray(values) || values.length === 0) return '<p class="muted">' + escapeHtml(emptyText) + '</p>';
  return '<ul>' + values.map((value) => {
    const text = typeof value === 'string' ? value : JSON.stringify(value);
    if (currentFile && typeof value === 'string') return '<li>' + linkFor(currentFile, value, value) + '</li>';
    return '<li>' + escapeHtml(text) + '</li>';
  }).join('') + '</ul>';
}

function sourceStats(source) {
  const lines = String(source || '').split(/\r?\n/).filter((line) => line.trim() !== '').length;
  const sectionCount = headings(source).filter((heading) => heading.level === 2).length;
  const tableCount = (String(source || '').match(/\n\|.+\|\r?\n/g) ?? []).length;
  return { lines, sectionCount, tableCount };
}

function parseInfographicSpecs(source, artifact) {
  const specs = [];
  if (Array.isArray(artifact.infographics)) specs.push(...artifact.infographics);
  const fence = String.fromCharCode(96).repeat(3);
  const pattern = new RegExp('^' + fence + '(?:artifact-infographic|infographic)\\s*\\r?\\n([\\s\\S]*?)\\r?\\n' + fence, 'gm');
  for (const match of String(source || '').matchAll(pattern)) {
    try {
      specs.push(JSON.parse(match[1]));
    } catch (error) {
      specs.push({
        title: 'Invalid Infographic Spec',
        tool: 'source-spec',
        kind: 'notice',
        summary: 'The source contains an infographic block that is not valid JSON: ' + error.message
      });
    }
  }
  return specs.map((spec, index) => ({ ...spec, id: spec.id ?? 'infographic-' + (index + 1) }));
}

function numericSeries(spec) {
  let values = spec.values ?? spec.data ?? spec.dataset ?? [];
  if (values && !Array.isArray(values) && Array.isArray(values.values)) values = values.values;
  if (!Array.isArray(values)) return [];
  const labelField = spec.labelField ?? spec.xField ?? spec.categoryField ?? 'label';
  const valueField = spec.valueField ?? spec.yField ?? spec.metricField ?? 'value';
  return values.map((item, index) => {
    if (typeof item === 'number') return { label: 'Item ' + (index + 1), value: item };
    if (Array.isArray(item)) return { label: String(item[0] ?? 'Item ' + (index + 1)), value: Number(item[1] ?? 0) || 0 };
    if (item && typeof item === 'object') {
      return {
        label: String(item[labelField] ?? item.name ?? item.id ?? 'Item ' + (index + 1)),
        value: Number(item[valueField] ?? item.value ?? item.count ?? 0) || 0
      };
    }
    return { label: String(item ?? 'Item ' + (index + 1)), value: 0 };
  }).slice(0, 10);
}

function graphData(spec) {
  const edges = Array.isArray(spec.edges) ? spec.edges.map((edge) => Array.isArray(edge)
    ? { from: String(edge[0] ?? ''), to: String(edge[1] ?? ''), label: String(edge[2] ?? '') }
    : { from: String(edge.from ?? edge.source ?? ''), to: String(edge.to ?? edge.target ?? ''), label: String(edge.label ?? '') }) : [];
  const nodeIds = new Set();
  for (const edge of edges) {
    if (edge.from) nodeIds.add(edge.from);
    if (edge.to) nodeIds.add(edge.to);
  }
  const nodes = Array.isArray(spec.nodes) && spec.nodes.length > 0
    ? spec.nodes.map((node) => typeof node === 'string' ? { id: node, label: node } : {
      id: String(node.id ?? node.name ?? node.label),
      label: String(node.label ?? node.name ?? node.id),
      route: node.route ? String(node.route) : '',
      facet: node.facet ? String(node.facet) : '',
      role: node.role ? String(node.role) : '',
      navigationClass: node.navigationClass || node.lane || node.class ? String(node.navigationClass ?? node.lane ?? node.class) : '',
      effect: node.effect || node.sideEffect ? String(node.effect ?? node.sideEffect) : '',
      screenshot: node.screenshot || node.image || node.visualEvidence || '',
      actions: node.actions || node.primaryActions || node.action ? String(node.actions ?? node.primaryActions ?? node.action) : '',
      stereotype: node.stereotype || node.stereo ? String(node.stereotype ?? node.stereo) : '',
      type: node.type || node.facet ? String(node.type ?? node.facet) : ''
    })
    : [...nodeIds].map((id) => ({ id, label: id }));
  return { nodes, edges: edges.filter((edge) => edge.from && edge.to) };
}

function renderBarSvg(spec) {
  const series = numericSeries(spec);
  if (series.length === 0) return '<p class="muted">No numeric series was provided for this infographic spec.</p>';
  const max = Math.max(...series.map((item) => item.value), 1);
  const width = 720;
  const height = 300;
  const chartTop = 28;
  const chartBottom = 248;
  const slot = 620 / series.length;
  const bars = series.map((item, index) => {
    const barHeight = Math.max(4, (item.value / max) * 180);
    const x = 72 + index * slot + slot * 0.15;
    const y = chartBottom - barHeight;
    const barWidth = Math.max(18, slot * 0.7);
    return '<rect x="' + x.toFixed(1) + '" y="' + y.toFixed(1) + '" width="' + barWidth.toFixed(1) + '" height="' + barHeight.toFixed(1) + '" rx="5" fill="#0f766e"></rect><text x="' + (x + barWidth / 2).toFixed(1) + '" y="' + (y - 8).toFixed(1) + '" text-anchor="middle" font-size="12" fill="#1f2937">' + escapeHtml(item.value) + '</text><text x="' + (x + barWidth / 2).toFixed(1) + '" y="274" text-anchor="middle" font-size="11" fill="#5b6472">' + escapeHtml(item.label.slice(0, 14)) + '</text>';
  }).join('');
  return '<svg class="infographic-chart" viewBox="0 0 ' + width + ' ' + height + '" role="img" aria-label="' + escapeAttribute(spec.title ?? 'bar chart') + '"><line x1="58" y1="' + chartBottom + '" x2="700" y2="' + chartBottom + '" stroke="#d8dee8"></line><line x1="58" y1="' + chartTop + '" x2="58" y2="' + chartBottom + '" stroke="#d8dee8"></line>' + bars + '</svg>';
}

function renderLineSvg(spec) {
  const series = numericSeries(spec);
  if (series.length === 0) return '<p class="muted">No numeric series was provided for this infographic spec.</p>';
  const max = Math.max(...series.map((item) => item.value), 1);
  const min = Math.min(...series.map((item) => item.value), 0);
  const span = Math.max(max - min, 1);
  const points = series.map((item, index) => {
    const x = 70 + (index * (620 / Math.max(series.length - 1, 1)));
    const y = 240 - ((item.value - min) / span) * 180;
    return { ...item, x, y };
  });
  const polyline = points.map((point) => point.x.toFixed(1) + ',' + point.y.toFixed(1)).join(' ');
  const dots = points.map((point) => '<circle cx="' + point.x.toFixed(1) + '" cy="' + point.y.toFixed(1) + '" r="5" fill="#2457c5"></circle><text x="' + point.x.toFixed(1) + '" y="274" text-anchor="middle" font-size="11" fill="#5b6472">' + escapeHtml(point.label.slice(0, 12)) + '</text>').join('');
  return '<svg class="infographic-chart" viewBox="0 0 720 300" role="img" aria-label="' + escapeAttribute(spec.title ?? 'line chart') + '"><line x1="58" y1="248" x2="700" y2="248" stroke="#d8dee8"></line><line x1="58" y1="28" x2="58" y2="248" stroke="#d8dee8"></line><polyline points="' + polyline + '" fill="none" stroke="#2457c5" stroke-width="4" stroke-linejoin="round" stroke-linecap="round"></polyline>' + dots + '</svg>';
}

async function renderGraphSvg(spec) {
  const graph = graphData(spec);
  if (graph.nodes.length === 0) return '<p class="muted">No graph nodes or edges were provided for this infographic spec.</p>';
  const kind = String(spec.kind ?? spec.mark ?? spec.type ?? '').toLowerCase();
  if (kind === 'uwe-navigation' || kind === 'uwe' || graph.nodes.some((node) => node.screenshot)) return await renderUweNavigationGraphvizSvg(spec, graph);
  const cx = 360;
  const cy = 170;
  const radius = 104;
  const positions = new Map(graph.nodes.map((node, index) => {
    const angle = (Math.PI * 2 * index) / Math.max(graph.nodes.length, 1) - Math.PI / 2;
    return [node.id, { x: cx + Math.cos(angle) * radius, y: cy + Math.sin(angle) * radius }];
  }));
  const edges = graph.edges.map((edge) => {
    const from = positions.get(edge.from);
    const to = positions.get(edge.to);
    if (!from || !to) return '';
    return '<line x1="' + from.x.toFixed(1) + '" y1="' + from.y.toFixed(1) + '" x2="' + to.x.toFixed(1) + '" y2="' + to.y.toFixed(1) + '" stroke="#8ea0b8" stroke-width="2"></line>';
  }).join('');
  const nodes = graph.nodes.map((node) => {
    const pos = positions.get(node.id);
    return '<g><circle cx="' + pos.x.toFixed(1) + '" cy="' + pos.y.toFixed(1) + '" r="30" fill="#effaf8" stroke="#0f766e" stroke-width="2"></circle><text x="' + pos.x.toFixed(1) + '" y="' + (pos.y + 4).toFixed(1) + '" text-anchor="middle" font-size="11" fill="#1f2937">' + escapeHtml(node.label.slice(0, 12)) + '</text></g>';
  }).join('');
  return '<svg class="infographic-chart" viewBox="0 0 720 340" role="img" aria-label="' + escapeAttribute(spec.title ?? 'relationship graph') + '">' + edges + nodes + '</svg>';
}

async function graphvizSvg(dot) {
  graphvizInstancePromise = graphvizInstancePromise || import('@viz-js/viz').then((module) => module.instance());
  const viz = await graphvizInstancePromise;
  return viz.renderString(dot, { format: 'svg' })
    .replace(/^<\?xml[\s\S]*?<svg/, '<svg')
    .replace(/<svg /, '<svg class="infographic-chart graphviz-render" ');
}

function uweGraphvizDot(spec, graph) {
  const classNames = Array.isArray(spec.navigationClasses ?? spec.lanes ?? spec.classes)
    ? (spec.navigationClasses ?? spec.lanes ?? spec.classes).map((item) => typeof item === 'string' ? item : (item.label ?? item.id ?? item.name)).filter(Boolean).map(String)
    : [];
  for (const node of graph.nodes) {
    const name = node.navigationClass || 'Navigation';
    if (!classNames.includes(name)) classNames.push(name);
  }
  const lines = [
    'digraph UweNavigation {',
    '  graph [rankdir=LR, bgcolor="transparent", pad="0.22", nodesep="0.42", ranksep="0.78", compound=true, fontname="Segoe UI", label=' + dotQuote(spec.title ?? 'UWE navigation model') + ', labelloc=t, fontcolor="#111827"];',
    '  node [shape=plain, margin=0, fontname="Segoe UI", fontsize=12];',
    '  edge [fontname="Segoe UI", fontsize=10, color="#111827", fontcolor="#111827", arrowsize=0.72, arrowhead=open];'
  ];
  for (const className of classNames) {
    const nodes = graph.nodes.filter((node) => (node.navigationClass || 'Navigation') === className);
    if (nodes.length === 0) continue;
    lines.push('  subgraph cluster_' + safeName(className).replaceAll('-', '_') + ' {');
    lines.push('    label=' + dotQuote('«navigation class» ' + className) + ';');
    lines.push('    color="#111827";');
    lines.push('    fillcolor="#ffffff";');
    lines.push('    style="filled";');
    for (const node of nodes) lines.push('    ' + dotQuote(node.id) + ' [label=<' + uweNodeHtmlLabel(node) + '>];');
    lines.push('  }');
  }
  for (const edge of graph.edges) lines.push('  ' + dotQuote(edge.from) + ' -> ' + dotQuote(edge.to) + (edge.label ? ' [label=' + dotQuote(edge.label) + ']' : '') + ';');
  lines.push('}');
  return lines.join('\n');
}

function uweNodeHtmlLabel(node) {
  const label = dotHtmlText(String(node.label || node.id).slice(0, 36));
  const route = dotHtmlText(String(node.route || node.facet || node.role || '').slice(0, 42));
  const screenshotToken = 'screenshot:' + dotHtmlText(node.id);
  return '<TABLE BORDER="1" CELLBORDER="0" CELLSPACING="0" CELLPADDING="6" COLOR="#111827">' +
    '<TR><TD BGCOLOR="#f3f4f6"><FONT FACE="Segoe UI" POINT-SIZE="11" COLOR="#111827">«navigation node»</FONT></TD></TR>' +
    '<TR><TD FIXEDSIZE="TRUE" WIDTH="320" HEIGHT="184" BGCOLOR="#ffffff"><FONT FACE="Segoe UI" POINT-SIZE="9" COLOR="#6b7280">' + screenshotToken + '</FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="12"><B>' + label + '</B></FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="10" COLOR="#334155">' + route + '</FONT></TD></TR>' +
    '</TABLE>';
}

function polygonBounds(group) {
  const match = group.match(/<polygon\b[^>]*\bpoints="([^"]+)"/);
  let points = [];
  if (match) {
    points = match[1].trim().split(/\s+/).map((point) => point.split(',').map(Number)).filter((point) => point.length === 2 && point.every(Number.isFinite));
  } else {
    const pathMatch = group.match(/<path\b[^>]*\bd="([^"]+)"/);
    if (!pathMatch) return null;
    const numbers = [...pathMatch[1].matchAll(/-?\d+(?:\.\d+)?/g)].map((item) => Number(item[0]));
    for (let index = 0; index + 1 < numbers.length; index += 2) points.push([numbers[index], numbers[index + 1]]);
  }
  if (points.length === 0) return null;
  const xs = points.map((point) => point[0]);
  const ys = points.map((point) => point[1]);
  return { minX: Math.min(...xs), maxX: Math.max(...xs), minY: Math.min(...ys), maxY: Math.max(...ys) };
}

function injectUweScreenshots(svg, graph) {
  let output = svg;
  for (const node of graph.nodes) {
    const token = 'screenshot:' + escapeHtml(node.id);
    const start = output.indexOf(token);
    if (start < 0) continue;
    const groupStart = output.lastIndexOf('<g ', start);
    const groupEnd = output.indexOf('</g>', start);
    if (groupStart < 0 || groupEnd < 0) continue;
    const group = output.slice(groupStart, groupEnd + 4);
    if (!group.includes('class="node"')) continue;
    const dataUrl = imageDataUrl(node.screenshot);
    if (!dataUrl) continue;
    const tokenStart = group.indexOf(token);
    const beforeToken = group.slice(0, tokenStart);
    const polygonStart = beforeToken.lastIndexOf('<polygon ');
    const polygonEnd = polygonStart >= 0 ? group.indexOf('>', polygonStart) : -1;
    if (polygonStart < 0 || polygonEnd < 0) continue;
    const bounds = polygonBounds(group.slice(polygonStart, polygonEnd + 1));
    if (!bounds) continue;
    const imageX = bounds.minX + 6;
    const imageY = bounds.minY + 6;
    const imageWidth = bounds.maxX - bounds.minX - 12;
    const imageHeight = bounds.maxY - bounds.minY - 12;
    const tokenTextStart = group.lastIndexOf('<text ', tokenStart);
    const tokenTextEnd = tokenTextStart >= 0 ? group.indexOf('</text>', tokenStart) + '</text>'.length : -1;
    if (tokenTextStart < 0 || tokenTextEnd < 0) continue;
    const replacement = group.slice(0, tokenTextStart) +
      '<image href="' + escapeAttribute(dataUrl) + '" x="' + imageX.toFixed(1) + '" y="' + imageY.toFixed(1) + '" width="' + imageWidth.toFixed(1) + '" height="' + imageHeight.toFixed(1) + '" preserveAspectRatio="xMidYMid meet"></image>' +
      group.slice(tokenTextEnd);
    output = output.slice(0, groupStart) + replacement + output.slice(groupEnd + 4);
  }
  return output.replace(/<polygon fill="none" stroke="black"/g, '<polygon fill="#ffffff" stroke="#111827"')
    .replace(/<polygon fill="none" stroke="#111827"/g, '<polygon fill="#ffffff" stroke="#111827"');
}

async function renderUweNavigationGraphvizSvg(spec, graph) {
  try {
    return injectUweScreenshots(await graphvizSvg(uweGraphvizDot(spec, graph)), graph);
  } catch (error) {
    return renderUweNavigationSvg(spec, graph) + '<p class="muted">Graphviz render failed; fallback renderer used: ' + escapeHtml(error.message) + '</p>';
  }
}

function renderZoomableUmlSvg(visual, index) {
  const name = 'uwe-zoom-' + index;
  return '<div class="uml-zoom" data-svg-pan-zoom="true" role="group" aria-label="Zoomable UML model rendered by open-source Graphviz via @viz-js/viz and viewed with svg-pan-zoom">' +
    '<input id="' + name + '-fit" name="' + name + '" type="radio" checked><input id="' + name + '-100" name="' + name + '" type="radio"><input id="' + name + '-150" name="' + name + '" type="radio"><input id="' + name + '-200" name="' + name + '" type="radio">' +
    '<div class="uml-zoom-controls" aria-label="CSS fallback zoom controls"><span>Zoom</span><label for="' + name + '-fit">Fit</label><label for="' + name + '-100">100%</label><label for="' + name + '-150">150%</label><label for="' + name + '-200">200%</label><span class="viewer-badge">svg-pan-zoom enabled when reviewed JS lane is allowed</span></div>' +
    '<div class="uml-zoom-frame"><div class="uml-zoom-canvas">' + visual + '</div></div>' +
    '</div>';
}

function uweStereotype(node) {
  if (node.stereotype) return node.stereotype;
  const type = String(node.type || node.facet || '').toLowerCase();
  if (type.includes('process')) return '«processClass»';
  if (type.includes('menu')) return '«menu»';
  if (type.includes('query')) return '«query»';
  if (type.includes('index')) return '«index»';
  if (type.includes('external')) return '«externalNode»';
  if (type.includes('adaptation')) return '«navigationClass» {adaptation}';
  if (type.includes('access')) return '«accessPrimitive»';
  return '«navigationClass»';
}

function renderUweWorkspace(spec, graph, index) {
  const workspaceId = 'uwe-workspace-' + index;
  const packages = [...new Set(graph.nodes.map((node) => node.navigationClass || 'Navigation'))];
  const first = graph.nodes[0] ?? {};
  const packageButtons = packages.map((name) => '<button type="button" data-uwe-action="package:' + escapeAttribute(name) + '">' + escapeHtml(name.slice(0, 28)) + '</button>').join('');
  const nodeButtons = graph.nodes.map((node) => '<button type="button" data-uwe-focus-node="' + escapeAttribute(node.id) + '">' + escapeHtml((node.label || node.id).slice(0, 24)) + '</button>').join('');
  const nodeData = graph.nodes.map((node) => {
    const dataUrl = imageDataUrl(node.screenshot) || '';
    return '<span data-uwe-node data-uwe-id="' + escapeAttribute(node.id) + '" data-uwe-label="' + escapeAttribute(node.label || node.id) + '" data-uwe-stereo="' + escapeAttribute(uweStereotype(node)) + '" data-uwe-type="' + escapeAttribute(node.type || node.facet || 'navigation') + '" data-uwe-package="' + escapeAttribute(node.navigationClass || 'Navigation') + '" data-uwe-route="' + escapeAttribute(node.route || '') + '" data-uwe-role="' + escapeAttribute(node.role || '') + '" data-uwe-actions="' + escapeAttribute(node.actions || 'Inspect this node and outgoing UWE links.') + '" data-uwe-effect="' + escapeAttribute(node.effect || '') + '" data-uwe-screenshot="' + escapeAttribute(dataUrl) + '"></span>';
  }).join('');
  const edgeData = graph.edges.map((edge) => '<span data-uwe-edge data-uwe-from="' + escapeAttribute(edge.from) + '" data-uwe-to="' + escapeAttribute(edge.to) + '" data-uwe-label="' + escapeAttribute(edge.label || '«navigationLink»') + '"></span>').join('');
  const firstImage = imageDataUrl(first.screenshot) || 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==';
  return '<style>.uwe-engine-workspace{border:1px solid #111827;background:#fff;margin:12px 0 18px}.uwe-engine-head{display:flex;flex-wrap:wrap;align-items:flex-start;justify-content:space-between;gap:12px;border-bottom:1px solid #111827;background:#f3f4f6;padding:10px 12px}.uwe-engine-head h4{margin:0;font-size:15px;line-height:1.2}.uwe-engine-toolbar,.uwe-node-map{display:flex;flex-wrap:wrap;gap:7px;align-items:center;padding:9px 12px;border-bottom:1px solid #d8dee8}.uwe-node-map{background:#fbfdff}.uwe-engine-toolbar button,.uwe-node-map button{cursor:pointer;border:1px solid #111827;background:#fff;color:#111827;padding:5px 9px;font-size:12px;font-weight:700}.uwe-engine-toolbar button:hover,.uwe-node-map button:hover{background:#111827;color:#fff}.uwe-runtime-badge{margin-left:auto;color:#374151;font-size:12px;font-weight:800}.uwe-engine-grid{display:grid;grid-template-columns:minmax(0,1fr) 300px;min-height:620px}.uwe-cy-graph{min-height:620px;background:linear-gradient(#f8fafc,#fff);position:relative}.uwe-cy-placeholder{position:absolute;inset:0;display:grid;place-items:center;color:#64748b;font-weight:800}.uwe-engine-inspector{border-left:1px solid #111827;background:#fbfdff;padding:12px}.uwe-engine-inspector img{display:block;width:100%;aspect-ratio:16/10;object-fit:cover;border:1px solid #111827;background:#fff;margin:8px 0}.uwe-inspector-kicker{display:block;color:#0f766e;font-size:12px;font-weight:900;text-transform:uppercase}.uwe-inspector-title{margin:2px 0 0;font-size:18px}.uwe-inspector-block{border-top:1px solid #d8dee8;padding:8px 0}.uwe-inspector-block strong{display:block;font-size:11px;text-transform:uppercase;color:#5b6472}.uwe-engine-stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px;padding:9px 12px;border-top:1px solid #d8dee8}.uwe-engine-stats div{border:1px solid #d8dee8;background:#f8fafc;padding:8px}.uwe-engine-stats strong{display:block;font-size:20px}@media(max-width:920px){.uwe-engine-grid{grid-template-columns:1fr}.uwe-engine-inspector{border-left:0;border-top:1px solid #111827}.uwe-runtime-badge{margin-left:0}}</style>' +
    '<div id="' + workspaceId + '" class="uwe-engine-workspace" role="group" aria-label="Engine-backed UWE navigation workspace">' +
    '<div class="uwe-engine-head"><div><h4>Engine-Backed UWE Navigation Workspace</h4><p class="muted">Primary view rendered from structured source data with Cytoscape.js and dagre. Screenshots are node backgrounds; click nodes to inspect actions and effects.</p></div><span class="tool-badge">Cytoscape + dagre</span></div>' +
    '<div class="uwe-engine-toolbar"><button type="button" data-uwe-action="fit">Fit graph</button><button type="button" data-uwe-action="layout">Re-run layout</button>' + packageButtons + '<span class="uwe-runtime-badge" data-uwe-runtime-badge>Workspace runtime pending</span></div><div class="uwe-node-map" aria-label="UWE node focus map">' + nodeButtons + '</div>' +
    '<div class="uwe-engine-grid"><div class="uwe-cy-graph" data-uwe-cy><div class="uwe-cy-placeholder">Rendering UWE graph workspace...</div></div>' +
    '<aside class="uwe-engine-inspector" aria-label="Selected UWE node inspector"><span class="uwe-inspector-kicker" data-uwe-inspector-stereo>' + escapeHtml(uweStereotype(first)) + '</span><h4 class="uwe-inspector-title" data-uwe-inspector-title>' + escapeHtml(first.label || first.id || 'UWE node') + '</h4><img data-uwe-inspector-image src="' + escapeAttribute(firstImage) + '" alt="Selected UWE node screenshot"><div class="uwe-inspector-block"><strong>Package</strong><span data-uwe-inspector-package>' + escapeHtml(first.navigationClass || 'Navigation') + '</span></div><div class="uwe-inspector-block"><strong>Route or state</strong><span data-uwe-inspector-route>' + escapeHtml(first.route || 'state') + '</span></div><div class="uwe-inspector-block"><strong>Role</strong><span data-uwe-inspector-role>' + escapeHtml(first.role || 'all roles') + '</span></div><div class="uwe-inspector-block"><strong>Available user action</strong><span data-uwe-inspector-actions>' + escapeHtml(first.actions || 'Inspect this node and outgoing UWE links.') + '</span></div><div class="uwe-inspector-block"><strong>System effect</strong><span data-uwe-inspector-effect>' + escapeHtml(first.effect || 'Effect not recorded.') + '</span></div></aside></div>' +
    '<div class="uwe-engine-stats"><div><strong data-uwe-stat="nodes">' + graph.nodes.length + '</strong>UWE nodes</div><div><strong data-uwe-stat="edges">' + graph.edges.length + '</strong>typed links</div><div><strong data-uwe-stat="packages">' + packages.length + '</strong>packages</div></div>' +
    '<div hidden>' + nodeData + edgeData + '</div></div>';
}

function renderUweNavigationSvg(spec, graph) {
  const cardWidth = 210;
  const cardHeight = 168;
  const gapX = 42;
  const gapY = 34;
  const laneGap = 32;
  const lanePadding = 18;
  const explicitClasses = Array.isArray(spec.navigationClasses ?? spec.lanes ?? spec.classes)
    ? (spec.navigationClasses ?? spec.lanes ?? spec.classes).map((item) => typeof item === 'string' ? item : (item.label ?? item.id ?? item.name)).filter(Boolean).map(String)
    : [];
  const discoveredClasses = graph.nodes.map((node) => node.navigationClass || 'Navigation').filter((value, index, all) => all.indexOf(value) === index);
  const classNames = [...explicitClasses, ...discoveredClasses.filter((name) => !explicitClasses.includes(name))];
  const lanes = classNames.map((name) => {
    const nodes = graph.nodes.filter((node) => (node.navigationClass || 'Navigation') === name);
    return { name, nodes: nodes.length > 0 ? nodes : [] };
  }).filter((lane) => lane.nodes.length > 0);
  const cols = Math.min(4, Math.max(1, ...lanes.map((lane) => lane.nodes.length)));
  const laneWidth = cols * cardWidth + (cols - 1) * gapX + lanePadding * 2;
  const positions = new Map();
  let yCursor = 50;
  for (const lane of lanes) {
    lane.y = yCursor;
    lane.rows = Math.max(1, Math.ceil(lane.nodes.length / cols));
    lane.height = lane.rows * cardHeight + (lane.rows - 1) * gapY + lanePadding * 2 + 34;
    lane.nodes.forEach((node, index) => {
      const col = index % cols;
      const row = Math.floor(index / cols);
      positions.set(node.id, {
        x: lanePadding + col * (cardWidth + gapX),
        y: lane.y + lanePadding + 34 + row * (cardHeight + gapY),
      });
    });
    yCursor += lane.height + laneGap;
  }
  const width = laneWidth;
  const height = Math.max(260, yCursor + 4);
  const laneRects = lanes.map((lane) =>
    '<g><rect x="4" y="' + lane.y + '" width="' + (laneWidth - 8) + '" height="' + lane.height + '" rx="8" fill="#f8fafc" stroke="#cbd5e1"></rect>' +
    '<text x="20" y="' + (lane.y + 24) + '" font-size="13" font-weight="800" fill="#334155">«navigation class» ' + escapeHtml(lane.name.slice(0, 80)) + '</text></g>'
  ).join('');
  const edges = graph.edges.map((edge) => {
    const from = positions.get(edge.from);
    const to = positions.get(edge.to);
    if (!from || !to) return '';
    const x1 = from.x + cardWidth;
    const y1 = from.y + cardHeight / 2;
    const x2 = to.x;
    const y2 = to.y + cardHeight / 2;
    const labelX = (x1 + x2) / 2;
    const labelY = (y1 + y2) / 2 - 5;
    return '<g><line x1="' + x1.toFixed(1) + '" y1="' + y1.toFixed(1) + '" x2="' + x2.toFixed(1) + '" y2="' + y2.toFixed(1) + '" stroke="#64748b" stroke-width="2" marker-end="url(#arrowHead)"></line>' +
      (edge.label ? '<text x="' + labelX.toFixed(1) + '" y="' + labelY.toFixed(1) + '" text-anchor="middle" font-size="10" fill="#334155">' + escapeHtml(edge.label.slice(0, 34)) + '</text>' : '') + '</g>';
  }).join('');
  const nodes = graph.nodes.map((node) => {
    const pos = positions.get(node.id);
    const dataUrl = imageDataUrl(node.screenshot);
    const image = dataUrl
      ? '<image href="' + escapeAttribute(dataUrl) + '" x="' + (pos.x + 10) + '" y="' + (pos.y + 34) + '" width="' + (cardWidth - 20) + '" height="96" preserveAspectRatio="xMidYMid meet"></image>'
      : '<rect x="' + (pos.x + 10) + '" y="' + (pos.y + 34) + '" width="' + (cardWidth - 20) + '" height="96" rx="6" fill="#eef2f6" stroke="#d8dee8"></rect><text x="' + (pos.x + cardWidth / 2) + '" y="' + (pos.y + 86) + '" text-anchor="middle" font-size="11" fill="#64748b">no screenshot</text>';
    return '<g><rect x="' + pos.x + '" y="' + pos.y + '" width="' + cardWidth + '" height="' + cardHeight + '" rx="8" fill="#ffffff" stroke="#9fb3c8" stroke-width="1.5"></rect>' +
      '<rect x="' + pos.x + '" y="' + pos.y + '" width="' + cardWidth + '" height="25" rx="8" fill="#effaf8"></rect>' +
      '<text x="' + (pos.x + 10) + '" y="' + (pos.y + 17) + '" font-size="11" font-weight="800" fill="#0f766e">«navigation node»</text>' +
      image +
      '<text x="' + (pos.x + 10) + '" y="' + (pos.y + 146) + '" font-size="12" font-weight="800" fill="#111827">' + escapeHtml(node.label.slice(0, 26)) + '</text>' +
      '<text x="' + (pos.x + 10) + '" y="' + (pos.y + 160) + '" font-size="10" fill="#334155">' + escapeHtml((node.route || node.facet || node.role || 'navigationNode').slice(0, 34)) + '</text></g>';
  }).join('');
  return '<svg class="infographic-chart" viewBox="0 0 ' + width + ' ' + height + '" role="img" aria-label="' + escapeAttribute(spec.title ?? 'UWE navigation screenshot model') + '"><defs><marker id="arrowHead" markerWidth="8" markerHeight="8" refX="7" refY="3.5" orient="auto"><path d="M0,0 L8,3.5 L0,7 Z" fill="#64748b"></path></marker></defs>' + laneRects + edges + nodes + '</svg>';
}

async function renderInfographicSpec(spec, index) {
  const tool = normalizeToolId(spec.tool ?? spec.renderer ?? 'source-spec');
  const allowed = infographicToolIds.has(tool) || tool === 'source-spec';
  const kind = String(spec.kind ?? spec.mark ?? spec.type ?? '').toLowerCase();
  const isUwe = kind === 'uwe-navigation' || kind === 'uwe';
  let visual = '';
  if (isUwe) {
    const graph = graphData(spec);
    const graphvizVisual = await renderUweNavigationGraphvizSvg(spec, graph);
    visual = renderUweWorkspace(spec, graph, index) +
      '<details class="uml-render-fallback"><summary>Graphviz UML fallback and source-rendered SVG</summary>' + renderZoomableUmlSvg(graphvizVisual, index) + '</details>';
  } else if (['graphviz', 'mermaid'].includes(tool) || ['graph', 'network', 'lineage', 'relationship'].includes(kind)) {
    visual = await renderGraphSvg(spec);
  } else if (kind === 'line' || kind === 'trend') {
    visual = renderLineSvg(spec);
  } else {
    visual = renderBarSvg(spec);
  }
  const articleClass = isUwe ? 'chart-panel uml-model-panel' : 'chart-panel';
  const badge = isUwe ? 'UWE workspace - Cytoscape + Graphviz' : (tool || 'source-spec');
  const rendererNote = isUwe ? '<p class="muted">Generated from source JSON into an engine-backed Cytoscape/Dagre workspace, with Graphviz kept as the UML renderer fallback. Screenshots remain evidence extensions inside UWE nodes, not replacements for UWE notation.</p>' : '';
  return '<article class="' + articleClass + '"><div class="chart-head"><h3>' + escapeHtml(spec.title ?? 'Infographic ' + (index + 1)) + '</h3><span class="tool-badge">' + escapeHtml(badge) + '</span></div><p>' + escapeHtml(spec.summary ?? spec.description ?? (allowed ? 'Rendered as static review markup from a source-declared infographic spec.' : 'Unknown tool requested; rendered with the static fallback.')) + '</p>' + rendererNote + visual + '</article>';
}

function renderInfographicToolkit() {
  return '<section class="panel"><h2>Open-Source Infographic Toolkit</h2><p class="muted">These tools are allowed as source/spec or generation-time renderers. The generated HTML does not load their browser runtimes.</p><div class="toolkit-grid">' + infographicTools.map((tool) => '<div class="tool-card"><strong>' + escapeHtml(tool.label) + '</strong><span>' + escapeHtml(tool.role) + '</span><em>' + escapeHtml(tool.output) + '</em></div>').join('') + '</div></section>';
}

async function renderInfographicSpecs(source, artifact) {
  const specs = parseInfographicSpecs(source, artifact);
  if (specs.length === 0) return '';
  const hasUwe = specs.some((spec) => ['uwe-navigation', 'uwe'].includes(String(spec.kind ?? spec.mark ?? spec.type ?? '').toLowerCase()));
  return '<section class="panel ' + (hasUwe ? 'uml-review-section' : '') + '"><h2>' + (hasUwe ? 'UWE Navigation Model' : 'Static Infographic Specs') + '</h2><div class="chart-grid">' + (await Promise.all(specs.map(renderInfographicSpec))).join('') + '</div></section>';
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
  if (figures.length === 0) return '';
  return '<section class="panel"><h2>Screenshots And Evidence Images</h2><div class="gallery">' + figures.join('\n') + '</div></section>';
}

function htmlPage(title, body, options = {}) {
  const svgPanZoomInitializer = 'document.querySelectorAll("[data-svg-pan-zoom=true] svg").forEach(function(svg){svgPanZoom(svg,{controlIconsEnabled:true,fit:true,center:true,minZoom:.1,maxZoom:20,zoomScaleSensitivity:.25});});document.querySelectorAll(".viewer-badge").forEach(function(el){el.textContent="svg-pan-zoom active: drag, wheel, +/- controls";});document.documentElement.classList.add("svg-pan-zoom-active");';
  const scripts = options.enableUweWorkspace && svgPanZoomRuntime && cytoscapeRuntime && dagreRuntime && cytoscapeDagreRuntime && uweWorkspaceRuntime
    ? '<script>' + svgPanZoomRuntime + '</script>\n<script>' + cytoscapeRuntime + '</script>\n<script>' + dagreRuntime + '</script>\n<script>' + cytoscapeDagreRuntime + '</script>\n<script>' + uweWorkspaceRuntime + '</script>\n'
    : options.enableSvgPanZoom && svgPanZoomRuntime
    ? '<script>' + svgPanZoomRuntime + '</script>\n<script>' + svgPanZoomInitializer + '</script>\n'
    : '';
  const csp = options.enableSvgPanZoom || options.enableUweWorkspace ? "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'" : requiredCsp;
  return '<!doctype html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n<meta name="viewport" content="width=device-width, initial-scale=1">\n<meta http-equiv="Content-Security-Policy" content="' + escapeAttribute(csp) + '">\n<title>' + escapeHtml(title) + '</title>\n<style>\n:root{color-scheme:light;font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;line-height:1.55;--bg:#eef2f6;--panel:#fff;--text:#1f2937;--muted:#5b6472;--line:#d8dee8;--navy:#172033;--teal:#0f766e;--blue:#2457c5;--amber:#a15c07;--green:#16794b;--red:#b42318;--violet:#6546a3}*{box-sizing:border-box}body{margin:0;color:var(--text);background:var(--bg)}a{color:#174ea6}header{background:var(--navy);color:#fff;padding:28px;border-bottom:6px solid var(--teal)}header h1{margin:0 0 8px;font-size:clamp(28px,4vw,42px);line-height:1.08;letter-spacing:0}header p{max-width:1040px;margin:0;color:#dce6f1}main{max-width:1240px;margin:0 auto;padding:18px}h2,h3{margin:0 0 10px;line-height:1.2;letter-spacing:0}p{margin:0 0 12px}.grid{display:grid;grid-template-columns:minmax(0,1.1fr) minmax(280px,.9fr);gap:16px}.panel{background:var(--panel);border:1px solid var(--line);border-radius:8px;padding:18px;margin-bottom:16px}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px}.metric{border:1px solid var(--line);border-top:5px solid var(--teal);border-radius:8px;padding:12px;background:#fff;min-height:104px}.metric strong{display:block;font-size:28px;line-height:1;margin-bottom:6px}.metric span{color:var(--muted);font-size:13px}.blue{border-top-color:var(--blue)}.amber{border-top-color:var(--amber)}.green{border-top-color:var(--green)}.violet{border-top-color:var(--violet)}.tabs>input{position:absolute;inline-size:1px;block-size:1px;overflow:hidden;clip:rect(0 0 0 0)}.tab-labels{display:flex;flex-wrap:wrap;gap:8px;border-bottom:1px solid var(--line);padding-bottom:10px}.tab-labels label{cursor:pointer;padding:8px 11px;border:1px solid var(--line);border-radius:7px;background:#fff;font-weight:650}.tab-panel{display:none;margin-top:14px}.tabs input:nth-of-type(1):checked~.tab-panels .tab-panel:nth-of-type(1),.tabs input:nth-of-type(2):checked~.tab-panels .tab-panel:nth-of-type(2),.tabs input:nth-of-type(3):checked~.tab-panels .tab-panel:nth-of-type(3),.tabs input:nth-of-type(4):checked~.tab-panels .tab-panel:nth-of-type(4){display:block}.flow{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:8px}.step{border:1px solid var(--line);border-radius:8px;background:#f9fbfd;padding:12px;min-height:96px}.step strong{display:block;margin-bottom:5px}.step span,.muted{color:var(--muted)}.bar-row{display:grid;grid-template-columns:172px minmax(0,1fr) 52px;gap:10px;align-items:center;margin:10px 0}.bar-track{height:18px;background:#e8edf4;border-radius:999px;overflow:hidden}.bar{display:block;height:100%;border-radius:999px;background:var(--teal)}.w100{width:100%}.w80{width:80%}.w60{width:60%}.w40{width:40%}.w20{width:20%}.toolkit-grid,.chart-grid,.gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px}.uml-review-section{background:#fdfdfb;border-color:#111827}.uml-review-section>h2{font-size:18px;text-transform:uppercase;letter-spacing:.04em;color:#111827}.uml-model-panel{grid-column:1/-1;border:1px solid #111827;border-radius:0;background:#fff;padding:14px}.uml-model-panel .chart-head{border-bottom:1px solid #111827;margin:-14px -14px 12px;padding:10px 12px;background:#f3f4f6}.uml-model-panel .tool-badge{border-color:#111827;border-radius:0;background:#fff;color:#111827}.uml-model-panel p{max-width:980px;color:#374151}.uml-model-panel .graphviz-render{border:0;border-radius:0;background:#fff;padding:0;margin:0}.uml-zoom{border:1px solid #111827;background:#fff;margin-top:10px}.uml-zoom>input{position:absolute;inline-size:1px;block-size:1px;overflow:hidden;clip:rect(0 0 0 0)}.uml-zoom-controls{display:flex;flex-wrap:wrap;gap:6px;align-items:center;border-bottom:1px solid #111827;background:#f3f4f6;padding:8px 10px}.uml-zoom-controls span{font-size:12px;font-weight:800;text-transform:uppercase;color:#111827}.uml-zoom-controls label{cursor:pointer;border:1px solid #111827;background:#fff;color:#111827;padding:4px 8px;font-size:12px;font-weight:700}.viewer-badge{text-transform:none!important;font-weight:700!important;color:#374151!important;margin-left:auto}.uml-zoom-frame{height:min(72vh,760px);overflow:auto;background:#fff}.uml-zoom-canvas{transform-origin:0 0;transition:transform .12s ease;width:max-content;min-width:100%}.uml-zoom input:nth-of-type(1):checked~.uml-zoom-controls label:nth-of-type(1),.uml-zoom input:nth-of-type(2):checked~.uml-zoom-controls label:nth-of-type(2),.uml-zoom input:nth-of-type(3):checked~.uml-zoom-controls label:nth-of-type(3),.uml-zoom input:nth-of-type(4):checked~.uml-zoom-controls label:nth-of-type(4){background:#111827;color:#fff}.uml-zoom input:nth-of-type(2):checked~.uml-zoom-frame .uml-zoom-canvas{transform:scale(1)}.uml-zoom input:nth-of-type(3):checked~.uml-zoom-frame .uml-zoom-canvas{transform:scale(1.5);padding-right:50%;padding-bottom:28%}.uml-zoom input:nth-of-type(4):checked~.uml-zoom-frame .uml-zoom-canvas{transform:scale(2);padding-right:100%;padding-bottom:55%}.tool-card,.chart-panel{border:1px solid var(--line);border-radius:8px;background:#fff;padding:12px}.chart-panel.uml-model-panel{border:1px solid #111827;border-radius:0;background:#fff;padding:14px}.tool-card strong,.tool-card span,.tool-card em{display:block}.tool-card span{color:var(--muted);font-size:13px}.tool-card em{margin-top:6px;color:#334155;font-size:12px;font-style:normal}.chart-head{display:flex;gap:8px;align-items:flex-start;justify-content:space-between}.tool-badge{display:inline-flex;border:1px solid #b9c7dc;background:#f4f7fb;border-radius:999px;padding:2px 8px;color:#334155;font-size:12px;font-weight:700;white-space:nowrap}.infographic-chart{width:100%;height:auto;border:1px solid var(--line);border-radius:8px;background:#fff;margin-top:8px}.graphviz-render .edge polygon{fill:#fff;stroke:#111827}.graphviz-render text{font-family:"Segoe UI",Arial,sans-serif}.graphviz-render .cluster text{font-weight:600}.graphviz-render .node>polygon:last-child{stroke:#111827}figure{margin:0;border:1px solid var(--line);border-radius:8px;background:#fff;overflow:hidden}figure img{display:block;width:100%;height:auto}figcaption{padding:9px 10px;color:var(--muted);font-size:13px}table{width:100%;border-collapse:collapse;margin-top:8px}th,td{border-bottom:1px solid var(--line);text-align:left;vertical-align:top;padding:9px}th{background:#f7f9fc}pre{white-space:pre-wrap;overflow:auto;background:#101828;color:#e5edf7;padding:14px;border-radius:8px}code{font-family:ui-monospace,SFMono-Regular,Consolas,monospace;background:#eef2f6;border:1px solid #dae2ec;border-radius:4px;padding:1px 4px}ul,ol{margin:8px 0 0 20px;padding:0}li{margin:4px 0}.callout{border-left:5px solid var(--teal);background:#effaf8;padding:12px 14px;border-radius:7px;margin:12px 0}@media(max-width:920px){.grid,.metrics{grid-template-columns:1fr}main{padding:12px}.bar-row{grid-template-columns:1fr}.chart-head{display:block}.tool-badge{margin-bottom:8px}.viewer-badge{margin-left:0}}\\n</style>\\n</head>\\n<body>\\n' + body + '\\n' + scripts + '</body>\\n</html>\\n';
}

async function renderArtifact(artifact, outPath) {
  const source = readSource(artifact);
  const summary = artifact.summary || artifact.purpose || firstParagraph(source) || 'Source-backed human review artifact.';
  const family = familyFor(artifact);
  const stats = sourceStats(source);
  const title = artifact.title || sourceTitle(source) || artifact.id;
  const sectionHeads = headings(source);
  const evidenceCount = Array.isArray(artifact.evidenceLinks) ? artifact.evidenceLinks.length : 0;
  const updateCount = Array.isArray(artifact.updateTriggers) ? artifact.updateTriggers.length : 0;
  const body = '<header><h1>' + escapeHtml(title) + '</h1><p>' + escapeHtml(summary) + '</p></header><main>' +
    '<section class="grid"><div class="panel"><h2>Review Verdict</h2><p>' + escapeHtml(summary) + '</p><div class="callout"><strong>Source first:</strong> edit ' + linkFor(outPath, artifact.source, artifact.source) + ' before regenerating this review surface.</div></div>' +
    '<div class="metrics"><div class="metric green"><strong>' + escapeHtml(artifact.status || 'draft') + '</strong><span>Status</span></div><div class="metric blue"><strong>' + evidenceCount + '</strong><span>Evidence links</span></div><div class="metric amber"><strong>' + stats.sectionCount + '</strong><span>Major sections</span></div><div class="metric violet"><strong>' + escapeHtml(family) + '</strong><span>Artifact family</span></div></div></section>' +
    '<section class="panel"><h2>Infographic Snapshot</h2><div class="bar-row"><span>Evidence coverage</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, evidenceCount * 20)) + '"></span></span><strong>' + evidenceCount + '</strong></div><div class="bar-row"><span>Source depth</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, stats.sectionCount * 20)) + '"></span></span><strong>' + stats.sectionCount + '</strong></div><div class="bar-row"><span>Update triggers</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, updateCount * 20)) + '"></span></span><strong>' + updateCount + '</strong></div></section>' +
    renderInfographicToolkit() +
    await renderInfographicSpecs(source, artifact) +
    gallerySection(artifact) +
    '<section class="panel"><h2>Source-To-Review Flow</h2><div class="flow"><div class="step"><strong>Canonical Source</strong><span>' + escapeHtml(artifact.source || '') + '</span></div><div class="step"><strong>Generated HTML</strong><span>' + escapeHtml(artifact.reviewSurface || '') + '</span></div><div class="step"><strong>Evidence</strong><span>' + evidenceCount + ' linked item(s)</span></div><div class="step"><strong>Freshness</strong><span>' + escapeHtml(artifact.generatedAt || artifact.freshness?.generatedAt || 'not-recorded') + '</span></div></div></section>' +
    '<section class="panel tabs"><input id="tab-overview" name="tabs" type="radio" checked><input id="tab-evidence" name="tabs" type="radio"><input id="tab-source" name="tabs" type="radio"><input id="tab-metadata" name="tabs" type="radio"><div class="tab-labels"><label for="tab-overview">Overview</label><label for="tab-evidence">Evidence</label><label for="tab-source">Source</label><label for="tab-metadata">Metadata</label></div><div class="tab-panels">' +
    '<div class="tab-panel"><h2>Review Sections</h2>' + (sectionHeads.length === 0 ? '<p class="muted">No headings found in source.</p>' : '<ol>' + sectionHeads.map((heading) => '<li>' + escapeHtml(heading.text) + '</li>').join('') + '</ol>') + '</div>' +
    '<div class="tab-panel"><h2>Evidence</h2>' + listItems(artifact.evidenceLinks, 'No evidence links are listed yet.', outPath) + '</div>' +
    '<div class="tab-panel"><h2>Canonical Source</h2><p>' + linkFor(outPath, artifact.source, artifact.source || 'source') + '</p><pre>' + escapeHtml(source || 'Source not found or not readable.') + '</pre></div>' +
    '<div class="tab-panel"><h2>Metadata</h2><table><tbody><tr><th>ID</th><td>' + escapeHtml(artifact.id) + '</td></tr><tr><th>Type</th><td>' + escapeHtml(artifact.type) + '</td></tr><tr><th>Owner</th><td>' + escapeHtml(artifact.owner) + '</td></tr><tr><th>Renderer</th><td>' + escapeHtml(artifact.renderer || rendererName) + '</td></tr><tr><th>Source hash</th><td>' + escapeHtml(artifact.sourceHash || '') + '</td></tr></tbody></table></div>' +
    '</div></section></main>';
  return htmlPage(String(title || 'Artifact Review'), body, {
    enableSvgPanZoom: artifact.htmlInteractionLane === 'reviewed-svg-pan-zoom',
    enableUweWorkspace: artifact.htmlInteractionLane === 'reviewed-uwe-workspace'
  });
}

function renderIndex(artifacts, outPath, hasModelArtifacts) {
  const rows = artifacts.map((artifact) => '<tr><td>' + escapeHtml(artifact.id) + '</td><td>' + escapeHtml(artifact.type) + '</td><td>' + escapeHtml(artifact.status) + '</td><td>' + linkFor(outPath, artifact.source, artifact.source) + '</td><td>' + linkFor(outPath, artifact.reviewSurface, artifact.reviewSurface) + '</td></tr>').join('\n');
  const modelLink = hasModelArtifacts ? '<section class="panel"><h2>Model Reviews</h2><p>' + linkFor(outPath, 'generated/review/models/index.html', 'Open the generated model review index') + '</p></section>' : '';
  return htmlPage('Artifact Review Index', '<header><h1>Artifact Review Index</h1><p>Static infographic review surfaces generated from canonical source artifacts.</p></header><main>' + modelLink + '<section class="panel"><table><thead><tr><th>ID</th><th>Type</th><th>Status</th><th>Source</th><th>HTML Review</th></tr></thead><tbody>' + rows + '</tbody></table></section></main>');
}

if (!fs.existsSync(manifestPath)) {
  console.error('Missing artifact manifest: ' + repoPath(manifestPath));
  process.exit(1);
}

fs.mkdirSync(reviewRoot, { recursive: true });
const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
const artifacts = Array.isArray(manifest.artifacts) ? manifest.artifacts.filter(isManagedArtifact) : [];
const hasModelArtifacts = Array.isArray(manifest.artifacts) && manifest.artifacts.some(isModelArtifact);
const expectedFiles = new Map();
const generationDate = new Date().toISOString().slice(0, 10);

for (const artifact of artifacts) {
  artifact.reviewSurface = artifact.reviewSurface || defaultReviewSurface(artifact);
  artifact.renderer = rendererName;
  artifact.sourceHash = hashSource(artifact) || artifact.sourceHash;
  artifact.generatedAt = checkOnly ? (artifact.generatedAt || artifact.freshness?.generatedAt || generationDate) : (sourceGeneratedAt(artifact) || generationDate);
  artifact.freshness = { ...(artifact.freshness ?? {}), generatedAt: artifact.generatedAt, sourceFirst: true };
  const outPath = resolveReviewSurface(artifact);
  expectedFiles.set(outPath, await renderArtifact(artifact, outPath));
}

if (artifacts.length > 0 || hasModelArtifacts) {
  const indexPath = path.join(reviewRoot, 'index.html');
  expectedFiles.set(indexPath, renderIndex(artifacts, indexPath, hasModelArtifacts));
}

if (checkOnly) {
  const failures = [];
  for (const [filePath, expected] of expectedFiles) {
    if (!fs.existsSync(filePath)) {
      failures.push('missing generated artifact review: ' + repoPath(filePath));
      continue;
    }
    if (fs.readFileSync(filePath, 'utf8') !== expected) failures.push('stale generated artifact review: ' + repoPath(filePath));
  }
  const currentManifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  if (JSON.stringify(currentManifest, null, 2) + '\n' !== JSON.stringify(manifest, null, 2) + '\n') failures.push('manifest artifact review metadata is stale; run node scripts/generate-artifact-review.mjs');
  if (failures.length > 0) {
    console.error('Artifact review drift check failed:');
    for (const failure of failures) console.error('- ' + failure);
    process.exit(1);
  }
  console.log('Artifact review drift check passed');
} else {
  for (const [filePath, html] of expectedFiles) {
    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    fs.writeFileSync(filePath, html);
  }
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n');
  console.log('Generated ' + artifacts.length + ' artifact review surface(s) in ' + repoPath(reviewRoot));
}
