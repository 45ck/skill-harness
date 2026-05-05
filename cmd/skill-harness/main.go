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
	installOnly := fs.Bool("install-only", false, "Install packages only; skip initialization commands.")
	scopeValue := fs.String("scope", string(projectScopeAuto), "Setup scope: auto, root, or workspace.")
	packageManagerValue := fs.String("package-manager", string(packageManagerAuto), "Package manager: auto, npm, pnpm, yarn, or bun.")
	fs.Parse(args)

	projectDir, err := filepath.Abs(*targetDir)
	exitOnErr(err)

	ctx, err := resolveProjectSetupContext(projectDir, *scopeValue, *packageManagerValue)
	exitOnErr(err)
	exitOnErr(requireSetupCommands(ctx.PackageManager))

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
	fmt.Println("  setup-project [--dir path] [--scope auto|root|workspace] [--package-manager auto|npm|pnpm|yarn|bun] [--install-only] [--skip-noslop] [--skip-agent-docs] [--skip-beads] [--beads-worktrees] [--skip-claude-settings]")
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
