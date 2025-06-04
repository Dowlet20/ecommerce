package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Dowlet_projects/ecommerce/models"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	rows, err := s.db.Query("SELECT id, name, thumbnail_url, username, full_name, phone FROM markets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []models.Market = []models.Market{}
	for rows.Next() {
		var m models.Market
		if err := rows.Scan(&m.ID, &m.Name, &m.ThumbnailURL, &m.Username, &m.FullName, &m.Phone); err != nil {
			return nil, err
		}
		markets = append(markets, m)
	}
	return markets, nil
}
// GetMarketProducts retrieves paginated products for a market
func (s *DBService) GetMarketProducts(marketID, page, limit int) ([]models.Product, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 {
        limit = 10
    }
    offset := (page - 1) * limit

    query := `
        SELECT p.id, p.market_id, m.name, p.category_id, p.name, p.price, p.discount, p.description, p.created_at
        FROM products p
        INNER JOIN markets m ON p.market_id = m.id
        WHERE p.market_id = ?
        ORDER BY p.id LIMIT ? OFFSET ?`

    rows, err := s.db.Query(query, marketID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to query products: %v", err)
    }
    defer rows.Close()

    var products []models.Product
    for rows.Next() {
        var p models.Product
        var createdAtStr string
        if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.CategoryID, &p.Name, &p.Price, &p.Discount, &p.Description, &createdAtStr); err != nil {
            return nil, fmt.Errorf("failed to scan product: %v", err)
        }
        p.CreatedAt = createdAtStr
        p.Thumbnails, err = s.getProductDetails(p.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get product details: %v", err)
        }
        products = append(products, p)
    }

    return products, nil
}
// GetMarketByID retrieves a market and its products by market ID
func (s *DBService) GetMarketByID(marketID int) (*models.Market, []models.Product, error) {
    var market models.Market
    err := s.db.QueryRow(`
        SELECT id, username, full_name, phone, name, thumbnail_url
        FROM markets WHERE id = ?`, marketID).
        Scan(&market.ID, &market.Username, &market.FullName, &market.Phone, &market.Name, &market.ThumbnailURL)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil, fmt.Errorf("market not found")
        }
        return nil, nil, fmt.Errorf("failed to query market: %v", err)
    }

    products, err := s.GetMarketProducts(marketID, 1, 1000)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to query products: %v", err)
    }

    return &market, products, nil
}


// UpdateProduct updates a product
func (s *DBService) UpdateProduct(marketID, productID int, name string, price float64, discount float64, description string) error {
	// priceFloat, err := strconv.ParseFloat(price, 64)
	// if err != nil {
	// 	return fmt.Errorf("invalid price: %v", err)
	// }

	// var discountFloat float64
	// if discount != "" {
	// 	discountFloat, err = strconv.ParseFloat(discount, 64)
	// 	if err != nil {
	// 		return fmt.Errorf("invalid discount: %v", err)
	// 	}
	// }

	// Verify product exists and belongs to market
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ? AND market_id = ?)", productID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate product: %v", err)
	}
	if !exists {
		return fmt.Errorf("product not found or unauthorized")
	}

	_, err = s.db.Exec(`
		UPDATE products 
		SET name = ?, price = ?, discount = ?, description = ?
		WHERE id = ? AND market_id = ?`,
		name, price, discount, description, productID, marketID)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}
