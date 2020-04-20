package config

// RuleSource object defining the source location of the files
type RuleSource struct {
	URL   string `yaml:"url"`
	Files string `yaml:"files"`
	Name  string
}

// Config of the cli
type Config struct {
	Rules struct {
		Presets     []string     `yaml:"presets"`
		Additionals []RuleSource `yaml:"additionals"`
	} `yaml:"rules"`
	TargetVersion string `yaml:"targetVersion"`
	Helm          struct {
		Repositories []struct {
			URL     string   `yaml:"url"`
			Chart   string   `yaml:"chart"`
			Version string   `yaml:"version"`
			Values  []string `yaml:"values"`
		} `yaml:"repositories"`
		Registries []struct {
			URL     string   `yaml:"url"`
			Version string   `yaml:"version"`
			Values  []string `yaml:"values"`
		} `yaml:"repositories"`
	} `yaml:"helm"`
	Files  []string `yaml:"files"`
	Ignore []string `yaml:"ignore"`
}

var (
	// Conf is the current loaded Configuration
	Conf *Config
)
