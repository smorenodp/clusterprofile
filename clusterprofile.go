package main

import (
	"fmt"

	"github.com/smorenodp/clusterprofile/config"
	"github.com/smorenodp/clusterprofile/providers"
)

type Profile struct {
	Name   string
	Export []string
	Creds  []string
}

type ClusterProfile struct {
	vaultClient    *providers.VaultClient
	profile        Profile
	profilesConfig map[string]config.ClusterConfig // TODO: Change to type
	profilesCreds  config.CredConfig
}

func NewClusterProfile(args CommandArgs) (*ClusterProfile, error) {
	profiles, err := config.ReadConfig(args.ProfilesConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file from folder %s - %s", args.ProfilesConfig, err)
	}
	creds, err := config.LoadCreds(args.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("Error parsing creds file from %s", args.CredentialsFile)
	}
	p := Profile{Name: args.Profile, Creds: []string{}, Export: []string{fmt.Sprintf("export %s=%s", clusterProfileEnv, args.Profile)}}
	cluster := &ClusterProfile{profilesConfig: profiles, profile: p, profilesCreds: creds}

	err = cluster.generateVaultClient()
	return cluster, err
}

// TODO: Refactor this function
func (cp *ClusterProfile) generateVaultClient() (err error) {
	var profile config.ClusterConfig
	var creds []string
	var client *providers.VaultClient
	if profile, creds, err = cp.GetProfile(cp.profile.Name); err != nil {
		return
	}
	client, err = providers.NewVaultClient(profile.Vault)

	cp.vaultClient = client

	if loaded := client.LoadProfileCreds(creds); loaded {
		cp.profile.Export = append(cp.profile.Export, cp.vaultClient.ExportCreds()...)
		cp.profile.Creds = append(cp.profile.Creds, cp.vaultClient.ProfileCreds()...)
		return
	}
	if profile.Vault.PivotProfile != "" {
		var pivotConfig config.ClusterConfig
		var pivotCreds []string
		if pivotConfig, pivotCreds, err = cp.GetProfile(profile.Vault.PivotProfile); err != nil {
			return
		}

		_, err := client.WithPivotRole(pivotConfig.Vault, pivotCreds)
		cp.profilesCreds[profile.Vault.PivotProfile] = client.Pivot.ProfileCreds()
		if err != nil {
			return err
		}
		_, err = client.GenerateCreds()
		if err != nil {
			return err
		}
	} else {
		_, err = client.GenerateCreds()
		if err != nil {
			return err
		}
	}
	if cp.vaultClient.CredsLoaded() {
		cp.profile.Export = append(cp.profile.Export, cp.vaultClient.ExportCreds()...)
		cp.profile.Creds = append(cp.profile.Creds, cp.vaultClient.ProfileCreds()...)
	}
	return
}

func (cp *ClusterProfile) GetProfile(name string) (pConfig config.ClusterConfig, pCreds []string, err error) {
	var ok bool
	if pConfig, ok = cp.profilesConfig[name]; !ok {
		err = fmt.Errorf("Error loading profile config %s", name)
		return
	}
	pCreds = cp.profilesCreds[name]

	return
}

func (cp *ClusterProfile) RemoveProfile(name string) (err error) {
	var ok bool
	if _, ok = cp.profilesCreds[name]; !ok {
		err = fmt.Errorf("Error removing profile config %s", name)
		return
	}
	delete(cp.profilesCreds, name)

	return
}

func (cp *ClusterProfile) ExecuteProviders() {
	pConfig, pCreds, _ := cp.GetProfile(cp.profile.Name)
	for _, p := range pConfig.Providers {
		provider := providers.NewProvider(cp.vaultClient, p)
		if provider != nil {
			provider.LoadProfileCreds(pCreds)
			if !provider.CredsLoaded() {
				provider.GenerateCreds()
			}
			if provider.CredsLoaded() {
				cp.profile.Export = append(cp.profile.Export, provider.ExportCreds()...)
				cp.profile.Creds = append(cp.profile.Creds, provider.ProfileCreds()...)
			}
		} else {
			errorLog.Println("Provider of type %s not implemented\n", p.Type)
		}
	}
}

func (cp *ClusterProfile) Run() {
	cp.ExecuteProviders()
	cp.profilesCreds[cp.profile.Name] = cp.profile.Creds //TODO: Change this, i don't like it
}
