package services

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"bonpreu-go/pkg/models"
	"bonpreu-go/pkg/utils"
)

// ProductService handles asynchronous fetching of product data from the Bonpreu API.
// It manages concurrent requests with rate limiting and provides progress tracking.
type ProductService struct {
	client      *http.Client
	logger      *utils.Logger
	semaphore   chan struct{}
	maxWorkers  int
	rateLimiter *time.Ticker
}

// ProductResult represents the result of a single product fetch operation.
// It contains the fetched product data, nutritional information, any errors,
// and the product ID for identification.
type ProductResult struct {
	Product         models.Product
	NutritionalData []models.ProductNutritionalData
	Error           error
	ProductID       int
}

// ProgressStats tracks the progress of the product fetching operation.
// It maintains atomic counters for thread-safe progress monitoring.
type ProgressStats struct {
	TotalProducts  int64
	ProcessedCount int64
	SuccessCount   int64
	NotFoundCount  int64
	ErrorCount     int64
	StartTime      time.Time
}

// NewProductService creates a new ProductService instance with the specified number of workers.
// The service uses a worker pool pattern to manage concurrent HTTP requests efficiently.
// maxWorkers determines the maximum number of concurrent requests that can be processed.
func NewProductService(maxWorkers int) *ProductService {
	if maxWorkers <= 0 {
		maxWorkers = 200
	}

	return &ProductService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:     utils.NewLogger("ProductService"),
		semaphore:  make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// FetchAllProductsData asynchronously fetches product data for all provided product IDs.
// It implements rate limiting when duration > 0, spreading requests over the specified duration.
// The function returns slices of successfully fetched products and nutritional data,
// along with any errors that occurred during the process.
func (p *ProductService) FetchAllProductsData(productIDs []int, duration time.Duration) ([]models.Product, []models.ProductNutritionalData, error) {
	start := time.Now()

	// Calculate rate limiting parameters
	totalRequests := len(productIDs)
	var requestsPerSecond float64
	var delayBetweenRequests time.Duration

	if duration > 0 {
		requestsPerSecond = float64(totalRequests) / duration.Seconds()
		delayBetweenRequests = time.Duration(float64(time.Second) / requestsPerSecond)
		p.logger.Info("Starting to fetch data for %d products with max %d concurrent workers", len(productIDs), p.maxWorkers)
		p.logger.Info("Rate limiting: %.2f requests/second (%.2f ms between requests) over %v",
			requestsPerSecond, float64(delayBetweenRequests.Microseconds())/1000, duration)
	} else {
		requestsPerSecond = 0
		delayBetweenRequests = 0
		p.logger.Info("Starting to fetch data for %d products with max %d concurrent workers (no rate limiting)", len(productIDs), p.maxWorkers)
	}

	// Initialize progress tracking
	stats := &ProgressStats{
		TotalProducts: int64(len(productIDs)),
		StartTime:     time.Now(),
	}

	// Create channels for results and coordination
	resultChan := make(chan ProductResult, len(productIDs))
	var wg sync.WaitGroup

	// Start progress monitoring goroutine
	done := make(chan bool)
	progressDone := make(chan bool)
	go p.monitorProgress(stats, done, progressDone)

	// Create rate limiter ticker (only if rate limiting is enabled)
	var rateLimiter *time.Ticker

	if duration > 0 {
		rateLimiter = time.NewTicker(delayBetweenRequests)
		defer rateLimiter.Stop()
	}

	// Create a job channel for the worker pool
	jobChan := make(chan int, len(productIDs))

	// Start worker goroutines
	for i := 0; i < p.maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for productID := range jobChan {
				// Wait for rate limiter tick (only if rate limiting is enabled)
				if duration > 0 {
					<-rateLimiter.C
				}

				p.fetchSingleProductData(productID, resultChan, stats)
			}
		}(i)
	}

	// Send jobs to workers
	go func() {
		for _, productID := range productIDs {
			jobChan <- productID
		}
		close(jobChan)
	}()

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
		done <- true
	}()

	// Collect results
	var products []models.Product
	var nutritionalData []models.ProductNutritionalData
	var errors []error

	for result := range resultChan {
		atomic.AddInt64(&stats.ProcessedCount, 1)

		if result.Error != nil {
			if result.Error.Error() == fmt.Sprintf("product %d not found", result.ProductID) {
				atomic.AddInt64(&stats.NotFoundCount, 1)
			} else {
				atomic.AddInt64(&stats.ErrorCount, 1)
				errors = append(errors, result.Error)
			}
		} else {
			atomic.AddInt64(&stats.SuccessCount, 1)
			products = append(products, result.Product)
			nutritionalData = append(nutritionalData, result.NutritionalData...)
		}
	}

	// Wait for progress monitoring to finish
	<-progressDone

	// Print final statistics
	p.logger.Info("Completed fetching products:")
	p.logger.Info("  - Total processed: %d", stats.ProcessedCount)
	p.logger.Info("  - Successful: %d", stats.SuccessCount)
	p.logger.Info("  - Not found (404): %d", stats.NotFoundCount)
	p.logger.Info("  - Errors: %d", stats.ErrorCount)
	p.logger.LogDuration("FetchAllProductsData", start)

	return products, nutritionalData, nil
}

