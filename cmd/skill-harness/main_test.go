package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestResolveAgentStackAppliesOverlay(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".skill-harness", "agent-stack.json"), `{
  "version": 1,
  "baseline": {
    "source": "https://github.com/45ck/skill-harness.git",
    "channel": "main"
  },
  "profile": "minimal",
  "disabledPacks": ["demo-production-skills"],
  "agents": {
    "software-architect": {
      "removeSkills": ["public-private-hybrid-cloud-reviewer"],
      "addSkills": ["repo-architecture-reviewer"]
    }
  },
  "repoLocalPacks": ["packs/project-skills"]
}
`)

	deps := dependencyConfig{
		Repos: map[string]repoConfig{
			"software-architecture-skills":  {Path: "packs/software-architecture-skills"},
			"documentation-evidence-skills": {Path: "packs/documentation-evidence-skills"},
			"coding-workflow-skills":        {Path: "packs/coding-workflow-skills"},
			"frontier-agent-playbook":       {URL: "https://example.invalid/frontier.git"},
			"demo-production-skills":        {Path: "packs/demo-production-skills"},
		},
		Agents: map[string]agentConfig{
			"requirements-analyst": {Repos: []string{"documentation-evidence-skills", "frontier-agent-playbook"}},
			"software-architect":   {Repos: []string{"software-architecture-skills", "demo-production-skills"}},
			"workflow-engineer":    {Repos: []string{"coding-workflow-skills", "frontier-agent-playbook"}},
		},
	}
	loadouts := loadoutConfig{
		"requirements-analyst": {Skills: []string{"problem-statement-refiner"}},
		"software-architect":   {Skills: []string{"architecture-option-generator", "public-private-hybrid-cloud-reviewer"}},
		"workflow-engineer":    {Skills: []string{"issue-driven-delivery"}},
	}

	resolution, err := resolveAgentStack(root, deps, loadouts)
	if err != nil {
		t.Fatalf("resolveAgentStack returned error: %v", err)
	}
	if resolution.Profile != "minimal" {
		t.Fatalf("expected minimal profile, got %q", resolution.Profile)
	}
	if !containsString(resolution.EffectiveAgents, "software-architect") || containsString(resolution.EffectiveAgents, "security-reviewer") {
		t.Fatalf("unexpected effective agents: %#v", resolution.EffectiveAgents)
	}
	if containsString(resolution.EffectivePacks, "demo-production-skills") {
		t.Fatalf("disabled pack should not be effective: %#v", resolution.EffectivePacks)
	}
	architectSkills := resolution.AgentSkills["software-architect"]
	if containsString(architectSkills, "public-private-hybrid-cloud-reviewer") {
		t.Fatalf("removed skill remained in software-architect loadout: %#v", architectSkills)
	}
	if !containsString(architectSkills, "repo-architecture-reviewer") {
		t.Fatalf("repo-local skill missing from software-architect loadout: %#v", architectSkills)
	}
	if resolution.State != "overridden" {
		t.Fatalf("expected overridden state, got %q", resolution.State)
	}
	if len(resolution.Diagnostics) == 0 {
		t.Fatal("expected disabled-pack-required-by-agent warning")
	}
}

func TestResolveAgentStackReportsUnknownReferences(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".skill-harness", "agent-stack.json"), `{
  "version": 1,
  "profile": "default",
  "enabledAgents": ["missing-agent"],
  "disabledPacks": ["missing-pack"]
}
`)
	resolution, err := resolveAgentStack(root, dependencyConfig{Repos: map[string]repoConfig{}, Agents: map[string]agentConfig{}}, loadoutConfig{})
	if err != nil {
		t.Fatalf("resolveAgentStack returned error: %v", err)
	}
	if resolution.State != "conflicted" {
		t.Fatalf("expected conflicted state, got %q", resolution.State)
	}
	if !agentStackHasErrors(resolution.Diagnostics) {
		t.Fatalf("expected error diagnostics, got %#v", resolution.Diagnostics)
	}
}

