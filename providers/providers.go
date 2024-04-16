package providers

import (
	"github.com/smorenodp/clusterprofile/config"
)

const (
	layout    = "2006-01-02 15:04:05"
	dataRegex = "%s=\"(?P<data>.*)\""
)

type Provider interface {
	GenerateCreds() (string, error)
	ExportCreds() []string
	LoadProfileCreds([]string)
	ProfileCreds() []string
	CredsLoaded() bool
}

func NewProvider(client *VaultClient, config config.ProviderConfig) Provider {
	switch config.Type {
	case "consul":
		return NewConsulProvider(client, config)
	case "nomad":
		return NewNomadProvider(client, config)
	case "secret":
		return NewSecretProvider(client, config)
	case "text":
		return NewTextProvider(client, config)
	default:
		return nil
	}
}
