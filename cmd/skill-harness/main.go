package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type dependencyConfig struct {
	Repos  map[string]repoConfig  `json:"repos"`
	Agents map[string]agentConfig `json:"agents"`
}

type repoConfig struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

type agentConfig struct {
	Repos []string `json:"repos"`
}

type loadoutConfig map[string]struct {
	Skills []string `json:"skills"`
}

type selection struct {
	All   bool
	Agent []string
	Repo  []string
}

type agentStackConfig struct {
	Version        int                            `json:"version"`
	Baseline       agentStackBaseline             `json:"baseline"`
	Profile        string                         `json:"profile"`
	EnabledAgents  []string                       `json:"enabledAgents,omitempty"`
	DisabledAgents []string                       `json:"disabledAgents,omitempty"`
	EnabledPacks   []string                       `json:"enabledPacks,omitempty"`
	DisabledPacks  []string                       `json:"disabledPacks,omitempty"`
	Agents         map[string]agentStackAgentRule `json:"agents,omitempty"`
	RepoLocalPacks []string                       `json:"repoLocalPacks,omitempty"`
	Policies       map[string]string              `json:"policies,omitempty"`
}

type agentStackBaseline struct {
	Source  string `json:"source"`
	Channel string `json:"channel"`
	Pin     string `json:"pin,omitempty"`
}

type agentStackAgentRule struct {
	AddSkills     []string          `json:"addSkills,omitempty"`
	RemoveSkills  []string          `json:"removeSkills,omitempty"`
	ReplaceSkills map[string]string `json:"replaceSkills,omitempty"`
}

type agentStackResolution struct {
	Version         int                            `json:"version"`
	Profile         string                         `json:"profile"`
	State           string                         `json:"state"`
	Baseline        agentStackBaseline             `json:"baseline"`
	EffectiveAgents []string                       `json:"effectiveAgents"`
	EffectivePacks  []string                       `json:"effectivePacks"`
	RepoLocalPacks  []string                       `json:"repoLocalPacks,omitempty"`
	AgentSkills     map[string][]string            `json:"agentSkills"`
	OptOuts         agentStackOptOuts              `json:"optOuts"`
	Overlays        map[string]agentStackAgentRule `json:"overlays,omitempty"`
	Diagnostics     []agentStackDiagnostic         `json:"diagnostics,omitempty"`
}

type agentStackOptOuts struct {
	DisabledAgents []string `json:"disabledAgents,omitempty"`
	DisabledPacks  []string `json:"disabledPacks,omitempty"`
}

type agentStackDiagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type agentStackLock struct {
	Version         int                     `json:"version"`
	Baseline        agentStackBaseline      `json:"baseline"`
	Profile         string                  `json:"profile"`
	ResolvedAt      string                  `json:"resolvedAt"`
	OverlayHash     string                  `json:"overlayHash,omitempty"`
	EffectiveAgents []string                `json:"effectiveAgents"`
	EffectivePacks  []string                `json:"effectivePacks"`
	RepoLocalPacks  []string                `json:"repoLocalPacks,omitempty"`
	AgentSkills     map[string][]string     `json:"agentSkills"`
	OptOuts         agentStackOptOuts       `json:"optOuts"`
	Diagnostics     []agentStackDiagnostic  `json:"diagnostics,omitempty"`
	Surfaces        []agentStackSurfaceLock `json:"surfaces"`
}

type agentStackSurfaceLock struct {
	Path        string `json:"path"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	Kind        string `json:"kind,omitempty"`
	ContentHash string `json:"contentHash,omitempty"`
}

type projectScope string

const (
	projectScopeAuto      projectScope = "auto"
	projectScopeRoot      projectScope = "root"
	projectScopeWorkspace projectScope = "workspace"
)

type packageManager string

const (
	packageManagerAuto packageManager = "auto"
	packageManagerNpm  packageManager = "npm"
	packageManagerPnpm packageManager = "pnpm"
	packageManagerYarn packageManager = "yarn"
	packageManagerBun  packageManager = "bun"
)

type artifactProfile string

const (
	artifactProfileAuto          artifactProfile = "auto"
	artifactProfileMarkdown      artifactProfile = "markdown"
	artifactProfileHTML          artifactProfile = "html"
	artifactProfileDual          artifactProfile = "dual"
	artifactProfileCodexApp      artifactProfile = "codex-app"
	artifactProfileClaudeDesktop artifactProfile = "claude-desktop"
	artifactProfileCLI           artifactProfile = "cli"
	artifactProfileTUI           artifactProfile = "tui"
	artifactProfileMedia         artifactProfile = "media"
	artifactProfileAgentLoop     artifactProfile = "agent-loop"
	artifactProfileNone          artifactProfile = "none"
)

type modelingMode string

const (
	modelingModeAuto     modelingMode = "auto"
	modelingModeOff      modelingMode = "off"
	modelingModeBaseline modelingMode = "baseline"
	modelingModeUMLFirst modelingMode = "uml-first"
)

type projectSetupContext struct {
	TargetDir      string
	OperationDir   string
	MonorepoRoot   string
	Monorepo       bool
	Scope          projectScope
	PackageManager packageManager
}

type projectSetupProof struct {
	Version        int                       `json:"version"`
	Project        projectSetupProofProject  `json:"project"`
	Profiles       projectSetupProofProfiles `json:"profiles"`
	Tools          map[string]toolProof      `json:"tools"`
	Checks         map[string]checkProof     `json:"checks"`
	GeneratedPaths []string                  `json:"generatedPaths"`
	Skipped        []string                  `json:"skipped"`
}

type projectSetupProofProject struct {
	TargetDir      string         `json:"targetDir"`
	OperationDir   string         `json:"operationDir"`
	MonorepoRoot   string         `json:"monorepoRoot,omitempty"`
	Monorepo       bool           `json:"monorepo"`
	Scope          projectScope   `json:"scope"`
	PackageManager packageManager `json:"packageManager"`
}

type projectSetupProofProfiles struct {
	RequestedDeveloperArtifacts artifactProfile `json:"requestedDeveloperArtifacts"`
	EffectiveDeveloperArtifacts artifactProfile `json:"effectiveDeveloperArtifacts"`
	RequestedModeling           modelingMode    `json:"requestedModeling"`
	EffectiveModeling           modelingMode    `json:"effectiveModeling"`
}

type toolProof struct {
	Status  string   `json:"status"`
	Mode    string   `json:"mode,omitempty"`
	Package string   `json:"package,omitempty"`
	Command string   `json:"command,omitempty"`
	Paths   []string `json:"paths,omitempty"`
}

type checkProof struct {
	Status  string `json:"status"`
	Command string `json:"command,omitempty"`
	Path    string `json:"path,omitempty"`
}

type projectSetupProofOptions struct {
	RequestedArtifactProfile artifactProfile
	EffectiveArtifactProfile artifactProfile
	RequestedModelingMode    modelingMode
	EffectiveModelingMode    modelingMode
	InstallOnly              bool
	SkipNoslop               bool
	SkipAgentDocs            bool
	SkipBeads                bool
	SkipClaudeSettings       bool
	SkipArtifacts            bool
	SkipDeveloperArtifacts   bool
	BeadsWorktrees           bool
	BeadsMode                beadsInstallMode
}

type beadsInstallMode string

const (
	beadsDisabled beadsInstallMode = "disabled"
	beadsSystem   beadsInstallMode = "system"
)

const (
	repoSurfaceGenerated      = "generated"
	repoSurfaceManagedSection = "managed-section"
	repoSurfaceOverlay        = "overlay"
	repoSurfaceOwned          = "owned"
	repoSurfaceIgnored        = "ignored"
)

type repoBaselineManifest struct {
	Version  int                     `json:"version"`
	Baseline repoBaselineSource      `json:"baseline"`
	Profile  string                  `json:"profile"`
	Surfaces map[string]surfaceOwner `json:"surfaces,omitempty"`
	Agents   repoSelectionConfig     `json:"agents,omitempty"`
	Packs    repoSelectionConfig     `json:"packs,omitempty"`
	Policies repoPolicyConfig        `json:"policies,omitempty"`
}

type repoBaselineSource struct {
	Source  string `json:"source"`
	Channel string `json:"channel"`
	Pin     string `json:"pin,omitempty"`
}

type surfaceOwner struct {
	Mode string `json:"mode"`
}

type repoSelectionConfig struct {
	Enabled  []string `json:"enabled,omitempty"`
	Disabled []string `json:"disabled,omitempty"`
}

type repoPolicyConfig struct {
	Beads           string `json:"beads,omitempty"`
	Closeout        string `json:"closeout,omitempty"`
	GlobalWrites    string `json:"globalWrites,omitempty"`
	PackageInstalls string `json:"packageInstalls,omitempty"`
	HookChanges     string `json:"hookChanges,omitempty"`
}

type repoAuditReport struct {
	Version         int                 `json:"version"`
	State           string              `json:"state"`
	ProjectDir      string              `json:"projectDir"`
	ManifestPath    string              `json:"manifestPath,omitempty"`
	LockPath        string              `json:"lockPath,omitempty"`
	ReportPath      string              `json:"reportPath,omitempty"`
	Profile         string              `json:"profile"`
	Baseline        repoBaselineSource  `json:"baseline"`
	Surfaces        []repoSurfaceReport `json:"surfaces"`
	LocalSkills     map[string][]string `json:"localSkills,omitempty"`
	LocalAgents     map[string][]string `json:"localAgents,omitempty"`
	EffectiveAgents []string            `json:"effectiveAgents"`
	EffectivePacks  []string            `json:"effectivePacks"`
	Findings        []repoFinding       `json:"findings"`
	Suggestions     []string            `json:"suggestions"`
	Policies        repoPolicyConfig    `json:"policies,omitempty"`
}

type repoSurfaceReport struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Status string `json:"status"`
	Kind   string `json:"kind,omitempty"`
	Count  int    `json:"count,omitempty"`
}

type repoFinding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}

type repoBaselineLock struct {
	Version         int                `json:"version"`
	Baseline        repoBaselineSource `json:"baseline"`
	Profile         string             `json:"profile"`
	ResolvedAt      string             `json:"resolvedAt"`
	EffectiveAgents []string           `json:"effectiveAgents"`
	EffectivePacks  []string           `json:"effectivePacks"`
	Surfaces        []repoSurfaceLock  `json:"surfaces"`
	Findings        []repoFinding      `json:"findings,omitempty"`
}

type repoSurfaceLock struct {
	Path        string `json:"path"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	ContentHash string `json:"contentHash,omitempty"`
}

func main() {
	root, err := findRepoRoot()
	exitOnErr(err)

	deps := loadDependencies(root)
	loadouts := loadLoadouts(root)

	if len(os.Args) < 2 {
		printUsage(loadouts, deps)
		return
	}

	switch os.Args[1] {
	case "list":
		runList(root, deps, loadouts, os.Args[2:])
	case "resolve":
		runResolve(root, deps, loadouts, os.Args[2:])
	case "audit-project":
		runAuditProject(root, deps, loadouts, os.Args[2:])
	case "bootstrap":
		runBootstrap(root, deps, loadouts, os.Args[2:])
	case "update-project":
		runUpdateProject(root, deps, loadouts, os.Args[2:])
	case "install":
		runInstall(root, deps, loadouts, os.Args[2:])
	case "setup-project":
		runSetupProject(root, os.Args[2:])
	case "repo":
		runRepo(root, deps, loadouts, os.Args[2:])
	case "update":
		runUpdate(root)
	case "check":
		runCheck(root, loadouts, os.Args[2:])
	case "render":
		runRender(root, loadouts, os.Args[2:])
	case "beads-worktrees":
		runBeadsWorktrees(root, os.Args[2:])
	case "uninstall":
		runUninstall(root, loadouts, os.Args[2:])
	case "help", "-h", "--help":
		printUsage(loadouts, deps)
	default:
		exitOnErr(fmt.Errorf("unknown command: %s", os.Args[1]))
	}
}

func runList(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	showAgents := fs.Bool("agents", false, "List agents.")
	showPacks := fs.Bool("packs", false, "List packs.")
	fs.Parse(args)

	if !*showAgents && !*showPacks {
		*showAgents = true
		*showPacks = true
	}

	if *showAgents {
		fmt.Println("Agents:")
		for _, name := range sortedKeys(loadouts) {
			fmt.Printf("- %s\n", name)
		}
	}
	if *showPacks {
		if *showAgents {
			fmt.Println()
		}
		fmt.Println("Packs:")
		for _, name := range sortedKeys(deps.Repos) {
			fmt.Printf("- %s\n", name)
		}
	}
}

func runResolve(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("resolve", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	strict := fs.Bool("strict", false, "Return an error when resolution diagnostics contain errors.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	resolution, err := resolveAgentStack(projectDir, deps, loadouts)
	exitOnErr(err)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, resolution))
	} else {
		printAgentStackResolution(resolution)
	}
	if *strict && agentStackHasErrors(resolution.Diagnostics) {
		exitOnErr(errors.New("agent stack resolution has errors"))
	}
}

func runAuditProject(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("audit-project", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	result := auditAgentStackProject(projectDir, deps, loadouts)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, result))
		return
	}
	fmt.Printf("State: %s\n", result.State)
	for _, reason := range result.Reasons {
		fmt.Printf("- %s\n", reason)
	}
	if len(result.Diagnostics) > 0 {
		fmt.Println("Diagnostics:")
		for _, diagnostic := range result.Diagnostics {
			fmt.Printf("- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
		}
	}
}

func runRepoLock(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo lock", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	lock := buildRepoBaselineLock(projectDir, report)
	lockPath := repoLockPath(projectDir)
	exitOnErr(writeJSONFile(lockPath, lock))
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, lock))
		return
	}
	fmt.Printf("Wrote %s\n", lockPath)
}

func runBootstrap(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	agentNative := fs.Bool("agent-native", false, "Scaffold agent-native overlay state.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	if !*agentNative {
		exitOnErr(errors.New("bootstrap currently requires --agent-native"))
	}
	exitOnErr(writeDefaultAgentStack(projectDir, false))
	resolution, lock, err := writeResolvedAgentStackLock(projectDir, deps, loadouts)
	exitOnErr(err)
	ctx, err := resolveProjectSetupContext(projectDir, string(projectScopeAuto), string(packageManagerAuto))
	exitOnErr(err)
	exitOnErr(writeAgentNativeSetupProof(ctx, resolution, lock))
	result := map[string]any{
		"state":      resolution.State,
		"resolution": resolution,
		"lockPath":   agentStackLockPath(projectDir),
		"proofPath":  filepath.Join(projectDir, ".skill-harness", "setup-proof.json"),
	}
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, result))
		return
	}
	fmt.Printf("Agent-native bootstrap scaffolded for %s\n", projectDir)
	fmt.Printf("State: %s\n", resolution.State)
	fmt.Printf("Wrote %s\n", agentStackLockPath(projectDir))
	fmt.Printf("Wrote %s\n", filepath.Join(projectDir, ".skill-harness", "setup-proof.json"))
	fmt.Println("Next: review the resolved effective agents and packs before package installs or global writes.")
}

