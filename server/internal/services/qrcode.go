package services

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// QRCodeService handles QR code generation
type QRCodeService struct {
	baseVerificationURL string
}

// NewQRCodeService creates a new QR code service
func NewQRCodeService(baseVerificationURL string) *QRCodeService {
	return &QRCodeService{
		baseVerificationURL: baseVerificationURL,
	}
}

// GenerateQRCode generates QR code for certificate verification
func (q *QRCodeService) GenerateQRCode(certID string) ([]byte, error) {
	// Build verification URL
	verifyURL := fmt.Sprintf("%s/verify/%s", q.baseVerificationURL, certID)

	// Generate QR code
	qr, err := qrcode.New(verifyURL, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %v", err)
	}

	// Disable border for cleaner look
	qr.DisableBorder = true

	// Generate PNG at 256x256
	return qr.PNG(256)
}

// GenerateQRCodeToFile generates QR code and saves to file
func (q *QRCodeService) GenerateQRCodeToFile(certID, filePath string) error {
	verifyURL := fmt.Sprintf("%s/verify/%s", q.baseVerificationURL, certID)
	return qrcode.WriteFile(verifyURL, qrcode.Medium, 256, filePath)
}
