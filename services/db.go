package services

import (
	"log"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"math"
	"math/big"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Dowlet_projects/ecommerce/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"context"
	"encoding/json"
)

// DBService handles database operations
type DBService struct {
	db    *sql.DB
	redis *redis.Client
}

// ThumbnailData represents data for a thumbnail to be inserted
type ThumbnailData struct {
	ProductID int
	Color     string
	ColorRu   string
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

// NewDBService creates a new database service with Redis
func NewDBService(user, password, dbname, redisAddr string) (*DBService, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", user, password, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}
	return &DBService{db: db, redis: redisClient}, nil
}

// Close closes the database and Redis connections
func (s *DBService) Close() {
	s.db.Close()
	s.redis.Close()
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

// GetMarkets retrieves all markets with caching
func (s *DBService) GetMarkets(ctx context.Context, isNew, isVip bool, duration int) ([]models.Market, error) {
    cacheKey := fmt.Sprintf("markets:new:%t:vip:%t:duration:%d", isNew, isVip, duration)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var markets []models.Market
        if err := json.Unmarshal([]byte(cached), &markets); err == nil {
            return markets, nil
        }
        log.Printf("Failed to unmarshal cached markets: %v", err)
    } else if err != redis.Nil {
        log.Printf("Redis get error: %v", err)
    }

    query := `SELECT id, name, name_ru, location, location_ru, thumbnail_url, delivery_price, phone, created_at,
              CASE WHEN DATEDIFF(CURDATE(), STR_TO_DATE(created_at, '%Y-%m-%d %H:%i:%s')) <= ? THEN true ELSE false END as isNew,
              isVIP FROM markets WHERE 1=1`
    args := []interface{}{duration}
    if isNew {
        query += " AND DATEDIFF(CURDATE(), STR_TO_DATE(created_at, '%Y-%m-%d %H:%i:%s')) <= ?"
        args = append(args, duration)
    }
    if isVip {
        query += " AND isVIP = ?"
        args = append(args, true)
    }
    query += " ORDER BY id DESC"

    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query markets: %v", err)
    }
    defer rows.Close()

    var markets []models.Market
    for rows.Next() {
        var m models.Market
        var createdAtStr string
        if err := rows.Scan(&m.ID, &m.Name, &m.NameRu, &m.Location, &m.LocationRu, &m.ThumbnailURL, &m.DeliveryPrice, &m.Phone, &createdAtStr, &m.IsNew, &m.IsVIP); err != nil {
            return nil, fmt.Errorf("failed to scan market: %v", err)
        }
        markets = append(markets, m)
    }

    if len(markets) > 0 {
        marketsJSON, err := json.Marshal(markets)
        if err != nil {
            log.Printf("Failed to marshal markets: %v", err)
        } else {
            pipe := s.redis.Pipeline()
            pipe.Set(ctx, cacheKey, marketsJSON, 1*time.Hour)
            pipe.SAdd(ctx, "markets_cache_keys", cacheKey)
            if _, err := pipe.Exec(ctx); err != nil {
                log.Printf("Failed to set cache: %v", err)
            }
        }
    }

    return markets, nil
}