// monitorProgress displays periodic status updates during the fetching process.
// It updates every minute and provides concise progress information.
func (p *ProductService) monitorProgress(stats *ProgressStats, done chan bool, progressDone chan bool) {
	ticker := time.NewTicker(1 * time.Minute) // Update every minute
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Print final status
			processed := atomic.LoadInt64(&stats.ProcessedCount)
			success := atomic.LoadInt64(&stats.SuccessCount)
			notFound := atomic.LoadInt64(&stats.NotFoundCount)
			errors := atomic.LoadInt64(&stats.ErrorCount)

			elapsed := time.Since(stats.StartTime)
			rate := float64(processed) / elapsed.Seconds()

			p.logger.Info("COMPLETED: %d/%d products (%.1f req/s) - Success: %d, 404: %d, Errors: %d, Time: %v",
				processed, stats.TotalProducts, rate, success, notFound, errors, elapsed)
			progressDone <- true
			return

		case <-ticker.C:
			processed := atomic.LoadInt64(&stats.ProcessedCount)
			success := atomic.LoadInt64(&stats.SuccessCount)
			notFound := atomic.LoadInt64(&stats.NotFoundCount)
			errors := atomic.LoadInt64(&stats.ErrorCount)

			if stats.TotalProducts > 0 {
				elapsed := time.Since(stats.StartTime)
				rate := float64(processed) / elapsed.Seconds()
				percentage := float64(processed) / float64(stats.TotalProducts) * 100

				// Calculate estimated time remaining
				var eta time.Duration
				if rate > 0 {
					remaining := stats.TotalProducts - processed
					eta = time.Duration(float64(remaining)/rate) * time.Second
				}

				if eta > 0 {
					p.logger.Info("Progress: %d/%d (%.1f%%) - Success: %d, 404: %d, Errors: %d - Rate: %.1f req/s - ETA: %v",
						processed, stats.TotalProducts, percentage, success, notFound, errors, rate, eta)
				} else {
					p.logger.Info("Progress: %d/%d (%.1f%%) - Success: %d, 404: %d, Errors: %d - Rate: %.1f req/s",
						processed, stats.TotalProducts, percentage, success, notFound, errors, rate)
				}
			}
		}
	}
}

// fetchSingleProductData fetches detailed product information for a single product ID.
// It handles HTTP requests, gzip decompression, JSON parsing, and error handling.
// The result is sent through the resultChan for collection by the main process.
func (p *ProductService) fetchSingleProductData(productID int, resultChan chan<- ProductResult, stats *ProgressStats) {
	result := ProductResult{
		ProductID: productID,
	}

	// Create request with headers
	url := fmt.Sprintf("https://www.compraonline.bonpreuesclat.cat/api/webproductpagews/v5/products/bop?retailerProductId=%d", productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request for product %d: %w", productID, err)
		resultChan <- result
		return
	}

	// Set headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "ca-ES,ca;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15")

	// Make the request
	resp, err := p.client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch product %d: %w", productID, err)
		resultChan <- result
		return
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == 404 {
		result.Error = fmt.Errorf("product %d not found", productID)
		resultChan <- result
		return
	} else if resp.StatusCode != 200 {
		result.Error = fmt.Errorf("failed to fetch product %d, status code: %d", productID, resp.StatusCode)
		resultChan <- result
		return
	}

	// Read and decompress response body
	var reader io.Reader = resp.Body

	// Handle gzip compression
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			result.Error = fmt.Errorf("failed to create gzip reader for product %d: %w", productID, err)
			resultChan <- result
			return
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body for product %d: %w", productID, err)
		resultChan <- result
		return
	}

	// Parse JSON response
	var responseJSON map[string]interface{}
	if err := json.Unmarshal(body, &responseJSON); err != nil {
		result.Error = fmt.Errorf("failed to parse JSON for product %d: %w", productID, err)
		resultChan <- result
		return
	}

	// Parse product data using the model structure
	result.Product = models.ParseProductFromResponse(responseJSON, productID)
	result.NutritionalData = models.ParseNutritionalDataFromResponse(responseJSON, productID)

	resultChan <- result
}

// FetchSingleProductData fetches data for a single product synchronously.
// This is a convenience method for testing or when only one product is needed.
func (p *ProductService) FetchSingleProductData(productID int) (models.Product, []models.ProductNutritionalData, error) {
	resultChan := make(chan ProductResult, 1)

	go p.fetchSingleProductData(productID, resultChan, nil)

	result := <-resultChan
	return result.Product, result.NutritionalData, result.Error
}