// GetPaginatedProducts retrieves products with pagination, optional category, and name search
func (s *DBService) GetPaginatedProducts(categoryID, page, limit int, search string) ([]models.Product, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.market_id, m.name as market_name, p.category_id, p.name, p.price, 
		p.discount, p.description, p.created_at, 
		IF(f.user_id IS NOT NULL, true, false) as is_favorite 
		FROM products p 
		LEFT JOIN markets m ON p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id 
		WHERE 1=1`

	args := []interface{}{}

	if categoryID != 0 {
		query += " AND p.category_id = ?"
		args = append(args, categoryID)
	}

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
		var createdAtStr string
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.CategoryID, &p.Name, &p.Price, &p.Discount, &p.Description, &createdAtStr, &p.IsFavorite); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.CreatedAt = createdAtStr
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
func (s *DBService) CreateProduct(marketID, categoryID int, name string, price float64, discount float64, description string) (int, error) {
    // priceFloat, err := strconv.ParseFloat(price, 64)
    // if err != nil {
    //     return 0, fmt.Errorf("invalid price: %v", err)
    // }

    // var discountFloat float64
    // if discount != "" {
    //     discountFloat, err = strconv.ParseFloat(discount, 64)
    //     if err != nil {
    //         return 0, fmt.Errorf("invalid discount: %v", err)
    //     }
    // }

    // Verify market exists
    var exists bool
    err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM markets WHERE id = ?)", marketID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to validate market: %v", err)
    }
    if !exists {
        return 0, fmt.Errorf("market not found")
    }

    // Verify category exists
    err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", categoryID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to validate category: %v", err)
    }
    if !exists {
        return 0, fmt.Errorf("category not found")
    }

    result, err := s.db.Exec(`
        INSERT INTO products (market_id, category_id, name, price, discount, description)
        VALUES (?, ?, ?, ?, ?, ?)`,
        marketID, categoryID, name, price, discount, description)
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
// DeleteProduct deletes a product and its thumbnails
func (s *DBService) DeleteProduct(marketID, productID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Verify product exists and belongs to market
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ? AND market_id = ?)", productID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate product: %v", err)
	}
	if !exists {
		return fmt.Errorf("product not found or unauthorized")
	}

	// Fetch thumbnails
	rows, err := tx.Query("SELECT image_url FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return fmt.Errorf("failed to query thumbnails: %v", err)
	}
	defer rows.Close()

	var imageURLs []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return fmt.Errorf("failed to scan thumbnail: %v", err)
		}
		if imageURL != "" {
			imageURLs = append(imageURLs, imageURL)
		}
	}

	// Delete thumbnails (sizes are deleted via CASCADE)
	_, err = tx.Exec("DELETE FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return fmt.Errorf("failed to delete thumbnails: %v", err)
	}

	// Delete product
	result, err := tx.Exec("DELETE FROM products WHERE id = ? AND market_id = ?", productID, marketID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("product not found or unauthorized")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Delete thumbnail files
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	for _, imageURL := range imageURLs {
		filePath := filepath.Join(uploadDir, filepath.Base(imageURL))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to delete file %s: %v\n", filePath, err)
		}
	}

	return nil
}

// CreateMarket creates a market
func (s *DBService) CreateMarket(name, thumbnailURL, username, fullName, phone, password string, deliveryPrice float64) (string, string, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM markets WHERE username = ? OR phone = ?)", username, phone).Scan(&exists)
	if err != nil {
		return "", "", fmt.Errorf("failed to check username/phone: %v", err)
	}
	if exists {
		return "", "", fmt.Errorf("username or phone already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %v", err)
	}

	result, err := s.db.Exec(`
		INSERT INTO markets (username, password, full_name, phone, name, thumbnail_url, delivery_price)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, passwordHash, fullName, phone, name, thumbnailURL, deliveryPrice)
	if err != nil {
		return "", "", fmt.Errorf("failed to create market: %v", err)
	}

	marketID, err := result.LastInsertId()
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve market ID: %v", err)
	}

	return username, strconv.Itoa(int(marketID)), nil
}

// AuthenticateMarket authenticates a market admin
func (s *DBService) AuthenticateMarket(username, password string) (int, int, error) {
	var userID, marketID int
	var passwordHash string
	err := s.db.QueryRow("SELECT id, id, password FROM markets WHERE username = ?", username).Scan(&userID, &marketID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, fmt.Errorf("invalid credentials")
		}
		return 0, 0, fmt.Errorf("failed to query market: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return 0, 0, fmt.Errorf("invalid credentials")
	}

	return userID, marketID, nil
}

