package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v59/github"
)

var packageName, org, state, visibility string

func init() {
	flag.StringVar(&packageName, "package", "", "the package to clean")
	flag.StringVar(&org, "org", "", "the organization")
	flag.StringVar(&visibility, "visibility", "public", "clean 'public' or 'private' images")
	flag.StringVar(&state, "state", "active", "must be 'active' or 'deleted'")
}

func main() {
	flag.Parse()

	if packageName == "" {
		missingFlag("package")
	}
	if org == "" {
		missingFlag("org")
	}

	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	packages, _, err := client.Organizations.ListPackages(ctx, org, &github.PackageListOptions{
		Visibility:  github.String(visibility),
		PackageType: github.String("container"),
		State:       github.String(state),
	})

	if err != nil {
		printError(err)
	}

	for _, p := range packages {
		if p.GetName() != packageName {
			continue
		}

		fmt.Printf("%s:\n", p.GetName())

		versions, _, err := client.Organizations.PackageGetAllVersions(ctx, org, *p.PackageType, *p.Name, &github.PackageListOptions{
			Visibility:  github.String(visibility),
			PackageType: p.PackageType,
			State:       github.String(state),
		})
		if err != nil {
			printError(err)
		}

		for _, v := range versions {
			tags := v.GetMetadata().GetContainer().Tags
			for _, t := range tags {
				if !strings.HasPrefix(t, "pr-") {
					continue
				}

				fmt.Printf("Deleting: %q\n", t)

				_, err = client.Organizations.PackageDeleteVersion(ctx, org, *p.PackageType, *p.Name, v.GetID())
				if err != nil {
					printError(err)
				}
			}
		}

		fmt.Println("")
	}
}

func missingFlag(flagName string) {
	fmt.Printf("missing --%s\n", flagName)
	os.Exit(1)
}

func printError(err error) {
	fmt.Printf("an error occurred: %s", err)
	os.Exit(2)
}
