package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"gorm.io/gorm"
)

// FieldMappingService handles field mapping operations
type FieldMappingService struct {
	db *gorm.DB
}

// NewFieldMappingService creates a new field mapping service
func NewFieldMappingService(db *gorm.DB) *FieldMappingService {
	return &FieldMappingService{
		db: db,
	}
}

// MappingResult contains the transformed data and any errors
type MappingResult struct {
	Data  map[string]interface{}
	Error error
}

// CreateFieldMapping creates a new field mapping
func (s *FieldMappingService) CreateFieldMapping(fieldMapping *models.FieldMapping) error {
	return s.db.Create(fieldMapping).Error
}

// GetFieldMapping gets a field mapping by ID
func (s *FieldMappingService) GetFieldMapping(id uint) (*models.FieldMapping, error) {
	var fieldMapping models.FieldMapping

	if err := s.db.First(&fieldMapping, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &fieldMapping, nil
}

// ListFieldMappings lists all field mappings for a dataflow
func (s *FieldMappingService) ListFieldMappings(dataflowID uint) ([]models.FieldMapping, error) {
	var fieldMappings []models.FieldMapping

	if err := s.db.Where("dataflow_id = ?", dataflowID).Find(&fieldMappings).Error; err != nil {
		return nil, err
	}

	return fieldMappings, nil
}

// UpdateFieldMapping updates a field mapping
func (s *FieldMappingService) UpdateFieldMapping(id uint, fieldMapping *models.FieldMapping) error {
	// Check if the field mapping exists
	existingFieldMapping, err := s.GetFieldMapping(id)
	if err != nil {
		return err
	}

	// Update the field mapping
	fieldMapping.ID = existingFieldMapping.ID
	return s.db.Save(fieldMapping).Error
}

// DeleteFieldMapping deletes a field mapping
func (s *FieldMappingService) DeleteFieldMapping(id uint) error {
	// Check if the field mapping exists
	existingFieldMapping, err := s.GetFieldMapping(id)
	if err != nil {
		return err
	}

	// Delete the field mapping
	return s.db.Delete(existingFieldMapping).Error
}

// TransformData transforms data based on field mappings
func (s *FieldMappingService) TransformData(dataflowID uint, sourceData []byte) (*MappingResult, error) {
	// Get field mappings for the dataflow
	fieldMappings, err := s.ListFieldMappings(dataflowID)
	if err != nil {
		return nil, fmt.Errorf("error getting field mappings: %w", err)
	}

	// Parse source data
	var sourceObj map[string]interface{}
	if err := json.Unmarshal(sourceData, &sourceObj); err != nil {
		return nil, fmt.Errorf("error parsing source data: %w", err)
	}

	// Create destination object
	destObj := make(map[string]interface{})

	// Apply field mappings
	for _, mapping := range fieldMappings {
		// Get source value using dot notation (supports nested fields)
		sourceValue, err := getNestedValue(sourceObj, mapping.SourceField)
		if err != nil {
			if mapping.IsRequired {
				return &MappingResult{Error: fmt.Errorf("required field %s not found in source data", mapping.SourceField)}, nil
			}
			// Use default value if provided
			if mapping.DefaultValue != "" {
				sourceValue = mapping.DefaultValue
			} else {
				// Skip this field
				continue
			}
		}

		// Apply transformation if needed
		transformedValue, err := s.applyTransformation(sourceValue, mapping)
		if err != nil {
			return &MappingResult{Error: fmt.Errorf("error transforming field %s: %w", mapping.SourceField, err)}, nil
		}

		// Set destination value (supports nested fields)
		if err := setNestedValue(destObj, mapping.DestField, transformedValue); err != nil {
			return &MappingResult{Error: fmt.Errorf("error setting field %s: %w", mapping.DestField, err)}, nil
		}
	}

	return &MappingResult{Data: destObj}, nil
}

// applyTransformation applies a transformation to a value
func (s *FieldMappingService) applyTransformation(value interface{}, mapping models.FieldMapping) (interface{}, error) {
	switch mapping.TransformType {
	case models.TransformationTypeNone:
		return value, nil

	case models.TransformationTypeFormat:
		// For date format transformations
		var config struct {
			SourceFormat string `json:"source_format"`
			DestFormat   string `json:"dest_format"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		if str, ok := value.(string); ok {
			t, err := time.Parse(config.SourceFormat, str)
			if err != nil {
				return nil, fmt.Errorf("error parsing date: %w", err)
			}
			return t.Format(config.DestFormat), nil
		}
		return nil, fmt.Errorf("value is not a string")

	case models.TransformationTypeConvert:
		// For type conversions
		var config struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		switch config.Type {
		case "string":
			return fmt.Sprintf("%v", value), nil
		case "int":
			if str, ok := value.(string); ok {
				i, err := strconv.Atoi(str)
				if err != nil {
					return nil, fmt.Errorf("error converting to int: %w", err)
				}
				return i, nil
			}
			return nil, fmt.Errorf("value is not a string")
		case "float":
			if str, ok := value.(string); ok {
				f, err := strconv.ParseFloat(str, 64)
				if err != nil {
					return nil, fmt.Errorf("error converting to float: %w", err)
				}
				return f, nil
			}
			return nil, fmt.Errorf("value is not a string")
		case "bool":
			if str, ok := value.(string); ok {
				b, err := strconv.ParseBool(str)
				if err != nil {
					return nil, fmt.Errorf("error converting to bool: %w", err)
				}
				return b, nil
			}
			return nil, fmt.Errorf("value is not a string")
		default:
			return nil, fmt.Errorf("unsupported conversion type: %s", config.Type)
		}

	case models.TransformationTypeMap:
		// For value mappings
		var config map[string]interface{}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		strValue := fmt.Sprintf("%v", value)
		if mappedValue, ok := config[strValue]; ok {
			return mappedValue, nil
		}

		// Check if there's a default value in the mapping
		if defaultValue, ok := config["_default"]; ok {
			return defaultValue, nil
		}

		return nil, fmt.Errorf("no mapping found for value: %v", value)

	case models.TransformationTypeTemplate:
		// For template-based transformations
		var config struct {
			Template string `json:"template"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		// Simple template replacement
		result := config.Template
		result = strings.ReplaceAll(result, "{{value}}", fmt.Sprintf("%v", value))
		return result, nil

	case models.TransformationTypeGraphQLID:
		var config struct {
			ResourceType string `json:"resource_type"`
			Direction    string `json:"direction"` // "to_global" or "from_global"
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		if str, ok := value.(string); ok {
			if config.Direction == "to_global" {
				return s.convertToGraphQLGlobalID(config.ResourceType, str), nil
			} else if config.Direction == "from_global" {
				return s.convertFromGraphQLGlobalID(str), nil
			}
			return nil, fmt.Errorf("invalid direction: %s", config.Direction)
		}
		return nil, fmt.Errorf("value is not a string")

	case models.TransformationTypeArrayMap:
		// Handle array to array mapping
		var config struct {
			SourcePath string            `json:"source_path"`
			DestPath   string            `json:"dest_path"`
			Mapping    map[string]string `json:"mapping"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		// Array mapping logic...
		return transformArray(value, config)

	case models.TransformationTypeJsonPath:
		// Extract value from JSON using path
		var config struct {
			Path string `json:"path"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		// JSONPath extraction logic...
		return extractJsonPath(value, config.Path)

	case models.TransformationTypeMediaMap:
		// Convert Shopware media to Shopify media
		var config struct {
			BaseURL string `json:"base_url"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		return transformMedia(value, config)

	case models.TransformationTypeMetafield:
		// Create Shopify metafield from Shopware value
		var config struct {
			Namespace string `json:"namespace"`
			Key       string `json:"key"`
			Type      string `json:"type"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		return createMetafield(value, config)

	case models.TransformationTypeEntityLookup:
		// Look up entity by ID and return a property
		var config struct {
			EntityType string `json:"entity_type"`
			Property   string `json:"property"`
		}

		if err := json.Unmarshal([]byte(mapping.TransformConfig), &config); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		return s.lookupEntity(value, config)

	default:
		return nil, fmt.Errorf("unsupported transformation type: %s", mapping.TransformType)
	}

	//default:
	//	return nil, fmt.Errorf("unsupported transformation type: %s", mapping.TransformType)
	//}
}

// transformMedia transforms Shopware media to Shopify media format
func transformMedia(value interface{}, config struct {
	BaseURL string `json:"base_url"`
}) (interface{}, error) {
	// Check if value is an array
	sourceArray, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("media value is not an array")
	}

	// Create destination array for media
	destMedia := make([]interface{}, 0, len(sourceArray))

	// Process each media item
	for i, item := range sourceArray {
		mediaMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract media properties
		var (
			url string
			alt string
			_   string
		)

		if urlVal, found := mediaMap["url"]; found && urlVal != nil {
			url = fmt.Sprintf("%s", urlVal)
		}

		if altVal, found := mediaMap["alt"]; found && altVal != nil {
			alt = fmt.Sprintf("%s", altVal)
		}

		if titleVal, found := mediaMap["title"]; found && titleVal != nil {
			_ = fmt.Sprintf("%s", titleVal)
		} else {
			// Use alt text as fallback for title
			_ = alt
		}

		// Construct the full URL
		fullURL := url
		if config.BaseURL != "" && !strings.HasPrefix(url, "http") {
			if strings.HasSuffix(config.BaseURL, "/") {
				fullURL = config.BaseURL + url
			} else {
				fullURL = config.BaseURL + "/" + url
			}
		}

		// Create the destination media object
		destItem := map[string]interface{}{
			"mediaContentType": "IMAGE",
			"originalSource":   fullURL,
			"alt":              alt,
			"position":         i + 1,
		}

		// Add the media item to the destination array
		destMedia = append(destMedia, destItem)
	}

	return destMedia, nil
}

// lookupEntity looks up an entity by ID and returns a property
func (s *FieldMappingService) lookupEntity(value interface{}, config struct {
	EntityType string `json:"entity_type"`
	Property   string `json:"property"`
}) (interface{}, error) {
	// Convert value to string ID
	strID := fmt.Sprintf("%v", value)
	if strID == "" {
		return nil, fmt.Errorf("empty entity ID")
	}

	// For Shopware, entity types include: product, category, manufacturer, etc.
	switch config.EntityType {
	case "manufacturer":
		// Look up manufacturer
		var manufacturer map[string]interface{}
		err := s.db.Table("manufacturer").
			Where("id = ?", strID).
			First(&manufacturer).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return strID, nil // Return original ID if not found
			}
			return nil, fmt.Errorf("error looking up manufacturer: %w", err)
		}

		// Extract the requested property
		if prop, ok := manufacturer[config.Property]; ok {
			return prop, nil
		}
		return strID, nil // Return original ID if property not found

	case "category":
		// Look up category
		var category map[string]interface{}
		err := s.db.Table("category").
			Where("id = ?", strID).
			First(&category).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return strID, nil
			}
			return nil, fmt.Errorf("error looking up category: %w", err)
		}

		if prop, ok := category[config.Property]; ok {
			return prop, nil
		}
		return strID, nil

	default:
		return strID, fmt.Errorf("unsupported entity type: %s", config.EntityType)
	}
}

