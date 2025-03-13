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

// ProductCreateRequest represents a request to create a Shopify product
type ProductCreateRequest struct {
	Product ShopifyProduct `json:"product"`
}

// OrderCreateRequest represents a request to create a Shopify order
type OrderCreateRequest struct {
	Order ShopifyOrder `json:"order"`
}

// ShopifyProduct represents a Shopify product
type ShopifyProduct struct {
	Title            string            `json:"title"`
	BodyHTML         string            `json:"bodyHtml,omitempty"`
	Vendor           string            `json:"vendor,omitempty"`
	ProductType      string            `json:"productType,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Status           string            `json:"status,omitempty"`
	Images           []ShopifyImage    `json:"images,omitempty"`
	Variants         []ShopifyVariant  `json:"variants,omitempty"`
	Options          []ShopifyOption   `json:"options,omitempty"`
	MetafieldsGlobal ShopifyMetafields `json:"metafields,omitempty"`
}

// ShopifyImage represents a Shopify product image
type ShopifyImage struct {
	Src      string `json:"src"`
	Alt      string `json:"altText,omitempty"`
	Position int    `json:"position,omitempty"`
}

// ShopifyVariant represents a Shopify product variant
type ShopifyVariant struct {
	Title               string  `json:"title,omitempty"`
	Price               string  `json:"price"`
	SKU                 string  `json:"sku,omitempty"`
	Position            int     `json:"position,omitempty"`
	InventoryPolicy     string  `json:"inventoryPolicy,omitempty"`
	CompareAtPrice      string  `json:"compareAtPrice,omitempty"`
	FulfillmentService  string  `json:"fulfillmentService,omitempty"`
	InventoryManagement string  `json:"inventoryManagement,omitempty"`
	Taxable             bool    `json:"taxable,omitempty"`
	Barcode             string  `json:"barcode,omitempty"`
	Weight              float64 `json:"weight,omitempty"`
	WeightUnit          string  `json:"weightUnit,omitempty"`
	InventoryQuantity   int     `json:"inventoryQuantity,omitempty"`
	RequiresShipping    bool    `json:"requiresShipping,omitempty"`
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
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Handle    string    `json:"handle"`
		Variants  []struct {
			ID        string `json:"id"`
			ProductID string `json:"productId"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		} `json:"variants"`
	} `json:"product"`
}

// ShopifyOrder represents a Shopify order
type ShopifyOrder struct {
	Email               string            `json:"email,omitempty"`
	Gateway             string            `json:"gateway,omitempty"`
	Test                bool              `json:"test,omitempty"`
	TotalPrice          string            `json:"totalPrice,omitempty"`
	SubtotalPrice       string            `json:"subtotalPrice,omitempty"`
	TotalTax            string            `json:"totalTax,omitempty"`
	Currency            string            `json:"currency,omitempty"`
	FinancialStatus     string            `json:"financialStatus,omitempty"`
	Confirmed           bool              `json:"confirmed,omitempty"`
	TotalDiscounts      string            `json:"totalDiscounts,omitempty"`
	TotalLineItemsPrice string            `json:"totalLineItemsPrice,omitempty"`
	BillingAddress      ShopifyAddress    `json:"billingAddress,omitempty"`
	ShippingAddress     ShopifyAddress    `json:"shippingAddress,omitempty"`
	LineItems           []ShopifyLineItem `json:"lineItems"`
	Customer            ShopifyCustomer   `json:"customer,omitempty"`
	Note                string            `json:"note,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
}

// ShopifyAddress represents a Shopify address
type ShopifyAddress struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Province     string `json:"province,omitempty"`
	Country      string `json:"country"`
	Zip          string `json:"zip"`
	Phone        string `json:"phone,omitempty"`
	ProvinceCode string `json:"provinceCode,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	Default      bool   `json:"default,omitempty"`
}

// ShopifyLineItem represents a Shopify order line item
type ShopifyLineItem struct {
	VariantID  string            `json:"variantId,omitempty"`
	ProductID  string            `json:"productId,omitempty"`
	Title      string            `json:"title"`
	Quantity   int               `json:"quantity"`
	Price      string            `json:"price"`
	Grams      int               `json:"grams,omitempty"`
	SKU        string            `json:"sku,omitempty"`
	Name       string            `json:"name,omitempty"`
	TaxLines   []ShopifyTaxLine  `json:"taxLines,omitempty"`
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
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	Phone            string `json:"phone,omitempty"`
	AcceptsMarketing bool   `json:"acceptsMarketing,omitempty"`
}

