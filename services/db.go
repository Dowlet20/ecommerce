package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"path/filepath"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"Dowlet_projects/ecommerce/models"
)

// DBService handles database operations
type DBService struct {
	db *sql.DB
}

// ThumbnailData represents data for a thumbnail to be inserted
type ThumbnailData struct {
	ProductID int
	Color     string
	ImageURL  string
}


// ThumbnailWithProduct represents a thumbnail with its product details
type ThumbnailWithProduct struct {
	ThumbnailID int     `json:"thumbnail_id"`
	ProductID   int     `json:"product_id"`
	Color       string  `json:"color"`
	ImageURL    string  `json:"image_url"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Discount    float64 `json:"discount"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"` // Changed to string
}

// NewDBService creates a new database service
func NewDBService(user, password, dbname string) (*DBService, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", user, password, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	return &DBService{db: db}, nil
}

// Close closes the database connection
func (s *DBService) Close() {
	s.db.Close()
}

// SaveVerificationCode stores a verification code with registration data
func (s *DBService) SaveVerificationCode(phone, code, fullName string) error {
	expiresAt := time.Now().Add(5 * time.Minute)
	_, err := s.db.Exec(
		`INSERT INTO verification_codes (phone, code, expires_at, full_name)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE code = ?, expires_at = ?, full_name = ?`,
		phone, code, expiresAt, fullName,
		code, expiresAt, fullName)
	if err != nil {
		return fmt.Errorf("failed to store code: %v", err)
	}
	return nil
}

func (s *DBService) GetVerificationCode(phone string) (string, time.Time, string, error) {
    var code, fullName string
    var expiresAtStr string // Temporary string to hold the expires_at value
    err := s.db.QueryRow(
        "SELECT code, expires_at, full_name FROM verification_codes WHERE phone = ?",
        phone).Scan(&code, &expiresAtStr, &fullName)
    if err != nil {
        return "", time.Time{}, "", err
    }

    // Parse the expires_at string into time.Time
    expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr) // Adjust format as needed
    if err != nil {
        return "", time.Time{}, "", fmt.Errorf("failed to parse expires_at: %w", err)
    }

    return code, expiresAt, fullName, nil
}

// DeleteVerificationCode deletes a verification code
func (s *DBService) DeleteVerificationCode(phone string) error {
	_, err := s.db.Exec("DELETE FROM verification_codes WHERE phone = ?", phone)
	return err
}

// SaveUser saves a new user
func (s *DBService) SaveUser(fullName, phone string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (full_name, phone, verified) VALUES (?, ?, ?)",
		fullName, phone, true)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetUserByPhone retrieves a user by phone number
func (s *DBService) GetUserByPhone(phone string) (int, error) {
	var userID int
	err := s.db.QueryRow("SELECT id FROM users WHERE phone = ?", phone).Scan(&userID)
	fmt.Println(userID)
	return userID, err
}

