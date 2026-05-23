package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveProjectSetupContextAutoUsesMonorepoRoot(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "web")
	mustMkdirAll(t, appDir)
	mustWriteFile(t, filepath.Join(root, "pnpm-workspace.yaml"), "packages:\n  - apps/*\n")
	mustWriteFile(t, filepath.Join(root, "pnpm-lock.yaml"), "lockfileVersion: '9.0'\n")
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	mustWriteFile(t, filepath.Join(appDir, "package.json"), "{\n  \"name\": \"web\",\n  \"private\": true\n}\n")

	ctx, err := resolveProjectSetupContext(appDir, "auto", "auto")
	if err != nil {
		t.Fatalf("resolveProjectSetupContext returned error: %v", err)
	}
	if ctx.OperationDir != root {
		t.Fatalf("expected operation dir %q, got %q", root, ctx.OperationDir)
	}
	if ctx.PackageManager != packageManagerPnpm {
		t.Fatalf("expected pnpm, got %q", ctx.PackageManager)
	}
}

func TestResolveProjectSetupContextWorkspaceScopeStaysLocal(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "api")
	mustMkdirAll(t, appDir)
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true,\n  \"workspaces\": [\"apps/*\"]\n}\n")
	mustWriteFile(t, filepath.Join(root, "yarn.lock"), "# lockfile\n")
	mustWriteFile(t, filepath.Join(appDir, "package.json"), "{\n  \"name\": \"api\",\n  \"private\": true\n}\n")

	ctx, err := resolveProjectSetupContext(appDir, "workspace", "auto")
	if err != nil {
		t.Fatalf("resolveProjectSetupContext returned error: %v", err)
	}
	if ctx.OperationDir != appDir {
		t.Fatalf("expected workspace operation dir %q, got %q", appDir, ctx.OperationDir)
	}
	if ctx.PackageManager != packageManagerYarn {
		t.Fatalf("expected yarn, got %q", ctx.PackageManager)
	}
}

func TestResolvePackageManagerFromPackageManagerField(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"packageManager\": \"bun@1.2.0\"\n}\n")

	manager, err := resolvePackageManager(packageManagerAuto, root)
	if err != nil {
		t.Fatalf("resolvePackageManager returned error: %v", err)
	}
	if manager != packageManagerBun {
		t.Fatalf("expected bun, got %q", manager)
	}
}

func TestParseArtifactProfileRejectsUnknownValue(t *testing.T) {
	if _, err := parseArtifactProfile("rich-ui"); err == nil {
		t.Fatal("expected unsupported artifact profile to return an error")
	}
}

func TestParseArtifactProfileAcceptsMedia(t *testing.T) {
	profile, err := parseArtifactProfile("media")
	if err != nil {
		t.Fatalf("parseArtifactProfile returned error: %v", err)
	}
	if profile != artifactProfileMedia {
		t.Fatalf("expected media profile, got %q", profile)
	}
	if effectiveArtifactProfile(profile) != artifactProfileDual {
		t.Fatalf("expected media to resolve to dual, got %q", effectiveArtifactProfile(profile))
	}
	if artifactOpenMode(profile) != "file-preview" {
		t.Fatalf("expected media open mode to use file-preview, got %q", artifactOpenMode(profile))
	}
}

func TestParseArtifactProfileAcceptsAgentLoopAliases(t *testing.T) {
	for _, value := range []string{"agent-loop", "governed-agent", "self-improving", "self-improving-agent-loop"} {
		profile, err := parseArtifactProfile(value)
		if err != nil {
			t.Fatalf("parseArtifactProfile(%q) returned error: %v", value, err)
		}
		if profile != artifactProfileAgentLoop {
			t.Fatalf("expected agent-loop profile for %q, got %q", value, profile)
		}
		if effectiveArtifactProfile(profile) != artifactProfileDual {
			t.Fatalf("expected agent-loop to resolve to dual, got %q", effectiveArtifactProfile(profile))
		}
		if artifactSpecialization(profile) != "self-improving-agent-loop" {
			t.Fatalf("expected self-improving-agent-loop specialization, got %q", artifactSpecialization(profile))
		}
		if artifactOpenMode(profile) != "file-preview" {
			t.Fatalf("expected agent-loop open mode to use file-preview, got %q", artifactOpenMode(profile))
		}
	}
}

