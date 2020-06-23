package service

import (
	"log"
	"os"

	paypalsdk "github.com/plutov/paypal"
)

// PayPalService interface
type PayPalService interface {
	CreateOrder(amount, description string) (Order, error)
	GetOrderStatus(orderID string) (string, error)
	CompleteOrder(orderID string) (string, error)
}

type payPalService struct {
	client       *paypalsdk.Client
	acessToken   *paypalsdk.TokenResponse
	emailAddress string
	serverURL    string
}

// Order struct
type Order struct {
	ID  string
	URL string
}

func (s *payPalService) CreateOrder(amount, description string) (Order, error) {
	intent := "CAPTURE"
	order, err := s.client.CreateOrder(
		intent,
		[]paypalsdk.PurchaseUnitRequest{
			{
				Amount: &paypalsdk.PurchaseUnitAmount{
					Currency: "BRL",
					Value:    amount,
				},
				Payee: &paypalsdk.PayeeForOrders{
					EmailAddress: s.emailAddress,
				},
				Description: description,
			},
		},
		&paypalsdk.CreateOrderPayer{},
		&paypalsdk.ApplicationContext{
			BrandName: "Cancanvas",
			ReturnURL: s.serverURL + "/return",
			CancelURL: s.serverURL + "/cancel",
		})
	if err != nil {
		return Order{}, err
	}
	return Order{
		ID:  order.ID,
		URL: order.Links[1].Href,
	}, nil
}

func (s *payPalService) GetOrderStatus(orderID string) (string, error) {
	order, err := s.client.GetOrder(orderID)
	if err != nil {
		return "", err
	}
	return order.Status, nil
}

func (s *payPalService) CompleteOrder(orderID string) (string, error) {
	order, err := s.client.CaptureOrder(orderID, paypalsdk.CaptureOrderRequest{})
	if err != nil {
		return "", err
	}
	return order.Status, nil
}

// NewPayPalService function
func NewPayPalService() PayPalService {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	secretID := os.Getenv("PAYPAL_SECRET_ID")
	paypalMode := os.Getenv("PAYPAL_MODE")
	email := os.Getenv("PAYPAL_EMAIL")
	serverURL := os.Getenv("SERVER_URL")
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
		client:       client,
		acessToken:   acessToken,
		emailAddress: email,
		serverURL:    serverURL,
	}
}
