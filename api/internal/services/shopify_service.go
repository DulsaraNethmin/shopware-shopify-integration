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

// ShopifyService handles Shopify API operations
type ShopifyService struct {
	db         *gorm.DB
	httpClient *http.Client
}

// NewShopifyService creates a new Shopify service
func NewShopifyService(db *gorm.DB) *ShopifyService {
	return &ShopifyService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProductCreateRequest represents a Shopify product create request
type ProductCreateRequest struct {
	Product ShopifyProduct `json:"product"`
}

// ShopifyProduct represents a Shopify product
type ShopifyProduct struct {
	Title            string            `json:"title"`
	BodyHTML         string            `json:"body_html"`
	Vendor           string            `json:"vendor,omitempty"`
	ProductType      string            `json:"product_type,omitempty"`
	Tags             string            `json:"tags,omitempty"`
	Status           string            `json:"status,omitempty"`
	Images           []ShopifyImage    `json:"images,omitempty"`
	Variants         []ShopifyVariant  `json:"variants,omitempty"`
	Options          []ShopifyOption   `json:"options,omitempty"`
	MetafieldsGlobal ShopifyMetafields `json:"metafields_global,omitempty"`
}

// ShopifyImage represents a Shopify product image
type ShopifyImage struct {
	Src      string `json:"src"`
	Alt      string `json:"alt,omitempty"`
	Position int    `json:"position,omitempty"`
}

// ShopifyVariant represents a Shopify product variant
type ShopifyVariant struct {
	Title               string  `json:"title,omitempty"`
	Price               string  `json:"price"`
	SKU                 string  `json:"sku,omitempty"`
	Position            int     `json:"position,omitempty"`
	InventoryPolicy     string  `json:"inventory_policy,omitempty"`
	CompareAtPrice      string  `json:"compare_at_price,omitempty"`
	FulfillmentService  string  `json:"fulfillment_service,omitempty"`
	InventoryManagement string  `json:"inventory_management,omitempty"`
	Taxable             bool    `json:"taxable,omitempty"`
	Barcode             string  `json:"barcode,omitempty"`
	Weight              float64 `json:"weight,omitempty"`
	WeightUnit          string  `json:"weight_unit,omitempty"`
	InventoryQuantity   int     `json:"inventory_quantity,omitempty"`
	RequiresShipping    bool    `json:"requires_shipping,omitempty"`
}

// ShopifyOption represents a Shopify product option
type ShopifyOption struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// ShopifyMetafields represents Shopify product metafields
type ShopifyMetafields struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProductCreateResponse represents a Shopify product create response
type ProductCreateResponse struct {
	Product struct {
		ID        int64     `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Handle    string    `json:"handle"`
		Variants  []struct {
			ID        int64  `json:"id"`
			ProductID int64  `json:"product_id"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		} `json:"variants"`
	} `json:"product"`
}

// OrderCreateRequest represents a Shopify order create request
type OrderCreateRequest struct {
	Order ShopifyOrder `json:"order"`
}

// ShopifyOrder represents a Shopify order
type ShopifyOrder struct {
	Email               string            `json:"email,omitempty"`
	Gateway             string            `json:"gateway,omitempty"`
	Test                bool              `json:"test,omitempty"`
	TotalPrice          string            `json:"total_price,omitempty"`
	SubtotalPrice       string            `json:"subtotal_price,omitempty"`
	TotalTax            string            `json:"total_tax,omitempty"`
	Currency            string            `json:"currency,omitempty"`
	FinancialStatus     string            `json:"financial_status,omitempty"`
	Confirmed           bool              `json:"confirmed,omitempty"`
	TotalDiscounts      string            `json:"total_discounts,omitempty"`
	TotalLineItemsPrice string            `json:"total_line_items_price,omitempty"`
	BillingAddress      ShopifyAddress    `json:"billing_address,omitempty"`
	ShippingAddress     ShopifyAddress    `json:"shipping_address,omitempty"`
	LineItems           []ShopifyLineItem `json:"line_items"`
	Customer            ShopifyCustomer   `json:"customer,omitempty"`
	Note                string            `json:"note,omitempty"`
	Tags                string            `json:"tags,omitempty"`
}

