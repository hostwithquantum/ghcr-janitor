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

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN")).Organizations

	packageOpt := &github.PackageListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
		Visibility:  github.String(visibility),
		PackageType: github.String("container"),
		State:       github.String(state),
	}

	var packages []*github.Package
	for {
		packagePage, packageResp, err := client.ListPackages(ctx, org, packageOpt)
		if err != nil {
			printError(err)
		}

		packages = append(packages, packagePage...)
		if packageResp.NextPage == 0 {
			break
		}
		packageOpt.Page = packageResp.NextPage
	}

	for _, p := range packages {
		if p.GetName() != packageName {
			continue
		}

		fmt.Printf("%s:\n", p.GetName())

		versionOpt := &github.PackageListOptions{
			ListOptions: github.ListOptions{PerPage: 10},
			Visibility:  github.String(visibility),
			PackageType: p.PackageType,
			State:       github.String(state),
		}

		var versions []*github.PackageVersion
		for {
			versionPage, versionResp, err := client.PackageGetAllVersions(ctx, org, p.GetPackageType(), p.GetName(), versionOpt)
			if err != nil {
				printError(err)
			}

			versions = append(versions, versionPage...)
			if versionResp.NextPage == 0 {
				break
			}

			versionOpt.Page = versionResp.NextPage
		}

		for _, v := range versions {
			version := v.GetID()
			tags := v.GetMetadata().GetContainer().Tags

			// look for pr- tags
			for _, t := range tags {
				if !strings.HasPrefix(t, "pr-") {
					continue
				}

				fmt.Printf("Deleting: %q (%d)\n", t, v.GetID())

				err := delete(ctx, client, p.GetPackageType(), p.GetName(), version)
				if err != nil {
					printError(err)
				}
				break
			}

			// clean-up untagged
			if len(tags) == 0 {
				fmt.Printf("untagged: %d\n", v.GetID())
				err := delete(ctx, client, p.GetPackageType(), p.GetName(), version)
				if err != nil {
					printError(err)
				}
			}

		}

		fmt.Println("")
	}
}

func delete(ctx context.Context, client *github.OrganizationsService, packageType, packageName string, version int64) error {
	_, err := client.PackageDeleteVersion(ctx, org, packageType, packageName, version)
	return err
}

func missingFlag(flagName string) {
	fmt.Printf("missing --%s\n", flagName)
	os.Exit(1)
}

func printError(err error) {
	fmt.Printf("an error occurred: %s", err)
	os.Exit(2)
}