func runUpdateProject(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("update-project", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	writeLock := fs.Bool("write-lock", false, "Write .skill-harness/agent-stack.lock.json after resolution.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	exitOnErr(requireAgentStack(projectDir))
	resolution, err := resolveAgentStack(projectDir, deps, loadouts)
	exitOnErr(err)
	report := map[string]any{
		"state":      resolution.State,
		"resolution": resolution,
		"lockPath":   agentStackLockPath(projectDir),
		"dryRun":     !*writeLock,
	}
	if *writeLock {
		if err := errorOnAgentStackDiagnostics(resolution.Diagnostics); err != nil {
			exitOnErr(err)
		}
		lock := buildAgentStackLock(projectDir, resolution)
		exitOnErr(writeJSONFile(agentStackLockPath(projectDir), lock))
		report["lock"] = lock
	}
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	fmt.Printf("State: %s\n", resolution.State)
	fmt.Printf("Profile: %s\n", resolution.Profile)
	for _, diagnostic := range resolution.Diagnostics {
		fmt.Printf("- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
	}
	if !*writeLock {
		fmt.Println("Dry run only. Pass --write-lock to refresh .skill-harness/agent-stack.lock.json.")
	} else {
		fmt.Printf("Wrote %s\n", agentStackLockPath(projectDir))
	}
}

func runInstall(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	all := fs.Bool("all", false, "Install all packs and all agents.")
	interactive := fs.Bool("interactive", false, "Pick packs and agents interactively.")
	packsOnly := fs.Bool("packs-only", false, "Install only dependency repos.")
	agentsOnly := fs.Bool("agents-only", false, "Install only agent files without bootstrapping dependency repos.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	packsCSV := fs.String("packs", "", "Comma-separated dependency repo names.")
	targetDir := fs.String("dir", "", "Target project directory with .skill-harness/agent-stack.json.")
	fs.Parse(args)

	if *packsOnly && *agentsOnly {
		exitOnErr(errors.New("cannot combine --packs-only and --agents-only"))
	}

	sel := selection{
		All:   *all,
		Agent: csvList(*agentsCSV),
		Repo:  csvList(*packsCSV),
	}

	if *interactive {
		sel = promptInstallSelection(loadouts, deps)
	}
	resolvedFromStack := false
	stackAgents := []string{}
	stackRepos := []string{}
	if *targetDir != "" && !*all && len(sel.Agent) == 0 && len(sel.Repo) == 0 && !*interactive {
		projectDir, err := filepath.Abs(*targetDir)
		exitOnErr(err)
		exitOnErr(requireAgentStack(projectDir))
		resolution, err := resolveAgentStack(projectDir, deps, loadouts)
		exitOnErr(err)
		exitOnErr(errorOnAgentStackDiagnostics(resolution.Diagnostics))
		resolvedFromStack = true
		stackAgents = resolution.EffectiveAgents
		stackRepos = resolution.EffectivePacks
	}

	agents := resolveAgents(sel, loadouts)
	repos := resolveRepos(sel, deps)
	if resolvedFromStack {
		agents = stackAgents
		repos = stackRepos
	}

	if !*agentsOnly {
		if resolvedFromStack {
			exitOnErr(runPython(root, "scripts/bootstrap_dependencies.py", repoArgs(repos)...))
		} else {
			exitOnErr(runPython(root, "scripts/bootstrap_dependencies.py", bootstrapArgs(sel)...))
		}
	}

	if !*packsOnly {
		if len(agents) == 0 {
			agents = sortedKeys(loadouts)
		}
		exitOnErr(runPython(root, "scripts/render_claude_agents.py", agentArgs(agents)...))
		exitOnErr(runPython(root, "scripts/render_codex_agents.py", agentArgs(agents)...))
		exitOnErr(runPython(root, "scripts/check_dependencies.py", agentArgs(agents)...))
	}

	fmt.Printf("Installed %d dependency repo(s) and %d agent(s)\n", len(repos), len(agents))
}

func runSetupProject(root string, args []string) {
	fs := flag.NewFlagSet("setup-project", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	skipNoslop := fs.Bool("skip-noslop", false, "Do not install @45ck/noslop.")
	skipAgentDocs := fs.Bool("skip-agent-docs", false, "Do not install 45ck/agent-docs.")
	skipBeads := fs.Bool("skip-beads", false, "Do not install or initialize Beads.")
	beadsWorktrees := fs.Bool("beads-worktrees", true, "Install repo-local Beads worktree wrapper script.")
	skipClaudeSettings := fs.Bool("skip-claude-settings", false, "Do not write .claude/settings.json.")
	skipArtifacts := fs.Bool("skip-artifacts", false, "Do not scaffold developer artifact guidance.")
	skipDeveloperArtifacts := fs.Bool("skip-developer-artifacts", false, "Do not scaffold developer artifact guidance.")
	enableModeling := fs.Bool("enable-modeling", false, "Scaffold source-first UML/UWE/C4 model artifacts and model review checks.")
	skipModeling := fs.Bool("skip-modeling", false, "Keep developer artifacts but skip UML-first model scaffolding.")
	modelingModeValue := fs.String("modeling-mode", string(modelingModeAuto), "Developer artifact modeling mode: auto, off, baseline, or uml-first.")
	artifactProfileValue := fs.String("artifact-profile", string(artifactProfileAuto), "Developer artifact profile: auto, media, agent-loop, markdown, html, dual, or none.")
	developerArtifactsProfileValue := fs.String("developer-artifacts-profile", string(artifactProfileAuto), "Developer artifact profile: auto, codex-app, claude-desktop, cli, tui, media, agent-loop, markdown, html, dual, or none.")
	installOnly := fs.Bool("install-only", false, "Install packages only; skip initialization commands.")
	scopeValue := fs.String("scope", string(projectScopeAuto), "Setup scope: auto, root, or workspace.")
	packageManagerValue := fs.String("package-manager", string(packageManagerAuto), "Package manager: auto, npm, pnpm, yarn, or bun.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)

	ctx, err := resolveProjectSetupContext(projectDir, *scopeValue, *packageManagerValue)
	exitOnErr(err)
	exitOnErr(requireSetupCommands(ctx.PackageManager))
	profileValue := *artifactProfileValue
	if flagWasSet(fs, "developer-artifacts-profile") {
		profileValue = *developerArtifactsProfileValue
	}
	artifactProfile, err := parseArtifactProfile(profileValue)
	exitOnErr(err)
	requestedModelingMode, err := parseModelingMode(*modelingModeValue)
	exitOnErr(err)
	if *enableModeling && *skipModeling {
		exitOnErr(errors.New("cannot combine --enable-modeling and --skip-modeling"))
	}
	if *enableModeling && flagWasSet(fs, "modeling-mode") && requestedModelingMode != modelingModeAuto && requestedModelingMode != modelingModeUMLFirst {
		exitOnErr(errors.New("--enable-modeling cannot be combined with --modeling-mode off or baseline"))
	}
	if *skipModeling && flagWasSet(fs, "modeling-mode") && requestedModelingMode != modelingModeAuto && requestedModelingMode != modelingModeOff {
		exitOnErr(errors.New("--skip-modeling cannot be combined with --modeling-mode baseline or uml-first"))
	}
	if *enableModeling {
		requestedModelingMode = modelingModeUMLFirst
	}
	if *skipModeling {
		requestedModelingMode = modelingModeOff
	}
	effectiveModelingMode := resolveEffectiveModelingMode(ctx.OperationDir, requestedModelingMode, artifactProfile, *skipArtifacts || *skipDeveloperArtifacts)

	if _, err := os.Stat(filepath.Join(ctx.OperationDir, "package.json")); errors.Is(err, os.ErrNotExist) {
		exitOnErr(writeMinimalPackageJSON(ctx.OperationDir))
	}

	packages := []string{}
	if !*skipNoslop {
		packages = append(packages, "@45ck/noslop")
	}
	if !*skipAgentDocs {
		packages = append(packages, "github:45ck/agent-docs")
	}
	if len(packages) > 0 {
		exitOnErr(installPackages(ctx.OperationDir, ctx.PackageManager, packages...))
	}

	beadsMode := beadsDisabled
	if !*skipBeads {
		mode, err := installBeads(ctx.OperationDir)
		exitOnErr(err)
		beadsMode = mode
	}

	if !*skipClaudeSettings {
		exitOnErr(allowAgentTeams(projectDir))
	}
	exitOnErr(writeDefaultAgentStack(ctx.OperationDir, false))

	proofOptions := projectSetupProofOptions{
		RequestedArtifactProfile: artifactProfile,
		EffectiveArtifactProfile: effectiveArtifactProfile(artifactProfile),
		RequestedModelingMode:    requestedModelingMode,
		EffectiveModelingMode:    effectiveModelingMode,
		InstallOnly:              *installOnly,
		SkipNoslop:               *skipNoslop,
		SkipAgentDocs:            *skipAgentDocs,
		SkipBeads:                *skipBeads,
		SkipClaudeSettings:       *skipClaudeSettings,
		SkipArtifacts:            *skipArtifacts,
		SkipDeveloperArtifacts:   *skipDeveloperArtifacts,
		BeadsWorktrees:           *beadsWorktrees,
		BeadsMode:                beadsMode,
	}

	if *installOnly {
		exitOnErr(writeProjectSetupProof(ctx.OperationDir, buildProjectSetupProof(ctx, proofOptions)))
		fmt.Println(projectSetupSummary("Installed project tooling", ctx))
		return
	}

	if !*skipAgentDocs {
		exitOnErr(runLocalTool(ctx.OperationDir, ctx.PackageManager, "agent-docs", "init"))
	}
	if !*skipArtifacts && !*skipDeveloperArtifacts && artifactProfile != artifactProfileNone {
		exitOnErr(writeDeveloperArtifactScaffold(ctx.OperationDir, artifactProfile, !*skipAgentDocs, !*skipBeads, effectiveModelingMode))
	}
	if !*skipNoslop {
		exitOnErr(runLocalTool(ctx.OperationDir, ctx.PackageManager, "noslop", "init"))
	}
	if beadsMode != beadsDisabled {
		exitOnErr(initBeads(ctx.OperationDir, beadsMode))
		if *beadsWorktrees {
			exitOnErr(installBeadsWorktreeWrapper(root, ctx.OperationDir, false))
		}
	}
	if !*skipAgentDocs {
		exitOnErr(runLocalTool(ctx.OperationDir, ctx.PackageManager, "agent-docs", "install-gates", "--quality"))
	}

	exitOnErr(writeProjectSetupProof(ctx.OperationDir, buildProjectSetupProof(ctx, proofOptions)))
	fmt.Println(projectSetupSummary("Project setup complete", ctx))
}

func runBeadsWorktrees(root string, args []string) {
	fs := flag.NewFlagSet("beads-worktrees", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory (repo root).")
	force := fs.Bool("force", false, "Overwrite existing scripts/beads/bd.mjs if present.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)

	exitOnErr(installBeadsWorktreeWrapper(root, projectDir, *force))
	fmt.Printf("Installed Beads worktree wrapper in %s\n", projectDir)
}

// allowAgentTeams writes or updates .claude/settings.json to permit the Agent
// tool without per-use prompts, enabling agent team workflows by default.
func allowAgentTeams(projectDir string) error {
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")

	existing := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	perms, _ := existing["permissions"].(map[string]any)
	if perms == nil {
		perms = map[string]any{}
	}
	allow, _ := perms["allow"].([]any)
	hasAgent := false
	for _, v := range allow {
		if s, ok := v.(string); ok && s == "Agent" {
			hasAgent = true
			break
		}
	}
	if !hasAgent {
		allow = append(allow, "Agent")
	}
	perms["allow"] = allow
	existing["permissions"] = perms

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(settingsPath, data, 0o644)
}

func installBeadsWorktreeWrapper(harnessRoot string, projectDir string, force bool) error {
	sourcePath := filepath.Join(harnessRoot, "scripts", "templates", "beads-worktrees", "bd.mjs")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read beads-worktrees template: %w", err)
	}

	destDir := filepath.Join(projectDir, "scripts", "beads")
	destPath := filepath.Join(destDir, "bd.mjs")
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return ensureGitignoreHasTrees(projectDir)
		}
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(destPath, source, 0o755); err != nil {
		return err
	}

	return ensureGitignoreHasTrees(projectDir)
}

func ensureGitignoreHasTrees(projectDir string) error {
	ignorePath := filepath.Join(projectDir, ".gitignore")
	data, err := os.ReadFile(ignorePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	content := string(data)
	if strings.Contains(content, "\n.trees/\n") || strings.HasSuffix(content, "\n.trees/") || strings.Contains(content, "\n.trees/\r\n") {
		return nil
	}

	trimmed := strings.TrimRight(content, "\r\n")
	lines := []string{}
	if trimmed != "" {
		lines = append(lines, trimmed)
	}
	lines = append(lines, "", "# Beads worktrees", ".trees/")
	out := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(ignorePath, []byte(out), 0o644)
}

func runCheck(root string, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	all := fs.Bool("all", false, "Check all agents.")
	interactive := fs.Bool("interactive", false, "Choose agents interactively.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	targetDir := fs.String("dir", "", "Target project directory with .skill-harness/agent-stack.json.")
	fs.Parse(args)

	sel := selection{All: *all, Agent: csvList(*agentsCSV)}
	if *interactive {
		sel.Agent = promptAgentList("Check which agents?", sortedKeys(loadouts))
	}
	if *targetDir != "" && !*all && len(sel.Agent) == 0 && !*interactive {
		projectDir, err := filepath.Abs(*targetDir)
		exitOnErr(err)
		exitOnErr(requireAgentStack(projectDir))
		deps := loadDependencies(root)
		resolution, err := resolveAgentStack(projectDir, deps, loadouts)
		exitOnErr(err)
		exitOnErr(errorOnAgentStackDiagnostics(resolution.Diagnostics))
		sel.Agent = resolution.EffectiveAgents
	}
	agents := resolveAgents(sel, loadouts)
	exitOnErr(runPython(root, "scripts/check_dependencies.py", agentArgs(agents)...))
}

func runRender(root string, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	all := fs.Bool("all", false, "Render all agents.")
	interactive := fs.Bool("interactive", false, "Choose agents interactively.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	targetDir := fs.String("dir", "", "Target project directory with .skill-harness/agent-stack.json.")
	fs.Parse(args)

	sel := selection{All: *all, Agent: csvList(*agentsCSV)}
	if *interactive {
		sel.Agent = promptAgentList("Render which agents?", sortedKeys(loadouts))
	}
	if *targetDir != "" && !*all && len(sel.Agent) == 0 && !*interactive {
		projectDir, err := filepath.Abs(*targetDir)
		exitOnErr(err)
		exitOnErr(requireAgentStack(projectDir))
		deps := loadDependencies(root)
		resolution, err := resolveAgentStack(projectDir, deps, loadouts)
		exitOnErr(err)
		exitOnErr(errorOnAgentStackDiagnostics(resolution.Diagnostics))
		sel.Agent = resolution.EffectiveAgents
	}
	agents := resolveAgents(sel, loadouts)
	exitOnErr(runPython(root, "scripts/render_codex_agents.py", agentArgs(agents)...))
}

func runUpdate(root string) {
	exitOnErr(runCommand(root, "git", "pull", "--ff-only"))
	fmt.Println("skill-harness updated. Rebuild the binary to apply changes:")
	fmt.Println("  go build ./cmd/skill-harness/")
}

func runUninstall(root string, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	all := fs.Bool("all", false, "Uninstall all agents.")
	interactive := fs.Bool("interactive", false, "Choose agents interactively.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	fs.Parse(args)

	sel := selection{All: *all, Agent: csvList(*agentsCSV)}
	if *interactive {
		sel.Agent = promptAgentList("Uninstall which agents?", sortedKeys(loadouts))
	}
	agents := resolveAgents(sel, loadouts)
	if len(agents) == 0 {
		agents = sortedKeys(loadouts)
	}
	home, err := os.UserHomeDir()
	exitOnErr(err)
	for _, agent := range agents {
		_ = os.Remove(filepath.Join(home, ".claude", "agents", agent+".md"))
		_ = os.Remove(filepath.Join(home, ".codex", "agents", agent+".toml"))
	}
	fmt.Printf("Removed %d agent(s)\n", len(agents))
}

func runRepo(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	if len(args) == 0 {
		printRepoUsage()
		return
	}
	switch args[0] {
	case "init":
		runRepoInit(root, deps, loadouts, args[1:])
	case "audit":
		runRepoAudit(root, deps, loadouts, args[1:])
	case "drift":
		runRepoDrift(root, deps, loadouts, args[1:])
	case "update":
		runRepoUpdate(root, deps, loadouts, args[1:])
	case "trim":
		runRepoTrim(root, deps, loadouts, args[1:])
	case "sync":
		runRepoSync(root, deps, loadouts, args[1:])
	case "lock":
		runRepoLock(root, deps, loadouts, args[1:])
	case "help", "-h", "--help":
		printRepoUsage()
	default:
		exitOnErr(fmt.Errorf("unknown repo command: %s", args[0]))
	}
}

func runRepoInit(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo init", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	profile := fs.String("profile", "team", "Baseline profile: minimal, team, or agent-native.")
	force := fs.Bool("force", false, "Overwrite an existing baseline manifest.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	if err := validateRepoProfile(*profile); err != nil {
		exitOnErr(err)
	}
	manifestPath := repoManifestPath(projectDir)
	if fileExists(manifestPath) && !*force {
		exitOnErr(fmt.Errorf("%s already exists; use --force to replace it", manifestPath))
	}
	manifest := defaultRepoBaselineManifest(root, *profile, projectDir)
	exitOnErr(validateRepoManifest(manifest, deps, loadouts))
	exitOnErr(writeJSONFile(manifestPath, manifest))
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	fmt.Printf("Wrote %s\n", manifestPath)
	printRepoReport(report)
}

func runRepoAudit(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo audit", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	printRepoReport(report)
}

func runRepoDrift(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo drift", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
	} else {
		printFindings(report.Findings)
	}
	if repoHasBlockingFindings(report.Findings) {
		exitOnErr(errors.New("repo drift found warnings or errors"))
	}
}

func runRepoUpdate(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo update", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	checkOnly := fs.Bool("check", false, "Check available baseline changes without writing project surfaces.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	if !*checkOnly {
		exitOnErr(errors.New("repo update currently requires --check; use repo sync to refresh lock/report files"))
	}
	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	printSuggestions(report.Suggestions)
}

func runRepoTrim(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo trim", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	dryRun := fs.Bool("dry-run", false, "Show cost-reduction candidates without changing the repo.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	if !*dryRun {
		exitOnErr(errors.New("repo trim currently requires --dry-run"))
	}
	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	printSuggestions(report.Suggestions)
}

func runRepoSync(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("repo sync", flag.ExitOnError)
	targetDir := fs.String("dir", ".", "Target project directory.")
	jsonOutput := fs.Bool("json", false, "Print JSON output.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)
	if !fileExists(repoManifestPath(projectDir)) {
		exitOnErr(fmt.Errorf("missing %s; run repo init first", repoManifestPath(projectDir)))
	}
	report := buildRepoAuditReport(root, deps, loadouts, projectDir)
	lock := buildRepoBaselineLock(projectDir, report)
	exitOnErr(writeJSONFile(repoLockPath(projectDir), lock))
	exitOnErr(writeJSONFile(repoUpdateReportPath(projectDir), report))
	if *jsonOutput {
		exitOnErr(writeJSON(os.Stdout, report))
		return
	}
	fmt.Printf("Wrote %s\n", repoLockPath(projectDir))
	fmt.Printf("Wrote %s\n", repoUpdateReportPath(projectDir))
	printRepoReport(report)
}

func buildRepoAuditReport(root string, deps dependencyConfig, loadouts loadoutConfig, projectDir string) repoAuditReport {
	manifest, manifestExists, manifestErr := readRepoBaselineManifest(projectDir)
	findings := []repoFinding{}
	if manifestErr != nil {
		findings = append(findings, repoFinding{
			Severity: "error",
			Code:     "manifest-invalid",
			Message:  manifestErr.Error(),
			Path:     repoManifestPath(projectDir),
		})
		manifest = inferredRepoManifest(root, projectDir)
	} else if !manifestExists {
		manifest = inferredRepoManifest(root, projectDir)
		findings = append(findings, repoFinding{
			Severity: "info",
			Code:     "manifest-missing",
			Message:  "repo is not pinned to a skill-harness baseline manifest",
			Path:     repoManifestPath(projectDir),
		})
	}
	if err := validateRepoManifest(manifest, deps, loadouts); err != nil {
		findings = append(findings, repoFinding{
			Severity: "error",
			Code:     "manifest-validation",
			Message:  err.Error(),
			Path:     repoManifestPath(projectDir),
		})
	}
	surfaces := buildRepoSurfaceReports(projectDir, manifest)
	for _, finding := range repoSurfaceFindings(surfaces) {
		findings = append(findings, finding)
	}
	localSkills := discoverLocalSkills(projectDir)
	localAgents := discoverLocalAgents(projectDir)
	if !manifestExists && len(localSkills) > 0 {
		findings = append(findings, repoFinding{
			Severity: "info",
			Code:     "local-skills-unmanaged",
			Message:  "repo-local skills were found; repo init will classify them as overlays rather than replacing them",
		})
	}
	effectiveAgents := resolveRepoEffectiveAgents(manifest, loadouts)
	effectivePacks := resolveRepoEffectivePacks(manifest, effectiveAgents, deps)
	report := repoAuditReport{
		Version:         1,
		State:           repoState(manifestExists, surfaces),
		ProjectDir:      projectDir,
		ManifestPath:    repoPathIfExists(repoManifestPath(projectDir)),
		LockPath:        repoPathIfExists(repoLockPath(projectDir)),
		ReportPath:      repoPathIfExists(repoUpdateReportPath(projectDir)),
		Profile:         manifest.Profile,
		Baseline:        manifest.Baseline,
		Surfaces:        surfaces,
		LocalSkills:     localSkills,
		LocalAgents:     localAgents,
		EffectiveAgents: effectiveAgents,
		EffectivePacks:  effectivePacks,
		Findings:        findings,
		Suggestions:     repoSuggestions(root, projectDir, manifest, surfaces, effectiveAgents, effectivePacks),
		Policies:        manifest.Policies,
	}
	return report
}

func defaultRepoBaselineManifest(root, profile, projectDir string) repoBaselineManifest {
	manifest := repoBaselineManifest{
		Version: 1,
		Baseline: repoBaselineSource{
			Source:  "skill-harness",
			Channel: "default",
			Pin:     currentGitRevision(root),
		},
		Profile: profile,
		Surfaces: map[string]surfaceOwner{
			"AGENTS.md":                              {Mode: repoSurfaceManagedSection},
			"CLAUDE.md":                              {Mode: repoSurfaceManagedSection},
			"AGENT_INSTRUCTIONS.md":                  {Mode: repoSurfaceManagedSection},
			"llms.txt":                               {Mode: repoSurfaceGenerated},
			".skill-harness/agent-stack.json":        {Mode: repoSurfaceOverlay},
			".skill-harness/setup-proof.json":        {Mode: repoSurfaceGenerated},
			".agent-docs":                            {Mode: repoSurfaceGenerated},
			".beads":                                 {Mode: repoSurfaceOwned},
			"scripts/beads":                          {Mode: repoSurfaceGenerated},
			".claude/agents":                         {Mode: repoSurfaceGenerated},
			".codex/agents":                          {Mode: repoSurfaceGenerated},
			".claude/skills":                         {Mode: repoSurfaceOverlay},
			".codex/skills":                          {Mode: repoSurfaceOverlay},
			".github/skills":                         {Mode: repoSurfaceOverlay},
			"packs":                                  {Mode: repoSurfaceOverlay},
			"docs/artifacts/source/models":           {Mode: repoSurfaceOverlay},
			"docs/artifacts/artifacts.manifest.json": {Mode: repoSurfaceManagedSection},
			"generated/review":                       {Mode: repoSurfaceGenerated},
		},
		Agents: repoSelectionConfig{
			Enabled: repoProfileAgents(profile, nil),
		},
		Policies: repoPolicyConfig{
			Beads:           "preserve",
			Closeout:        "repo-owned",
			GlobalWrites:    "ask",
			PackageInstalls: "ask",
			HookChanges:     "ask",
		},
	}
	for path := range manifest.Surfaces {
		if !pathExists(filepath.Join(projectDir, filepath.FromSlash(path))) {
			manifest.Surfaces[path] = surfaceOwner{Mode: repoSurfaceIgnored}
		}
	}
	return manifest
}

func inferredRepoManifest(root, projectDir string) repoBaselineManifest {
	manifest := defaultRepoBaselineManifest(root, "unmanaged", projectDir)
	manifest.Agents.Enabled = nil
	for path, owner := range manifest.Surfaces {
		if !pathExists(filepath.Join(projectDir, filepath.FromSlash(path))) {
			delete(manifest.Surfaces, path)
			continue
		}
		if owner.Mode == repoSurfaceGenerated || owner.Mode == repoSurfaceManagedSection {
			manifest.Surfaces[path] = surfaceOwner{Mode: repoSurfaceOwned}
		}
	}
	return manifest
}

func readRepoBaselineManifest(projectDir string) (repoBaselineManifest, bool, error) {
	path := repoManifestPath(projectDir)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return repoBaselineManifest{}, false, nil
	}
	if err != nil {
		return repoBaselineManifest{}, false, err
	}
	var manifest repoBaselineManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return repoBaselineManifest{}, true, err
	}
	if manifest.Version == 0 {
		manifest.Version = 1
	}
	if manifest.Profile == "" {
		manifest.Profile = "team"
	}
	if manifest.Baseline.Source == "" {
		manifest.Baseline.Source = "skill-harness"
	}
	if manifest.Baseline.Channel == "" {
		manifest.Baseline.Channel = "default"
	}
	if manifest.Surfaces == nil {
		manifest.Surfaces = map[string]surfaceOwner{}
	}
	return manifest, true, nil
}

func validateRepoManifest(manifest repoBaselineManifest, deps dependencyConfig, loadouts loadoutConfig) error {
	if manifest.Version != 1 {
		return fmt.Errorf("unsupported baseline manifest version %d", manifest.Version)
	}
	for path, owner := range manifest.Surfaces {
		if strings.TrimSpace(path) == "" {
			return errors.New("surface path cannot be empty")
		}
		if strings.Contains(filepath.Clean(filepath.FromSlash(path)), "..") {
			return fmt.Errorf("surface path %q must stay inside the project", path)
		}
		if !validRepoSurfaceMode(owner.Mode) {
			return fmt.Errorf("surface %q has unknown mode %q", path, owner.Mode)
		}
	}
	for _, agent := range manifest.Agents.Enabled {
		if _, ok := loadouts[agent]; !ok {
			return fmt.Errorf("enabled agent %q is not in loadouts.json", agent)
		}
	}
	for _, agent := range manifest.Agents.Disabled {
		if _, ok := loadouts[agent]; !ok {
			return fmt.Errorf("disabled agent %q is not in loadouts.json", agent)
		}
	}
	for _, pack := range append([]string{}, manifest.Packs.Enabled...) {
		if _, ok := deps.Repos[pack]; !ok {
			return fmt.Errorf("enabled pack %q is not in dependencies.json", pack)
		}
	}
	for _, pack := range manifest.Packs.Disabled {
		if _, ok := deps.Repos[pack]; !ok {
			return fmt.Errorf("disabled pack %q is not in dependencies.json", pack)
		}
	}
	return nil
}

func buildRepoSurfaceReports(projectDir string, manifest repoBaselineManifest) []repoSurfaceReport {
	paths := knownRepoSurfacePaths()
	for path := range manifest.Surfaces {
		if !stringSliceContains(paths, path) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	reports := []repoSurfaceReport{}
	for _, path := range paths {
		owner, ok := manifest.Surfaces[path]
		if !ok {
			continue
		}
		if owner.Mode == "" {
			owner.Mode = repoSurfaceOwned
		}
		reports = append(reports, describeRepoSurface(projectDir, path, owner.Mode))
	}
	return reports
}

func describeRepoSurface(projectDir, relPath, mode string) repoSurfaceReport {
	path := filepath.Join(projectDir, filepath.FromSlash(relPath))
	report := repoSurfaceReport{Path: relPath, Mode: mode, Status: "missing"}
	info, err := os.Stat(path)
	if err != nil {
		return report
	}
	report.Status = "present"
	if info.IsDir() {
		report.Kind = "dir"
		report.Count = countImmediateChildren(path)
	} else {
		report.Kind = "file"
		report.Count = 1
	}
	return report
}

func repoSurfaceFindings(surfaces []repoSurfaceReport) []repoFinding {
	findings := []repoFinding{}
	for _, surface := range surfaces {
		if surface.Status == "missing" && surface.Mode != repoSurfaceIgnored {
			findings = append(findings, repoFinding{
				Severity: "warning",
				Code:     "surface-missing",
				Message:  fmt.Sprintf("%s is declared as %s but is missing", surface.Path, surface.Mode),
				Path:     surface.Path,
			})
		}
	}
	return findings
}

func resolveRepoEffectiveAgents(manifest repoBaselineManifest, loadouts loadoutConfig) []string {
	agents := manifest.Agents.Enabled
	if len(agents) == 0 {
		agents = repoProfileAgents(manifest.Profile, loadouts)
	}
	disabled := toSet(manifest.Agents.Disabled)
	out := []string{}
	for _, agent := range agents {
		if disabled[agent] {
			continue
		}
		if _, ok := loadouts[agent]; ok {
			out = append(out, agent)
		}
	}
	return unique(out)
}

func resolveRepoEffectivePacks(manifest repoBaselineManifest, agents []string, deps dependencyConfig) []string {
	packs := []string{}
	for _, agent := range agents {
		packs = append(packs, deps.Agents[agent].Repos...)
	}
	packs = append(packs, manifest.Packs.Enabled...)
	disabled := toSet(manifest.Packs.Disabled)
	out := []string{}
	for _, pack := range packs {
		if disabled[pack] {
			continue
		}
		if _, ok := deps.Repos[pack]; ok {
			out = append(out, pack)
		}
	}
	return unique(out)
}

func repoProfileAgents(profile string, loadouts loadoutConfig) []string {
	var candidates []string
	switch profile {
	case "minimal":
		candidates = []string{"delivery-manager", "quality-reviewer"}
	case "agent-native":
		candidates = []string{"requirements-analyst", "software-architect", "delivery-manager", "quality-reviewer", "workflow-engineer"}
	case "team", "unmanaged", "":
		candidates = []string{"requirements-analyst", "software-architect", "delivery-manager", "quality-reviewer"}
	default:
		candidates = []string{"requirements-analyst", "software-architect", "delivery-manager", "quality-reviewer"}
	}
	if loadouts == nil {
		return candidates
	}
	out := []string{}
	for _, agent := range candidates {
		if _, ok := loadouts[agent]; ok {
			out = append(out, agent)
		}
	}
	return out
}

func validateRepoProfile(profile string) error {
	switch profile {
	case "minimal", "team", "agent-native":
		return nil
	default:
		return fmt.Errorf("unknown repo profile %q", profile)
	}
}

func repoSuggestions(root, projectDir string, manifest repoBaselineManifest, surfaces []repoSurfaceReport, agents, packs []string) []string {
	suggestions := []string{}
	if current := currentGitRevision(root); current != "" && manifest.Baseline.Pin != "" && current != manifest.Baseline.Pin {
		suggestions = append(suggestions, fmt.Sprintf("baseline pin differs from current skill-harness revision: %s -> %s", manifest.Baseline.Pin, current))
	}
	if len(agents) > 4 {
		suggestions = append(suggestions, "consider disabling unused agents in .skill-harness/baseline.manifest.json to reduce context and install cost")
	}
	if len(packs) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("effective pack set has %d pack(s); repo trim --dry-run can identify packs to opt out of after workflow review", len(packs)))
	}
	for _, surface := range surfaces {
		if surface.Mode == repoSurfaceOverlay && surface.Status == "present" {
			suggestions = append(suggestions, fmt.Sprintf("preserve %s as a repo-local overlay during baseline updates", surface.Path))
		}
	}
	if remote := gitRemoteURL(root); remote != "" {
		suggestions = append(suggestions, fmt.Sprintf("baseline source can be reviewed from %s", remote))
	}
	return unique(suggestions)
}

func repoState(manifestExists bool, surfaces []repoSurfaceReport) string {
	if manifestExists {
		return "managed"
	}
	for _, surface := range surfaces {
		if surface.Status == "present" {
			return "unmanaged"
		}
	}
	return "empty"
}

func discoverLocalSkills(projectDir string) map[string][]string {
	roots := []string{".claude/skills", ".codex/skills", ".github/skills", "agent/workspace/skills"}
	result := map[string][]string{}
	for _, root := range roots {
		path := filepath.Join(projectDir, filepath.FromSlash(root))
		names := listImmediateDirs(path)
		names = append(names, listStemFiles(path, ".md")...)
		if len(names) > 0 {
			result[root] = unique(names)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func discoverLocalAgents(projectDir string) map[string][]string {
	roots := map[string]string{
		".claude/agents": ".md",
		".codex/agents":  ".toml",
	}
	result := map[string][]string{}
	for root, suffix := range roots {
		names := listStemFiles(filepath.Join(projectDir, filepath.FromSlash(root)), suffix)
		if len(names) > 0 {
			result[root] = unique(names)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func buildRepoBaselineLock(projectDir string, report repoAuditReport) repoBaselineLock {
	surfaces := []repoSurfaceLock{}
	for _, surface := range report.Surfaces {
		lock := repoSurfaceLock{
			Path:   surface.Path,
			Mode:   surface.Mode,
			Status: surface.Status,
		}
		if surface.Status == "present" && surface.Kind == "file" {
			lock.ContentHash = hashRepoSurface(filepath.Join(projectDir, filepath.FromSlash(surface.Path)))
		}
		surfaces = append(surfaces, lock)
	}
	return repoBaselineLock{
		Version:         1,
		Baseline:        report.Baseline,
		Profile:         report.Profile,
		ResolvedAt:      time.Now().UTC().Format(time.RFC3339),
		EffectiveAgents: report.EffectiveAgents,
		EffectivePacks:  report.EffectivePacks,
		Surfaces:        surfaces,
		Findings:        report.Findings,
	}
}

func hashRepoSurface(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeJSON(file, value)
}

func printRepoReport(report repoAuditReport) {
	fmt.Printf("State: %s\n", report.State)
	fmt.Printf("Profile: %s\n", report.Profile)
	fmt.Printf("Effective agents: %d\n", len(report.EffectiveAgents))
	fmt.Printf("Effective packs: %d\n", len(report.EffectivePacks))
	if len(report.Surfaces) > 0 {
		fmt.Println("Surfaces:")
		for _, surface := range report.Surfaces {
			fmt.Printf("- %s [%s] %s\n", surface.Path, surface.Mode, surface.Status)
		}
	}
	printFindings(report.Findings)
	if len(report.Suggestions) > 0 {
		printSuggestions(report.Suggestions)
	}
}

func printFindings(findings []repoFinding) {
	if len(findings) == 0 {
		fmt.Println("Findings: none")
		return
	}
	fmt.Println("Findings:")
	for _, finding := range findings {
		if finding.Path != "" {
			fmt.Printf("- [%s] %s: %s (%s)\n", finding.Severity, finding.Code, finding.Message, finding.Path)
		} else {
			fmt.Printf("- [%s] %s: %s\n", finding.Severity, finding.Code, finding.Message)
		}
	}
}

func printSuggestions(suggestions []string) {
	if len(suggestions) == 0 {
		fmt.Println("Suggestions: none")
		return
	}
	fmt.Println("Suggestions:")
	for _, suggestion := range suggestions {
		fmt.Printf("- %s\n", suggestion)
	}
}

func repoHasBlockingFindings(findings []repoFinding) bool {
	for _, finding := range findings {
		if finding.Severity == "error" || finding.Severity == "warning" {
			return true
		}
	}
	return false
}

func knownRepoSurfacePaths() []string {
	return []string{
		".agent-docs",
		".beads",
		".claude/agents",
		".claude/skills",
		".codex/agents",
		".codex/skills",
		".github/skills",
		".skill-harness/agent-stack.json",
		".skill-harness/setup-proof.json",
		"AGENT_INSTRUCTIONS.md",
		"AGENTS.md",
		"CLAUDE.md",
		"docs/artifacts/artifacts.manifest.json",
		"docs/artifacts/source/models",
		"generated/review",
		"llms.txt",
		"packs",
		"scripts/beads",
	}
}

func validRepoSurfaceMode(mode string) bool {
	switch mode {
	case repoSurfaceGenerated, repoSurfaceManagedSection, repoSurfaceOverlay, repoSurfaceOwned, repoSurfaceIgnored:
		return true
	default:
		return false
	}
}

func repoManifestPath(projectDir string) string {
	return filepath.Join(projectDir, ".skill-harness", "baseline.manifest.json")
}

func agentStackPath(projectDir string) string {
	return filepath.Join(projectDir, ".skill-harness", "agent-stack.json")
}

func agentStackLockPath(projectDir string) string {
	return filepath.Join(projectDir, ".skill-harness", "agent-stack.lock.json")
}

func repoLockPath(projectDir string) string {
	return filepath.Join(projectDir, ".skill-harness", "baseline.lock.json")
}

func repoUpdateReportPath(projectDir string) string {
	return filepath.Join(projectDir, ".skill-harness", "update-report.json")
}

func repoPathIfExists(path string) string {
	if pathExists(path) {
		return path
	}
	return ""
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func countImmediateChildren(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(entries)
}

func listImmediateDirs(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	out := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			out = append(out, entry.Name())
		}
	}
	return unique(out)
}

func listStemFiles(path, suffix string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	out := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		out = append(out, strings.TrimSuffix(entry.Name(), suffix))
	}
	return unique(out)
}

func gitRemoteURL(root string) string {
	command := exec.Command("git", "config", "--get", "remote.origin.url")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func currentGitRevision(root string) string {
	command := exec.Command("git", "rev-parse", "--short=12", "HEAD")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func printRepoUsage() {
	fmt.Println("skill-harness repo")
	fmt.Println("  init   --dir <project> [--profile minimal|team|agent-native]")
	fmt.Println("  audit  --dir <project> [--json]")
	fmt.Println("  drift  --dir <project> [--json]")
	fmt.Println("  update --dir <project> --check [--json]")
	fmt.Println("  trim   --dir <project> --dry-run [--json]")
	fmt.Println("  sync   --dir <project> [--json]")
	fmt.Println("  lock   --dir <project> [--json]")
}

func promptInstallSelection(loadouts loadoutConfig, deps dependencyConfig) selection {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Install mode:")
	fmt.Println("1. Everything")
	fmt.Println("2. Selected agents")
	fmt.Println("3. Selected packs")
	fmt.Println("4. Selected agents and packs")
	fmt.Print("> ")
	mode := readLine(reader)

	switch mode {
	case "", "1":
		return selection{All: true}
	case "2":
		return selection{Agent: promptAgentList("Choose agents", sortedKeys(loadouts))}
	case "3":
		return selection{Repo: promptPackList("Choose packs", sortedKeys(deps.Repos))}
	case "4":
		return selection{
			Agent: promptAgentList("Choose agents", sortedKeys(loadouts)),
			Repo:  promptPackList("Choose packs", sortedKeys(deps.Repos)),
		}
	default:
		exitOnErr(fmt.Errorf("unknown mode: %s", mode))
		return selection{}
	}
}

func promptAgentList(title string, options []string) []string {
	fmt.Println(title + ":")
	return promptIndexedSelection(options)
}

func promptPackList(title string, options []string) []string {
	fmt.Println(title + ":")
	return promptIndexedSelection(options)
}

func promptIndexedSelection(options []string) []string {
	reader := bufio.NewReader(os.Stdin)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	fmt.Print("Enter comma-separated numbers or 'all': ")
	input := readLine(reader)
	if strings.EqualFold(input, "all") || strings.TrimSpace(input) == "" {
		return append([]string(nil), options...)
	}

	selected := []string{}
	seen := map[string]bool{}
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		index, err := strconv.Atoi(part)
		exitOnErr(err)
		if index < 1 || index > len(options) {
			exitOnErr(fmt.Errorf("selection out of range: %d", index))
		}
		name := options[index-1]
		if !seen[name] {
			selected = append(selected, name)
			seen[name] = true
		}
	}
	return selected
}

func bootstrapArgs(sel selection) []string {
	if sel.All || (len(sel.Agent) == 0 && len(sel.Repo) == 0) {
		return []string{"--all"}
	}
	args := []string{}
	for _, agent := range sel.Agent {
		args = append(args, "--agent", agent)
	}
	for _, repo := range sel.Repo {
		args = append(args, "--repo", repo)
	}
	return args
}

func repoArgs(repos []string) []string {
	if len(repos) == 0 {
		return []string{"--all"}
	}
	args := []string{}
	for _, repo := range repos {
		args = append(args, "--repo", repo)
	}
	return args
}

func agentArgs(agents []string) []string {
	if len(agents) == 0 {
		return []string{"--all"}
	}
	args := []string{}
	for _, agent := range agents {
		args = append(args, "--agent", agent)
	}
	return args
}

func resolveAgents(sel selection, loadouts loadoutConfig) []string {
	if sel.All || (len(sel.Agent) == 0 && len(sel.Repo) == 0) {
		return sortedKeys(loadouts)
	}
	for _, agent := range sel.Agent {
		if _, ok := loadouts[agent]; !ok {
			exitOnErr(fmt.Errorf("unknown agent: %s", agent))
		}
	}
	return unique(sel.Agent)
}

func resolveRepos(sel selection, deps dependencyConfig) []string {
	if sel.All || (len(sel.Agent) == 0 && len(sel.Repo) == 0) {
		return sortedKeys(deps.Repos)
	}
	repos := append([]string(nil), sel.Repo...)
	for _, agent := range sel.Agent {
		cfg, ok := deps.Agents[agent]
		if !ok {
			exitOnErr(fmt.Errorf("unknown agent: %s", agent))
		}
		repos = append(repos, cfg.Repos...)
	}
	for _, repo := range repos {
		if _, ok := deps.Repos[repo]; !ok {
			exitOnErr(fmt.Errorf("unknown repo: %s", repo))
		}
	}
	return unique(repos)
}

func resolveAgentStack(projectDir string, deps dependencyConfig, loadouts loadoutConfig) (agentStackResolution, error) {
	stack, exists, err := readAgentStack(projectDir)
	if err != nil {
		return agentStackResolution{}, err
	}
	if !exists {
		stack = defaultAgentStackConfig()
	}
	normalizeAgentStack(&stack)

	diagnostics := []agentStackDiagnostic{}
	agents := resolveAgentStackAgents(stack, loadouts, &diagnostics)
	agentSkills := map[string][]string{}
	skillCatalog := knownSkills(loadouts)

	for _, agent := range agents {
		base := append([]string(nil), loadouts[agent].Skills...)
		rule := stack.Agents[agent]
		skills := applyAgentStackSkillRule(agent, base, rule, skillCatalog, len(stack.RepoLocalPacks) > 0, &diagnostics)
		agentSkills[agent] = skills
	}

	packs := resolveAgentStackPacks(stack, agents, deps, &diagnostics)
	for _, pack := range stack.DisabledPacks {
		for _, agent := range agents {
			if stringSliceContains(deps.Agents[agent].Repos, pack) {
				diagnostics = append(diagnostics, agentStackDiagnostic{
					Severity: "warning",
					Code:     "disabled-pack-required-by-agent",
					Message:  fmt.Sprintf("agent %s normally depends on disabled pack %s", agent, pack),
				})
			}
		}
	}

	state := "clean"
	if len(stack.Agents) > 0 || len(stack.DisabledAgents) > 0 || len(stack.DisabledPacks) > 0 || len(stack.EnabledPacks) > 0 || len(stack.RepoLocalPacks) > 0 {
		state = "overridden"
	}
	if agentStackHasErrors(diagnostics) {
		state = "conflicted"
	}

	return agentStackResolution{
		Version:         1,
		Profile:         stack.Profile,
		State:           state,
		Baseline:        stack.Baseline,
		EffectiveAgents: agents,
		EffectivePacks:  packs,
		RepoLocalPacks:  unique(stack.RepoLocalPacks),
		AgentSkills:     agentSkills,
		OptOuts: agentStackOptOuts{
			DisabledAgents: unique(stack.DisabledAgents),
			DisabledPacks:  unique(stack.DisabledPacks),
		},
		Overlays:    stack.Agents,
		Diagnostics: diagnostics,
	}, nil
}

func requireAgentStack(projectDir string) error {
	if fileExists(agentStackPath(projectDir)) {
		return nil
	}
	return fmt.Errorf("missing %s; run skill-harness bootstrap --agent-native --dir %s first", agentStackPath(projectDir), projectDir)
}

func readAgentStack(projectDir string) (agentStackConfig, bool, error) {
	path := agentStackPath(projectDir)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return agentStackConfig{}, false, nil
	}
	if err != nil {
		return agentStackConfig{}, false, err
	}
	var stack agentStackConfig
	if err := json.Unmarshal(data, &stack); err != nil {
		return agentStackConfig{}, true, fmt.Errorf("invalid agent stack config %s: %w", path, err)
	}
	return stack, true, nil
}

func normalizeAgentStack(stack *agentStackConfig) {
	if stack.Version == 0 {
		stack.Version = 1
	}
	if stack.Profile == "" {
		stack.Profile = "default"
	}
	if stack.Baseline.Source == "" {
		stack.Baseline.Source = "https://github.com/45ck/skill-harness.git"
	}
	if stack.Baseline.Channel == "" {
		stack.Baseline.Channel = "stable"
	}
	if stack.Agents == nil {
		stack.Agents = map[string]agentStackAgentRule{}
	}
	if stack.Policies == nil {
		stack.Policies = defaultAgentStackPolicies()
	}
	stack.EnabledAgents = unique(stack.EnabledAgents)
	stack.DisabledAgents = unique(stack.DisabledAgents)
	stack.EnabledPacks = unique(stack.EnabledPacks)
	stack.DisabledPacks = unique(stack.DisabledPacks)
	stack.RepoLocalPacks = unique(stack.RepoLocalPacks)
}

func defaultAgentStackConfig() agentStackConfig {
	return agentStackConfig{
		Version: 1,
		Baseline: agentStackBaseline{
			Source:  "https://github.com/45ck/skill-harness.git",
			Channel: "stable",
		},
		Profile:  "default",
		Policies: defaultAgentStackPolicies(),
	}
}

func defaultAgentStackPolicies() map[string]string {
	return map[string]string{
		"packageInstalls":   "ask",
		"globalWrites":      "ask",
		"hookChanges":       "ask",
		"ciChanges":         "ask",
		"guardrailOverride": "ask",
	}
}

func writeDefaultAgentStack(projectDir string, force bool) error {
	path := filepath.Join(projectDir, ".skill-harness", "agent-stack.json")
	if !force && fileExists(path) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(defaultAgentStackConfig(), "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeResolvedAgentStackLock(projectDir string, deps dependencyConfig, loadouts loadoutConfig) (agentStackResolution, agentStackLock, error) {
	resolution, err := resolveAgentStack(projectDir, deps, loadouts)
	if err != nil {
		return agentStackResolution{}, agentStackLock{}, err
	}
	if err := errorOnAgentStackDiagnostics(resolution.Diagnostics); err != nil {
		return resolution, agentStackLock{}, err
	}
	lock := buildAgentStackLock(projectDir, resolution)
	if err := writeJSONFile(agentStackLockPath(projectDir), lock); err != nil {
		return resolution, agentStackLock{}, err
	}
	return resolution, lock, nil
}

func buildAgentStackLock(projectDir string, resolution agentStackResolution) agentStackLock {
	return agentStackLock{
		Version:         1,
		Baseline:        resolution.Baseline,
		Profile:         resolution.Profile,
		ResolvedAt:      time.Now().UTC().Format(time.RFC3339),
		OverlayHash:     hashRepoSurface(agentStackPath(projectDir)),
		EffectiveAgents: resolution.EffectiveAgents,
		EffectivePacks:  resolution.EffectivePacks,
		RepoLocalPacks:  resolution.RepoLocalPacks,
		AgentSkills:     resolution.AgentSkills,
		OptOuts:         resolution.OptOuts,
		Diagnostics:     resolution.Diagnostics,
		Surfaces:        buildAgentStackSurfaceLocks(projectDir),
	}
}

func buildAgentStackSurfaceLocks(projectDir string) []agentStackSurfaceLock {
	paths := []struct {
		path string
		mode string
	}{
		{".skill-harness/agent-stack.json", repoSurfaceOverlay},
		{".skill-harness/agent-stack.lock.json", repoSurfaceGenerated},
		{".skill-harness/setup-proof.json", repoSurfaceGenerated},
		{".claude/agents", repoSurfaceGenerated},
		{".codex/agents", repoSurfaceGenerated},
		{".claude/skills", repoSurfaceOverlay},
		{".codex/skills", repoSurfaceOverlay},
		{".github/skills", repoSurfaceOverlay},
		{"packs", repoSurfaceOverlay},
	}
	surfaces := []agentStackSurfaceLock{}
	for _, item := range paths {
		fullPath := filepath.Join(projectDir, filepath.FromSlash(item.path))
		surface := agentStackSurfaceLock{Path: item.path, Mode: item.mode, Status: "missing"}
		info, err := os.Stat(fullPath)
		if err == nil {
			surface.Status = "present"
			if info.IsDir() {
				surface.Kind = "dir"
			} else {
				surface.Kind = "file"
				surface.ContentHash = hashRepoSurface(fullPath)
			}
		} else if item.path == ".skill-harness/agent-stack.lock.json" {
			surface.Status = "present"
			surface.Kind = "file"
		}
		surfaces = append(surfaces, surface)
	}
	return surfaces
}

func resolveAgentStackAgents(stack agentStackConfig, loadouts loadoutConfig, diagnostics *[]agentStackDiagnostic) []string {
	agents := []string{}
	if len(stack.EnabledAgents) > 0 {
		agents = append(agents, stack.EnabledAgents...)
	} else {
		agents = defaultProfileAgents(stack.Profile, loadouts)
	}
	disabled := toSet(stack.DisabledAgents)
	filtered := []string{}
	for _, agent := range unique(agents) {
		if _, ok := loadouts[agent]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-agent",
				Message:  fmt.Sprintf("agent %s is not defined in the baseline loadouts", agent),
			})
			continue
		}
		if disabled[agent] {
			continue
		}
		filtered = append(filtered, agent)
	}
	for _, agent := range stack.DisabledAgents {
		if _, ok := loadouts[agent]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-disabled-agent",
				Message:  fmt.Sprintf("disabled agent %s is not defined in the baseline loadouts", agent),
			})
		}
	}
	for agent := range stack.Agents {
		if _, ok := loadouts[agent]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-agent-overlay",
				Message:  fmt.Sprintf("overlay references unknown agent %s", agent),
			})
		}
	}
	return unique(filtered)
}

func defaultProfileAgents(profile string, loadouts loadoutConfig) []string {
	switch profile {
	case "minimal":
		return []string{"requirements-analyst", "software-architect", "workflow-engineer"}
	case "security":
		return []string{"security-reviewer", "pentest-reviewer", "software-architect", "workflow-engineer"}
	default:
		return sortedKeys(loadouts)
	}
}

func resolveAgentStackPacks(stack agentStackConfig, agents []string, deps dependencyConfig, diagnostics *[]agentStackDiagnostic) []string {
	packs := []string{}
	for _, pack := range stack.EnabledPacks {
		if _, ok := deps.Repos[pack]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-enabled-pack",
				Message:  fmt.Sprintf("enabled pack %s is not defined in the baseline dependencies", pack),
			})
			continue
		}
		packs = append(packs, pack)
	}
	for _, agent := range agents {
		cfg, ok := deps.Agents[agent]
		if !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "missing-agent-dependencies",
				Message:  fmt.Sprintf("agent %s has no dependency mapping", agent),
			})
			continue
		}
		packs = append(packs, cfg.Repos...)
	}
	disabled := toSet(stack.DisabledPacks)
	filtered := []string{}
	for _, pack := range unique(packs) {
		if _, ok := deps.Repos[pack]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-pack",
				Message:  fmt.Sprintf("pack %s is not defined in the baseline dependencies", pack),
			})
			continue
		}
		if disabled[pack] {
			continue
		}
		filtered = append(filtered, pack)
	}
	for _, pack := range stack.DisabledPacks {
		if _, ok := deps.Repos[pack]; !ok {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "error",
				Code:     "unknown-disabled-pack",
				Message:  fmt.Sprintf("disabled pack %s is not defined in the baseline dependencies", pack),
			})
		}
	}
	return unique(filtered)
}

