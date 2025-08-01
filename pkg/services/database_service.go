package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bonpreu-go/pkg/config"
	"bonpreu-go/pkg/models"
	"bonpreu-go/pkg/utils"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// DatabaseService handles all database operations for the Bonpreu application.
// It provides methods for connecting to PostgreSQL, saving products and nutritional data,
// and retrieving database statistics. The service uses bulk operations for optimal performance.
type DatabaseService struct {
	db     *sql.DB
	logger *utils.Logger
}

// NewDatabaseService creates a new DatabaseService instance with the provided configuration.
// It establishes a connection to the PostgreSQL database using the connection details
// from the configuration. The connection is tested with a ping before returning.
func NewDatabaseService(cfg *config.Configuration) (*DatabaseService, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseService{
		db:     db,
		logger: utils.NewLogger("DatabaseService"),
	}, nil
}

// Close closes the database connection and releases associated resources.
// This method should be called when the DatabaseService is no longer needed.
func (d *DatabaseService) Close() error {
	return d.db.Close()
}

// SaveProducts saves multiple products to the database using bulk insert operations.
// It uses PostgreSQL's VALUES clause for optimal performance and handles conflicts
// with ON CONFLICT DO UPDATE to update existing records. The operation is performed
// within a transaction for data consistency.
func (d *DatabaseService) SaveProducts(products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	start := time.Now()
	d.logger.Info("Saving %d products to database...", len(products))

	// Begin transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use bulk insert with batching to respect PostgreSQL parameter limits
	// PostgreSQL supports max 65535 parameters, so max ~4000 products per batch (16 params each)
	maxParamsPerBatch := 60000
	maxProductsPerBatch := maxParamsPerBatch / 16

	for i := 0; i < len(products); i += maxProductsPerBatch {
		end := i + maxProductsPerBatch
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		values := make([]string, 0, len(batch))
		args := make([]interface{}, 0, len(batch)*16)
		argIndex := 1

		for _, product := range batch {
			values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5, argIndex+6, argIndex+7,
				argIndex+8, argIndex+9, argIndex+10, argIndex+11, argIndex+12, argIndex+13, argIndex+14, argIndex+15))

			args = append(args,
				product.ProductID,
				product.ProductType,
				product.ProductName,
				product.ProductDescription,
				product.ProductBrand,
				product.ProductPackSizeDescription,
				product.ProductPriceAmount,
				product.ProductCurrency,
				product.ProductUnitPriceAmount,
				product.ProductUnitPriceCurrency,
				product.ProductUnitPriceUnit,
				product.ProductAvailable,
				product.ProductAlcohol,
				product.ProductCookingGuidelines,
				pq.Array(product.ProductCategories),
				product.CreatedAt,
			)
			argIndex += 16
		}

		query := fmt.Sprintf(`
			INSERT INTO products (
				product_id, product_type, product_name, product_description, 
				product_brand, product_pack_size_description, product_price_amount, 
				product_currency, product_unit_price_amount, product_unit_price_currency, 
				product_unit_price_unit, product_available, product_alcohol, 
				product_cooking_guidelines, product_categories, created_at
			) VALUES %s
			ON CONFLICT (product_id) DO UPDATE SET
				product_type = EXCLUDED.product_type,
				product_name = EXCLUDED.product_name,
				product_description = EXCLUDED.product_description,
				product_brand = EXCLUDED.product_brand,
				product_pack_size_description = EXCLUDED.product_pack_size_description,
				product_price_amount = EXCLUDED.product_price_amount,
				product_currency = EXCLUDED.product_currency,
				product_unit_price_amount = EXCLUDED.product_unit_price_amount,
				product_unit_price_currency = EXCLUDED.product_unit_price_currency,
				product_unit_price_unit = EXCLUDED.product_unit_price_unit,
				product_available = EXCLUDED.product_available,
				product_alcohol = EXCLUDED.product_alcohol,
				product_cooking_guidelines = EXCLUDED.product_cooking_guidelines,
				product_categories = EXCLUDED.product_categories,
				updated_at = CURRENT_TIMESTAMP
		`, strings.Join(values, ","))

		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("failed to bulk insert products batch %d-%d: %w", i+1, end, err)
		}

		// Log progress for large datasets
		if len(products) > 1000 {
			d.logger.Info("Inserted batch %d-%d of %d products", i+1, end, len(products))
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	d.logger.Info("Successfully saved %d products in %v", len(products), time.Since(start))
	return nil
}

