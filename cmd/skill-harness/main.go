package main

import (
	"bufio"
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
	case "install":
		runInstall(root, deps, loadouts, os.Args[2:])
	case "setup-project":
		runSetupProject(root, os.Args[2:])
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

func runInstall(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	all := fs.Bool("all", false, "Install all packs and all agents.")
	interactive := fs.Bool("interactive", false, "Pick packs and agents interactively.")
	packsOnly := fs.Bool("packs-only", false, "Install only dependency repos.")
	agentsOnly := fs.Bool("agents-only", false, "Install only agent files without bootstrapping dependency repos.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	packsCSV := fs.String("packs", "", "Comma-separated dependency repo names.")
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

	agents := resolveAgents(sel, loadouts)
	repos := resolveRepos(sel, deps)

	if !*agentsOnly {
		exitOnErr(runPython(root, "scripts/bootstrap_dependencies.py", bootstrapArgs(sel)...))
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
	fs.Parse(args)

	sel := selection{All: *all, Agent: csvList(*agentsCSV)}
	if *interactive {
		sel.Agent = promptAgentList("Check which agents?", sortedKeys(loadouts))
	}
	agents := resolveAgents(sel, loadouts)
	exitOnErr(runPython(root, "scripts/check_dependencies.py", agentArgs(agents)...))
}

func runRender(root string, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	all := fs.Bool("all", false, "Render all agents.")
	interactive := fs.Bool("interactive", false, "Choose agents interactively.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	fs.Parse(args)

	sel := selection{All: *all, Agent: csvList(*agentsCSV)}
	if *interactive {
		sel.Agent = promptAgentList("Render which agents?", sortedKeys(loadouts))
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
		GeneratedPaths: []string{".skill-harness/setup-proof.json"},
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
	defaultScripts := map[string]string{
		"artifacts:check":          "node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs",
		"artifacts:html:check":     "node scripts/check-artifact-html-policy.mjs",
		"artifacts:manifest:check": "node scripts/check-artifact-manifest.mjs",
		"artifacts:open":           "node scripts/open-artifact-review.mjs",
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
		defaultScripts["artifacts:check"] = "node scripts/check-artifact-manifest.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["artifacts:model:generate"] = "node scripts/generate-model-review.mjs"
		defaultScripts["artifacts:model:check"] = "node scripts/check-model-artifact-policy.mjs"
		defaultScripts["artifacts:model:review"] = "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["models:generate"] = "node scripts/generate-model-review.mjs"
		defaultScripts["models:check"] = "node scripts/check-model-artifact-policy.mjs"
		defaultScripts["models:diff:check"] = "node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-html-policy.mjs"
		defaultScripts["models:open"] = "node scripts/open-artifact-review.mjs generated/review/models/index.html"
		defaultScripts["models:review"] = "node scripts/generate-model-review.mjs && node scripts/check-model-artifact-policy.mjs && node scripts/check-artifact-manifest.mjs && node scripts/check-artifact-html-policy.mjs"
	}
	for name, command := range defaultScripts {
		if _, exists := scripts[name]; !exists {
			scripts[name] = command
		}
	}
	metadata["scripts"] = scripts
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
		filepath.Join(projectDir, "docs", "artifacts", "templates"),
		filepath.Join(projectDir, "generated", "review"),
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
					"formats": []string{"markdown", "toon"},
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
				"modeling":    artifactModelingConfig(mode),
				"modelPolicy": artifactModelPolicy(mode),
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

- Keep canonical decisions, specs, investigations, and handoff notes in Markdown, TOON, or specgraph-compatible sources.
- Treat HTML as a generated review surface for scanning, comparison, diagrams, prototypes, and desktop app previews.
- Do not make generated HTML the only durable source for a decision.
- Record source-backed review surfaces in artifacts.manifest.json so agents and humans can detect stale output.

## Layout

- source/ - canonical artifact sources when they do not belong in a domain-specific docs folder.
- templates/ - local templates for recurring artifact types.
- artifacts.manifest.json - provenance and freshness index for source-backed review artifacts.
- ../../generated/review/ - generated HTML or rich review artifacts for humans.
- ../../generated/media/ - generated demo media for media profile projects.
- ../../generated/agent-runs/ - generated trace receipts and eval summaries for agent-loop profile projects.

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
- No inline JavaScript unless the project explicitly opts in and reviews the script.
- Every HTML review artifact must include the required CSP meta tag from .skill-harness/project.json.
- Use semantic headings, landmarks, meaningful link text, and alt text for embedded images.
- No secrets, credentials, tokens, private logs, or customer data.
- Link back to the canonical source artifact and issue.
- Regenerate or discard HTML when the source changes.
- Open generated HTML with the best human review surface for the current environment: Codex Browser plugin in Codex app, Claude desktop preview/browser in Claude desktop, or node scripts/open-artifact-review.mjs for CLI/system-browser fallback.

Run this policy check before handing off generated HTML:

    node scripts/check-artifact-manifest.mjs
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

func developerModelReviewGeneratorScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
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
    const arrow = line.match(/^"?([^"-]+?)"?\s*(?:-->|->>|--|-\)|-\])\s*"?([^":]+?)"?(?::.*)?$/);
    if (arrow) edges.push([arrow[1].trim(), arrow[2].trim()]);
  }
  if (edges.length === 0) {
    return '<pre>' + escapeHtml(source || 'No diagram source found.') + '</pre>';
  }
  let html = '<div class="flow">';
  for (const [index, edge] of edges.entries()) {
    if (index > 0) html += '<span class="arrow">/</span>';
    html += '<span class="node">' + escapeHtml(edge[0]) + '</span><span class="arrow">-></span><span class="node">' + escapeHtml(edge[1]) + '</span>';
  }
  return html + '</div>';
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

function listItems(values, emptyText) {
  if (!Array.isArray(values) || values.length === 0) return '<p class="muted">' + escapeHtml(emptyText) + '</p>';
  return '<ul>' + values.map((value) => '<li>' + escapeHtml(typeof value === 'string' ? value : JSON.stringify(value)) + '</li>').join('') + '</ul>';
}

function renderArtifact(artifact, byId) {
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
  body += '<section class="tab-panel"><h2>Overview</h2><p>' + escapeHtml(summary) + '</p><div class="meta"><div><strong>Abstraction</strong><br>' + escapeHtml(artifact.abstractionLevel) + '</div><div><strong>Notation</strong><br>' + escapeHtml(artifact.notation) + '</div><div><strong>Canonical Source</strong><br>' + escapeHtml(artifact.source) + '</div><div><strong>Review Surface</strong><br>' + escapeHtml(artifact.reviewSurface || '') + '</div></div></section>';
  body += '<section class="tab-panel"><h2>Visuals</h2>' + diagramSection(source, artifact) + '<h3>Screenshots And Evidence Images</h3>' + gallerySection(artifact) + '</section>';
  body += '<section class="tab-panel"><h2>Canonical Source</h2><pre>' + escapeHtml(source || 'Source not found or not readable.') + '</pre></section>';
  body += '<section class="tab-panel"><h2>Evidence</h2>' + listItems(artifact.evidenceLinks, 'No evidence links are listed yet.') + '<h3>Freshness</h3><div class="meta"><div><strong>Source Hash</strong><br>' + escapeHtml(artifact.sourceHash || '') + '</div><div><strong>Renderer</strong><br>' + escapeHtml(artifact.renderer || 'skill-harness model review generator') + '</div><div><strong>Generated</strong><br>' + escapeHtml(new Date().toISOString()) + '</div></div></section>';
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
for (const artifact of artifacts) {
  const outPath = artifactPath(artifact);
  fs.writeFileSync(outPath, renderArtifact(artifact, byId));
  const reviewSurface = repoPath(outPath);
  artifact.reviewSurface = reviewSurface;
  if (artifact.type === 'model-diff') {
    artifact.diff = artifact.diff ?? {};
    artifact.diff.reviewSurface = reviewSurface;
  }
  indexRows.push('<tr><td>' + escapeHtml(artifact.id) + '</td><td>' + escapeHtml(artifact.modelKind) + '</td><td>' + escapeHtml(artifact.method) + '</td><td>' + escapeHtml(artifact.status) + '</td><td>' + escapeHtml(artifact.source) + '</td><td>' + escapeHtml(reviewSurface) + '</td></tr>');
}

const indexBody = '<section><h1>Model Review Index</h1><p>Static human review surfaces generated from canonical model sources. Edit source first, then regenerate these pages.</p></section><section><table><thead><tr><th>ID</th><th>Kind</th><th>Method</th><th>Status</th><th>Source</th><th>HTML Review</th></tr></thead><tbody>' + indexRows.join('\n') + '</tbody></table></section>';
fs.writeFileSync(path.join(modelReviewDir, 'index.html'), htmlPage('Model Review Index', indexBody));
fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2) + '\n');
console.log('Generated ' + artifacts.length + ' model review artifact(s) in ' + repoPath(modelReviewDir));
`
}

func developerArtifactOpenScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { pathToFileURL } from 'node:url';

const root = process.cwd();
const args = process.argv.slice(2);
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
  if (explicitTarget) {
    const resolved = resolveReviewPath(explicitTarget);
    if (resolved && fs.existsSync(resolved)) return resolved;
    throw new Error('review artifact not found or outside repo: ' + explicitTarget);
  }

  const config = readJSON(path.join(root, '.skill-harness', 'project.json')) ?? {};
  const developerArtifacts = config.capabilities?.developerArtifacts ?? {};
  const reviewDir = developerArtifacts.reviewSurface?.outDir ?? 'generated/review';
  const modelReviewDir = developerArtifacts.modeling?.reviewDir ?? developerArtifacts.modelPolicy?.uml?.reviewDir ?? path.join(reviewDir, 'models');
  const manifestPath = path.join(root, developerArtifacts.manifest?.path ?? 'docs/artifacts/artifacts.manifest.json');
  const manifest = readJSON(manifestPath);
  const manifestTargets = [];
  for (const artifact of Array.isArray(manifest?.artifacts) ? manifest.artifacts : []) {
    if (typeof artifact?.reviewSurface === 'string' && artifact.reviewSurface.endsWith('.html')) {
      const resolved = resolveReviewPath(artifact.reviewSurface);
      if (resolved) manifestTargets.push(resolved);
    }
  }

  const discovered = firstExisting([
    path.join(root, modelReviewDir, 'index.html'),
    path.join(root, reviewDir, 'index.html'),
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

func developerArtifactPolicyScript() string {
	return `import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const configPath = path.join(root, '.skill-harness', 'project.json');
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
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
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...walk(fullPath));
    } else if (entry.isFile() && entry.name.endsWith('.html')) {
      files.push(fullPath);
    }
  }
  return files;
}