func TestParseAndResolveModelingModes(t *testing.T) {
	for value, want := range map[string]modelingMode{
		"":             modelingModeAuto,
		"auto":         modelingModeAuto,
		"off":          modelingModeOff,
		"none":         modelingModeOff,
		"baseline":     modelingModeBaseline,
		"source-first": modelingModeBaseline,
		"uml":          modelingModeUMLFirst,
		"uml-first":    modelingModeUMLFirst,
	} {
		got, err := parseModelingMode(value)
		if err != nil {
			t.Fatalf("parseModelingMode(%q) returned error: %v", value, err)
		}
		if got != want {
			t.Fatalf("parseModelingMode(%q) = %q, want %q", value, got, want)
		}
	}

	fresh := t.TempDir()
	if got := resolveEffectiveModelingMode(fresh, modelingModeAuto, artifactProfileDual, false); got != modelingModeUMLFirst {
		t.Fatalf("fresh auto should resolve to uml-first, got %q", got)
	}

	legacy := t.TempDir()
	mustWriteFile(t, filepath.Join(legacy, ".skill-harness", "project.json"), "{\n  \"version\": 1,\n  \"capabilities\": {\"developerArtifacts\": {\"enabled\": true}}\n}\n")
	if got := resolveEffectiveModelingMode(legacy, modelingModeAuto, artifactProfileDual, false); got != modelingModeOff {
		t.Fatalf("legacy project without modeling mode should preserve off, got %q", got)
	}

	if got := resolveEffectiveModelingMode(fresh, modelingModeAuto, artifactProfileNone, false); got != modelingModeOff {
		t.Fatalf("profile none should resolve to off, got %q", got)
	}
}

func TestWriteDeveloperArtifactScaffold(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	for _, path := range []string{
		filepath.Join(root, "docs", "artifacts", "source"),
		filepath.Join(root, "docs", "artifacts", "templates"),
		filepath.Join(root, "generated", "review"),
		filepath.Join(root, ".skill-harness"),
		filepath.Join(root, "scripts"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected scaffold path %s: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", path)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	if artifacts["profile"] != string(artifactProfileDual) {
		t.Fatalf("expected dual profile, got %#v", artifacts["profile"])
	}
	if !fileExists(filepath.Join(root, "scripts", "check-artifact-html-policy.mjs")) {
		t.Fatal("expected HTML policy checker script")
	}
	if !fileExists(filepath.Join(root, "scripts", "check-artifact-manifest.mjs")) {
		t.Fatal("expected artifact manifest checker script")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "artifacts.manifest.json")) {
		t.Fatal("expected artifact manifest")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "templates", "model-artifact.md")) {
		t.Fatal("expected model artifact template")
	}
	if !gitignoreHasLine(mustReadText(t, filepath.Join(root, ".gitignore")), "generated/review/") {
		t.Fatal("expected generated review output to be gitignored")
	}
	manifest, ok := artifacts["manifest"].(map[string]any)
	if !ok {
		t.Fatalf("expected manifest config, got %#v", artifacts["manifest"])
	}
	if manifest["path"] != "docs/artifacts/artifacts.manifest.json" {
		t.Fatalf("expected manifest path, got %#v", manifest["path"])
	}
	modelPolicy, ok := artifacts["modelPolicy"].(map[string]any)
	if !ok {
		t.Fatalf("expected model policy config, got %#v", artifacts["modelPolicy"])
	}
	if modelPolicy["defaultReviewEmbedding"] != "inline-svg" {
		t.Fatalf("expected inline-svg review embedding, got %#v", modelPolicy["defaultReviewEmbedding"])
	}
	notations, ok := modelPolicy["allowedNotations"].([]any)
	if !ok || !containsJSONValue(notations, "mermaid") {
		t.Fatalf("expected mermaid notation in model policy, got %#v", modelPolicy["allowedNotations"])
	}
	scripts := packageScripts(t, filepath.Join(root, "package.json"))
	if scripts["artifacts:manifest:check"] != "node scripts/check-artifact-manifest.mjs" {
		t.Fatalf("expected artifact manifest check script, got %#v", scripts["artifacts:manifest:check"])
	}
}

