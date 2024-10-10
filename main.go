package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smorenodp/clusterprofile/config"
	"github.com/urfave/cli/v3"
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

func load(args CommandArgs) error {
	cp, err := NewClusterProfile(args)
	if err != nil {
		return fmt.Errorf("error generating clusterprofile - %s", err)
	}
	cp.Run()

	if err = config.SaveCreds(args.CredentialsFile, cp.profilesCreds); err != nil {
		return fmt.Errorf("error saving credentials in %s - %s", args.CredentialsFile, err)
	}

	if args.Echo {
		config.GenerateExportContent(args.Profile, cp.profile.Export, args.Banner)
	} else {
		config.CreateExecFile(args.ExecutableFile, args.Profile, cp.profile.Export, args.Banner)
	}
	return nil
}

func show(args CommandArgs) error {

	cp, err := NewClusterProfile(args)
	if err != nil {
		return fmt.Errorf("error generating clusterprofile - %s", err)
	}
	_, creds, err := cp.GetProfile(args.Profile)

	if err != nil {
		return fmt.Errorf("error getting profile - %s", err)
	}
	if len(creds) > 0 {
		fmt.Printf("[%s]\n", args.Profile)
		for _, c := range creds {
			fmt.Println(c)
		}
	}

	return nil
}

func remove(args CommandArgs) error {

	cp, err := NewClusterProfile(args)
	if err != nil {
		return fmt.Errorf("error generating clusterprofile - %s", err)
	}
	err = cp.RemoveProfile(args.Profile)

	if err != nil {
		return fmt.Errorf("error removing profile - %s", err)
	}

	if err = config.SaveCreds(args.CredentialsFile, cp.profilesCreds); err != nil {
		return fmt.Errorf("error saving credentials in %s - %s", args.CredentialsFile, err)
	}

	return nil
}

func main() {
	var args CommandArgs = CommandArgs{}
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Error obtaining home directory.")
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "profileFolder",
				Aliases:     []string{"pf"},
				Value:       getOrElse("CLUSTERID_CONFIG_FOLDER", fmt.Sprintf("%s/.clusterid/profiles/", home)),
				Usage:       "Config folder for program.",
				Destination: &args.ProfilesConfig,
			},
			&cli.StringFlag{
				Name:        "creds",
				Aliases:     []string{"c"},
				Value:       getOrElse("CLUSTERID_PROFILE_FILE", fmt.Sprintf("%s/.clusterid/credentials", home)),
				Usage:       "Creds file for program.",
				Destination: &args.CredentialsFile,
			},
			&cli.StringFlag{
				Name:        "exec",
				Aliases:     []string{"e"},
				Value:       getOrElse("CLUSTERID_EXEC_FILE", fmt.Sprintf("%s/.clusterid/export.sh", home)),
				Usage:       "Bash file to export the configuration for the profile.",
				Destination: &args.ExecutableFile,
			},
			&cli.StringFlag{
				Name:        "profile",
				Aliases:     []string{"p"},
				Value:       getOrElse("PROFILE_NAME", ""),
				Usage:       "Name of the profile to load",
				Destination: &args.Profile,
			},
			&cli.BoolFlag{
				Name:        "echo",
				Value:       false,
				Usage:       "Output the export instructions like an echo",
				Destination: &args.Echo,
			},
			&cli.BoolFlag{
				Name:        "banner",
				Aliases:     []string{"b"},
				Value:       false,
				Usage:       "Show banner outputing the profile loaded",
				Destination: &args.Banner.Enable,
			},
			&cli.StringFlag{
				Name:        "banner-cmd",
				Aliases:     []string{"bc"},
				Value:       "figlet",
				Usage:       "Command for the banner",
				Destination: &args.Banner.Command,
			}, &cli.StringSliceFlag{
				Name:        "banner-args",
				Aliases:     []string{"ba"},
				Usage:       "Arguments for the banner",
				Destination: &args.Banner.Args,
			},
		},
		DefaultCommand: "load",
		Commands: []*cli.Command{
			{
				Name:    "load",
				Aliases: []string{"l"},
				Usage:   "Load credentials for profile. Generate them if they don't exist or are expired.",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Load")
					return load(args)
				},
			},
			{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "Show credencials if exist",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Show")
					return show(args)
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"r"},
				Usage:   "Remove credencials if exist",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Remove")
					return remove(args)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
