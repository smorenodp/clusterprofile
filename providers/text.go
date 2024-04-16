package providers

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/smorenodp/clusterprofile/config"
)

const (
	exportRegex = "(?P<data>[^=]*)[\\ ]*=[\\ ]*(?P<data>.*)"
)

type TextProvider struct {
	client     *VaultClient
	config     config.ProviderConfig
	load       bool
	mapEnvVars map[string]string
	regex      *regexp.Regexp
}

func NewTextProvider(client *VaultClient, config config.ProviderConfig) *TextProvider {
	p := TextProvider{client: client, config: config, regex: regexp.MustCompile(exportRegex), mapEnvVars: make(map[string]string)}
	return &p
}

func (p *TextProvider) LoadProfileCreds(info []string) {
	// Not needed
}

func (p *TextProvider) generateFromFile() (string, error) {
	f, err := os.Open(p.config.Config.File)
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(f)

	for line, _, _ := r.ReadLine(); line != nil; line, _, _ = r.ReadLine() {
		stringLine := string(line)
		if matches := p.regex.FindStringSubmatch(stringLine); matches != nil {
			name := matches[1] // The name is always this index
			value := matches[2]
			p.mapEnvVars[name] = value
		}
	}
	return "", nil
}

func (p *TextProvider) generateFromData() (string, error) {
	r := bufio.NewReader(strings.NewReader(p.config.Config.Data))
	for line, _, _ := r.ReadLine(); line != nil; line, _, _ = r.ReadLine() {
		stringLine := string(line)
		if matches := p.regex.FindStringSubmatch(stringLine); matches != nil {
			name := matches[1] // The name is always this index
			value := matches[2]
			p.mapEnvVars[name] = value
		}
	}
	return "", nil
}

func (p *TextProvider) GenerateCreds() (string, error) {
	switch p.config.Method {
	case "file":
		return p.generateFromFile()
	case "data":
		return p.generateFromData()
	default:
		return "", fmt.Errorf("method %s not implemented yet", p.config.Method)
	}
}

func (p *TextProvider) ExportCreds() (export []string) {
	for envName, envValue := range p.mapEnvVars {
		export = append(export, fmt.Sprintf("export %s=%s", envName, envValue))
	}
	return
}

func (p *TextProvider) CredsLoaded() bool {
	return p.load
}

func (p *TextProvider) ProfileCreds() (creds []string) {
	for envName, envValue := range p.mapEnvVars {
		creds = append(creds, fmt.Sprintf("export %s=%s", envName, envValue))
	}
	return
}