func TestWriteDeveloperArtifactScaffoldMediaProfile(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileMedia, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	if !dirExists(filepath.Join(root, "generated", "media")) {
		t.Fatal("expected generated media directory")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "templates", "demo-artifact.md")) {
		t.Fatal("expected demo artifact template")
	}
	if !gitignoreHasLine(mustReadText(t, filepath.Join(root, ".gitignore")), "generated/media/") {
		t.Fatal("expected generated media output to be gitignored")
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	if artifacts["requestedProfile"] != string(artifactProfileMedia) {
		t.Fatalf("expected requested media profile, got %#v", artifacts["requestedProfile"])
	}
	if artifacts["profile"] != string(artifactProfileDual) {
		t.Fatalf("expected effective dual profile, got %#v", artifacts["profile"])
	}
	if artifacts["specialization"] != "media-demo" {
		t.Fatalf("expected media-demo specialization, got %#v", artifacts["specialization"])
	}
	mediaOutputs, ok := artifacts["mediaOutputs"].(map[string]any)
	if !ok {
		t.Fatalf("expected mediaOutputs config, got %#v", artifacts["mediaOutputs"])
	}
	if mediaOutputs["enabled"] != true {
		t.Fatalf("expected media outputs enabled, got %#v", mediaOutputs["enabled"])
	}
	if mediaOutputs["outDir"] != "generated/media" {
		t.Fatalf("expected generated/media outDir, got %#v", mediaOutputs["outDir"])
	}
	statuses, ok := mediaOutputs["allowedStatuses"].([]any)
	if !ok || !containsJSONValue(statuses, "needs-evidence") || !containsJSONValue(statuses, "inconclusive") {
		t.Fatalf("expected conservative media statuses, got %#v", mediaOutputs["allowedStatuses"])
	}
	exclusions, ok := mediaOutputs["defaultExclusions"].([]any)
	if !ok || !containsJSONValue(exclusions, "trace.zip") || !containsJSONValue(exclusions, "network.json") {
		t.Fatalf("expected sensitive default exclusions, got %#v", mediaOutputs["defaultExclusions"])
	}
}

