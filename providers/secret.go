package providers

import (
	"fmt"
	"regexp"

	"github.com/smorenodp/clusterprofile/config"
)

type SecretEnvVar struct {
	regex *regexp.Regexp
	value string
}

type SecretProvider struct {
	client     *VaultClient
	config     config.ProviderConfig
	mapEnvVars map[string]SecretEnvVar
	load       bool
}

func (p *SecretProvider) generateMap() {
	mapEnvVars := map[string]SecretEnvVar{}
	for _, envValue := range p.config.Config.SecretMap {
		mapEnvVars[envValue] = SecretEnvVar{regex: regexp.MustCompile(fmt.Sprintf(dataRegex, envValue))}
	}
	p.mapEnvVars = mapEnvVars
}

func NewSecretProvider(client *VaultClient, config config.ProviderConfig) *SecretProvider {
	p := SecretProvider{client: client, config: config}
	p.generateMap()
	return &p
}

func (p *SecretProvider) LoadProfileCreds(info []string) {
	envNames := []string{}
	for key, _ := range p.mapEnvVars {
		envNames = append(envNames, key)
	}
	for _, i := range info {
		for envIndex, envName := range envNames {
			envVar := p.mapEnvVars[envName]
			if matches := envVar.regex.FindStringSubmatch(i); matches != nil {
				envVar.value = matches[1]
				p.mapEnvVars[envName] = envVar
				envNames = remove(envNames, envIndex)
			}
		}
	}
	p.load = (len(envNames) == 0)
}

func (p *SecretProvider) GenerateCreds() (string, error) {
	secret, err := p.client.Logical().Read(p.config.Config.SecretPath)
	if err != nil {
		return "", err
	}
	if secret != nil {
		envVars2Load := len(p.mapEnvVars)
		for key, value := range secret.Data {
			if envName, ok := p.config.Config.SecretMap[key]; ok {
				envVar := p.mapEnvVars[envName]
				envVar.value = value.(string)
				p.mapEnvVars[envName] = envVar
				envVars2Load--
			}
		}
		p.load = (envVars2Load == 0)
	}
	return "", nil
}

func (p *SecretProvider) ExportCreds() (export []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			export = append(export, fmt.Sprintf("export %s=%s", envName, envValue.value))
		}

	}
	return
}

func (p *SecretProvider) CredsLoaded() bool {
	return p.load
}

func (p *SecretProvider) ProfileCreds() (creds []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			creds = append(creds, fmt.Sprintf("%s=%s", envName, envValue.value))
		}
	}
	return
}
