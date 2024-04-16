package providers

import (
	"fmt"
	"regexp"

	"github.com/smorenodp/clusterprofile/config"
)

type InfobloxEnvVar struct {
	regex *regexp.Regexp
	value string
}

type InfobloxProvider struct {
	client     *VaultClient
	config     config.ProviderConfig
	mapEnvVars map[string]InfobloxEnvVar
	load       bool
}

func (p *InfobloxProvider) generateMap() {
	mapEnvVars := map[string]InfobloxEnvVar{}
	for _, envValue := range p.config.Config.SecretMap {
		mapEnvVars[envValue] = InfobloxEnvVar{regex: regexp.MustCompile(fmt.Sprintf(dataRegex, envValue))}
	}
	p.mapEnvVars = mapEnvVars
}

func NewInfobloxProvider(client *VaultClient, config config.ProviderConfig) *InfobloxProvider {
	p := InfobloxProvider{client: client, config: config}
	p.generateMap()
	return &p
}

func (p *InfobloxProvider) LoadProfileCreds(info []string) {
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

func (p *InfobloxProvider) credsFromSecret() (string, error) {
	secret, err := p.client.Logical().Read(p.config.Config.SecretPath)
	if err != nil {
		return "", err
	}

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
	return "", nil
}

func (p *InfobloxProvider) GenerateCreds() (string, error) {
	switch p.config.Method {
	case "secret":
		return p.credsFromSecret()
	default:
		return "", fmt.Errorf("method %s not implemented yet", p.config.Method)
	}
}

func (p *InfobloxProvider) ExportCreds() (export []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			export = append(export, fmt.Sprintf("export %s=%s", envName, envValue.value))
		}

	}
	return
}

func (p *InfobloxProvider) CredsLoaded() bool {
	return p.load
}

func (p *InfobloxProvider) ProfileCreds() (creds []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			creds = append(creds, fmt.Sprintf("%s=%s", envName, envValue.value))
		}
	}
	return
}