func TestWriteDeveloperArtifactScaffoldAgentLoopProfile(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileAgentLoop, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	for _, path := range []string{
		filepath.Join(root, "generated", "agent-runs"),
		filepath.Join(root, "docs", "artifacts", "templates", "agent-loop-artifact.md"),
		filepath.Join(root, "docs", "artifacts", "source", "agent-loop-playbook.md"),
		filepath.Join(root, "scripts", "check-agent-loop-policy.mjs"),
	} {
		if !fileExists(path) && !dirExists(path) {
			t.Fatalf("expected agent-loop scaffold path %s", path)
		}
	}
	if !gitignoreHasLine(mustReadText(t, filepath.Join(root, ".gitignore")), "generated/agent-runs/") {
		t.Fatal("expected generated agent run output to be gitignored")
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	if artifacts["requestedProfile"] != string(artifactProfileAgentLoop) {
		t.Fatalf("expected requested agent-loop profile, got %#v", artifacts["requestedProfile"])
	}
	if artifacts["profile"] != string(artifactProfileDual) {
		t.Fatalf("expected effective dual profile, got %#v", artifacts["profile"])
	}
	if artifacts["specialization"] != "self-improving-agent-loop" {
		t.Fatalf("expected self-improving-agent-loop specialization, got %#v", artifacts["specialization"])
	}
	agentLoop, ok := artifacts["agentLoop"].(map[string]any)
	if !ok {
		t.Fatalf("expected agentLoop config, got %#v", artifacts["agentLoop"])
	}
	if agentLoop["enabled"] != true {
		t.Fatalf("expected agent loop enabled, got %#v", agentLoop["enabled"])
	}
	if agentLoop["traceDir"] != "generated/agent-runs" {
		t.Fatalf("expected generated/agent-runs trace dir, got %#v", agentLoop["traceDir"])
	}
	if agentLoop["defaultIssueTool"] != "beads" {
		t.Fatalf("expected beads issue tool, got %#v", agentLoop["defaultIssueTool"])
	}
	phases, ok := agentLoop["phases"].([]any)
	if !ok || !containsJSONValue(phases, "sense") || !containsJSONValue(phases, "learn") {
		t.Fatalf("expected sense/learn phases, got %#v", agentLoop["phases"])
	}
	approvalBoundaries, ok := agentLoop["humanApprovalRequiredFor"].([]any)
	if !ok || !containsJSONValue(approvalBoundaries, "permission expansion") {
		t.Fatalf("expected permission expansion approval boundary, got %#v", agentLoop["humanApprovalRequiredFor"])
	}
	scripts := packageScripts(t, filepath.Join(root, "package.json"))
	if scripts["agent-loop:check"] != "node scripts/check-agent-loop-policy.mjs" {
		t.Fatalf("expected agent loop check script, got %#v", scripts["agent-loop:check"])
	}
	if scripts["agent-loop:review"] != "node scripts/check-agent-loop-policy.mjs && node scripts/check-artifact-manifest.mjs" {
		t.Fatalf("expected agent loop review script, got %#v", scripts["agent-loop:review"])
	}
}

func TestWriteDeveloperArtifactScaffoldAgentLoopProfileWithoutBeads(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileAgentLoop, true, false, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	agentLoop := developerArtifactsConfig(t, config)["agentLoop"].(map[string]any)
	if agentLoop["defaultIssueTool"] != "explicit-human-request" {
		t.Fatalf("expected explicit request issue tool, got %#v", agentLoop["defaultIssueTool"])
	}
	learningOutputs, ok := agentLoop["learningOutputs"].([]any)
	if !ok {
		t.Fatalf("expected learningOutputs array, got %#v", agentLoop["learningOutputs"])
	}
	if containsJSONValue(learningOutputs, "bd remember insight") {
		t.Fatalf("did not expect Beads memory output when Beads is skipped: %#v", learningOutputs)
	}
	if !containsJSONValue(learningOutputs, "follow-up issue or source artifact") {
		t.Fatalf("expected non-Beads fallback learning output, got %#v", learningOutputs)
	}
}

func TestWriteDeveloperArtifactScaffoldModeling(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeUMLFirst); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	for _, path := range []string{
		filepath.Join(root, "docs", "artifacts", "source", "models"),
		filepath.Join(root, "generated", "review", "models"),
		filepath.Join(root, "docs", "artifacts", "templates", "model-diff-artifact.md"),
		filepath.Join(root, "scripts", "check-model-artifact-policy.mjs"),
	} {
		if !fileExists(path) && !dirExists(path) {
			t.Fatalf("expected modeling scaffold path %s", path)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	artifactTypes, ok := artifacts["artifactTypes"].([]any)
	if !ok || !containsJSONValue(artifactTypes, "model-diff") {
		t.Fatalf("expected model-diff artifact type, got %#v", artifacts["artifactTypes"])
	}
	modelPolicy, ok := artifacts["modelPolicy"].(map[string]any)
	if !ok {
		t.Fatalf("expected model policy config, got %#v", artifacts["modelPolicy"])
	}
	umlPolicy, ok := modelPolicy["uml"].(map[string]any)
	if !ok || umlPolicy["enabled"] != true {
		t.Fatalf("expected enabled UML policy, got %#v", modelPolicy["uml"])
	}
	if umlPolicy["sourceDir"] != "docs/artifacts/source/models" {
		t.Fatalf("expected model source dir, got %#v", umlPolicy["sourceDir"])
	}
	scripts := packageScripts(t, filepath.Join(root, "package.json"))
	if !strings.Contains(fmt.Sprint(scripts["artifacts:check"]), "check-model-artifact-policy.mjs") {
		t.Fatalf("expected artifacts:check to include model policy, got %#v", scripts["artifacts:check"])
	}
	if scripts["artifacts:model:check"] != "node scripts/check-model-artifact-policy.mjs" {
		t.Fatalf("expected model check script, got %#v", scripts["artifacts:model:check"])
	}
	if scripts["models:review"] != "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs" {
		t.Fatalf("expected model review script, got %#v", scripts["models:review"])
	}
}

func TestModelArtifactPolicyScriptAcceptsValidModelView(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeUMLFirst); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	sourcePath := filepath.Join(root, "docs", "artifacts", "source", "models", "enrolment-flow.md")
	source := "# Enrolment Flow\n\n```mermaid\nsequenceDiagram\n  actor Parent\n  Parent->>Portal: Apply\n```\n"
	mustWriteFile(t, sourcePath, source)
	mustWriteFile(t, filepath.Join(root, "generated", "review", "models", "enrolment-flow.html"), "<!doctype html><html><body><main><h1>Enrolment Flow</h1></main></body></html>")
	writeManifest(t, root, []map[string]any{
		{
			"id":               "model-enrolment-flow",
			"type":             "model-view",
			"source":           "docs/artifacts/source/models/enrolment-flow.md",
			"status":           "ready",
			"modelId":          "enrolment-flow",
			"modelKind":        "sequence",
			"notation":         "mermaid",
			"method":           "uml",
			"abstractionLevel": "runtime",
			"owner":            "system-modeler",
			"reviewSurface":    "generated/review/models/enrolment-flow.html",
			"sourceHash":       sha256Hex(source),
			"evidenceLinks":    []string{"tests/enrolment.test.ts"},
		},
	})

	runNodeScript(t, root, "scripts/check-model-artifact-policy.mjs", true)
}

func TestModelReviewGeneratorCreatesHumanHTML(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeUMLFirst); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	sourcePath := filepath.Join(root, "docs", "artifacts", "source", "models", "architecture.md")
	source := "# Architecture\n\n```mermaid\nflowchart LR\n  Web --> API\n```\n"
	mustWriteFile(t, sourcePath, source)
	mustWriteFile(t, filepath.Join(root, "docs", "artifacts", "source", "models", "architecture-screen.svg"), `<svg xmlns="http://www.w3.org/2000/svg" width="160" height="80"><rect width="160" height="80" fill="#eef6f6"/><text x="16" y="44" font-size="18">Review</text></svg>`)
	writeManifest(t, root, []map[string]any{
		{
			"id":               "model-architecture",
			"title":            "Architecture Review",
			"type":             "model-view",
			"source":           "docs/artifacts/source/models/architecture.md",
			"status":           "ready",
			"modelId":          "architecture",
			"modelKind":        "context",
			"notation":         "mermaid",
			"method":           "c4",
			"abstractionLevel": "runtime",
			"owner":            "system-modeler",
			"sourceHash":       sha256Hex(source),
			"evidenceLinks":    []string{"docs/architecture.md"},
			"screenshots": []map[string]any{
				{
					"path":    "docs/artifacts/source/models/architecture-screen.svg",
					"caption": "Generated review screenshot",
				},
			},
		},
	})

	runNodeScript(t, root, "scripts/generate-model-review.mjs", true)
	htmlPath := filepath.Join(root, "generated", "review", "models", "model-architecture.html")
	if !fileExists(htmlPath) {
		t.Fatal("expected generated model review HTML")
	}
	html := mustReadText(t, htmlPath)
	for _, want := range []string{"Overview", "Visuals", "Source", "Evidence", "Diff", "diagram-card", "data:image/svg+xml;base64"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected generated human HTML to contain %q", want)
		}
	}
	runNodeScript(t, root, "scripts/check-model-artifact-policy.mjs", true)
	runNodeScript(t, root, "scripts/check-artifact-html-policy.mjs", true)
}