// AuthenticateSuperadmin authenticates a superadmin
func (s *DBService) AuthenticateSuperadmin(username, password string) (int, error) {
	var userID int
	var passwordHash string
	err := s.db.QueryRow("SELECT id, password FROM superadmins WHERE username = ?", username).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid credentials")
		}
		return 0, fmt.Errorf("failed to query superadmin: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return 0, fmt.Errorf("invalid credentials")
	}

	return userID, nil
}

// RegisterSuperadmin registers a new superadmin
func (s *DBService) RegisterSuperadmin(username, fullName, phone, password string) (int, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM superadmins WHERE username = ? OR phone = ?)", username, phone).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to check username/phone: %v", err)
	}
	if exists {
		return 0, fmt.Errorf("username or phone already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %v", err)
	}

	result, err := s.db.Exec(`
		INSERT INTO superadmins (username, full_name, phone, password)
		VALUES (?, ?, ?, ?)`,
		username, fullName, phone, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("failed to create superadmin: %v", err)
	}

	superadminID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve superadmin ID: %v", err)
	}

	return int(superadminID), nil
}

// RegisterUser registers a user
func (s *DBService) RegisterUser(fullName, phone string) (int, string, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE phone = ?)", phone).Scan(&exists)
	if err != nil {
		return 0, "", fmt.Errorf("failed to check phone: %v", err)
	}
	if exists {
		return 0, "", fmt.Errorf("phone number already registered")
	}

	otp := generateOTP(4)
	expiresAt := time.Now().Add(5 * time.Minute)

	result, err := s.db.Exec("INSERT INTO verification_codes (phone, code, expires_at, full_name) VALUES (?, ?, ?, ?)",
		phone, otp, expiresAt, fullName)
	if err != nil {
		return 0, "", fmt.Errorf("failed to store verification: %v", err)
	}

	verifID, err := result.LastInsertId()
	if err != nil {
		return 0, "", fmt.Errorf("failed to retrieve verification ID: %v", err)
	}

	return int(verifID), otp, nil
}

// VerifyUserOTP verifies OTP and creates/updates a user
func (s *DBService) VerifyUserOTP(phone, otp string) (int, error) {
	var code string
	var expiresAt time.Time
	var fullName string
	err := s.db.QueryRow("SELECT code, expires_at, full_name FROM verification_codes WHERE phone = ?", phone).Scan(&code, &expiresAt, &fullName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no verification code found")
		}
		return 0, fmt.Errorf("failed to query verification code: %v", err)
	}

	if code != otp || time.Now().After(expiresAt) {
		return 0, fmt.Errorf("invalid or expired OTP")
	}

	var userID int
	err = s.db.QueryRow("SELECT id FROM users WHERE phone = ?", phone).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check user: %v", err)
	}

	if err == sql.ErrNoRows {
		result, err := s.db.Exec("INSERT INTO users (full_name, phone, verified) VALUES (?, ?, 1)", fullName, phone)
		if err != nil {
			return 0, fmt.Errorf("failed to create user: %v", err)
		}
		userID64, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve user ID: %v", err)
		}
		userID = int(userID64)
	} else {
		_, err = s.db.Exec("UPDATE users SET verified = 1 WHERE id = ?", userID)
		if err != nil {
			return 0, fmt.Errorf("failed to update user: %v", err)
		}
	}

	_, err = s.db.Exec("DELETE FROM verification_codes WHERE phone = ?", phone)
	if err != nil {
		return 0, fmt.Errorf("failed to clear verification code: %v", err)
	}

	return userID, nil
}