// OrderCreateResponse represents a Shopify order create response
type OrderCreateResponse struct {
	Order struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
		TotalPrice      string    `json:"totalPrice"`
		SubtotalPrice   string    `json:"subtotalPrice"`
		TotalTax        string    `json:"totalTax"`
		FinancialStatus string    `json:"financialStatus"`
	} `json:"order"`
}

// GraphQLResponse is a generic GraphQL response structure
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// ProductCreateRequest represents a request to create a Shopify product
//type ProductCreateRequest struct {
//	Product ShopifyProduct `json:"product"`
//}
//
//// OrderCreateRequest represents a request to create a Shopify order
//type OrderCreateRequest struct {
//	Order ShopifyOrder `json:"order"`
//}

// TestConnection tests the connection to Shopify using GraphQL API
func (s *ShopifyService) TestConnection(connector *models.Connector) error {
	// Use the GraphQL API to test the connection by fetching shop information
	query := `{
		shop {
			name
			id
		}
	}`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, query, nil, &response); err != nil {
		return err
	}

	// Check if there are any errors in the response
	if len(response.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	return nil
}

// CreateProduct creates a product in Shopify using GraphQL
func (s *ShopifyService) CreateProduct(connector *models.Connector, productRequest *ProductCreateRequest) (*ProductCreateResponse, error) {
	product := productRequest.Product

	// Prepare variables for the GraphQL mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":           product.Title,
			"descriptionHtml": product.BodyHTML,
			"vendor":          product.Vendor,
			"productType":     product.ProductType,
			"tags":            product.Tags, // This is already a []string
			"status":          product.Status,
		},
	}

	// Create the GraphQL mutation
	mutation := `
		mutation createProduct($input: ProductInput!) {
			productCreate(input: $input) {
				product {
					id
					title
					createdAt
					updatedAt
					handle
					variants(first: 10) {
						edges {
							node {
								id
								title
								price
							}
						}
					}
				}
				userErrors {
					field
					message
				}
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, mutation, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		ProductCreate struct {
			Product struct {
				ID        string    `json:"id"`
				Title     string    `json:"title"`
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
				Handle    string    `json:"handle"`
				Variants  struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Title string `json:"title"`
							Price string `json:"price"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"variants"`
			} `json:"product"`
			UserErrors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"productCreate"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Check for user errors
	if len(result.ProductCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("error creating product: %s", result.ProductCreate.UserErrors[0].Message)
	}

	// Convert the GraphQL response to our expected response format
	productResponse := &ProductCreateResponse{}
	productResponse.Product.ID = result.ProductCreate.Product.ID
	productResponse.Product.Title = result.ProductCreate.Product.Title
	productResponse.Product.CreatedAt = result.ProductCreate.Product.CreatedAt
	productResponse.Product.UpdatedAt = result.ProductCreate.Product.UpdatedAt
	productResponse.Product.Handle = result.ProductCreate.Product.Handle

	// Convert variants
	for _, edge := range result.ProductCreate.Product.Variants.Edges {
		variant := struct {
			ID        string `json:"id"`
			ProductID string `json:"productId"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		}{
			ID:        edge.Node.ID,
			ProductID: result.ProductCreate.Product.ID,
			Title:     edge.Node.Title,
			Price:     edge.Node.Price,
		}
		productResponse.Product.Variants = append(productResponse.Product.Variants, variant)
	}

	return productResponse, nil
}