// GetMarketProducts retrieves paginated products for a market with caching
func (s *DBService) GetMarketProducts(ctx context.Context, marketID, categoryID, page, limit int) ([]models.Product, error) {
	cacheKey := fmt.Sprintf("market:%d:products:page:%d:limit:%d", marketID, page, limit)

	shouldCache := categoryID == 0
	if shouldCache {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var products []models.Product
			if err := json.Unmarshal([]byte(cached), &products); err == nil {
				return products, nil
			}
			log.Printf("Failed to unmarshal cached products: %v", err)
		} else if err != redis.Nil {
			log.Printf("Redis get error: %v", err)
		}
	}
	

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
        SELECT 
            p.id, 
            p.market_id, 
            m.name as market_name, 
            m.name_ru as market_name_ru, 
            p.category_id, 
            c.name as category_name,
            c.name_ru as category_name_ru,
            p.name, 
            p.name_ru, 
            p.price, 
            p.discount, 
            p.description, 
            p.description_ru, 
            p.created_at, 
            IF(f.user_id IS NOT NULL, true, false) as is_favorite,
            COALESCE(t.image_url, '') as thumbnail_url,
            CASE 
                WHEN DATEDIFF(CURDATE(), p.created_at) <= 7 THEN true 
                ELSE false 
            END as isNew,
            CASE 
                WHEN p.discount IS NOT NULL AND p.discount > 0 
                THEN p.price - (p.price * p.discount / 100) 
                ELSE p.price 
            END as final_price
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id 
        LEFT JOIN markets m ON p.market_id = m.id 
        LEFT JOIN favorites f ON p.id = f.product_id 
        LEFT JOIN thumbnails t ON p.thumbnail_id = t.id 
        WHERE p.market_id = ?`
		args := []interface{}{marketID}

		if categoryID != 0 {
		query += " AND p.category_id = ?"
		args = append(args, categoryID)
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
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.MarketNameRu, &p.CategoryID, &p.CategoryName, &p.CategoryNameRu, &p.Name, &p.NameRu, &p.Price, &p.Discount, &p.Description, &p.DescriptionRu, &createdAtStr, &p.IsFavorite, &p.ThumbnailURL, &p.IsNew, &p.FinalPrice); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.CreatedAt = createdAtStr
		p.Thumbnails = []models.Thumbnail{}
		products = append(products, p)
	}

	// Cache the result only if shouldCache is true
	if shouldCache && len(products) > 0 {
		productsJSON, err := json.Marshal(products)
		if err != nil {
			log.Printf("Failed to marshal products: %v", err)
		} else {
			pipe := s.redis.Pipeline()
			pipe.Set(ctx, cacheKey, productsJSON, 5*time.Minute)
			pipe.SAdd(ctx, fmt.Sprintf("market:%d:products_cache_keys", marketID), cacheKey)
			if _, err := pipe.Exec(ctx); err != nil {
				log.Printf("Failed to set cache: %v", err)
			}
		}
	}

	

	return products, nil
}

// GetMarketByID retrieves a market and its products by market ID with pagination
func (s *DBService) GetMarketByID(ctx context.Context, marketID, categoryID, page, limit int) (*models.Market, []models.Product, int, error) {
	var market models.Market
	err := s.db.QueryRow(`
        SELECT id, phone, name, name_ru, location, location_ru, delivery_price, thumbnail_url
        FROM markets WHERE id = ?`, marketID).
		Scan(&market.ID, &market.Phone, &market.Name, &market.NameRu, &market.Location, &market.LocationRu, &market.DeliveryPrice, &market.ThumbnailURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, 0, fmt.Errorf("market not found")
		}
		return nil, nil, 0, fmt.Errorf("failed to query market: %v", err)
	}

	products, err := s.GetMarketProducts(ctx, marketID, categoryID, page, limit)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to query products: %v", err)
	}

	// Get total product count
	var totalCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM products WHERE market_id = ?", marketID).Scan(&totalCount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to query product count: %v", err)
	}

	return &market, products, totalCount, nil
}

// UpdateProduct updates a product, its thumbnail, and invalidates cache
func (s *DBService) UpdateProduct(ctx context.Context, marketID, productID, categoryID int, name, nameRu string, price, discount float64, description, descriptionRu string, isActive bool, imageURL string) (oldImageURL string, err error) {
    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
		fmt.Println(err)
        return "", fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()
	fmt.Println(productID, marketID)
    // Verify product exists and belongs to market
    var thumbnailID *int
    err = tx.QueryRowContext(ctx, "SELECT thumbnail_id FROM products WHERE id = ? AND market_id = ?", productID, marketID).Scan(&thumbnailID)
    if err != nil {
		fmt.Println(err)
        return "", fmt.Errorf("failed to validate product: %v", err)
    }
    if thumbnailID == nil {
		fmt.Println("error thumbnail")
        return "", fmt.Errorf("product not found or unauthorized")
    }

    // Fetch old image_url if thumbnail_id exists
    oldImageURL = ""
    if *thumbnailID != 0 {
        err = tx.QueryRowContext(ctx, "SELECT COALESCE(image_url, '') FROM thumbnails WHERE id = ?", *thumbnailID).Scan(&oldImageURL)
        if err != nil {
			fmt.Println(err)
            return "", fmt.Errorf("failed to fetch old image_url: %v", err)
        }
    }

	thumbnailURL := filepath.Join("/uploads/products", imageURL)
    thumbnailURL = strings.ReplaceAll(thumbnailURL, string(filepath.Separator), "/")
	
    // Update thumbnail if new image_url is provided
    if imageURL != "" && *thumbnailID != 0 {
		fmt.Println(*thumbnailID)
        result, err := tx.ExecContext(ctx, "UPDATE thumbnails SET image_url = ? WHERE id = ?", thumbnailURL, *thumbnailID)
        if err != nil {
			fmt.Println(err)
            return "", fmt.Errorf("failed to update thumbnail: %v", err)
        }
        rowsAffected, err := result.RowsAffected()
        if err != nil {
			fmt.Println(err)
            return "", fmt.Errorf("failed to check thumbnail update: %v", err)
        }
        if rowsAffected == 0 {
            return "", fmt.Errorf("thumbnail not found")
        }
    }

    // Update product
    _, err = tx.ExecContext(ctx, `
        UPDATE products 
        SET category_id = ?, name = ?, name_ru = ?, price = ?, discount = ?, description = ?, description_ru = ?, is_active = ?
        WHERE id = ? AND market_id = ?`,
        categoryID, name, nameRu, price, discount, description, descriptionRu, isActive, productID, marketID)
    if err != nil {
        return "", fmt.Errorf("failed to update product: %v", err)
    }

    // rowsAffected, err := result.RowsAffected()
    // if err != nil {
    //     return "", fmt.Errorf("failed to check product update: %v", err)
    // }
    // if rowsAffected == 0 {
    //     return "", fmt.Errorf("product not found or unauthorized")
    // }

    // Commit transaction
    if err := tx.Commit(); err != nil {
		fmt.Println(err)
        return "", fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Invalidate caches using pipeline
    pipe := s.redis.Pipeline()
    marketKeys, err := s.redis.SMembers(ctx, fmt.Sprintf("market:%d:products_cache_keys", marketID)).Result()
    if err == nil && len(marketKeys) > 0 {
		fmt.Println(err)
        pipe.Del(ctx, marketKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch market cache keys: %v", err)
    }

    globalKeys, err := s.redis.SMembers(ctx, "global_products_cache_keys").Result()
    if err == nil && len(globalKeys) > 0 {
		fmt.Println(err)
        pipe.Del(ctx, globalKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch global cache keys: %v", err)
    }

    if _, err := pipe.Exec(ctx); err != nil {
        log.Printf("Failed to invalidate caches: %v", err)
    }

    return oldImageURL, nil
}

// GetPaginatedProducts retrieves products with pagination, optional filters, and sorting with caching
func (s *DBService) GetPaginatedProducts(ctx context.Context, categoryID, marketID, duration, page, limit int, search string, random bool, startPrice, endPrice float64, sort string, hasDiscount, isNew bool) ([]models.Product, error) {
	// Generate cache key (exclude duration, search, startPrice, endPrice)
	cacheKey := fmt.Sprintf("products:cat:%d:page:%d:limit:%d:random:%t:sort:%s:discount:%t:new:%t",
		categoryID, page, limit, random, sort, hasDiscount, isNew)

	// Check cache only if no search, no price filters, and default duration (7 days)
	defaultDuration := 7
	shouldCache := search == "" && startPrice == 0 && endPrice == 0 && duration == defaultDuration
	if shouldCache {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var products []models.Product
			if err := json.Unmarshal([]byte(cached), &products); err == nil {
				return products, nil
			}
			log.Printf("Failed to unmarshal cached products: %v", err)
		} else if err != redis.Nil {
			log.Printf("Redis get error: %v", err)
		}
	}

	// Pagination setup
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Use default duration for caching, otherwise use provided duration
	queryDuration := defaultDuration
	if !shouldCache {
		queryDuration = duration
	}

	// Build SQL query
	query := `
        SELECT 
            p.id, 
            p.market_id, 
            m.name as market_name, 
            m.name_ru as market_name_ru, 
            p.category_id, 
            c.name as category_name,
            c.name_ru as category_name_ru,
            p.name, 
            p.name_ru, 
            p.price, 
            p.discount, 
            p.description, 
            p.description_ru, 
            p.created_at, 
            IF(f.user_id IS NOT NULL, true, false) as is_favorite,
            COALESCE(t.image_url, '') as thumbnail_url,
            CASE 
                WHEN DATEDIFF(CURDATE(), p.created_at) <= ? THEN true 
                ELSE false 
            END as isNew,
            CASE 
                WHEN p.discount IS NOT NULL AND p.discount > 0 
                THEN p.price - (p.price * p.discount / 100) 
                ELSE p.price 
            END as final_price
        FROM products p 
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN markets m ON p.market_id = m.id 
        LEFT JOIN favorites f ON p.id = f.product_id 
        LEFT JOIN thumbnails t ON p.thumbnail_id = t.id 
        WHERE p.is_active = true`
	args := []interface{}{queryDuration}

	if categoryID != 0 {
		query += " AND p.category_id = ?"
		args = append(args, categoryID)
	}
	fmt.Println(marketID)
	if marketID != 0 {
		query += " AND p.market_id = ?"
		args = append(args, marketID)
	}

	if search != "" {
    query += " AND (name_lower LIKE ? OR name_ru_lower LIKE ?)"
    args = append(args, "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%")
}
	if startPrice > 0 {
		query += " AND (CASE WHEN p.discount IS NOT NULL AND p.discount > 0 THEN p.price - (p.price * p.discount / 100) ELSE p.price END) >= ?"
		args = append(args, startPrice)
	}
	if endPrice > 0 {
		query += " AND (CASE WHEN p.discount IS NOT NULL AND p.discount > 0 THEN p.price - (p.price * p.discount / 100) ELSE p.price END) <= ?"
		args = append(args, endPrice)
	}
	if hasDiscount {
		query += " AND p.discount IS NOT NULL AND p.discount > 0"
	}
	if isNew {
		query += " AND DATEDIFF(CURDATE(), p.created_at) <= ?"
		args = append(args, queryDuration)
	}
	if random {
		query += " ORDER BY RAND()"
	} else {
		switch sort {
		case "cheap_to_expensive":
			query += " ORDER BY (CASE WHEN p.discount IS NOT NULL AND p.discount > 0 THEN p.price - (p.price * p.discount / 100) ELSE p.price END) ASC"
		case "expensive_to_cheap":
			query += " ORDER BY (CASE WHEN p.discount IS NOT NULL AND p.discount > 0 THEN p.price - (p.price * p.discount / 100) ELSE p.price END) DESC"
		default:
			query += " ORDER BY p.id DESC"
		}
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %v", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.MarketNameRu, &p.CategoryID, &p.CategoryName, &p.CategoryNameRu, &p.Name, &p.NameRu, &p.Price, &p.Discount,
			&p.Description, &p.DescriptionRu, &p.CreatedAt, &p.IsFavorite, &p.ThumbnailURL, &p.IsNew, &p.FinalPrice); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		p.Thumbnails = []models.Thumbnail{}
		products = append(products, p)
	}

	// Cache the result only if shouldCache is true
	if shouldCache && len(products) > 0 {
		productsJSON, err := json.Marshal(products)
		if err != nil {
			log.Printf("Failed to marshal products: %v", err)
		} else {
			pipe := s.redis.Pipeline()
			pipe.Set(ctx, cacheKey, productsJSON, 3*time.Minute)
			pipe.SAdd(ctx, "global_products_cache_keys", cacheKey)
			if _, err := pipe.Exec(ctx); err != nil {
				log.Printf("Failed to set cache: %v", err)
			}
		}
	}

	return products, nil
}

// GetProduct retrieves a single product by ID
func (s *DBService) GetProduct(id string) (models.Product, error) {
	var p models.Product
	err := s.db.QueryRow(`
		SELECT 
			p.id, 
			p.market_id, 
			m.name as market_name, 
			m.name_ru as market_name_ru, 
			p.category_id,
			c.name as category_name,
			c.name_ru as category_name_ru, 
			p.name, 
			p.name_ru, 
			p.price, 
			p.discount, 
			p.description, 
			p.description_ru, 
			p.created_at, 
			IF(f.user_id IS NOT NULL, true, false) as is_favorite,
			COALESCE(t.image_url, '') as thumbnail_url,
			CASE 
				WHEN DATEDIFF(CURDATE(), p.created_at) <= ? THEN true 
				ELSE false 
			END as isNew,
			CASE 
				WHEN p.discount IS NOT NULL AND p.discount > 0 
				THEN p.price - (p.price * p.discount / 100) 
				ELSE p.price 
			END as final_price
		FROM products p 
		LEFT JOIN categories c ON p.category_id = c.id 
		LEFT JOIN markets m ON p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id 
		LEFT JOIN thumbnails t ON p.thumbnail_id = t.id 
		WHERE p.id = ?`, 7, id).Scan(&p.ID, &p.MarketID, &p.MarketName, &p.MarketNameRu, &p.CategoryID, &p.CategoryName, &p.CategoryNameRu, &p.Name, &p.NameRu, &p.Price, &p.Discount, &p.Description, &p.DescriptionRu, &p.CreatedAt, &p.IsFavorite, &p.ThumbnailURL, &p.IsNew, &p.FinalPrice)
	if err != nil {
		return p, err
	}

	p.Thumbnails, err = s.getProductDetails(p.ID)
	return p, err
}

// getProductDetails retrieves thumbnails and sizes for a product
func (s *DBService) getProductDetails(productID int) ([]models.Thumbnail, error) {
	thumbRows, err := s.db.Query("SELECT id, product_id, color, color_ru, image_url FROM thumbnails WHERE product_id = ?", productID)
	if err != nil {
		return nil, err
	}
	defer thumbRows.Close()

	var thumbnails []models.Thumbnail = []models.Thumbnail{}
	for thumbRows.Next() {
		var sizes []models.Size = []models.Size{}
		var t models.Thumbnail
		if err := thumbRows.Scan(&t.ID, &t.ProductID, &t.Color, &t.ColorRu, &t.ImageURL); err != nil {
			return nil, err
		}
		sizeRows, err := s.db.Query("SELECT id, thumbnail_id, size, count, price FROM sizes WHERE thumbnail_id = ?", t.ID)
		if err != nil {
			return nil, err
		}
		defer sizeRows.Close()

		for sizeRows.Next() {
			var s models.Size
			if err := sizeRows.Scan(&s.ID, &s.ThumbnailID, &s.Size, &s.Count, &s.Price); err != nil {
				return nil, err
			}
			sizes = append(sizes, s)
			t.Sizes = sizes
		}
		thumbnails = append(thumbnails, t)
	}
	return thumbnails, nil
}

// CreateProduct creates a new product and invalidates cache
func (s *DBService) CreateProduct(ctx context.Context, marketID, categoryID int, name, name_ru string, price, discount float64, description, description_ru string, is_active bool, urlPath, filePath, filename string) (int, error) {
    // Verify market exists
    var exists bool
    err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM markets WHERE id = ?)", marketID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to validate market: %v", err)
    }
    if !exists {
        return 0, fmt.Errorf("market not found")
    }

    // Verify category exists
    err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", categoryID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to validate category: %v", err)
    }
    if !exists {
        return 0, fmt.Errorf("category not found")
    }

    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return 0, fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Insert into thumbnails table
    result, err := tx.ExecContext(ctx, "INSERT INTO thumbnails (image_url) VALUES (?)", urlPath)
    if err != nil {
        os.Remove(filePath) // Clean up file on error
        return 0, fmt.Errorf("failed to save thumbnail URL: %v", err)
    }
    thumbnailID64, err := result.LastInsertId()
    if err != nil {
        os.Remove(filePath)
        return 0, fmt.Errorf("failed to retrieve thumbnail ID: %v", err)
    }
    thumbnailID := int(thumbnailID64)

    // Verify thumbnail_id
    if thumbnailID != 0 {
        err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM thumbnails WHERE id = ?)", thumbnailID).Scan(&exists)
        if err != nil {
            os.Remove(filePath)
            return 0, fmt.Errorf("failed to validate thumbnail: %v", err)
        }
        if !exists {
            os.Remove(filePath)
            return 0, fmt.Errorf("thumbnail not found")
        }
    }

    // Insert product
    result, err = tx.ExecContext(ctx, `
        INSERT INTO products (market_id, category_id, name, name_ru, price, discount, description, description_ru, is_active, thumbnail_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        marketID, categoryID, name, name_ru, price, discount, description, description_ru, is_active, thumbnailID)
    if err != nil {
        if thumbnailID != 0 {
            tx.ExecContext(ctx, "DELETE FROM thumbnails WHERE id = ?", thumbnailID)
            os.Remove(filepath.Join("Uploads", "products", "main", filename))
        }
        return 0, fmt.Errorf("failed to create product: %v", err)
    }

    productID, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to retrieve product ID: %v", err)
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return 0, fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Invalidate caches using pipeline
    pipe := s.redis.Pipeline()
    marketKeys, err := s.redis.SMembers(ctx, fmt.Sprintf("market:%d:products_cache_keys", marketID)).Result()
    if err == nil && len(marketKeys) > 0 {
        pipe.Del(ctx, marketKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch market cache keys: %v", err)
    }

    globalKeys, err := s.redis.SMembers(ctx, "global_products_cache_keys").Result()
    if err == nil && len(globalKeys) > 0 {
        pipe.Del(ctx, globalKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch global cache keys: %v", err)
    }

    if _, err := pipe.Exec(ctx); err != nil {
        log.Printf("Failed to invalidate caches: %v", err)
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
// DeleteProduct deletes a product and its thumbnails and invalidates cache
func (s *DBService) DeleteProduct(ctx context.Context, marketID, productID int) error {
    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Verify product exists and belongs to market, fetch thumbnail_id
    var exists bool
    var thumbnailID int
    err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = ? AND market_id = ?), COALESCE(thumbnail_id, 0) FROM products WHERE id = ? AND market_id = ?",
        productID, marketID, productID, marketID).Scan(&exists, &thumbnailID)
    if err != nil {
        return fmt.Errorf("failed to validate product: %v", err)
    }
    if !exists {
        return fmt.Errorf("product not found or unauthorized")
    }

    // Fetch legacy thumbnails (linked via product_id)
    rows, err := tx.QueryContext(ctx, "SELECT image_url FROM thumbnails WHERE product_id = ?", productID)
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

    // Delete legacy thumbnails (sizes are deleted via CASCADE)
    _, err = tx.ExecContext(ctx, "DELETE FROM thumbnails WHERE product_id = ?", productID)
    if err != nil {
        return fmt.Errorf("failed to delete thumbnails: %v", err)
    }

    // Delete main thumbnail (via thumbnail_id)
    var thumbnailURL string
    if thumbnailID != 0 {
        err = tx.QueryRowContext(ctx, "SELECT image_url FROM thumbnails WHERE id = ?", thumbnailID).Scan(&thumbnailURL)
        if err != nil && err != sql.ErrNoRows {
            return fmt.Errorf("failed to query main thumbnail: %v", err)
        }
        if thumbnailURL != "" {
            imageURLs = append(imageURLs, thumbnailURL)
            _, err = tx.ExecContext(ctx, "DELETE FROM thumbnails WHERE id = ?", thumbnailID)
            if err != nil {
                return fmt.Errorf("failed to delete main thumbnail: %v", err)
            }
        }
    }

    // Delete product
    result, err := tx.ExecContext(ctx, "DELETE FROM products WHERE id = ? AND market_id = ?", productID, marketID)
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
        uploadDir = "./Uploads/products"
    }
    for _, imageURL := range imageURLs {
        filePath := filepath.Join(uploadDir, filepath.Base(imageURL))
        if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
            log.Printf("Warning: failed to delete file %s: %v", filePath, err)
        }
    }

    // Invalidate caches using pipeline
    pipe := s.redis.Pipeline()
    marketKeys, err := s.redis.SMembers(ctx, fmt.Sprintf("market:%d:products_cache_keys", marketID)).Result()
    if err == nil && len(marketKeys) > 0 {
        pipe.Del(ctx, marketKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch market cache keys: %v", err)
    }

    globalKeys, err := s.redis.SMembers(ctx, "global_products_cache_keys").Result()
    if err == nil && len(globalKeys) > 0 {
        pipe.Del(ctx, globalKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch global cache keys: %v", err)
    }

    if _, err := pipe.Exec(ctx); err != nil {
        log.Printf("Failed to invalidate caches: %v", err)
    }

    return nil
}
// CreateMarket creates a market and invalidates related caches
func (s *DBService) CreateMarket(ctx context.Context, name, name_ru, location, location_ru, thumbnailURL, phone, password string, deliveryPrice float64) (string, string, error) {
    // Verify phone doesn't exist
    var exists bool
    err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM markets WHERE phone = ?)", phone).Scan(&exists)
    if err != nil {
        return "", "", fmt.Errorf("failed to check phone: %v", err)
    }
    if exists {
        return "", "", fmt.Errorf("phone already exists")
    }

    // Hash password
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", "", fmt.Errorf("failed to hash password: %v", err)
    }

    // Insert market
    result, err := s.db.ExecContext(ctx, `
        INSERT INTO markets (password, phone, name, name_ru, location, location_ru, thumbnail_url, delivery_price)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        passwordHash, phone, name, name_ru, location, location_ru, thumbnailURL, deliveryPrice)
    if err != nil {
        return "", "", fmt.Errorf("failed to create market: %v", err)
    }

    marketID, err := result.LastInsertId()
    if err != nil {
        return "", "", fmt.Errorf("failed to retrieve market ID: %v", err)
    }

    // Invalidate caches using pipeline
    pipe := s.redis.Pipeline()

    // Invalidate market-specific product caches
    marketCacheKey := fmt.Sprintf("market:%d:products_cache_keys", marketID)
    marketKeys, err := s.redis.SMembers(ctx, marketCacheKey).Result()
    if err == nil && len(marketKeys) > 0 {
        pipe.Del(ctx, marketKeys...)
        pipe.Del(ctx, marketCacheKey)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch market cache keys: %v", err)
    }

    // Invalidate global markets caches
    globalMarketKeys, err := s.redis.SMembers(ctx, "markets_cache_keys").Result()
    if err == nil && len(globalMarketKeys) > 0 {
        pipe.Del(ctx, globalMarketKeys...)
        pipe.Del(ctx, "markets_cache_keys")
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch global markets cache keys: %v", err)
    }

    if _, err := pipe.Exec(ctx); err != nil {
        log.Printf("Failed to invalidate caches: %v", err)
    }

    return phone, strconv.Itoa(int(marketID)), nil
}

// AuthenticateMarket authenticates a market admin
func (s *DBService) AuthenticateMarket(phone, password string) (int, int, error) {
	var userID, marketID int
	var passwordHash string
	err := s.db.QueryRow("SELECT id, id, password FROM markets WHERE phone = ?", phone).Scan(&userID, &marketID, &passwordHash)
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
		filePath := filepath.Join(uploadDir+"/products", filepath.Base(imageURL))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to delete file %s: %v", filePath, err)
		}
	}

	// Delete market thumbnail file
	if marketThumbnailURL != "" {
		filePath := filepath.Join(uploadDir+"/markets", filepath.Base(marketThumbnailURL))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to delete market thumbnail %s: %v", filePath, err)
		}
	}

	// Invalidate cache
	marketKeys, err := s.redis.SMembers(context.Background(), "markets_cache_keys").Result()
	if err == nil && len(marketKeys) > 0 {
		s.redis.Del(context.Background(), marketKeys...)
	} else if err != nil && err != redis.Nil {
		log.Printf("Failed to invalidate markets cache: %v", err)
	}

	// Also invalidate market-specific product caches
	marketProductKeys, err := s.redis.SMembers(context.Background(), fmt.Sprintf("market:%d:products_cache_keys", marketID)).Result()
	if err == nil && len(marketProductKeys) > 0 {
		s.redis.Del(context.Background(), marketProductKeys...)
	} else if err != nil && err != redis.Nil {
		log.Printf("Failed to invalidate market product cache: %v", err)
	}

	return nil
}

func (s *DBService) CreateThumbnails(thumbnails []ThumbnailData) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	var lastInsertID int64
	for _, thumb := range thumbnails {
		result, err := tx.Exec("INSERT INTO thumbnails (product_id, color, color_ru, image_url) VALUES (?, ?, ?, ?)",
			thumb.ProductID, thumb.Color, thumb.ColorRu, thumb.ImageURL)
		if err != nil {
			return 0, fmt.Errorf("failed to insert thumbnail: %v", err)
		}
		// Get the last inserted ID for this insert
		id, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID: %v", err)
		}
		lastInsertID = id // Update to the most recent ID
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return lastInsertID, nil
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

	uploadDir := os.Getenv("UPLOAD_DIR")
    if uploadDir == "" {
        uploadDir = "./Uploads/products"
    }

	
	if imageURL != "" {
		filePath := filepath.Join(uploadDir, filepath.Base(imageURL))
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: failed to delete file %s: %v", filePath, err)
			}
	}

	// Invalidate cache
	s.redis.Del(context.Background(), fmt.Sprintf("market:%d:products:*", marketID))

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

	var thumbnails []ThumbnailWithProduct = []ThumbnailWithProduct{}
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
func (s *DBService) CreateSizeByThumbnailID(marketID int, thumbnailID string, size string, count int, price float64) error {
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

	_, err = s.db.Exec("INSERT INTO sizes (thumbnail_id, size, count, price) VALUES (?, ?, ?, ?)",
		thumbnailIDInt, size, count, price)
	if err != nil {
		return fmt.Errorf("failed to insert size: %v", err)
	}

	// Invalidate cache
	s.redis.Del(context.Background(), fmt.Sprintf("market:%d:products:*", marketID))

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

	// Invalidate cache
	s.redis.Del(context.Background(), fmt.Sprintf("market:%d:products:*", marketID))

	return nil
}



// DeleteOrderByID deletes a size by its ID
func (s *DBService) DeleteOrderByID(marketID int, orderID string) error {
	// Verify size exists and belongs to market's product
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM orders 
			WHERE id = ? AND market_id = ?
		)`, orderID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate order: %v", err)
	}
	if !exists {
		return fmt.Errorf("order not found or unauthorized")
	}

	result, err := s.db.Exec("DELETE FROM orders WHERE id = ?", orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	// Invalidate cache
	s.redis.Del(context.Background(), fmt.Sprintf("market:%d:products:*", marketID))

	return nil
}

