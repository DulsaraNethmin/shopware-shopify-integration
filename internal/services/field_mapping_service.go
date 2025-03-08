package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/shopware-shopify-integration/internal/models"
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

	default:
		return nil, fmt.Errorf("unsupported transformation type: %s", mapping.TransformType)
	}
}

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