// DeleteMarket deletes a market, its products, and thumbnails
func (s *DBService) DeleteMarket(marketID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Verify market exists
	var marketThumbnailURL string
	err = tx.QueryRow("SELECT thumbnail_url FROM markets WHERE id = ?", marketID).Scan(&marketThumbnailURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("market not found")
		}
		return fmt.Errorf("failed to validate market: %v", err)
	}

	// Fetch all products and their thumbnails
	rows, err := tx.Query(`
		SELECT t.image_url
		FROM thumbnails t
		JOIN products p ON t.product_id = p.id
		WHERE p.market_id = ?`, marketID)
	if err != nil {
		return fmt.Errorf("failed to query thumbnails: %v", err)
	}
	defer rows.Close()

	var imageURLs []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return fmt.Errorf("failed to scan thumbnail: %v", err)
		}
		if imageURL != "" {
			imageURLs = append(imageURLs, imageURL)
		}
	}

	// Delete thumbnails (sizes are deleted via CASCADE)
	_, err = tx.Exec(`
		DELETE t FROM thumbnails t
		JOIN products p ON t.product_id = p.id
		WHERE p.market_id = ?`, marketID)
	if err != nil {
		return fmt.Errorf("failed to delete thumbnails: %v", err)
	}

	// Delete products
	_, err = tx.Exec("DELETE FROM products WHERE market_id = ?", marketID)
	if err != nil {
		return fmt.Errorf("failed to delete products: %v", err)
	}

	// Delete market
	result, err := tx.Exec("DELETE FROM markets WHERE id = ?", marketID)
	if err != nil {
		return fmt.Errorf("failed to delete market: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("market not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Delete thumbnail files
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	for _, imageURL := range imageURLs {
		filePath := filepath.Join(uploadDir, filepath.Base(imageURL))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to delete file %s: %v\n", filePath, err)
		}
	}

	// Delete market thumbnail file
	if marketThumbnailURL != "" {
		filePath := filepath.Join(uploadDir, filepath.Base(marketThumbnailURL))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to delete market thumbnail %s: %v\n", filePath, err)
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
func (s *DBService) DeleteThumbnail(marketID int, thumbnailID string) error {
	// Verify thumbnail exists and belongs to market's product
	var imageURL string
	err := s.db.QueryRow(`
		SELECT t.image_url 
		FROM thumbnails t 
		JOIN products p ON t.product_id = p.id 
		WHERE t.id = ? AND p.market_id = ?`, thumbnailID, marketID).Scan(&imageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("thumbnail not found or unauthorized")
		}
		return fmt.Errorf("failed to retrieve thumbnail: %v", err)
	}

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

// CreateSizeByThumbnailID creates a size linked to a thumbnail
func (s *DBService) CreateSizeByThumbnailID(marketID int, thumbnailID string, size string, count int) error {
	if size == "" {
		return fmt.Errorf("size cannot be empty")
	}
	if count < 0 {
		return fmt.Errorf("count cannot be negative")
	}

	// Verify thumbnail exists and belongs to market's product
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM thumbnails t 
			JOIN products p ON t.product_id = p.id 
			WHERE t.id = ? AND p.market_id = ?
		)`, thumbnailID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate thumbnail: %v", err)
	}
	if !exists {
		return fmt.Errorf("thumbnail not found or unauthorized")
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
func (s *DBService) DeleteSizeByID(marketID int, sizeID string) error {
	// Verify size exists and belongs to market's product
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM sizes s 
			JOIN thumbnails t ON s.thumbnail_id = t.id 
			JOIN products p ON t.product_id = p.id 
			WHERE s.id = ? AND p.market_id = ?
		)`, sizeID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate size: %v", err)
	}
	if !exists {
		return fmt.Errorf("size not found or unauthorized")
	}

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




// GetUserFavoriteProducts retrieves paginated favorite products
func (s *DBService) GetUserFavoriteProducts(userID, page, limit int) ([]models.Product, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.market_id, m.name, p.name, p.price, p.discount, p.description, p.created_at, true as is_favorite
		FROM products p
		INNER JOIN markets m ON p.market_id = m.id
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
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.Name, &p.Price, &p.Discount, &p.Description, &createdAt, &p.IsFavorite); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		p.Thumbnails, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details: %v", err)
		}
		products = append(products, p)
	}

	return products, nil
}