func (s *DBService) GetUserFavoriteProducts(userID, page, limit int) ([]models.Product, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 {
        limit = 10
    }
    offset := (page - 1) * limit

    query := `
        SELECT 
			p.id, 
			p.market_id, 
			m.name as market_name, 
			m.name_ru as market_name_ru, 
			p.category_id, 
			p.name, 
			p.name_ru, 
			p.price, 
			p.discount, 
			p.description, 
			p.description_ru, 
			p.created_at, 
			IF(f.user_id IS NOT NULL, true, false) as is_favorite,
			COALESCE(t.image_url, '') as thumbnail_url,
			CASE 
				WHEN DATEDIFF(CURDATE(), p.created_at) <= 7 THEN true 
				ELSE false 
			END as isNew,
			CASE 
				WHEN p.discount IS NOT NULL AND p.discount > 0 
				THEN p.price - (p.price * p.discount / 100) 
				ELSE p.price 
			END as final_price
		FROM products p 
		LEFT JOIN markets m ON p.market_id = m.id 
		LEFT JOIN favorites f ON p.id = f.product_id 
		LEFT JOIN thumbnails t ON p.thumbnail_id = t.id 
        WHERE f.user_id = ?
        ORDER BY p.id DESC LIMIT ? OFFSET ?`

    rows, err := s.db.Query(query, userID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to query favorite products: %v", err)
    }
    defer rows.Close()

    var products []models.Product = []models.Product{}
    for rows.Next() {
        var p models.Product
        var createdAtStr string // Changed to string
        if err := rows.Scan(&p.ID, &p.MarketID, &p.MarketName, &p.MarketNameRu, &p.CategoryID, &p.Name, &p.NameRu, &p.Price, &p.Discount, &p.Description, &p.DescriptionRu, &createdAtStr, &p.IsFavorite, &p.ThumbnailURL, &p.IsNew, &p.FinalPrice); err != nil {
            return nil, fmt.Errorf("failed to scan product: %v", err)
        }
        // Parse the string to time.Time
        createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr) // Adjust format based on your DB
        if err != nil {
            return nil, fmt.Errorf("failed to parse created_at: %v", err)
        }
        p.CreatedAt = createdAt.Format(time.RFC3339)
        // p.Thumbnails, err = s.getProductDetails(p.ID)
        // if err != nil {
        //     return nil, fmt.Errorf("failed to get product details: %v", err)
        // }
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
func (s *DBService) CreateCategory(name, name_ru, thumbnailURL string) (int, error) {
    if name == "" {
        return 0, fmt.Errorf("category name is required")
    }

    result, err := s.db.Exec("INSERT INTO categories (name, name_ru, thumbnail_url) VALUES (?, ?, ?)", name, name_ru, thumbnailURL)
    if err != nil {
        return 0, fmt.Errorf("failed to create category: %v", err)
    }

    categoryID, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to retrieve category ID: %v", err)
    }

    // Invalidate caches
    ctx := context.Background()

    // Invalidate global product caches
    globalKeys, err := s.redis.SMembers(ctx, "global_products_cache_keys").Result()
    if err == nil && len(globalKeys) > 0 {
        s.redis.Del(ctx, globalKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to invalidate global product cache: %v", err)
    }

    // Invalidate category caches
    categoryKeys, err := s.redis.SMembers(ctx, "categories_cache_keys").Result()
    if err == nil && len(categoryKeys) > 0 {
        s.redis.Del(ctx, categoryKeys...)
        s.redis.Del(ctx, "categories_cache_keys") // Clear the tracking set
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to invalidate category cache: %v", err)
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
            log.Printf("Warning: failed to delete category thumbnail %s: %v", filePath, err)
        }
    }

    // Invalidate caches
    ctx := context.Background()

    // Invalidate global product caches
    globalKeys, err := s.redis.SMembers(ctx, "global_products_cache_keys").Result()
    if err == nil && len(globalKeys) > 0 {
        s.redis.Del(ctx, globalKeys...)
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to invalidate global product cache: %v", err)
    }

    // Invalidate category caches
    categoryKeys, err := s.redis.SMembers(ctx, "categories_cache_keys").Result()
    if err == nil && len(categoryKeys) > 0 {
        s.redis.Del(ctx, categoryKeys...)
        s.redis.Del(ctx, "categories_cache_keys") // Clear the tracking set
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to invalidate category cache: %v", err)
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

    // Generate cache key (exclude search)
    cacheKey := fmt.Sprintf("categories:page:%d:limit:%d", page, limit)
    ctx := context.Background()

    // Check Redis cache only if search is empty
    var categories []models.Category
    if search == "" {
        cached, err := s.redis.Get(ctx, cacheKey).Result()
        if err == nil {
            if err := json.Unmarshal([]byte(cached), &categories); err != nil {
                log.Printf("Failed to unmarshal cached categories: %v", err)
            } else {
                return categories, nil
            }
        } else if err != redis.Nil {
            log.Printf("Redis error: %v", err)
        }
    }

    // Query the database
    query := `
        SELECT id, name, name_ru, thumbnail_url
        FROM categories
        WHERE 1=1`
    args := []interface{}{}
    if search != "" {
        query += " AND LOWER(name) LIKE ?"
        args = append(args, "%"+strings.ToLower(search)+"%")
    }
    query += " ORDER BY id DESC LIMIT ? OFFSET ?"
    args = append(args, limit, offset)

    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query categories: %v", err)
    }
    defer rows.Close()

    categories = []models.Category{}
    for rows.Next() {
        var c models.Category
        if err := rows.Scan(&c.ID, &c.Name, &c.NameRu, &c.ThumbnailURL); err != nil {
            return nil, fmt.Errorf("failed to scan category: %v", err)
        }
        categories = append(categories, c)
    }

    // Cache the result only if search is empty
    if search == "" && len(categories) > 0 {
        jsonData, err := json.Marshal(categories)
        if err != nil {
            log.Printf("Failed to marshal categories for caching: %v", err)
        } else {
            ttl := 3600 * time.Second // 1 hour TTL
            pipe := s.redis.Pipeline()
            pipe.Set(ctx, cacheKey, jsonData, ttl)
            pipe.SAdd(ctx, "categories_cache_keys", cacheKey)
            if _, err := pipe.Exec(ctx); err != nil {
                log.Printf("Failed to cache categories: %v", err)
            }
        }
    }

    return categories, nil
}