function checkFile(filePath) {
  const html = fs.readFileSync(filePath, 'utf8');
  const failures = [];
  if (!html.includes('Content-Security-Policy') || !html.includes(requiredCsp)) {
    failures.push('missing required CSP meta tag');
  }
  for (const pattern of blockedTagPatterns) {
    if (pattern.test(html)) failures.push('blocked tag or preload pattern: ' + pattern);
  }
  for (const pattern of blockedApiPatterns) {
    if (pattern.test(html)) failures.push('blocked browser API: ' + pattern);
  }
  if (externalReferencePattern.test(html)) {
    failures.push('external src/href/action reference');
  }
  return failures;
}

const failures = [];
for (const filePath of walk(reviewRoot)) {
  const fileFailures = checkFile(filePath);
  for (const failure of fileFailures) {
    failures.push(path.relative(root, filePath) + ': ' + failure);
  }
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
	fmt.Println("  install [--all] [--interactive] [--packs-only] [--agents-only] [--agents=a,b] [--packs=x,y]")
	fmt.Println("  setup-project [--dir path] [--scope auto|root|workspace] [--package-manager auto|npm|pnpm|yarn|bun] [--developer-artifacts-profile auto|codex-app|claude-desktop|cli|tui|media|agent-loop|none] [--modeling-mode auto|off|baseline|uml-first] [--enable-modeling] [--skip-modeling] [--install-only] [--skip-noslop] [--skip-agent-docs] [--skip-beads] [--beads-worktrees] [--skip-developer-artifacts] [--skip-claude-settings]")
	fmt.Println("  beads-worktrees [--dir path] [--force]")
	fmt.Println("  update")
	fmt.Println("  check [--all] [--interactive] [--agents=a,b]")
	fmt.Println("  render [--all] [--interactive] [--agents=a,b]")
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
