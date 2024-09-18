package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/smorenodp/clusterprofile/config"
)

const (
	clusterProfileEnv = "CLUSTERID_PROFILE"
)

type CommandArgs struct {
	ProfilesConfig  string
	CredentialsFile string
	ExecutableFile  string
	Profile         string
	Echo            bool
	Banner          config.Banner
}

var (
	errorLog = log.New(os.Stderr, "", 0)
)

func getOrElse(env, valueDefault string) string {
	if value := os.Getenv("env"); value == "" {
		return valueDefault
	} else {
		return value
	}
}

func parse() (args CommandArgs) {
	var banner config.Banner
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Error obtaining home directory.")
	}

	flag.StringVar(&args.ProfilesConfig, "profileFolder", getOrElse("CLUSTERID_CONFIG_FOLDER", fmt.Sprintf("%s/.clusterid/profiles/", home)), "Config folder for program.")
	flag.StringVar(&args.CredentialsFile, "creds", getOrElse("CLUSTERID_PROFILE_FILE", fmt.Sprintf("%s/.clusterid/credentials", home)), "Creds file for program.")
	flag.StringVar(&args.ExecutableFile, "exec", getOrElse("CLUSTERID_EXEC_FILE", fmt.Sprintf("%s/.clusterid/export.sh", home)), "Bash file to export the configuration for the profile.")
	flag.StringVar(&args.Profile, "profile", getOrElse("PROFILE_NAME", ""), "Name of the profile to load")
	flag.BoolVar(&args.Echo, "echo", false, "Output the export instructions like an echo")
	flag.BoolVar(&banner.Enable, "banner", false, "Show banner outputing the profile loaded")
	flag.StringVar(&banner.Command, "banner-cmd", "figlet", "Command for the banner")
	flag.StringVar(&banner.Args, "banner-args", "", "Arguments for the banner")

	flag.Parse()
	args.Banner = banner
	return
}

func main() {

	args := parse()

	cp, err := NewClusterProfile(args)
	fmt.Println(err)
	if err != nil {
		log.Fatalf("Error generating clusterprofile - %s", err)
	}
	cp.Run()

	if err = config.SaveCreds(args.CredentialsFile, cp.profilesCreds); err != nil {
		log.Fatalf("Error saving credentials in %s - %s", args.CredentialsFile, err)
	}

	if args.Echo {
		config.GenerateExportContent(args.Profile, cp.profile.Export, args.Banner)
	} else {
		config.CreateExecFile(args.ExecutableFile, args.Profile, cp.profile.Export, args.Banner)
	}
}
