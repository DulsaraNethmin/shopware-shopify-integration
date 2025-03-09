package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductsHandler handles product-related API requests
type ProductsHandler struct {
	connectorService *services.ConnectorService
	shopwareService  *services.ShopwareService
}

// NewProductsHandler creates a new products handler
func NewProductsHandler(connectorService *services.ConnectorService, shopwareService *services.ShopwareService) *ProductsHandler {
	return &ProductsHandler{
		connectorService: connectorService,
		shopwareService:  shopwareService,
	}
}

// GetAllProducts returns all products from the Shopware connector
func (h *ProductsHandler) GetAllProducts(c *gin.Context) {
	// Parse connector ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	// Get the connector from the service
	connector, err := h.connectorService.GetConnector(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check that it's a Shopware connector
	if connector.Type != models.ConnectorTypeShopware {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Connector is not a Shopware connector",
		})
		return
	}

	// Get products from Shopware
	products, err := h.shopwareService.GetAllProducts(connector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get products: " + err.Error(),
		})
		return
	}

	// Return the products
	c.JSON(http.StatusOK, gin.H{
		"message": "Products retrieved successfully",
		"data": gin.H{
			"connector": connector.Name,
			"products":  products,
			"count":     len(products),
		},
	})
}

// GetProduct gets a specific product from the Shopware connector
func (h *ProductsHandler) GetProduct(c *gin.Context) {
	// Parse connector ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	// Get product ID from URL
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product ID is required",
		})
		return
	}

	// Get the connector from the service
	connector, err := h.connectorService.GetConnector(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check that it's a Shopware connector
	if connector.Type != models.ConnectorTypeShopware {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Connector is not a Shopware connector",
		})
		return
	}

	// Get the product from Shopware
	product, err := h.shopwareService.GetProduct(connector, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get product: " + err.Error(),
		})
		return
	}

	// Return the product
	c.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"data":    product,
	})
}

// RegisterRoutes registers the product routes
func (h *ProductsHandler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/connectors/:id/products")
	{
		products.GET("", h.GetAllProducts)
		products.GET("/:productId", h.GetProduct)
	}
}
