//package middleware
//
//import (
//	"context"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"net/http"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/config"
//	"github.com/gin-gonic/gin"
//	"github.com/golang-jwt/jwt/v4"
//)
//
//// KeycloakMiddleware provides JWT validation for Keycloak tokens
//type KeycloakMiddleware struct {
//	config        config.KeycloakConfig
//	publicKeys    map[string]interface{}
//	publicKeyLock sync.RWMutex
//}
//
//// NewKeycloakMiddleware creates a new Keycloak middleware instance
//func NewKeycloakMiddleware(config config.KeycloakConfig) *KeycloakMiddleware {
//	middleware := &KeycloakMiddleware{
//		config:     config,
//		publicKeys: make(map[string]interface{}),
//	}
//
//	// Fetch public keys on startup
//	if err := middleware.fetchPublicKeys(); err != nil {
//		// Log error but don't fail startup
//		fmt.Printf("Warning: Failed to fetch Keycloak public keys: %v\n", err)
//	}
//
//	return middleware
//}
//
//// AuthRequired is a Gin middleware that validates Keycloak JWT tokens
//func (m *KeycloakMiddleware) AuthRequired(c *gin.Context) {
//	// Get the Authorization header
//	authHeader := c.GetHeader("Authorization")
//	if authHeader == "" {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Authorization header is required",
//		})
//		return
//	}
//
//	// Check if using Bearer authentication
//	parts := strings.Split(authHeader, " ")
//	if len(parts) != 2 || parts[0] != "Bearer" {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Authorization header format must be Bearer {token}",
//		})
//		return
//	}
//
//	tokenString := parts[1]
//
//	// Parse the JWT token without validating the signature yet
//	token, err := jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
//		// Don't verify the signature in this step
//		return nil, nil
//	})
//	if err != nil && !errors.Is(err, jwt.ErrInvalidKeyType) {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Invalid token",
//		})
//		return
//	}
//
//	claims, ok := token.Claims.(*KeycloakClaims)
//	if !ok {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Invalid token claims",
//		})
//		return
//	}
//
//	// Check if the token is expired
//	if claims.ExpiresAt == nil || time.Now().After(time.Unix(claims.ExpiresAt.Unix(), 0)) {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Token is expired",
//		})
//		return
//	}
//
//	// Get the key ID from the token header
//	kid, ok := token.Header["kid"].(string)
//	if !ok {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "No key ID found in token",
//		})
//		return
//	}
//
//	// Get the public key for this kid
//	publicKey, err := m.getPublicKey(kid)
//	if err != nil {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Failed to get public key",
//		})
//		return
//	}
//
//	// Validate the token with the correct public key
//	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
//		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
//			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//		}
//		return publicKey, nil
//	})
//
//	if err != nil {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": fmt.Sprintf("Invalid token signature: %v", err),
//		})
//		return
//	}
//
//	// Validate the audience
//	if !contains(claims.Audience, m.config.ClientID) {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Invalid audience",
//		})
//		return
//	}
//
//	// Validate the issuer
//	expectedIssuer := fmt.Sprintf("%s/realms/%s", m.config.URL, m.config.Realm)
//	if claims.Issuer != expectedIssuer {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
//			"error": "Invalid issuer",
//		})
//		return
//	}
//
//	// Add user info to the context
//	c.Set("userID", claims.Subject)
//	c.Set("username", claims.PreferredUsername)
//	c.Set("email", claims.Email)
//	c.Set("roles", claims.RealmAccess.Roles)
//
//	c.Next()
//}
//
//// KeycloakClaims represents the claims in a Keycloak JWT token
//type KeycloakClaims struct {
//	jwt.RegisteredClaims
//	Name              string `json:"name"`
//	PreferredUsername string `json:"preferred_username"`
//	GivenName         string `json:"given_name"`
//	FamilyName        string `json:"family_name"`
//	Email             string `json:"email"`
//	EmailVerified     bool   `json:"email_verified"`
//	RealmAccess       struct {
//		Roles []string `json:"roles"`
//	} `json:"realm_access"`
//	ResourceAccess map[string]struct {
//		Roles []string `json:"roles"`
//	} `json:"resource_access"`
//}
//
//// fetchPublicKeys fetches the public keys from Keycloak
//func (m *KeycloakMiddleware) fetchPublicKeys() error {
//	// Create a context with timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	// Construct the URL for the Keycloak public keys
//	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", m.config.URL, m.config.Realm)
//
//	// Create a new HTTP request
//	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
//	if err != nil {
//		return fmt.Errorf("error creating request: %w", err)
//	}
//
//	// Send the request
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return fmt.Errorf("error fetching public keys: %w", err)
//	}
//	defer resp.Body.Close()
//
//	// Check the status code
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
//	}
//
//	// Parse the response
//	var jwks struct {
//		Keys []struct {
//			Kid string   `json:"kid"`
//			Kty string   `json:"kty"`
//			Alg string   `json:"alg"`
//			Use string   `json:"use"`
//			N   string   `json:"n"`
//			E   string   `json:"e"`
//			X5c []string `json:"x5c"`
//		} `json:"keys"`
//	}
//
//	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
//		return fmt.Errorf("error decoding response: %w", err)
//	}
//
//	// Store the public keys
//	m.publicKeyLock.Lock()
//	defer m.publicKeyLock.Unlock()
//
//	for _, key := range jwks.Keys {
//		if key.Use == "sig" {
//			// Create a public key from the modulus and exponent
//			publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", key.X5c[0])))
//			if err != nil {
//				return fmt.Errorf("error parsing public key: %w", err)
//			}
//			m.publicKeys[key.Kid] = publicKey
//		}
//	}
//
//	return nil
//}
//
//// getPublicKey gets the public key for the given key ID
//func (m *KeycloakMiddleware) getPublicKey(kid string) (interface{}, error) {
//	m.publicKeyLock.RLock()
//	publicKey, ok := m.publicKeys[kid]
//	m.publicKeyLock.RUnlock()
//
//	if !ok {
//		// Key not found, try to fetch new keys
//		if err := m.fetchPublicKeys(); err != nil {
//			return nil, err
//		}
//
//		// Check if the key is now available
//		m.publicKeyLock.RLock()
//		publicKey, ok = m.publicKeys[kid]
//		m.publicKeyLock.RUnlock()
//
//		if !ok {
//			return nil, errors.New("key not found")
//		}
//	}
//
//	return publicKey, nil
//}
//
//// contains checks if a string is contained in a slice
//func contains(slice []string, item string) bool {
//	for _, s := range slice {
//		if s == item {
//			return true
//		}
//	}
//	return false
//}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware for authentication
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check if using Bearer authentication
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})
			return
		}

		// In a real application, you would validate a JWT token here
		// For simplicity, we're just checking against a secret
		if parts[1] != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authentication token",
			})
			return
		}

		// Authentication successful, proceed
		c.Next()
	}
}
