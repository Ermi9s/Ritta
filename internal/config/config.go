package config

type Config struct {
	LocalProjectRoot  string      `yaml:"local_project_root"`
	RemoteProjectRoot string      `yaml:"remote_project_root"`
	Source            Source      `yaml:"source"`
	Server            Server      `yaml:"server"`
	SetupConfig       SetupConfig `yaml:"setup_config"`
	ScanEnv           bool        `yaml:"scan_env"`
	File              []File      `yaml:"file"`
	Build             *Command    `yaml:"build"`
	Run               *Command    `yaml:"run"`
	Health            *Health     `yaml:"health"`
	Proxy             *Proxy      `yaml:"proxy"`
	Domains           []Domain    `yaml:"domains"`
	TLS               *TLS        `yaml:"tls"`
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

type File struct {
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
	Script string `yaml:"script"`
}