func TestWriteDefaultAgentStackDoesNotOverwrite(t *testing.T) {
	root := t.TempDir()
	stackPath := filepath.Join(root, ".skill-harness", "agent-stack.json")
	mustWriteFile(t, stackPath, "{\"version\":1,\"profile\":\"minimal\"}\n")
	if err := writeDefaultAgentStack(root, false); err != nil {
		t.Fatalf("writeDefaultAgentStack returned error: %v", err)
	}
	if got := mustReadText(t, stackPath); got != "{\"version\":1,\"profile\":\"minimal\"}\n" {
		t.Fatalf("expected existing stack to be preserved, got %q", got)
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
		filepath.Join(root, "docs", "artifacts", "source", "product"),
		filepath.Join(root, "docs", "artifacts", "source", "business"),
		filepath.Join(root, "docs", "artifacts", "source", "data"),
		filepath.Join(root, "docs", "artifacts", "source", "research"),
		filepath.Join(root, "docs", "artifacts", "source", "ux"),
		filepath.Join(root, "docs", "artifacts", "templates"),
		filepath.Join(root, "generated", "review"),
		filepath.Join(root, "generated", "review", "product"),
		filepath.Join(root, "generated", "review", "business"),
		filepath.Join(root, "generated", "review", "data"),
		filepath.Join(root, "generated", "review", "research"),
		filepath.Join(root, "generated", "review", "ux"),
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
	if !fileExists(filepath.Join(root, "scripts", "generate-artifact-review.mjs")) {
		t.Fatal("expected generic artifact review generator script")
	}
	if !fileExists(filepath.Join(root, "scripts", "open-artifact-review.mjs")) {
		t.Fatal("expected artifact review opener script")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "artifacts.manifest.json")) {
		t.Fatal("expected artifact manifest")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "templates", "model-artifact.md")) {
		t.Fatal("expected model artifact template")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "templates", "visual-source-artifact.md")) {
		t.Fatal("expected visual source artifact template")
	}
	if !fileExists(filepath.Join(root, "docs", "artifacts", "templates", "e2e-product-system-atlas.md")) {
		t.Fatal("expected E2E product system atlas template")
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
	visualPolicy, ok := artifacts["visualArtifactPolicy"].(map[string]any)
	if !ok {
		t.Fatalf("expected visual artifact policy config, got %#v", artifacts["visualArtifactPolicy"])
	}
	if visualPolicy["doctrine"] != "visual-source-first" {
		t.Fatalf("expected visual-source-first doctrine, got %#v", visualPolicy["doctrine"])
	}
	if visualPolicy["defaultHumanSurface"] != "high-fidelity-html" {
		t.Fatalf("expected high-fidelity default surface, got %#v", visualPolicy["defaultHumanSurface"])
	}
	if visualPolicy["lowFidelityPolicy"] != "scratch-only-not-canonical" {
		t.Fatalf("expected low-fi scratch policy, got %#v", visualPolicy["lowFidelityPolicy"])
	}
	infographicPolicy, ok := artifacts["infographicPolicy"].(map[string]any)
	if !ok {
		t.Fatalf("expected infographic policy config, got %#v", artifacts["infographicPolicy"])
	}
	if infographicPolicy["defaultMode"] != "source-spec-to-static-review" || infographicPolicy["browserRuntime"] != "blocked-by-default-html-policy" {
		t.Fatalf("expected static infographic policy, got %#v", infographicPolicy)
	}
	infographicTools, ok := infographicPolicy["tools"].([]any)
	if !ok {
		t.Fatalf("expected infographic tools, got %#v", infographicPolicy["tools"])
	}
	for _, tool := range []string{"mermaid", "vega-lite", "observable-plot", "d3", "graphviz", "echarts", "rawgraphs", "chartjs"} {
		found := false
		for _, item := range infographicTools {
			toolConfig, _ := item.(map[string]any)
			if toolConfig["id"] == tool {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected infographic tool %q in %#v", tool, infographicTools)
		}
	}
	families, ok := visualPolicy["families"].([]any)
	if !ok || len(families) != 5 {
		t.Fatalf("expected five visual artifact families, got %#v", visualPolicy["families"])
	}
	artifactTypes, ok := artifacts["artifactTypes"].([]any)
	if !ok || !containsJSONValue(artifactTypes, "high-fidelity-prototype") || !containsJSONValue(artifactTypes, "claim-evidence-matrix") {
		t.Fatalf("expected visual artifact types, got %#v", artifacts["artifactTypes"])
	}
	scripts := packageScripts(t, filepath.Join(root, "package.json"))
	if scripts["artifacts:manifest:check"] != "node scripts/check-artifact-manifest.mjs" {
		t.Fatalf("expected artifact manifest check script, got %#v", scripts["artifacts:manifest:check"])
	}
	if scripts["artifacts:generate"] != "node scripts/generate-artifact-review.mjs" {
		t.Fatalf("expected generic artifact generate script, got %#v", scripts["artifacts:generate"])
	}
	if scripts["artifacts:review"] != "node scripts/generate-artifact-review.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs" {
		t.Fatalf("expected generic artifact review script, got %#v", scripts["artifacts:review"])
	}
	if scripts["artifacts:open"] != "node scripts/open-artifact-review.mjs" {
		t.Fatalf("expected artifact open script, got %#v", scripts["artifacts:open"])
	}
	reviewSurface, ok := artifacts["reviewSurface"].(map[string]any)
	if !ok {
		t.Fatalf("expected reviewSurface config, got %#v", artifacts["reviewSurface"])
	}
	openPolicy, ok := reviewSurface["openPolicy"].(map[string]any)
	if !ok || openPolicy["preferHostBrowserTool"] != true {
		t.Fatalf("expected host browser open policy, got %#v", reviewSurface["openPolicy"])
	}
	htmlPolicy, ok := artifacts["htmlPolicy"].(map[string]any)
	if !ok {
		t.Fatalf("expected htmlPolicy config, got %#v", artifacts["htmlPolicy"])
	}
	interactionLanes, ok := htmlPolicy["interactionLanes"].(map[string]any)
	if !ok || interactionLanes["default"] == nil || interactionLanes["reviewed-inline-js"] == nil {
		t.Fatalf("expected HTML interaction lanes, got %#v", htmlPolicy["interactionLanes"])
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
		filepath.Join(root, "scripts", "check-model-inventory.mjs"),
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
	if scripts["artifacts:generate"] != "node scripts/generate-artifact-review.mjs && node scripts/generate-model-review.mjs" {
		t.Fatalf("expected combined artifact generate script, got %#v", scripts["artifacts:generate"])
	}
	if scripts["artifacts:review"] != "node scripts/generate-artifact-review.mjs && node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs" {
		t.Fatalf("expected combined artifact review script, got %#v", scripts["artifacts:review"])
	}
	if scripts["artifacts:model:review"] != "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs" {
		t.Fatalf("expected strict model artifact review script, got %#v", scripts["artifacts:model:review"])
	}
	if scripts["artifacts:model:check"] != "node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/generate-model-review.mjs --check" {
		t.Fatalf("expected model check script, got %#v", scripts["artifacts:model:check"])
	}
	if scripts["models:check"] != "node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/generate-model-review.mjs --check" {
		t.Fatalf("expected models check script, got %#v", scripts["models:check"])
	}
	if scripts["models:drift"] != "node scripts/generate-model-review.mjs --check" {
		t.Fatalf("expected model drift script, got %#v", scripts["models:drift"])
	}
	if scripts["models:review"] != "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs" {
		t.Fatalf("expected model review script, got %#v", scripts["models:review"])
	}
	if scripts["models:open"] != "node scripts/open-artifact-review.mjs generated/review/models/index.html" {
		t.Fatalf("expected model open script, got %#v", scripts["models:open"])
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
	runNodeScript(t, root, "scripts/generate-model-review.mjs", true, "--check")
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

	mustWriteFile(t, sourcePath, source+"\nChanged without regeneration.\n")
	runNodeScript(t, root, "scripts/generate-model-review.mjs", false, "--check")
}

func TestArtifactReviewGeneratorCreatesInfographicHTML(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	sourcePath := filepath.Join(root, "docs", "artifacts", "source", "research", "html-artifacts.md")
	source := strings.Join([]string{
		"# HTML Artifact Discovery",
		"",
		"## Purpose",
		"",
		"Decide how human review artifacts should be generated.",
		"",
		"```artifact-infographic",
		"{\"title\":\"Renderer Fit\",\"tool\":\"vega-lite\",\"kind\":\"bar\",\"values\":[{\"label\":\"Source\",\"value\":3},{\"label\":\"HTML\",\"value\":2}]}",
		"```",
		"",
		"```artifact-infographic",
		"{\"title\":\"Review Flow\",\"tool\":\"graphviz\",\"kind\":\"graph\",\"edges\":[[\"Source\",\"Manifest\"],[\"Manifest\",\"HTML\"]]}",
		"```",
		"",
		"## Evidence",
		"",
		"- docs/developer-artifacts.md",
		"",
	}, "\n")
	mustWriteFile(t, sourcePath, source)
	writeManifest(t, root, []map[string]any{
		{
			"id":             "html-artifact-discovery",
			"type":           "research-synthesis",
			"source":         "docs/artifacts/source/research/html-artifacts.md",
			"status":         "ready",
			"owner":          "research-writer",
			"reviewRequired": true,
			"evidenceLinks":  []string{"docs/developer-artifacts.md"},
		},
	})

	runNodeScript(t, root, "scripts/generate-artifact-review.mjs", true)
	runNodeScript(t, root, "scripts/generate-artifact-review.mjs", true, "--check")
	htmlPath := filepath.Join(root, "generated", "review", "research", "html-artifact-discovery.html")
	if !fileExists(htmlPath) {
		t.Fatal("expected generated infographic HTML")
	}
	html := mustReadText(t, htmlPath)
	for _, want := range []string{"Infographic Snapshot", "Open-Source Infographic Toolkit", "Static Infographic Specs", "Renderer Fit", "Review Flow", "vega-lite", "graphviz", "Source-To-Review Flow", "Canonical Source", "Evidence", "Content-Security-Policy"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected generated artifact HTML to contain %q", want)
		}
	}
	if !strings.Contains(html, "<h1>HTML Artifact Discovery</h1>") {
		t.Fatalf("expected generated artifact HTML to use source H1 as fallback title")
	}
	runNodeScript(t, root, "scripts/check-artifact-manifest.mjs", true)
	runNodeScript(t, root, "scripts/check-artifact-html-policy.mjs", true)

	output := runNodeScript(t, root, "scripts/open-artifact-review.mjs", true, "--json", "--print", "html-artifact-discovery")
	if !strings.Contains(output, `"repoPath": "generated/review/research/html-artifact-discovery.html"`) {
		t.Fatalf("expected opener to resolve artifact id, got:\n%s", output)
	}

	mustWriteFile(t, sourcePath, source+"\nChanged without regeneration.\n")
	runNodeScript(t, root, "scripts/generate-artifact-review.mjs", false, "--check")
}

func TestArtifactReviewGeneratorEmbedsScreenshotEvidence(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	mustWriteFile(t, filepath.Join(root, "node_modules", "@viz-js", "viz", "package.json"), "{\"type\":\"module\"}\n")
	mustWriteFile(t, filepath.Join(root, "node_modules", "@viz-js", "viz", "index.js"), `
export async function instance() {
  return {
    renderString() {
      return '<svg width="320pt" height="180pt" viewBox="0.00 0.00 320.00 180.00" xmlns="http://www.w3.org/2000/svg"><g id="node1" class="node"><title>Landing</title><polygon fill="#effaf8" stroke="none" points="8,-150 8,-130 142,-130 142,-150 8,-150"/><text x="20" y="-136">«navigation node»</text><polygon fill="#eef2f6" stroke="none" points="8,-130 8,-42 142,-42 142,-130 8,-130"/><text x="32" y="-80">screenshot:Landing</text><polygon fill="none" stroke="#9fb3c8" points="8,-150 142,-150 142,-20 8,-20 8,-150"/></g><g id="node2" class="node"><title>Auth</title><polygon fill="#effaf8" stroke="none" points="178,-150 178,-130 312,-130 312,-150 178,-150"/><text x="190" y="-136">«navigation node»</text><polygon fill="#eef2f6" stroke="none" points="178,-130 178,-42 312,-42 312,-130 178,-130"/><text x="202" y="-80">screenshot:Auth</text><text x="190" y="-28">/login</text><polygon fill="none" stroke="#9fb3c8" points="178,-150 312,-150 312,-20 178,-20 178,-150"/></g></svg>';
    }
  };
}
`)
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	sourcePath := filepath.Join(root, "docs", "artifacts", "source", "product", "atlas.md")
	source := strings.Join([]string{
		"# App Atlas",
		"",
		"## Purpose",
		"",
		"Inspect a UWE navigation model with screenshot evidence.",
		"",
		"```artifact-infographic",
		"{\"title\":\"UWE Screenshot Nodes\",\"tool\":\"graphviz\",\"kind\":\"uwe-navigation\",\"navigationClasses\":[\"Visitor acquisition\"],\"nodes\":[{\"id\":\"Landing\",\"label\":\"Landing\",\"route\":\"/\",\"navigationClass\":\"Visitor acquisition\",\"facet\":\"navigation\",\"screenshot\":\"generated/review/evidence/atlas/landing.svg\"},{\"id\":\"Auth\",\"label\":\"Auth\",\"route\":\"/login\",\"navigationClass\":\"Visitor acquisition\",\"facet\":\"access\"}],\"edges\":[[\"Landing\",\"Auth\",\"sign in\"]]}",
		"```",
		"",
	}, "\n")
	mustWriteFile(t, sourcePath, source)
	imagePath := filepath.Join(root, "generated", "review", "evidence", "atlas", "landing.svg")
	mustWriteFile(t, imagePath, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 80"><rect width="120" height="80" fill="#eef2f6"/><text x="16" y="42">Landing</text></svg>`)
	writeManifest(t, root, []map[string]any{
		{
			"id":             "app-atlas",
			"type":           "e2e-product-system-atlas",
			"family":         "product",
			"source":         "docs/artifacts/source/product/atlas.md",
			"status":         "ready",
			"owner":          "system-modeler",
			"reviewRequired": true,
			"evidenceLinks":  []string{"docs/developer-artifacts.md"},
			"screenshots": []map[string]any{
				{
					"path":    "generated/review/evidence/atlas/landing.svg",
					"caption": "Landing page",
					"alt":     "Landing page screenshot",
				},
			},
		},
	})

	runNodeScript(t, root, "scripts/generate-artifact-review.mjs", true)
	htmlPath := filepath.Join(root, "generated", "review", "product", "app-atlas.html")
	html := mustReadText(t, htmlPath)
	for _, want := range []string{"Screenshots And Evidence Images", "Landing page", "data:image/svg+xml;base64,", "UWE Screenshot Nodes", "graphviz-render", "«navigation node»", "/login", "<image href=\"data:image/svg+xml;base64,"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected generated artifact HTML to contain %q", want)
		}
	}
	runNodeScript(t, root, "scripts/check-artifact-html-policy.mjs", true)
}

func TestArtifactReviewGeneratorRejectsUnsafeReviewSurface(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeOff); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	source := "# Unsafe Surface\n\n## Purpose\n\nExercise write-boundary checks.\n"
	mustWriteFile(t, filepath.Join(root, "docs", "artifacts", "source", "research", "unsafe.md"), source)
	writeManifest(t, root, []map[string]any{
		{
			"id":             "unsafe-surface",
			"type":           "research-synthesis",
			"source":         "docs/artifacts/source/research/unsafe.md",
			"status":         "ready",
			"owner":          "security-reviewer",
			"reviewRequired": true,
			"reviewSurface":  "../escaped.html",
			"renderer":       "skill-harness artifact review generator",
			"sourceHash":     sha256Hex(source),
			"evidenceLinks":  []string{"docs/developer-artifacts.md"},
		},
	})

	output := runNodeScript(t, root, "scripts/generate-artifact-review.mjs", false)
	if !strings.Contains(output, "reviewSurface must be an HTML file under generated/review") {
		t.Fatalf("expected unsafe reviewSurface failure, got:\n%s", output)
	}
	if fileExists(filepath.Join(root, "escaped.html")) || fileExists(filepath.Join(root, "..", "escaped.html")) {
		t.Fatal("unsafe reviewSurface created an escaped file")
	}
}

func TestArtifactReviewOpenScriptPrintsDiscoveredTarget(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true, true, modelingModeUMLFirst); err != nil {
		t.Fatalf("writeDeveloperArtifactScaffold returned error: %v", err)
	}

	mustWriteFile(t, filepath.Join(root, "generated", "review", "index.html"), "<!doctype html><html><body><main><h1>Artifacts</h1></main></body></html>")
	mustWriteFile(t, filepath.Join(root, "generated", "review", "models", "index.html"), "<!doctype html><html><body><main><h1>Models</h1></main></body></html>")
	output := runNodeScript(t, root, "scripts/open-artifact-review.mjs", true, "--print")
	if !strings.Contains(output, "file://") || !strings.Contains(output, "generated/review/index.html") {
		t.Fatalf("expected printed file URL for artifact review index, got:\n%s", output)
	}
	jsonOutput := runNodeScript(t, root, "scripts/open-artifact-review.mjs", true, "--json", "--print")
	for _, want := range []string{`"hostAction":`, `"openMode": "print"`, `"repoPath": "generated/review/index.html"`, `"url": "file://`} {
		if !strings.Contains(jsonOutput, want) {
			t.Fatalf("expected JSON opener output to contain %q, got:\n%s", want, jsonOutput)
		}
	}
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

func TestCLISetupProjectMediaProfileScaffoldsAndWiresTools(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "package.json"), "{\n  \"name\": \"fixture\",\n  \"private\": true\n}\n")
	tooling := fakeTooling(t, "npm", "npx")

	runSkillHarnessCLI(t, binary, tooling, "setup-project",
		"--dir", projectDir,
		"--package-manager", "npm",
		"--developer-artifacts-profile", "media",
		"--skip-beads",
		"--skip-claude-settings",
	)

	commands := tooling.commands(t)
	assertCommandContains(t, commands, "npm", "install -D @45ck/noslop github:45ck/agent-docs")
	assertCommandContains(t, commands, "npx", "agent-docs init")
	assertCommandContains(t, commands, "npx", "noslop init")
	assertCommandContains(t, commands, "npx", "agent-docs install-gates --quality")

	for _, path := range []string{
		filepath.Join(projectDir, ".skill-harness", "agent-stack.json"),
		filepath.Join(projectDir, ".skill-harness", "project.json"),
		filepath.Join(projectDir, ".skill-harness", "setup-proof.json"),
		filepath.Join(projectDir, "docs", "artifacts", "artifacts.manifest.json"),
		filepath.Join(projectDir, "docs", "artifacts", "templates", "demo-artifact.md"),
		filepath.Join(projectDir, "generated", "media"),
		filepath.Join(projectDir, "scripts", "check-artifact-manifest.mjs"),
		filepath.Join(projectDir, "scripts", "check-artifact-html-policy.mjs"),
		filepath.Join(projectDir, "scripts", "open-artifact-review.mjs"),
	} {
		if !fileExists(path) && !dirExists(path) {
			t.Fatalf("expected setup-project to create %s", path)
		}
	}

	var config map[string]any
	if err := json.Unmarshal([]byte(mustReadText(t, filepath.Join(projectDir, ".skill-harness", "project.json"))), &config); err != nil {
		t.Fatalf("project config should be valid JSON: %v", err)
	}
	artifacts := developerArtifactsConfig(t, config)
	if artifacts["requestedProfile"] != string(artifactProfileMedia) {
		t.Fatalf("expected requested media profile, got %#v", artifacts["requestedProfile"])
	}
	if artifacts["profile"] != string(artifactProfileDual) {
		t.Fatalf("expected media to resolve to dual, got %#v", artifacts["profile"])
	}

	var proof projectSetupProof
	if err := json.Unmarshal([]byte(mustReadText(t, filepath.Join(projectDir, ".skill-harness", "setup-proof.json"))), &proof); err != nil {
		t.Fatalf("setup proof should be valid JSON: %v", err)
	}
	if proof.Tools["noslop"].Status != "initialized" {
		t.Fatalf("expected noslop initialized proof, got %#v", proof.Tools["noslop"])
	}
	if proof.Tools["agentDocs"].Status != "quality-gates-installed" {
		t.Fatalf("expected agent-docs quality gate proof, got %#v", proof.Tools["agentDocs"])
	}
	if proof.Tools["beads"].Status != "skipped" {
		t.Fatalf("expected beads skipped proof, got %#v", proof.Tools["beads"])
	}
	if proof.Tools["claudeSettings"].Status != "skipped" {
		t.Fatalf("expected claude settings skipped proof, got %#v", proof.Tools["claudeSettings"])
	}
	if proof.Tools["agentStack"].Status != "scaffolded" {
		t.Fatalf("expected agent stack scaffolded proof, got %#v", proof.Tools["agentStack"])
	}
	if !containsString(proof.GeneratedPaths, "generated/media/") {
		t.Fatalf("expected media generated path in proof, got %#v", proof.GeneratedPaths)
	}

	scripts := packageScripts(t, filepath.Join(projectDir, "package.json"))
	if scripts["artifacts:manifest:check"] != "node scripts/check-artifact-manifest.mjs" {
		t.Fatalf("expected artifact manifest script, got %#v", scripts["artifacts:manifest:check"])
	}
	if scripts["artifacts:open"] != "node scripts/open-artifact-review.mjs" {
		t.Fatalf("expected artifact open script, got %#v", scripts["artifacts:open"])
	}
}

