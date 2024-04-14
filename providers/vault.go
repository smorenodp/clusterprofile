package providers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/smorenodp/clusterprofile/config"
)

const (
	vaultEnvTokenVar = "VAULT_TOKEN"
	vaultEnvTTLVar   = "VAULT_TTL"
	vaultEnvAddrVar  = "VAULT_ADDR"
)

type VaultClient struct {
	config config.VaultConfig
	TTL    time.Time
	*vault.Client
}

func (c *VaultClient) LoadProfileCreds(info []string) {
	var token string
	var ttl time.Time
	tokenRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, vaultEnvTokenVar))
	ttlRegex := regexp.MustCompile(fmt.Sprintf(dataRegex, vaultEnvTTLVar))
	for _, i := range info {
		if matches := tokenRegex.FindStringSubmatch(i); matches != nil {
			token = matches[1]
		} else if matches := ttlRegex.FindStringSubmatch(i); matches != nil {
			ttl, _ = time.Parse(layout, matches[1])
		}
	}
	if time.Now().Before(ttl) {
		c.SetToken(token)
		c.TTL = ttl
	}

}

func NewVaultClient(config config.VaultConfig) (*VaultClient, error) {
	defaultConfig := vault.DefaultConfig()
	if config.Addr != "" {
		defaultConfig.Address = config.Addr
	}
	client, err := vault.NewClient(defaultConfig)
	client.SetToken("")
	if err != nil {
		return nil, err
	}
	c := &VaultClient{config: config, Client: client}
	return c, nil
}

func (c *VaultClient) loginOidc() error {
	os.Setenv("VAULT_ADDR", c.config.Addr)
	cmd := exec.Command("vault", "login", "-method", "oidc", "-token-only")
	var outb bytes.Buffer
	cmd.Stdout = &outb
	err := cmd.Run()
	if err != nil {
		return err
	}
	token := outb.String()
	c.SetToken(token)
	secret, err := c.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}
	duration, _ := secret.TokenTTL()
	c.TTL = time.Now().Add(duration)
	return nil
}

func (c *VaultClient) WithPivotRole(pivotConfig config.VaultConfig, profile []string) *VaultClient {
	pivotC := &VaultClient{config: pivotConfig, Client: c.Client}
	pivotC.LoadProfileCreds(profile)
	pivotC.GenerateCreds()
	pivotC.config = c.config
	return pivotC
}

func (c *VaultClient) loginToken() error {
	if c.config.Config.Role != "" {
		tokenSecret, err := c.Auth().Token().CreateWithRole(&vault.TokenCreateRequest{Policies: c.config.Config.Policies}, c.config.Config.Role)
		if err == nil {
			c.SetToken(tokenSecret.Auth.ClientToken)
			dur, _ := time.ParseDuration(fmt.Sprintf("%ds", tokenSecret.Auth.LeaseDuration))
			c.TTL = time.Now().Add(dur)
		} else {
			return err
		}
	} else if c.config.Config.Token != "" {
		c.SetToken(c.config.Config.Token)
	}
	return nil
}

func (c *VaultClient) GenerateCreds() (string, error) {
	var err error
	if c.config.Method == "oidc" && c.Token() == "" {
		err = c.loginOidc()
	} else if c.config.Method == "token" {
		err = c.loginToken()
	}
	return c.Token(), err
}

func (c *VaultClient) ExportCreds() []string {
	return []string{fmt.Sprintf("export %s=\"%s\"", vaultEnvTokenVar, c.Token()),
		fmt.Sprintf("export %s=\"%s\"", vaultEnvTTLVar, c.TTL.Format(layout)),
		fmt.Sprintf("export %s=\"%s\"", vaultEnvAddrVar, c.config.Addr)}
}

func (c *VaultClient) CredsLoaded() bool {
	return c.Token() != ""
}

func (c *VaultClient) ProfileCreds() []string {
	return []string{fmt.Sprintf("%s=\"%s\"", vaultEnvTokenVar, c.Token()),
		fmt.Sprintf("%s=\"%s\"", vaultEnvTTLVar, c.TTL.Format(layout)),
		fmt.Sprintf("%s=\"%s\"", vaultEnvAddrVar, c.config.Addr)}
}
