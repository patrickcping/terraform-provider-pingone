terraform {
  required_providers {
    pingone = {
      version = "0.2"
      source  = "patrickcping/pingone"
    }
  }
}

provider "pingone" {
  client_id       = var.p1_adminClientId
  client_secret   = var.p1_adminClientSecret
	environment_id  = var.p1_adminEnvId
	region          = var.p1_region
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

### Application
resource "pingone_application_oidc" "oidc_web_app" {
  environment_id = pingone_environment.test.environment_id

  name = "Test App 1"
  description = "Test app 1 description"
  enabled = true

  type = "WEB_APP"
  grant_types = ["AUTHORIZATION_CODE"]
  response_types = ["CODE"]
  token_endpoint_authn_method = "CLIENT_SECRET_BASIC"

  redirect_uris = ["https://localhost"]

  access_control {
    group {
      type = "ANY_GROUP"
      groups = [pingone_group.test_group.id]
    }
  }

}

data "pingone_application_oidc_secret" "oidc_web_app_secret" {
  environment_id = pingone_environment.test.environment_id
  application_id = pingone_application_oidc.oidc_web_app.id
}

resource "pingone_application_oidc" "worker_app" {
  environment_id = pingone_environment.test.environment_id

  name = "Test App 2"
  description = "Test app 2 description"
  enabled = true

  type = "WORKER"
  grant_types = ["CLIENT_CREDENTIALS"]
  token_endpoint_authn_method = "CLIENT_SECRET_BASIC"

}

data "pingone_resource" "openid_resource" {
  environment_id = pingone_environment.test.environment_id

  name = "openid"
}

data "pingone_resource_scope" "openid_profile" {
  environment_id = pingone_environment.test.environment_id
  resource_id = data.pingone_resource.openid_resource.id

  name = "profile"

}

data "pingone_resource_scope" "openid_email" {
  environment_id = pingone_environment.test.environment_id
  resource_id = data.pingone_resource.openid_resource.id

  name = "email"

}

resource "pingone_application_resource_grant" "oidc_web_app" {
  environment_id = pingone_environment.test.environment_id
  application_id = pingone_application_oidc.oidc_web_app.id

  resource_id = data.pingone_resource.openid_resource.id
  scopes = [
    data.pingone_resource_scope.openid_profile.id,
    data.pingone_resource_scope.openid_email.id
  ]
}

resource "pingone_application_attribute_mapping" "username" {
  environment_id = pingone_environment.test.environment_id
  application_id = pingone_application_oidc.oidc_web_app.id

  name = "test"
  value = "$${user.email}"
  required = false
}