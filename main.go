package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/subut0n/mk/internal/ansi"
	"github.com/subut0n/mk/internal/config"
	"github.com/subut0n/mk/internal/history"
	"github.com/subut0n/mk/internal/i18n"
	"github.com/subut0n/mk/internal/parser"
	"github.com/subut0n/mk/internal/ui"
)

// Version is set via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, ansi.Red+format+ansi.Reset+"\n", args...)
	os.Exit(1)
}

// getPalette returns the ANSI color codes for the given color scheme.
func getPalette(scheme config.ColorScheme) []string {
	switch scheme {
	case config.ColorSchemeDeuteranopia:
		return []string{
			"\033[34m", // Blue
			"\033[33m", // Yellow
			"\033[36m", // Cyan
			"\033[35m", // Magenta
			"\033[94m", // Bright Blue
			"\033[93m", // Bright Yellow
		}
	case config.ColorSchemeTritanopia:
		return []string{
			"\033[31m", // Red
			"\033[35m", // Magenta
			"\033[32m", // Green
			"\033[91m", // Bright Red
			"\033[95m", // Bright Magenta
			"\033[92m", // Bright Green
		}
	case config.ColorSchemeHighContrast:
		return []string{
			"\033[91m", // Bright Red
			"\033[93m", // Bright Yellow
			"\033[92m", // Bright Green
			"\033[96m", // Bright Cyan
			"\033[94m", // Bright Blue
			"\033[95m", // Bright Magenta
		}
	default: // rainbow
		return []string{
			"\033[31m", // Red
			"\033[33m", // Yellow
			"\033[32m", // Green
			"\033[36m", // Cyan
			"\033[34m", // Blue
			"\033[35m", // Magenta
		}
	}
}

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "--version", "-v":
			loadConfigAndSetLang()
			fmt.Printf(i18n.Get().VersionFormat+"\n", Version)
			return
		case "--help", "-h":
			cfg := loadConfigAndSetLang()
			printHelp(getPalette(cfg.Config.ColorScheme))
			return
		case "--config":
			runConfigSetup()
			return
		case "--lang":
			runLangSetup()
			return
		case "--colors":
			runColorSetup()
			return
		case "--keys":
			runKeysSetup()
			return
		case "--history", "-hist":
			loadConfigAndSetLang()
			showHistory()
			return
		case "init", "config":
			runConfigSetup()
			return
		default:
			// Non-flag argument: treat as a direct target name
			if !strings.HasPrefix(arg, "-") {
				loadConfigAndSetLang()
				runDirectTarget(arg)
				return
			}
			// Unknown flag: show help and exit
			cfg := loadConfigAndSetLang()
			printHelp(getPalette(cfg.Config.ColorScheme))
			os.Exit(1)
		}
	}

	// No arguments: launch the interactive menu
	cfg := loadConfigAndSetLang()

	// First launch: run the initial setup wizard
	if !cfg.Exists() {
		result, err := config.RunSetup()
		if err != nil {
			fatal(i18n.Get().ErrConfig, err)
		}
		cfg.Config.KeyScheme = result.KeyScheme
		cfg.Config.Language = result.Language
		cfg.Config.ColorScheme = result.ColorScheme
		if result.KeyScheme == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		}
		_ = cfg.Save()
		fmt.Println()
	}

	makefilePath := findMakefile()
	if makefilePath == "" {
		fatal("%s", i18n.Get().ErrNoMakefile)
	}

	m := i18n.Get()
	fmt.Printf("%s%s%s\n\n", ansi.Gray, fmt.Sprintf(m.MakefileFound, makefilePath), ansi.Reset)

	targets, err := parser.ParseMakefile(makefilePath)
	if err != nil {
		fatal(m.ErrReadMakefile, err)
	}

	if len(targets) == 0 {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Red, m.ErrNoTargets, ansi.Reset)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Gray, m.HintAddDoc, ansi.Reset)
		os.Exit(1)
	}

	opts := ui.Options{
		KeyScheme:     cfg.Config.KeyScheme,
		ColorPalette:  getPalette(cfg.Config.ColorScheme),
		CustomUpKey:   cfg.Config.CustomUpKey,
		CustomDownKey: cfg.Config.CustomDownKey,
	}
	result := ui.Run(targets, opts)

	if !result.Confirmed || result.Target == nil {
		fmt.Printf("%s%s%s\n", ansi.Gray, m.Cancelled, ansi.Reset)
		return
	}

	executeTarget(makefilePath, result.Target.Name)
}

