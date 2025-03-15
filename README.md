# Field Mapping

## Data Models

### Shopware Product Model

```json
{
  "id": "string",
  "productNumber": "string",
  "name": "string",
  "description": "string",
  "active": "boolean",
  "stock": "integer",
  "availableStock": "integer" ,
  "price": [
    {
      "currencyId": "string",
      "gross": "float",
      "net": "float",
      "linked": "boolean"
    }
  ],
  "purchasePrices": [
    {
      "currencyId": "string",
      "gross": "float",
      "net": "float",
      "linked": "boolean"
    }
  ],
  "taxId": "string",
  "manufacturerId": "string",
  "categoryIds": [
    "string"
  ],
  "media": [
    {
      "mediaId": "string",
      "position": "integer"
    }
  ],
  "visibilities": [
    {
      "salesChannelId": "string",
      "visibility": "integer"
    }
  ],
  "metaTitle": "string",
  "metaDescription": "string",
  "keywords": "string",
  "minPurchase": "integer",
  "maxPurchase": "integer",
  "isCloseout": "boolean",
  "purchaseSteps": "integer",
  "unitId": "string",
  "weight": "float",
  "width": "float",
  "height": "float",
  "length": "float",
  "releaseDate": "date-time",
  "manufacturer": {
    "id": "string",
    "name": "string"
  },
  "categories": [
    {
      "id": "string",
      "name": "string"
    }
  ],
  "tags": [
    {
      "id": "string",
      "name": "string"
    }
  ],
  "properties": [
    {
      "id": "string",
      "name": "string",
      "group": {
        "id": "string",
        "name": "string"
      }
    }
  ],
  "options": [
    {
      "id": "string",
      "name": "string"
    }
  ],
  "customFields": {
    "custom_field_1": "value",
    "custom_field_2": "integer"
  },
  "createdAt": "date-time",
  "updatedAt": "date-time"
}
```

### Shopify Product Model

```json
{
  "id": "string",
  "title": "string",
  "descriptionHtml": "<p>Product description</p>",
  "vendor": "string",
  "productType": "string",
  "tags": ["tag1", "tag2"],
  "status": "ACTIVE",
  "publishedAt": "2024-03-14T12:00:00.000Z",
  "templateSuffix": "string",
  "seo": {
    "title": "string",
    "description": "string"
  },
  "options": [
    {
      "name": "Size",
      "values": ["Small", "Medium", "Large"]
    }
  ],
  "variants": [
    {
      "id": "string",
      "title": "string",
      "sku": "SKU123",
      "price": "19.99",
      "compareAtPrice": "24.99",
      "weight": "float",
      "weightUnit": "KG",
      "barcode": "string",
      "inventoryItemId": "string",
      "inventoryPolicy": "CONTINUE",
      "inventoryQuantity": "integer",
      "requiresShipping": "boolean",
      "taxable": "boolean",
      "selectedOptions": [
        {
          "name": "Size",
          "value": "Medium"
        }
      ]
    }
  ],
  "media": [
    {
      "altText": "string",
      "mediaContentType": "IMAGE",
      "previewImage": {
        "src": "https://cdn.shopify.com/path-to-image.jpg",
        "altText": "Product image"
      }
    }
  ],
  "collections": [
    {
      "id": "string",
      "title": "string"
    }
  ],
  "metafields": [
    {
      "namespace": "custom",
      "key": "material",
      "value": "Cotton",
      "type": "string"
    }
  ],
  "customProductType": "string",
  "createdAt": "date-time",
  "updatedAt": "date-time"
}
```

### Field Mapping Model

```go
type FieldMapping struct {
    ID             uint                `json:"id" gorm:"primaryKey"`
    DataflowID     uint                `json:"dataflow_id" gorm:"not null"`
    SourceField    string              `json:"source_field" gorm:"not null"`
    DestField      string              `json:"dest_field" gorm:"not null"`
    IsRequired     bool                `json:"is_required" gorm:"default:false"`
    DefaultValue   string              `json:"default_value"`
    TransformType  TransformationType  `json:"transform_type" gorm:"default:'none'"`
    TransformConfig string              `json:"transform_config"`
}
```

### Transformation Types

The system supports multiple transformation types:

| Type | Description | Example |
|------|-------------|---------|
| `TransformationTypeNone` | Direct mapping without changes | Copying title directly |
| `TransformationTypeFormat` | Format transformation (e.g., date formats) | Converting date formats |
| `TransformationTypeConvert` | Type conversion | String to int/float conversion |
| `TransformationTypeMap` | Value mapping | Mapping status codes between systems |
| `TransformationTypeTemplate` | Template-based transformations | Creating text with value insertions |
| `TransformationTypeArrayMap` | Array-to-array mapping | Transforming arrays of objects |
| `TransformationTypeJsonPath` | Extract values using JSON paths | Accessing deeply nested values |
| `TransformationTypeMediaMap` | Converting media formats | Shopware media to Shopify format |
## Transformation Process

The transformation is handled by the `TransformData` method in `FieldMappingService`:

1. **Retrieve all field mappings** for the specific dataflow
2. **Parse the source data** from JSON
3. **Create an empty destination object**
4. For each field mapping:
    - **Extract the source value** using dot notation
    - **Apply the specified transformation**
    - **Set the transformed value** in the destination object
5. **Return the transformed data**

## Key Features

### Nested Field Support

The system supports both dot notation and array indexing:

- Dot notation: `product.price.gross`
- Array indexing: `variants[0].price`

### Default Product Mappings

Default mappings are provided for common Shopware to Shopify field mappings:

```go
// Example of default product mappings
{
    SourceField:     "id",
    DestField:       "id",
    IsRequired:      true,
    TransformType:   TransformationTypeGraphQLID,
    TransformConfig: `{"resource_type": "Product", "direction": "to_global"}`,
},
{
    SourceField:   "name",
    DestField:     "title",
    IsRequired:    true,
    TransformType: TransformationTypeNone,
},
// ...additional mappings
```

### Helper Functions

The system includes several helper functions:

- `getNestedValue`: Extracts values from nested objects
- `setNestedValue`: Sets values in nested objects
- `applyTransformation`: Applies the requested transformation
- `transformArray`, `createMetafield`, `transformMedia`: Specialized transformation functions

## Example Mappings

| Shopware Field | Shopify Field | Transformation | Notes |
|----------------|---------------|----------------|-------|
| `name` | `title` | None | Simple field rename |
| `description` | `descriptionHtml` | None | Direct mapping |
| `productNumber` | `variants[0].sku` | None | Mapping to nested field |
| `stock` | `variants[0].inventoryQuantity` | Convert to int | Type conversion |
| `price[0].gross` | `variants[0].price` | Convert to string | Type conversion |
| `active` | `status` | Map values | `true` → `'ACTIVE'`, `false` → `'DRAFT'` |
| `manufacturerId` | `vendor` | Entity lookup | Convert ID to manufacturer name |
| `media` | `media` | Media transformation | Complex media object transformation |
| `width` | `metafields[0]` | Metafield creation | Create width dimension metafield |
| `metaTitle` | `seo.title` | None | Mapping to nested SEO field |
| `metaDescription` | `seo.description` | None | Mapping to nested SEO field |
| `categoryIds` | `collections` | Array mapping | Convert category IDs to collections |
| `weight` | `variants[0].weight` | Convert to float | Weight conversion |