func TestCLIInstallPacksOnlyBootstrapsSelectedPack(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	tooling := fakeTooling(t, "python")

	runSkillHarnessCLI(t, binary, tooling, "install", "--packs-only", "--packs", "developer-artifact-skills")

	commands := tooling.commands(t)
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"bootstrap_dependencies.py --repo developer-artifact-skills")
	assertNoCommandContains(t, commands, "render_claude_agents.py")
	assertNoCommandContains(t, commands, "render_codex_agents.py")
	assertNoCommandContains(t, commands, "check_dependencies.py")
}

func TestCLIInstallAgentsOnlyRendersAndChecksSelectedAgent(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	tooling := fakeTooling(t, "python")

	runSkillHarnessCLI(t, binary, tooling, "install", "--agents-only", "--agents", "requirements-analyst")

	commands := tooling.commands(t)
	assertNoCommandContains(t, commands, "bootstrap_dependencies.py")
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"render_claude_agents.py --agent requirements-analyst")
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"render_codex_agents.py --agent requirements-analyst")
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"check_dependencies.py --agent requirements-analyst")
}

func TestCLIRenderAndCheckForwardAgentArgs(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	tooling := fakeTooling(t, "python")

	runSkillHarnessCLI(t, binary, tooling, "render", "--agents", "requirements-analyst")
	runSkillHarnessCLI(t, binary, tooling, "check", "--agents", "requirements-analyst")

	commands := tooling.commands(t)
	assertNoCommandContains(t, commands, "render_claude_agents.py")
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"render_codex_agents.py --agent requirements-analyst")
	assertCommandContains(t, commands, "python", "scripts"+string(os.PathSeparator)+"check_dependencies.py --agent requirements-analyst")
	assertNoCommandContains(t, commands, "bootstrap_dependencies.py")
}