func TestModelArtifactPolicyBaselineAcceptsEvidenceModelsWithoutHTML(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileCLI, true, true, modelingModeBaseline); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	source := "{\"dependencies\":[]}\n"
	mustWriteFile(t, filepath.Join(root, "docs", "artifacts", "source", "models", "dependencies.json"), source)
	writeManifest(t, root, []map[string]any{
		{
			"id":               "model-dependencies",
			"type":             "model-view",
			"source":           "docs/artifacts/source/models/dependencies.json",
			"status":           "ready",
			"modelId":          "dependencies",
			"modelKind":        "dependency",
			"notation":         "markdown",
			"method":           "evidence",
			"abstractionLevel": "runtime",
			"owner":            "system-modeler",
			"sourceHash":       sha256Hex(source),
			"evidenceLinks":    []string{"package.json"},
		},
	})

	runNodeScript(t, root, "scripts/check-model-artifact-policy.mjs", true)
}

func TestModelArtifactPolicyScriptRejectsInvalidFacetAndBrokenDiff(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeUMLFirst); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	source := "# Navigation\n"
	mustWriteFile(t, filepath.Join(root, "docs", "artifacts", "source", "models", "navigation.md"), source)
	writeManifest(t, root, []map[string]any{
		{
			"id":               "model-navigation",
			"type":             "model-view",
			"source":           "docs/artifacts/source/models/navigation.md",
			"status":           "draft",
			"modelId":          "navigation",
			"modelKind":        "activity",
			"notation":         "mermaid",
			"method":           "uwe",
			"facets":           []string{"unknown"},
			"abstractionLevel": "design",
			"owner":            "system-modeler",
			"reviewSurface":    "generated/review/models/navigation.html",
			"sourceHash":       sha256Hex(source),
		},
		{
			"id":               "model-navigation-diff",
			"type":             "model-diff",
			"source":           "docs/artifacts/source/models/navigation.md",
			"status":           "draft",
			"modelId":          "navigation",
			"modelKind":        "activity",
			"notation":         "mermaid",
			"method":           "uwe",
			"facets":           []string{"navigation"},
			"abstractionLevel": "design",
			"owner":            "system-modeler",
			"reviewSurface":    "generated/review/models/navigation-diff.html",
			"diff": map[string]any{
				"beforeArtifactId": "missing-before",
				"afterArtifactId":  "model-navigation",
				"method":           "source",
				"reviewSurface":    "generated/review/models/navigation-diff.html",
			},
		},
	})

	output := runNodeScript(t, root, "scripts/check-model-artifact-policy.mjs", false)
	if !strings.Contains(output, "unsupported UWE facet") || !strings.Contains(output, "references unknown artifact") {
		t.Fatalf("expected facet and diff failures, got:\n%s", output)
	}
}

