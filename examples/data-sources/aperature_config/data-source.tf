data "aperature_config" "main" {
  providers = {
    openai = {
      baseurl = "https://api.openai.com/v1"
      models  = ["openai/gpt-5.5", "openai/gpt-5.2"]
      apikey  = "env:OPENAI_API_KEY"
    }
    anthropic = {
      baseurl = "https://api.anthropic.com/v1"
      models  = ["anthropic/claude-opus-4-7", "anthropic/claude-sonnet-4-6"]
      apikey  = "env:ANTHROPIC_API_KEY"
    }
  }

  grants = [
    {
      src = ["*"]
      capabilities = [
        { role = "user" },
        { models = "anthropic/**" },
      ]
    },
    {
      src = ["group:admins"]
      capabilities = [
        { role = "admin" },
        { models = "**" },
      ]
    },
  ]

  quotas = {
    monthly = {
      capacity  = 100
      rate      = 100
      on_exceed = "block"
    }
  }
}

resource "local_file" "aperture_json" {
  filename        = "${path.module}/aperture.json"
  content         = data.aperature_config.main.json
  file_permission = "0600"
}
