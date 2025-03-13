package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
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
//type ProductResponse struct {
//	ID          string    `json:"id"`
//	Name        string    `json:"name"`
//	Description string    `json:"description"`
//	Price       []Price   `json:"price"`
//	Stock       int       `json:"stock"`
//	Categories  []string  `json:"categories"`
//	Images      []Image   `json:"media"`
//	CreatedAt   time.Time `json:"createdAt"`
//	UpdatedAt   time.Time `json:"updatedAt"`
//}

type ShopwareResponse struct {
	Data ProductResponse `json:"data"`
}

// ProductResponse represents a Shopware product response
type ProductResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          []Price   `json:"price"`
	Stock          int       `json:"stock"`
	AvailableStock int       `json:"availableStock"`
	ProductNumber  string    `json:"productNumber"`
	Categories     []string  `json:"categoryIds"`
	Media          []Image   `json:"media"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Translated     struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"translated"`
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

	fmt.Printf("URL: %s", url)

	fmt.Printf("API Key: %s", connector.ApiKey)
	fmt.Printf("API Secret: %s", connector.ApiSecret)

	requestBody, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"scopes":        "write",
		"client_id":     connector.ApiKey,
		"client_secret": connector.ApiSecret,
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
		"grant_type":    "client_credentials",
		"scopes":        "write",
		"client_id":     connector.ApiKey,
		"client_secret": connector.ApiSecret,
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
//func (s *ShopwareService) GetProduct(connector *models.Connector, productID string) (*ProductResponse, error) {
//	accessToken, err := s.GetAccessToken(connector)
//	if err != nil {
//		return nil, err
//	}
//
//	url := fmt.Sprintf("%s/api/product/%s", connector.URL, productID)
//
//	req, err := http.NewRequest(http.MethodGet, url, nil)
//	if err != nil {
//		return nil, fmt.Errorf("error creating request: %w", err)
//	}
//
//	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
//	req.Header.Set("Accept", "application/json")
//
//	resp, err := s.httpClient.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("error making request: %w", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		body, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
//	}
//
//	var product ProductResponse
//	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
//		return nil, fmt.Errorf("error decoding response: %w", err)
//	}
//
//	return &product, nil
//}

// GetProduct gets a product from Shopware
//func (s *ShopwareService) GetProduct(connector *models.Connector, productID string) (*ProductResponse, error) {
//	accessToken, err := s.GetAccessToken(connector)
//	if err != nil {
//		fmt.Printf("Failed to get access token: %v\n", err)
//		return nil, err
//	}
//
//	fmt.Printf("Using access token: %s\n", accessToken)
//	fmt.Printf("Getting product with ID: %s\n", productID)
//
//	// Make sure we're using the correct API endpoint format
//	url := fmt.Sprintf("%s/api/product/%s", connector.URL, productID)
//	fmt.Printf("API URL: %s\n", url)
//
//	req, err := http.NewRequest(http.MethodGet, url, nil)
//	if err != nil {
//		fmt.Printf("Error creating request: %v\n", err)
//		return nil, fmt.Errorf("error creating request: %w", err)
//	}
//
//	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
//	req.Header.Set("Accept", "application/json")
//
//	resp, err := s.httpClient.Do(req)
//	if err != nil {
//		fmt.Printf("Error making request: %v\n", err)
//		return nil, fmt.Errorf("error making request: %w", err)
//	}
//	defer resp.Body.Close()
//
//	// Read the full response body for logging
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		fmt.Printf("Error reading response body: %v\n", err)
//		return nil, fmt.Errorf("error reading response body: %w", err)
//	}
//
//	fmt.Printf("API response status: %d\n", resp.StatusCode)
//	fmt.Printf("API response body: %s\n", string(body))
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
//	}
//
//	// Create a new reader from the body bytes for JSON decoding
//	bodyReader := bytes.NewReader(body)
//
//	var product ProductResponse
//	if err := json.NewDecoder(bodyReader).Decode(&product); err != nil {
//		fmt.Printf("Error decoding response: %v\n", err)
//		return nil, fmt.Errorf("error decoding response: %w", err)
//	}
//
//	fmt.Printf("Decoded product: %+v\n", product)
//
//	return &product, nil
//}

