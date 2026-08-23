package config

type Config struct {
	Version       int        `yaml:"version"`
	RootDirectory string     `yaml:"root_directory"`
	SetupConfig   SetupConfig `yaml:"setup_config"`
	Source        Source     `yaml:"source"`
	Server        Server     `yaml:"server"`
	Health        *Health    `yaml:"health"`
	Env           EnvConfig  `yaml:"env"`
	Build         *Command   `yaml:"build"`
	Run           *Command   `yaml:"run"`
	Domains       []Domain   `yaml:"domains"`
	Proxy         *Proxy     `yaml:"proxy"`
	TLS           *TLS       `yaml:"tls"`
}

type Source struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

type Server struct {
	Host string `yaml:"host"`
	User string `yaml:"user"`
	Port int    `yaml:"port"`
	Key  string `yaml:"key"`
}

type EnvConfig struct {
	Scan  bool      `yaml:"scan"`
	Files []EnvFile `yaml:"files"`
}

type EnvFile struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Command struct {
	Command string `yaml:"command"`
}

type Domain struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  bool   `yaml:"tls"`
}

type Proxy struct {
	Provider string `yaml:"provider"`
}

type TLS struct {
	Provider string `yaml:"provider"`
	Email    string `yaml:"email"`
}

type Health struct {
	Command string `yaml:"command"`
}

type SetupConfig struct {
	PackageManager string `yaml:"package_manager"`
	Script         string `yaml:"script"`
}