// transformArray transforms an array based on the mapping configuration
func transformArray(value interface{}, config struct {
	SourcePath string            `json:"source_path"`
	DestPath   string            `json:"dest_path"`
	Mapping    map[string]string `json:"mapping"`
}) (interface{}, error) {
	// Check if value is an array
	sourceArray, ok := value.([]interface{})
	if !ok {
		// Try to convert from map to array if it's a single object
		if sourceMap, mapOk := value.(map[string]interface{}); mapOk {
			return []interface{}{transformSingleObject(sourceMap, config)}, nil
		}
		return nil, fmt.Errorf("value is not an array or object")
	}

	// Create destination array
	destArray := make([]interface{}, 0, len(sourceArray))

	// Process each item in the source array
	for _, item := range sourceArray {
		// Skip if not an object
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Transform the item
		destItem := transformSingleObject(itemMap, config)
		destArray = append(destArray, destItem)
	}

	return destArray, nil
}

// createMetafield transforms a value into a Shopify metafield
func createMetafield(value interface{}, config struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Type      string `json:"type"`
}) (interface{}, error) {
	if config.Namespace == "" || config.Key == "" {
		return nil, fmt.Errorf("metafield namespace and key are required")
	}

	// Determine the value type
	metafieldType := config.Type
	if metafieldType == "" {
		metafieldType = "string"
	}

	// Convert value based on metafield type
	var metafieldValue interface{}
	switch metafieldType {
	case "string":
		metafieldValue = fmt.Sprintf("%v", value)
	case "number_integer":
		if intVal, err := strconv.Atoi(fmt.Sprintf("%v", value)); err == nil {
			metafieldValue = intVal
		} else {
			metafieldValue = 0
		}
	case "number_decimal":
		if floatVal, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
			metafieldValue = floatVal
		} else {
			metafieldValue = 0.0
		}
	case "boolean":
		boolStr := strings.ToLower(fmt.Sprintf("%v", value))
		metafieldValue = boolStr == "true" || boolStr == "1" || boolStr == "yes"
	case "json_string":
		// For JSON, we just convert the value to a string
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			metafieldValue = "{}"
		} else {
			metafieldValue = string(jsonBytes)
		}
	default:
		metafieldValue = fmt.Sprintf("%v", value)
	}

	// Create the metafield object
	metafield := map[string]interface{}{
		"namespace": config.Namespace,
		"key":       config.Key,
		"value":     metafieldValue,
		"type":      metafieldType,
	}

	return metafield, nil
}

