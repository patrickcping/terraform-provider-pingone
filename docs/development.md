# Developing the Provider

## Requirements
* Terraform 1.0.8
* Go 1.15

## Quick Start

If you wish to work on the provider, you'll first need [Go](https://golang.org) installed on your machine (version 1.15+ is preferred). You'll also need to correctly setup a [GOPATH](https://golang.org/doc/code#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

See above for which option suits your workflow for building the provider.

In order to build the provider, you can simply run `make build`.  This will build the provider and put the provider binary in the local directory

```shell
make build
```

To install the provider locally and test, you can run `make install`

```shell
make install
```

Provider testing is TBD.

This provider uses the Terraform Plugin SDKv2, documentation on developing Terraform providers can be found on the [Terraform website](https://www.terraform.io/docs/extend/sdkv2-intro.html)

## Using the Provider

Terraform provide, [development overrides for provider developers](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers) to test providers built from source.

To do this, populate a Terraform CLI configuration file (`~/.terraformrc` for
all platforms other than Windows; `terraform.rc` in the `%APPDATA%` directory
when using Windows) with at least the following options:

```
provider_installation {
  dev_overrides {
    "patrickcping/pingone" = "/path/to/github/clone/pingone-terraform-provider"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

You will need to replace the path with the **full path** where the cloned repository lives, no `~` shorthand.

Once you have this file in place, you can run `make build` which will
build a development version of the binary in the repository that Terraform
will use instead of the version from the remote registry.

## Test sample configuration

First, build and install the provider.

```shell
make build
```

Set the required provider setup variables.  This requires a trial or licensed PingOne instance with a `client credentials` admin client created with `Client Secret Basic` authentication.
Note that the admin client used must have the `ORGANIZATION ADMIN` role in the PingOne organisation to create environments and ideally the `ENVIRONMENT_ADMIN` role scoped to the organisation to be able to create resources in the newly created environment

```shell
export TF_VAR_p1_adminClientId=$YOUR_ADMIN_CLIENT_ID
export TF_VAR_p1_adminClientSecret=$YOUR_ADMIN_CLIENT_SECRET
export TF_VAR_p1_adminEnvId=$YOUR_ADMIN_CLIENT_ENVIRONMENT_ID
export TF_VAR_p1_region=EU
export TF_VAR_p1_licenseId=$YOUR_LICENSE_ID_FOR_ENV_CREATION
```

Set the region to one of `EU`, `US`, `ASIA`, `CA`

Run the provider configuration
```shell
terraform apply
```