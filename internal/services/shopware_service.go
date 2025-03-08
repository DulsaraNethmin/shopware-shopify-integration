package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/shopware-shopify-integration/internal/models"
	"gorm.io/gorm"
)

// ShopwareService handles Shopware API operations
type ShopwareService struct {
	db         *gorm.DB
	httpClient *http.Client
}

// NewShopwareService creates a new Shopware service
func NewShopwareService(db *gorm.DB) *ShopwareService {
	return &ShopwareService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProductResponse represents a Shopware product response
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       []Price   `json:"price"`
	Stock       int       `json:"stock"`
	Categories  []string  `json:"categories"`
	Images      []Image   `json:"media"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Price represents a Shopware product price
type Price struct {
	CurrencyID string  `json:"currencyId"`
	Net        float64 `json:"net"`
	Gross      float64 `json:"gross"`
}

// Image represents a Shopware product image
type Image struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Alt  string `json:"alt"`
	Size int    `json:"fileSize"`
}

// OrderResponse represents a Shopware order response
type OrderResponse struct {
	ID              string        `json:"id"`
	OrderNumber     string        `json:"orderNumber"`
	Customer        Customer      `json:"orderCustomer"`
	BillingAddress  Address       `json:"billingAddress"`
	ShippingAddress Address       `json:"shippingAddress"`
	LineItems       []OrderItem   `json:"lineItems"`
	TotalPrice      float64       `json:"amountTotal"`
	TaxStatus       string        `json:"taxStatus"`
	PaymentStatus   PaymentStatus `json:"stateMachineState"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// Customer represents a Shopware customer
type Customer struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Address represents a Shopware address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	PostalCode string `json:"zipcode"`
	Country    string `json:"countryId"`
}

// OrderItem represents a Shopware order item
type OrderItem struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"productId"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unitPrice"`
	TotalPrice float64 `json:"totalPrice"`
}

// PaymentStatus represents a Shopware payment status
type PaymentStatus struct {
	TechnicalName string `json:"technicalName"`
	Name          string `json:"name"`
}

// TestConnection tests the connection to Shopware
func (s *ShopwareService) TestConnection(connector *models.Connector) error {
	url := fmt.Sprintf("%s/api/oauth/token", connector.URL)

	requestBody, err := json.Marshal(map[string]string{
		"client_id":  "administration",
		"grant_type": "password",
		"scopes":     "write",
		"username":   connector.Username,
		"password":   connector.Password,
	})

	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetAccessToken gets an access token from Shopware
func (s *ShopwareService) GetAccessToken(connector *models.Connector) (string, error) {
	url := fmt.Sprintf("%s/api/oauth/token", connector.URL)

	requestBody, err := json.Marshal(map[string]string{
		"client_id":  "administration",
		"grant_type": "password",
		"scopes":     "write",
		"username":   connector.Username,
		"password":   connector.Password,
	})

	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

// GetProduct gets a product from Shopware
func (s *ShopwareService) GetProduct(connector *models.Connector, productID string) (*ProductResponse, error) {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/product/%s", connector.URL, productID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	var product ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &product, nil
}

// GetOrder gets an order from Shopware
func (s *ShopwareService) GetOrder(connector *models.Connector, orderID string) (*OrderResponse, error) {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/order/%s", connector.URL, orderID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	var order OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &order, nil
}

// RegisterWebhooks registers webhooks with Shopware
func (s *ShopwareService) RegisterWebhooks(connector *models.Connector, callbackURL string) error {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		return err
	}

	// Register product webhook
	if err := s.registerWebhook(connector, accessToken, "product.written", callbackURL+"/product"); err != nil {
		return err
	}

	// Register order webhook
	if err := s.registerWebhook(connector, accessToken, "order.placed", callbackURL+"/order"); err != nil {
		return err
	}

	return nil
}

// registerWebhook registers a webhook with Shopware
func (s *ShopwareService) registerWebhook(connector *models.Connector, accessToken, event, url string) error {
	webhookURL := fmt.Sprintf("%s/api/webhook", connector.URL)

	requestBody, err := json.Marshal(map[string]string{
		"name":      fmt.Sprintf("Integration Webhook - %s", event),
		"url":       url,
		"eventName": event,
	})

	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	return nil
}