// transformSingleObject transforms a single object in an array
func transformSingleObject(item map[string]interface{}, config struct {
	SourcePath string            `json:"source_path"`
	DestPath   string            `json:"dest_path"`
	Mapping    map[string]string `json:"mapping"`
}) map[string]interface{} {
	result := make(map[string]interface{})

	// Extract source value using source path
	var sourceValue interface{}
	if config.SourcePath == "" {
		sourceValue = item // Use entire item if no source path
	} else {
		paths := strings.Split(config.SourcePath, ".")
		current := item
		for i, path := range paths {
			if i == len(paths)-1 {
				sourceValue = current[path]
				break
			}
			if nextMap, ok := current[path].(map[string]interface{}); ok {
				current = nextMap
			} else {
				return result // Path doesn't exist, return empty result
			}
		}
	}

	// Apply mapping if present
	mappedValue := sourceValue
	if config.Mapping != nil {
		if sourceStr, ok := sourceValue.(string); ok {
			if mapped, exists := config.Mapping[sourceStr]; exists {
				mappedValue = mapped
			}
		}
	}

	// Set destination value using dest path
	if config.DestPath == "" {
		// If no dest path, copy the whole item structure
		for k, v := range item {
			result[k] = v
		}
	} else {
		// Put the mapped value at the destination path
		paths := strings.Split(config.DestPath, ".")
		current := result
		for i, path := range paths {
			if i == len(paths)-1 {
				// Final path component, set the value
				current[path] = mappedValue
			} else {
				// Create intermediate objects if needed
				if _, exists := current[path]; !exists {
					current[path] = make(map[string]interface{})
				}
				current = current[path].(map[string]interface{})
			}
		}
	}

	return result
}

