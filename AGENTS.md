# AGENTS.md

Context for AI coding agents working in this repo.

## What this is

A Terraform provider, written in Go using the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework),
for [Aperture by Tailscale](https://tailscale.com/docs/aperture). See
[`README.md`](./README.md) for the user-facing description and the
two-layer split (`aperature_config` data source today, resource CRUD
once upstream ships a management API).

## Upstream surface (as of 2026-05)

Aperture is in open alpha. Configuration today is a single JSON
document — there is **no public administrative HTTP API**. The
documented top-level keys of that document are:

- `providers` — map of LLM provider configs (`baseurl`, `models[]`,
  `apikey`, `authorization`, `compatibility{}`, `cost_basis`,
  `preference`, `disabled`, `add_headers`, `model_cost_map`)
- `grants[]` — Tailscale grant entries: `src[]` and
  `app["tailscale.com/cap/aperture"][]` capabilities (`role: user|admin`,
  `models: <glob>`)
- `quotas` — token-bucket spend limits (`capacity`, `rate`, `on_exceed`)
- `hooks` — webhook configs (`url`, `apikey`, `authorization`,
  `timeout`, `disabled`, `fail_policy`, `preference`)
- `exporters` — S3-compatible session-log export
- `mcp` — Model Context Protocol proxy (`accept_registrations`,
  `servers{}`)
- `database` — `retention` config
- `auto_cost_basis` — boolean

The provider's schema mirrors these names verbatim. **Do not invent
fields.** If upstream doesn't document it, leave it out and add a TODO.

## Layout

```
cmd/terraform-provider-aperature/   # main.go — providerserver entrypoint
internal/aperture/                  # HTTP client (currently stub-only)
internal/provider/                  # plugin-framework provider + resources + data sources
examples/                           # terraform-plugin-docs example layout
docs/                               # generated docs (terraform-plugin-docs)
.github/workflows/                  # CI
```

Resources and data sources each get one file under
`internal/provider/`: `<name>_resource.go` or `<name>_data_source.go`,
plus a sibling `_test.go`. Schemas live in the same file as the
implementation.

## Conventions

- **Atomic commits.** One logical change per commit.
- **No idempotent fluff.** No `cmd || true`, no swallowed errors, no
  "tolerate pre-existing state" patterns. Strict failure with a clear
  error beats silent recovery.
- **Don't speculate, verify.** Before claiming "Aperture supports X",
  check `tailscale.com/docs/aperture` or `tailscale/aperture-cli`.
- **Comments explain *why*, not *what*.** The reader can see what the
  code does. They can't see why a field is `Optional` instead of
  `Required`, or why a resource Create returns an error today.
- **No emojis** unless explicitly requested.
- **Field names mirror upstream JSON.** `baseurl`, not `base_url`;
  `apikey`, not `api_key`. Keep HCL → JSON one-to-one wherever possible
  so users can cross-reference Aperture docs without translation.

## Common commands

```sh
go mod tidy
go vet ./...
go test ./...

# Format example HCL
terraform fmt -recursive examples/

# Generate docs (once tfplugindocs is wired up)
go generate ./...
```

## Pre-alpha caveats

- Resource Create/Update/Delete return a typed "upstream API not yet
  public" error. Don't paper over this with mock backends or
  speculative endpoint paths — the error is the feature until Tailscale
  ships the API.
- The HTTP client in `internal/aperture/` is a placeholder. When the
  real API arrives, replace `Client.do` and add typed methods; keep the
  package boundary so the provider package never touches `net/http`
  directly.

## When in doubt

Re-read `tailscale.com/docs/aperture/configuration` and
`tailscale.com/docs/aperture/how-to/grant-model-access`. Those two
pages are the source of truth for the schema.
