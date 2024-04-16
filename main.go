package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/smorenodp/clusterprofile/config"
	"github.com/smorenodp/clusterprofile/providers"
)

const (
	clusterProfileEnv = "CLUSTERID_PROFILE"
)

func getOrElse(env, valueDefault string) string {
	if value := os.Getenv("env"); value == "" {
		return valueDefault
	} else {
		return value
	}
}

func main() {
	var configFolder, credsFile, profileName, execFile string
	var echo bool
	var client *providers.VaultClient
	var profileCreds []string
	errorLog := log.New(os.Stderr, "", 0)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error obtaining home directory.")
	}
	flag.StringVar(&configFolder, "profileFolder", getOrElse("CLUSTERID_CONFIG_FOLDER", fmt.Sprintf("%s/.clusterid/profiles/", home)), "Config folder for program.")
	flag.StringVar(&credsFile, "creds", getOrElse("CLUSTERID_PROFILE_FILE", fmt.Sprintf("%s/.clusterid/credentials", home)), "Creds file for program.")
	flag.StringVar(&execFile, "exec", getOrElse("CLUSTERID_EXEC_FILE", fmt.Sprintf("%s/.clusterid/export.sh", home)), "Bash file to export the configuration for the profile.")
	flag.StringVar(&profileName, "profile", getOrElse("PROFILE_NAME", ""), "Name of the profile to load")
	flag.BoolVar(&echo, "echo", false, "Output the export instructions like an echo")

	flag.Parse()

	profiles, err := config.ReadConfig(configFolder)

	if err != nil {
		log.Fatalf("Error parsing config file from folder %s - %s\n", configFolder, err)
	}

	creds, err := config.LoadCreds(credsFile)
	if err != nil {
		log.Fatalf("Error parsing creds file from %s\n", credsFile)
	}

	c := profiles[profileName]
	client, err = providers.NewVaultClient(c.Vault)
	client.LoadProfileCreds(creds[profileName])
	if c.Vault.PivotProfile != "" {
		if clusterConfig, ok := profiles[c.Vault.PivotProfile]; !ok {
			log.Fatalf("Error loading profile pivoting %s in profile %s configuration", c.Vault.PivotProfile, c.Name)
		} else {
			client = client.WithPivotRole(clusterConfig.Vault, creds[c.Vault.PivotProfile])
			creds[c.Vault.PivotProfile] = client.ProfileCreds()
			client.GenerateCreds()
		}
	} else {
		client.GenerateCreds()
	}
	if err != nil {
		log.Fatal("Vault creation failed")
	}
	exportCreds := []string{fmt.Sprintf("export %s=%s", clusterProfileEnv, profileName)}
	profileCreds = client.ProfileCreds()
	exportCreds = client.ExportCreds()
	for _, p := range c.Providers {
		provider := providers.NewProvider(client, p)
		if provider != nil {
			provider.LoadProfileCreds(creds[profileName])
			if !provider.CredsLoaded() {
				provider.GenerateCreds()
			}
			exportCreds = append(exportCreds, provider.ExportCreds()...)
			profileCreds = append(profileCreds, provider.ProfileCreds()...)
		} else {
			errorLog.Println("Provider of type %s not implemented\n", p.Type)
		}
	}
	creds[profileName] = profileCreds

	if err = config.SaveCreds(credsFile, creds); err != nil {
		log.Fatalf("Error saving credentials in %s - %s", credsFile, err)
	}

	if echo {
		config.GenerateExportContent(profileName, exportCreds)
	} else {
		config.CreateExecFile(execFile, profileName, exportCreds)
	}
}