// ShopifyAddress represents a Shopify address
type ShopifyAddress struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Province     string `json:"province,omitempty"`
	Country      string `json:"country"`
	Zip          string `json:"zip"`
	Phone        string `json:"phone,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	Default      bool   `json:"default,omitempty"`
}

// ShopifyLineItem represents a Shopify order line item
type ShopifyLineItem struct {
	VariantID  int64             `json:"variant_id,omitempty"`
	ProductID  int64             `json:"product_id,omitempty"`
	Title      string            `json:"title"`
	Quantity   int               `json:"quantity"`
	Price      string            `json:"price"`
	Grams      int               `json:"grams,omitempty"`
	SKU        string            `json:"sku,omitempty"`
	Name       string            `json:"name,omitempty"`
	TaxLines   []ShopifyTaxLine  `json:"tax_lines,omitempty"`
	Properties []ShopifyProperty `json:"properties,omitempty"`
}

// ShopifyTaxLine represents a Shopify tax line
type ShopifyTaxLine struct {
	Title string `json:"title"`
	Price string `json:"price"`
	Rate  string `json:"rate"`
}

// ShopifyProperty represents a Shopify line item property
type ShopifyProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ShopifyCustomer represents a Shopify customer
type ShopifyCustomer struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Email            string `json:"email"`
	Phone            string `json:"phone,omitempty"`
	AcceptsMarketing bool   `json:"accepts_marketing,omitempty"`
}

// OrderCreateResponse represents a Shopify order create response
type OrderCreateResponse struct {
	Order struct {
		ID              int64     `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		TotalPrice      string    `json:"total_price"`
		SubtotalPrice   string    `json:"subtotal_price"`
		TotalTax        string    `json:"total_tax"`
		FinancialStatus string    `json:"financial_status"`
	} `json:"order"`
}

// TestConnection tests the connection to Shopify
func (s *ShopifyService) TestConnection(connector *models.Connector) error {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/shop.json", connector.URL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	return nil
}

// CreateProduct creates a product in Shopify
func (s *ShopifyService) CreateProduct(connector *models.Connector, product *ProductCreateRequest) (*ProductCreateResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/products.json", connector.URL)

	requestBody, err := json.Marshal(product)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var response ProductCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// UpdateProduct updates a product in Shopify
func (s *ShopifyService) UpdateProduct(connector *models.Connector, productID int64, product *ProductCreateRequest) (*ProductCreateResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/products/%d.json", connector.URL, productID)

	requestBody, err := json.Marshal(product)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var response ProductCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// CreateOrder creates an order in Shopify
func (s *ShopifyService) CreateOrder(connector *models.Connector, order *OrderCreateRequest) (*OrderCreateResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/orders.json", connector.URL)

	requestBody, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var response OrderCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// RegisterWebhooks registers webhooks with Shopify
func (s *ShopifyService) RegisterWebhooks(connector *models.Connector, callbackURL string) error {
	// Shopify webhooks are not needed for this integration as it's one-way
	// from Shopware to Shopify, but we'll implement the method for completeness
	return nil
}

// GetProductByID gets a product from Shopify by ID
func (s *ShopifyService) GetProductByID(connector *models.Connector, productID int64) (*ProductCreateResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/products/%d.json", connector.URL, productID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var response ProductCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// GetOrderByID gets an order from Shopify by ID
func (s *ShopifyService) GetOrderByID(connector *models.Connector, orderID int64) (*OrderCreateResponse, error) {
	url := fmt.Sprintf("https://%s/admin/api/2023-07/orders/%d.json", connector.URL, orderID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var response OrderCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// FindProductBySKU finds a product in Shopify by SKU
func (s *ShopifyService) FindProductBySKU(connector *models.Connector, sku string) (*ProductCreateResponse, error) {
	// Shopify doesn't have a direct endpoint to search by SKU, so we need to search for variants
	url := fmt.Sprintf("https://%s/admin/api/2023-07/variants.json?sku=%s", connector.URL, sku)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from Shopify: %s - %s", resp.Status, string(body))
	}

	var variantsResponse struct {
		Variants []struct {
			ID        int64  `json:"id"`
			ProductID int64  `json:"product_id"`
			SKU       string `json:"sku"`
		} `json:"variants"`
	}

	if err := json.Unmarshal(body, &variantsResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(variantsResponse.Variants) == 0 {
		return nil, fmt.Errorf("no product found with SKU: %s", sku)
	}

	// Get the product details using the first matching variant's product ID
	productID := variantsResponse.Variants[0].ProductID
	return s.GetProductByID(connector, productID)
}