// AddToCart adds or updates a product in the user's cart under a single cart order
func (s *DBService) AddToCart(userID int, req models.CartRequest) (int, error) {
	if req.Count <= 0 {
		return 0, fmt.Errorf("count must be positive for product_id %d", req.ProductID)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Derive market_id from product_id
	var marketID int
	err = tx.QueryRow("SELECT market_id FROM products WHERE id = ?", req.ProductID).Scan(&marketID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid product_id %d", req.ProductID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get market_id for product_id %d: %v", req.ProductID, err)
	}

	// Validate relationships
	var exists bool
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 
            FROM products p
            JOIN markets m ON p.market_id = m.id
            JOIN thumbnails t ON t.id = ?
            JOIN sizes s ON s.id = ? AND s.thumbnail_id = t.id
            WHERE p.id = ? AND m.id = ?
        )`, req.ThumbnailID, req.SizeID, req.ProductID, marketID).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to validate product_id %d: %v", req.ProductID, err)
	}
	if !exists {
		return 0, fmt.Errorf("invalid market, product_id %d, thumbnail, or size", req.ProductID)
	}

	// Find existing cart_order_id for user and market
	var cartOrderID int
	err = tx.QueryRow(`
        SELECT DISTINCT cart_order_id 
        FROM carts 
        WHERE user_id = ? AND market_id = ? 
        LIMIT 1`, userID, marketID).Scan(&cartOrderID)
	if err == sql.ErrNoRows {
		// Generate new cart_order_id (using max + 1 for simplicity)
		err = tx.QueryRow(`
            SELECT COALESCE(MAX(cart_order_id), 0) + 1 
            FROM carts`).Scan(&cartOrderID)
		if err != nil {
			return 0, fmt.Errorf("failed to generate cart_order_id: %v", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to check existing cart: %v", err)
	}

	// Check if cart entry exists
	var cartID int
	err = tx.QueryRow(`
        SELECT id FROM carts 
        WHERE user_id = ? AND market_id = ? AND product_id = ? AND thumbnail_id = ? AND size_id = ?`,
		userID, marketID, req.ProductID, req.ThumbnailID, req.SizeID).Scan(&cartID)
	if err == sql.ErrNoRows {
		// Insert new cart entry
		result, err := tx.Exec(`
            INSERT INTO carts (user_id, market_id, product_id, thumbnail_id, size_id, count, cart_order_id)
            VALUES (?, ?, ?, ?, ?, ?, ?)`,
			userID, marketID, req.ProductID, req.ThumbnailID, req.SizeID, req.Count, cartOrderID)
		if err != nil {
			return 0, fmt.Errorf("failed to insert product_id %d: %v", req.ProductID, err)
		}
		cartID64, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve cart ID for product_id %d: %v", req.ProductID, err)
		}
		cartID = int(cartID64)
	} else if err != nil {
		return 0, fmt.Errorf("failed to check cart entry for product_id %d: %v", req.ProductID, err)
	} else {
		// Update existing cart entry
		_, err := tx.Exec(`
            UPDATE carts 
            SET count = count + ?
            WHERE id = ?`,
			req.Count, cartID)
		if err != nil {
			return 0, fmt.Errorf("failed to update cart entry for product_id %d: %v", req.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return cartOrderID, nil
}

// GetUserCart retrieves a user's cart grouped by cart order and markets
func (s *DBService) GetUserCart(userID int) ([]models.CartMarket, error) {
	query := `
		SELECT 
			c.cart_order_id, 
			c.user_id, 
			m.id, 
			m.name, 
			m.name_ru as market_name_ru, 
			m.delivery_price, 
			p.id, 
			p.name, 
			p.name_ru, 
			p.price, 
			p.discount, 
			t.id, 
			t.image_url, 
			t.color, 
			t.color_ru, 
			s.id, 
			s.size, 
			s.price, 
			c.count, 
			s.price*(1-COALESCE(p.discount,0)/100)*c.count AS sum
		FROM carts c
		JOIN markets m ON c.market_id = m.id
		JOIN products p ON c.product_id = p.id
		JOIN thumbnails t ON c.thumbnail_id = t.id
		JOIN sizes s ON c.size_id = s.id
		WHERE c.user_id = ?
		ORDER BY c.cart_order_id, m.id, p.id DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart: %v", err)
	}
	defer rows.Close()

	// Group by cart_order_id and market_id
	type cartOrder struct {
		CartOrderID int
		Markets     map[int]*models.CartMarket
	}
	cartOrders := make(map[int]*cartOrder)
	for rows.Next() {
		var cartOrderID, userID, marketID, productID, thumbnailID, sizeID int
		var marketName, productName, thumbnailURL, color, size, marketNameRu, productNameRu, colorRu string
		var price, discount, deliveryPrice, sizePrice, sum float64
		var count int

		if err := rows.Scan(&cartOrderID, &userID, &marketID, &marketName, &marketNameRu, &deliveryPrice,
			&productID, &productName, &productNameRu, &price, &discount,
			&thumbnailID, &thumbnailURL, &color, &colorRu,
			&sizeID, &size, &sizePrice, &count, &sum); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %v", err)
		}

		// Create or update cart order
		order, exists := cartOrders[cartOrderID]
		if !exists {
			order = &cartOrder{
				CartOrderID: cartOrderID,
				Markets:     make(map[int]*models.CartMarket),
			}
			cartOrders[cartOrderID] = order
		}

		// Create or update market entry
		market, exists := order.Markets[marketID]
		if !exists {
			market = &models.CartMarket{
				MarketID:      marketID,
				MarketName:    marketName,
				MarketNameRu:  marketNameRu,
				UserID:        userID,
				CartID:        cartOrderID,
				DeliveryPrice: deliveryPrice,
				Products:      []models.CartProduct{},
			}
			order.Markets[marketID] = market
		}

		// Add product to market
		market.Products = append(market.Products, models.CartProduct{
			SizeID:       sizeID,
			ProductID:    productID,
			ThumbnailURL: thumbnailURL,
			Name:         productName,
			NameRu:       productNameRu,
			Price:        price,
			Discount:     discount,
			Color:        color,
			ColorRu:      colorRu,
			Size:         size,
			SizePrice:    sizePrice,
			Count:        count,
			Sum:          sum,
		})
	}

	// Convert to final output
	var cart []models.CartMarket = []models.CartMarket{}
	for _, order := range cartOrders {
		for _, market := range order.Markets {
			cart = append(cart, *market)
		}
	}

	return cart, nil
}