// ToggleFavoriteProduct adds or removes a favorite
func (s *DBService) ToggleFavoriteProduct(userID, productID int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to validate product: %v", err)
	}
	if !exists {
		return false, fmt.Errorf("product not found")
	}

	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND product_id = ?)", userID, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite status: %v", err)
	}

	if exists {
		_, err = s.db.Exec("DELETE FROM favorites WHERE user_id = ? AND product_id = ?", userID, productID)
		if err != nil {
			return false, fmt.Errorf("failed to remove favorite: %v", err)
		}
		return false, nil
	}

	_, err = s.db.Exec("INSERT INTO favorites (user_id, product_id) VALUES (?, ?)", userID, productID)
	if err != nil {
		return false, fmt.Errorf("failed to add favorite: %v", err)
	}
	return true, nil
}




// generateOTP generates a random OTP
func generateOTP(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		result, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			b[i] = digits[result.Int64()]
		}
	}
	return string(b)
}

// CreateCategory creates a new category
func (s *DBService) CreateCategory(name, thumbnailURL string) (int, error) {
    if name == "" {
        return 0, fmt.Errorf("category name is required")
    }

    result, err := s.db.Exec("INSERT INTO categories (name, thumbnail_url) VALUES (?, ?)", name, thumbnailURL)
    if err != nil {
        return 0, fmt.Errorf("failed to create category: %v", err)
    }

    categoryID, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to retrieve category ID: %v", err)
    }

    return int(categoryID), nil
}



// DeleteCategory deletes a category and its thumbnail
func (s *DBService) DeleteCategory(categoryID int) error {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Fetch thumbnail URL
    var thumbnailURL string
    err = tx.QueryRow("SELECT thumbnail_url FROM categories WHERE id = ?", categoryID).Scan(&thumbnailURL)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("category not found")
        }
        return fmt.Errorf("failed to retrieve category: %v", err)
    }

    // Delete category
    result, err := tx.Exec("DELETE FROM categories WHERE id = ?", categoryID)
    if err != nil {
        return fmt.Errorf("failed to delete category: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check rows affected: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("category not found")
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Delete thumbnail file
    if thumbnailURL != "" {
        uploadDir := os.Getenv("UPLOAD_DIR")
        if uploadDir == "" {
            uploadDir = "./uploads/categories"
        }
        filePath := filepath.Join(uploadDir, filepath.Base(thumbnailURL))
        if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
            fmt.Printf("Warning: failed to delete category thumbnail %s: %v\n", filePath, err)
        }
    }

    return nil
}


