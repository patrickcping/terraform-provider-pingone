package pingone

import (
	"context"
	"fmt"
	"log"

	"github.com/patrickcping/pingone-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type p1ClientConfig struct {
	ClientId      string
	ClientSecret  string
	EnvironmentID string
	Region        string
}

type p1Client struct {
	APIClient    *pingone.APIClient
	regionSuffix string
}

func (c *p1ClientConfig) ApiClient(ctx context.Context) (*p1Client, error) {
	// var err error
	var client *pingone.APIClient

	token, err := getToken(ctx, c)
	if err != nil {
		return nil, err
	}

	clientcfg := pingone.NewConfiguration()
	clientcfg.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	client = pingone.NewAPIClient(clientcfg)

	var regionSuffix string
	switch p1Region := c.Region; p1Region {
	case "EU":
		regionSuffix = "eu"
	case "US":
		regionSuffix = "com"
	case "ASIA":
		regionSuffix = "asia"
	case "CA":
		regionSuffix = "ca"
	default:
		regionSuffix = "com"
	}

	log.Printf("[INFO] PingOne Client using region suffix %s", regionSuffix)

	apiClient := &p1Client{
		APIClient:    client,
		regionSuffix: regionSuffix,
	}

	log.Printf("[INFO] PingOne Client configured")
	return apiClient, nil
}

func getToken(ctx context.Context, c *p1ClientConfig) (*oauth2.Token, error) {

	//Get URL from SDK
	authUrl := "https://auth.pingone"
	switch p1Region := c.Region; p1Region {
	case "EU":
		authUrl = fmt.Sprintf("%s.eu", authUrl)
	case "US":
		authUrl = fmt.Sprintf("%s.com", authUrl)
	case "ASIA":
		authUrl = fmt.Sprintf("%s.asia", authUrl)
	case "CA":
		authUrl = fmt.Sprintf("%s.ca", authUrl)
	default:
		authUrl = fmt.Sprintf("%s.com", authUrl)
	}

	log.Printf("[INFO] Getting token from %s", authUrl)

	//OAuth 2.0 config for client creds
	config := clientcredentials.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/%s/as/token", authUrl, c.EnvironmentID),
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] Token retrieved")

	return token, nil
}