// executeTarget runs a make target, records it in history, and prints the outcome.
func executeTarget(makefilePath, targetName string) {
	m := i18n.Get()

	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Green, fmt.Sprintf(m.Executing, makefilePath, targetName), ansi.Reset)

	cmd := exec.Command("make", "-f", makefilePath, targetName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	hist, err := history.New()
	if err == nil {
		_ = hist.Add(targetName)
	}

	if err := cmd.Run(); err != nil {
		fatal("\n"+m.ErrCommandFailed, err)
	}

	fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, m.Success, ansi.Reset)
}

// printHelp renders the help text (always in English) using the given color palette.
func printHelp(palette []string) {
	type helpEntry struct {
		cmd  string
		desc string
	}
	entries := []helpEntry{
		{"mk", "Interactive menu"},
		{"mk <target>", "Run a target directly"},
		{"mk --help, -h", "Show this help"},
		{"mk --version, -v", "Show version"},
		{"mk --config", "Configure language, colors and key scheme"},
		{"mk --lang", "Change language"},
		{"mk --colors", "Change color scheme"},
		{"mk --keys", "Change key scheme"},
		{"mk --history, -hist", "Show execution history"},
	}

	fmt.Printf("\n  %s%s🔧 mk%s %s— interactive Makefile runner%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset, ansi.Gray, ansi.Reset)
	fmt.Printf("  %sUsage:%s mk %s[options]%s %s[target]%s\n\n", ansi.Bold, ansi.Reset, ansi.Gray, ansi.Reset, ansi.Gray, ansi.Reset)

	for i, e := range entries {
		c := palette[i%len(palette)]
		fmt.Printf("    %s%-24s%s  %s%s%s\n", c, e.cmd, ansi.Reset, ansi.Gray, e.desc, ansi.Reset)
	}
	fmt.Println()
}

// loadConfigAndSetLang loads the configuration and activates the saved language.
func loadConfigAndSetLang() *config.Manager {
	cfg, err := config.New()
	if err != nil {
		cfg = &config.Manager{Config: config.Config{
			KeyScheme:   config.KeySchemeArrows,
			Language:    i18n.LangEN,
			ColorScheme: config.ColorSchemeRainbow,
		}}
	}
	i18n.Set(cfg.Config.Language)
	return cfg
}

// withConfig loads the configuration, sets the language, runs fn, and saves.
func withConfig(fn func(cfg *config.Manager)) {
	cfg, err := config.New()
	if err != nil {
		fatal(i18n.Get().ErrGeneric, err)
	}
	i18n.Set(cfg.Config.Language)
	fn(cfg)
	if err := cfg.Save(); err != nil {
		fatal(i18n.Get().ErrSaveConfig, err)
	}
}

func runConfigSetup() {
	withConfig(func(cfg *config.Manager) {
		result, err := config.RunSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.KeyScheme = result.KeyScheme
		cfg.Config.Language = result.Language
		cfg.Config.ColorScheme = result.ColorScheme
		if result.KeyScheme == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		} else {
			cfg.Config.CustomUpKey = 0
			cfg.Config.CustomDownKey = 0
		}
	})
}

func runLangSetup() {
	withConfig(func(cfg *config.Manager) {
		lang, err := config.RunLangSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.Language = lang
	})
}

func runColorSetup() {
	withConfig(func(cfg *config.Manager) {
		cs, err := config.RunColorSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.ColorScheme = cs
	})
}

func runKeysSetup() {
	withConfig(func(cfg *config.Manager) {
		ks, err := config.RunKeysSetup()
		if err != nil {
			fatal(i18n.Get().ErrGeneric, err)
		}
		cfg.Config.KeyScheme = ks
		if ks == config.KeySchemeCustom {
			upKey, downKey := captureCustomKeys()
			cfg.Config.CustomUpKey = upKey
			cfg.Config.CustomDownKey = downKey
		} else {
			cfg.Config.CustomUpKey = 0
			cfg.Config.CustomDownKey = 0
		}
	})
}

