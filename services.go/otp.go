package services

import (
	"math/rand"
	"time"
	"fmt"
)

// GenerateOTP generates a 6-digit OTP
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}