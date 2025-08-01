package models

import (
	"strconv"
	"strings"
	"time"
)

// Product represents a product from the Bonpreu API.
// It contains all the essential product information including pricing,
// availability, categories, and metadata.
type Product struct {
	ProductID                  int       `json:"product_id"`
	ProductType                string    `json:"product_type"`
	ProductName                string    `json:"product_name"`
	ProductDescription         string    `json:"product_description"`
	ProductBrand               string    `json:"product_brand"`
	ProductPackSizeDescription string    `json:"product_pack_size_description"`
	ProductPriceAmount         float64   `json:"product_price_amount"`
	ProductCurrency            string    `json:"product_currency"`
	ProductUnitPriceAmount     float64   `json:"product_unit_price_amount"`
	ProductUnitPriceCurrency   string    `json:"product_unit_price_currency"`
	ProductUnitPriceUnit       string    `json:"product_unit_price_unit"`
	ProductAvailable           bool      `json:"product_available"`
	ProductAlcohol             bool      `json:"product_alcohol"`
	ProductCookingGuidelines   string    `json:"product_cooking_guidelines"`
	ProductCategories          []string  `json:"product_categories"`
	CreatedAt                  time.Time `json:"created_at"`
}

// ProductNutritionalData represents nutritional information for a product.
// It contains the nutritional value name and quantity for a specific product.
type ProductNutritionalData struct {
	ID                         *int      `json:"id,omitempty"`
	ProductID                  int       `json:"product_id"`
	ProductNutritionalValue    string    `json:"product_nutritional_value"`
	ProductNutritionalQuantity string    `json:"product_nutritional_quantity"`
	CreatedAt                  time.Time `json:"created_at"`
}

// APIResponse represents the raw JSON response from the Bonpreu API.
// It contains the product data and additional BOP (Bonpreu) specific information.
type APIResponse struct {
	Product ProductData `json:"product"`
	BopData BopData     `json:"bopData"`
}

