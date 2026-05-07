# Examples

Layout follows [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs)
conventions, so `tfplugindocs generate` can pick them up unmodified
once docs generation is wired up.

```
provider/                          provider block example -> docs/index.md
data-sources/<name>/data-source.tf -> docs/data-sources/<name>.md
resources/<name>/resource.tf       -> docs/resources/<name>.md
```

## Running against a local build

The provider is not published to a registry yet. Use a `dev_overrides`
block in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "langri-sha/aperature" = "/path/to/aperature/dist"
  }
  direct {}
}
```

Then `go build -o dist/terraform-provider-aperature ./cmd/terraform-provider-aperature`
and run `terraform plan` from any of the example directories. Skip
`terraform init` — dev overrides bypass the lock file.