func TestWriteDeveloperArtifactScaffoldResolvesAutoProfile(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileAuto, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "project.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	if artifacts["requestedProfile"] != string(artifactProfileAuto) {
		t.Fatalf("expected requested auto profile, got %#v", artifacts["requestedProfile"])
	}
	if artifacts["profile"] != string(artifactProfileDual) {
		t.Fatalf("expected effective dual profile, got %#v", artifacts["profile"])
	}
}

func TestWriteProjectSetupProof(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "apps", "web")
	mustMkdirAll(t, target)
	ctx := projectSetupContext{
		TargetDir:      target,
		OperationDir:   root,
		MonorepoRoot:   root,
		Monorepo:       true,
		Scope:          projectScopeAuto,
		PackageManager: packageManagerNpm,
	}
	proof := buildProjectSetupProof(ctx, projectSetupProofOptions{
		RequestedArtifactProfile: artifactProfileAgentLoop,
		EffectiveArtifactProfile: artifactProfileDual,
		RequestedModelingMode:    modelingModeUMLFirst,
		EffectiveModelingMode:    modelingModeUMLFirst,
		BeadsMode:                beadsSystem,
		BeadsWorktrees:           true,
	})

	if err := writeProjectSetupProof(root, proof); err != nil {
		t.Fatalf("writeProjectSetupProof returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".skill-harness", "setup-proof.json"))
	if err != nil {
		t.Fatalf("read setup proof: %v", err)
	}
	var decoded projectSetupProof
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("setup proof should be valid JSON: %v", err)
	}
	if decoded.Version != 1 {
		t.Fatalf("expected proof version 1, got %d", decoded.Version)
	}
	if decoded.Profiles.RequestedDeveloperArtifacts != artifactProfileAgentLoop {
		t.Fatalf("expected requested agent-loop profile, got %q", decoded.Profiles.RequestedDeveloperArtifacts)
	}
	if decoded.Tools["beads"].Status != "initialized" {
		t.Fatalf("expected beads initialized, got %#v", decoded.Tools["beads"])
	}
	if decoded.Checks["agentLoopPolicy"].Command != "node scripts/check-agent-loop-policy.mjs" {
		t.Fatalf("expected agent-loop check command, got %#v", decoded.Checks["agentLoopPolicy"])
	}
	if decoded.Checks["modelArtifactPolicy"].Command != "node scripts/check-model-artifact-policy.mjs" {
		t.Fatalf("expected model artifact check command, got %#v", decoded.Checks["modelArtifactPolicy"])
	}
	if !containsString(decoded.GeneratedPaths, ".skill-harness/setup-proof.json") {
		t.Fatalf("expected proof path in generated paths, got %#v", decoded.GeneratedPaths)
	}
	if !containsString(decoded.GeneratedPaths, "docs/artifacts/source/models/") {
		t.Fatalf("expected model source path in generated paths, got %#v", decoded.GeneratedPaths)
	}
}

