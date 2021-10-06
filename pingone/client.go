package pingone

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/patrickcping/pingone-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type p1ClientConfig struct {
	ClientId      string
	ClientSecret  string
	AccessToken   string
	EnvironmentID string
	DomainSuffix  string
	APIToken      string
}

type p1Client struct {
	//ConfigStore                               configStore.ConfigStoreAPI
}

func (c *p1ClientConfig) ApiClient() (*pingone.APIClient, diag.Diagnostics) {
	// var err error
	var client *pingone.APIClient

	clientcfg := pingone.NewConfiguration()
	clientcfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	client = pingone.NewAPIClient(clientcfg)

	log.Printf("[INFO] PingOne Client configured")
	return client, nil
}

func (c *p1ClientConfig) GetToken(apiClient *pingone.APIClient) (interface{}, error) {

	ctx := context.Background()

	//Get URL from SDK
	url, err := apiClient.GetConfig().ServerURL(1, map[string]string{
		"Description": fmt.Sprintf("PingOne %s AUTH", strings.ToUpper(c.DomainSuffix)), //POC Code
	})
	if err != nil {
		return nil, err
	}

	//OAuth 2.0 config for client creds
	config := clientcredentials.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/%s/as/token", url, c.EnvironmentID),
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	return token, nil
}