// UpdateProduct updates a product in Shopify using GraphQL
func (s *ShopifyService) UpdateProduct(connector *models.Connector, productID string, productRequest *ProductCreateRequest) (*ProductCreateResponse, error) {
	product := productRequest.Product

	// Prepare variables for the GraphQL mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":           product.Title,
			"descriptionHtml": product.BodyHTML,
			"vendor":          product.Vendor,
			"productType":     product.ProductType,
			"tags":            product.Tags, // This is already a []string
			"status":          product.Status,
		},
	}

	// Create the GraphQL mutation
	mutation := `
		mutation updateProduct($input: ProductInput!) {
			productUpdate(input: $input) {
				product {
					id
					title
					createdAt
					updatedAt
					handle
					variants(first: 10) {
						edges {
							node {
								id
								title
								price
							}
						}
					}
				}
				userErrors {
					field
					message
				}
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, mutation, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		ProductUpdate struct {
			Product struct {
				ID        string    `json:"id"`
				Title     string    `json:"title"`
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
				Handle    string    `json:"handle"`
				Variants  struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Title string `json:"title"`
							Price string `json:"price"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"variants"`
			} `json:"product"`
			UserErrors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"productUpdate"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Check for user errors
	if len(result.ProductUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("error updating product: %s", result.ProductUpdate.UserErrors[0].Message)
	}

	// Convert the GraphQL response to our expected response format
	productResponse := &ProductCreateResponse{}
	productResponse.Product.ID = result.ProductUpdate.Product.ID
	productResponse.Product.Title = result.ProductUpdate.Product.Title
	productResponse.Product.CreatedAt = result.ProductUpdate.Product.CreatedAt
	productResponse.Product.UpdatedAt = result.ProductUpdate.Product.UpdatedAt
	productResponse.Product.Handle = result.ProductUpdate.Product.Handle

	// Convert variants
	for _, edge := range result.ProductUpdate.Product.Variants.Edges {
		variant := struct {
			ID        string `json:"id"`
			ProductID string `json:"productId"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		}{
			ID:        edge.Node.ID,
			ProductID: result.ProductUpdate.Product.ID,
			Title:     edge.Node.Title,
			Price:     edge.Node.Price,
		}
		productResponse.Product.Variants = append(productResponse.Product.Variants, variant)
	}

	return productResponse, nil
}

