package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Tutor    TutorConfig   `toml:"tutor"`
	Agent    AgentConfig   `toml:"agent"`
	Watchers WatcherConfig `toml:"watchers"`
	Tmux     TmuxConfig    `toml:"tmux"`
}

type TutorConfig struct {
	Intensity string `toml:"intensity"`
	Level     string `toml:"level"`
}

type AgentConfig struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

type WatcherConfig struct {
	FilePatterns         []string `toml:"file_patterns"`
	IgnorePatterns       []string `toml:"ignore_patterns"`
	TerminalPollInterval string   `toml:"terminal_poll_interval"`
	GitPollInterval      string   `toml:"git_poll_interval"`
}

type TmuxConfig struct {
	Layout       string `toml:"layout"`
	UserPaneSize int    `toml:"user_pane_size"`
}

func Default() *Config {
	return &Config{
		Tutor: TutorConfig{
			Intensity: "on-demand",
			Level:     "auto",
		},
		Agent: AgentConfig{
			Command: "claude",
			Args:    []string{},
		},
		Watchers: WatcherConfig{
			FilePatterns:         []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts", "**/*.rs"},
			IgnorePatterns:       []string{"node_modules", ".git", "vendor", "target"},
			TerminalPollInterval: "2s",
			GitPollInterval:      "5s",
		},
		Tmux: TmuxConfig{
			Layout:       "horizontal",
			UserPaneSize: 50,
		},
	}
}

func configPath(projectDir string) string {
	return filepath.Join(projectDir, ".agent-tutor", "config.toml")
}

func Load(projectDir string) (*Config, error) {
	path := configPath(projectDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := Save(projectDir, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(projectDir string, cfg *Config) error {
	path := configPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
