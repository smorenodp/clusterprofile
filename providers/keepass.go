package providers

import (
	"fmt"
	"os"
	"strings"

	"github.com/smorenodp/clusterprofile/config"
	"github.com/tobischo/gokeepasslib/v3"
)

type KeepassProvider struct {
	vault  *VaultClient
	config config.ProviderConfig
	data   map[string]string
	db     *gokeepasslib.Database
}

func NewKeePassProvider(vault *VaultClient, config config.ProviderConfig) *KeepassProvider {
	file, err := os.Open(config.Config.File)
	if err != nil {
		return nil
	}
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(config.Config.Password)
	err = gokeepasslib.NewDecoder(file).Decode(db)
	if err != nil {
		return nil
	}
	provider := &KeepassProvider{vault: vault, config: config, data: make(map[string]string), db: db}
	return provider
}

func (k *KeepassProvider) getData() {
	groups := k.db.Content.Root.Groups
	var group *gokeepasslib.Group
	if k.config.Config.Group == "" {
		group = &groups[0]
	} else {
		group = getGroup(groups[0].Groups, k.config.Config.Group)
	}
	if group == nil {
		return
	}

	for _, entry := range group.Entries {
		k.data[entry.GetTitle()] = entry.GetPassword()
	}
	return
}

func getGroup(groups []gokeepasslib.Group, name string) (result *gokeepasslib.Group) {
	innerGroups := strings.Split(name, ".")
	for _, iGroup := range innerGroups {
		for _, group := range groups {
			if group.Name == iGroup {
				result = &group
				groups = group.Groups
				break
			}
		}
	}
	return
}

func (k *KeepassProvider) GenerateCreds() (string, error) {
	k.getData()
	return "", nil
}

func (k *KeepassProvider) ExportCreds() []string {
	result := []string{}
	for dbKey, osEnv := range k.config.Config.SecretMap {
		if value, ok := k.data[dbKey]; ok {
			result = append(result, fmt.Sprintf("export %s=%s", osEnv, value))
		}
	}
	return result
}

func (k *KeepassProvider) LoadProfileCreds(info []string) {
	values := make([]string, 0, len(k.config.Config.SecretMap))
	for _, v2 := range k.config.Config.SecretMap {
		values = append(values, v2)
	}

	for _, i := range info {
		parts := strings.Split(i, "=")
		if contains(values, parts[0]) {
			k.data[parts[0]] = parts[1]
		}
	}
}

func (k *KeepassProvider) ProfileCreds() []string {
	result := []string{}
	for dbKey, osEnv := range k.config.Config.SecretMap {
		if value, ok := k.data[dbKey]; ok {
			result = append(result, fmt.Sprintf("%s=%s", osEnv, value))
		}
	}
	return result
}

func (k *KeepassProvider) CredsLoaded() bool {
	for _, os := range k.config.Config.SecretMap {
		if _, ok := k.data[os]; !ok {
			return false
		}
	}
	return true
}
