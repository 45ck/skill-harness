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
	"sort"
	"strconv"
	"strings"
)

type dependencyConfig struct {
	Repos  map[string]repoConfig  `json:"repos"`
	Agents map[string]agentConfig `json:"agents"`
}

type repoConfig struct {
	URL string `json:"url"`
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
	case "check":
		runCheck(root, loadouts, os.Args[2:])
	case "render":
		runRender(root, loadouts, os.Args[2:])
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
	showPacks := fs.Bool("packs", false, "List pack repos.")
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
		fmt.Println("Pack repos:")
		for _, name := range sortedKeys(deps.Repos) {
			fmt.Printf("- %s\n", name)
		}
	}
}

func runInstall(root string, deps dependencyConfig, loadouts loadoutConfig, args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	all := fs.Bool("all", false, "Install all packs and all agents.")
	interactive := fs.Bool("interactive", false, "Pick packs and agents interactively.")
	packsOnly := fs.Bool("packs-only", false, "Install only skill packs.")
	agentsOnly := fs.Bool("agents-only", false, "Install only agent files without bootstrapping packs.")
	agentsCSV := fs.String("agents", "", "Comma-separated agent names.")
	packsCSV := fs.String("packs", "", "Comma-separated pack repo names.")
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
		exitOnErr(copyClaudeAgents(root, agents))
		exitOnErr(runPython(root, "scripts/render_codex_agents.py", agentArgs(agents)...))
		exitOnErr(runPython(root, "scripts/check_dependencies.py", agentArgs(agents)...))
	}

	fmt.Printf("Installed %d pack repo(s) and %d agent(s)\n", len(repos), len(agents))
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