// DeleteCart deletes all entries for a user's cart order
func (s *DBService) DeleteCart(userID, cartOrderID int) error {
	result, err := s.db.Exec(`
		DELETE FROM carts 
		WHERE cart_order_id = ? AND user_id = ?`, cartOrderID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete cart: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("cart not found or not owned by user")
	}

	return nil
}

// CreateLocation adds a new location for a user
func (s *DBService) CreateLocation(userID int, locationName, locationNameRu, locationAddress, locationAddressRu string) (int, error) {
	if locationName == "" || locationAddress == "" || locationNameRu == "" || locationAddressRu == "" {
		return 0, fmt.Errorf("location name and address are required")
	}

	result, err := s.db.Exec(`
        INSERT INTO locations (user_id, location_name, location_name_ru, location_address, location_address_ru)
        VALUES (?, ?, ?, ?, ?)`, userID, locationName, locationNameRu, locationAddress, locationAddressRu)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, fmt.Errorf("location name already exists")
		}
		return 0, fmt.Errorf("failed to create location: %v", err.Error())
	}

	locationID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve location ID: %v", err.Error())
	}

	return int(locationID), nil
}

// GetUserLocations retrieves all locations for a user
func (s *DBService) GetUserLocations(userID int) ([]models.Location, error) {
	rows, err := s.db.Query(`
        SELECT id, user_id, location_name, location_name_ru, location_address, location_address_ru
        FROM locations
        WHERE user_id = ?
        ORDER BY id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query locations: %v", err)
	}
	defer rows.Close()

	var locations []models.Location = []models.Location{}
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(&loc.ID, &loc.UserID, &loc.LocationName, &loc.LocationNameRu,
			 &loc.LocationAddress, &loc.LocationAddressRu); err != nil {
			return nil, fmt.Errorf("failed to scan location: %v", err)
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// CreateOrder creates a new order for a user's cart
func (s *DBService) CreateOrder(userID, cartOrderID, locationID int, name, phone, notes string) (int, error) {
	if name == "" || phone == "" {
		return 0, fmt.Errorf("name and phone are required")
	}

	// Validate cart_order_id and user ownership
	var exists bool
	err := s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 
            FROM carts 
            WHERE cart_order_id = ? AND user_id = ?
        )`, cartOrderID, userID).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to validate cart: %v", err)
	}
	if !exists {
		return 0, fmt.Errorf("cart not found or not owned by user")
	}
	var forPostOrders models.ForPostOrders
	err = s.db.QueryRow(`
		SELECT 	market_id, product_id, thumbnail_id, size_id, count 
		FROM carts WHERE cart_order_id = ? AND user_id = ?
	`, cartOrderID, userID).Scan(&forPostOrders.MarketID, &forPostOrders.ProductID,
		&forPostOrders.ThumbnailID, &forPostOrders.SizeID, &forPostOrders.Count,
		)	

	if err != nil {
		return 0, fmt.Errorf("failed to validate cart: %v", err)
	}		

	// Validate location_id and user ownership
	err = s.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 
            FROM locations 
            WHERE id = ? AND user_id = ?
        )`, locationID, userID).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to validate location: %v", err)
	}
	if !exists {
		return 0, fmt.Errorf("location not found or not owned by user")
	}

	result, err := s.db.Exec(`
        INSERT INTO orders (user_id, cart_order_id, location_id, name, phone, notes, 
		market_id, product_id, thumbnail_id, size_id, count)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, cartOrderID, locationID, name, phone, notes, forPostOrders.MarketID, 
		forPostOrders.ProductID, forPostOrders.ThumbnailID, forPostOrders.SizeID, 
		forPostOrders.Count,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, fmt.Errorf("order already exists for this cart")
		}
		return 0, fmt.Errorf("failed to create order: %v", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve order ID: %v", err)
	}

	return int(orderID), nil
}