// extractJsonPath extracts a value from an object using a JSON path
func extractJsonPath(value interface{}, path string) (interface{}, error) {
	if path == "" {
		return value, nil
	}

	// Parse the JSON path
	components := strings.Split(path, ".")
	current := value

	for _, component := range components {
		// Handle array indexing in the path (e.g., items[0])
		var index int = -1
		var key string

		if match := arrayIndexRegex.FindStringSubmatch(component); len(match) > 0 {
			key = match[1]
			indexStr := match[2]
			idx, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index in path: %s", component)
			}
			index = idx
		} else {
			key = component
		}

		// Navigate the object
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[key]; exists {
				if index >= 0 {
					// Try to access array element
					if arr, ok := val.([]interface{}); ok {
						if index < len(arr) {
							current = arr[index]
						} else {
							return nil, fmt.Errorf("array index out of bounds: %d", index)
						}
					} else {
						return nil, fmt.Errorf("value at key %s is not an array", key)
					}
				} else {
					current = val
				}
			} else {
				return nil, fmt.Errorf("key %s not found in object", key)
			}
		case []interface{}:
			return nil, fmt.Errorf("cannot access property %s of an array", key)
		default:
			return nil, fmt.Errorf("cannot access property %s of a non-object", key)
		}
	}

	return current, nil
}

