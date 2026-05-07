# NOTE: This resource currently errors on apply — Aperture's
# management API is not yet public. The schema is finalized so plans
# work today; CRUD lights up once Tailscale ships the API.

resource "aperature_grant" "all_users_anthropic" {
  src = ["*"]
  capabilities = [
    { role = "user" },
    { models = "anthropic/**" },
  ]
}
