terraform {
  required_providers {
    pingone = {
      version = "0.2"
      source  = "pingidentity.com/edu/pingone"
    }
  }
}

provider "pingone" {
  client_id       = var.p1_adminClientId
  client_secret   = var.p1_adminClientSecret
	access_token    = "<<fill this in if you want to test - client creds routine tbc>>"
	environment_id  = var.p1_adminEnvId
	domain_suffix   = "EU"
}

data "pingone_environment" "admin_env" {
  name = "Administrators"
}

data "pingone_role" "identity_admin" {
  name = "Identity Data Admin"
}

resource "pingone_environment" "test" {
  name = "WUT THIS IS TERRAFORMED"
  description = "Description of WUT1"
  type = "SANDBOX"
  region = "EU"

  license_id = var.p1_licenseId

  default_population {
    description = "tbc"
  }

  product {
    type = "PING_ONE_BASE"
  }

  product {
    type = "PING_ONE_RISK"
  }

  product {
    type = "PING_ONE_MFA"
  }

  product {
    type = "PING_FEDERATE"
    console_href = "https://thisismylink.com"

    bookmark {
      name = "Test Bookmark"
      href = "https://www.pingidentity.com"
    }

    bookmark {
      name = "Test Bookmark2"
      href = "https://www.google.co.uk"
    }
  }

  product {
    type = "PING_ACCESS"
    console_href = "https://thisismylink.com"

    bookmark {
      name = "Test Bookmark"
      href = "https://www.pingidentity.com"
    }
  }
  
}

resource "pingone_role_assignment" "admin_role_assignment" {
  environment_id = data.pingone_environment.admin_env.id
  user_id = "3944a587-a1aa-4378-9933-e9f9a2ad59fe" // My test user
  role_id = data.pingone_role.identity_admin.id
  scope_id = pingone_environment.test.id
  scope_type = "ENVIRONMENT"
}

resource "pingone_population" "customers_a" {
  environment_id = pingone_environment.test.id

  name = "Customers A"
  description = "WUT this is terraformed"
}

resource "pingone_population" "customers_b" {
  environment_id = pingone_environment.test.id

  name = "Customers B"
  description = "WUT this is terraformed"
}

resource "pingone_group" "test_group" {
  environment_id = pingone_environment.test.id

  name = "Test"
  description = "Test group"
  population_id = pingone_population.customers_a.id
}