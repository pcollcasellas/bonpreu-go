package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bonpreu-go/pkg/models"
	"bonpreu-go/pkg/utils"
)

// SitemapService handles sitemap operations
type SitemapService struct {
	client *http.Client
	logger *utils.Logger
}

// NewSitemapService creates a new SitemapService instance
func NewSitemapService() *SitemapService {
	return &SitemapService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: utils.NewLogger("SitemapService"),
	}
}

// FetchProductIds fetches product IDs from the sitemap XML
func (s *SitemapService) FetchProductIds(sitemapURL string) ([]models.ItemIds, error) {
	start := time.Now()
	s.logger.Info("Starting to fetch product IDs from sitemap: %s", sitemapURL)

	// Make HTTP request to the sitemap
	resp, err := s.client.Get(sitemapURL)
	if err != nil {
		s.logger.Error("Failed to fetch sitemap: %v", err)
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.logger.Error("Failed to fetch URL list, status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch URL list, status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	s.logger.Info("Successfully downloaded sitemap, size: %d bytes", len(body))

	// Parse the XML
	var sitemap models.Sitemap
	if err := xml.Unmarshal(body, &sitemap); err != nil {
		s.logger.Error("Failed to parse XML: %v", err)
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	s.logger.Info("Found %d URLs in sitemap", len(sitemap.URLs))

	var itemIdsToInsert []models.ItemIds

	// Process each URL in the sitemap
	for _, urlEntry := range sitemap.URLs {
		if urlEntry.Loc == "" {
			continue
		}

		// URL decode the location
		decodedURL, err := url.QueryUnescape(urlEntry.Loc)
		if err != nil {
			// If decoding fails, use the original URL
			decodedURL = urlEntry.Loc
		}

		// Split the URL by "/" and get the last part as product ID
		parts := strings.Split(strings.TrimSuffix(decodedURL, "/"), "/")
		if len(parts) == 0 {
			continue
		}

		// Try to convert the last part to an integer
		productIDStr := parts[len(parts)-1]
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			// Skip if the last part is not a valid integer
			continue
		}

		itemIdsToInsert = append(itemIdsToInsert, models.ItemIds{ProductID: productID})
	}

	s.logger.Info("Successfully extracted %d product IDs", len(itemIdsToInsert))
	s.logger.LogDuration("FetchProductIds", start)

	return itemIdsToInsert, nil
}
