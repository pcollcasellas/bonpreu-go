# Bonpreu Go Application

This Go application fetches product data from the Bonpreu sitemap XML and stores it in a PostgreSQL database.

## Features

- Fetches product IDs from the Bonpreu sitemap
- Parses XML sitemap data
- Extracts product IDs from URLs
- Asynchronously fetches product data from API endpoints
- Rate limiting to be respectful to servers
- Stores data in PostgreSQL database (Neon)
- Progress tracking with visual progress bar
- Structured logging and monitoring
- Clean architecture with separation of concerns

## Prerequisites

- Go 1.21 or later
- PostgreSQL database (Neon recommended)
- Neon account (free tier available)

## Database Setup

### 1. Create Neon Database

1. Sign up at [neon.tech](https://neon.tech)
2. Create a new project
3. Note your connection details:
   - Host
   - Port (usually 5432)
   - Username
   - Password
   - Database name

### 2. Run Database Schema

Execute the SQL script to create the necessary tables:

```bash
# Connect to your Neon database and run:
psql "postgresql://username:password@host:port/database?sslmode=require" -f scripts/schema.sql
```

Or copy and paste the contents of `scripts/schema.sql` into your Neon SQL editor.

### 3. Configure Environment Variables

Copy the example environment file and update it with your Neon credentials:

```bash
cp env.example .env
```

Edit the `.env` file with your Neon database credentials:

```bash
# Database Configuration (Neon PostgreSQL)
DB_HOST=your-neon-host.neon.tech
DB_PORT=5432
DB_USER=your-username
DB_PASSWORD=your-password
DB_NAME=your-database-name
DB_SSL_MODE=require
```

**Important**: Never commit your `.env` file to version control. It's already in `.gitignore`.

## Installation

1. Clone or download this repository
2. Navigate to the project directory
3. Run the following command to download dependencies:

```bash
go mod tidy
```

## Usage

### Quick Start

Run the application using Go directly:

```bash
go run cmd/bonpreu/main.go
```

Or use the Makefile for convenience:

```bash
make run
```

### Configuration

### Environment Variables

The application uses environment variables for configuration. Copy `env.example` to `.env` and update the values:

```bash
cp env.example .env
```

**Available Environment Variables:**
- `SITEMAP_URL`: Bonpreu sitemap URL
- `REQUEST_DURATION_MINUTES`: Rate limiting duration in minutes
- `HTTP_TIMEOUT_SECONDS`: HTTP client timeout
- `DB_HOST`: Database host (Neon host)
- `DB_PORT`: Database port (usually 5432)
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSL_MODE`: SSL mode (usually "require")

### Configuration Modes

The application supports two modes:

#### Production Mode (Rate Limited)
```go
cfg := config.DefaultConfig()    // Uses REQUEST_DURATION_MINUTES from env
```

#### Testing Mode (No Rate Limiting)
```go
cfg := config.TestingConfig()   // No rate limiting for testing
```

To switch modes, edit `cmd/bonpreu/main.go` and change the configuration line.

### Available Commands

#### Using Go directly:
```bash
# Run the application
go run cmd/bonpreu/main.go

# Build the application
go build -o bonpreu-go cmd/bonpreu/main.go

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Format code
go fmt ./...
```

#### Using Makefile:
```bash
# Show all available commands
make help

# Run the application
make run

# Build the application
make build

# Install dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Clean build artifacts
make clean
```

## What the application does:

1. Fetch the sitemap from Bonpreu's website
2. Parse the XML to extract product URLs
3. Extract product IDs from the URLs
4. Asynchronously fetch product data from API endpoints
5. Store products and nutritional data in PostgreSQL database
6. Provide detailed logging and progress tracking throughout the process

## Database Schema

### Products Table
- `product_id` (PRIMARY KEY): Unique product identifier
- `product_type`: Type of product
- `product_name`: Product name
- `product_description`: Product description
- `product_brand`: Brand name
- `product_pack_size_description`: Package size description
- `product_price_amount`: Price amount
- `product_currency`: Currency code
- `product_unit_price_amount`: Unit price amount
- `product_unit_price_currency`: Unit price currency
- `product_unit_price_unit`: Unit price unit
- `product_available`: Availability status
- `product_alcohol`: Alcohol content flag
- `product_cooking_guidelines`: Cooking instructions
- `product_categories`: Array of category strings
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

### Product Nutritional Data Table
- `id` (PRIMARY KEY): Auto-incrementing ID
- `product_id` (FOREIGN KEY): Reference to products table
- `product_nutritional_value`: Nutritional value name
- `product_nutritional_quantity`: Nutritional quantity
- `created_at`: Creation timestamp

## Project Structure

```
bonpreu-go/
├── cmd/
│   └── bonpreu/
│       └── main.go          # Application entry point
├── pkg/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── models/
│   │   ├── item.go          # Sitemap data structures
│   │   └── product.go       # Product data structures
│   ├── services/
│   │   ├── sitemap_service.go    # Sitemap fetching
│   │   ├── product_service.go    # Product data fetching
│   │   └── database_service.go   # Database operations
│   └── utils/
│       └── logger.go        # Logging utilities
├── scripts/
│   └── schema.sql          # Database schema
├── go.mod                  # Go module dependencies
├── Makefile               # Build and run commands
└── README.md             # This file
```

## Architecture

The application follows clean architecture principles:

- **cmd/**: Contains the main application entry points
- **pkg/config/**: Configuration management
- **pkg/models/**: Data structures and domain models
- **pkg/services/**: Business logic and external service interactions
- **pkg/utils/**: Utility functions and helpers
- **scripts/**: Database and deployment scripts

## Performance Features

- **Asynchronous Processing**: Uses Go goroutines for concurrent API requests
- **Rate Limiting**: Configurable rate limiting to be respectful to servers
- **Progress Tracking**: Real-time progress bar with statistics
- **Database Transactions**: Efficient batch inserts with transaction support
- **Connection Pooling**: Optimized database connections

## Error Handling

- Graceful handling of 404 errors (products not found)
- Network timeout handling
- Database connection error recovery
- Comprehensive logging throughout the process

## Monitoring

The application provides detailed logging including:
- Request rates and progress
- Success/error statistics
- Database operation timing
- Overall execution duration 