// GetMarkets retrieves all markets
func (s *DBService) GetMarkets() ([]models.Market, error) {
	rows, err := s.db.Query("SELECT id, name, thumbnail_url FROM markets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []models.Market =[]models.Market{}
	for rows.Next() {
		var m models.Market
		if err := rows.Scan(&m.ID, &m.Name, &m.ThumbnailURL); err != nil {
			return nil, err
		}
		markets = append(markets, m)
	}
	return markets, nil
}

// GetMarketProducts retrieves products for a specific market
func (s *DBService) GetMarketProducts(marketID string) ([]models.Product, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.market_id, m.name as market_name, p.name, p.price, p.discount, p.description, p.created_at, 
		       IF(f.user_id IS NOT NULL, true, false) as is_favorite
		FROM products p LEFT JOIN markets m on p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id
		WHERE p.market_id = ?`, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, err
		}
		p.Thumbnails, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// GetPaginatedProducts retrieves products with pagination and optional name search
func (s *DBService) GetPaginatedProducts(page, limit int, search string) ([]models.Product, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.market_id, m.name as market_name, p.name, p.price, 
		p.discount, p.description, p.created_at, 
		IF(f.user_id IS NOT NULL, true, false) as is_favorite 
		FROM products p LEFT JOIN markets m on p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id WHERE 1=1`

	args := []interface{}{}

	if search != "" {
		query += " AND LOWER(p.name) LIKE ?"
		args = append(args, "%"+strings.ToLower(search)+"%")
	}

	query += " ORDER BY p.id LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %v", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.Thumbnails, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details: %v", err)
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProduct retrieves a single product by ID
func (s *DBService) GetProduct(id string) (models.Product, error) {
	var p models.Product
	err := s.db.QueryRow(`
		SELECT p.id, p.market_id, m.name as market_name, p.name, p.price, p.discount, p.description, p.created_at, 
		       IF(f.user_id IS NOT NULL, true, false) as is_favorite
		FROM products p LEFT JOIN markets m on p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id
		WHERE p.id = ?`, id).Scan(&p.ID, &p.MarketID, &p.MarketName, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite)
	if err != nil {
		return p, err
	}

	p.Thumbnails,  err = s.getProductDetails(p.ID)
	return p, err
}

// getProductDetails retrieves thumbnails and sizes for a product
func (s *DBService) getProductDetails(productID int) ([]models.Thumbnail, error) {
	thumbRows, err := s.db.Query("SELECT id, product_id, color, image_url FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return nil, err
	}
	defer thumbRows.Close()

	var thumbnails []models.Thumbnail
	for thumbRows.Next() {
		var sizes []models.Size
		var t models.Thumbnail
		if err := thumbRows.Scan(&t.ID, &t.ProductID, &t.Color, &t.ImageURL); err != nil {
			return nil,  err
		}
		sizeRows, err := s.db.Query("SELECT id, thumbnail_id, size, count FROM sizes WHERE thumbnail_id = ?", t.ID)
		if err != nil {
			return nil, err
		}
		defer sizeRows.Close()

		for sizeRows.Next() {
			var s models.Size
			if err := sizeRows.Scan(&s.ID, &s.ThumbnailID, &s.Size, &s.Count); err != nil {
				return nil, err
			}
			sizes = append(sizes, s)
			t.Sizes = sizes
		}
		thumbnails = append(thumbnails, t)
	}
	return thumbnails,  nil
}

// CreateProduct creates a new product 
func (s *DBService) CreateProduct(marketID, name, price, discount, description string) (int, error) {
	marketIDInt, err := strconv.Atoi(marketID)
	if err != nil {
		return 0, fmt.Errorf("invalid market ID: %v", err)
	}

	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price: %v", err)
	}

	var discountFloat float64
	if discount != "" {
		discountFloat, err = strconv.ParseFloat(discount, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid discount: %v", err)
		}
	}

	result, err := s.db.Exec(`
		INSERT INTO products (market_id, name, price, discount, description)
		VALUES (?, ?, ?, ?, ?)`,
		marketIDInt, name, priceFloat, discountFloat, description)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %v", err)
	}

	productID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve product ID: %v", err)
	}

	return int(productID), nil
}

// UpdateMarketThumbnail updates the thumbnail for a market
func (s *DBService) UpdateMarketThumbnail(marketID, thumbnailURL string) error {
	_, err := s.db.Exec("UPDATE markets SET thumbnail_url = ? WHERE id = ?", thumbnailURL, marketID)
	return err
}

// GetProductThumbnails retrieves thumbnails for a product
func (s *DBService) GetProductThumbnails(productID string) ([]models.Thumbnail, error) {
	rows, err := s.db.Query("SELECT id, product_id, color, image_url FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var thumbnails []models.Thumbnail
	for rows.Next() {
		var t models.Thumbnail
		if err := rows.Scan(&t.ID, &t.ProductID, &t.Color, &t.ImageURL); err != nil {
			return nil, err
		}
		thumbnails = append(thumbnails, t)
	}
	return thumbnails, nil
}

// DeleteProduct deletes a product and its associated thumbnails and sizes
func (s *DBService) DeleteProduct(productID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM sizes WHERE thumbnail_id IN (SELECT id FROM thumbnails WHERE product_id = ?)", productID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateMarket creates a new market with the given name and thumbnail URL
func (s *DBService) CreateMarket(name, thumbnailURL string) (int64, error) {
	result, err := s.db.Exec("INSERT INTO markets (name, thumbnail_url) VALUES (?, ?)", name, thumbnailURL)
	if err != nil {
		return 0, fmt.Errorf("failed to create market: %v", err)
	}
	marketID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve market ID: %v", err)
	}
	return marketID, nil
}

// DeleteMarket deletes a market by ID and its associated thumbnail file
func (s *DBService) DeleteMarket(marketID string) error {
	// Get thumbnail URL for file deletion
	var thumbnailURL string
	err := s.db.QueryRow("SELECT thumbnail_url FROM markets WHERE id = ?", marketID).Scan(&thumbnailURL)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to retrieve market thumbnail: %v", err)
	}

	// Delete market (assumes ON DELETE CASCADE for products, thumbnails, sizes)
	_, err = s.db.Exec("DELETE FROM markets WHERE id = ?", marketID)
	if err != nil {
		return fmt.Errorf("failed to delete market: %v", err)
	}

	// Delete thumbnail file if it exists
	if thumbnailURL != "" {
		filePath := filepath.Join(".", thumbnailURL[1:]) // Remove leading '/'
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting file %s: %v\n", filePath, err)
		}
	}

	return nil
}


// CreateThumbnails creates multiple thumbnails for a product
func (s *DBService) CreateThumbnails(thumbnails []ThumbnailData) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	for _, thumb := range thumbnails {
		_, err := tx.Exec("INSERT INTO thumbnails (product_id, color, image_url) VALUES (?, ?, ?)",
			thumb.ProductID, thumb.Color, thumb.ImageURL)
		if err != nil {
			return fmt.Errorf("failed to insert thumbnail: %v", err)
		}
	}

	return tx.Commit()
}