// Add a regex for parsing array indices in JSON paths
var arrayIndexRegex = regexp.MustCompile(`^([^\[]+)\[(\d+)\]$`)

// getNestedValue gets a value from a nested object using dot notation
func getNestedValue(obj map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	var current interface{} = obj

	for _, part := range parts {
		// Handle array access, e.g. "items[0]"
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			key := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			if currentMap, ok := current.(map[string]interface{}); ok {
				if arr, ok := currentMap[key].([]interface{}); ok {
					if index < 0 || index >= len(arr) {
						return nil, fmt.Errorf("array index out of bounds: %d", index)
					}
					current = arr[index]
				} else {
					return nil, fmt.Errorf("field %s is not an array", key)
				}
			} else {
				return nil, fmt.Errorf("cannot access %s: parent is not an object", key)
			}
		} else {
			// Regular object access
			if currentMap, ok := current.(map[string]interface{}); ok {
				var exists bool
				current, exists = currentMap[part]
				if !exists {
					return nil, fmt.Errorf("field %s not found", part)
				}
			} else {
				return nil, fmt.Errorf("cannot access %s: parent is not an object", part)
			}
		}
	}

	return current, nil
}

// setNestedValue sets a value in a nested object using dot notation
func setNestedValue(obj map[string]interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")

	// For all but the last part, ensure the path exists
	current := obj
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		// Handle array access, e.g. "items[0]"
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			key := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", indexStr)
			}

			// Ensure the key exists and is an array
			if _, ok := current[key]; !ok {
				current[key] = make([]interface{}, index+1)
			}

			arr, ok := current[key].([]interface{})
			if !ok {
				return fmt.Errorf("field %s is not an array", key)
			}

			// Ensure the array is big enough
			if index >= len(arr) {
				newArr := make([]interface{}, index+1)
				copy(newArr, arr)
				arr = newArr
				current[key] = arr
			}

			// If this is not the last part, ensure the array element is an object
			if i < len(parts)-2 {
				if arr[index] == nil {
					arr[index] = make(map[string]interface{})
				}

				if nextMap, ok := arr[index].(map[string]interface{}); ok {
					current = nextMap
				} else {
					return fmt.Errorf("array element at index %d is not an object", index)
				}
			}
		} else {
			// Regular object access
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]interface{})
			}

			nextMap, ok := current[part].(map[string]interface{})
			if !ok {
				return fmt.Errorf("field %s is not an object", part)
			}

			current = nextMap
		}
	}

	// Set the value at the last part
	lastPart := parts[len(parts)-1]

	// Handle array access for the last part
	if strings.Contains(lastPart, "[") && strings.Contains(lastPart, "]") {
		key := lastPart[:strings.Index(lastPart, "[")]
		indexStr := lastPart[strings.Index(lastPart, "[")+1 : strings.Index(lastPart, "]")]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", indexStr)
		}

		// Ensure the key exists and is an array
		if _, ok := current[key]; !ok {
			current[key] = make([]interface{}, index+1)
		}

		arr, ok := current[key].([]interface{})
		if !ok {
			return fmt.Errorf("field %s is not an array", key)
		}

		// Ensure the array is big enough
		if index >= len(arr) {
			newArr := make([]interface{}, index+1)
			copy(newArr, arr)
			arr = newArr
			current[key] = arr
		}

		// Set the value at the specified index
		arr[index] = value
	} else {
		// Regular object access
		current[lastPart] = value
	}

	return nil

}

// convertToGraphQLGlobalID converts a regular ID to a Shopify GraphQL Global ID
func (s *FieldMappingService) convertToGraphQLGlobalID(resourceType string, id string) string {
	return fmt.Sprintf("gid://shopify/%s/%s", resourceType, id)
}

