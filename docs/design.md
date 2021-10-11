# PingOne Terraform Provider Design

## PingOne API and SDK

The PingOne Terraform provider leverages the [PingOne Platform API](https://apidocs.pingidentity.com/pingone/platform/v1/api/), via an automatically generated [PingOne Go SDK](https://github.com/patrickcping/pingone-go).  The resources in this provider must use the Go SDK to call PingOne platform endpoints, rather than call API endpoints directly.

Like this provider, the PingOne Go SDK is also in an early build/test phase and is generated from a common [PingOne OpenAPI v3 specification](https://github.com/patrickcping/pingone-openapi-specs/blob/main/managementAPIs.yml).  Most PingOne endpoints exist in the SDK, however some have been 'cleansed' and others have not.  You can usually tell from the SDK function whether a function has been cleansed in the OpenAPI spec, as it will have a more 'human-readable' function name and will not begin with `V1EnvironmentsEnvID....`.  Examples are:

`api_client.ManagementAPIsEnvironmentsApi.ReadAllEnvironments(...)` - Cleansed SDK
`api_client.ManagementAPIsAgreementManagementAgreementsResourcesApi.V1EnvironmentsEnvIDAgreementsAgreementIDGet(...)` - Yet to be cleansed in the OpenAPI spec

For PingOne API resources not present in the SDK, in either cleansed or uncleansed form, these will need to be added to the [PingOne OpenAPI v3 specification](https://github.com/patrickcping/pingone-openapi-specs/blob/main/managementAPIs.yml) so that it can be generated into code form into the SDK.