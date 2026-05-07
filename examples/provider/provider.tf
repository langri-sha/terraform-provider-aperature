terraform {
  required_providers {
    aperature = {
      source  = "langri-sha/aperature"
      version = "~> 0.1"
    }
  }
}

provider "aperature" {
  endpoint = "http://aperture.tailnet.ts.net"
  # auth_token defaults to $APERATURE_AUTH_TOKEN.
}