// SaveNutritionalData saves nutritional data to the database using bulk insert operations.
// It uses PostgreSQL's VALUES clause for optimal performance and handles conflicts
// with ON CONFLICT DO NOTHING to avoid duplicate entries. The operation is performed
// within a transaction for data consistency.
func (d *DatabaseService) SaveNutritionalData(nutritionalData []models.ProductNutritionalData) error {
	if len(nutritionalData) == 0 {
		return nil
	}

	start := time.Now()
	d.logger.Info("Saving %d nutritional data entries to database...", len(nutritionalData))

	// Begin transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Use bulk insert for nutritional data with batching
	// Nutritional data has 4 parameters per record, so max ~15000 records per batch
	maxParamsPerBatch := 60000
	maxNutritionalPerBatch := maxParamsPerBatch / 4

	for i := 0; i < len(nutritionalData); i += maxNutritionalPerBatch {
		end := i + maxNutritionalPerBatch
		if end > len(nutritionalData) {
			end = len(nutritionalData)
		}

		batch := nutritionalData[i:end]
		values := make([]string, 0, len(batch))
		args := make([]interface{}, 0, len(batch)*4)
		argIndex := 1

		for _, data := range batch {
			values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)",
				argIndex, argIndex+1, argIndex+2, argIndex+3))

			args = append(args,
				data.ProductID,
				data.ProductNutritionalValue,
				data.ProductNutritionalQuantity,
				data.CreatedAt,
			)
			argIndex += 4
		}

		query := fmt.Sprintf(`
			INSERT INTO product_nutritional_data (
				product_id, product_nutritional_value, product_nutritional_quantity, created_at
			) VALUES %s
			ON CONFLICT DO NOTHING
		`, strings.Join(values, ","))

		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("failed to bulk insert nutritional data batch %d-%d: %w", i+1, end, err)
		}

		// Log progress for large datasets
		if len(nutritionalData) > 1000 {
			d.logger.Info("Inserted batch %d-%d of %d nutritional data entries", i+1, end, len(nutritionalData))
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	d.logger.Info("Successfully saved %d nutritional data entries in %v", len(nutritionalData), time.Since(start))
	return nil
}

// SaveAllData saves both products and nutritional data to the database.
// It first saves all products, then saves all nutritional data. This ensures
// that foreign key constraints are satisfied. The operation is optimized for
// large datasets with bulk insert operations.
func (d *DatabaseService) SaveAllData(products []models.Product, nutritionalData []models.ProductNutritionalData) error {
	start := time.Now()
	d.logger.Info("Saving all data to database...")

	// Save products first
	if err := d.SaveProducts(products); err != nil {
		return fmt.Errorf("failed to save products: %w", err)
	}

	// Save nutritional data
	if err := d.SaveNutritionalData(nutritionalData); err != nil {
		return fmt.Errorf("failed to save nutritional data: %w", err)
	}

	d.logger.Info("Successfully saved all data in %v", time.Since(start))
	return nil
}

// GetProductCount returns the total number of products in the database.
// This method provides a quick way to check the current state of the products table.
func (d *DatabaseService) GetProductCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get product count: %w", err)
	}
	return count, nil
}

// GetNutritionalDataCount returns the total number of nutritional data entries in the database.
// This method provides a quick way to check the current state of the nutritional data table.
func (d *DatabaseService) GetNutritionalDataCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM product_nutritional_data").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get nutritional data count: %w", err)
	}
	return count, nil
}