func TestCLIResolvePrintsAgentStackJSON(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, ".skill-harness", "agent-stack.json"), `{
  "version": 1,
  "profile": "minimal",
  "disabledAgents": ["workflow-engineer"]
}
`)
	output := runSkillHarnessCLI(t, binary, fakeTooling(t), "resolve", "--dir", projectDir, "--json")
	var resolution agentStackResolution
	if err := json.Unmarshal([]byte(output), &resolution); err != nil {
		t.Fatalf("resolve output should be JSON: %v\n%s", err, output)
	}
	if resolution.Profile != "minimal" {
		t.Fatalf("expected minimal profile, got %q", resolution.Profile)
	}
	if containsString(resolution.EffectiveAgents, "workflow-engineer") {
		t.Fatalf("disabled workflow-engineer should not be effective: %#v", resolution.EffectiveAgents)
	}
	if !containsString(resolution.OptOuts.DisabledAgents, "workflow-engineer") {
		t.Fatalf("expected disabled agent in opt-outs, got %#v", resolution.OptOuts.DisabledAgents)
	}
}

func TestCLIBootstrapAgentNativeWritesStackLockAndProof(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()

	output := runSkillHarnessCLI(t, binary, fakeTooling(t), "bootstrap", "--agent-native", "--dir", projectDir, "--json")
	var result struct {
		State      string               `json:"state"`
		LockPath   string               `json:"lockPath"`
		ProofPath  string               `json:"proofPath"`
		Resolution agentStackResolution `json:"resolution"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("bootstrap output should be JSON: %v\n%s", err, output)
	}
	if result.State != "clean" || result.Resolution.Profile != "default" {
		t.Fatalf("expected clean default bootstrap result, got state=%q profile=%q", result.State, result.Resolution.Profile)
	}
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "agent-stack.json")) {
		t.Fatalf("expected agent stack config")
	}
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "agent-stack.lock.json")) {
		t.Fatalf("expected agent stack lock")
	}
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "setup-proof.json")) {
		t.Fatalf("expected setup proof")
	}
	var proof projectSetupProof
	if err := json.Unmarshal([]byte(mustReadText(t, filepath.Join(projectDir, ".skill-harness", "setup-proof.json"))), &proof); err != nil {
		t.Fatalf("setup proof should be JSON: %v", err)
	}
	if proof.Tools["agentStackLock"].Status != "written" {
		t.Fatalf("expected agentStackLock proof, got %#v", proof.Tools["agentStackLock"])
	}
}

func TestCLIBootstrapAgentNativeWritesProofBesideStackInMonorepo(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "web")
	mustWriteFile(t, filepath.Join(root, "pnpm-workspace.yaml"), "packages:\n  - apps/*\n")
	mustWriteFile(t, filepath.Join(root, "pnpm-lock.yaml"), "lockfileVersion: '9.0'\n")
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")
	mustWriteFile(t, filepath.Join(appDir, "package.json"), "{\n  \"name\": \"web\",\n  \"private\": true\n}\n")

	runSkillHarnessCLI(t, binary, fakeTooling(t), "bootstrap", "--agent-native", "--dir", appDir)

	if !fileExists(filepath.Join(appDir, ".skill-harness", "agent-stack.json")) {
		t.Fatalf("expected agent stack in app dir")
	}
	if !fileExists(filepath.Join(appDir, ".skill-harness", "setup-proof.json")) {
		t.Fatalf("expected setup proof beside app stack")
	}
	if fileExists(filepath.Join(root, ".skill-harness", "setup-proof.json")) {
		t.Fatalf("did not expect bootstrap proof to be written at monorepo root")
	}
}

func TestCLIUpdateProjectWritesAgentStackLock(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, ".skill-harness", "agent-stack.json"), `{
  "version": 1,
  "profile": "minimal",
  "disabledAgents": ["workflow-engineer"]
}
`)

	output := runSkillHarnessCLI(t, binary, fakeTooling(t), "update-project", "--dir", projectDir, "--write-lock", "--json")
	var result struct {
		State      string               `json:"state"`
		Resolution agentStackResolution `json:"resolution"`
		Lock       agentStackLock       `json:"lock"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("update-project output should be JSON: %v\n%s", err, output)
	}
	if result.State != "overridden" || result.Resolution.Profile != "minimal" {
		t.Fatalf("expected overridden minimal update, got state=%q profile=%q", result.State, result.Resolution.Profile)
	}
	if result.Lock.Profile != "minimal" || result.Lock.OverlayHash == "" {
		t.Fatalf("expected agent-stack lock with profile and overlay hash, got %#v", result.Lock)
	}
	if !agentStackSurfaceStatus(result.Lock.Surfaces, ".skill-harness/agent-stack.lock.json", "present") {
		t.Fatalf("expected lock surface to be present, got %#v", result.Lock.Surfaces)
	}
	if !containsString(result.Lock.OptOuts.DisabledAgents, "workflow-engineer") {
		t.Fatalf("expected disabled agent in lock opt-outs, got %#v", result.Lock.OptOuts)
	}
	var lock agentStackLock
	if err := json.Unmarshal([]byte(mustReadText(t, filepath.Join(projectDir, ".skill-harness", "agent-stack.lock.json"))), &lock); err != nil {
		t.Fatalf("agent stack lock should be JSON: %v", err)
	}
	if lock.Profile != "minimal" || len(lock.AgentSkills) == 0 {
		t.Fatalf("expected resolved lock content, got %#v", lock)
	}
}