func developerArtifactsConfig(t *testing.T, config map[string]any) map[string]any {
	t.Helper()
	capabilities, ok := config["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities map, got %#v", config["capabilities"])
	}
	artifacts, ok := capabilities["developerArtifacts"].(map[string]any)
	if !ok {
		t.Fatalf("expected developerArtifacts map, got %#v", capabilities["developerArtifacts"])
	}
	return artifacts
}

func packageScripts(t *testing.T, packagePath string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("package.json should be valid JSON: %v", err)
	}
	scripts, ok := metadata["scripts"].(map[string]any)
	if !ok {
		t.Fatalf("expected package scripts, got %#v", metadata["scripts"])
	}
	return scripts
}

func containsJSONValue(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func mustReadText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}

func writeManifest(t *testing.T, root string, artifacts []map[string]any) {
	t.Helper()
	data, err := json.MarshalIndent(map[string]any{
		"version": 1,
		"rules": map[string]any{
			"editSourceFirst":            true,
			"generatedReviewIsCanonical": false,
			"hashAlgorithm":              "sha256",
		},
		"artifacts": artifacts,
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	mustWriteFile(t, filepath.Join(root, "docs", "artifacts", "artifacts.manifest.json"), string(data)+"\n")
}

func runNodeScript(t *testing.T, root string, script string, wantSuccess bool) string {
	t.Helper()
	cmd := exec.Command("node", script)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if wantSuccess && err != nil {
		t.Fatalf("%s failed unexpectedly: %v\n%s", script, err, output)
	}
	if !wantSuccess && err == nil {
		t.Fatalf("%s succeeded unexpectedly:\n%s", script, output)
	}
	return string(output)
}