// CreateOrder creates an order in Shopify using GraphQL
func (s *ShopifyService) CreateOrder(connector *models.Connector, orderRequest *OrderCreateRequest) (*OrderCreateResponse, error) {
	// Implement GraphQL mutation for order creation
	// This will be a complex mutation as orders have many fields and relationships
	// This is a simplified example - you'll need to adapt it for your specific requirements

	order := orderRequest.Order

	// Prepare line items
	lineItems := make([]map[string]interface{}, len(order.LineItems))
	for i, item := range order.LineItems {
		lineItem := map[string]interface{}{
			"title":    item.Title,
			"quantity": item.Quantity,
			"price":    item.Price,
		}
		if item.VariantID != "" {
			lineItem["variantId"] = item.VariantID
		}
		lineItems[i] = lineItem
	}

	// Prepare variables for the GraphQL mutation
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"email":       order.Email,
			"lineItems":   lineItems,
			"note":        order.Note,
			"tags":        order.Tags,
			"customerId":  nil, // You'd need to fetch or create a customer ID
			"processedAt": time.Now().Format(time.RFC3339),
		},
	}

	if order.BillingAddress.FirstName != "" {
		variables["input"].(map[string]interface{})["billingAddress"] = map[string]interface{}{
			"firstName": order.BillingAddress.FirstName,
			"lastName":  order.BillingAddress.LastName,
			"address1":  order.BillingAddress.Address1,
			"address2":  order.BillingAddress.Address2,
			"city":      order.BillingAddress.City,
			"province":  order.BillingAddress.Province,
			"country":   order.BillingAddress.Country,
			"zip":       order.BillingAddress.Zip,
			"phone":     order.BillingAddress.Phone,
		}
	}

	if order.ShippingAddress.FirstName != "" {
		variables["input"].(map[string]interface{})["shippingAddress"] = map[string]interface{}{
			"firstName": order.ShippingAddress.FirstName,
			"lastName":  order.ShippingAddress.LastName,
			"address1":  order.ShippingAddress.Address1,
			"address2":  order.ShippingAddress.Address2,
			"city":      order.ShippingAddress.City,
			"province":  order.ShippingAddress.Province,
			"country":   order.ShippingAddress.Country,
			"zip":       order.ShippingAddress.Zip,
			"phone":     order.ShippingAddress.Phone,
		}
	}

	// Create the GraphQL mutation
	mutation := `
		mutation createOrder($input: OrderInput!) {
			orderCreate(input: $input) {
				order {
					id
					name
					email
					createdAt
					updatedAt
					totalPrice
					subtotalPrice
					totalTax
					displayFinancialStatus
				}
				userErrors {
					field
					message
				}
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, mutation, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		OrderCreate struct {
			Order struct {
				ID                     string    `json:"id"`
				Name                   string    `json:"name"`
				Email                  string    `json:"email"`
				CreatedAt              time.Time `json:"createdAt"`
				UpdatedAt              time.Time `json:"updatedAt"`
				TotalPrice             string    `json:"totalPrice"`
				SubtotalPrice          string    `json:"subtotalPrice"`
				TotalTax               string    `json:"totalTax"`
				DisplayFinancialStatus string    `json:"displayFinancialStatus"`
			} `json:"order"`
			UserErrors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"orderCreate"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Check for user errors
	if len(result.OrderCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("error creating order: %s", result.OrderCreate.UserErrors[0].Message)
	}

	// Convert the GraphQL response to our expected response format
	orderResponse := &OrderCreateResponse{}
	orderResponse.Order.ID = result.OrderCreate.Order.ID
	orderResponse.Order.Name = result.OrderCreate.Order.Name
	orderResponse.Order.Email = result.OrderCreate.Order.Email
	orderResponse.Order.CreatedAt = result.OrderCreate.Order.CreatedAt
	orderResponse.Order.UpdatedAt = result.OrderCreate.Order.UpdatedAt
	orderResponse.Order.TotalPrice = result.OrderCreate.Order.TotalPrice
	orderResponse.Order.SubtotalPrice = result.OrderCreate.Order.SubtotalPrice
	orderResponse.Order.TotalTax = result.OrderCreate.Order.TotalTax
	orderResponse.Order.FinancialStatus = result.OrderCreate.Order.DisplayFinancialStatus

	return orderResponse, nil
}

// RegisterWebhooks registers webhooks with Shopify using GraphQL
func (s *ShopifyService) RegisterWebhooks(connector *models.Connector, callbackURL string) error {
	// Shopify webhooks are not needed for this integration as it's one-way
	// from Shopware to Shopify, but we'll implement the method for completeness

	// Note: In GraphQL, you would use the webhookSubscriptionCreate mutation
	return nil
}

// GetProductByID gets a product from Shopify by ID using GraphQL
func (s *ShopifyService) GetProductByID(connector *models.Connector, productID string) (*ProductCreateResponse, error) {
	// Create GraphQL query
	variables := map[string]interface{}{
		"id": productID,
	}

	query := `
		query getProduct($id: ID!) {
			product(id: $id) {
				id
				title
				createdAt
				updatedAt
				handle
				variants(first: 10) {
					edges {
						node {
							id
							title
							price
						}
					}
				}
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, query, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		Product struct {
			ID        string    `json:"id"`
			Title     string    `json:"title"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
			Handle    string    `json:"handle"`
			Variants  struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Title string `json:"title"`
						Price string `json:"price"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"variants"`
		} `json:"product"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Convert the GraphQL response to our expected response format
	productResponse := &ProductCreateResponse{}
	productResponse.Product.ID = result.Product.ID
	productResponse.Product.Title = result.Product.Title
	productResponse.Product.CreatedAt = result.Product.CreatedAt
	productResponse.Product.UpdatedAt = result.Product.UpdatedAt
	productResponse.Product.Handle = result.Product.Handle

	// Convert variants
	for _, edge := range result.Product.Variants.Edges {
		variant := struct {
			ID        string `json:"id"`
			ProductID string `json:"productId"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		}{
			ID:        edge.Node.ID,
			ProductID: result.Product.ID,
			Title:     edge.Node.Title,
			Price:     edge.Node.Price,
		}
		productResponse.Product.Variants = append(productResponse.Product.Variants, variant)
	}

	return productResponse, nil
}

// GetOrderByID gets an order from Shopify by ID using GraphQL
func (s *ShopifyService) GetOrderByID(connector *models.Connector, orderID string) (*OrderCreateResponse, error) {
	// Create GraphQL query
	variables := map[string]interface{}{
		"id": orderID,
	}

	query := `
		query getOrder($id: ID!) {
			order(id: $id) {
				id
				name
				email
				createdAt
				updatedAt
				totalPrice
				subtotalPrice
				totalTax
				displayFinancialStatus
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, query, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		Order struct {
			ID                     string    `json:"id"`
			Name                   string    `json:"name"`
			Email                  string    `json:"email"`
			CreatedAt              time.Time `json:"createdAt"`
			UpdatedAt              time.Time `json:"updatedAt"`
			TotalPrice             string    `json:"totalPrice"`
			SubtotalPrice          string    `json:"subtotalPrice"`
			TotalTax               string    `json:"totalTax"`
			DisplayFinancialStatus string    `json:"displayFinancialStatus"`
		} `json:"order"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Convert the GraphQL response to our expected response format
	orderResponse := &OrderCreateResponse{}
	orderResponse.Order.ID = result.Order.ID
	orderResponse.Order.Name = result.Order.Name
	orderResponse.Order.Email = result.Order.Email
	orderResponse.Order.CreatedAt = result.Order.CreatedAt
	orderResponse.Order.UpdatedAt = result.Order.UpdatedAt
	orderResponse.Order.TotalPrice = result.Order.TotalPrice
	orderResponse.Order.SubtotalPrice = result.Order.SubtotalPrice
	orderResponse.Order.TotalTax = result.Order.TotalTax
	orderResponse.Order.FinancialStatus = result.Order.DisplayFinancialStatus

	return orderResponse, nil
}

// FindProductBySKU finds a product in Shopify by SKU using GraphQL
func (s *ShopifyService) FindProductBySKU(connector *models.Connector, sku string) (*ProductCreateResponse, error) {
	// GraphQL query to search for a product variant by SKU
	variables := map[string]interface{}{
		"query": fmt.Sprintf("sku:%s", sku),
	}

	query := `
		query findProductBySKU($query: String!) {
			productVariants(first: 1, query: $query) {
				edges {
					node {
						id
						sku
						product {
							id
							title
							createdAt
							updatedAt
							handle
							variants(first: 10) {
								edges {
									node {
										id
										title
										price
									}
								}
							}
						}
					}
				}
			}
		}
	`

	var response GraphQLResponse
	if err := s.executeGraphQL(connector, query, variables, &response); err != nil {
		return nil, err
	}

	// Unmarshal the GraphQL response
	var result struct {
		ProductVariants struct {
			Edges []struct {
				Node struct {
					ID      string `json:"id"`
					SKU     string `json:"sku"`
					Product struct {
						ID        string    `json:"id"`
						Title     string    `json:"title"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
						Handle    string    `json:"handle"`
						Variants  struct {
							Edges []struct {
								Node struct {
									ID    string `json:"id"`
									Title string `json:"title"`
									Price string `json:"price"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"variants"`
					} `json:"product"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"productVariants"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("error parsing GraphQL response: %w", err)
	}

	// Check if any variants were found
	if len(result.ProductVariants.Edges) == 0 {
		return nil, fmt.Errorf("no product found with SKU: %s", sku)
	}

	// Extract the product from the first variant
	product := result.ProductVariants.Edges[0].Node.Product

	// Convert the GraphQL response to our expected response format
	productResponse := &ProductCreateResponse{}
	productResponse.Product.ID = product.ID
	productResponse.Product.Title = product.Title
	productResponse.Product.CreatedAt = product.CreatedAt
	productResponse.Product.UpdatedAt = product.UpdatedAt
	productResponse.Product.Handle = product.Handle

	// Convert variants
	for _, edge := range product.Variants.Edges {
		variant := struct {
			ID        string `json:"id"`
			ProductID string `json:"productId"`
			Title     string `json:"title"`
			Price     string `json:"price"`
		}{
			ID:        edge.Node.ID,
			ProductID: product.ID,
			Title:     edge.Node.Title,
			Price:     edge.Node.Price,
		}
		productResponse.Product.Variants = append(productResponse.Product.Variants, variant)
	}

	return productResponse, nil
}

// executeGraphQL is a helper method to execute GraphQL queries and mutations
func (s *ShopifyService) executeGraphQL(connector *models.Connector, query string, variables map[string]interface{}, response interface{}) error {
	// Prepare the request body
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling GraphQL request: %w", err)
	}

	// Create the GraphQL endpoint URL
	url := fmt.Sprintf("https://%s/admin/api/2025-04/graphql.json", connector.URL)

	// Create the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating GraphQL request: %w", err)
	}

	// Set headers
	req.Header.Set("X-Shopify-Access-Token", connector.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading GraphQL response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Unmarshal the response
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("error unmarshaling GraphQL response: %w", err)
	}

	return nil
}