func TestRepoAuditDetectsRepoLocalOverlays(t *testing.T) {
	root := repoRootForTest(t)
	deps := loadDependencies(root)
	loadouts := loadLoadouts(root)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "AGENTS.md"), "# Project rules\n")
	mustWriteFile(t, filepath.Join(projectDir, ".codex", "skills", "custom-debug", "SKILL.md"), "# Custom debug\n")

	report := buildRepoAuditReport(root, deps, loadouts, projectDir)

	if report.State != "unmanaged" {
		t.Fatalf("expected unmanaged repo, got %q", report.State)
	}
	if !containsString(report.LocalSkills[".codex/skills"], "custom-debug") {
		t.Fatalf("expected repo-local skill discovery, got %#v", report.LocalSkills)
	}
	var agentsSurface repoSurfaceReport
	for _, surface := range report.Surfaces {
		if surface.Path == "AGENTS.md" {
			agentsSurface = surface
			break
		}
	}
	if agentsSurface.Status != "present" || agentsSurface.Mode != repoSurfaceOwned {
		t.Fatalf("expected AGENTS.md to be present and owned, got %#v", agentsSurface)
	}
}

func TestCLIRepoInitAndSyncWritesManifestLockAndReport(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "AGENTS.md"), "# Project rules\n")

	runSkillHarnessCLI(t, binary, fakeTooling(t), "repo", "init", "--dir", projectDir, "--profile", "minimal")
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "baseline.manifest.json")) {
		t.Fatalf("expected baseline manifest to be written")
	}
	var manifest repoBaselineManifest
	if err := json.Unmarshal([]byte(mustReadText(t, filepath.Join(projectDir, ".skill-harness", "baseline.manifest.json"))), &manifest); err != nil {
		t.Fatalf("baseline manifest should be JSON: %v", err)
	}
	if manifest.Surfaces[".skill-harness/setup-proof.json"].Mode != repoSurfaceIgnored {
		t.Fatalf("absent generated setup proof should default to ignored, got %#v", manifest.Surfaces[".skill-harness/setup-proof.json"])
	}

	runSkillHarnessCLI(t, binary, fakeTooling(t), "repo", "sync", "--dir", projectDir)
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "baseline.lock.json")) {
		t.Fatalf("expected baseline lock to be written")
	}
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "update-report.json")) {
		t.Fatalf("expected update report to be written")
	}

	output := runSkillHarnessCLI(t, binary, fakeTooling(t), "repo", "audit", "--dir", projectDir, "--json")
	var report repoAuditReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("repo audit output should be JSON: %v\n%s", err, output)
	}
	if report.State != "managed" || report.Profile != "minimal" {
		t.Fatalf("expected managed minimal report, got state=%q profile=%q", report.State, report.Profile)
	}
	if hasRepoFinding(report.Findings, "surface-missing") {
		t.Fatalf("repo init should not create immediate drift for absent default surfaces, got %#v", report.Findings)
	}

	runSkillHarnessCLI(t, binary, fakeTooling(t), "repo", "lock", "--dir", projectDir)
	if !fileExists(filepath.Join(projectDir, ".skill-harness", "baseline.lock.json")) {
		t.Fatalf("expected repo lock to write baseline lock")
	}
}

