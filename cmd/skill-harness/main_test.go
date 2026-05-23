package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestWriteDeveloperArtifactScaffold(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileDual, true); err != nil {
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
	if !gitignoreHasLine(mustReadText(t, filepath.Join(root, ".gitignore")), "generated/review/") {
		t.Fatal("expected generated review output to be gitignored")
	}
}

func TestWriteDeveloperArtifactScaffoldResolvesAutoProfile(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "package.json"), "{\n  \"name\": \"repo\",\n  \"private\": true\n}\n")

	if err := writeDeveloperArtifactScaffold(root, artifactProfileAuto, true); err != nil {
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
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
