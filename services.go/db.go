package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Dowlet-projects/ecommerce/models"
	_ "github.com/go-sql-driver/mysql"
)

// DBService handles database operations
type DBService struct {
	db *sql.DB
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

// SaveUserAndOTP saves a user and generates an OTP
func (s *DBService) SaveUserAndOTP(user models.User, otp string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Save user
	result, err := tx.Exec("INSERT INTO users (full_name, phone, password, verified) VALUES (?, ?, ?, ?)",
		user.FullName, user.Phone, user.Password, false)
	if err != nil {
		return err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Save OTP
	_, err = tx.Exec("INSERT INTO otps (user_id, phone, code, expires_at) VALUES (?, ?, ?, ?)",
		userID, user.Phone, otp, time.Now().Add(5*time.Minute))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// VerifyOTP verifies the OTP for a phone number
func (s *DBService) VerifyOTP(phone, otp string) error {
	var userID int
	var expiresAt time.Time
	err := s.db.QueryRow("SELECT user_id, expires_at FROM otps WHERE phone = ? AND code = ?", phone, otp).Scan(&userID, &expiresAt)
	if err != nil {
		return fmt.Errorf("invalid OTP")
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("OTP expired")
	}

	_, err = s.db.Exec("UPDATE users SET verified = true WHERE id = ?", userID)
	if err != nil {
		return err
	}

	return nil
}

// AuthenticateUser checks user credentials and verification status
func (s *DBService) AuthenticateUser(phone, password string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow("SELECT id, full_name, phone, verified FROM users WHERE phone = ? AND password = ?",
		phone, password).Scan(&user.ID, &user.FullName, &user.Phone, &user.Verified)
	if err != nil {
		return user, err
	}
	if !user.Verified {
		return user, fmt.Errorf("phone not verified")
	}
	return user, nil
}

// GetMarkets retrieves all markets
func (s *DBService) GetMarkets() ([]models.Market, error) {
	rows, err := s.db.Query("SELECT id, name, thumbnail_url FROM markets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []models.Market
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
		SELECT p.id, p.market_id, p.name, p.price, p.discount, p.description, p.created_at, 
		       IF(f.user_id IS NOT NULL, true, false) as is_favorite
		FROM products p
		LEFT JOIN favorites f ON p.id = f.product_id
		WHERE p.market_id = ?`, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, err
		}
		p.Thumbnails, p.Sizes, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// GetAllProducts retrieves all products
func (s *DBService) GetAllProducts() ([]models.Product, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.market_id, p.name, p.price, p.discount, p.description, p.created_at, 
		       IF(f.user_id IS NOT NULL, true, false) as is_favorite
		FROM products p
		LEFT JOIN favorites f ON p.id = f.product_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite); err != nil {
			return nil, err
		}
		p.Thumbnails, p.Sizes, err = s.getProductDetails(p.ID)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// GetProduct retrieves a single product by ID
func (s *DBService) GetProduct(id string) (models.Product, error) {
	var p models.Product
	err := s.db.QueryRow(`
		SELECT p.id, p.market_id, p.name, p.price, p.discount, p.description, p.created_at, 
		       IF(f.user_id IS NOT NULL, true, false) as is_favorite
		FROM products p
		LEFT JOIN favorites f ON p.id = f.product_id
		WHERE p.id = ?`, id).Scan(&p.ID, &p.MarketID, &p.Name, &p.Price, &p.Discount, &p.Description, &p.CreatedAt, &p.IsFavorite)
	if err != nil {
		return p, err
	}

	p.Thumbnails, p.Sizes, err = s.getProductDetails(p.ID)
	return p, err
}

// getProductDetails retrieves thumbnails and sizes for a product
func (s *DBService) getProductDetails(productID int) ([]models.Thumbnail, []models.Size, error) {
	thumbRows, err := s.db.Query("SELECT id, product_id, color, image_url FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return nil, nil, err
	}
	defer thumbRows.Close()

	var thumbnails []models.Thumbnail
	var sizes []models.Size
	for thumbRows.Next() {
		var t models.Thumbnail
		if err := thumbRows.Scan(&t.ID, &t.ProductID, &t.Color, &t.ImageURL); err != nil {
			return nil, nil, err
		}
		sizeRows, err := s.db.Query("SELECT id, thumbnail_id, size FROM sizes WHERE thumbnail_id = ?", t.ID)
		if err != nil {
			return nil, nil, err
		}
		defer sizeRows.Close()

		for sizeRows.Next() {
			var s models.Size
			if err := sizeRows.Scan(&s.ID, &s.ThumbnailID, &s.Size); err != nil {
				return nil, nil, err
			}
			sizes = append(sizes, s)
		}
		thumbnails = append(thumbnails, t)
	}
	return thumbnails, sizes, nil
}

// CreateProduct creates a new product with thumbnails and sizes
func (s *DBService) CreateProduct(marketID, name, price, discount, description string, thumbnailURLs []string, colors, sizes string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO products (market_id, name, price, discount, description) VALUES (?, ?, ?, ?, ?)",
		marketID, name, price, discount, description)
	if err != nil {
		return 0, err
	}

	productID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	colorList := strings.Split(colors, ",")
	sizeList := strings.Split(sizes, ",")
	for i, url := range thumbnailURLs {
		color := colorList[i]
		thumbResult, err := tx.Exec("INSERT INTO thumbnails (product_id, color, image_url) VALUES (?, ?, ?)", productID, color, url)
		if err != nil {
			return 0, err
		}
		thumbID, err := thumbResult.LastInsertId()
		if err != nil {
			return 0, err
		}
		for _, size := range sizeList {
			_, err = tx.Exec("INSERT INTO sizes (thumbnail_id, size) VALUES (?, ?)", thumbID, size)
			if err != nil {
				return 0, err
			}
		}
	}

	return int(productID), tx.Commit()
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

	// Delete sizes
	_, err = tx.Exec("DELETE FROM sizes WHERE thumbnail_id IN (SELECT id FROM thumbnails WHERE product_id = ?)", productID)
	if err != nil {
		return err
	}

	// Delete thumbnails
	_, err = tx.Exec("DELETE FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return err
	}

	// Delete product
	_, err = tx.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
