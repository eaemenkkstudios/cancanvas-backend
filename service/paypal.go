package service

import (
	"log"
	"os"

	paypalsdk "github.com/plutov/paypal"
)

// PayPalService interface
type PayPalService interface {
	Pay() (bool, error)
}

type payPalService struct {
	client     *paypalsdk.Client
	acessToken *paypalsdk.TokenResponse
}

func (s *payPalService) Pay() (bool, error) {
	return true, nil
}

// NewPayPalService function
func NewPayPalService() PayPalService {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	secretID := os.Getenv("PAYPAL_SECRET_ID")
	paypalMode := os.Getenv("PAYPAL_MODE")
	apiBase := paypalsdk.APIBaseSandBox
	if paypalMode == "prod" {
		apiBase = paypalsdk.APIBaseLive
	}
	client, err := paypalsdk.NewClient(clientID, secretID, apiBase)
	client.SetLog(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	acessToken, err := client.GetAccessToken()
	if err != nil {
		log.Fatal(err)
	}
	return &payPalService{
		client:     client,
		acessToken: acessToken,
	}
}