// ProductData represents the core product information from the API response.
type ProductData struct {
	RetailerProductID   int       `json:"retailerProductId"`
	Type                string    `json:"type"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Brand               string    `json:"brand"`
	PackSizeDescription string    `json:"packSizeDescription"`
	Price               Price     `json:"price"`
	UnitPrice           UnitPrice `json:"unitPrice"`
	Available           bool      `json:"available"`
	Alcohol             bool      `json:"alcohol"`
	CookingGuidelines   string    `json:"cookingGuidelines"`
	CategoryPath        []string  `json:"categoryPath"`
}

// Price represents the price information for a product.
type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// UnitPrice represents the unit price information for a product.
type UnitPrice struct {
	Price Price  `json:"price"`
	Unit  string `json:"unit"`
}

// BopData represents additional Bonpreu-specific product information.
// It contains detailed descriptions and custom fields like nutritional data.
type BopData struct {
	DetailedDescription string  `json:"detailedDescription"`
	Fields              []Field `json:"fields"`
}

// Field represents a custom field in the BOP data.
// It can contain various types of information like nutritional data or cooking guidelines.
type Field struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ParseProductFromResponse parses a raw JSON response from the Bonpreu API
// and converts it into a Product struct. It handles the complex nested structure
// of the API response and extracts all relevant product information.
func ParseProductFromResponse(responseJSON map[string]interface{}, productID int) Product {
	product := Product{
		ProductID:         productID,
		CreatedAt:         time.Now(),
		ProductCategories: []string{},
	}

	// Extract product data
	if productData, ok := responseJSON["product"].(map[string]interface{}); ok {
		// Basic product information
		if productType, ok := productData["type"].(string); ok {
			product.ProductType = productType
		}
		if productName, ok := productData["name"].(string); ok {
			product.ProductName = strings.ReplaceAll(productName, "<br />", "")
		}
		if productDescription, ok := productData["description"].(string); ok {
			product.ProductDescription = strings.ReplaceAll(productDescription, "<br />", "")
		}
		if productBrand, ok := productData["brand"].(string); ok {
			product.ProductBrand = productBrand
		}
		if packSizeDescription, ok := productData["packSizeDescription"].(string); ok {
			product.ProductPackSizeDescription = packSizeDescription
		}
		if available, ok := productData["available"].(bool); ok {
			product.ProductAvailable = available
		}
		if alcohol, ok := productData["alcohol"].(bool); ok {
			product.ProductAlcohol = alcohol
		}

		// Price information
		if priceData, ok := productData["price"].(map[string]interface{}); ok {
			if amount, ok := priceData["amount"].(float64); ok {
				product.ProductPriceAmount = amount
			} else if amountStr, ok := priceData["amount"].(string); ok {
				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					product.ProductPriceAmount = amount
				}
			}
			if currency, ok := priceData["currency"].(string); ok {
				product.ProductCurrency = currency
			}
		}

		// Unit price information
		if unitPriceData, ok := productData["unitPrice"].(map[string]interface{}); ok {
			if unitPricePrice, ok := unitPriceData["price"].(map[string]interface{}); ok {
				if amount, ok := unitPricePrice["amount"].(float64); ok {
					product.ProductUnitPriceAmount = amount
				} else if amountStr, ok := unitPricePrice["amount"].(string); ok {
					if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
						product.ProductUnitPriceAmount = amount
					}
				}
				if currency, ok := unitPricePrice["currency"].(string); ok {
					product.ProductUnitPriceCurrency = currency
				}
			}
			if unit, ok := unitPriceData["unit"].(string); ok {
				product.ProductUnitPriceUnit = unit
			}
		}

		// Categories
		if categoryPath, ok := productData["categoryPath"].([]interface{}); ok {
			for _, cat := range categoryPath {
				if catStr, ok := cat.(string); ok {
					product.ProductCategories = append(product.ProductCategories, catStr)
				}
			}
		}

		// Extract description and cooking guidelines from bopData
		if bopData, ok := responseJSON["bopData"].(map[string]interface{}); ok {
			if detailedDesc, ok := bopData["detailedDescription"].(string); ok {
				product.ProductDescription = strings.ReplaceAll(detailedDesc, "<br />", "")
			}

			// Extract cooking guidelines
			if fields, ok := bopData["fields"].([]interface{}); ok {
				for _, field := range fields {
					if fieldMap, ok := field.(map[string]interface{}); ok {
						if title, ok := fieldMap["title"].(string); ok && title == "cookingGuidelines" {
							if content, ok := fieldMap["content"].(string); ok {
								product.ProductCookingGuidelines = strings.ReplaceAll(content, "<br />", "")
							}
							break
						}
					}
				}
			}
		}
	}

	return product
}

// ParseNutritionalDataFromResponse parses nutritional data from the API response.
// It looks for the "nutritionalData" field in the BOP data and extracts
// nutritional information from the HTML table content.
func ParseNutritionalDataFromResponse(responseJSON map[string]interface{}, productID int) []ProductNutritionalData {
	var nutritionalData []ProductNutritionalData

	if bopData, ok := responseJSON["bopData"].(map[string]interface{}); ok {
		if fields, ok := bopData["fields"].([]interface{}); ok {
			for _, field := range fields {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if title, ok := fieldMap["title"].(string); ok && title == "nutritionalData" {
						if content, ok := fieldMap["content"].(string); ok {
							nutritionalData = parseNutritionalDataTable(content, productID)
						}
						break
					}
				}
			}
		}
	}

	return nutritionalData
}

// parseNutritionalDataTable parses the HTML table containing nutritional data.
// It extracts nutritional values and quantities from HTML table rows and cells.
// The function handles basic HTML table parsing for nutritional information.
func parseNutritionalDataTable(html string, productID int) []ProductNutritionalData {
	var nutritionalData []ProductNutritionalData

	// Simple regex-based parser for HTML table
	// Look for patterns like: <td>Nutrient Name</td><td>Value</td>
	rows := strings.Split(html, "<tr>")

	for _, row := range rows {
		// Skip header rows and empty rows
		if strings.Contains(row, "<th>") || strings.TrimSpace(row) == "" {
			continue
		}

		// Extract cells
		cells := strings.Split(row, "<td>")
		if len(cells) >= 3 { // At least 2 data cells + empty first element
			// Extract nutritional value (first cell)
			valueCell := cells[1]
			value := strings.TrimSpace(strings.ReplaceAll(valueCell, "</td>", ""))
			value = strings.ReplaceAll(value, "<br />", "")

			// Extract quantity (second cell)
			quantityCell := cells[2]
			quantity := strings.TrimSpace(strings.ReplaceAll(quantityCell, "</td>", ""))
			quantity = strings.ReplaceAll(quantity, "<br />", "")

			// Only add if we have both value and quantity
			if value != "" && quantity != "" {
				nutritionalData = append(nutritionalData, ProductNutritionalData{
					ProductID:                  productID,
					ProductNutritionalValue:    value,
					ProductNutritionalQuantity: quantity,
					CreatedAt:                  time.Now(),
				})
			}
		}
	}

	return nutritionalData
}
