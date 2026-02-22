package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/subut0n/mk/internal/ansi"
	"github.com/subut0n/mk/internal/i18n"
)

// KeyScheme defines the keyboard navigation scheme.
type KeyScheme string

const (
	KeySchemeArrows KeyScheme = "arrows"
	KeySchemeWASD   KeyScheme = "wasd"
	KeySchemeCustom KeyScheme = "custom"
)

// ColorScheme defines a color palette for the UI.
type ColorScheme string

const (
	ColorSchemeRainbow      ColorScheme = "rainbow"
	ColorSchemeDeuteranopia ColorScheme = "deuteranopia"
	ColorSchemeTritanopia   ColorScheme = "tritanopia"
	ColorSchemeHighContrast  ColorScheme = "high-contrast"
)

// Config holds the user configuration.
type Config struct {
	KeyScheme     KeyScheme   `json:"key_scheme"`
	Language      i18n.Lang   `json:"language"`
	ColorScheme   ColorScheme `json:"color_scheme"`
	CustomUpKey   byte        `json:"custom_up_key,omitempty"`
	CustomDownKey byte        `json:"custom_down_key,omitempty"`
}

// Manager handles persistent configuration.
type Manager struct {
	filePath string
	Config   Config
}

// New creates a new configuration Manager.
func New() (*Manager, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}

	configDir := filepath.Join(dir, "mk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		filePath: filepath.Join(configDir, "config.json"),
		Config: Config{
			KeyScheme:   KeySchemeArrows,
			Language:    i18n.LangEN,
			ColorScheme: ColorSchemeRainbow,
		},
	}

	_ = m.load()

	return m, nil
}

// Exists reports whether the configuration file exists on disk.
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.filePath)
	return err == nil
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m.Config); err != nil {
		return err
	}
	// Default missing language to EN.
	if m.Config.Language == "" {
		m.Config.Language = i18n.LangEN
	}
	// Default missing color scheme to rainbow.
	if m.Config.ColorScheme == "" {
		m.Config.ColorScheme = ColorSchemeRainbow
	}
	// Migrate legacy "zqsd" key scheme to "custom" with z/s keys.
	if m.Config.KeyScheme == "zqsd" {
		m.Config.KeyScheme = KeySchemeCustom
		m.Config.CustomUpKey = 'z'
		m.Config.CustomDownKey = 's'
	}
	return nil
}

// Save writes the configuration to disk.
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.Config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0600)
}

// SetupResult holds the result of the interactive setup.
type SetupResult struct {
	KeyScheme     KeyScheme
	Language      i18n.Lang
	ColorScheme   ColorScheme
	CustomUpKey   byte
	CustomDownKey byte
}