func TestRepoInitPinsSkillHarnessRevisionFromRoot(t *testing.T) {
	root := repoRootForTest(t)
	deps := loadDependencies(root)
	loadouts := loadLoadouts(root)
	projectDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir target: %v", err)
	}

	manifest := defaultRepoBaselineManifest(root, "minimal", projectDir)
	expected := currentGitRevision(root)
	if expected == "" {
		t.Skip("skill-harness root is not a git checkout")
	}
	if manifest.Baseline.Pin != expected {
		t.Fatalf("expected baseline pin %q from skill-harness root, got %q", expected, manifest.Baseline.Pin)
	}
	if err := validateRepoManifest(manifest, deps, loadouts); err != nil {
		t.Fatalf("default manifest should validate: %v", err)
	}
}

func TestRepoDriftReportsMissingManagedSurface(t *testing.T) {
	root := repoRootForTest(t)
	deps := loadDependencies(root)
	loadouts := loadLoadouts(root)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, ".skill-harness", "baseline.manifest.json"), `{
  "version": 1,
  "profile": "minimal",
  "baseline": {"source": "skill-harness", "channel": "default"},
  "surfaces": {
    "AGENTS.md": {"mode": "managed-section"}
  }
}
`)

	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if !hasRepoFinding(report.Findings, "surface-missing") {
		t.Fatalf("expected surface-missing finding, got %#v", report.Findings)
	}
}

