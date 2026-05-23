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
	artifactProfileNone          artifactProfile = "none"
)

type projectSetupContext struct {
	TargetDir      string
	OperationDir   string
	MonorepoRoot   string
	Monorepo       bool
	Scope          projectScope
	PackageManager packageManager
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
	artifactProfileValue := fs.String("artifact-profile", string(artifactProfileAuto), "Developer artifact profile: auto, markdown, html, dual, or none.")
	developerArtifactsProfileValue := fs.String("developer-artifacts-profile", string(artifactProfileAuto), "Developer artifact profile: auto, codex-app, claude-desktop, cli, tui, markdown, html, dual, or none.")
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

	if *installOnly {
		fmt.Println(projectSetupSummary("Installed project tooling", ctx))
		return
	}

	if !*skipAgentDocs {
		exitOnErr(runLocalTool(ctx.OperationDir, ctx.PackageManager, "agent-docs", "init"))
	}
	if !*skipArtifacts && !*skipDeveloperArtifacts && artifactProfile != artifactProfileNone {
		exitOnErr(writeDeveloperArtifactScaffold(ctx.OperationDir, artifactProfile, !*skipAgentDocs))
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
	case artifactProfileNone:
		return artifactProfileNone, nil
	default:
		return "", fmt.Errorf("unsupported artifact profile: %s", value)
	}
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

func updatePackageScripts(projectDir string, agentDocsEnabled bool) error {
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
	}
	if agentDocsEnabled {
		defaultScripts["docs:check"] = "agent-docs check"
		defaultScripts["docs:generate"] = "agent-docs generate"
		defaultScripts["docs:report"] = "agent-docs report"
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

func writeDeveloperArtifactScaffold(projectDir string, profile artifactProfile, agentDocsEnabled bool) error {
	effectiveProfile := effectiveArtifactProfile(profile)
	dirs := []string{
		filepath.Join(projectDir, "docs", "artifacts", "source"),
		filepath.Join(projectDir, "docs", "artifacts", "templates"),
		filepath.Join(projectDir, "generated", "review"),
		filepath.Join(projectDir, ".skill-harness"),
		filepath.Join(projectDir, "scripts"),
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
				"canonical": map[string]any{
					"formats": []string{"markdown", "toon"},
					"tooling": canonicalTooling,
					"paths":   []string{"docs", "docs/artifacts/source"},
				},
				"artifactTypes": []string{
					"decision",
					"plan",
					"spec",
					"handoff",
					"evidence-pack",
					"blast-radius",
					"architecture-view",
					"model-view",
					"review-dashboard",
				},
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
				"modelPolicy": map[string]any{
					"canonicalSource":        true,
					"generatedReviewOnly":    true,
					"renderDiagramsOffline":  true,
					"defaultReviewEmbedding": "inline-svg",
					"allowedNotations":       []string{"mermaid", "markdown", "toon", "plantuml"},
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
				},
				"reviewSurface": map[string]any{
					"format":          "html",
					"outDir":          "generated/review",
					"commitGenerated": false,
					"openMode":        artifactOpenMode(profile),
				},
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
	if err := updatePackageScripts(projectDir, agentDocsEnabled); err != nil {
		return err
	}
	if err := ensureGitignoreLines(projectDir, []string{"generated/review/"}); err != nil {
		return err
	}

	readmePath := filepath.Join(projectDir, "docs", "artifacts", "README.md")
	if !fileExists(readmePath) {
		if err := os.WriteFile(readmePath, []byte(developerArtifactReadme(profile)), 0o644); err != nil {
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
		return os.WriteFile(checkerPath, []byte(developerArtifactPolicyScript()), 0o644)
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
	default:
		return profile
	}
}

func artifactOpenMode(profile artifactProfile) string {
	switch profile {
	case artifactProfileCodexApp, artifactProfileClaudeDesktop, artifactProfileHTML:
		return "file-preview"
	case artifactProfileCLI, artifactProfileTUI, artifactProfileMarkdown:
		return "path-or-command"
	default:
		return "when-supported"
	}
}

func artifactRequiredCSP() string {
	return "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
}

func developerArtifactReadme(profile artifactProfile) string {
	effectiveProfile := effectiveArtifactProfile(profile)
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

## Model And Diagram Policy

- Keep Mermaid, C4, UML-style, and architecture-space sources in Markdown, TOON, or specgraph-compatible source artifacts.
- Pre-render diagrams into generated HTML as inline SVG or static markup; do not load a browser Mermaid runtime by default.
- Treat Mermaid C4 as a review notation and record the level explicitly: context, container, component, dynamic, or deployment.
- Treat dependency graphs as generated evidence unless the project has a separate model source of truth.
- Link every generated model view back to its source artifact, issue, and evidence.

## HTML Review Policy

- Self-contained static HTML only by default.
- No external scripts, external assets, or network calls unless the project explicitly opts in.
- No inline JavaScript unless the project explicitly opts in and reviews the script.
- Every HTML review artifact must include the required CSP meta tag from .skill-harness/project.json.
- Use semantic headings, landmarks, meaningful link text, and alt text for embedded images.
- No secrets, credentials, tokens, private logs, or customer data.
- Link back to the canonical source artifact and issue.
- Regenerate or discard HTML when the source changes.

Run this policy check before handing off generated HTML:

    node scripts/check-artifact-manifest.mjs
    node scripts/check-artifact-html-policy.mjs
`, profile, effectiveProfile)
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
**Model kind:** sequence | state | class | domain | context | container | component | dynamic | deployment | dependency | use-case | activity | architecture-space
**Notation:** mermaid | markdown | toon | plantuml
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
	fmt.Println("  setup-project [--dir path] [--scope auto|root|workspace] [--package-manager auto|npm|pnpm|yarn|bun] [--developer-artifacts-profile auto|codex-app|claude-desktop|cli|tui|none] [--install-only] [--skip-noslop] [--skip-agent-docs] [--skip-beads] [--beads-worktrees] [--skip-developer-artifacts] [--skip-claude-settings]")
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
