# ghcr-janitor

Clean-up `pr-` images in your org's Github Container Registry.

```sh
❯ GITHUB_TOKEN=123 ghcr-janitor --org hostwithquantum --package hugo-docker
hugo-docker:
Deleting: "pr-6"
```

## Usage

```shell
❯ ghcr-janitor --help
Usage of ghcr-janitor:
  -org string
    	the organization
  -package string
    	the package to clean
  -state string
    	must be 'active' or 'deleted' (default "active")
  -visibility string
    	clean 'public' or 'private' images (default "public")
```