// GetMarketAdminOrders retrieves orders for a market admin's market, optionally filtered by status
func (s *DBService) GetMarketAdminOrders(marketID int, status string) ([]models.MarketAdminOrder, error) {
	query := `
		SELECT 
			o.id,
			o.cart_order_id, 
			l.location_address, 
			l.location_address_ru, 
			o.status,
			o.name, 
			o.created_at,
			SUM(s.price * (1 - COALESCE(p.discount, 0)/100) * o.count) as sum
		FROM orders o
		JOIN locations l ON o.location_id = l.id
		JOIN sizes s ON o.size_id = s.id
		JOIN products p ON o.product_id = p.id
		WHERE o.market_id = ?`
	args := []interface{}{marketID}

	if status != "" {
		query += ` AND o.status = ?`
		args = append(args, status)
	}

	query += `
		GROUP BY o.id
		ORDER BY o.created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []models.MarketAdminOrder = []models.MarketAdminOrder{}
	for rows.Next() {
		var o models.MarketAdminOrder
		if err := rows.Scan(&o.ID, &o.CartOrderID, &o.LocationAddress, &o.LocationAddressRu, &o.Status, &o.Name, &o.CreatedAt, &o.Sum); err != nil {
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}
		o.Sum = math.Round(o.Sum*100) / 100 // Round to 2 decimal places
		orders = append(orders, o)
	}

	return orders, nil
}

// GetMarketAdminOrderByID retrieves a specific order by cart_order_id for a market admin
func (s *DBService) GetMarketAdminOrderByID(marketID, orderID int) (*models.MarketAdminOrderDetail, error) {
	// Fetch order details
	var order models.MarketAdminOrderDetail = models.MarketAdminOrderDetail{}
	err := s.db.QueryRow(`
		SELECT 
			o.id,
			o.cart_order_id, 
			o.name, 
			o.phone,
			o.status, 
			l.location_address, 
			l.location_address_ru,
			o.created_at,
			SUM(s.price * (1 - COALESCE(p.discount, 0)/100) * o.count) as sum
		FROM orders o
		JOIN locations l ON o.location_id = l.id
		JOIN sizes s ON o.size_id = s.id
		JOIN products p ON o.product_id = p.id
		WHERE o.market_id = ? AND o.id = ?
		GROUP BY o.id`,
		marketID, orderID,
	).Scan(&order.ID, &order.CartOrderID, &order.Name, &order.Phone, &order.Status, &order.LocationAddress, &order.LocationAddressRu, &order.CreatedAt, &order.Sum)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found or not for this market")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query order: %v", err)
	}
	order.Sum = math.Round(order.Sum*100) / 100 // Round to 2 decimal places

	// Fetch products
	rows, err := s.db.Query(`
		SELECT 
			p.id, 
			p.name, 
			p.name_ru,
			p.price, 
			t.image_url, 
			COALESCE(p.discount, 0), 
			p.created_at,
			s.size, s.price, o.count,
			(s.price * (1 - COALESCE(p.discount, 0)/100) * o.count) as product_sum
		FROM orders o
		JOIN products p ON o.product_id = p.id
		JOIN sizes s ON o.size_id = s.id
		JOIN thumbnails t ON o.thumbnail_id = t.id
		WHERE o.id = ? AND o.market_id = ?
		ORDER BY p.id`,
		orderID, marketID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %v", err)
	}
	defer rows.Close()

	var products []models.MarketAdminOrderProduct
	for rows.Next() {
		var prod models.MarketAdminOrderProduct
		if err := rows.Scan(&prod.ID, &prod.Name, &prod.NameRu, &prod.Price, &prod.ImageURL, &prod.Discount, &prod.CreatedAt,
			&prod.Size, &prod.SizePrice, &prod.Count, &prod.Sum); err != nil {
			return nil, fmt.Errorf("failed to scan product: %v", err)
		}
		prod.Sum = math.Round(prod.Sum*100) / 100 // Round to 2 decimal places
		products = append(products, prod)
	}
	order.Products = products

	return &order, nil
}

// DeleteLocation deletes a user's location if no orders reference it
func (s *DBService) DeleteLocation(userID, locationID int) error {
	// Verify location exists and belongs to user
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM locations 
			WHERE id = ? AND user_id = ?
		)`, locationID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify location: %v", err)
	}
	if !exists {
		return fmt.Errorf("location not found or not owned by user")
	}

	// Check for orders referencing the location
	err = s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM orders 
			WHERE location_id = ?
		)`, locationID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check orders: %v", err)
	}
	if exists {
		return fmt.Errorf("location is referenced by orders and cannot be deleted")
	}

	// Delete the location
	_, err = s.db.Exec(`
		DELETE FROM locations 
		WHERE id = ? AND user_id = ?`, locationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete location: %v", err)
	}

	return nil
}

// ClearCart deletes all cart entries for a user
func (s *DBService) ClearCart(userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete all cart entries for the user
	_, err = tx.Exec("DELETE FROM carts WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// DeleteCartBySizeID deletes a cart entry for a user based on size_id
func (s *DBService) DeleteCartBySizeID(userID, sizeID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete cart entry
	result, err := tx.Exec("DELETE FROM carts WHERE user_id = ? AND size_id = ?", userID, sizeID)
	if err != nil {
		return fmt.Errorf("failed to delete cart entry: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("cart entry not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// UpdateCartCountBySizeID updates the count of a cart entry for a user based on size_id
func (s *DBService) UpdateCartCountBySizeID(userID, sizeID, countChange int) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	

	// Update count
	result, err := tx.Exec("UPDATE carts SET count = ? WHERE user_id = ? AND size_id = ?", countChange, userID, sizeID)
	if err != nil {
		return 0, fmt.Errorf("failed to update cart entry: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("cart entry not found")
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return countChange, nil
}

// UpdateLocationByID updates a location entry for a user based on location_id
func (s *DBService) UpdateLocationByID(userID, locationID int, req models.UpdateLocationRequest) (models.Location, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Build update query dynamically
	query := "UPDATE locations SET "
	var args []interface{}
	var updates []string

	if req.LocationName != "" {
		updates = append(updates, "location_name = ?")
		args = append(args, req.LocationName)
	}
	if req.LocationAddress != "" {
		updates = append(updates, "location_address = ?")
		args = append(args, req.LocationAddress)
	}

	if len(updates) == 0 {
		return models.Location{}, fmt.Errorf("no fields provided to update")
	}

	query += strings.Join(updates, ", ") + " WHERE id = ? AND user_id = ?"
	args = append(args, locationID, userID)

	// Execute update
	result, err := tx.Exec(query, args...)
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to update location: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return models.Location{}, fmt.Errorf("location not found or unauthorized")
	}

	// Fetch updated location
	var loc models.Location
	err = tx.QueryRow("SELECT id, user_id, location_name, location_address FROM locations WHERE id = ? AND user_id = ?",
		locationID, userID).Scan(&loc.ID, &loc.UserID, &loc.LocationName, &loc.LocationAddress)
	if err == sql.ErrNoRows {
		return models.Location{}, fmt.Errorf("location not found after update")
	}
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to fetch updated location: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return models.Location{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return loc, nil
}

// GetUserProfile retrieves the profile data for a user
func (s *DBService) GetUserProfile(userID int) (models.UserProfile, error) {
	var profile models.UserProfile
	err := s.db.QueryRow("SELECT id, COALESCE(full_name, ''), COALESCE(phone, '') FROM users WHERE id = ?", userID).
		Scan(&profile.ID, &profile.FullName, &profile.Phone)
	if err == sql.ErrNoRows {
		return models.UserProfile{}, fmt.Errorf("user not found")
	}
	if err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to fetch user profile: %v", err)
	}
	return profile, nil
}

// UpdateUserProfile updates the profile data for a user
func (s *DBService) UpdateUserProfile(userID int, req models.UpdateProfileRequest) (models.UserProfile, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Build update query dynamically
	query := "UPDATE users SET "
	var args []interface{}
	var updates []string

	if req.FullName != "" {
		updates = append(updates, "full_name = ?")
		args = append(args, req.FullName)
	}
	if req.Phone != "" {
		updates = append(updates, "phone = ?")
		args = append(args, req.Phone)
	}

	if len(updates) == 0 {
		return models.UserProfile{}, fmt.Errorf("no fields provided to update")
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, userID)

	// Execute update
	result, err := tx.Exec(query, args...)
	if err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to update profile: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return models.UserProfile{}, fmt.Errorf("user not found")
	}

	// Fetch updated profile
	var profile models.UserProfile
	err = tx.QueryRow("SELECT id, COALESCE(full_name, ''), COALESCE(phone, '') FROM users WHERE id = ?", userID).
		Scan(&profile.ID, &profile.FullName, &profile.Phone)
	if err == sql.ErrNoRows {
		return models.UserProfile{}, fmt.Errorf("user not found after update")
	}
	if err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to fetch updated profile: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return models.UserProfile{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return profile, nil
}

func (s *DBService) GetMarketProfile(marketID int) (models.MarketProfile, error) {
	var profile models.MarketProfile
	err := s.db.QueryRow(`
		SELECT id, COALESCE(phone, ''), delivery_price, COALESCE(name, ''), COALESCE(name_ru, ''),
			COALESCE(location, ''), COALESCE(location_ru, ''), COALESCE(thumbnail_url, '')
		FROM markets WHERE id = ?`, marketID).
		Scan(&profile.ID, &profile.Phone, &profile.DeliveryPrice, &profile.Name, &profile.NameRu,
			&profile.Location, &profile.LocationRu, &profile.ThumbnailURL)
	if err == sql.ErrNoRows {
		return models.MarketProfile{}, fmt.Errorf("market not found")
	}
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to fetch market profile: %v", err)
	}
	return profile, nil
}

// UpdateMarketProfile updates the profile data for a market
func (s *DBService) UpdateMarketProfile(marketID int, req models.UpdateMarketProfileRequest, thumbnailURL string) (models.MarketProfile, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Fetch current thumbnail_url
	var oldThumbnailURL string
	err = tx.QueryRow("SELECT COALESCE(thumbnail_url, '') FROM markets WHERE id = ?", marketID).Scan(&oldThumbnailURL)
	if err == sql.ErrNoRows {
		return models.MarketProfile{}, fmt.Errorf("market not found")
	}
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to fetch current thumbnail: %v", err)
	}

	// Build update query dynamically
	query := "UPDATE markets SET "
	var args []interface{}
	var updates []string

	if req.DeliveryPrice != 0 {
		updates = append(updates, "delivery_price = ?")
		args = append(args, req.DeliveryPrice)
	}
	if req.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, req.Name)
	}
	if req.NameRu != "" {
		updates = append(updates, "name_ru = ?")
		args = append(args, req.NameRu)
	}
	if req.Location != "" {
		updates = append(updates, "location = ?")
		args = append(args, req.Location)
	}
	if req.LocationRu != "" {
		updates = append(updates, "location_ru = ?")
		args = append(args, req.LocationRu)
	}
	if thumbnailURL != "" {
		updates = append(updates, "thumbnail_url = ?")
		args = append(args, thumbnailURL)
	}

	if len(updates) == 0 {
		return models.MarketProfile{}, fmt.Errorf("no fields provided to update")
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, marketID)

	// Execute update
	result, err := tx.Exec(query, args...)
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to update market profile: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return models.MarketProfile{}, fmt.Errorf("market not found")
	}

	// Delete old thumbnail if new one is provided and old one exists
	if thumbnailURL != "" && oldThumbnailURL != "" {
		filePath := filepath.Join(".", oldThumbnailURL)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return models.MarketProfile{}, fmt.Errorf("failed to delete old thumbnail: %v", err)
		}
	}

	// Fetch updated profile
	var profile models.MarketProfile
	err = tx.QueryRow(`
		SELECT id, COALESCE(phone, ''), delivery_price, COALESCE(name, ''), COALESCE(name_ru, ''),
			COALESCE(location, ''), COALESCE(location_ru, ''), COALESCE(thumbnail_url, '')
		FROM markets WHERE id = ?`, marketID).
		Scan(&profile.ID, &profile.Phone, &profile.DeliveryPrice, &profile.Name, &profile.NameRu,
			&profile.Location, &profile.LocationRu, &profile.ThumbnailURL)
	if err == sql.ErrNoRows {
		return models.MarketProfile{}, fmt.Errorf("market not found after update")
	}
	if err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to fetch updated market profile: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return models.MarketProfile{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Invalidate cache
	s.redis.Del(context.Background(), "markets:*")

	return profile, nil
}

// CreateBanner inserts a new banner into the banners table
func (s *DBService) CreateBanner(req models.CreateBannerRequest, thumbnailURL string) (models.Banner, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Banner{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert banner
	result, err := tx.Exec("INSERT INTO banners (description, thumbnail_url) VALUES (?, ?)", req.Description, thumbnailURL)
	if err != nil {
		return models.Banner{}, fmt.Errorf("failed to insert banner: %v", err)
	}

	// Get inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return models.Banner{}, fmt.Errorf("failed to get inserted banner ID: %v", err)
	}

	// Fetch created banner
	var banner models.Banner
	err = tx.QueryRow("SELECT id, COALESCE(description, ''), COALESCE(thumbnail_url, '') FROM banners WHERE id = ?", id).
		Scan(&banner.ID, &banner.Description, &banner.ThumbnailURL)
	if err != nil {
		return models.Banner{}, fmt.Errorf("failed to fetch created banner: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return models.Banner{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return banner, nil
}

// DeleteBanner deletes a banner by ID and returns its thumbnail URL
func (s *DBService) DeleteBanner(ctx context.Context, bannerID int) (string, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return "", fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Fetch thumbnail_url
    var thumbnailURL string
    err = tx.QueryRowContext(ctx, "SELECT COALESCE(thumbnail_url, '') FROM banners WHERE id = ?", bannerID).Scan(&thumbnailURL)
    if err == sql.ErrNoRows {
        return "", fmt.Errorf("banner not found")
    }
    if err != nil {
        return "", fmt.Errorf("failed to fetch banner: %v", err)
    }

    // Delete banner
    result, err := tx.ExecContext(ctx, "DELETE FROM banners WHERE id = ?", bannerID)
    if err != nil {
        return "", fmt.Errorf("failed to delete banner: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return "", fmt.Errorf("failed to check rows affected: %v", err)
    }
    if rowsAffected == 0 {
        return "", fmt.Errorf("banner not found")
    }

    if err := tx.Commit(); err != nil {
        return "", fmt.Errorf("failed to commit transaction: %v", err)
    }

    return thumbnailURL, nil
}

// GetAllBanners retrieves all banners from the banners table
func (s *DBService) GetAllBanners() ([]models.Banner, error) {
	query := "SELECT id, COALESCE(description, ''), COALESCE(thumbnail_url, '') FROM banners ORDER BY id DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query banners: %v", err)
	}
	defer rows.Close()

	var banners []models.Banner = []models.Banner{}
	for rows.Next() {
		var banner models.Banner
		if err := rows.Scan(&banner.ID, &banner.Description, &banner.ThumbnailURL); err != nil {
			return nil, fmt.Errorf("failed to scan banner: %v", err)
		}
		banners = append(banners, banner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating banners: %v", err)
	}

	return banners, nil
}

// UpdateOrderStatus updates the status of an order
func (s *DBService) UpdateOrderStatus(orderID, marketID int, status string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate status
	if status != "canceled" && status != "delivered" {
		return fmt.Errorf("invalid status; must be 'canceled' or 'delivered'")
	}

	// Verify order exists and is associated with the market
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM orders o 
			WHERE o.id = ? AND o.market_id = ?
		)`, orderID, marketID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify order: %v", err)
	}
	if !exists {
		return fmt.Errorf("order not found or not associated with this market")
	}

	// Update order status
	result, err := tx.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}



