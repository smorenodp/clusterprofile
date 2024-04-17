package providers

import (
	"fmt"
	"regexp"
	"time"

	"github.com/smorenodp/clusterprofile/config"
)

const (
	nomadRoleDataKey = "secret_id"
	nomadEnvTokenVar = "NOMAD_TOKEN"
	nomadEnvAddrVar  = "NOMAD_ADDR"
	nomadEnvTTLVar   = "NOMAD_TTL"
)

type NomadProvider struct {
	client *VaultClient
	config config.ProviderConfig
	token  string
	TTL    time.Time
}

func NewNomadProvider(client *VaultClient, config config.ProviderConfig) *NomadProvider {
	return &NomadProvider{client: client, config: config}
}

func (p *NomadProvider) LoadProfileCreds(info []string) {
	var token string
	var ttl time.Time
	tokenRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, nomadEnvTokenVar))
	ttlRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, nomadEnvTTLVar))
	for _, i := range info {
		if matches := tokenRegex.FindStringSubmatch(i); matches != nil {
			token = matches[1]
		} else if matches := ttlRegex.FindStringSubmatch(i); matches != nil {
			ttl, _ = time.Parse(time.Layout, matches[1])
		}
	}
	if time.Now().Before(ttl) {
		p.token = token
		p.TTL = ttl
	}

}

func (p *NomadProvider) credsFromRole() (string, error) {
	if p.token != "" {
		return p.token, nil
	}
	path := fmt.Sprintf("%s/creds/%s", p.config.Backend, p.config.Config.Role)
	secret, err := p.client.Logical().Read(path)
	if err != nil {
		return "", err
	}
	token := secret.Data[nomadRoleDataKey].(string)
	p.token = token
	duration := secret.LeaseDuration
	TTL, _ := time.ParseDuration(fmt.Sprintf("%ds", duration))

	p.TTL = time.Now().Add(TTL)
	return token, nil
}

func (p *NomadProvider) credsFromToken() (string, error) {
	p.token = p.config.Config.Token
	return p.token, nil
}

func (p *NomadProvider) GenerateCreds() (string, error) {
	switch p.config.Method {
	case "role":
		return p.credsFromRole()
	case "token":
		return p.credsFromToken()
	default:
		return "", fmt.Errorf("method %s not implemented yet", p.config.Method)
	}
}

func (p *NomadProvider) ExportCreds() []string {
	return []string{fmt.Sprintf("export %s=%q", nomadEnvTokenVar, p.token),
		fmt.Sprintf("export %s=%q", nomadEnvTTLVar, p.TTL.Format(layout)),
		fmt.Sprintf("export %s=%q", nomadEnvAddrVar, p.config.Addr)}
}

func (p *NomadProvider) CredsLoaded() bool {
	return p.token != ""
}

func (p *NomadProvider) ProfileCreds() []string {
	return []string{fmt.Sprintf("%s=%q", nomadEnvTokenVar, p.token),
		fmt.Sprintf("%s=%q", nomadEnvTTLVar, p.TTL.Format(layout)),
		fmt.Sprintf("%s=%q", nomadEnvAddrVar, p.config.Addr)}
}
