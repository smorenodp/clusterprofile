package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type InnerProviderConfig struct {
	Role     string   `yaml:"role"`
	Token    string   `yaml:"token"`
	Policies []string `yaml:"policies"`
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
	var content []byte
	var profileConfig []ClusterConfig
	config = make(map[string]ClusterConfig)
	profiles, _ := os.ReadDir(folder)
	for _, f := range profiles {
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
	return
}