// DeleteUserHistory updates the status of an order
func (s *DBService) DeleteUserHistory(orderID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()


	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM orders o 
			WHERE o.id = ? AND o.user_id = ?
		)`, orderID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify order: %v", err)
	}
	if !exists {
		return fmt.Errorf("order not found or not associated with this user")
	}

	// Update order status
	result, err := tx.Exec("UPDATE orders SET is_active = ? WHERE id = ?", false, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order is_active: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetMarketAdminOrders retrieves orders for a market admin's market, optionally filtered by status
func (s *DBService) GetUserOrders(UserID int, status string) ([]models.UserOrder, error) {
	query := `
		SELECT 
			o.id,
			o.cart_order_id,
			o.status,
			o.created_at,
			m.name as market_name,
			m.name_ru as market_name_ru,
			p.name,
			p.name_ru,
			s.size,
			t.color,
			t.color_ru,
			t.image_url,
			COALESCE(th.image_url, '') as product_image_url,
			SUM(s.price * (1 - COALESCE(p.discount, 0)/100) * o.count + m.delivery_price) as sum
		FROM orders o
		JOIN markets m ON o.market_id = m.id
		JOIN sizes s ON o.size_id = s.id
		JOIN products p ON o.product_id = p.id
		JOIN thumbnails t ON t.id = o.thumbnail_id
		JOIN thumbnails th ON th.id = o.thumbnail_id
		WHERE o.is_active = true AND o.user_id = ?`
	args := []interface{}{UserID}

	if status != "" {
		query += ` AND o.status = ?`
		args = append(args, status)
	}

	query += `
		GROUP BY o.id
		ORDER BY o.created_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []models.UserOrder = []models.UserOrder{}
	for rows.Next() {
		var o models.UserOrder
		if err := rows.Scan(&o.ID, &o.CartOrderID, &o.Status, &o.CreatedAt, &o.MarketName, &o.MarketNameRu, &o.ProductName, &o.ProductNameRu, &o.Size, &o.Color, &o.ColorRu, &o.ImageURL, &o.ProductImageURL, &o.Sum); err != nil {
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}
		o.Sum = math.Round(o.Sum*100) / 100 // Round to 2 decimal places
		orders = append(orders, o)
	}

	return orders, nil
}