// GetCategories retrieves paginated categories with optional name search
func (s *DBService) GetCategories(page, limit int, search string) ([]models.Category, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, name, thumbnail_url
		FROM categories
		WHERE name LIKE ?
		ORDER BY id
		LIMIT ? OFFSET ?`
	searchParam := "%" + search + "%"

	rows, err := s.db.Query(query, searchParam, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %v", err)
	}
	defer rows.Close()

	var categories []models.Category = []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.ThumbnailURL); err != nil {
			return nil, fmt.Errorf("failed to scan category: %v", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}


// AddToCart adds or updates a product in the user's cart
func (s *DBService) AddToCart(userID, marketID, productID, thumbnailID, sizeID, count int) (int, error) {
    if count <= 0 {
        return 0, fmt.Errorf("count must be positive")
    }

    // Validate relationships
    var exists bool
    err := s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 
            FROM products p
            JOIN markets m ON p.market_id = m.id
            JOIN thumbnails t ON t.product_id = p.id
            JOIN sizes s ON s.thumbnail_id = t.id
            WHERE p.id = ? AND m.id = ? AND t.id = ? AND s.id = ?
        )`, productID, marketID, thumbnailID, sizeID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to validate cart entry: %v", err)
    }
    if !exists {
        return 0, fmt.Errorf("invalid market, product, thumbnail, or size")
    }

    tx, err := s.db.Begin()
    if err != nil {
        return 0, fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Check if cart entry exists
    var cartID int
    err = tx.QueryRow(`
        SELECT id FROM carts 
        WHERE user_id = ? AND market_id = ? AND product_id = ? AND thumbnail_id = ? AND size_id = ?`,
        userID, marketID, productID, thumbnailID, sizeID).Scan(&cartID)
    if err == sql.ErrNoRows {
        // Insert new cart entry
        result, err := tx.Exec(`
            INSERT INTO carts (user_id, market_id, product_id, thumbnail_id, size_id, count)
            VALUES (?, ?, ?, ?, ?, ?)`,
            userID, marketID, productID, thumbnailID, sizeID, count)
        if err != nil {
            return 0, fmt.Errorf("failed to insert cart entry: %v", err)
        }
        cartID64, err := result.LastInsertId()
        if err != nil {
            return 0, fmt.Errorf("failed to retrieve cart ID: %v", err)
        }
        cartID = int(cartID64)
    } else if err != nil {
        return 0, fmt.Errorf("failed to check cart entry: %v", err)
    } else {
        // Update existing cart entry
        _, err := tx.Exec(`
            UPDATE carts 
            SET count = count + ?
            WHERE id = ?`,
            count, cartID)
        if err != nil {
            return 0, fmt.Errorf("failed to update cart entry: %v", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return 0, fmt.Errorf("failed to commit transaction: %v", err)
    }

    return cartID, nil
}

// GetUserCart retrieves a user's cart grouped by markets
func (s *DBService) GetUserCart(userID int) ([]models.CartMarket, error) {
	query := `
		SELECT c.id, c.user_id, m.id, m.name, m.delivery_price, 
			p.id, p.name, p.price, p.discount, 
			t.id, t.thumbnail_url, t.color, 
			s.id, s.size, c.count
		FROM carts c
		JOIN markets m ON c.market_id = m.id
		JOIN products p ON c.product_id = p.id
		JOIN thumbnails t ON c.thumbnail_id = t.id
		JOIN sizes s ON c.size_id = s.id
		WHERE c.user_id = ?
		ORDER BY m.id, p.id`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart: %v", err)
	}
	defer rows.Close()

	// Group by market
	marketMap := make(map[int]*models.CartMarket)
	for rows.Next() {
		var cartID, userID, marketID, productID, thumbnailID, sizeID int
		var marketName, productName, thumbnailURL, color, size string
		var price, discount, deliveryPrice float64
		var count int

		if err := rows.Scan(&cartID, &userID, &marketID, &marketName, &deliveryPrice, 
			&productID, &productName, &price, &discount, 
			&thumbnailID, &thumbnailURL, &color, 
			&sizeID, &size, &count); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %v", err)
		}

		// Create or update market entry
		market, exists := marketMap[marketID]
		if !exists {
			market = &models.CartMarket{
				MarketName: marketName,
				UserID:     userID,
				CartID:     cartID,
				DeliveryPrice: deliveryPrice,
				Products:   []models.CartProduct{},
			}
			marketMap[marketID] = market
		}

		// Add product to market
		market.Products = append(market.Products, models.CartProduct{
			ThumbnailURL: thumbnailURL,
			Name:         productName,
			Price:        price,
			Discount:     discount,
			Color:        color,
			Size:         size,
			Count:        count,
		})
	}

	// Convert map to slice
	var cart []models.CartMarket
	for _, market := range marketMap {
		cart = append(cart, *market)
	}

	return cart, nil
}
