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

## Get Roles
data "pingone_role" "organisation_admin" {
  name = "Organization Admin"
}

data "pingone_role" "environment_admin" {
  name = "Environment Admin"
}

data "pingone_role" "identity_data_admin" {
  name = "Identity Data Admin"
}

data "pingone_role" "client_application_developer" {
  name = "Client Application Developer"
}

data "pingone_role" "identity_data_admin_ro" {
  name = "Identity Data Read Only"
}
## End

data "pingone_group" "test_group" {
  environment_id = data.pingone_environment.admin_env.id
  name = "Advanced Services Admin"
}

resource "pingone_environment" "test" {
  name = "AAA WUT THIS IS TERRAFORMED"
  description = "Description of WUT1"
  type = "SANDBOX"
  region = "EU"

  license_id = var.p1_licenseId

  default_population_name = "Default2"
  default_population_description = "tbc"

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
  role_id = data.pingone_role.identity_data_admin.id
  scope_id = pingone_environment.test.environment_id
  scope_type = "ENVIRONMENT"
}

resource "pingone_population" "customers_a" {
  environment_id = pingone_environment.test.environment_id

  name = "Customers A"
  description = "WUT this is terraformed"
}

resource "pingone_population" "customers_b" {
  environment_id = pingone_environment.test.environment_id

  name = "Customers B"
  description = "WUT this is terraformed"
}

resource "pingone_group" "test_group" {
  environment_id = pingone_environment.test.environment_id

  name = "Test"
  description = "Test group"
  population_id = pingone_population.customers_a.id
}

### Attributes
data "pingone_schema" "attribute_schema" {
  environment_id = pingone_environment.test.environment_id
  name = "User"
}

resource "pingone_schema_attribute" "test_attribute" {
  environment_id = pingone_environment.test.environment_id
  schema_id = data.pingone_schema.attribute_schema.id

  name = "testAttribute"
  display_name = "Test Attribute for TF"
  description = "A description"

  enabled = true
  unique = false
  required = false
  
}