// RunSetup runs the full interactive setup (language, colors, keys).
func RunSetup() (SetupResult, error) {
	result := SetupResult{}

	fmt.Printf("%s%s⚙  mk — setup%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset)

	lang, err := promptLang()
	if err != nil {
		return result, err
	}
	result.Language = lang

	cs, err := promptColor()
	if err != nil {
		return result, err
	}
	result.ColorScheme = cs

	ks, err := promptKeys()
	if err != nil {
		return result, err
	}
	result.KeyScheme = ks

	return result, nil
}

// RunLangSetup prompts the user to change the language.
func RunLangSetup() (i18n.Lang, error) {
	fmt.Printf("%s%s⚙  mk — language%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset)
	return promptLang()
}

// RunColorSetup prompts the user to change the color scheme.
func RunColorSetup() (ColorScheme, error) {
	fmt.Printf("%s%s⚙  mk — colors%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset)
	return promptColor()
}

// RunKeysSetup prompts the user to change the key scheme.
func RunKeysSetup() (KeyScheme, error) {
	fmt.Printf("%s%s⚙  mk — keys%s\n\n", ansi.Bold, ansi.Purple, ansi.Reset)
	return promptKeys()
}

func promptLang() (i18n.Lang, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("  Choose your language / Choisis ta langue:\n\n")
	fmt.Printf("  %s1.%s English\n", ansi.Purple, ansi.Reset)
	fmt.Printf("  %s2.%s Français\n", ansi.Purple, ansi.Reset)
	fmt.Printf("  %s3.%s Español\n", ansi.Purple, ansi.Reset)
	fmt.Printf("  %s4.%s Deutsch\n", ansi.Purple, ansi.Reset)
	fmt.Printf("\n%sChoice [1/2/3/4] (default: 1): %s", ansi.Gray, ansi.Reset)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var lang i18n.Lang
		switch input {
		case "", "1":
			lang = i18n.LangEN
		case "2":
			lang = i18n.LangFR
		case "3":
			lang = i18n.LangES
		case "4":
			lang = i18n.LangDE
		default:
			fmt.Printf("%s1, 2, 3 or/ou 4: %s", ansi.Gray, ansi.Reset)
			continue
		}

		i18n.Set(lang)
		m := i18n.Get()
		fmt.Printf("\n%s%s%s%s\n\n", ansi.Bold, ansi.Green, m.ConfigLangConfirm, ansi.Reset)
		return lang, nil
	}
}

func promptColor() (ColorScheme, error) {
	reader := bufio.NewReader(os.Stdin)
	m := i18n.Get()

	fmt.Printf("  %s\n\n", m.ConfigColorPrompt)
	fmt.Printf("  %s1.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigColorRainbow, ansi.Gray, m.ConfigColorRainbowHint, ansi.Reset)
	fmt.Printf("  %s2.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigColorDeuteranopia, ansi.Gray, m.ConfigColorDeuteranopiaHint, ansi.Reset)
	fmt.Printf("  %s3.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigColorTritanopia, ansi.Gray, m.ConfigColorTritanopiaHint, ansi.Reset)
	fmt.Printf("  %s4.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigColorHighContrast, ansi.Gray, m.ConfigColorHighContrastHint, ansi.Reset)
	fmt.Printf("\n%s%s%s", ansi.Gray, m.ConfigColorChoice, ansi.Reset)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var cs ColorScheme
		var label string
		switch input {
		case "", "1":
			cs = ColorSchemeRainbow
			label = m.ConfigColorRainbow
		case "2":
			cs = ColorSchemeDeuteranopia
			label = m.ConfigColorDeuteranopia
		case "3":
			cs = ColorSchemeTritanopia
			label = m.ConfigColorTritanopia
		case "4":
			cs = ColorSchemeHighContrast
			label = m.ConfigColorHighContrast
		default:
			fmt.Printf("%s%s%s", ansi.Gray, m.ConfigColorInvalid, ansi.Reset)
			continue
		}
		fmt.Printf("\n%s%s%s%s\n\n", ansi.Bold, ansi.Green, fmt.Sprintf(m.ConfigColorConfirm, label), ansi.Reset)
		return cs, nil
	}
}

func promptKeys() (KeyScheme, error) {
	reader := bufio.NewReader(os.Stdin)
	m := i18n.Get()

	fmt.Printf("  %s\n\n", m.ConfigKeyPrompt)
	fmt.Printf("  %s1.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigKeyArrows, ansi.Gray, m.ConfigKeyArrowsHint, ansi.Reset)
	fmt.Printf("  %s2.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigKeyWASD, ansi.Gray, m.ConfigKeyWASDHint, ansi.Reset)
	fmt.Printf("  %s3.%s %s %s%s%s\n", ansi.Purple, ansi.Reset, m.ConfigKeyCustom, ansi.Gray, m.ConfigKeyCustomHint, ansi.Reset)
	fmt.Printf("\n%s%s%s", ansi.Gray, m.ConfigKeyChoice, ansi.Reset)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "", "1":
			fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, m.ConfigKeyConfirmArrows, ansi.Reset)
			return KeySchemeArrows, nil
		case "2":
			fmt.Printf("\n%s%s%s%s\n", ansi.Bold, ansi.Green, m.ConfigKeyConfirmWASD, ansi.Reset)
			return KeySchemeWASD, nil
		case "3":
			return KeySchemeCustom, nil
		default:
			fmt.Printf("%s%s%s", ansi.Gray, m.ConfigKeyInvalid, ansi.Reset)
		}
	}
}
