package services

import (
	"crypto/rand"
	//"database/sql"
	"fmt"
	"math/big"
	"time"
)

// GenerateVerificationCode generates and stores a 4-digit code
func GenerateVerificationCode(db *DBService, phone, fullName string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %v", err)
	}
	code := fmt.Sprintf("%04d", n)

	err = saveVerificationCode(db, phone, code, fullName)
	if err != nil {
		return "", err
	}

	return code, nil
}

// saveVerificationCode stores the verification code in the database
func saveVerificationCode(db *DBService, phone, code, fullName string) error {
	expiresAt := time.Now().Add(5 * time.Minute)
	_, err := db.db.Exec(
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