package main

import (
	"log"
	"time"

	"bonpreu-go/pkg/config"
	"bonpreu-go/pkg/services"
	"bonpreu-go/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	start := time.Now()
	logger := utils.NewLogger("Main")

	logger.Info("Starting Bonpreu Go application")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using system environment variables")
	}

	// Load configuration
	// Choose your mode:
	// cfg := config.DefaultConfig() // Production: Uses REQUEST_DURATION_MINUTES from env
	cfg := config.TestingConfig() // Testing: no rate limiting

	logger.Info("Loaded configuration")

	// Initialize services
	sitemapService := services.NewSitemapService()
	productService := services.NewProductService(200) // Increased to 200 concurrent workers

	// Initialize database service
	dbService, err := services.NewDatabaseService(cfg)
	if err != nil {
		logger.Error("Error initializing database service: %v", err)
		log.Fatalf("Error initializing database service: %v", err)
	}
	defer dbService.Close()

	logger.Info("Initialized services")

	// Fetch product IDs
	logger.Info("Fetching product IDs from sitemap...")

	productIDs, err := sitemapService.FetchProductIds(cfg.SitemapURL)
	if err != nil {
		logger.Error("Error fetching product IDs: %v", err)
		log.Fatalf("Error fetching product IDs: %v", err)
	}

	logger.Info("Successfully fetched %d product IDs", len(productIDs))

	// Extract product IDs as integers for the product service
	var productIDInts []int
	for _, item := range productIDs {
		productIDInts = append(productIDInts, item.ProductID)
	}

	// Fetch product data using duration from config
	if cfg.RequestDuration > 0 {
		logger.Info("Fetching product data for %d products over %v...", len(productIDInts), cfg.RequestDuration)
	} else {
		logger.Info("Fetching product data for %d products (no rate limiting)...", len(productIDInts))
	}

	products, nutritionalData, err := productService.FetchAllProductsData(productIDInts, cfg.RequestDuration)
	if err != nil {
		logger.Error("Error fetching product data: %v", err)
		log.Fatalf("Error fetching product data: %v", err)
	}

	logger.Info("Successfully fetched data for %d products", len(products))
	logger.Info("Total nutritional data entries: %d", len(nutritionalData))

	// Save data to database
	logger.Info("Saving data to database...")
	if err := dbService.SaveAllData(products, nutritionalData); err != nil {
		logger.Error("Error saving data to database: %v", err)
		log.Fatalf("Error saving data to database: %v", err)
	}

	// Get database statistics
	productCount, err := dbService.GetProductCount()
	if err != nil {
		logger.Error("Error getting product count: %v", err)
	} else {
		logger.Info("Total products in database: %d", productCount)
	}

	nutritionalCount, err := dbService.GetNutritionalDataCount()
	if err != nil {
		logger.Error("Error getting nutritional data count: %v", err)
	} else {
		logger.Info("Total nutritional data entries in database: %d", nutritionalCount)
	}

	logger.LogDuration("Application execution", start)
}
