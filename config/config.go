package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const (
	yamlRegex = ".*\\.yaml"
)

type InnerProviderConfig struct {
	Role       string            `yaml:"role"`
	Token      string            `yaml:"token"`
	Policies   []string          `yaml:"policies"`
	SecretPath string            `yaml:"path"`
	SecretMap  map[string]string `yaml:"secret_map"`
}

type ProviderConfig struct {
	Type    string              `yaml:"type"`
	Backend string              `yaml:"backend"`
	Method  string              `yaml:"method"`
	Config  InnerProviderConfig `yaml:"config"`
	Addr    string              `yaml:"addr"`
}

type VaultConfig struct {
	Addr         string              `yaml:"addr"`
	Method       string              `yaml:"method"`
	Config       InnerProviderConfig `yaml:"config"`
	PivotProfile string              `yaml:"pivoting_profile"`
}

type ClusterConfig struct {
	Name      string           `yaml:"name"`
	Vault     VaultConfig      `yaml:"vault"`
	Providers []ProviderConfig `yaml:"providers"`
}

func ReadConfig(folder string) (config map[string]ClusterConfig, err error) {
	fileRegex := regexp.MustCompile(yamlRegex)
	var content []byte
	var profileConfig []ClusterConfig
	config = make(map[string]ClusterConfig)
	profiles, _ := os.ReadDir(folder)
	for _, f := range profiles {
		if fileRegex.MatchString(f.Name()) {
			content, err = os.ReadFile(fmt.Sprintf("%s/%s", folder, f.Name()))
			if err != nil {
				return
			}
			err = yaml.Unmarshal(content, &profileConfig)
			if err != nil {
				return
			}
			for _, p := range profileConfig {
				config[p.Name] = p
			}
		}
	}
	return
}