// GetProduct gets a product from Shopware
func (s *ShopwareService) GetProduct(connector *models.Connector, productID string) (*ProductResponse, error) {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		fmt.Printf("Failed to get access token: %v\n", err)
		return nil, err
	}

	fmt.Printf("Using access token: %s\n", accessToken)
	fmt.Printf("Getting product with ID: %s\n", productID)

	url := fmt.Sprintf("%s/api/product/%s", connector.URL, productID)
	fmt.Printf("API URL: %s\n", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	fmt.Printf("API response status: %d\n", resp.StatusCode)
	fmt.Printf("API response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	// Parse the nested response
	var response ShopwareResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Extract the product data from the nested structure
	product := response.Data

	// If name is empty in the main object but exists in 'translated', use that
	if product.Name == "" && product.Translated.Name != "" {
		product.Name = product.Translated.Name
	}

	// If description is empty in the main object but exists in 'translated', use that
	if product.Description == "" && product.Translated.Description != "" {
		product.Description = product.Translated.Description
	}

	fmt.Printf("Decoded product: %+v\n", product)

	return &product, nil
}

// Get All Products
func (s *ShopwareService) GetAllProducts(connector *models.Connector) ([]ProductResponse, error) {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/api/product", connector.URL)
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
	var products []ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return products, nil
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
//func (s *ShopwareService) RegisterWebhooks(connector *models.Connector, callbackURL string) error {
//	accessToken, err := s.GetAccessToken(connector)
//	fmt.Printf("Access token: %s\n", accessToken)
//	fmt.Printf("accesstoken: %s", accessToken)
//	if err != nil {
//		return err
//	}
//
//	// Register product webhook
//	if err := s.registerWebhook(connector, accessToken, "product.written", callbackURL+"/product"); err != nil {
//		return err
//	}
//
//	// Register order webhook
//	if err := s.registerWebhook(connector, accessToken, "order.placed", callbackURL+"/order"); err != nil {
//		return err
//	}
//
//	return nil
//}

// registerWebhook registers a webhook with Shopware

// RegisterWebhooks registers webhooks with Shopware
func (s *ShopwareService) RegisterWebhooks(connector *models.Connector, callbackURL string) error {
	accessToken, err := s.GetAccessToken(connector)
	fmt.Printf("access token: %s\n", accessToken)
	if err != nil {
		return err
	}

	// Register product webhook
	if err := s.registerWebhook(connector, accessToken, "product.written", callbackURL); err != nil {
		return err
	}

	// Register order webhook
	if err := s.registerWebhook(connector, accessToken, "order.placed", callbackURL); err != nil {
		return err
	}

	return nil
}

func (s *ShopwareService) registerWebhook(connector *models.Connector, accessToken, event, url string) error {
	webhookURL := fmt.Sprintf("%s/api/webhook", connector.URL)

	requestBody, err := json.Marshal(map[string]string{
		"name":      fmt.Sprintf("Integration Webhook - %s", event),
		"url":       url,
		"eventName": event,
	})

	fmt.Printf("body: %s", requestBody)
	println(webhookURL)

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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from Shopware: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetWebhooks retrieves all webhooks registered with Shopware
// GetWebhooks retrieves all webhooks registered with Shopware
func (s *ShopwareService) GetWebhooks(connector *models.Connector) ([]map[string]interface{}, error) {
	accessToken, err := s.GetAccessToken(connector)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/webhook", connector.URL)

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

	// Read the response body to inspect the structure
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// First try parsing as a response wrapper object
	var responseWrapper struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &responseWrapper); err == nil && len(responseWrapper.Data) > 0 {
		return responseWrapper.Data, nil
	}

	// If that fails, try parsing as a direct array
	var webhooksArray []map[string]interface{}
	if err := json.Unmarshal(body, &webhooksArray); err == nil {
		return webhooksArray, nil
	}

	// If both fail, try parsing as a single object
	var webhookObject map[string]interface{}
	if err := json.Unmarshal(body, &webhookObject); err == nil {
		return []map[string]interface{}{webhookObject}, nil
	}

	// If all parsing attempts fail, return error with response content
	return nil, fmt.Errorf("unable to parse webhook response: %s", string(body))
}