// CreateMessage inserts a new message into the user_messages table
func (s *DBService) CreateMessage(userID int, req models.CreateMessageRequest) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert message
	result, err := tx.Exec(
		"INSERT INTO user_messages (user_id, full_name, phone, message) VALUES (?, ?, ?, ?)",
		userID, req.FullName, req.Phone, req.Message,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert message: %v", err)
	}

	// Get inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted message ID: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return int(id), nil
}
// UpdateSize updates a size entry
func (s *DBService) UpdateSize(sizeID int, req models.UpdateSizeRequest) (models.SizeUpdate, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.SizeUpdate{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Update size
	result, err := tx.Exec(
		"UPDATE sizes SET count = ?, price = ?, size = ? WHERE id = ?",
		req.Count, req.Price, req.Size, sizeID,
	)
	if err != nil {
		return models.SizeUpdate{}, fmt.Errorf("failed to update size: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.SizeUpdate{}, fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return models.SizeUpdate{}, fmt.Errorf("size not found or not associated with this market")
	}

	// Fetch updated size
	var size models.SizeUpdate
	err = tx.QueryRow("SELECT id, count, price, COALESCE(size, '') FROM sizes WHERE id = ?", sizeID).
		Scan(&size.ID, &size.Count, &size.Price, &size.Size)
	if err == sql.ErrNoRows {
		return models.SizeUpdate{}, fmt.Errorf("size not found after update")
	}
	if err != nil {
		return models.SizeUpdate{}, fmt.Errorf("failed to fetch updated size: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return models.SizeUpdate{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Fetch marketID to invalidate market-specific caches
	var marketID int
	err = s.db.QueryRow(`
        SELECT p.market_id
        FROM sizes s
        JOIN thumbnails t ON s.thumbnail_id = t.id
        JOIN products p ON t.product_id = p.id
        WHERE s.id = ?
    `, sizeID).Scan(&marketID)
	if err != nil {
		log.Printf("Failed to fetch marketID for size %d: %v", sizeID, err)
	} else {
		// Invalidate market-specific product caches
		marketKeys, err := s.redis.SMembers(context.Background(), fmt.Sprintf("market:%d:products_cache_keys", marketID)).Result()
		if err == nil && len(marketKeys) > 0 {
			s.redis.Del(context.Background(), marketKeys...)
		} else if err != nil && err != redis.Nil {
			log.Printf("Failed to invalidate market cache: %v", err)
		}
	}

	return size, nil
}

// UpdateThumbnail updates a thumbnail entry and returns the old image URL
func (s *DBService) UpdateThumbnail(thumbnailID int, color, color_ru string, marketID int, files []*multipart.FileHeader) (string, string, string, error) {
	var productID int
	var filePath string
	var imageCreated string
	err := s.db.QueryRow("SELECT product_id FROM thumbnails WHERE id = ?", thumbnailID).Scan(&productID)
	if err == sql.ErrNoRows {
		return "", "", "", fmt.Errorf("thumbnail not found: %v", err)
	}
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch product ID: %v", err)
	} 

	// Create uploads directory
	uploadsDir := filepath.Join("uploads/products", "")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("error creating directory: %v", err)
	}

	// Handle file upload
	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		return "", "", "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), fileHeader.Filename)
	filePath = filepath.Join(uploadsDir, filename)
	out, err := os.Create(filePath)
	if err != nil {
		return "", "", "", fmt.Errorf("error saving file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", "", "", fmt.Errorf("error copying file: %v", err)
	}
	imageCreated = "created"

	// Construct image URL
	imageURL := filepath.Join("/uploads/products", filename)
	imageURL = strings.ReplaceAll(imageURL, string(filepath.Separator), "/")

	// Update thumbnail in database
	thumb := ThumbnailData{
		ProductID: productID,
		Color:     color,
		ColorRu:   color_ru,
		ImageURL:  imageURL,
	}
	tx, err := s.db.Begin()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Fetch old image URL and verify market ownership
	var oldImageURL string
	err = tx.QueryRow(`
		SELECT t.image_url 
		FROM thumbnails t
		JOIN products p ON t.product_id = p.id
		WHERE t.id = ? AND p.market_id = ?
	`, thumbnailID, marketID).Scan(&oldImageURL)
	if err == sql.ErrNoRows {
		return "", "", "", fmt.Errorf("thumbnail not found or not associated with this market")
	}
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch thumbnail: %v", err)
	}

	// Update thumbnail
	result, err := tx.Exec(
		"UPDATE thumbnails SET color = ?, color_ru = ?, image_url = ? WHERE id = ?",
		thumb.Color, thumb.ColorRu, thumb.ImageURL, thumbnailID,
	)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to update thumbnail: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return "", "", "", fmt.Errorf("thumbnail not found")
	}

	if err := tx.Commit(); err != nil {
		return "", "", "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Invalidate cache
	s.redis.Del(context.Background(), fmt.Sprintf("market:%d:products:*", marketID))

	return oldImageURL, filePath, imageCreated, nil
}

// GetAllUserMessages retrieves all user messages
func (s *DBService) GetAllUserMessages() ([]models.UserMessage, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			COALESCE(full_name, ''), 
			COALESCE(phone, ''),
			COALESCE(message, '')
		FROM user_messages
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query user messages: %v", err)
	}
	defer rows.Close()

	var messages []models.UserMessage = []models.UserMessage{}
	for rows.Next() {
		var msg models.UserMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.FullName, &msg.Phone, &msg.Message); err != nil {
			return nil, fmt.Errorf("failed to scan user message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user messages: %v", err)
	}

	return messages, nil
}

// DeleteUserMessage deletes a user message by ID
func (s *DBService) DeleteUserMessage(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec("DELETE FROM user_messages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user message: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// CreateMarketMessage inserts a new message into the market_messages table
func (s *DBService) CreateMarketMessage(marketID int, req models.CreateMarketMessageRequest) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"INSERT INTO market_messages (market_id, full_name, phone, message) VALUES (?, ?, ?, ?)",
		marketID, req.FullName, req.Phone, req.Message,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert market message: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted message ID: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return int(id), nil
}

// GetAllMarketMessages retrieves all market messages
func (s *DBService) GetAllMarketMessages() ([]models.MarketMessage, error) {
	query := `
		SELECT id, market_id, COALESCE(full_name, ''), COALESCE(phone, ''), COALESCE(message, '')
		FROM market_messages
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query market messages: %v", err)
	}
	defer rows.Close()

	var messages []models.MarketMessage = []models.MarketMessage{}
	for rows.Next() {
		var msg models.MarketMessage
		if err := rows.Scan(&msg.ID, &msg.MarketID, &msg.FullName, &msg.Phone, &msg.Message); err != nil {
			return nil, fmt.Errorf("failed to scan market message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating market messages: %v", err)
	}

	return messages, nil
}

// DeleteMarketMessage deletes a market message by ID
func (s *DBService) DeleteMarketMessage(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec("DELETE FROM market_messages WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete market message: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
// GetUsers retrieves paginated users from the database with optional search
func (s *DBService) GetUsers(ctx context.Context, page, limit int, search string) ([]models.User, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 {
        limit = 10
    }
    offset := (page - 1) * limit

    // Build SQL query
    query := `
        SELECT id, full_name, phone, verified, created_at
        FROM users`
    args := []interface{}{}

    // Add search filter if provided
    if search != "" {
        query += ` WHERE (LOWER(full_name) LIKE ? OR LOWER(phone) LIKE ? OR CAST(id AS CHAR) LIKE ?)`
        searchTerm := "%" + strings.ToLower(search) + "%"
        args = append(args, searchTerm, searchTerm, searchTerm)
    }

    query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
    args = append(args, limit, offset)

    // Execute query
    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %v", err)
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var u models.User
        var createdAtStr string
        if err := rows.Scan(&u.ID, &u.FullName, &u.Phone, &u.Verified, &createdAtStr); err != nil {
            return nil, fmt.Errorf("failed to scan user: %v", err)
        }
        // Parse string to time.Time
        parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
        if err != nil {
            return nil, fmt.Errorf("failed to parse created_at: %v", err)
        }
        u.CreatedAt = parsedTime
        users = append(users, u)
    }

    return users, nil
}


// DeleteUser deletes a user by ID
func (s *DBService) DeleteUser(ctx context.Context, userID int) error {
    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Verify user exists
    var exists bool
    err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("failed to validate user: %v", err)
    }
    if !exists {
        return fmt.Errorf("user not found")
    }

    // Delete user
    result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        return fmt.Errorf("failed to delete user: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check rows affected: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    return nil
}

// UpdateUserVerified updates a user's verified status
func (s *DBService) UpdateUserVerified(ctx context.Context, userID int, verified bool) error {
    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Verify user exists
    var exists bool
    err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("failed to validate user: %v", err)
    }
    if !exists {
        return fmt.Errorf("user not found")
    }

    // Update verified status
    result, err := tx.ExecContext(ctx, "UPDATE users SET verified = ? WHERE id = ?", verified, userID)
    if err != nil {
        return fmt.Errorf("failed to update user: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check rows affected: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    return nil
}

// UpdateMarket updates a market and invalidates cache
func (s *DBService) UpdateMarket(ctx context.Context, marketID int, password string, deliveryPrice *float64, phone, name, nameRu, location, locationRu string, isVIP *bool, thumbnailURL string) (oldThumbnailURL string, newThumbnailURL string, err error) {
    thumbnailURL = filepath.Join("/uploads/markets", thumbnailURL)
    thumbnailURL = strings.ReplaceAll(thumbnailURL, string(filepath.Separator), "/")
	// Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return "", "", fmt.Errorf("failed to start transaction: %v", err)
    }
    defer tx.Rollback()

    // Verify market exists and fetch old thumbnail_url
    err = tx.QueryRowContext(ctx, "SELECT COALESCE(thumbnail_url, '') FROM markets WHERE id = ?", marketID).Scan(&oldThumbnailURL)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", "", fmt.Errorf("market not found")
        }
        return "", "", fmt.Errorf("failed to validate market: %v", err)
    }

    // Build dynamic UPDATE query
    var setClauses []string
    var args []interface{}

    if password != "" {
        passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            return "", "", fmt.Errorf("failed to hash password: %v", err)
        }
        setClauses = append(setClauses, "password = ?")
        args = append(args, passwordHash)
    }
    if deliveryPrice != nil {
        setClauses = append(setClauses, "delivery_price = ?")
        args = append(args, *deliveryPrice)
    }
    if phone != "" {
        setClauses = append(setClauses, "phone = ?")
        args = append(args, phone)
    }
    if name != "" {
        setClauses = append(setClauses, "name = ?")
        args = append(args, name)
    }
    if nameRu != "" {
        setClauses = append(setClauses, "name_ru = ?")
        args = append(args, nameRu)
    }
    if location != "" {
        setClauses = append(setClauses, "location = ?")
        args = append(args, location)
    }
    if locationRu != "" {
        setClauses = append(setClauses, "location_ru = ?")
        args = append(args, locationRu)
    }
    if isVIP != nil {
        setClauses = append(setClauses, "isVIP = ?")
        args = append(args, *isVIP)
    }
    if thumbnailURL != "" {
        setClauses = append(setClauses, "thumbnail_url = ?")
        args = append(args, thumbnailURL)
    }

    // If no fields to update, return early
    if len(setClauses) == 0 {
        if err := tx.Commit(); err != nil {
            return "", "", fmt.Errorf("failed to commit transaction: %v", err)
        }
        return oldThumbnailURL, thumbnailURL, nil
    }

    // Construct and execute UPDATE query
    query := fmt.Sprintf("UPDATE markets SET %s WHERE id = ?", strings.Join(setClauses, ", "))
    args = append(args, marketID)
    _, err = tx.ExecContext(ctx, query, args...)
    if err != nil {
        return "", "", fmt.Errorf("failed to update market: %v", err)
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return "", "", fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Invalidate caches
    pipe := s.redis.Pipeline()
    cacheKeys, err := s.redis.SMembers(ctx, "markets_cache_keys").Result()
    if err == nil && len(cacheKeys) > 0 {
        pipe.Del(ctx, cacheKeys...)
        pipe.Del(ctx, "markets_cache_keys")
    } else if err != nil && err != redis.Nil {
        log.Printf("Failed to fetch markets cache keys: %v", err)
    }
    if _, err := pipe.Exec(ctx); err != nil {
        log.Printf("Failed to invalidate caches: %v", err)
    }

    return oldThumbnailURL, thumbnailURL, nil
}