func applyAgentStackSkillRule(agent string, base []string, rule agentStackAgentRule, skillCatalog map[string]bool, hasRepoLocalPacks bool, diagnostics *[]agentStackDiagnostic) []string {
	remove := toSet(rule.RemoveSkills)
	out := []string{}
	for _, skill := range base {
		if remove[skill] {
			continue
		}
		if replacement, ok := rule.ReplaceSkills[skill]; ok {
			out = append(out, replacement)
			continue
		}
		out = append(out, skill)
	}
	for _, skill := range rule.RemoveSkills {
		if !stringSliceContains(base, skill) {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "warning",
				Code:     "remove-skill-not-in-agent",
				Message:  fmt.Sprintf("agent %s does not include removed skill %s in the baseline", agent, skill),
			})
		}
	}
	for oldSkill, newSkill := range rule.ReplaceSkills {
		if !stringSliceContains(base, oldSkill) {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "warning",
				Code:     "replace-skill-not-in-agent",
				Message:  fmt.Sprintf("agent %s does not include replaced skill %s in the baseline", agent, oldSkill),
			})
		}
		if !skillCatalog[newSkill] && !hasRepoLocalPacks {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "warning",
				Code:     "unknown-replacement-skill",
				Message:  fmt.Sprintf("replacement skill %s is not in the baseline skill catalog and no repoLocalPacks are configured", newSkill),
			})
		}
	}
	for _, skill := range rule.AddSkills {
		if !skillCatalog[skill] && !hasRepoLocalPacks {
			*diagnostics = append(*diagnostics, agentStackDiagnostic{
				Severity: "warning",
				Code:     "unknown-added-skill",
				Message:  fmt.Sprintf("added skill %s is not in the baseline skill catalog and no repoLocalPacks are configured", skill),
			})
		}
		out = append(out, skill)
	}
	return unique(out)
}

func knownSkills(loadouts loadoutConfig) map[string]bool {
	out := map[string]bool{}
	for _, loadout := range loadouts {
		for _, skill := range loadout.Skills {
			out[skill] = true
		}
	}
	return out
}

func auditAgentStackProject(projectDir string, deps dependencyConfig, loadouts loadoutConfig) agentStackAudit {
	stackPath := filepath.Join(projectDir, ".skill-harness", "agent-stack.json")
	lockPath := filepath.Join(projectDir, ".skill-harness", "agent-stack.lock.json")
	proofPath := filepath.Join(projectDir, ".skill-harness", "setup-proof.json")
	hasStack := fileExists(stackPath)
	hasLock := fileExists(lockPath)
	hasProof := fileExists(proofPath)
	hasClaudeAgents := dirHasFiles(filepath.Join(projectDir, ".claude", "agents"))
	hasCodexAgents := dirHasFiles(filepath.Join(projectDir, ".codex", "agents"))

	result := agentStackAudit{
		State: "unmanaged",
		Paths: map[string]bool{
			".skill-harness/agent-stack.json":      hasStack,
			".skill-harness/agent-stack.lock.json": hasLock,
			".skill-harness/setup-proof.json":      hasProof,
			".claude/agents":                       hasClaudeAgents,
			".codex/agents":                        hasCodexAgents,
		},
	}
	if !hasStack {
		if hasClaudeAgents || hasCodexAgents {
			result.State = "generated-only"
			result.Reasons = append(result.Reasons, "agent files exist but no repo-local agent stack overlay was found")
		} else {
			result.Reasons = append(result.Reasons, "no agent stack overlay or generated agent files were found")
		}
		return result
	}
	resolution, err := resolveAgentStack(projectDir, deps, loadouts)
	if err != nil {
		result.State = "conflicted"
		result.Reasons = append(result.Reasons, err.Error())
		return result
	}
	result.Resolution = &resolution
	result.Diagnostics = resolution.Diagnostics
	if agentStackHasErrors(resolution.Diagnostics) {
		result.State = "conflicted"
		result.Reasons = append(result.Reasons, "agent stack overlay has resolution errors")
		return result
	}
	result.State = "agent-native"
	result.Reasons = append(result.Reasons, "repo-local agent stack overlay is present")
	if !hasLock {
		result.Reasons = append(result.Reasons, "lockfile is missing; run resolve/update once lockfile support is enabled")
	}
	if !hasProof {
		result.Reasons = append(result.Reasons, "setup proof is missing")
	}
	return result
}

type agentStackAudit struct {
	State       string                 `json:"state"`
	Reasons     []string               `json:"reasons"`
	Paths       map[string]bool        `json:"paths"`
	Diagnostics []agentStackDiagnostic `json:"diagnostics,omitempty"`
	Resolution  *agentStackResolution  `json:"resolution,omitempty"`
}

func printAgentStackResolution(resolution agentStackResolution) {
	fmt.Printf("Profile: %s\n", resolution.Profile)
	fmt.Printf("State: %s\n", resolution.State)
	fmt.Printf("Baseline: %s (%s)\n", resolution.Baseline.Source, resolution.Baseline.Channel)
	if resolution.Baseline.Pin != "" {
		fmt.Printf("Pin: %s\n", resolution.Baseline.Pin)
	}
	fmt.Println("Effective agents:")
	for _, agent := range resolution.EffectiveAgents {
		fmt.Printf("- %s (%d skills)\n", agent, len(resolution.AgentSkills[agent]))
	}
	fmt.Println("Effective packs:")
	for _, pack := range resolution.EffectivePacks {
		fmt.Printf("- %s\n", pack)
	}
	if len(resolution.RepoLocalPacks) > 0 {
		fmt.Println("Repo-local packs:")
		for _, pack := range resolution.RepoLocalPacks {
			fmt.Printf("- %s\n", pack)
		}
	}
	if len(resolution.Diagnostics) > 0 {
		fmt.Println("Diagnostics:")
		for _, diagnostic := range resolution.Diagnostics {
			fmt.Printf("- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
		}
	}
}

func agentStackHasErrors(diagnostics []agentStackDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return true
		}
	}
	return false
}

func errorOnAgentStackDiagnostics(diagnostics []agentStackDiagnostic) error {
	if !agentStackHasErrors(diagnostics) {
		return nil
	}
	messages := []string{}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			messages = append(messages, diagnostic.Code+": "+diagnostic.Message)
		}
	}
	return errors.New(strings.Join(messages, "; "))
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func toSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func dirHasFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return true
		}
	}
	return false
}

func copyClaudeAgents(root string, agents []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	targetDir := filepath.Join(home, ".claude", "agents")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	for _, agent := range agents {
		src := filepath.Join(root, ".claude", "agents", agent+".md")
		dst := filepath.Join(targetDir, agent+".md")
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Close()
}

func runPython(root, script string, args ...string) error {
	command := exec.Command("python", append([]string{filepath.Join(root, script)}, args...)...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = root
	return command.Run()
}

func runCommand(dir, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = dir
	return command.Run()
}

func loadDependencies(root string) dependencyConfig {
	var cfg dependencyConfig
	data := mustRead(filepath.Join(root, "scripts", "dependencies.json"))
	exitOnErr(json.Unmarshal(data, &cfg))
	return cfg
}

func loadLoadouts(root string) loadoutConfig {
	var cfg loadoutConfig
	data := mustRead(filepath.Join(root, "scripts", "agent_loadouts.json"))
	exitOnErr(json.Unmarshal(data, &cfg))
	return cfg
}

func mustRead(path string) []byte {
	data, err := os.ReadFile(path)
	exitOnErr(err)
	return data
}

func requireCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required on PATH", name)
	}
	return nil
}

