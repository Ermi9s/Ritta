package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadEnvFile reads a KEY=VALUE environment file and stores variables in envMap.
func loadEnvFile(path string, envMap map[string]string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if _, exists := envMap[k]; !exists {
				envMap[k] = v
			}
		}
	}
}

func getEnv(key string, envMap map[string]string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return envMap[key]
}

func LoadConfig(path string) (*Config, error) {
	targetFile := path
	dir := filepath.Dir(path)
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		targetFile = filepath.Join(path, filename)
		dir = path
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, err
	}

	envMap := make(map[string]string)
	loadEnvFile(filepath.Join(dir, ".env.deploy"), envMap)
	loadEnvFile(filepath.Join(dir, "deploy", ".env.deploy"), envMap)
	loadEnvFile(filepath.Join(dir, ".env"), envMap)

	expanded := os.Expand(string(data), func(key string) string {
		parts := strings.SplitN(key, ":-", 2)
		val := getEnv(parts[0], envMap)
		if val == "" && len(parts) == 2 {
			return parts[1]
		}
		return val
	})

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	if cfg.DeployDir == "" {
		cfg.DeployDir = "./deploy"
	}

	// Smart fallbacks for server credentials
	if cfg.Server.Host == "" {
		if h := getEnv("VPS_HOST", envMap); h != "" {
			cfg.Server.Host = h
		} else {
			cfg.Server.Host = getEnv("SERVER_HOST", envMap)
		}
	}
	if cfg.Server.User == "" {
		if u := getEnv("VPS_USER", envMap); u != "" {
			cfg.Server.User = u
		} else if u := getEnv("SERVER_USER", envMap); u != "" {
			cfg.Server.User = u
		} else {
			cfg.Server.User = "root"
		}
	}
	if cfg.Server.Port == 0 {
		portStr := getEnv("VPS_PORT", envMap)
		if portStr == "" {
			portStr = getEnv("SERVER_PORT", envMap)
		}
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.Server.Port = p
		} else {
			cfg.Server.Port = 22
		}
	}
	if cfg.Server.Key == "" {
		if k := getEnv("VPS_KEY", envMap); k != "" {
			cfg.Server.Key = k
		} else if k := getEnv("SERVER_SSH_KEY_PATH", envMap); k != "" {
			cfg.Server.Key = k
		}
	}

	return &cfg, nil
}
