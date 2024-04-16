package providers

import (
	"fmt"
	"regexp"

	"github.com/smorenodp/clusterprofile/config"
)

type OpenstackEnvVar struct {
	regex *regexp.Regexp
	value string
}

type OpenstackProvider struct {
	client     *VaultClient
	config     config.ProviderConfig
	mapEnvVars map[string]OpenstackEnvVar
	load       bool
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (p *OpenstackProvider) generateMap() {
	mapEnvVars := map[string]OpenstackEnvVar{}
	for _, envValue := range p.config.Config.SecretMap {
		mapEnvVars[envValue] = OpenstackEnvVar{regex: regexp.MustCompile(fmt.Sprintf(dataRegex, envValue))}
	}
	p.mapEnvVars = mapEnvVars
}

func NewOSProvider(client *VaultClient, config config.ProviderConfig) *OpenstackProvider {
	p := OpenstackProvider{client: client, config: config}
	p.generateMap()
	return &p
}

func (p *OpenstackProvider) LoadProfileCreds(info []string) {
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

func (p *OpenstackProvider) credsFromSecret() (string, error) {
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

func (p *OpenstackProvider) GenerateCreds() (string, error) {
	switch p.config.Method {
	case "secret":
		return p.credsFromSecret()
	default:
		return "", fmt.Errorf("method %s not implemented yet", p.config.Method)
	}
}

func (p *OpenstackProvider) ExportCreds() (export []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			export = append(export, fmt.Sprintf("export %s=%s", envName, envValue.value))
		}

	}
	return
}

func (p *OpenstackProvider) CredsLoaded() bool {
	return p.load
}

func (p *OpenstackProvider) ProfileCreds() (creds []string) {
	for envName, envValue := range p.mapEnvVars {
		if envValue.value != "" {
			creds = append(creds, fmt.Sprintf("%s=%s", envName, envValue.value))
		}
	}
	return
}