// captureCustomKeys prompts the user to assign custom UP and DOWN navigation keys.
func captureCustomKeys() (upKey, downKey byte) {
	m := i18n.Get()
	fmt.Println()

	// Capture the UP key
	fmt.Printf("  %s%s%s", ansi.Purple, m.ConfigKeyUpPrompt, ansi.Reset)
	up, err := ui.CaptureKey()
	if err != nil {
		// Fall back to z/s if raw mode is unavailable
		fmt.Printf("\n%s%s%s\n", ansi.Gray, "  (raw mode unavailable, defaulting to z/s)", ansi.Reset)
		return 'z', 's'
	}
	fmt.Printf("%s%s%s\n", ansi.Bold, ui.KeyDisplayName(up), ansi.Reset)

	// Capture the DOWN key (must differ from UP)
	for {
		fmt.Printf("  %s%s%s", ansi.Purple, m.ConfigKeyDownPrompt, ansi.Reset)
		down, err := ui.CaptureKey()
		if err != nil {
			return up, 's'
		}
		if down == up {
			fmt.Printf("%s(same as up key, try again)%s\n", ansi.Red, ansi.Reset)
			continue
		}
		fmt.Printf("%s%s%s\n", ansi.Bold, ui.KeyDisplayName(down), ansi.Reset)

		upName := ui.KeyDisplayName(up)
		downName := ui.KeyDisplayName(down)
		fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, fmt.Sprintf(m.ConfigKeyConfirmCustom, upName, downName), ansi.Reset)
		return up, down
	}
}

func runDirectTarget(target string) {
	makefilePath := findMakefile()
	if makefilePath == "" {
		fatal("%s", i18n.Get().ErrNoMakefile)
	}

	m := i18n.Get()

	targets, err := parser.ParseMakefile(makefilePath)
	if err != nil {
		fatal(m.ErrReadMakefile, err)
	}

	// Verify the target exists among documented targets
	found := false
	for _, t := range targets {
		if t.Name == target {
			found = true
			break
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Red, fmt.Sprintf(m.ErrUnknownTarget, target), ansi.Reset)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", ansi.Gray, m.AvailableTargets, ansi.Reset)
		for _, t := range targets {
			desc := ""
			if t.Description != "" {
				desc = fmt.Sprintf("  %s%s%s", ansi.Gray, t.Description, ansi.Reset)
			}
			fmt.Fprintf(os.Stderr, "  %s•%s %s%s\n", ansi.Purple, ansi.Reset, t.Name, desc)
		}
		os.Exit(1)
	}

	executeTarget(makefilePath, target)
}

func findMakefile() string {
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	prefixes := []string{"Makefile", "makefile", "GNUmakefile"}
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}
	for _, prefix := range prefixes {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, prefix+".") && filepath.Ext(name) != "" {
				return name
			}
		}
	}

	return ""
}

func showHistory() {
	m := i18n.Get()
	hist, err := history.New()
	if err != nil {
		fatal(m.ErrReadHistory, err)
	}

	entries := hist.Recent(20)
	if len(entries) == 0 {
		fmt.Printf("%s%s%s\n", ansi.Gray, m.HistoryEmpty, ansi.Reset)
		return
	}

	fmt.Printf("%s%s%s%s\n\n", ansi.Bold, ansi.Purple, m.HistoryTitle, ansi.Reset)
	for i, e := range entries {
		age := formatAge(e.ExecutedAt)
		fmt.Printf("  %s%2d.%s  %-30s  %s%s  %s%s\n",
			ansi.Purple, i+1, ansi.Reset,
			fmt.Sprintf("%s%s%s", ansi.Bold, e.Target, ansi.Reset),
			ansi.Gray, age,
			e.Directory, ansi.Reset,
		)
	}
}

func formatAge(t time.Time) string {
	m := i18n.Get()
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return m.TimeJustNow
	case d < time.Hour:
		return fmt.Sprintf(m.TimeMinutesAgo, int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf(m.TimeHoursAgo, int(d.Hours()))
	default:
		return t.Format("02/01 15:04")
	}
}