// DeleteThumbnail deletes a thumbnail by its ID
func (s *DBService) DeleteThumbnail(thumbnailID string) error {
	// Get image_url to delete the file
	var imageURL string
	err := s.db.QueryRow("SELECT image_url FROM thumbnails WHERE id = ?", thumbnailID).Scan(&imageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("thumbnail not found")
		}
		return fmt.Errorf("failed to retrieve thumbnail: %v", err)
	}

	// Delete thumbnail from database
	result, err := s.db.Exec("DELETE FROM thumbnails WHERE id = ?", thumbnailID)
	if err != nil {
		return fmt.Errorf("failed to delete thumbnail: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("thumbnail not found")
	}

	// Delete file from uploads
	if imageURL != "" {
		filePath := filepath.Join(".", imageURL)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	}

	return nil
}


// GetAllThumbnailsWithProducts retrieves all thumbnails with associated product information
func (s *DBService) GetAllThumbnailsWithProducts() ([]ThumbnailWithProduct, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.product_id, t.color, t.image_url, 
		       p.name, p.price, p.discount, p.description, p.created_at
		FROM thumbnails t
		JOIN products p ON t.product_id = p.id`)
	if err != nil {
		return nil, fmt.Errorf("failed to query thumbnails: %v", err)
	}
	defer rows.Close()

	var thumbnails []ThumbnailWithProduct
	for rows.Next() {
		var t ThumbnailWithProduct
		if err := rows.Scan(&t.ThumbnailID, &t.ProductID, &t.Color, &t.ImageURL,
			&t.ProductName, &t.Price, &t.Discount, &t.Description, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan thumbnail: %v", err)
		}
		thumbnails = append(thumbnails, t)
	}

	return thumbnails, nil
}


// CreateSizeByThumbnailID creates a single size with count linked to a thumbnail
func (s *DBService) CreateSizeByThumbnailID(thumbnailID, size string, count int) error {
	if size == "" {
		return fmt.Errorf("size cannot be empty")
	}
	if count < 0 {
		return fmt.Errorf("count cannot be negative")
	}

	// Validate thumbnail_id exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM thumbnails WHERE id = ?)", thumbnailID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate thumbnail: %v", err)
	}
	if !exists {
		return fmt.Errorf("thumbnail not found")
	}

	thumbnailIDInt, err := strconv.Atoi(thumbnailID)
	if err != nil {
		return fmt.Errorf("invalid thumbnail ID: %v", err)
	}

	_, err = s.db.Exec("INSERT INTO sizes (thumbnail_id, size, count) VALUES (?, ?, ?)",
		thumbnailIDInt, size, count)
	if err != nil {
		return fmt.Errorf("failed to insert size: %v", err)
	}

	return nil
}



// DeleteSizeByID deletes a size by its ID
func (s *DBService) DeleteSizeByID(sizeID string) error {
	result, err := s.db.Exec("DELETE FROM sizes WHERE id = ?", sizeID)
	if err != nil {
		return fmt.Errorf("failed to delete size: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("size not found")
	}

	return nil
}




// GetUserFavoriteProducts retrieves paginated favorite products for a user
func (s *DBService) GetUserFavoriteProducts(userID int, page, limit int) ([]models.Product, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.market_id, m.name as market_name, p.name, p.price, p.discount, p.description, p.created_at, 
		       true as is_favorite
		FROM products p INNER JOIN markets m on p.market_id = m.id 
		INNER JOIN favorites f ON p.id = f.product_id
		WHERE f.user_id = ?
		ORDER BY p.id LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query favorite products: %v", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.Thumbnails, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details: %v", err)
		}
		products = append(products, p)
	}

	return products, nil
}


// ToggleFavoriteProduct adds or removes a product from the user's favorites
func (s *DBService) ToggleFavoriteProduct(userID, productID int) (bool, error) {
	// Check if product exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to validate product: %v", err)
	}
	if !exists {
		return false, fmt.Errorf("product not found")
	}

	// Check if already favorited
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND product_id = ?)", userID, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite status: %v", err)
	}

	if exists {
		// Remove from favorites
		_, err = s.db.Exec("DELETE FROM favorites WHERE user_id = ? AND product_id = ?", userID, productID)
		if err != nil {
			return false, fmt.Errorf("failed to remove favorite: %v", err)
		}
		return false, nil // is_favorite = false
	}

	// Add to favorites
	_, err = s.db.Exec("INSERT INTO favorites (user_id, product_id) VALUES (?, ?)", userID, productID)
	if err != nil {
		return false, fmt.Errorf("failed to add favorite: %v", err)
	}
	return true, nil // is_favorite = true
}