func requireSetupCommands(manager packageManager) error {
	switch manager {
	case packageManagerNpm:
		if err := requireCommand("npm"); err != nil {
			return err
		}
		return requireCommand("npx")
	case packageManagerPnpm:
		return requireCommand("pnpm")
	case packageManagerYarn:
		return requireCommand("yarn")
	case packageManagerBun:
		return requireCommand("bun")
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func installPackages(dir string, manager packageManager, packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	switch manager {
	case packageManagerNpm:
		return runCommand(dir, "npm", append([]string{"install", "-D"}, packages...)...)
	case packageManagerPnpm:
		return runCommand(dir, "pnpm", append([]string{"add", "-D"}, packages...)...)
	case packageManagerYarn:
		return runCommand(dir, "yarn", append([]string{"add", "-D"}, packages...)...)
	case packageManagerBun:
		return runCommand(dir, "bun", append([]string{"add", "-d"}, packages...)...)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func runLocalTool(dir string, manager packageManager, tool string, args ...string) error {
	switch manager {
	case packageManagerNpm:
		return runCommand(dir, "npx", append([]string{tool}, args...)...)
	case packageManagerPnpm:
		return runCommand(dir, "pnpm", append([]string{"exec", tool}, args...)...)
	case packageManagerYarn:
		return runCommand(dir, "yarn", append([]string{"exec", tool}, args...)...)
	case packageManagerBun:
		return runCommand(dir, "bun", append([]string{"x", tool}, args...)...)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func resolveProjectSetupContext(targetDir, scopeValue, packageManagerValue string) (projectSetupContext, error) {
	scope, err := parseProjectScope(scopeValue)
	if err != nil {
		return projectSetupContext{}, err
	}
	managerPreference, err := parsePackageManager(packageManagerValue)
	if err != nil {
		return projectSetupContext{}, err
	}

	monorepoRoot := findMonorepoRoot(targetDir)
	operationDir := targetDir
	if scope != projectScopeWorkspace && monorepoRoot != "" {
		operationDir = monorepoRoot
	}

	manager, err := resolvePackageManager(managerPreference, operationDir)
	if err != nil {
		return projectSetupContext{}, err
	}

	return projectSetupContext{
		TargetDir:      targetDir,
		OperationDir:   operationDir,
		MonorepoRoot:   monorepoRoot,
		Monorepo:       monorepoRoot != "",
		Scope:          scope,
		PackageManager: manager,
	}, nil
}

func parseProjectScope(value string) (projectScope, error) {
	switch projectScope(strings.ToLower(strings.TrimSpace(value))) {
	case projectScopeAuto:
		return projectScopeAuto, nil
	case projectScopeRoot:
		return projectScopeRoot, nil
	case projectScopeWorkspace:
		return projectScopeWorkspace, nil
	default:
		return "", fmt.Errorf("unsupported setup scope: %s", value)
	}
}

func parsePackageManager(value string) (packageManager, error) {
	switch packageManager(strings.ToLower(strings.TrimSpace(value))) {
	case packageManagerAuto:
		return packageManagerAuto, nil
	case packageManagerNpm:
		return packageManagerNpm, nil
	case packageManagerPnpm:
		return packageManagerPnpm, nil
	case packageManagerYarn:
		return packageManagerYarn, nil
	case packageManagerBun:
		return packageManagerBun, nil
	default:
		return "", fmt.Errorf("unsupported package manager: %s", value)
	}
}

func parseArtifactProfile(value string) (artifactProfile, error) {
	switch artifactProfile(strings.ToLower(strings.TrimSpace(value))) {
	case artifactProfileAuto:
		return artifactProfileAuto, nil
	case artifactProfileMarkdown:
		return artifactProfileMarkdown, nil
	case artifactProfileHTML:
		return artifactProfileHTML, nil
	case artifactProfileDual:
		return artifactProfileDual, nil
	case artifactProfileCodexApp:
		return artifactProfileCodexApp, nil
	case artifactProfileClaudeDesktop:
		return artifactProfileClaudeDesktop, nil
	case artifactProfileCLI:
		return artifactProfileCLI, nil
	case artifactProfileTUI:
		return artifactProfileTUI, nil
	case artifactProfileMedia:
		return artifactProfileMedia, nil
	case artifactProfileAgentLoop:
		return artifactProfileAgentLoop, nil
	case artifactProfileNone:
		return artifactProfileNone, nil
	default:
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "governed-agent" || normalized == "self-improving" || normalized == "self-improving-agent-loop" {
			return artifactProfileAgentLoop, nil
		}
		return "", fmt.Errorf("unsupported artifact profile: %s", value)
	}
}

func parseModelingMode(value string) (modelingMode, error) {
	switch modelingMode(strings.ToLower(strings.TrimSpace(value))) {
	case "", modelingModeAuto:
		return modelingModeAuto, nil
	case modelingModeOff, "none", "skip", "disabled":
		return modelingModeOff, nil
	case modelingModeBaseline, "source", "source-first":
		return modelingModeBaseline, nil
	case modelingModeUMLFirst, "uml", "umlfirst", "review":
		return modelingModeUMLFirst, nil
	default:
		return "", fmt.Errorf("unsupported modeling mode %q: expected auto, off, baseline, or uml-first", value)
	}
}

func resolveEffectiveModelingMode(projectDir string, requested modelingMode, profile artifactProfile, artifactsDisabled bool) modelingMode {
	if artifactsDisabled || profile == artifactProfileNone {
		return modelingModeOff
	}
	if requested != modelingModeAuto {
		return requested
	}
	if existing := existingModelingMode(projectDir); existing != modelingModeAuto {
		return existing
	}
	if fileExists(filepath.Join(projectDir, ".skill-harness", "project.json")) {
		return modelingModeOff
	}
	return modelingModeUMLFirst
}

func existingModelingMode(projectDir string) modelingMode {
	configPath := filepath.Join(projectDir, ".skill-harness", "project.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return modelingModeAuto
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return modelingModeAuto
	}
	capabilities, _ := config["capabilities"].(map[string]any)
	developerArtifacts, _ := capabilities["developerArtifacts"].(map[string]any)
	modeling, _ := developerArtifacts["modeling"].(map[string]any)
	if value, ok := modeling["mode"].(string); ok {
		if mode, err := parseModelingMode(value); err == nil && mode != modelingModeAuto {
			return mode
		}
	}
	modelPolicy, _ := developerArtifacts["modelPolicy"].(map[string]any)
	uml, _ := modelPolicy["uml"].(map[string]any)
	if enabled, ok := uml["enabled"].(bool); ok && enabled {
		return modelingModeUMLFirst
	}
	return modelingModeAuto
}

func flagWasSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func resolvePackageManager(preferred packageManager, startDir string) (packageManager, error) {
	if preferred != packageManagerAuto {
		return preferred, nil
	}
	for dir := startDir; ; dir = filepath.Dir(dir) {
		if manager := detectPackageManagerInDir(dir); manager != "" {
			return manager, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return packageManagerNpm, nil
}

func detectPackageManagerInDir(dir string) packageManager {
	if metadata, ok := readPackageJSONMetadata(dir); ok {
		if manager := packageManagerFromMetadata(metadata.PackageManager); manager != "" {
			return manager
		}
	}
	switch {
	case fileExists(filepath.Join(dir, "bun.lockb")), fileExists(filepath.Join(dir, "bun.lock")):
		return packageManagerBun
	case fileExists(filepath.Join(dir, "pnpm-lock.yaml")), fileExists(filepath.Join(dir, "pnpm-workspace.yaml")):
		return packageManagerPnpm
	case fileExists(filepath.Join(dir, "yarn.lock")):
		return packageManagerYarn
	case fileExists(filepath.Join(dir, "package-lock.json")), fileExists(filepath.Join(dir, "npm-shrinkwrap.json")):
		return packageManagerNpm
	default:
		return ""
	}
}

func packageManagerFromMetadata(value string) packageManager {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch {
	case normalized == "":
		return ""
	case strings.HasPrefix(normalized, "npm@"), normalized == "npm":
		return packageManagerNpm
	case strings.HasPrefix(normalized, "pnpm@"), normalized == "pnpm":
		return packageManagerPnpm
	case strings.HasPrefix(normalized, "yarn@"), normalized == "yarn":
		return packageManagerYarn
	case strings.HasPrefix(normalized, "bun@"), normalized == "bun":
		return packageManagerBun
	default:
		return ""
	}
}

type packageJSONMetadata struct {
	Workspaces     json.RawMessage `json:"workspaces"`
	PackageManager string          `json:"packageManager"`
}

func readPackageJSONMetadata(dir string) (packageJSONMetadata, bool) {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return packageJSONMetadata{}, false
	}
	var metadata packageJSONMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return packageJSONMetadata{}, false
	}
	return metadata, true
}

func findMonorepoRoot(start string) string {
	for dir := start; ; dir = filepath.Dir(dir) {
		if isMonorepoRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
	}
}

func isMonorepoRoot(dir string) bool {
	if fileExists(filepath.Join(dir, "pnpm-workspace.yaml")) ||
		fileExists(filepath.Join(dir, "lerna.json")) ||
		fileExists(filepath.Join(dir, "nx.json")) ||
		fileExists(filepath.Join(dir, "turbo.json")) ||
		fileExists(filepath.Join(dir, "rush.json")) {
		return true
	}
	metadata, ok := readPackageJSONMetadata(dir)
	return ok && len(metadata.Workspaces) > 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func projectSetupSummary(prefix string, ctx projectSetupContext) string {
	if ctx.OperationDir == ctx.TargetDir {
		return fmt.Sprintf("%s in %s (scope=%s, package-manager=%s)", prefix, ctx.OperationDir, ctx.Scope, ctx.PackageManager)
	}
	return fmt.Sprintf(
		"%s in %s (target=%s, scope=%s, package-manager=%s)",
		prefix,
		ctx.OperationDir,
		ctx.TargetDir,
		ctx.Scope,
		ctx.PackageManager,
	)
}

func buildProjectSetupProof(ctx projectSetupContext, options projectSetupProofOptions) projectSetupProof {
	artifactScaffolded := !options.InstallOnly &&
		!options.SkipArtifacts &&
		!options.SkipDeveloperArtifacts &&
		options.RequestedArtifactProfile != artifactProfileNone

	proof := projectSetupProof{
		Version: 1,
		Project: projectSetupProofProject{
			TargetDir:      ctx.TargetDir,
			OperationDir:   ctx.OperationDir,
			MonorepoRoot:   ctx.MonorepoRoot,
			Monorepo:       ctx.Monorepo,
			Scope:          ctx.Scope,
			PackageManager: ctx.PackageManager,
		},
		Profiles: projectSetupProofProfiles{
			RequestedDeveloperArtifacts: options.RequestedArtifactProfile,
			EffectiveDeveloperArtifacts: options.EffectiveArtifactProfile,
			RequestedModeling:           options.RequestedModelingMode,
			EffectiveModeling:           options.EffectiveModelingMode,
		},
		Tools: map[string]toolProof{
			"packageManager": {
				Status:  "available",
				Command: string(ctx.PackageManager),
			},
			"agentStack": {
				Status: "scaffolded",
				Paths:  []string{".skill-harness/agent-stack.json"},
			},
			"setupProof": {
				Status: "written",
				Paths:  []string{".skill-harness/setup-proof.json"},
			},
		},
		Checks: map[string]checkProof{
			"setupProof": {
				Status: "written",
				Path:   ".skill-harness/setup-proof.json",
			},
		},
		GeneratedPaths: []string{".skill-harness/agent-stack.json", ".skill-harness/setup-proof.json"},
		Skipped:        []string{},
	}

	if options.SkipNoslop {
		proof.Tools["noslop"] = toolProof{Status: "skipped", Package: "@45ck/noslop"}
		proof.Skipped = append(proof.Skipped, "noslop")
	} else if options.InstallOnly {
		proof.Tools["noslop"] = toolProof{Status: "installed", Package: "@45ck/noslop"}
	} else {
		proof.Tools["noslop"] = toolProof{Status: "initialized", Package: "@45ck/noslop"}
	}

	if options.SkipAgentDocs {
		proof.Tools["agentDocs"] = toolProof{Status: "skipped", Package: "github:45ck/agent-docs"}
		proof.Checks["agentDocs"] = checkProof{Status: "skipped"}
		proof.Skipped = append(proof.Skipped, "agent-docs")
	} else if options.InstallOnly {
		proof.Tools["agentDocs"] = toolProof{Status: "installed", Package: "github:45ck/agent-docs"}
		proof.Checks["agentDocs"] = checkProof{Status: "not-run", Command: localToolCommand(ctx.PackageManager, "agent-docs", "check")}
	} else {
		proof.Tools["agentDocs"] = toolProof{
			Status:  "quality-gates-installed",
			Package: "github:45ck/agent-docs",
			Command: localToolCommand(ctx.PackageManager, "agent-docs", "install-gates", "--quality"),
		}
		proof.Checks["agentDocs"] = checkProof{Status: "available", Command: localToolCommand(ctx.PackageManager, "agent-docs", "check")}
	}

	if options.SkipBeads {
		proof.Tools["beads"] = toolProof{Status: "skipped"}
		proof.Skipped = append(proof.Skipped, "beads")
	} else if options.InstallOnly {
		proof.Tools["beads"] = toolProof{Status: "installed", Mode: string(options.BeadsMode), Command: "bd init"}
	} else {
		proof.Tools["beads"] = toolProof{Status: "initialized", Mode: string(options.BeadsMode), Command: "bd init", Paths: []string{".beads/"}}
		proof.GeneratedPaths = append(proof.GeneratedPaths, ".beads/")
	}

	if options.SkipClaudeSettings {
		proof.Tools["claudeSettings"] = toolProof{Status: "skipped"}
		proof.Skipped = append(proof.Skipped, "claude-settings")
	} else {
		proof.Tools["claudeSettings"] = toolProof{Status: "written", Paths: []string{relativeProofPath(ctx.OperationDir, filepath.Join(ctx.TargetDir, ".claude", "settings.json"))}}
		proof.GeneratedPaths = append(proof.GeneratedPaths, relativeProofPath(ctx.OperationDir, filepath.Join(ctx.TargetDir, ".claude", "settings.json")))
	}

	if artifactScaffolded {
		proof.Tools["developerArtifacts"] = toolProof{
			Status: "scaffolded",
			Paths: []string{
				".skill-harness/project.json",
				"docs/artifacts/",
				"generated/review/",
				"scripts/check-artifact-manifest.mjs",
				"scripts/check-artifact-html-policy.mjs",
				"scripts/open-artifact-review.mjs",
			},
		}
		proof.GeneratedPaths = append(proof.GeneratedPaths,
			".skill-harness/project.json",
			"docs/artifacts/",
			"generated/review/",
			"scripts/check-artifact-manifest.mjs",
			"scripts/check-artifact-html-policy.mjs",
			"scripts/open-artifact-review.mjs",
		)
		proof.Checks["artifactManifest"] = checkProof{Status: "available", Command: "node scripts/check-artifact-manifest.mjs", Path: "scripts/check-artifact-manifest.mjs"}
		proof.Checks["artifactHtmlPolicy"] = checkProof{Status: "available", Command: "node scripts/check-artifact-html-policy.mjs", Path: "scripts/check-artifact-html-policy.mjs"}
		proof.Checks["artifactReviewOpen"] = checkProof{Status: "available", Command: "node scripts/open-artifact-review.mjs", Path: "scripts/open-artifact-review.mjs"}
		if options.EffectiveModelingMode != modelingModeOff {
			proof.GeneratedPaths = append(proof.GeneratedPaths,
				"docs/artifacts/source/models/",
				"docs/artifacts/source/models/model-inventory.md",
				"docs/artifacts/templates/model-diff-artifact.md",
				"generated/review/models/",
				"scripts/check-model-artifact-policy.mjs",
				"scripts/generate-model-review.mjs",
			)
			proof.Checks["modelArtifactPolicy"] = checkProof{Status: "available", Command: "node scripts/check-model-artifact-policy.mjs", Path: "scripts/check-model-artifact-policy.mjs"}
			proof.Checks["modelReviewGenerate"] = checkProof{Status: "available", Command: "node scripts/generate-model-review.mjs", Path: "scripts/generate-model-review.mjs"}
		}
		if options.RequestedArtifactProfile == artifactProfileMedia {
			proof.GeneratedPaths = append(proof.GeneratedPaths, "generated/media/", "docs/artifacts/templates/demo-artifact.md")
		}
		if options.RequestedArtifactProfile == artifactProfileAgentLoop {
			proof.GeneratedPaths = append(proof.GeneratedPaths,
				"generated/agent-runs/",
				"docs/artifacts/templates/agent-loop-artifact.md",
				"docs/artifacts/source/agent-loop-playbook.md",
				"scripts/check-agent-loop-policy.mjs",
			)
			proof.Checks["agentLoopPolicy"] = checkProof{Status: "available", Command: "node scripts/check-agent-loop-policy.mjs", Path: "scripts/check-agent-loop-policy.mjs"}
		}
	} else {
		status := "skipped"
		if options.RequestedArtifactProfile == artifactProfileNone {
			status = "disabled"
		}
		proof.Tools["developerArtifacts"] = toolProof{Status: status}
		proof.Checks["artifactManifest"] = checkProof{Status: status}
		proof.Checks["artifactHtmlPolicy"] = checkProof{Status: status}
		proof.Skipped = append(proof.Skipped, "developer-artifacts")
	}

	if !options.SkipBeads && !options.InstallOnly && options.BeadsWorktrees {
		proof.Tools["beadsWorktrees"] = toolProof{Status: "installed", Paths: []string{"scripts/beads/bd.mjs", ".gitignore"}}
		proof.GeneratedPaths = append(proof.GeneratedPaths, "scripts/beads/bd.mjs")
	} else if !options.SkipBeads && !options.InstallOnly {
		proof.Tools["beadsWorktrees"] = toolProof{Status: "skipped"}
		proof.Skipped = append(proof.Skipped, "beads-worktrees")
	}

	sort.Strings(proof.GeneratedPaths)
	proof.GeneratedPaths = unique(proof.GeneratedPaths)
	sort.Strings(proof.Skipped)
	proof.Skipped = unique(proof.Skipped)
	return proof
}

func writeProjectSetupProof(projectDir string, proof projectSetupProof) error {
	proofDir := filepath.Join(projectDir, ".skill-harness")
	if err := os.MkdirAll(proofDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(proofDir, "setup-proof.json"), data, 0o644)
}

func writeAgentNativeSetupProof(ctx projectSetupContext, resolution agentStackResolution, lock agentStackLock) error {
	proof := projectSetupProof{
		Version: 1,
		Project: projectSetupProofProject{
			TargetDir:      ctx.TargetDir,
			OperationDir:   ctx.OperationDir,
			MonorepoRoot:   ctx.MonorepoRoot,
			Monorepo:       ctx.Monorepo,
			Scope:          ctx.Scope,
			PackageManager: ctx.PackageManager,
		},
		Profiles: projectSetupProofProfiles{
			RequestedDeveloperArtifacts: artifactProfileNone,
			EffectiveDeveloperArtifacts: artifactProfileNone,
			RequestedModeling:           modelingModeOff,
			EffectiveModeling:           modelingModeOff,
		},
		Tools: map[string]toolProof{
			"agentStack": {
				Status: "resolved",
				Paths:  []string{".skill-harness/agent-stack.json"},
			},
			"agentStackLock": {
				Status: "written",
				Paths:  []string{".skill-harness/agent-stack.lock.json"},
			},
			"packageManager": {
				Status:  "available",
				Command: string(ctx.PackageManager),
			},
			"setupProof": {
				Status: "written",
				Paths:  []string{".skill-harness/setup-proof.json"},
			},
		},
		Checks: map[string]checkProof{
			"agentStack": {
				Status:  resolution.State,
				Command: "skill-harness resolve --dir . --strict",
				Path:    ".skill-harness/agent-stack.json",
			},
			"agentStackLock": {
				Status: "written",
				Path:   ".skill-harness/agent-stack.lock.json",
			},
			"setupProof": {
				Status: "written",
				Path:   ".skill-harness/setup-proof.json",
			},
		},
		GeneratedPaths: []string{
			".skill-harness/agent-stack.json",
			".skill-harness/agent-stack.lock.json",
			".skill-harness/setup-proof.json",
		},
		Skipped: []string{
			"package-installs",
			"global-writes",
		},
	}
	if len(lock.EffectiveAgents) == 0 {
		proof.Skipped = append(proof.Skipped, "agent-render")
	}
	sort.Strings(proof.GeneratedPaths)
	proof.GeneratedPaths = unique(proof.GeneratedPaths)
	sort.Strings(proof.Skipped)
	proof.Skipped = unique(proof.Skipped)
	return writeProjectSetupProof(ctx.TargetDir, proof)
}

func localToolCommand(manager packageManager, tool string, args ...string) string {
	switch manager {
	case packageManagerNpm:
		return strings.Join(append([]string{"npx", tool}, args...), " ")
	case packageManagerPnpm:
		return strings.Join(append([]string{"pnpm", "exec", tool}, args...), " ")
	case packageManagerYarn:
		return strings.Join(append([]string{"yarn", "exec", tool}, args...), " ")
	case packageManagerBun:
		return strings.Join(append([]string{"bun", "x", tool}, args...), " ")
	default:
		return strings.Join(append([]string{string(manager), tool}, args...), " ")
	}
}

func relativeProofPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return filepath.ToSlash(rel)
}

func writeMinimalPackageJSON(projectDir string) error {
	base := strings.ToLower(filepath.Base(projectDir))
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-", "/", "-", "\\", "-")
	name := replacer.Replace(base)
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	name = strings.Trim(name, "-")
	if name == "" {
		name = "skill-harness-project"
	}
	content, err := json.MarshalIndent(map[string]any{
		"name":    name,
		"private": true,
		"version": "0.0.0",
	}, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(filepath.Join(projectDir, "package.json"), content, 0o644)
}

func updatePackageScripts(projectDir string, agentDocsEnabled bool, profile artifactProfile, mode modelingMode) error {
	if mode == modelingModeAuto {
		mode = modelingModeUMLFirst
	}
	packagePath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}
	metadata := map[string]any{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}
	scripts, _ := metadata["scripts"].(map[string]any)
	if scripts == nil {
		scripts = map[string]any{}
	}
	devDependencies, _ := metadata["devDependencies"].(map[string]any)
	if devDependencies == nil {
		devDependencies = map[string]any{}
	}
	if _, exists := devDependencies["@viz-js/viz"]; !exists {
		devDependencies["@viz-js/viz"] = "3.27.0"
	}
	if _, exists := devDependencies["svg-pan-zoom"]; !exists {
		devDependencies["svg-pan-zoom"] = "^3.6.2"
	}
	if _, exists := devDependencies["cytoscape"]; !exists {
		devDependencies["cytoscape"] = "3.29.2"
	}
	if _, exists := devDependencies["dagre"]; !exists {
		devDependencies["dagre"] = "0.8.5"
	}
	if _, exists := devDependencies["cytoscape-dagre"]; !exists {
		devDependencies["cytoscape-dagre"] = "2.5.0"
	}
	defaultScripts := map[string]string{
		"artifacts:check":          "node scripts/generate-artifact-review.mjs --check && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs",
		"artifacts:generate":       "node scripts/generate-artifact-review.mjs",
		"artifacts:html:check":     "node scripts/check-artifact-html-policy.mjs",
		"artifacts:manifest:check": "node scripts/check-artifact-manifest.mjs",
		"artifacts:open":           "node scripts/open-artifact-review.mjs",
		"artifacts:review":         "node scripts/generate-artifact-review.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs",
	}
	if agentDocsEnabled {
		defaultScripts["docs:check"] = "agent-docs check"
		defaultScripts["docs:generate"] = "agent-docs generate"
		defaultScripts["docs:report"] = "agent-docs report"
	}
	if profile == artifactProfileAgentLoop {
		defaultScripts["agent-loop:check"] = "node scripts/check-agent-loop-policy.mjs"
		defaultScripts["agent-loop:review"] = "node scripts/check-agent-loop-policy.mjs && node scripts/check-artifact-manifest.mjs"
	}
	if mode != modelingModeOff {
		defaultScripts["artifacts:check"] = "node scripts/generate-artifact-review.mjs --check && node scripts/check-artifact-manifest.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/generate-model-review.mjs --check && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["artifacts:generate"] = "node scripts/generate-artifact-review.mjs && node scripts/generate-model-review.mjs"
		defaultScripts["artifacts:model:generate"] = "node scripts/generate-model-review.mjs"
		defaultScripts["artifacts:model:check"] = "node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/generate-model-review.mjs --check"
		defaultScripts["artifacts:model:drift"] = "node scripts/generate-model-review.mjs --check"
		defaultScripts["artifacts:model:review"] = "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["artifacts:review"] = "node scripts/generate-artifact-review.mjs && node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["models:generate"] = "node scripts/generate-model-review.mjs"
		defaultScripts["models:check"] = "node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/generate-model-review.mjs --check"
		defaultScripts["models:drift"] = "node scripts/generate-model-review.mjs --check"
		defaultScripts["models:diff:check"] = "node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["models:open"] = "node scripts/open-artifact-review.mjs generated/review/models/index.html"
		defaultScripts["models:review"] = "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-model-inventory.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs"
	}
	for name, command := range defaultScripts {
		if _, exists := scripts[name]; !exists {
			scripts[name] = command
		}
	}
	metadata["scripts"] = scripts
	metadata["devDependencies"] = devDependencies
	updated, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	updated = append(updated, '\n')
	return os.WriteFile(packagePath, updated, 0o644)
}

func ensureGitignoreLines(projectDir string, lines []string) error {
	path := filepath.Join(projectDir, ".gitignore")
	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	out := strings.TrimRight(existing, "\r\n")
	for _, line := range lines {
		if gitignoreHasLine(existing, line) {
			continue
		}
		if out != "" {
			out += "\n"
		}
		out += line
	}
	if out != "" {
		out += "\n"
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func gitignoreHasLine(text, expected string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}
	return false
}

func writeDeveloperArtifactScaffold(projectDir string, profile artifactProfile, agentDocsEnabled bool, beadsEnabled bool, mode modelingMode) error {
	if mode == modelingModeAuto {
		mode = modelingModeUMLFirst
	}
	effectiveProfile := effectiveArtifactProfile(profile)
	modelingEnabled := mode != modelingModeOff
	dirs := []string{
		filepath.Join(projectDir, "docs", "artifacts", "source"),
		filepath.Join(projectDir, "docs", "artifacts", "source", "business"),
		filepath.Join(projectDir, "docs", "artifacts", "source", "data"),
		filepath.Join(projectDir, "docs", "artifacts", "templates"),
		filepath.Join(projectDir, "docs", "artifacts", "source", "product"),
		filepath.Join(projectDir, "docs", "artifacts", "source", "research"),
		filepath.Join(projectDir, "docs", "artifacts", "source", "ux"),
		filepath.Join(projectDir, "generated", "review"),
		filepath.Join(projectDir, "generated", "review", "business"),
		filepath.Join(projectDir, "generated", "review", "data"),
		filepath.Join(projectDir, "generated", "review", "product"),
		filepath.Join(projectDir, "generated", "review", "research"),
		filepath.Join(projectDir, "generated", "review", "ux"),
		filepath.Join(projectDir, ".skill-harness"),
		filepath.Join(projectDir, "scripts"),
	}
	if profile == artifactProfileMedia {
		dirs = append(dirs, filepath.Join(projectDir, "generated", "media"))
	}
	if profile == artifactProfileAgentLoop {
		dirs = append(dirs, filepath.Join(projectDir, "generated", "agent-runs"))
	}
	if modelingEnabled {
		dirs = append(dirs,
			filepath.Join(projectDir, "docs", "artifacts", "source", "models"),
			filepath.Join(projectDir, "generated", "review", "models"),
		)
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	canonicalTooling := []string{"specgraph", "agent-docs"}
	if !agentDocsEnabled {
		canonicalTooling = []string{}
	}
	config := map[string]any{
		"version": 1,
		"capabilities": map[string]any{
			"developerArtifacts": map[string]any{
				"enabled":          true,
				"requestedProfile": string(profile),
				"profile":          string(effectiveProfile),
				"specialization":   artifactSpecialization(profile),
				"canonical": map[string]any{
					"formats": []string{"markdown", "toon", "json", "yaml"},
					"tooling": canonicalTooling,
					"paths":   []string{"docs", "docs/artifacts/source"},
				},
				"artifactTypes": artifactTypes(modelingEnabled),
				"manifest": map[string]any{
					"path":            "docs/artifacts/artifacts.manifest.json",
					"schemaVersion":   1,
					"requireSource":   true,
					"requireEvidence": false,
					"freshness": map[string]any{
						"trackSourceHash": true,
						"hashAlgorithm":   "sha256",
					},
				},
				"modeling":             artifactModelingConfig(mode),
				"modelPolicy":          artifactModelPolicy(mode),
				"visualArtifactPolicy": artifactVisualPolicy(),
				"infographicPolicy":    artifactInfographicPolicy(),
				"reviewSurface": map[string]any{
					"format":          "html",
					"outDir":          "generated/review",
					"commitGenerated": false,
					"openMode":        artifactOpenMode(profile),
					"openPolicy": map[string]any{
						"preferHostBrowserTool": true,
						"codexApp":              "use Browser plugin when available",
						"claudeDesktop":         "use built-in browser or preview tool when available",
						"cliFallback":           "system-default-browser",
						"headlessFallback":      "print-file-url",
					},
				},
				"mediaOutputs": artifactMediaOutputs(profile),
				"agentLoop":    artifactAgentLoopConfig(profile, beadsEnabled),
				"htmlPolicy": map[string]any{
					"role":                  "generated-review-surface",
					"selfContained":         true,
					"allowInlineJavaScript": false,
					"allowExternalScripts":  false,
					"allowExternalAssets":   false,
					"allowNetworkCalls":     false,
					"requireSemanticHTML":   true,
					"interactionLanes":      artifactHTMLInteractionLanes(),
					"requiredCSP":           artifactRequiredCSP(),
				},
			},
		},
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(projectDir, ".skill-harness", "project.json"), data, 0o644); err != nil {
		return err
	}
	if err := updatePackageScripts(projectDir, agentDocsEnabled, profile, mode); err != nil {
		return err
	}
	gitignoreLines := []string{"generated/review/"}
	if profile == artifactProfileMedia {
		gitignoreLines = append(gitignoreLines, "generated/media/")
	}
	if profile == artifactProfileAgentLoop {
		gitignoreLines = append(gitignoreLines, "generated/agent-runs/")
	}
	if err := ensureGitignoreLines(projectDir, gitignoreLines); err != nil {
		return err
	}

	readmePath := filepath.Join(projectDir, "docs", "artifacts", "README.md")
	if !fileExists(readmePath) {
		if err := os.WriteFile(readmePath, []byte(developerArtifactReadme(profile, mode)), 0o644); err != nil {
			return err
		}
	}

	templatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "review-artifact.md")
	if !fileExists(templatePath) {
		if err := os.WriteFile(templatePath, []byte(developerArtifactTemplate()), 0o644); err != nil {
			return err
		}
	}

	visualTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "visual-source-artifact.md")
	if !fileExists(visualTemplatePath) {
		if err := os.WriteFile(visualTemplatePath, []byte(developerVisualSourceArtifactTemplate()), 0o644); err != nil {
			return err
		}
	}

	atlasTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "e2e-product-system-atlas.md")
	if !fileExists(atlasTemplatePath) {
		if err := os.WriteFile(atlasTemplatePath, []byte(developerE2EProductSystemAtlasTemplate()), 0o644); err != nil {
			return err
		}
	}

	modelTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "model-artifact.md")
	if !fileExists(modelTemplatePath) {
		if err := os.WriteFile(modelTemplatePath, []byte(developerModelArtifactTemplate()), 0o644); err != nil {
			return err
		}
	}
	if modelingEnabled {
		modelInventoryPath := filepath.Join(projectDir, "docs", "artifacts", "source", "models", "model-inventory.md")
		if !fileExists(modelInventoryPath) {
			if err := os.WriteFile(modelInventoryPath, []byte(developerModelInventoryTemplate()), 0o644); err != nil {
				return err
			}
		}
		modelDiffTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "model-diff-artifact.md")
		if !fileExists(modelDiffTemplatePath) {
			if err := os.WriteFile(modelDiffTemplatePath, []byte(developerModelDiffArtifactTemplate()), 0o644); err != nil {
				return err
			}
		}
	}

	if profile == artifactProfileMedia {
		demoTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "demo-artifact.md")
		if !fileExists(demoTemplatePath) {
			if err := os.WriteFile(demoTemplatePath, []byte(developerDemoArtifactTemplate()), 0o644); err != nil {
				return err
			}
		}
	}

	if profile == artifactProfileAgentLoop {
		loopTemplatePath := filepath.Join(projectDir, "docs", "artifacts", "templates", "agent-loop-artifact.md")
		if !fileExists(loopTemplatePath) {
			if err := os.WriteFile(loopTemplatePath, []byte(developerAgentLoopArtifactTemplate()), 0o644); err != nil {
				return err
			}
		}
		playbookPath := filepath.Join(projectDir, "docs", "artifacts", "source", "agent-loop-playbook.md")
		if !fileExists(playbookPath) {
			if err := os.WriteFile(playbookPath, []byte(developerAgentLoopPlaybook()), 0o644); err != nil {
				return err
			}
		}
	}

	manifestPath := filepath.Join(projectDir, "docs", "artifacts", "artifacts.manifest.json")
	if !fileExists(manifestPath) {
		if err := os.WriteFile(manifestPath, []byte(developerArtifactManifest()), 0o644); err != nil {
			return err
		}
	}

	manifestCheckerPath := filepath.Join(projectDir, "scripts", "check-artifact-manifest.mjs")
	if !fileExists(manifestCheckerPath) {
		if err := os.WriteFile(manifestCheckerPath, []byte(developerArtifactManifestScript()), 0o644); err != nil {
			return err
		}
	}

	checkerPath := filepath.Join(projectDir, "scripts", "check-artifact-html-policy.mjs")
	if !fileExists(checkerPath) {
		if err := os.WriteFile(checkerPath, []byte(developerArtifactPolicyScript()), 0o644); err != nil {
			return err
		}
	}
	reviewGeneratorPath := filepath.Join(projectDir, "scripts", "generate-artifact-review.mjs")
	if !fileExists(reviewGeneratorPath) {
		if err := os.WriteFile(reviewGeneratorPath, []byte(developerArtifactReviewGeneratorScript()), 0o644); err != nil {
			return err
		}
	}
	uweWorkspaceRuntimePath := filepath.Join(projectDir, "scripts", "uwe-workspace-runtime.js")
	if !fileExists(uweWorkspaceRuntimePath) {
		if err := os.WriteFile(uweWorkspaceRuntimePath, []byte(developerArtifactUweWorkspaceRuntimeScript()), 0o644); err != nil {
			return err
		}
	}
	openerPath := filepath.Join(projectDir, "scripts", "open-artifact-review.mjs")
	if !fileExists(openerPath) {
		if err := os.WriteFile(openerPath, []byte(developerArtifactOpenScript()), 0o644); err != nil {
			return err
		}
	}
	if modelingEnabled {
		modelCheckerPath := filepath.Join(projectDir, "scripts", "check-model-artifact-policy.mjs")
		if !fileExists(modelCheckerPath) {
			if err := os.WriteFile(modelCheckerPath, []byte(developerModelArtifactPolicyScript()), 0o644); err != nil {
				return err
			}
		}
		modelInventoryCheckerPath := filepath.Join(projectDir, "scripts", "check-model-inventory.mjs")
		if !fileExists(modelInventoryCheckerPath) {
			if err := os.WriteFile(modelInventoryCheckerPath, []byte(developerModelInventoryPolicyScript()), 0o644); err != nil {
				return err
			}
		}
		modelReviewGeneratorPath := filepath.Join(projectDir, "scripts", "generate-model-review.mjs")
		if !fileExists(modelReviewGeneratorPath) {
			if err := os.WriteFile(modelReviewGeneratorPath, []byte(developerModelReviewGeneratorScript()), 0o644); err != nil {
				return err
			}
		}
	}

	if profile == artifactProfileAgentLoop {
		loopCheckerPath := filepath.Join(projectDir, "scripts", "check-agent-loop-policy.mjs")
		if !fileExists(loopCheckerPath) {
			return os.WriteFile(loopCheckerPath, []byte(developerAgentLoopPolicyScript()), 0o644)
		}
	}
	return nil
}

func effectiveArtifactProfile(profile artifactProfile) artifactProfile {
	switch profile {
	case artifactProfileAuto:
		return artifactProfileDual
	case artifactProfileCodexApp, artifactProfileClaudeDesktop:
		return artifactProfileHTML
	case artifactProfileCLI, artifactProfileTUI:
		return artifactProfileMarkdown
	case artifactProfileMedia, artifactProfileAgentLoop:
		return artifactProfileDual
	default:
		return profile
	}
}

func artifactOpenMode(profile artifactProfile) string {
	switch profile {
	case artifactProfileCodexApp, artifactProfileClaudeDesktop, artifactProfileHTML, artifactProfileMedia, artifactProfileAgentLoop:
		return "file-preview"
	case artifactProfileCLI, artifactProfileTUI, artifactProfileMarkdown:
		return "path-or-command"
	default:
		return "when-supported"
	}
}

func artifactSpecialization(profile artifactProfile) string {
	switch profile {
	case artifactProfileMedia:
		return "media-demo"
	case artifactProfileAgentLoop:
		return "self-improving-agent-loop"
	default:
		return ""
	}
}

func artifactMediaOutputs(profile artifactProfile) map[string]any {
	if profile != artifactProfileMedia {
		return map[string]any{
			"enabled": false,
		}
	}
	return map[string]any{
		"enabled":             true,
		"outDir":              "generated/media",
		"commitGenerated":     false,
		"sourceBacked":        true,
		"requireEvidence":     true,
		"defaultStatus":       "draft",
		"allowedStatuses":     []string{"draft", "needs-evidence", "approved", "rejected", "stale", "inconclusive"},
		"defaultExclusions":   []string{"trace.zip", "*.har", "network.json", "console.json", "page-errors.json"},
		"upstreamEvidence":    []string{"events.json", "verification.json", "quality.json", "qa-report.json", "review-bundle.json", "segment.evidence.json", "layout-safety.report.json"},
		"approvedOutputKinds": []string{"mp4", "webm", "gif", "webp", "poster-frame", "frame-strip", "html-review"},
	}
}

func artifactTypes(enableModeling bool) []string {
	types := []string{
		"decision",
		"plan",
		"spec",
		"product-brief",
		"e2e-product-system-atlas",
		"opportunity-brief",
		"business-case",
		"stakeholder-map",
		"data-dictionary",
		"metric-definition",
		"lineage-map",
		"research-synthesis",
		"claim-evidence-matrix",
		"high-fidelity-prototype",
		"interaction-state-board",
		"journey-map",
		"visual-review",
		"handoff",
		"evidence-pack",
		"blast-radius",
		"architecture-view",
		"model-view",
		"review-dashboard",
		"agent-loop",
		"trace-review",
		"eval-report",
		"learning-proposal",
	}
	if enableModeling {
		types = append(types, "model-diff")
	}
	return types
}

func artifactModelingConfig(mode modelingMode) map[string]any {
	if mode == modelingModeOff {
		return map[string]any{
			"mode":    string(modelingModeOff),
			"enabled": false,
		}
	}
	return map[string]any{
		"mode":                  string(mode),
		"enabled":               true,
		"defaultForFreshSetup":  mode == modelingModeUMLFirst,
		"autoDetectModelImpact": true,
		"canonicalInventory":    "docs/artifacts/source/models/model-inventory.md",
		"sourceDir":             "docs/artifacts/source/models",
		"reviewDir":             "generated/review/models",
		"humanReviewFormat":     "html",
	}
}

func artifactModelPolicy(mode modelingMode) map[string]any {
	policy := map[string]any{
		"canonicalSource":        true,
		"generatedReviewOnly":    true,
		"renderDiagramsOffline":  true,
		"defaultReviewEmbedding": "inline-svg",
		"allowedNotations":       []string{"mermaid", "markdown", "toon", "plantuml", "structurizr"},
		"allowedModelKinds": []string{
			"sequence",
			"state",
			"class",
			"domain",
			"context",
			"container",
			"component",
			"dynamic",
			"deployment",
			"dependency",
			"use-case",
			"activity",
			"architecture-space",
		},
		"c4": map[string]any{
			"notation":     "mermaid",
			"experimental": true,
			"levels":       []string{"context", "container", "component", "dynamic", "deployment"},
		},
	}
	if mode != modelingModeOff {
		policy["uml"] = map[string]any{
			"enabled":                 true,
			"mode":                    string(mode),
			"methods":                 []string{"uml", "uwe", "c4", "evidence"},
			"allowedSourceExtensions": []string{".md", ".toon", ".mmd", ".puml", ".plantuml", ".dsl", ".json", ".yaml", ".yml"},
			"defaultDiffMethod":       "source",
			"semanticDiff":            false,
			"sourceDir":               "docs/artifacts/source/models",
			"reviewDir":               "generated/review/models",
			"inventory":               "docs/artifacts/source/models/model-inventory.md",
			"generatedReviewRequired": mode == modelingModeUMLFirst,
			"reviewRequirement":       map[string]any{"baseline": "optional", "uml-first": "required-for-ready-model-artifacts"},
			"allowedFacets": map[string]any{
				"uwe": []string{"content", "navigation", "presentation", "process", "access", "adaptation"},
			},
			"methodModelKinds": map[string]any{
				"uml":      []string{"use-case", "activity", "sequence", "state", "class", "domain"},
				"uwe":      []string{"use-case", "activity", "sequence", "state", "domain", "component"},
				"c4":       []string{"context", "container", "component", "dynamic", "deployment", "architecture-space"},
				"evidence": []string{"dependency", "architecture-space", "class", "dynamic", "deployment"},
			},
			"evidenceDefaultKinds":    []string{"dependency"},
			"authoredOrEvidenceKinds": []string{"class", "dynamic"},
		}
	}
	return policy
}

func artifactVisualPolicy() map[string]any {
	return map[string]any{
		"enabled":             true,
		"doctrine":            "visual-source-first",
		"canonicalFirst":      true,
		"generatedReviewOnly": true,
		"defaultHumanSurface": "high-fidelity-html",
		"lowFidelityPolicy":   "scratch-only-not-canonical",
		"sourceDirs":          []string{"docs/artifacts/source/product", "docs/artifacts/source/business", "docs/artifacts/source/data", "docs/artifacts/source/research", "docs/artifacts/source/ux"},
		"reviewDirs":          []string{"generated/review/product", "generated/review/business", "generated/review/data", "generated/review/research", "generated/review/ux"},
		"families": []map[string]any{
			{
				"name":          "product",
				"sourceDir":     "docs/artifacts/source/product",
				"reviewDir":     "generated/review/product",
				"sourceKinds":   []string{"prd", "opportunity-brief", "feature-map", "acceptance-criteria", "roadmap", "e2e-product-system-atlas"},
				"reviewKinds":   []string{"product-brief", "feature-map", "decision-dashboard", "uwe-navigation-atlas"},
				"primaryAgents": []string{"requirements-analyst", "delivery-manager", "system-modeler", "ux-researcher", "test-designer"},
			},
			{
				"name":          "business",
				"sourceDir":     "docs/artifacts/source/business",
				"reviewDir":     "generated/review/business",
				"sourceKinds":   []string{"business-model", "pricing-assumptions", "stakeholder-map", "risk-register"},
				"reviewKinds":   []string{"strategy-review", "assumption-dashboard", "stakeholder-map"},
				"primaryAgents": []string{"requirements-analyst", "delivery-manager"},
			},
			{
				"name":          "data",
				"sourceDir":     "docs/artifacts/source/data",
				"reviewDir":     "generated/review/data",
				"sourceKinds":   []string{"schema", "data-dictionary", "metric-definition", "lineage-map", "quality-rules"},
				"reviewKinds":   []string{"schema-map", "metric-dashboard", "data-quality-review"},
				"primaryAgents": []string{"backend-engineer", "test-designer"},
			},
			{
				"name":          "research",
				"sourceDir":     "docs/artifacts/source/research",
				"reviewDir":     "generated/review/research",
				"sourceKinds":   []string{"claim-evidence-matrix", "literature-theme-map", "interview-synthesis", "assumption-register"},
				"reviewKinds":   []string{"research-board", "evidence-map", "confidence-dashboard"},
				"primaryAgents": []string{"research-writer", "ux-researcher"},
			},
			{
				"name":          "ux",
				"sourceDir":     "docs/artifacts/source/ux",
				"reviewDir":     "generated/review/ux",
				"sourceKinds":   []string{"design-brief", "component-state-spec", "interaction-flow", "prototype-source"},
				"reviewKinds":   []string{"high-fidelity-prototype", "state-board", "journey-map", "accessibility-review"},
				"primaryAgents": []string{"ux-researcher", "web-engineer"},
			},
		},
		"agentTeam": []map[string]any{
			{"agent": "research-writer", "owns": []string{"evidence", "citations", "claim strength", "research gaps"}},
			{"agent": "requirements-analyst", "owns": []string{"product intent", "requirements", "acceptance criteria", "assumptions"}},
			{"agent": "delivery-manager", "owns": []string{"business viability", "stakeholder impact", "risk and rollout"}},
			{"agent": "backend-engineer", "owns": []string{"data structures", "schemas", "integrity constraints"}},
			{"agent": "ux-researcher", "owns": []string{"task evidence", "prototype critique", "high-fidelity review"}},
			{"agent": "system-modeler", "owns": []string{"model impact", "workflow and structure diagrams"}},
			{"agent": "quality-reviewer", "owns": []string{"evidence gates", "freshness", "readiness risks"}},
		},
		"readinessGates": []string{
			"canonical source exists before generated review",
			"manifest entry links source, review surface, owner, evidence, and freshness",
			"human review surface is visual when product, business, data, research, or UX comprehension benefits from layout",
			"source-declared infographic specs render to static review markup without browser runtimes",
			"high-fidelity is required for UI and customer-facing workflow review",
			"synthetic user or agent simulation evidence is labelled separately from real user evidence",
			"HTML policy passes before handoff",
			"UWE navigation atlases cover navigable nodes, actions, screenshots, side effects, and untested branches without claiming exhaustive state-space coverage",
		},
	}
}

func artifactInfographicPolicy() map[string]any {
	return map[string]any{
		"enabled":              true,
		"defaultMode":          "source-spec-to-static-review",
		"browserRuntime":       "blocked-by-default-html-policy",
		"specFence":            "artifact-infographic",
		"allowedStaticOutputs": []string{"inline-svg", "static-html", "data-url-image"},
		"tools": []map[string]any{
			{"id": "mermaid", "label": "Mermaid", "role": "architecture, workflow, sequence, and model diagrams", "output": "pre-rendered inline SVG or static markup"},
			{"id": "vega-lite", "label": "Vega-Lite", "role": "default declarative charts for metrics, comparisons, and evidence dashboards", "output": "static SVG generated from source specs"},
			{"id": "observable-plot", "label": "Observable Plot", "role": "compact exploratory charts and statistical views", "output": "static SVG generated from source specs"},
			{"id": "d3", "label": "D3", "role": "custom infographic layouts when canned charts are not expressive enough", "output": "static SVG generated during artifact generation"},
			{"id": "graphviz", "label": "Graphviz", "role": "node-edge dependency, lineage, and relationship maps", "output": "static SVG generated from DOT or structured edges"},
			{"id": "echarts", "label": "Apache ECharts", "role": "dashboard-style chart families when a richer chart grammar is useful", "output": "static SVG or PNG generated outside the browser runtime"},
			{"id": "rawgraphs", "label": "RAWGraphs", "role": "design-led or unusual infographic forms using tabular data", "output": "exported SVG copied into the generated review surface"},
			{"id": "chartjs", "label": "Chart.js", "role": "simple familiar charts when existing source data already matches Chart.js conventions", "output": "server-rendered image or static SVG equivalent"},
		},
		"selectionRules": []string{
			"use Mermaid for authored architecture/process/model diagrams already represented as Mermaid text",
			"use Vega-Lite as the default chart grammar for source-backed metrics and comparison charts",
			"use Observable Plot for compact exploratory or statistical chart specs",
			"use D3 when the artifact needs a bespoke static infographic layout",
			"use Graphviz for dependency, lineage, and relationship graphs",
			"use ECharts only as a generation-time renderer or static equivalent, not as a browser runtime",
			"use RAWGraphs for design-led exported SVGs with tabular source data",
			"use Chart.js only through server-rendered/static output or an equivalent static chart",
		},
	}
}

func artifactAgentLoopConfig(profile artifactProfile, beadsEnabled bool) map[string]any {
	if profile != artifactProfileAgentLoop {
		return map[string]any{
			"enabled": false,
		}
	}
	defaultIssueTool := "explicit-human-request"
	learningOutputs := []string{
		"follow-up issue or source artifact",
		"agent-loop artifact",
		"skill or loadout change proposal",
		"regression test or checker update",
	}
	if beadsEnabled {
		defaultIssueTool = "beads"
		learningOutputs = append([]string{"beads issue", "bd remember insight"}, learningOutputs...)
	}
	return map[string]any{
		"enabled":          true,
		"loopName":         "self-improving-agent-loop",
		"traceDir":         "generated/agent-runs",
		"playbook":         "docs/artifacts/source/agent-loop-playbook.md",
		"defaultIssueTool": defaultIssueTool,
		"ownerModel":       "human-dri-plus-agent-team",
		"agents": []map[string]any{
			{
				"name": "research-model-agent",
				"recommendedLoadouts": []string{
					"research-writer",
					"requirements-analyst-beads",
				},
				"owns": []string{"external-research", "internal-context-map", "problem-framing", "candidate-loop-brief"},
			},
			{
				"name": "workflow-loop-agent",
				"recommendedLoadouts": []string{
					"workflow-engineer",
					"software-architect-beads",
				},
				"owns": []string{"implementation-slice", "quality-gates", "trace-review", "learning-proposal"},
			},
		},
		"phases": []string{
			"sense",
			"model",
			"plan",
			"act",
			"gate",
			"learn",
		},
		"sensors": []string{
			"beads issues",
			"git diffs",
			"test and build output",
			"artifact manifests",
			"agent handoff notes",
			"runtime traces",
		},
		"qualityGates": []string{
			"issue or explicit request captured before mutation",
			"tests or executable validators run",
			"artifact manifest policy passes",
			"HTML policy passes for generated review surfaces",
			"human approval for irreversible or high-risk actions",
		},
		"learningOutputs": learningOutputs,
		"humanApprovalRequiredFor": []string{
			"secret or credential handling",
			"production data access",
			"permission expansion",
			"destructive filesystem or database action",
			"network-visible publishing",
			"autonomous merge or deployment",
		},
		"riskControls": []string{
			"least-privilege tools",
			"bounded token and time budgets",
			"reversible first slice",
			"trace receipts for agent runs",
			"explicit rollback or follow-up issue",
		},
	}
}

func artifactRequiredCSP() string {
	return "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
}

func artifactHTMLInteractionLanes() map[string]any {
	return map[string]any{
		"default": map[string]any{
			"name":                  "css-only",
			"allowInlineJavaScript": false,
			"allowedPatterns":       []string{"radio tabs", "details summary", "anchor navigation", "inline svg states"},
			"csp":                   artifactRequiredCSP(),
		},
		"reviewed-inline-js": map[string]any{
			"name":                   "reviewed-inline-js",
			"allowInlineJavaScript":  false,
			"requiresManifestFlag":   "htmlInteractionLane: reviewed-inline-js",
			"requiresHumanReview":    true,
			"requiresCheckerSupport": true,
			"requiresCspChange":      true,
			"blockedApis":            []string{"fetch", "XMLHttpRequest", "WebSocket", "EventSource", "sendBeacon", "serviceWorker", "document.cookie", "localStorage", "sessionStorage"},
			"status":                 "reserved-not-enabled",
		},
	}
}

func developerArtifactReadme(profile artifactProfile, mode modelingMode) string {
	effectiveProfile := effectiveArtifactProfile(profile)
	modelingSection := ""
	if mode != modelingModeOff {
		modelingSection = fmt.Sprintf(`
## Source-First Modeling

- Modeling mode: %s.
- Auto-detect model impact for engineering changes. If code, API, workflow, dependency, deployment, or UX structure changes, update the relevant model source or record why no model change is needed.
- Use source/models/ as the default home for canonical UML/UWE/C4/evidence model sources when no domain-specific docs folder is better.
- Use generated/review/models/ for generated human HTML before/after review surfaces.
- Keep model-view and model-diff entries in artifacts.manifest.json; the manifest carries modelId, method, facets, lineage, diff metadata, evidence links, renderer data, and source hashes.
- Treat source diffs as canonical. HTML, SVG, PNG, and screenshots are review surfaces only.
- UWE facets are content, navigation, presentation, process, access, and adaptation. Access is the local security/access-control facet; adaptation covers personalization/context variation.
- Keep model-inventory.md current as the canonical index of model ids, owners, sources, generated reviews, and implementation touchpoints.
- Run node scripts/generate-model-review.mjs to refresh static HTML review pages for humans.
- Run node scripts/check-model-artifact-policy.mjs before handing off model-backed engineering artifacts.
`, mode)
	}
	return fmt.Sprintf(`# Developer Artifacts

Requested profile: %s
Effective profile: %s

Use this directory for durable developer artifacts and generated review surfaces.

## Source Of Truth

- Keep canonical decisions, specs, investigations, product briefs, business notes, data definitions, research syntheses, UX flows, and handoff notes in Markdown, TOON, JSON/YAML, or specgraph-compatible sources.
- Treat HTML as a generated review surface for scanning, comparison, diagrams, dashboards, prototypes, mockups, and desktop app previews.
- Do not make generated HTML the only durable source for a decision.
- Record source-backed review surfaces in artifacts.manifest.json so agents and humans can detect stale output.
- For human-facing discovery, planning, product, business, data, research, UX, and mockup artifacts, set reviewRequired: true in artifacts.manifest.json and generate infographic HTML with node scripts/generate-artifact-review.mjs.

## Layout

- source/ - canonical artifact sources when they do not belong in a domain-specific docs folder.
- source/product/ - product briefs, feature maps, roadmaps, and acceptance matrices.
- source/business/ - business models, pricing assumptions, stakeholder maps, and risk registers.
- source/data/ - schemas, data dictionaries, metric definitions, lineage, and quality rules.
- source/research/ - claim-evidence matrices, literature maps, interviews, and assumption registers.
- source/ux/ - design briefs, interaction flows, component states, and prototype sources.
- templates/ - local templates for recurring artifact types.
- artifacts.manifest.json - provenance and freshness index for source-backed review artifacts.
- ../../generated/review/ - generated HTML or rich review artifacts for humans.
- ../../generated/review/product/ - generated product review surfaces.
- ../../generated/review/business/ - generated business review surfaces.
- ../../generated/review/data/ - generated data review surfaces.
- ../../generated/review/research/ - generated research review surfaces.
- ../../generated/review/ux/ - generated UX mockups, prototypes, and state boards.
- ../../generated/media/ - generated demo media for media profile projects.
- ../../generated/agent-runs/ - generated trace receipts and eval summaries for agent-loop profile projects.

## Visual Source-First Policy

- Use visual-source-first artifacts for product, business, data, research, and UX work when humans need to inspect structure, evidence, states, tradeoffs, or mockups.
- Keep source artifacts agent-readable and diffable. Generated HTML, screenshots, videos, SVGs, PNGs, and comparison pages are review surfaces only.
- High-fidelity HTML is the default human review surface for UI, customer-facing workflow, product, and mockup reviews. Low-fidelity sketches are scratch only and should not become canonical approval surfaces.
- Visual review surfaces should show realistic data, states, error paths, assumptions, evidence strength, source links, and freshness metadata.
- Non-UI human review surfaces should be infographic-style by default: summary metrics, charts or timelines, evidence/freshness panels, review verdicts, and links back to source.
- Label synthetic user, simulated customer, or agent-generated evidence separately from real user or customer evidence.
- Use a team of agents when ownership crosses a real boundary: requirements-analyst for product intent, delivery-manager for business constraints, backend-engineer for data shape, research-writer for evidence, ux-researcher for high-fidelity UX review, system-modeler for structural impact, and quality-reviewer for readiness gates.
- Record every durable generated visual artifact in artifacts.manifest.json with source, reviewSurface, owner, evidenceLinks, status, and freshness.

## Model And Diagram Policy

- Keep Mermaid, C4, UML-style, and architecture-space sources in Markdown, TOON, or specgraph-compatible source artifacts.
- Pre-render diagrams into generated HTML as inline SVG or static markup; do not load a browser Mermaid runtime by default.
- Treat Mermaid C4 as a review notation and record the level explicitly: context, container, component, dynamic, or deployment.
- Treat dependency graphs as generated evidence unless the project has a separate model source of truth.
- Link every generated model view back to its source artifact, issue, and evidence.
%s

## Media And Demo Policy

- Keep .demo.yaml, QA flows, reports, and source notes as canonical artifacts.
- Treat MP4s, GIF/WebP previews, poster frames, and frame strips as generated outputs.
- Store generated media under generated/media/ for media profile projects and keep it out of git by default.
- Do not turn failed or inconclusive QA evidence into an approved product demo.
- Exclude raw traces, HAR/network dumps, console logs, page errors, secrets, and customer data from handoff bundles unless explicitly redacted and approved.

## Agent Loop Policy

- Start from a Beads issue or an explicit human request before changing files.
- Keep the durable loop playbook in source/agent-loop-playbook.md for agent-loop profile projects.
- Store trace receipts, eval summaries, and run evidence under generated/agent-runs/ and keep them out of git by default.
- Treat learning outputs as proposals until tests, policy checks, and the human DRI approve high-risk changes.
- Record reusable lessons with the project memory mechanism instead of unmanaged memory files.
- Do not expand tool permissions, publish, deploy, or run irreversible actions without explicit human approval.

## HTML Review Policy

- Self-contained static HTML only by default.
- No external scripts, external assets, or network calls unless the project explicitly opts in.
- Default interaction is CSS/HTML only: radio tabs, details/summary, anchor navigation, and inline SVG states.
- Do not use inline JavaScript. The reviewed inline-JS lane is reserved until manifest metadata, CSP/checker support, and human approval requirements are implemented together.
- Every HTML review artifact must include the required CSP meta tag from .skill-harness/project.json.
- Use semantic headings, landmarks, meaningful link text, and alt text for embedded images.
- No secrets, credentials, tokens, private logs, or customer data.
- Link back to the canonical source artifact and issue.
- Regenerate or discard HTML when the source changes.
- Open generated HTML with the best human review surface for the current environment: Codex Browser plugin in Codex app, Claude desktop preview/browser in Claude desktop, or node scripts/open-artifact-review.mjs for CLI/system-browser fallback.

Run this policy check before handing off generated HTML:

    node scripts/check-artifact-manifest.mjs
    node scripts/generate-artifact-review.mjs --check
    node scripts/check-artifact-html-policy.mjs
    node scripts/open-artifact-review.mjs --print
`, profile, effectiveProfile, modelingSection)
}

func developerArtifactTemplate() string {
	return `# Review Artifact: [Title]

**Status:** Draft
**Canonical source:** [path or issue]
**Review surface:** Markdown | HTML | Dual

## Purpose

What decision, review, or handoff this artifact supports.

## Evidence

- Source files:
- Issues:
- Test or verification output:

## Summary

The smallest useful explanation for a reviewer.

## Follow-Up

- [ ] Action item
`
}

func developerVisualSourceArtifactTemplate() string {
	return `# Visual Source Artifact: [Title]

**Status:** Draft
**Artifact type:** product-brief | business-case | data-dictionary | research-synthesis | high-fidelity-prototype | visual-review
**Family:** product | business | data | research | ux
**Canonical source:** [docs/artifacts/source/family/path.md]
**Generated review:** [generated/review/family/path.html]
**Owner agent:** requirements-analyst | ux-researcher | research-writer | backend-engineer | delivery-manager | system-modeler | quality-reviewer

## Purpose

What product, business, data, research, or UX decision this visual review supports.

## Source Contract

- Source format:
- Structured data or model inputs:
- Update rule: edit source first, regenerate review second.

## Visual Review Surface

- Human surface type: high-fidelity HTML | dashboard | state board | journey map | prototype | evidence board
- Required states or views:
- Accessibility and readability checks:

## Evidence

- Real user or customer evidence:
- Synthetic user or agent simulation evidence:
- Data, metrics, or source files:
- Issues, tests, or review notes:

## Readiness

- Manifest entry:
- Source hash:
- Generated at:
- Open risks:
`
}

func developerE2EProductSystemAtlasTemplate() string {
	return `# E2E Product System Atlas: [App Name]

**Status:** Draft
**Artifact type:** e2e-product-system-atlas
**Family:** product
**Model method:** UWE navigation model with screenshot-backed evidence
**Canonical source:** docs/artifacts/source/product/[app]-e2e-product-system-atlas.md
**Generated review:** generated/review/product/[app]-e2e-product-system-atlas.html
**Owner agents:** requirements-analyst, system-modeler, ux-researcher, test-designer, web-engineer, backend-engineer, software-architect, security-reviewer, quality-reviewer

## Purpose

Create a source-first atlas for inspecting the whole app from landing page to deployed workload behavior. The review surface should show the UWE navigation structure, screenshots for navigable nodes, manual QA evidence for actions, and runtime side effects.

## Scope

- Product boundary:
- Deployed target or environment:
- User roles:
- Included entry points:
- Excluded or unreachable areas:
- Authorization and data-safety limits:

## UWE Navigation Nodes

| Node ID | Route or state | Role(s) | Screenshot evidence | Primary actions | Expected side effects |
| --- | --- | --- | --- | --- | --- |
| landing | / | anonymous | generated/review/evidence/[app]/landing.png | sign in, sign up, browse | session unchanged |

## Navigation Links

artifact-infographic:

    {
      "title": "UWE Navigation Graph",
      "tool": "graphviz",
      "kind": "uwe-navigation",
      "summary": "Navigable app nodes grouped by UWE navigation class with screenshots embedded inside the UWE navigation nodes. Keep this bounded, not a giant whole-system UML diagram.",
      "navigationClasses": [
        "Visitor acquisition and access",
        "Authenticated app flow",
        "Utilities and admin"
      ],
      "nodes": [
        {
          "id": "Landing",
          "label": "Landing",
          "route": "/",
          "navigationClass": "Visitor acquisition and access",
          "facet": "navigation",
          "role": "anonymous",
          "effect": "session unchanged",
          "screenshot": "generated/review/evidence/[app]/landing.png"
        },
        {
          "id": "Auth",
          "label": "Auth",
          "route": "/login",
          "navigationClass": "Visitor acquisition and access",
          "facet": "access",
          "role": "anonymous",
          "effect": "session created on success",
          "screenshot": "generated/review/evidence/[app]/auth.png"
        },
        {
          "id": "Dashboard",
          "label": "Dashboard",
          "route": "/app",
          "navigationClass": "Authenticated app flow",
          "facet": "content",
          "role": "member",
          "effect": "account/project data read",
          "screenshot": "generated/review/evidence/[app]/dashboard.png"
        }
      ],
      "edges": [
        ["Landing", "Auth", "sign in"],
        ["Auth", "Dashboard", "valid session"],
        ["Dashboard", "Primary Workflow", "primary action"],
        ["Primary Workflow", "Result State", "success"]
      ]
    }

## Action And Side-Effect Matrix

| Action ID | Node | Trigger | Expected UI result | Data effect | Runtime effect | Evidence | Verdict |
| --- | --- | --- | --- | --- | --- | --- | --- |
| ACT-001 | landing | click sign in | auth form visible | none | route transition only | screenshot + manual QA note | untested |

## Manual QA Sequence

1. Inventory public routes and capture desktop/mobile screenshots.
2. Authenticate with each authorized role and capture post-login navigation nodes.
3. Exercise every primary visible action once with safe test data.
4. Exercise important invalid, empty, denied, and recovery paths.
5. Record data, event, job, email, webhook, or deployed workload side effects.
6. Mark untested branches explicitly instead of implying full coverage.

## Deployment And Runtime Evidence

| Area | Evidence to capture | Notes |
| --- | --- | --- |
| Deployed URL | URL, commit, build id | avoid secrets |
| Health | health check, uptime check, smoke result | link logs only after redaction |
| Data stores | tables/collections touched | use test data |
| Jobs/events | queue/event/log observation | no raw private logs |
| Integrations | outbound calls/webhooks/emails | redact tokens and customer data |

## Screenshot Manifest

List screenshots in docs/artifacts/artifacts.manifest.json under screenshots, images, or visualEvidence. Generated HTML embeds small local images as data URLs.

json:

    {
      "screenshots": [
        {
          "path": "generated/review/evidence/[app]/landing.png",
          "caption": "Landing page",
          "alt": "Landing page screenshot"
        }
      ]
    }

## Readiness Gate

- Canonical source exists and names scope, roles, and exclusions.
- UWE navigation nodes cover all known routable or user-reachable states.
- Each primary action has expected UI result, side effect, evidence, and verdict.
- Screenshots are local, redacted, and linked from the manifest.
- Runtime claims are backed by logs, traces, health checks, tests, or deployment metadata.
- Untested branches are labelled untested or inconclusive.
- Generated HTML passes manifest, drift, and HTML policy checks.
`
}

func developerModelArtifactTemplate() string {
	return `# Model Artifact: [Title]

**Status:** Draft
**Artifact type:** model-view
**Model ID:** [stable-model-id]
**Model kind:** sequence | state | class | domain | context | container | component | dynamic | deployment | dependency | use-case | activity | architecture-space
**Notation:** mermaid | markdown | toon | plantuml | structurizr
**Method:** uml | uwe | c4 | evidence
**Facets:** [uwe: content, navigation, presentation, process, access, adaptation]
**Canonical source:** [path or issue]
**Generated review:** [generated/review/path.html]

## Purpose

What system behavior, boundary, dependency, or architecture question this model helps review.

## Scope

- System or subsystem:
- Abstraction level: domain | design | runtime | deployment | decision
- Owner:

## Source Model

Keep diagram source here or link to the canonical source artifact. Generated HTML should render from this source.

## Evidence

- Specs:
- Code:
- Tests:
- Runtime evidence:

## Freshness

- Source hash:
- Renderer and version:
- Last generated:
`
}

func developerModelInventoryTemplate() string {
	return `# Model Inventory

This is the canonical index for source-backed system models and human HTML review surfaces.

## Rules

- Auto-detect model impact for each engineering change.
- Update the relevant canonical model source when code, API, workflow, dependency, deployment, or UX structure changes.
- Keep generated HTML under ` + "`generated/review/models/`" + ` and treat it as a review surface, not source of truth.
- Record every durable model in ` + "`docs/artifacts/artifacts.manifest.json`" + ` with source, modelId, method, modelKind, owner, and reviewSurface.

## Inventory

| Model ID | Kind | Method | Canonical Source | Human HTML Review | Owner | Implementation Touchpoints |
| --- | --- | --- | --- | --- | --- | --- |
| example-system-context | context | c4 | docs/artifacts/source/models/example-system-context.md | generated/review/models/example-system-context.html | system-modeler | apps, packages, deployment |

## Change Log

| Date | Model ID | Change | Source/Evidence |
| --- | --- | --- | --- |
`
}

func developerModelDiffArtifactTemplate() string {
	return `# Model Diff Artifact: [Title]

**Status:** Draft
**Artifact type:** model-diff
**Model ID:** [stable-model-id]
**Canonical source:** [path to source diff note or model source]
**Before artifact:** [manifest artifact id]
**After artifact:** [manifest artifact id]
**Diff method:** source | semantic
**Generated review:** [generated/review/models/path.html]

## Purpose

What changed in the model, why it changed, and what engineering decision this review supports.

## Source Diff

The git/source diff is canonical. Link the model source files and summarize the meaningful change here.

## Before And After

- Before model:
- After model:
- Rendered before:
- Rendered after:

## Evidence

- Issue:
- Specs:
- Code:
- Tests or traces:

## Residual Risks

- Stale implementation risk:
- Missing evidence:
- Human review focus:

## Freshness

- Source hash:
- Renderer and version:
- Last generated:
`
}

func developerDemoArtifactTemplate() string {
	return `# Demo Artifact: [Title]

**Status:** Draft
**Artifact type:** plan | handoff | evidence-pack | review-dashboard
**Audience:** review | docs | release | social | repro
**Canonical source:** [path to .demo.yaml, QA flow, report, or source artifact]
**Run directory:** [path]
**Generated media:** [generated/media/path]
**Generated review:** [generated/review/path.html]

## Purpose

What this demo, silent cut, slideshow, or repro clip is meant to prove.

## Source And Evidence

- Demo spec:
- QA report:
- Events:
- Verification:
- Quality:
- Analyzer evidence:

## Media Outputs

- MP4:
- Poster:
- Frame strip:
- Review surface:

## Safety And Promotion

- Verdict source:
- Exclusions reviewed:
- Redactions:
- Promotion destination:
`
}

func developerAgentLoopArtifactTemplate() string {
	return `# Agent Loop Artifact: [Title]

**Status:** Draft
**Artifact type:** agent-loop | trace-review | eval-report | learning-proposal
**Issue:** [beads issue id]
**Human DRI:** [name]
**Primary agents:** research-model-agent | workflow-loop-agent
**Trace directory:** [generated/agent-runs/path]

## Purpose

What workflow, failure pattern, or improvement loop this artifact supports.

## Loop Definition

- Sensor:
- Policy:
- Tools:
- Quality gate:
- Learning output:

## Evidence

- Issues:
- Source files:
- Tests or validators:
- Trace receipts:
- Human approvals:

## Findings

- What worked:
- What failed:
- Missing tool, skill, data, or policy:
- Regression risk:

## Learning Proposal

- Proposed change:
- Expected measurable improvement:
- Required gate before adoption:
- Rollback:
`
}

func developerAgentLoopPlaybook() string {
	return `# Agent Loop Playbook

This project uses the skill-harness agent-loop profile for governed, self-improving agent workflows.

## Operating Model

Use two agents by default:

- Research/model agent: gathers external and internal evidence, maps the current workflow, and frames candidate loop improvements.
- Workflow/loop agent: implements the smallest reversible slice, runs gates, records evidence, and proposes durable learning.

The human DRI owns scope, risk acceptance, permission expansion, and final adoption.

## Loop Shape

1. Sense: gather issues or explicit requests, git diffs, test output, artifact manifests, traces, and handoff notes.
2. Model: identify the task type, expected trace shape, quality bar, and current failure mode.
3. Plan: choose one reversible improvement with explicit done criteria.
4. Act: make the smallest scoped change using existing repo patterns.
5. Gate: run tests, artifact checks, and risk checks before claiming improvement.
6. Learn: file follow-up issues or source artifacts, record durable memories with the project memory mechanism, and propose skill/loadout/checker updates only when evidence supports them.

## Policy Boundaries

Agents may propose improvements to prompts, skills, docs, tests, manifests, and workflow scripts.

Agents must require human approval before:

- expanding tool permissions
- handling secrets or production data
- deploying, publishing, or merging autonomously
- running destructive actions
- changing security, privacy, or compliance policy

## Trace Receipts

Store generated run evidence under generated/agent-runs/. A useful receipt captures:

- issue or request id and human DRI
- agents used
- tools called
- files changed
- gates run
- failures observed
- learning proposed
- token or time budget notes when available

Keep generated receipts out of git by default. Promote only summarized, redacted evidence into durable docs, the issue tracker, or project memory when it matters.
`
}

func developerArtifactManifest() string {
	return `{
  "version": 1,
  "rules": {
    "editSourceFirst": true,
    "generatedReviewIsCanonical": false,
    "hashAlgorithm": "sha256"
  },
  "artifacts": []
}
`
}

func developerArtifactManifestScript() string {
	return `import crypto from 'node:crypto';
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

function relativeForMessage(filePath) {
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
  failures.push('missing artifact manifest: ' + relativeForMessage(manifestPath));
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
      if (typeof artifact[field] !== 'string' || artifact[field].trim() === '') {
        failures.push(label + ' missing required string field: ' + field);
      }
    }
    if (artifact.type && allowedTypes.size > 0 && !allowedTypes.has(artifact.type)) {
      failures.push(label + ' has unsupported type: ' + artifact.type);
    }
    if (artifact.modelKind && allowedModelKinds.size > 0 && !allowedModelKinds.has(artifact.modelKind)) {
      failures.push(label + ' has unsupported modelKind: ' + artifact.modelKind);
    }
    if (artifact.notation && allowedNotations.size > 0 && !allowedNotations.has(artifact.notation)) {
      failures.push(label + ' has unsupported notation: ' + artifact.notation);
    }

    const sourcePath = resolveInsideRoot(artifact.source, label + '.source', failures);
    if (sourcePath && !fs.existsSync(sourcePath)) {
      failures.push(label + ' source does not exist: ' + artifact.source);
    }
    if (sourcePath && artifact.sourceHash) {
      const actualHash = hashFile(sourcePath);
      if (artifact.sourceHash !== actualHash) {
        failures.push(label + ' sourceHash is stale for ' + artifact.source);
      }
    }

    if (artifact.reviewSurface) {
      const reviewPath = resolveInsideRoot(artifact.reviewSurface, label + '.reviewSurface', failures);
      if (reviewPath && path.extname(reviewPath) === '.html' && !reviewPath.startsWith(reviewRoot + path.sep)) {
        failures.push(label + ' HTML review surface must be under ' + relativeForMessage(reviewRoot));
      }
      if (artifact.status === 'ready' && reviewPath && !fs.existsSync(reviewPath)) {
        failures.push(label + ' ready review surface does not exist: ' + artifact.reviewSurface);
      }
    } else if (artifact.status === 'ready' && artifact.reviewRequired === true) {
      failures.push(label + ' ready artifact with reviewRequired=true needs a generated HTML reviewSurface');
    }

    if (artifact.status === 'ready' && artifact.reviewRequired === true && artifact.reviewSurface && path.extname(artifact.reviewSurface) !== '.html') {
      failures.push(label + ' ready artifact with reviewRequired=true must use an HTML reviewSurface');
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
`
}

func developerArtifactReviewGeneratorScript() string {
	return `import crypto from 'node:crypto';
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
    .map((match) => ({ level: match[1].length, text: match[2].trim() }));
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

function stringList(value) {
  if (Array.isArray(value)) return value.map((item) => String(item)).filter(Boolean);
  if (typeof value === 'string') return value.split(/\s*,\s*/).filter(Boolean);
  if (value === undefined || value === null) return [];
  return [String(value)];
}

function normalizeBounds(value) {
  if (!value) return null;
  const raw = Array.isArray(value)
    ? { x: value[0], y: value[1], w: value[2], h: value[3] }
    : value;
  const bounds = {
    x: Number(raw.x ?? raw.left ?? 0),
    y: Number(raw.y ?? raw.top ?? 0),
    w: Number(raw.w ?? raw.width ?? 0),
    h: Number(raw.h ?? raw.height ?? 0),
    units: String(raw.units ?? 'relative')
  };
  if (![bounds.x, bounds.y, bounds.w, bounds.h].every(Number.isFinite) || bounds.w <= 0 || bounds.h <= 0) return null;
  return bounds;
}

function normalizeAnnotation(annotation, index) {
  const bounds = normalizeBounds(annotation?.bounds ?? annotation?.rect ?? annotation?.crop);
  if (!bounds) return null;
  return {
    id: String(annotation.id ?? 'ann-' + (index + 1)),
    kind: String(annotation.kind ?? annotation.type ?? 'highlight'),
    label: String(annotation.label ?? annotation.caption ?? 'Evidence focus'),
    color: String(annotation.color ?? '#e11d48'),
    bounds,
    relatesTo: annotation.relatesTo ?? {},
    semantics: String(annotation.semantics ?? 'evidence-only')
  };
}

function normalizeEvidence(evidence, index) {
  const annotations = Array.isArray(evidence?.annotations)
    ? evidence.annotations.map(normalizeAnnotation).filter(Boolean)
    : [];
  return {
    id: String(evidence?.id ?? 'ev-' + (index + 1)),
    kind: String(evidence?.kind ?? evidence?.type ?? 'screenshot'),
    path: evidence?.path || evidence?.screenshot || evidence?.image || '',
    sourceRef: evidence?.sourceRef ? String(evidence.sourceRef) : '',
    state: evidence?.state ? String(evidence.state) : '',
    viewport: evidence?.viewport ? String(evidence.viewport) : '',
    caption: evidence?.caption || evidence?.label ? String(evidence.caption ?? evidence.label) : '',
    primaryFor: stringList(evidence?.primaryFor ?? evidence?.nodeId ?? evidence?.node),
    crop: normalizeBounds(evidence?.crop),
    annotations
  };
}

function graphData(spec) {
  const evidence = Array.isArray(spec.evidence) ? spec.evidence.map(normalizeEvidence).filter(Boolean) : [];
  const evidenceById = new Map(evidence.map((item) => [item.id, item]));
  const resolvedEvidence = evidence.map((item) => {
    const source = item.sourceRef ? evidenceById.get(item.sourceRef) : null;
    return { ...item, path: item.path || source?.path || '', annotations: item.annotations.length > 0 ? item.annotations : (source?.annotations ?? []) };
  });
  const edges = Array.isArray(spec.edges) ? spec.edges.map((edge) => Array.isArray(edge)
    ? { id: '', from: String(edge[0] ?? ''), to: String(edge[1] ?? ''), label: String(edge[2] ?? ''), stereotype: 'navigationLink', evidenceRefs: [] }
    : {
      id: edge.id ? String(edge.id) : '',
      from: String(edge.from ?? edge.source ?? ''),
      to: String(edge.to ?? edge.target ?? ''),
      label: String(edge.label ?? edge.action ?? ''),
      stereotype: String(edge.stereotype ?? edge.type ?? 'navigationLink'),
      guard: edge.guard ? String(edge.guard) : '',
      effect: edge.effect ? String(edge.effect) : '',
      evidenceRefs: stringList(edge.evidenceRefs ?? edge.evidenceRef ?? edge.evidence)
    }) : [];
  const nodeIds = new Set();
  for (const edge of edges) {
    if (edge.from) nodeIds.add(edge.from);
    if (edge.to) nodeIds.add(edge.to);
  }
  const nodes = Array.isArray(spec.nodes) && spec.nodes.length > 0
    ? spec.nodes.map((node) => {
      if (typeof node === 'string') return { id: node, label: node, evidence: [], annotations: [], crop: null, focus: '' };
      const id = String(node.id ?? node.name ?? node.label);
      const evidenceRefs = stringList(node.evidenceRefs ?? node.evidenceRef ?? node.evidence);
      const nodeEvidence = resolvedEvidence.filter((item) => evidenceRefs.includes(item.id) || item.primaryFor.includes(id));
      const primaryEvidence = nodeEvidence.find((item) => item.path) ?? null;
      const annotations = [
        ...(Array.isArray(node.annotations) ? node.annotations.map(normalizeAnnotation).filter(Boolean) : []),
        ...nodeEvidence.flatMap((item) => item.annotations)
      ].filter((annotation, index, all) => all.findIndex((candidate) => candidate.id === annotation.id) === index);
      const crop = normalizeBounds(node.crop) ?? nodeEvidence.find((item) => item.crop)?.crop ?? annotations[0]?.bounds ?? null;
      return {
        id,
        label: String(node.label ?? node.name ?? node.id),
        route: node.route ? String(node.route) : '',
        facet: Array.isArray(node.facets) ? node.facets.join(', ') : (node.facet ? String(node.facet) : ''),
        facets: Array.isArray(node.facets) ? node.facets.map(String) : (node.facet ? String(node.facet).split(/\s*\+\s*|\s*,\s*/).filter(Boolean) : []),
        role: Array.isArray(node.role) ? node.role.join(', ') : (node.role ? String(node.role) : ''),
        packageName: node.package || node.packageName || node.lane || node.class || node.group ? String(node.package ?? node.packageName ?? node.lane ?? node.class ?? node.group) : '',
        navigationClass: node.navigationClass ? String(node.navigationClass) : '',
        effect: node.effect || node.sideEffect ? String(node.effect ?? node.sideEffect) : '',
        screenshot: node.screenshot || node.image || node.visualEvidence || primaryEvidence?.path || '',
        actions: Array.isArray(node.actions) ? node.actions.join(', ') : (node.actions || node.primaryActions || node.action ? String(node.actions ?? node.primaryActions ?? node.action) : ''),
        stereotype: node.stereotype || node.stereo ? String(node.stereotype ?? node.stereo) : '',
        type: node.type || node.facet ? String(node.type ?? node.facet) : '',
        evidenceRefs,
        evidence: nodeEvidence,
        annotations,
        crop,
        focus: String(node.focus ?? primaryEvidence?.caption ?? annotations[0]?.label ?? '')
      };
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
  if (kind === 'uwe-navigation' || kind === 'uwe') return await renderUweNavigationGraphvizSvg(spec, graph);
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
  const classNames = Array.isArray(spec.packages ?? spec.navigationPackages ?? spec.lanes ?? spec.classes)
    ? (spec.packages ?? spec.navigationPackages ?? spec.lanes ?? spec.classes).map((item) => typeof item === 'string' ? item : (item.label ?? item.id ?? item.name)).filter(Boolean).map(String)
    : [];
  for (const node of graph.nodes) {
    const name = node.packageName || 'Navigation';
    if (!classNames.includes(name)) classNames.push(name);
  }
  const lines = [
    'digraph UweNavigation {',
    '  graph [rankdir=LR, bgcolor="transparent", pad="0.22", nodesep="0.42", ranksep="0.78", compound=true, fontname="Segoe UI", label=' + dotQuote(spec.title ?? 'UWE navigation model') + ', labelloc=t, fontcolor="#111827"];',
    '  node [shape=plain, margin=0, fontname="Segoe UI", fontsize=12];',
    '  edge [fontname="Segoe UI", fontsize=10, color="#111827", fontcolor="#111827", arrowsize=0.72, arrowhead=open];'
  ];
  for (const className of classNames) {
    const nodes = graph.nodes.filter((node) => (node.packageName || 'Navigation') === className);
    if (nodes.length === 0) continue;
    lines.push('  subgraph cluster_' + safeName(className).replaceAll('-', '_') + ' {');
    lines.push('    label=' + dotQuote('package ' + className) + ';');
    lines.push('    color="#111827";');
    lines.push('    fillcolor="#ffffff";');
    lines.push('    style="filled";');
    for (const node of nodes) lines.push('    ' + dotQuote(node.id) + ' [label=<' + uweNodeHtmlLabel(node) + '>];');
    lines.push('  }');
  }
  for (const edge of graph.edges) {
    const label = [edge.stereotype ? '«' + edge.stereotype + '»' : '', edge.label, edge.guard ? '[' + edge.guard + ']' : ''].filter(Boolean).join(' ');
    lines.push('  ' + dotQuote(edge.from) + ' -> ' + dotQuote(edge.to) + (label ? ' [label=' + dotQuote(label) + ']' : '') + ';');
  }
  lines.push('}');
  return lines.join('\n');
}

function uweNodeHtmlLabel(node) {
  const label = dotHtmlText(String(node.label || node.id).slice(0, 36));
  const route = dotHtmlText(String(node.route || node.facet || node.role || '').slice(0, 42));
  const role = dotHtmlText(String(node.role || 'all roles').slice(0, 42));
  const action = dotHtmlText(String(node.actions || 'action not recorded').slice(0, 52));
  const effect = dotHtmlText(String(node.effect || 'effect not recorded').slice(0, 52));
  const screenshotToken = 'screenshot:' + dotHtmlText(node.id);
  return '<TABLE BORDER="1" CELLBORDER="0" CELLSPACING="0" CELLPADDING="6" COLOR="#111827">' +
    '<TR><TD BGCOLOR="#f3f4f6"><FONT FACE="Segoe UI" POINT-SIZE="11" COLOR="#111827">' + dotHtmlText(uweStereotype(node)) + '</FONT></TD></TR>' +
    '<TR><TD FIXEDSIZE="TRUE" WIDTH="320" HEIGHT="184" BGCOLOR="#ffffff"><FONT FACE="Segoe UI" POINT-SIZE="9" COLOR="#6b7280">' + screenshotToken + '</FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="12"><B>' + label + '</B></FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="10" COLOR="#334155">' + route + '</FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="9" COLOR="#475569">role: ' + role + '</FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="9" COLOR="#475569">action: ' + action + '</FONT></TD></TR>' +
    '<TR><TD ALIGN="LEFT"><FONT FACE="Segoe UI" POINT-SIZE="9" COLOR="#475569">effect: ' + effect + '</FONT></TD></TR>' +
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
  if (node.stereotype) return String(node.stereotype).startsWith('«') ? node.stereotype : '«' + node.stereotype + '»';
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
  const packages = [...new Set(graph.nodes.map((node) => node.packageName || 'Navigation'))];
  const first = graph.nodes[0] ?? {};
  const attrJson = (value) => escapeAttribute(JSON.stringify(value ?? null));
  const packageButtons = packages.map((name) => '<button type="button" title="' + escapeAttribute(name) + '" data-uwe-action="package:' + escapeAttribute(name) + '">' + escapeHtml(name) + '</button>').join('');
  const nodeButtons = graph.nodes.map((node) => '<button type="button" title="' + escapeAttribute((node.label || node.id) + ' - ' + uweStereotype(node)) + '" data-uwe-focus-node="' + escapeAttribute(node.id) + '">' + escapeHtml(node.label || node.id) + '</button>').join('');
  const guidedButtons = graph.nodes.filter((node) => String(node.type || node.facet || '').match(/navigation|access|adaptation|process/i)).slice(0, 4)
    .map((node) => '<button type="button" title="Inspect ' + escapeAttribute(node.label || node.id) + '" data-uwe-focus-node="' + escapeAttribute(node.id) + '">' + escapeHtml('Inspect ' + (node.label || node.id)) + '</button>').join('');
  const nodeData = graph.nodes.map((node) => {
    const dataUrl = imageDataUrl(node.screenshot) || '';
    return '<span data-uwe-node data-uwe-id="' + escapeAttribute(node.id) + '" data-uwe-label="' + escapeAttribute(node.label || node.id) + '" data-uwe-stereo="' + escapeAttribute(uweStereotype(node)) + '" data-uwe-type="' + escapeAttribute(node.type || node.facet || 'navigation') + '" data-uwe-package="' + escapeAttribute(node.packageName || 'Navigation') + '" data-uwe-route="' + escapeAttribute(node.route || '') + '" data-uwe-role="' + escapeAttribute(node.role || '') + '" data-uwe-actions="' + escapeAttribute(node.actions || 'Inspect this node and outgoing UWE links.') + '" data-uwe-effect="' + escapeAttribute(node.effect || '') + '" data-uwe-focus="' + escapeAttribute(node.focus || '') + '" data-uwe-crop="' + attrJson(node.crop) + '" data-uwe-annotations="' + attrJson(node.annotations || []) + '" data-uwe-screenshot="' + escapeAttribute(dataUrl) + '"></span>';
  }).join('');
  const edgeData = graph.edges.map((edge) => '<span data-uwe-edge data-uwe-from="' + escapeAttribute(edge.from) + '" data-uwe-to="' + escapeAttribute(edge.to) + '" data-uwe-label="' + escapeAttribute([edge.stereotype ? '«' + edge.stereotype + '»' : '', edge.label, edge.guard ? '[' + edge.guard + ']' : ''].filter(Boolean).join(' ') || '«navigationLink»') + '"></span>').join('');
  const firstImage = imageDataUrl(first.screenshot) || 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==';
  const profileRows = graph.nodes.map((node) => '<span class="uwe-profile-chip"><strong>' + escapeHtml(uweStereotype(node)) + '</strong>' + escapeHtml(node.label || node.id) + '</span>').join('');
  const edgeRows = graph.edges.map((edge) => '<tr><td>' + escapeHtml(edge.from) + '</td><td>' + escapeHtml(edge.stereotype ? '«' + edge.stereotype + '»' : '«navigationLink»') + '</td><td>' + escapeHtml(edge.label || '') + (edge.guard ? '<br><span class="muted">guard: ' + escapeHtml(edge.guard) + '</span>' : '') + '</td><td>' + escapeHtml(edge.to) + '</td></tr>').join('');
  const workspaceCss = 'html.uwe-focus-active body{overflow:hidden}.uwe-engine-workspace{border:1px solid #223049;background:#fff;margin:12px 0 18px;box-shadow:0 1px 0 rgba(15,23,42,.04)}.uwe-engine-workspace.uwe-focus-mode{position:fixed;inset:14px;z-index:40;margin:0;display:flex;flex-direction:column;box-shadow:0 18px 60px rgba(15,23,42,.28)}.uwe-engine-head{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:12px;border-bottom:1px solid #223049;background:#f7f9fc;padding:14px 16px}.uwe-engine-kicker{display:block;margin-bottom:4px;color:#0f766e;font-size:11px;font-weight:900;text-transform:uppercase;letter-spacing:.08em}.uwe-engine-head h4{margin:0;font-size:19px;line-height:1.18}.uwe-engine-head p{max-width:920px}.uwe-conformance{border-bottom:1px solid #d8dee8;background:#fff;padding:10px 16px;color:#334155}.uwe-profile-strip{display:flex;flex-wrap:wrap;gap:7px;padding:10px 16px;border-bottom:1px solid #d8dee8;background:#fbfdff}.uwe-profile-chip{display:inline-flex;gap:6px;align-items:center;border:1px solid #cbd5e1;background:#fff;padding:4px 8px;font-size:12px}.uwe-profile-chip strong{color:#0f766e}.uwe-engine-toolbar,.uwe-node-map,.uwe-guided-map{display:flex;flex-wrap:wrap;gap:7px;align-items:center;padding:9px 16px;border-bottom:1px solid #d8dee8}.uwe-node-map{background:#fbfdff}.uwe-guided-map{background:#f8fafc}.uwe-map-label{font-size:11px;font-weight:900;text-transform:uppercase;letter-spacing:.08em;color:#5b6472}.uwe-engine-toolbar button,.uwe-node-map button,.uwe-guided-map button,.uwe-inspector-button{cursor:pointer;border:1px solid #223049;background:#fff;color:#111827;padding:6px 10px;font-size:12px;font-weight:800;max-width:260px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.uwe-engine-toolbar button:hover,.uwe-node-map button:hover,.uwe-guided-map button:hover,.uwe-inspector-button:hover,.uwe-engine-toolbar button.active,.uwe-node-map button.active,.uwe-guided-map button.active{background:#223049;color:#fff}.uwe-runtime-badge{margin-left:auto;color:#374151;font-size:12px;font-weight:800}.uwe-engine-grid{display:grid;grid-template-columns:minmax(0,1fr) 380px;min-height:700px}.uwe-focus-mode .uwe-engine-grid{min-height:0;height:calc(100vh - 372px);flex:1}.uwe-cy-graph{min-height:700px;background:linear-gradient(180deg,#fbfdff,#f8fafc);position:relative}.uwe-focus-mode .uwe-cy-graph{min-height:0}.uwe-cy-placeholder{position:absolute;inset:0;display:grid;place-items:center;color:#64748b;font-weight:800}.uwe-engine-inspector{border-left:1px solid #223049;background:#fbfdff;padding:14px;overflow:auto}.uwe-evidence-frame{position:relative;width:100%;aspect-ratio:16/9;border:1px solid #223049;background:#eef2f6;margin:10px 0;overflow:hidden;cursor:zoom-in}.uwe-evidence-frame img{display:block;width:100%;height:100%;object-fit:contain;background:#fff}.uwe-annotation-layer{position:absolute;inset:0;pointer-events:none}.uwe-screenshot-focus-box{position:absolute;border:3px solid #e11d48;border-radius:6px;box-shadow:0 0 0 2px rgba(255,255,255,.95),0 0 0 999px rgba(15,23,42,.18),0 8px 28px rgba(225,29,72,.32)}.uwe-screenshot-focus-box span{position:absolute;left:-3px;top:-25px;background:#e11d48;color:#fff;border-radius:4px 4px 0 0;padding:3px 7px;font-size:10px;font-weight:900;letter-spacing:.04em;text-transform:uppercase;white-space:nowrap}.uwe-focus-crop{display:none;position:relative;height:150px;border:1px solid #d1d9e3;background-color:#f8fafc;background-repeat:no-repeat;overflow:hidden;margin:10px 0}.uwe-focus-crop.active{display:block}.uwe-focus-crop:after{content:"";position:absolute;inset:8px;border:3px solid #e11d48;border-radius:6px;box-shadow:0 0 0 2px rgba(255,255,255,.95);pointer-events:none}.uwe-focus-crop-label{position:absolute;left:8px;top:8px;z-index:1;background:#e11d48;color:#fff;border-radius:4px;padding:4px 7px;font-size:10px;font-weight:900;letter-spacing:.04em;text-transform:uppercase}.uwe-inspector-kicker{display:block;color:#0f766e;font-size:12px;font-weight:900;text-transform:uppercase}.uwe-inspector-title{margin:2px 0 0;font-size:20px}.uwe-inspector-block{border-top:1px solid #d8dee8;padding:9px 0}.uwe-inspector-block strong{display:block;font-size:11px;text-transform:uppercase;color:#5b6472}.uwe-edge-inventory{border-top:1px solid #d8dee8;background:#fff;padding:10px 16px}.uwe-edge-inventory summary{cursor:pointer;font-weight:800}.uwe-edge-inventory table{font-size:12px}.uwe-engine-stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px;padding:10px 16px;border-top:1px solid #d8dee8}.uwe-engine-stats div{border:1px solid #d8dee8;background:#f8fafc;padding:9px}.uwe-engine-stats strong{display:block;font-size:22px;line-height:1}.uwe-engine-stats span{display:block;margin-top:4px;color:#5b6472}.uwe-lightbox{position:fixed;inset:0;z-index:80;display:none;align-items:center;justify-content:center;background:rgba(15,23,42,.78);padding:22px}.uwe-lightbox.active{display:flex}.uwe-lightbox-panel{max-width:min(1180px,96vw);max-height:94vh}.uwe-lightbox-frame{position:relative;max-width:min(1180px,96vw);max-height:82vh;background:#fff;border:1px solid #fff;overflow:hidden}.uwe-lightbox-frame img{display:block;max-width:100%;max-height:82vh}.uwe-lightbox-caption{color:#fff;margin-top:8px;font-size:14px}.uwe-lightbox-close{cursor:pointer;margin-bottom:8px;border:1px solid #fff;background:#fff;color:#111827;padding:6px 10px;font-weight:800}@media(max-width:920px){.uwe-engine-head{grid-template-columns:1fr}.uwe-engine-toolbar,.uwe-node-map,.uwe-guided-map{flex-wrap:nowrap;overflow-x:auto}.uwe-engine-grid{grid-template-columns:1fr}.uwe-focus-mode .uwe-engine-grid{height:calc(100vh - 470px)}.uwe-cy-graph{min-height:420px}.uwe-engine-inspector{border-left:0;border-top:1px solid #223049}.uwe-runtime-badge{margin-left:0}.uwe-engine-stats{grid-template-columns:1fr}}';
  return '<style>' + workspaceCss + '</style>' +
    '<div id="' + workspaceId + '" class="uwe-engine-workspace" role="group" aria-label="Engine-backed UWE navigation workspace">' +
    '<div class="uwe-engine-head"><div><span class="uwe-engine-kicker">UWE navigation model with screenshot evidence</span><h4>Engine-Backed UWE Navigation Workspace</h4><p class="muted">Primary view rendered from structured source data with Cytoscape.js and dagre. Screenshots are embedded evidence; UWE stereotypes, access scope, actions, links, and effects remain explicit.</p></div><span class="tool-badge">Cytoscape + dagre</span></div>' +
    '<div class="uwe-conformance"><strong>UWE conformance:</strong> screen-level destinations are modeled as <code>«navigationClass»</code>, process/access states retain their own stereotypes, and screenshot thumbnails are review evidence attached to those nodes.</div>' +
    '<div class="uwe-profile-strip" aria-label="UWE profile summary">' + profileRows + '</div>' +
    '<div class="uwe-engine-toolbar"><button type="button" data-uwe-action="fit">Fit graph</button><button type="button" data-uwe-action="layout">Re-run layout</button><button type="button" data-uwe-action="workspace-focus">Focus workspace</button>' + packageButtons + '<span class="uwe-runtime-badge" data-uwe-runtime-badge>Workspace runtime pending</span></div><div class="uwe-node-map" aria-label="UWE node focus map">' + nodeButtons + '</div>' +
    (guidedButtons ? '<div class="uwe-guided-map" aria-label="Guided UWE inspection paths"><span class="uwe-map-label">Review path</span>' + guidedButtons + '</div>' : '') +
    '<div class="uwe-engine-grid"><div class="uwe-cy-graph" data-uwe-cy><div class="uwe-cy-placeholder">Rendering UWE graph workspace...</div></div>' +
    '<aside class="uwe-engine-inspector" aria-label="Selected UWE node inspector"><span class="uwe-inspector-kicker" data-uwe-inspector-stereo>' + escapeHtml(uweStereotype(first)) + '</span><h4 class="uwe-inspector-title" data-uwe-inspector-title>' + escapeHtml(first.label || first.id || 'UWE node') + '</h4><div class="uwe-evidence-frame" data-uwe-evidence-frame><img data-uwe-inspector-image src="' + escapeAttribute(firstImage) + '" alt="Selected UWE node screenshot"><div class="uwe-annotation-layer" data-uwe-annotation-layer></div></div><div class="uwe-focus-crop" data-uwe-focus-crop><span class="uwe-focus-crop-label">zoomed focus</span></div><button type="button" class="uwe-inspector-button" data-uwe-open-screenshot>Open annotated screenshot</button><div class="uwe-inspector-block"><strong>Screenshot focus</strong><span data-uwe-inspector-focus>' + escapeHtml(first.focus || 'Red highlights are evidence-only annotations for the selected UWE node.') + '</span></div><div class="uwe-inspector-block"><strong>Package</strong><span data-uwe-inspector-package>' + escapeHtml(first.packageName || 'Navigation') + '</span></div><div class="uwe-inspector-block"><strong>Route or state</strong><span data-uwe-inspector-route>' + escapeHtml(first.route || 'state') + '</span></div><div class="uwe-inspector-block"><strong>Role</strong><span data-uwe-inspector-role>' + escapeHtml(first.role || 'all roles') + '</span></div><div class="uwe-inspector-block"><strong>Available user action</strong><span data-uwe-inspector-actions>' + escapeHtml(first.actions || 'Inspect this node and outgoing UWE links.') + '</span></div><div class="uwe-inspector-block"><strong>System effect</strong><span data-uwe-inspector-effect>' + escapeHtml(first.effect || 'Effect not recorded.') + '</span></div></aside></div>' +
    '<details class="uwe-edge-inventory"><summary>Typed UWE link inventory</summary><table><thead><tr><th>From</th><th>Type</th><th>Action / guard</th><th>To</th></tr></thead><tbody>' + edgeRows + '</tbody></table></details>' +
    '<div class="uwe-engine-stats"><div><strong data-uwe-stat="nodes">' + graph.nodes.length + '</strong><span>UWE nodes</span></div><div><strong data-uwe-stat="edges">' + graph.edges.length + '</strong><span>typed links</span></div><div><strong data-uwe-stat="packages">' + packages.length + '</strong><span>packages</span></div></div>' +
    '<div hidden>' + nodeData + edgeData + '</div><div class="uwe-lightbox" data-uwe-lightbox role="dialog" aria-modal="true" aria-label="UWE screenshot preview"><div class="uwe-lightbox-panel"><button type="button" class="uwe-lightbox-close" data-uwe-lightbox-close>Close</button><div class="uwe-lightbox-frame" data-uwe-lightbox-frame><img data-uwe-lightbox-image alt="Selected UWE screenshot"><div class="uwe-annotation-layer" data-uwe-lightbox-layer></div></div><div class="uwe-lightbox-caption" data-uwe-lightbox-caption></div></div></div></div>';
}

function renderUweNavigationSvg(spec, graph) {
  const cardWidth = 210;
  const cardHeight = 168;
  const gapX = 42;
  const gapY = 34;
  const laneGap = 32;
  const lanePadding = 18;
  const explicitClasses = Array.isArray(spec.packages ?? spec.navigationPackages ?? spec.lanes ?? spec.classes)
    ? (spec.packages ?? spec.navigationPackages ?? spec.lanes ?? spec.classes).map((item) => typeof item === 'string' ? item : (item.label ?? item.id ?? item.name)).filter(Boolean).map(String)
    : [];
  const discoveredClasses = graph.nodes.map((node) => node.packageName || 'Navigation').filter((value, index, all) => all.indexOf(value) === index);
  const classNames = [...explicitClasses, ...discoveredClasses.filter((name) => !explicitClasses.includes(name))];
  const lanes = classNames.map((name) => {
    const nodes = graph.nodes.filter((node) => (node.packageName || 'Navigation') === name);
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
    '<text x="20" y="' + (lane.y + 24) + '" font-size="13" font-weight="800" fill="#334155">package ' + escapeHtml(lane.name.slice(0, 80)) + '</text></g>'
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
      '<text x="' + (pos.x + 10) + '" y="' + (pos.y + 17) + '" font-size="11" font-weight="800" fill="#0f766e">' + escapeHtml(uweStereotype(node)) + '</text>' +
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
  if (options.enableUweWorkspace && !(svgPanZoomRuntime && cytoscapeRuntime && dagreRuntime && cytoscapeDagreRuntime && uweWorkspaceRuntime)) {
    throw new Error('reviewed-uwe-workspace requires bundled svg-pan-zoom, cytoscape, dagre, cytoscape-dagre, and scripts/uwe-workspace-runtime.js');
  }
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
  const infographicSpecs = parseInfographicSpecs(source, artifact);
  const hasUweInfographic = infographicSpecs.some((spec) => ['uwe-navigation', 'uwe'].includes(String(spec.kind ?? spec.mark ?? spec.type ?? '').toLowerCase()));
  const infographicSection = await renderInfographicSpecs(source, artifact);
  const toolkitSection = hasUweInfographic
    ? '<details class="panel"><summary><strong>Open-Source Infographic Toolkit</strong></summary>' + renderInfographicToolkit().replace(/^<section class="panel">|<\/section>$/g, '') + '</details>'
    : renderInfographicToolkit();
  const evidenceCount = Array.isArray(artifact.evidenceLinks) ? artifact.evidenceLinks.length : 0;
  const updateCount = Array.isArray(artifact.updateTriggers) ? artifact.updateTriggers.length : 0;
  const body = '<header><h1>' + escapeHtml(title) + '</h1><p>' + escapeHtml(summary) + '</p></header><main>' +
    (hasUweInfographic ? infographicSection : '') +
    '<section class="grid"><div class="panel"><h2>Review Verdict</h2><p>' + escapeHtml(summary) + '</p><div class="callout"><strong>Source first:</strong> edit ' + linkFor(outPath, artifact.source, artifact.source) + ' before regenerating this review surface.</div></div>' +
    '<div class="metrics"><div class="metric green"><strong>' + escapeHtml(artifact.status || 'draft') + '</strong><span>Status</span></div><div class="metric blue"><strong>' + evidenceCount + '</strong><span>Evidence links</span></div><div class="metric amber"><strong>' + stats.sectionCount + '</strong><span>Major sections</span></div><div class="metric violet"><strong>' + escapeHtml(family) + '</strong><span>Artifact family</span></div></div></section>' +
    '<section class="panel"><h2>Infographic Snapshot</h2><div class="bar-row"><span>Evidence coverage</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, evidenceCount * 20)) + '"></span></span><strong>' + evidenceCount + '</strong></div><div class="bar-row"><span>Source depth</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, stats.sectionCount * 20)) + '"></span></span><strong>' + stats.sectionCount + '</strong></div><div class="bar-row"><span>Update triggers</span><span class="bar-track"><span class="bar w' + Math.min(100, Math.max(20, updateCount * 20)) + '"></span></span><strong>' + updateCount + '</strong></div></section>' +
    toolkitSection +
    (hasUweInfographic ? '' : infographicSection) +
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
`
}

func developerModelReviewGeneratorScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const args = process.argv.slice(2);
const checkOnly = args.includes('--check');
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const modelPolicy = developerArtifacts.modelPolicy ?? {};
const modeling = developerArtifacts.modeling ?? modelPolicy.uml ?? {};
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const modelReviewDir = path.join(root, modeling.reviewDir ?? modelPolicy.uml?.reviewDir ?? 'generated/review/models');
const requiredCsp = developerArtifacts.htmlPolicy?.requiredCSP ?? "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'";

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function safeFileName(value) {
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

function htmlPage(title, body) {
  return '<!doctype html>\n<html lang="en">\n<head>\n<meta charset="utf-8">\n<meta name="viewport" content="width=device-width, initial-scale=1">\n<meta http-equiv="Content-Security-Policy" content="' + escapeAttribute(requiredCsp) + '">\n<title>' + escapeHtml(title) + '</title>\n<style>\n:root{color-scheme:light dark;font-family:Inter,Segoe UI,Arial,sans-serif;line-height:1.5;--bg:#f6f8fb;--panel:#fff;--text:#1f2933;--muted:#52616b;--line:#d9e2ec;--accent:#0f766e;--code:#102a43;--codeText:#f0f4f8}body{margin:0;color:var(--text);background:var(--bg)}main{max-width:1180px;margin:0 auto;padding:28px 18px 44px}.hero{display:grid;grid-template-columns:minmax(0,1.2fr) minmax(240px,.8fr);gap:18px;align-items:start}.panel,section{margin:18px 0;padding:18px;background:var(--panel);border:1px solid var(--line);border-radius:8px}h1,h2,h3{line-height:1.2;margin:0 0 10px}p{margin:0 0 12px}.muted{color:var(--muted)}table{width:100%;border-collapse:collapse;background:var(--panel)}th,td{text-align:left;vertical-align:top;border-bottom:1px solid var(--line);padding:10px}code,pre{font-family:ui-monospace,SFMono-Regular,Consolas,monospace}pre{white-space:pre-wrap;overflow:auto;background:var(--code);color:var(--codeText);padding:16px;border-radius:6px}.meta{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:10px}.meta div,.pill{padding:9px 10px;background:#eef6f6;border:1px solid #c8e7e3;border-radius:6px}.tabs{margin-top:18px}.tabs>input{position:absolute;inline-size:1px;block-size:1px;overflow:hidden;clip:rect(0 0 0 0)}.tab-labels{display:flex;flex-wrap:wrap;gap:8px;border-bottom:1px solid var(--line);padding-bottom:10px}.tab-labels label{cursor:pointer;padding:8px 11px;border:1px solid var(--line);border-radius:6px;background:var(--panel);font-weight:600}.tab-panel{display:none}.tabs input:nth-of-type(1):checked~.tab-panels .tab-panel:nth-of-type(1),.tabs input:nth-of-type(2):checked~.tab-panels .tab-panel:nth-of-type(2),.tabs input:nth-of-type(3):checked~.tab-panels .tab-panel:nth-of-type(3),.tabs input:nth-of-type(4):checked~.tab-panels .tab-panel:nth-of-type(4),.tabs input:nth-of-type(5):checked~.tab-panels .tab-panel:nth-of-type(5){display:block}.diagram-card{border:1px solid var(--line);border-radius:8px;overflow:hidden;background:#fbfdff}.diagram-header{display:flex;justify-content:space-between;gap:10px;padding:10px 12px;background:#eef2f7;border-bottom:1px solid var(--line)}.diagram-body{padding:14px}.flow{display:flex;flex-wrap:wrap;gap:10px;align-items:center}.node{padding:10px 12px;border:1px solid #9fb3c8;border-radius:6px;background:#fff;min-width:88px;text-align:center}.arrow{color:var(--accent);font-weight:700}.gallery{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px}.gallery figure{margin:0;border:1px solid var(--line);border-radius:8px;overflow:hidden;background:#fff}.gallery img{display:block;width:100%;height:auto}.gallery figcaption{padding:9px 10px;color:var(--muted)}.compare{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:14px}@media (max-width:760px){.hero{grid-template-columns:1fr}.tab-labels label{flex:1 1 auto;text-align:center}}@media (prefers-color-scheme:dark){:root{--bg:#102a43;--panel:#1f2933;--text:#d9e2ec;--muted:#bcccdc;--line:#334e68;--accent:#5eead4}.meta div,.pill{background:#243b53;border-color:#486581}.diagram-header{background:#243b53}.diagram-card,.gallery figure,.node{background:#1f2933}}\n</style>\n</head>\n<body>\n<main>\n' + body + '\n</main>\n</body>\n</html>\n';
}

function readSource(artifact) {
  if (typeof artifact.source !== 'string') return '';
  const sourcePath = path.resolve(root, artifact.source);
  if (!sourcePath.startsWith(root + path.sep) && sourcePath !== root) return '';
  if (!fs.existsSync(sourcePath) || !fs.statSync(sourcePath).isFile()) return '';
  return fs.readFileSync(sourcePath, 'utf8');
}

function artifactPath(artifact) {
  const fileName = safeFileName(artifact.reviewSurface ? path.basename(artifact.reviewSurface, '.html') : artifact.id || artifact.modelId) + '.html';
  return path.join(modelReviewDir, fileName);
}

function firstParagraph(markdown) {
  const fence = String.fromCharCode(96).repeat(3);
  const withoutFences = String(markdown || '').replace(new RegExp(fence + '[\\s\\S]*?' + fence, 'g'), '');
  const lines = withoutFences.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  for (const line of lines) {
    if (!line.startsWith('#') && !line.startsWith(fence)) return line;
  }
  return '';
}

function fencedBlocks(markdown) {
  const blocks = [];
  const fence = String.fromCharCode(96).repeat(3);
  const pattern = new RegExp(fence + '([a-zA-Z0-9_-]*)\\r?\\n([\\s\\S]*?)' + fence, 'g');
  let match;
  while ((match = pattern.exec(markdown)) !== null) {
    blocks.push({ language: match[1] || 'text', body: match[2].trim() });
  }
  return blocks;
}

function compactDiagramMarkup(source) {
  const lines = String(source || '').split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  const edges = [];
  for (const line of lines) {
    const arrow = line.match(/^"?([^"\[\](){}:;-][^":;-]*?)"?\s*(?:-->|->>|--|-\)|-\])\s*"?([^":;]+?)"?(?::.*)?$/);
    if (arrow) edges.push([arrow[1].replace(/\[.*$/, '').trim(), arrow[2].replace(/\[.*$/, '').trim()]);
  }
  if (edges.length === 0) {
    return '<pre>' + escapeHtml(source || 'No diagram source found.') + '</pre>';
  }
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
  const diagramSource = preferred?.body || source;
  return '<div class="diagram-card"><div class="diagram-header"><strong>' + escapeHtml(artifact.notation || 'model') + ' ' + escapeHtml(artifact.modelKind || 'view') + '</strong><span class="muted">static preview, source-backed</span></div><div class="diagram-body">' + compactDiagramMarkup(diagramSource) + '</div></div>';
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

function renderArtifact(artifact, byId, outPath) {
  const source = readSource(artifact);
  const summary = artifact.summary || artifact.purpose || firstParagraph(source) || 'Source-backed model review artifact.';
  const rows = [
    ['ID', artifact.id],
    ['Type', artifact.type],
    ['Status', artifact.status],
    ['Model ID', artifact.modelId],
    ['Model Kind', artifact.modelKind],
    ['Method', artifact.method],
    ['Notation', artifact.notation],
    ['Source', artifact.source],
    ['Owner', artifact.owner],
  ];
  let body = '<div class="hero"><section><h1>' + escapeHtml(artifact.title || artifact.modelId || artifact.id || 'Model Review') + '</h1><p>' + escapeHtml(summary) + '</p></section><section><h2>Review Focus</h2><div class="meta">';
  body += '<div><strong>Status</strong><br>' + escapeHtml(artifact.status) + '</div><div><strong>Kind</strong><br>' + escapeHtml(artifact.modelKind) + '</div><div><strong>Method</strong><br>' + escapeHtml(artifact.method) + '</div><div><strong>Owner</strong><br>' + escapeHtml(artifact.owner) + '</div>';
  body += '</div></section></div>\n<section class="meta">';
  for (const [label, value] of rows) body += '<div><strong>' + escapeHtml(label) + '</strong><br>' + escapeHtml(value) + '</div>';
  body += '</section>\n<div class="tabs"><input id="tab-overview" name="tabs" type="radio" checked><input id="tab-visual" name="tabs" type="radio"><input id="tab-source" name="tabs" type="radio"><input id="tab-evidence" name="tabs" type="radio"><input id="tab-diff" name="tabs" type="radio"><div class="tab-labels"><label for="tab-overview">Overview</label><label for="tab-visual">Visuals</label><label for="tab-source">Source</label><label for="tab-evidence">Evidence</label><label for="tab-diff">Diff</label></div><div class="tab-panels">';
  body += '<section class="tab-panel"><h2>Overview</h2><p>' + escapeHtml(summary) + '</p><div class="meta"><div><strong>Abstraction</strong><br>' + escapeHtml(artifact.abstractionLevel) + '</div><div><strong>Notation</strong><br>' + escapeHtml(artifact.notation) + '</div><div><strong>Canonical Source</strong><br>' + linkFor(outPath, artifact.source, artifact.source || 'source') + '</div><div><strong>Review Surface</strong><br>' + linkFor(outPath, artifact.reviewSurface, artifact.reviewSurface || '') + '</div><div><strong>Implementation Touchpoints</strong><br>' + (artifact.implementationTouchpoints || []).map((item) => linkFor(outPath, item, item)).join(', ') + '</div><div><strong>Doc Touchpoints</strong><br>' + (artifact.docTouchpoints || []).map((item) => linkFor(outPath, item, item)).join(', ') + '</div></div></section>';
  body += '<section class="tab-panel"><h2>Visuals</h2>' + diagramSection(source, artifact) + '<h3>Screenshots And Evidence Images</h3>' + gallerySection(artifact) + '</section>';
  body += '<section class="tab-panel"><h2>Canonical Source</h2><p>' + linkFor(outPath, artifact.source, artifact.source || 'source') + '</p><pre>' + escapeHtml(source || 'Source not found or not readable.') + '</pre></section>';
  body += '<section class="tab-panel"><h2>Evidence</h2>' + listItems(artifact.evidenceLinks, 'No evidence links are listed yet.', outPath) + '<h3>Update Triggers</h3>' + listItems(artifact.updateTriggers, 'No update triggers are listed yet.') + '<h3>Freshness</h3><div class="meta"><div><strong>Source Hash</strong><br>' + escapeHtml(artifact.sourceHash || '') + '</div><div><strong>Renderer</strong><br>' + escapeHtml(artifact.renderer || 'skill-harness model review generator') + '</div><div><strong>Generated</strong><br>' + escapeHtml(artifact.generatedAt || artifact.freshness?.generatedAt || 'not-recorded') + '</div></div></section>';
  if (artifact.type === 'model-diff') {
    const diff = artifact.diff ?? {};
    const before = byId.get(diff.beforeArtifactId);
    const after = byId.get(diff.afterArtifactId);
    body += '<section class="tab-panel"><h2>Before And After</h2><div class="compare"><div class="panel"><h3>Before</h3><p><strong>' + escapeHtml(diff.beforeArtifactId) + '</strong></p><p class="muted">' + escapeHtml(before?.source ?? '') + '</p><pre>' + escapeHtml(before ? readSource(before) : 'Missing before artifact.') + '</pre></div><div class="panel"><h3>After</h3><p><strong>' + escapeHtml(diff.afterArtifactId) + '</strong></p><p class="muted">' + escapeHtml(after?.source ?? '') + '</p><pre>' + escapeHtml(after ? readSource(after) : 'Missing after artifact.') + '</pre></div></div></section>';
  } else {
    body += '<section class="tab-panel"><h2>Diff</h2><p class="muted">This is a model-view artifact. Create a model-diff artifact to show before/after model changes.</p></section>';
  }
  body += '</div></div>';
  return htmlPage('Model Review: ' + (artifact.modelId || artifact.id || 'model'), body);
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

const indexRows = [];
const expectedFiles = new Map();
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
    const actual = fs.readFileSync(filePath, 'utf8');
    if (actual !== expected) failures.push('stale generated model review: ' + repoPath(filePath));
  }
  const currentManifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  if (JSON.stringify(currentManifest, null, 2) + '\n' !== JSON.stringify(manifest, null, 2) + '\n') {
    failures.push('manifest reviewSurface entries are stale; run node scripts/generate-model-review.mjs');
  }
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
`
}

func developerArtifactOpenScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { pathToFileURL } from 'node:url';

const root = process.cwd();
const args = process.argv.slice(2);
const jsonMode = args.includes('--json');
const printOnly = args.includes('--print') || args.includes('--dry-run') || process.env.CI === 'true';
const explicitTarget = args.find((arg) => !arg.startsWith('--'));

function repoPath(filePath) {
  return path.relative(root, filePath).replaceAll(path.sep, '/');
}

function isInsideRoot(filePath) {
  return filePath === root || filePath.startsWith(root + path.sep);
}

function resolveReviewPath(value) {
  if (typeof value !== 'string' || value.trim() === '') return null;
  const resolved = path.resolve(root, value);
  if (!isInsideRoot(resolved)) return null;
  return resolved;
}

function readJSON(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch {
    return null;
  }
}

function firstExisting(paths) {
  for (const candidate of paths) {
    if (candidate && fs.existsSync(candidate) && fs.statSync(candidate).isFile()) return candidate;
  }
  return null;
}

function discoverTarget() {
  const config = readJSON(path.join(root, '.skill-harness', 'project.json')) ?? {};
  const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
  const reviewDir = developerArtifacts.reviewSurface?.outDir ?? 'generated/review';
  const modelReviewDir = developerArtifacts.modeling?.reviewDir ?? developerArtifacts.modelPolicy?.uml?.reviewDir ?? path.join(reviewDir, 'models');
  const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
  const manifest = readJSON(manifestPath);

  if (explicitTarget) {
    const resolved = resolveReviewPath(explicitTarget);
    if (resolved && fs.existsSync(resolved)) return resolved;
    const artifact = (manifest?.artifacts ?? []).find((item) => item?.id === explicitTarget || item?.modelId === explicitTarget);
    if (artifact?.reviewSurface) {
      const reviewPath = resolveReviewPath(artifact.reviewSurface);
      if (reviewPath && fs.existsSync(reviewPath)) return reviewPath;
    }
    throw new Error('review artifact not found by path or artifact id: ' + explicitTarget);
  }

  const manifestTargets = [];
  for (const artifact of Array.isArray(manifest?.artifacts) ? manifest.artifacts : []) {
    if (typeof artifact?.reviewSurface === 'string' && artifact.reviewSurface.endsWith('.html')) {
      const resolved = resolveReviewPath(artifact.reviewSurface);
      if (resolved) manifestTargets.push(resolved);
    }
  }

  const discovered = firstExisting([
    path.join(root, reviewDir, 'index.html'),
    path.join(root, modelReviewDir, 'index.html'),
    ...manifestTargets,
  ]);
  if (discovered) return discovered;
  throw new Error('no generated HTML review artifact found; generate one first');
}

function hostHint() {
  const originator = process.env.CODEX_INTERNAL_ORIGINATOR_OVERRIDE ?? '';
  if (process.env.CODEX_THREAD_ID || /codex/i.test(originator)) {
    return 'Codex app detected: prefer opening this file with the Browser plugin when the agent has it.';
  }
  if (process.env.CLAUDE_DESKTOP || /claude/i.test(originator)) {
    return 'Claude desktop context detected: prefer the built-in browser or preview tool when available.';
  }
  return '';
}

function hostAction() {
  const originator = process.env.CODEX_INTERNAL_ORIGINATOR_OVERRIDE ?? '';
  if (process.env.CODEX_THREAD_ID || /codex/i.test(originator)) return 'codex-browser-plugin';
  if (process.env.CLAUDE_DESKTOP || /claude/i.test(originator)) return 'claude-desktop-preview';
  if (printOnly) return 'print-file-url';
  return 'system-default-browser';
}

function openSystemDefault(filePath) {
  const url = pathToFileURL(filePath).href;
  let command;
  let commandArgs;
  if (process.platform === 'win32') {
    command = 'cmd';
    commandArgs = ['/c', 'start', '', url];
  } else if (process.platform === 'darwin') {
    command = 'open';
    commandArgs = [url];
  } else {
    command = 'xdg-open';
    commandArgs = [url];
  }
  const child = spawn(command, commandArgs, { detached: true, stdio: 'ignore' });
  child.unref();
  return url;
}

try {
  const target = discoverTarget();
  const url = pathToFileURL(target).href;
  const hint = hostHint();
  if (jsonMode) {
    console.log(JSON.stringify({
      path: target,
      repoPath: repoPath(target),
      url,
      hostAction: hostAction(),
      openMode: printOnly ? 'print' : 'open',
      hint
    }, null, 2));
    process.exit(0);
  }
  if (hint) console.log(hint);
  if (printOnly) {
    console.log(url);
  } else {
    console.log('Opening ' + repoPath(target));
    console.log(openSystemDefault(target));
  }
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
`
}

func developerModelArtifactPolicyScript() string {
	return `import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
const modelPolicy = developerArtifacts.modelPolicy ?? {};
const modeling = modelPolicy.uml ?? developerArtifacts.modeling ?? {};
const modelingMode = developerArtifacts.modeling?.mode ?? modeling.mode ?? 'baseline';
const generatedReviewRequired = modeling.generatedReviewRequired === true || modelingMode === 'uml-first';
const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const reviewRoot = path.resolve(root, developerArtifacts.reviewSurface?.outDir ?? 'generated/review');
const modelReviewRoot = path.resolve(root, modeling.reviewDir ?? path.join(developerArtifacts.reviewSurface?.outDir ?? 'generated/review', 'models'));
const allowedMethods = new Set(modeling.methods ?? []);
const allowedSourceExtensions = new Set(modeling.allowedSourceExtensions ?? ['.md', '.toon', '.mmd', '.puml', '.plantuml', '.dsl', '.json', '.yaml', '.yml']);
const allowedModelKinds = new Set(modelPolicy.allowedModelKinds ?? []);
const allowedNotations = new Set(modelPolicy.allowedNotations ?? []);
const evidenceDefaultKinds = new Set(modeling.evidenceDefaultKinds ?? []);
const authoredOrEvidenceKinds = new Set(modeling.authoredOrEvidenceKinds ?? []);
const methodModelKinds = modeling.methodModelKinds ?? {};
const allowedFacets = modeling.allowedFacets ?? {};

function relativeForMessage(filePath) {
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
if (!fs.existsSync(manifestPath)) failures.push('missing artifact manifest: ' + relativeForMessage(manifestPath));

if (fs.existsSync(manifestPath)) {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  const artifacts = Array.isArray(manifest.artifacts) ? manifest.artifacts : [];
  const byId = new Map();
  for (const artifact of artifacts) {
    if (artifact?.id) byId.set(artifact.id, artifact);
  }

  for (const [index, artifact] of artifacts.entries()) {
    if (!isModelArtifact(artifact)) continue;
    const label = artifact?.id ? 'artifact ' + artifact.id : 'artifact #' + index;

    for (const field of ['id', 'type', 'source', 'status', 'modelId', 'modelKind', 'notation', 'method', 'abstractionLevel', 'owner']) {
      if (typeof artifact[field] !== 'string' || artifact[field].trim() === '') {
        failures.push(label + ' missing required model field: ' + field);
      }
    }

    if (artifact.type && !['model-view', 'model-diff'].includes(artifact.type)) {
      failures.push(label + ' model artifacts must use type model-view or model-diff');
    }
    if (artifact.modelKind && allowedModelKinds.size > 0 && !allowedModelKinds.has(artifact.modelKind)) {
      failures.push(label + ' has unsupported modelKind: ' + artifact.modelKind);
    }
    if (artifact.notation && allowedNotations.size > 0 && !allowedNotations.has(artifact.notation)) {
      failures.push(label + ' has unsupported notation: ' + artifact.notation);
    }
    if (artifact.method && allowedMethods.size > 0 && !allowedMethods.has(artifact.method)) {
      failures.push(label + ' has unsupported method: ' + artifact.method);
    }
    if (artifact.method && artifact.modelKind && Array.isArray(methodModelKinds[artifact.method]) && !methodModelKinds[artifact.method].includes(artifact.modelKind)) {
      failures.push(label + ' method ' + artifact.method + ' does not allow modelKind ' + artifact.modelKind);
    }

    const facets = asArray(artifact.facets);
    if (artifact.method === 'uwe') {
      if (facets.length === 0) failures.push(label + ' method uwe requires facets');
      const allowed = new Set(allowedFacets.uwe ?? []);
      for (const facet of facets) {
        if (!allowed.has(facet)) failures.push(label + ' has unsupported UWE facet: ' + facet);
      }
    } else if (facets.length > 0 && !allowedFacets[artifact.method]) {
      failures.push(label + ' facets are only configured for methods with an allowedFacets entry');
    }

    const sourcePath = resolveInsideRoot(artifact.source, label + '.source', failures);
    if (sourcePath) {
      if (!fs.existsSync(sourcePath)) {
        failures.push(label + ' source does not exist: ' + artifact.source);
      } else {
        const ext = path.extname(sourcePath).toLowerCase();
        if (!allowedSourceExtensions.has(ext)) {
          failures.push(label + ' source extension is not allowed for canonical model source: ' + ext);
        }
        if (artifact.sourceHash) {
          const actualHash = hashFile(sourcePath);
          if (artifact.sourceHash !== actualHash) failures.push(label + ' sourceHash is stale for ' + artifact.source);
        }
      }
    }

    if (artifact.reviewSurface) {
      const reviewPath = resolveInsideRoot(artifact.reviewSurface, label + '.reviewSurface', failures);
      if (reviewPath) {
        if (path.extname(reviewPath) !== '.html') failures.push(label + ' reviewSurface must be an HTML review artifact');
        if (!reviewPath.startsWith(reviewRoot + path.sep) && reviewPath !== reviewRoot) {
          failures.push(label + ' reviewSurface must be under ' + relativeForMessage(reviewRoot));
        }
        if (!reviewPath.startsWith(modelReviewRoot + path.sep) && reviewPath !== modelReviewRoot) {
          failures.push(label + ' modeling reviewSurface should be under ' + relativeForMessage(modelReviewRoot));
        }
        if (artifact.status === 'ready' && !fs.existsSync(reviewPath)) {
          failures.push(label + ' ready model review surface does not exist: ' + artifact.reviewSurface);
        }
      }
    } else if (artifact.status === 'ready' && (generatedReviewRequired || artifact.type === 'model-diff' || artifact.reviewRequired === true)) {
      failures.push(label + ' needs a generated HTML reviewSurface');
    }

    if (artifact.status === 'ready' && (!Array.isArray(artifact.evidenceLinks) || artifact.evidenceLinks.length === 0)) {
      failures.push(label + ' ready model artifact needs evidenceLinks');
    }

    if (evidenceDefaultKinds.has(artifact.modelKind) && artifact.canonical === true && artifact.authored !== true) {
      failures.push(label + ' modelKind ' + artifact.modelKind + ' defaults to generated evidence; set authored=true before marking it canonical');
    }
    if (authoredOrEvidenceKinds.has(artifact.modelKind) && artifact.canonical === true && artifact.authored !== true) {
      failures.push(label + ' modelKind ' + artifact.modelKind + ' must be explicitly authored=true before canonical=true');
    }

    if (artifact.type === 'model-diff') {
      if (artifact.canonical === true) failures.push(label + ' model-diff cannot be canonical; the source diff is canonical');
      const diff = artifact.diff ?? {};
      for (const field of ['beforeArtifactId', 'afterArtifactId', 'method', 'reviewSurface']) {
        if (typeof diff[field] !== 'string' || diff[field].trim() === '') {
          failures.push(label + ' missing diff.' + field);
        }
      }
      if (diff.method && !['source', 'semantic'].includes(diff.method)) {
        failures.push(label + ' diff.method must be source or semantic');
      }
      for (const field of ['beforeArtifactId', 'afterArtifactId']) {
        if (diff[field] && !byId.has(diff[field])) {
          failures.push(label + ' diff.' + field + ' references unknown artifact: ' + diff[field]);
        }
      }
      if (diff.reviewSurface && artifact.reviewSurface && diff.reviewSurface !== artifact.reviewSurface) {
        failures.push(label + ' diff.reviewSurface must match artifact.reviewSurface');
      }
    }
  }
}

if (failures.length > 0) {
  console.error('Model artifact policy failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Model artifact policy passed');
`
}

func developerModelInventoryPolicyScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = fs.existsSync(configPath) ? JSON.parse(fs.readFileSync(configPath, 'utf8')) : {};
const manifestPath = path.join(root, config.capabilities?.developerArtifacts?.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
const inventoryPath = path.join(root, 'docs/artifacts/source/models/model-inventory.md');
const failures = [];
const tick = String.fromCharCode(96);

function read(filePath) {
  return fs.existsSync(filePath) ? fs.readFileSync(filePath, 'utf8') : '';
}

function tableRows(markdown) {
  return markdown
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.startsWith('| ' + tick))
    .map((line) => line.split('|').slice(1, -1).map((cell) => cell.trim()));
}

function stripCode(value) {
  const text = String(value ?? '');
  return text.startsWith(tick) && text.endsWith(tick) ? text.slice(1, -1) : text;
}

if (!fs.existsSync(manifestPath)) failures.push('missing artifact manifest');
if (!fs.existsSync(inventoryPath)) failures.push('missing model inventory');

const manifest = fs.existsSync(manifestPath) ? JSON.parse(read(manifestPath)) : { artifacts: [] };
const inventory = read(inventoryPath);
const modelArtifacts = (manifest.artifacts ?? []).filter((artifact) => artifact.type === 'model-view' || artifact.type === 'model-diff');
const rows = tableRows(inventory);
const byModelId = new Map(rows.map((row) => [stripCode(row[0]), row]));

for (const artifact of modelArtifacts) {
  const row = byModelId.get(artifact.modelId);
  if (!row) {
    failures.push('inventory missing modelId: ' + artifact.modelId);
    continue;
  }
  if (row.length >= 8) {
    const [modelId, kind, method, owner, source, touchpoints, evidence, reviewSurface] = row;
    if (stripCode(modelId) !== artifact.modelId) failures.push(artifact.modelId + ' inventory modelId mismatch');
    if (kind !== artifact.modelKind) failures.push(artifact.modelId + ' inventory kind mismatch: ' + kind + ' != ' + artifact.modelKind);
    if (method !== artifact.method) failures.push(artifact.modelId + ' inventory method mismatch: ' + method + ' != ' + artifact.method);
    if (owner !== artifact.owner) failures.push(artifact.modelId + ' inventory owner mismatch: ' + owner + ' != ' + artifact.owner);
    if (stripCode(source) !== artifact.source) failures.push(artifact.modelId + ' inventory source mismatch: ' + source + ' != ' + artifact.source);
    if (artifact.reviewSurface && stripCode(reviewSurface) !== artifact.reviewSurface) failures.push(artifact.modelId + ' inventory review surface mismatch: ' + reviewSurface + ' != ' + artifact.reviewSurface);
    if (Array.isArray(artifact.implementationTouchpoints) && artifact.implementationTouchpoints.length > 0) {
      const hasPrimaryTouchpoint = artifact.implementationTouchpoints.some((touchpoint) => touchpoints.includes(touchpoint));
      if (!hasPrimaryTouchpoint) failures.push(artifact.modelId + ' inventory touchpoints do not include any manifest implementation touchpoint');
    }
    if (Array.isArray(artifact.evidenceLinks) && artifact.evidenceLinks.length > 0) {
      const hasEvidence = artifact.evidenceLinks.some((link) => evidence.includes(link));
      if (!hasEvidence) failures.push(artifact.modelId + ' inventory evidence does not include any manifest evidence link');
    }
  }
}

for (const modelId of byModelId.keys()) {
  if (!modelArtifacts.some((artifact) => artifact.modelId === modelId)) failures.push('manifest missing inventory modelId: ' + modelId);
}

if (failures.length > 0) {
  console.error('Model inventory check failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Model inventory check passed');
`
}

func developerArtifactUweWorkspaceRuntimeScript() string {
	return `(function () {
  function text(value, fallback) {
    return value || fallback || "";
  }

  function readJson(value, fallback) {
    if (!value) return fallback;
    try {
      return JSON.parse(value);
    } catch (error) {
      return fallback;
    }
  }

  function nodeDataFrom(element) {
    return {
      id: element.dataset.uweId,
      name: text(element.dataset.uweLabel, element.dataset.uweId),
      stereo: text(element.dataset.uweStereo, "«navigationClass»"),
      type: text(element.dataset.uweType, "navigation"),
      packageName: text(element.dataset.uwePackage, "Navigation"),
      route: text(element.dataset.uweRoute, ""),
      role: text(element.dataset.uweRole, ""),
      actions: text(element.dataset.uweActions, "Inspect available actions in the source matrix."),
      effect: text(element.dataset.uweEffect, "Effect not recorded."),
      focus: text(element.dataset.uweFocus, ""),
      crop: readJson(element.dataset.uweCrop, null),
      annotations: readJson(element.dataset.uweAnnotations, []),
      screenshot: text(element.dataset.uweScreenshot, "")
    };
  }

  function edgeDataFrom(element) {
    return {
      id: element.dataset.uweFrom + "__" + element.dataset.uweTo + "__" + text(element.dataset.uweLabel, "link"),
      source: element.dataset.uweFrom,
      target: element.dataset.uweTo,
      label: text(element.dataset.uweLabel, "«navigationLink»")
    };
  }

  function classFor(node) {
    var stereo = String(node.stereo || "").toLowerCase();
    var facets = String(node.type || "").toLowerCase();
    if (stereo.indexOf("processclass") >= 0) return "uwe-process";
    if (facets.indexOf("adaptation") >= 0 || facets.indexOf("denied") >= 0) return "uwe-adaptation";
    if (facets.indexOf("access") >= 0) return "uwe-access";
    return "uwe-navigation";
  }

  function percent(value) {
    return (Math.max(0, Math.min(1, Number(value) || 0)) * 100).toFixed(2) + "%";
  }

  function primaryBounds(data) {
    if (data && data.crop) return data.crop;
    if (data && Array.isArray(data.annotations) && data.annotations[0]) return data.annotations[0].bounds;
    return null;
  }

  function renderAnnotationLayer(layer, data) {
    if (!layer) return;
    layer.innerHTML = "";
    var annotations = Array.isArray(data.annotations) ? data.annotations : [];
    annotations.forEach(function (annotation, index) {
      var bounds = annotation.bounds;
      if (!bounds) return;
      var box = document.createElement("span");
      box.className = "uwe-screenshot-focus-box";
      box.style.left = percent(bounds.x);
      box.style.top = percent(bounds.y);
      box.style.width = percent(bounds.w);
      box.style.height = percent(bounds.h);
      var label = document.createElement("span");
      label.textContent = String(index + 1) + ". " + (annotation.label || "focus");
      box.appendChild(label);
      layer.appendChild(box);
    });
  }

  function renderCrop(workspace, data) {
    var crop = workspace.querySelector("[data-uwe-focus-crop]");
    var bounds = primaryBounds(data);
    if (!crop || !data.screenshot || !bounds) {
      if (crop) crop.classList.remove("active");
      return;
    }
    crop.style.backgroundImage = "url(\"" + data.screenshot.replace(/"/g, "%22") + "\")";
    crop.style.backgroundSize = (100 / bounds.w).toFixed(2) + "% " + (100 / bounds.h).toFixed(2) + "%";
    crop.style.backgroundPosition = (bounds.x >= 1 ? "100" : ((bounds.x / Math.max(1 - bounds.w, 0.001)) * 100).toFixed(2)) + "% " + (bounds.y >= 1 ? "100" : ((bounds.y / Math.max(1 - bounds.h, 0.001)) * 100).toFixed(2)) + "%";
    crop.classList.add("active");
  }

  function setInspector(workspace, data) {
    workspace.uweSelectedNode = data;
    workspace.querySelector("[data-uwe-inspector-stereo]").textContent = data.stereo || "UWE node";
    workspace.querySelector("[data-uwe-inspector-title]").textContent = data.name || data.id;
    workspace.querySelector("[data-uwe-inspector-package]").textContent = data.packageName || "Navigation";
    workspace.querySelector("[data-uwe-inspector-route]").textContent = data.route || "state";
    workspace.querySelector("[data-uwe-inspector-role]").textContent = data.role || "all roles";
    workspace.querySelector("[data-uwe-inspector-actions]").textContent = data.actions || "No action inventory recorded.";
    workspace.querySelector("[data-uwe-inspector-effect]").textContent = data.effect || "No side effect recorded.";
    workspace.querySelector("[data-uwe-inspector-focus]").textContent = data.focus || (data.annotations && data.annotations[0] && data.annotations[0].label) || "No focused screenshot annotation recorded.";
    var img = workspace.querySelector("[data-uwe-inspector-image]");
    if (img && data.screenshot) {
      img.src = data.screenshot;
      img.alt = (data.name || data.id) + " screenshot";
      img.dataset.uweCaption = (data.stereo || "UWE node") + " " + (data.name || data.id) + ": " + (data.effect || "");
    }
    renderAnnotationLayer(workspace.querySelector("[data-uwe-annotation-layer]"), data);
    renderCrop(workspace, data);
  }

  function setActiveNodeButton(workspace, id) {
    workspace.querySelectorAll("[data-uwe-focus-node]").forEach(function (button) {
      button.classList.toggle("active", button.dataset.uweFocusNode === id);
    });
  }

  function setActivePackageButton(workspace, packageName) {
    workspace.querySelectorAll("[data-uwe-action]").forEach(function (button) {
      var action = button.dataset.uweAction || "";
      button.classList.toggle("active", action === "package:" + packageName);
    });
  }

  function openLightbox(workspace) {
    var img = workspace.querySelector("[data-uwe-inspector-image]");
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    var lightboxImg = workspace.querySelector("[data-uwe-lightbox-image]");
    var caption = workspace.querySelector("[data-uwe-lightbox-caption]");
    if (!img || !lightbox || !lightboxImg || !img.src) return;
    lightboxImg.src = img.src;
    lightboxImg.alt = img.alt || "Selected UWE screenshot";
    if (caption) caption.textContent = img.dataset.uweCaption || img.alt || "Selected UWE screenshot";
    renderAnnotationLayer(workspace.querySelector("[data-uwe-lightbox-layer]"), workspace.uweSelectedNode || {});
    lightbox.classList.add("active");
  }

  function closeLightbox(workspace) {
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    if (lightbox) lightbox.classList.remove("active");
  }

  function setFocusMode(workspace, enabled, cy) {
    workspace.classList.toggle("uwe-focus-mode", enabled);
    document.documentElement.classList.toggle("uwe-focus-active", enabled);
    var button = workspace.querySelector("[data-uwe-action=workspace-focus]");
    if (button) button.textContent = enabled ? "Exit focus" : "Focus workspace";
    setTimeout(function () {
      if (cy) {
        cy.resize();
        cy.fit(undefined, 42);
      }
    }, 80);
  }

  function initPanZoom() {
    if (!window.svgPanZoom) return;
    document.querySelectorAll("[data-svg-pan-zoom=true] svg").forEach(function (svg) {
      window.svgPanZoom(svg, {
        controlIconsEnabled: true,
        fit: true,
        center: true,
        minZoom: 0.1,
        maxZoom: 20,
        zoomScaleSensitivity: 0.25
      });
    });
    document.querySelectorAll(".viewer-badge").forEach(function (el) {
      el.textContent = "svg-pan-zoom active: drag, wheel, +/- controls";
    });
  }

  function initWorkspace(workspace) {
    var graphHost = workspace.querySelector("[data-uwe-cy]");
    var badge = workspace.querySelector("[data-uwe-runtime-badge]");
    var nodeElements = Array.prototype.slice.call(workspace.querySelectorAll("[data-uwe-node]"));
    var edgeElements = Array.prototype.slice.call(workspace.querySelectorAll("[data-uwe-edge]"));
    var nodes = nodeElements.map(nodeDataFrom);
    var edges = edgeElements.map(edgeDataFrom).filter(function (edge) {
      return nodes.some(function (node) { return node.id === edge.source; }) && nodes.some(function (node) { return node.id === edge.target; });
    });
    if (!graphHost || !window.cytoscape) {
      if (badge) badge.textContent = "Graph workspace fallback: Cytoscape runtime unavailable";
      return;
    }
    if (window.cytoscapeDagre) window.cytoscape.use(window.cytoscapeDagre);
    graphHost.innerHTML = "";
    var cy = window.cytoscape({
      container: graphHost,
      elements: nodes.map(function (node) {
        return { group: "nodes", data: node, classes: classFor(node) };
      }).concat(edges.map(function (edge) {
        return { group: "edges", data: edge };
      })),
      minZoom: 0.12,
      maxZoom: 2.8,
      wheelSensitivity: 0.18,
      style: [
        {
          selector: "node",
          style: {
            shape: "round-rectangle",
            width: 232,
            height: 168,
            "background-color": "#ffffff",
            "background-image": "data(screenshot)",
            "background-fit": "contain",
            "background-opacity": 0.92,
            "border-color": "#111827",
            "border-width": 1.5,
            label: "data(name)",
            "font-family": "Segoe UI, Arial, sans-serif",
            "font-weight": 800,
            "font-size": 12,
            color: "#111827",
            "text-background-color": "#ffffff",
            "text-background-opacity": 0.92,
            "text-background-padding": 5,
            "text-valign": "bottom",
            "text-halign": "center",
            "overlay-opacity": 0
          }
        },
        { selector: ".uwe-process", style: { "border-color": "#a15c07", "border-style": "dashed" } },
        { selector: ".uwe-access", style: { "border-color": "#2457c5" } },
        { selector: ".uwe-adaptation", style: { "border-color": "#b42318", "border-style": "dashed" } },
        {
          selector: "edge",
          style: {
            width: 2,
            "line-color": "#64748b",
            "target-arrow-color": "#64748b",
            "target-arrow-shape": "triangle",
            "curve-style": "bezier",
            label: "data(label)",
            "font-size": 10,
            "font-family": "Segoe UI, Arial, sans-serif",
            color: "#334155",
            "text-background-color": "#ffffff",
            "text-background-opacity": 0.85,
            "text-background-padding": 2
          }
        },
        { selector: ":selected", style: { "border-width": 4, "border-color": "#0f766e", "line-color": "#0f766e", "target-arrow-color": "#0f766e" } }
      ],
      layout: {
        name: window.cytoscapeDagre ? "dagre" : "grid",
        rankDir: "LR",
        nodeSep: 70,
        edgeSep: 24,
        rankSep: 132,
        fit: true,
        padding: 36
      }
    });
    workspace.uweCy = cy;
    workspace.querySelector("[data-uwe-stat=nodes]").textContent = String(nodes.length);
    workspace.querySelector("[data-uwe-stat=edges]").textContent = String(edges.length);
    workspace.querySelector("[data-uwe-stat=packages]").textContent = String(new Set(nodes.map(function (node) { return node.packageName; })).size);
    if (badge) badge.textContent = "Cytoscape + dagre active: wheel zoom, drag pan, click inspect";
    if (nodes[0]) setInspector(workspace, nodes[0]);
    if (nodes[0]) setActiveNodeButton(workspace, nodes[0].id);
    if (nodes[0]) setActivePackageButton(workspace, nodes[0].packageName);
    cy.on("tap", "node", function (event) {
      var data = event.target.data();
      setInspector(workspace, data);
      setActiveNodeButton(workspace, data.id);
      setActivePackageButton(workspace, data.packageName);
    });
    workspace.querySelectorAll("[data-uwe-action]").forEach(function (button) {
      button.addEventListener("click", function () {
        var action = button.dataset.uweAction;
        if (action === "fit") cy.fit(undefined, 36);
        if (action === "layout") cy.layout({ name: window.cytoscapeDagre ? "dagre" : "grid", rankDir: "LR", nodeSep: 70, edgeSep: 24, rankSep: 132, fit: true, padding: 36 }).run();
        if (action === "workspace-focus") setFocusMode(workspace, !workspace.classList.contains("uwe-focus-mode"), cy);
        if (action && action.indexOf("package:") === 0) {
          var packageName = action.slice("package:".length);
          var collection = cy.nodes().filter(function (node) { return node.data("packageName") === packageName; });
          if (collection.length > 0) cy.fit(collection.union(collection.connectedEdges()), 56);
          setActivePackageButton(workspace, packageName);
        }
      });
    });
    workspace.querySelectorAll("[data-uwe-focus-node]").forEach(function (button) {
      button.addEventListener("click", function () {
        var id = button.dataset.uweFocusNode;
        var node = nodes.find(function (candidate) { return candidate.id === id; });
        if (!node) return;
        setInspector(workspace, node);
        setActiveNodeButton(workspace, node.id);
        setActivePackageButton(workspace, node.packageName);
        var cyNode = cy.getElementById(id);
        if (cyNode && cyNode.length > 0) {
          cy.nodes().unselect();
          cyNode.select();
          cy.fit(cyNode.union(cyNode.connectedEdges()), 72);
        }
      });
    });
    var inspectorImage = workspace.querySelector("[data-uwe-inspector-image]");
    if (inspectorImage) inspectorImage.addEventListener("click", function () { openLightbox(workspace); });
    var openScreenshotButton = workspace.querySelector("[data-uwe-open-screenshot]");
    if (openScreenshotButton) openScreenshotButton.addEventListener("click", function () { openLightbox(workspace); });
    workspace.querySelectorAll("[data-uwe-lightbox-close]").forEach(function (button) {
      button.addEventListener("click", function () { closeLightbox(workspace); });
    });
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    if (lightbox) {
      lightbox.addEventListener("click", function (event) {
        if (event.target === lightbox) closeLightbox(workspace);
      });
    }
    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        closeLightbox(workspace);
        if (workspace.classList.contains("uwe-focus-mode")) setFocusMode(workspace, false, cy);
      }
    });
  }

  initPanZoom();
  document.querySelectorAll(".uwe-engine-workspace").forEach(initWorkspace);
  document.documentElement.classList.add("uwe-workspace-active");
})();
`
}

func developerArtifactPolicyScript() string {
	return `import fs from 'node:fs';
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
const reviewedUweWorkspaceSurfaces = new Set((manifest.artifacts ?? [])
  .filter((artifact) => artifact.htmlInteractionLane === 'reviewed-uwe-workspace' && typeof artifact.reviewSurface === 'string')
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
const allowedSvgPanZoomInitializer = 'document.querySelectorAll("[data-svg-pan-zoom=true] svg").forEach(function(svg){svgPanZoom(svg,{controlIconsEnabled:true,fit:true,center:true,minZoom:.1,maxZoom:20,zoomScaleSensitivity:.25});});document.querySelectorAll(".viewer-badge").forEach(function(el){el.textContent="svg-pan-zoom active: drag, wheel, +/- controls";});document.documentElement.classList.add("svg-pan-zoom-active");';
const allowedCytoscapeRuntime = fs.existsSync(path.join(root, 'node_modules', 'cytoscape', 'dist', 'cytoscape.min.js'))
  ? fs.readFileSync(path.join(root, 'node_modules', 'cytoscape', 'dist', 'cytoscape.min.js'), 'utf8')
  : '';
const allowedDagreRuntime = fs.existsSync(path.join(root, 'node_modules', 'dagre', 'dist', 'dagre.min.js'))
  ? fs.readFileSync(path.join(root, 'node_modules', 'dagre', 'dist', 'dagre.min.js'), 'utf8')
  : '';
const allowedCytoscapeDagreRuntime = fs.existsSync(path.join(root, 'node_modules', 'cytoscape-dagre', 'cytoscape-dagre.js'))
  ? fs.readFileSync(path.join(root, 'node_modules', 'cytoscape-dagre', 'cytoscape-dagre.js'), 'utf8')
  : '';
const allowedUweWorkspaceRuntime = fs.existsSync(path.join(root, 'scripts', 'uwe-workspace-runtime.js'))
  ? fs.readFileSync(path.join(root, 'scripts', 'uwe-workspace-runtime.js'), 'utf8')
  : '';

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
  const reviewedUweWorkspace = reviewedUweWorkspaceSurfaces.has(path.normalize(path.resolve(filePath)));
  const expectedCsp = reviewedSvgPanZoom || reviewedUweWorkspace ? reviewedSvgPanZoomCsp : requiredCsp;
  if (!html.includes('Content-Security-Policy') || !html.includes(expectedCsp)) failures.push('missing required CSP meta tag');
  const scriptMatches = [...html.matchAll(/<script>([\s\S]*?)<\/script>/gi)].map((match) => match[1]);
  if (!reviewedSvgPanZoom && !reviewedUweWorkspace && scriptMatches.length > 0) failures.push('blocked script tag');
  if (reviewedSvgPanZoom) {
    if (scriptMatches.length !== 2) failures.push('reviewed-svg-pan-zoom requires exactly two inline scripts');
    if (scriptMatches[0] !== allowedSvgPanZoomRuntime) failures.push('reviewed-svg-pan-zoom runtime does not match bundled svg-pan-zoom');
    if (scriptMatches[1] !== allowedSvgPanZoomInitializer) failures.push('reviewed-svg-pan-zoom initializer is not the approved static initializer');
  }
  if (reviewedUweWorkspace) {
    if (scriptMatches.length !== 5) failures.push('reviewed-uwe-workspace requires exactly five inline scripts');
    if (scriptMatches[0] !== allowedSvgPanZoomRuntime) failures.push('reviewed-uwe-workspace svg-pan-zoom runtime does not match bundled dependency');
    if (scriptMatches[1] !== allowedCytoscapeRuntime) failures.push('reviewed-uwe-workspace Cytoscape runtime does not match bundled dependency');
    if (scriptMatches[2] !== allowedDagreRuntime) failures.push('reviewed-uwe-workspace dagre runtime does not match bundled dependency');
    if (scriptMatches[3] !== allowedCytoscapeDagreRuntime) failures.push('reviewed-uwe-workspace cytoscape-dagre runtime does not match bundled dependency');
    if (scriptMatches[4] !== allowedUweWorkspaceRuntime) failures.push('reviewed-uwe-workspace initializer does not match scripts/uwe-workspace-runtime.js');
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
`
}

func developerAgentLoopPolicyScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const configPath = path.join(root, '.skill-harness', 'project.json');
const failures = [];

function readJSON(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch (error) {
    failures.push('invalid or unreadable JSON: ' + path.relative(root, filePath) + ' (' + error.message + ')');
    return {};
  }
}

function expectArrayIncludes(values, required, label) {
  if (!Array.isArray(values)) {
    failures.push(label + ' must be an array');
    return;
  }
  for (const value of required) {
    if (!values.includes(value)) failures.push(label + ' missing required value: ' + value);
  }
}

function resolveInsideRoot(relativePath, label) {
  if (typeof relativePath !== 'string' || relativePath.trim() === '') {
    failures.push(label + ' must be a non-empty repo-relative path');
    return null;
  }
  if (path.isAbsolute(relativePath)) {
    failures.push(label + ' must be repo-relative: ' + relativePath);
    return null;
  }
  const resolved = path.resolve(root, relativePath);
  if (!resolved.startsWith(root + path.sep) && resolved !== root) {
    failures.push(label + ' escapes the repo root: ' + relativePath);
    return null;
  }
  return resolved;
}

if (!fs.existsSync(configPath)) {
  failures.push('missing .skill-harness/project.json');
} else {
  const config = readJSON(configPath);
  const developerArtifacts = config.capabilities?.developerArtifacts;
  const agentLoop = developerArtifacts?.agentLoop;

  if (!developerArtifacts?.enabled) failures.push('developerArtifacts must be enabled');
  if (developerArtifacts?.requestedProfile !== 'agent-loop') {
    failures.push('requestedProfile must be agent-loop for this checker');
  }
  if (developerArtifacts?.profile !== 'dual') failures.push('agent-loop profile must resolve to dual');
  if (developerArtifacts?.specialization !== 'self-improving-agent-loop') {
    failures.push('specialization must be self-improving-agent-loop');
  }
  expectArrayIncludes(developerArtifacts?.artifactTypes, ['agent-loop', 'trace-review', 'eval-report', 'learning-proposal'], 'developerArtifacts.artifactTypes');

  if (!agentLoop?.enabled) failures.push('agentLoop.enabled must be true');
  if (!['beads', 'explicit-human-request'].includes(agentLoop?.defaultIssueTool)) {
    failures.push('agentLoop.defaultIssueTool must be beads or explicit-human-request');
  }
  expectArrayIncludes(agentLoop?.phases, ['sense', 'model', 'plan', 'act', 'gate', 'learn'], 'agentLoop.phases');
  expectArrayIncludes(agentLoop?.qualityGates, ['issue or explicit request captured before mutation', 'tests or executable validators run'], 'agentLoop.qualityGates');
  expectArrayIncludes(agentLoop?.learningOutputs, ['agent-loop artifact'], 'agentLoop.learningOutputs');
  if (agentLoop?.defaultIssueTool === 'beads') {
    expectArrayIncludes(agentLoop?.learningOutputs, ['beads issue', 'bd remember insight'], 'agentLoop.learningOutputs');
  }
  expectArrayIncludes(agentLoop?.humanApprovalRequiredFor, ['permission expansion', 'destructive filesystem or database action'], 'agentLoop.humanApprovalRequiredFor');

  const traceDir = resolveInsideRoot(agentLoop?.traceDir, 'agentLoop.traceDir');
  if (traceDir && !fs.existsSync(traceDir)) failures.push('missing agentLoop.traceDir: ' + agentLoop.traceDir);

  const playbook = resolveInsideRoot(agentLoop?.playbook, 'agentLoop.playbook');
  if (playbook && !fs.existsSync(playbook)) failures.push('missing agent loop playbook: ' + agentLoop.playbook);

  const template = path.join(root, 'docs', 'artifacts', 'templates', 'agent-loop-artifact.md');
  if (!fs.existsSync(template)) failures.push('missing agent-loop artifact template');

  const gitignorePath = path.join(root, '.gitignore');
  const gitignore = fs.existsSync(gitignorePath) ? fs.readFileSync(gitignorePath, 'utf8') : '';
  if (!gitignore.split(/\r?\n/).map((line) => line.trim()).includes('generated/agent-runs/')) {
    failures.push('.gitignore must include generated/agent-runs/');
  }
}

if (failures.length > 0) {
  console.error('Agent loop policy failed:');
  for (const failure of failures) console.error('- ' + failure);
  process.exit(1);
}

console.log('Agent loop policy passed');
`
}

func installBeads(projectDir string) (beadsInstallMode, error) {
	if _, err := findBeadsBinary(); err == nil {
		return beadsSystem, nil
	}

	if goCmd, err := findGoCommand(); err == nil {
		if err := runCommand(projectDir, goCmd, "install", "github.com/steveyegge/beads/cmd/bd@latest"); err == nil {
			return beadsSystem, nil
		}
	}

	if runtime.GOOS == "windows" {
		powerShellCmd, err := findPowerShellCommand()
		if err != nil {
			return beadsDisabled, fmt.Errorf("failed to find PowerShell for Beads installation")
		}
		err = runCommand(
			projectDir,
			powerShellCmd,
			"-NoProfile",
			"-ExecutionPolicy",
			"Bypass",
			"-Command",
			"irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex",
		)
		if err == nil {
			return beadsSystem, nil
		}
		return beadsDisabled, fmt.Errorf("failed to install Beads via npm, go, or PowerShell installer")
	}

	err := runCommand(projectDir, "bash", "-lc", "curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash")
	if err == nil {
		return beadsSystem, nil
	}
	return beadsDisabled, fmt.Errorf("failed to install Beads via npm, go, or shell installer")
}

func initBeads(projectDir string, mode beadsInstallMode) error {
	bdPath, err := findBeadsBinary()
	if err != nil {
		return err
	}
	return runCommand(projectDir, bdPath, "init")
}

func findBeadsBinary() (string, error) {
	if path, err := exec.LookPath("bd"); err == nil {
		return path, nil
	}
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			candidate := filepath.Join(local, "Programs", "bd", "bd.exe")
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "go", "bin", binaryName("bd"))
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("bd binary not found after Beads installation")
}

func binaryName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func findGoCommand() (string, error) {
	if path, err := exec.LookPath("go"); err == nil {
		return path, nil
	}
	if runtime.GOOS == "windows" {
		candidate := filepath.Join(os.Getenv("ProgramFiles"), "Go", "bin", "go.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("go not found")
}

func findPowerShellCommand() (string, error) {
	if path, err := exec.LookPath("pwsh"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("powershell"); err == nil {
		return path, nil
	}
	return "", errors.New("PowerShell not found")
}

func findRepoRoot() (string, error) {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(exe))
	}
	for _, start := range candidates {
		if root := walkForRoot(start); root != "" {
			return root, nil
		}
	}
	return "", errors.New("could not locate skill-harness repo root")
}

func walkForRoot(start string) string {
	dir := start
	for {
		if isRepoRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isRepoRoot(dir string) bool {
	_, depErr := os.Stat(filepath.Join(dir, "scripts", "dependencies.json"))
	_, agentErr := os.Stat(filepath.Join(dir, ".claude", "agents"))
	return depErr == nil && agentErr == nil
}

func csvList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return unique(out)
}

func unique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sortedKeys[T any](input map[string]T) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func readLine(reader *bufio.Reader) string {
	text, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		exitOnErr(err)
	}
	return strings.TrimSpace(text)
}

func printUsage(loadouts loadoutConfig, deps dependencyConfig) {
	fmt.Println("skill-harness")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list [--agents] [--packs]")
	fmt.Println("  resolve [--dir path] [--json] [--strict]")
	fmt.Println("  audit-project [--dir path] [--json]")
	fmt.Println("  bootstrap --agent-native [--dir path] [--json]")
	fmt.Println("  update-project [--dir path] [--json] [--write-lock]")
	fmt.Println("  repo init|audit|drift|update|trim|sync [--dir path] [--json]")
	fmt.Println("  install [--dir path] [--all] [--interactive] [--packs-only] [--agents-only] [--agents=a,b] [--packs=x,y]")
	fmt.Println("  setup-project [--dir path] [--scope auto|root|workspace] [--package-manager auto|npm|pnpm|yarn|bun] [--developer-artifacts-profile auto|codex-app|claude-desktop|cli|tui|media|agent-loop|none] [--modeling-mode auto|off|baseline|uml-first] [--enable-modeling] [--skip-modeling] [--install-only] [--skip-noslop] [--skip-agent-docs] [--skip-beads] [--beads-worktrees] [--skip-developer-artifacts] [--skip-claude-settings]")
	fmt.Println("  beads-worktrees [--dir path] [--force]")
	fmt.Println("  update")
	fmt.Println("  check [--dir path] [--all] [--interactive] [--agents=a,b]")
	fmt.Println("  render [--dir path] [--all] [--interactive] [--agents=a,b]")
	fmt.Println("  uninstall [--all] [--interactive] [--agents=a,b]")
	fmt.Println()
	fmt.Printf("Configured agents: %d\n", len(loadouts))
	fmt.Printf("Configured packs:  %d\n", len(deps.Repos))
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
