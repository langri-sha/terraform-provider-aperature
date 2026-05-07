# NOTE: This resource currently errors on apply — see
# examples/resources/aperature_grant/resource.tf for the rationale.

resource "aperature_llm_provider" "openai" {
  name    = "openai"
  baseurl = "https://api.openai.com/v1"
  models  = ["openai/gpt-5.5", "openai/gpt-5.2"]
  apikey  = "env:OPENAI_API_KEY"
}