func TestCLIInstallUsesResolvedAgentStackRepos(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, ".skill-harness", "agent-stack.json"), `{
  "version": 1,
  "profile": "minimal",
  "disabledPacks": ["frontier-agent-playbook"]
}
`)
	tooling := fakeTooling(t, "python")

	runSkillHarnessCLI(t, binary, tooling, "install", "--dir", projectDir, "--packs-only")

	commands := tooling.commands(t)
	assertCommandContains(t, commands, "python", "bootstrap_dependencies.py")
	assertNoCommandContains(t, commands, "--repo frontier-agent-playbook")
}

func TestCLIInstallWithDirRequiresAgentStack(t *testing.T) {
	binary := buildSkillHarnessBinary(t)
	projectDir := t.TempDir()

	output, err := runSkillHarnessCLIExpectError(t, binary, fakeTooling(t), "install", "--dir", projectDir)
	if err == nil {
		t.Fatalf("expected install --dir to fail without agent-stack.json")
	}
	if !strings.Contains(output, "missing") || !strings.Contains(output, "agent-stack.json") {
		t.Fatalf("expected missing agent-stack error, got %q", output)
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

func hasRepoFinding(findings []repoFinding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func agentStackSurfaceStatus(surfaces []agentStackSurfaceLock, path, status string) bool {
	for _, surface := range surfaces {
		if surface.Path == path && surface.Status == status {
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

func runNodeScript(t *testing.T, root string, script string, wantSuccess bool, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{script}, args...)
	cmd := exec.Command("node", cmdArgs...)
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

type fakeToolingEnv struct {
	binDir  string
	logPath string
}

func buildSkillHarnessBinary(t *testing.T) string {
	t.Helper()
	outDir := t.TempDir()
	binary := filepath.Join(outDir, "skill-harness-test"+exeSuffix())
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/skill-harness")
	cmd.Dir = repoRootForTest(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build skill-harness test binary: %v\n%s", err, output)
	}
	return binary
}

func runSkillHarnessCLI(t *testing.T, binary string, tooling fakeToolingEnv, args ...string) string {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = repoRootForTest(t)
	cmd.Env = append(os.Environ(),
		"PATH="+tooling.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SKILL_HARNESS_COMMAND_LOG="+tooling.logPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skill-harness %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func runSkillHarnessCLIExpectError(t *testing.T, binary string, tooling fakeToolingEnv, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = repoRootForTest(t)
	cmd.Env = append(os.Environ(),
		"PATH="+tooling.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SKILL_HARNESS_COMMAND_LOG="+tooling.logPath,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}
	return string(output), err
}

func fakeTooling(t *testing.T, names ...string) fakeToolingEnv {
	t.Helper()
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	mustMkdirAll(t, binDir)
	logPath := filepath.Join(root, "commands.log")
	for _, name := range names {
		writeFakeTool(t, binDir, name)
	}
	return fakeToolingEnv{binDir: binDir, logPath: logPath}
}

func writeFakeTool(t *testing.T, binDir string, name string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		path := filepath.Join(binDir, name+".cmd")
		content := "@echo off\r\n" +
			"echo %~n0 %*>>\"%SKILL_HARNESS_COMMAND_LOG%\"\r\n" +
			"exit /b 0\r\n"
		mustWriteFile(t, path, content)
		return
	}
	path := filepath.Join(binDir, name)
	content := "#!/bin/sh\n" +
		"printf '%s' \"$(basename \"$0\")\" >> \"$SKILL_HARNESS_COMMAND_LOG\"\n" +
		"for arg in \"$@\"; do printf ' %s' \"$arg\" >> \"$SKILL_HARNESS_COMMAND_LOG\"; done\n" +
		"printf '\\n' >> \"$SKILL_HARNESS_COMMAND_LOG\"\n"
	mustWriteFile(t, path, content)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("chmod fake tool %s: %v", path, err)
	}
}

func (env fakeToolingEnv) commands(t *testing.T) []string {
	t.Helper()
	if !fileExists(env.logPath) {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(mustReadText(t, env.logPath)), "\n")
	out := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func assertCommandContains(t *testing.T, commands []string, tool string, want string) {
	t.Helper()
	normalizedWant := normalizeCommandForAssert(want)
	for _, command := range commands {
		normalized := normalizeCommandForAssert(command)
		if strings.HasPrefix(normalized, tool+" ") && strings.Contains(normalized, normalizedWant) {
			return
		}
	}
	t.Fatalf("expected %s command containing %q, got:\n%s", tool, want, strings.Join(commands, "\n"))
}

func assertNoCommandContains(t *testing.T, commands []string, want string) {
	t.Helper()
	normalizedWant := normalizeCommandForAssert(want)
	for _, command := range commands {
		if strings.Contains(normalizeCommandForAssert(command), normalizedWant) {
			t.Fatalf("expected no command containing %q, got:\n%s", want, strings.Join(commands, "\n"))
		}
	}
}

func normalizeCommandForAssert(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	return strings.ReplaceAll(normalized, "\"", "")
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
