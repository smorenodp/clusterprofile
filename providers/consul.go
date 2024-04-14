package providers

import (
	"fmt"
	"regexp"
	"time"

	"github.com/smorenodp/clusterprofile/config"
)

const (
	consulRoleTokenKey = "token"
	consulEnvTokenVar  = "CONSUL_HTTP_TOKEN"
	consulEnvTTLVar    = "CONSUL_TTL"
	consulEnvAddrVar   = "CONSUL_HTTP_ADDR"
)

type ConsulProvider struct {
	vault  *VaultClient
	config config.ProviderConfig
	token  string
	TTL    time.Time
}

func NewConsulProvider(vault *VaultClient, config config.ProviderConfig) *ConsulProvider {
	return &ConsulProvider{vault: vault, config: config}
}

func (p *ConsulProvider) LoadProfileCreds(info []string) {
	var token string
	var ttl time.Time
	tokenRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, consulEnvTokenVar))
	ttlRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, consulEnvTTLVar))
	for _, i := range info {
		if matches := tokenRegex.FindStringSubmatch(i); matches != nil {
			token = matches[1]
		} else if matches := ttlRegex.FindStringSubmatch(i); matches != nil {
			ttl, _ = time.Parse(layout, matches[1])
		}
	}
	if time.Now().Before(ttl) {
		p.token = token
		p.TTL = ttl
	}

}

func (p *ConsulProvider) credsFromRole() (string, error) {
	path := fmt.Sprintf("%s/creds/%s", p.config.Backend, p.config.Config.Role)
	secret, err := p.vault.Logical().Read(path)
	if err != nil {
		return "", err
	}

	token := secret.Data[consulRoleTokenKey].(string)
	duration := secret.LeaseDuration
	TTL, _ := time.ParseDuration(fmt.Sprintf("%ds", duration))
	p.token = token

	p.TTL = time.Now().Add(TTL)
	return token, nil
}

func (p *ConsulProvider) credsFromToken() (string, error) {
	p.token = p.config.Config.Token
	return p.token, nil
}

func (p *ConsulProvider) GenerateCreds() (string, error) {
	switch p.config.Method {
	case "role":
		return p.credsFromRole()
	case "token":
		return p.credsFromToken()
	default:
		return "", fmt.Errorf("method %s not implemented yet", p.config.Method)
	}
}

func (p *ConsulProvider) ExportCreds() []string {
	return []string{fmt.Sprintf("export %s=\"%s\"", consulEnvTokenVar, p.token),
		fmt.Sprintf("export %s=\"%s\"", consulEnvTTLVar, p.TTL.Format(layout)),
		fmt.Sprintf("export %s=\"%s\"", consulEnvAddrVar, p.config.Addr)}
}

func (p *ConsulProvider) CredsLoaded() bool {
	return p.token != ""
}

func (p *ConsulProvider) ProfileCreds() []string {
	return []string{fmt.Sprintf("%s=\"%s\"", consulEnvTokenVar, p.token),
		fmt.Sprintf("%s=\"%s\"", consulEnvTTLVar, p.TTL.Format(layout)),
		fmt.Sprintf("%s=\"%s\"", consulEnvAddrVar, p.config.Addr)}
}
