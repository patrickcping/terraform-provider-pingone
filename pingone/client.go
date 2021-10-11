package pingone

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go"
)

type providerData struct {
	ClientId      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	AccessToken   types.String `tfsdk:"access_token"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	DomainSuffix  types.String `tfsdk:"domain_suffix"`
}

type p1Client struct {
	//ConfigStore                               configStore.ConfigStoreAPI
}

func (c *providerData) ApiClient() (*pingone.APIClient, diag.Diagnostics) {
	// var err error
	var client *pingone.APIClient

	clientcfg := pingone.NewConfiguration()
	clientcfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken.Value))
	client = pingone.NewAPIClient(clientcfg)

	log.Printf("[INFO] PingOne Client configured")
	return client, nil
}

// func (c *providerData) GetToken(apiClient *pingone.APIClient) (interface{}, error) {

// 	ctx := context.Background()

// 	//Get URL from SDK
// 	url, err := apiClient.GetConfig().ServerURL(1, map[string]string{
// 		"Description": fmt.Sprintf("PingOne %s AUTH", strings.ToUpper(c.DomainSuffix)), //POC Code
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	//OAuth 2.0 config for client creds
// 	config := clientcredentials.Config{
// 		ClientID:     c.ClientId,
// 		ClientSecret: c.ClientSecret,
// 		TokenURL:     fmt.Sprintf("%s/%s/as/token", url, c.EnvironmentID),
// 		AuthStyle:    oauth2.AuthStyleAutoDetect,
// 	}

// 	token, err := config.Token(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return token, nil
// }