// convertFromGraphQLGlobalID extracts the original ID from a Shopify GraphQL Global ID
func (s *FieldMappingService) convertFromGraphQLGlobalID(globalID string) string {
	// GraphQL IDs are in the format: gid://shopify/ResourceType/ID
	parts := strings.Split(globalID, "/")
	if len(parts) < 4 {
		return globalID // Not a valid global ID, return as is
	}
	return parts[len(parts)-1]
}

// GetDefaultProductMappings returns a set of default field mappings for product migration
func (s *FieldMappingService) GetDefaultProductMappings(dataflowID uint) []models.FieldMapping {
	return []models.FieldMapping{
		{
			DataflowID:      dataflowID,
			SourceField:     "id",
			DestField:       "id",
			IsRequired:      true,
			TransformType:   models.TransformationTypeGraphQLID,
			TransformConfig: `{"resource_type": "Product", "direction": "to_global"}`,
		},
		{
			DataflowID:    dataflowID,
			SourceField:   "name",
			DestField:     "title",
			IsRequired:    true,
			TransformType: models.TransformationTypeNone,
		},
		{
			DataflowID:    dataflowID,
			SourceField:   "description",
			DestField:     "descriptionHtml",
			IsRequired:    false,
			TransformType: models.TransformationTypeNone,
		},
		{
			DataflowID:    dataflowID,
			SourceField:   "productNumber",
			DestField:     "variants[0].sku",
			IsRequired:    false,
			TransformType: models.TransformationTypeNone,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "stock",
			DestField:       "variants[0].inventoryQuantity",
			IsRequired:      false,
			TransformType:   models.TransformationTypeConvert,
			TransformConfig: `{"type": "int"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "price[0].gross",
			DestField:       "variants[0].price",
			IsRequired:      true,
			TransformType:   models.TransformationTypeConvert,
			TransformConfig: `{"type": "string"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "active",
			DestField:       "status",
			IsRequired:      false,
			DefaultValue:    "ACTIVE",
			TransformType:   models.TransformationTypeMap,
			TransformConfig: `{"true": "ACTIVE", "false": "DRAFT", "_default": "DRAFT"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "manufacturerId",
			DestField:       "vendor",
			IsRequired:      false,
			TransformType:   models.TransformationTypeEntityLookup,
			TransformConfig: `{"entity_type": "manufacturer", "property": "name"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "categoryIds",
			DestField:       "collections",
			IsRequired:      false,
			TransformType:   models.TransformationTypeArrayMap,
			TransformConfig: `{"source_path": "id", "dest_path": "id"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "media",
			DestField:       "media",
			IsRequired:      false,
			TransformType:   models.TransformationTypeMediaMap,
			TransformConfig: `{"base_url": ""}`,
		},
		{
			DataflowID:    dataflowID,
			SourceField:   "metaTitle",
			DestField:     "seo.title",
			IsRequired:    false,
			TransformType: models.TransformationTypeNone,
		},
		{
			DataflowID:    dataflowID,
			SourceField:   "metaDescription",
			DestField:     "seo.description",
			IsRequired:    false,
			TransformType: models.TransformationTypeNone,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "weight",
			DestField:       "variants[0].weight",
			IsRequired:      false,
			TransformType:   models.TransformationTypeConvert,
			TransformConfig: `{"type": "float"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "width",
			DestField:       "metafields[0]",
			IsRequired:      false,
			TransformType:   models.TransformationTypeMetafield,
			TransformConfig: `{"namespace": "dimensions", "key": "width", "type": "number_decimal"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "height",
			DestField:       "metafields[1]",
			IsRequired:      false,
			TransformType:   models.TransformationTypeMetafield,
			TransformConfig: `{"namespace": "dimensions", "key": "height", "type": "number_decimal"}`,
		},
		{
			DataflowID:      dataflowID,
			SourceField:     "length",
			DestField:       "metafields[2]",
			IsRequired:      false,
			TransformType:   models.TransformationTypeMetafield,
			TransformConfig: `{"namespace": "dimensions", "key": "length", "type": "number_decimal"}`,
		},
	}
}
