package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

// PDFGeneratorService handles certificate PDF generation
type PDFGeneratorService struct {
	qrService *QRCodeService
}

// NewPDFGeneratorService creates a new PDF generator
func NewPDFGeneratorService(qrService *QRCodeService) *PDFGeneratorService {
	return &PDFGeneratorService{
		qrService: qrService,
	}
}

// CertificateData contains all data for certificate generation
type CertificateData struct {
	CertID           string
	StudentName      string
	Degree           string
	University       string
	UniversityDomain string
	IssueDate        time.Time
	GPA              float64
}

// GenerateCertificatePDF generates a filled, professional certificate
func (p *PDFGeneratorService) GenerateCertificatePDF(data CertificateData) ([]byte, error) {
	// Create new PDF (L = Landscape, mm = millimeters, A4)
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pageWidth := 297.0
	pageHeight := 210.0

	// 1. Decorative Borders
	p.drawBorder(pdf, pageWidth, pageHeight)

	// 2. University Header (Top)
	p.drawHeader(pdf, data.University, pageWidth)

	// 3. Title Section
	p.drawTitle(pdf, pageWidth)

	// 4. Main Body Content (Expanded to fill space)
	p.drawBody(pdf, data, pageWidth)

	// 5. QR Code (Bottom Right)
	qrBytes, err := p.generateQRCode(data.CertID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %v", err)
	}
	p.addQRCode(pdf, qrBytes, pageWidth, pageHeight)

	// 6. Technical Footer (Bottom Left)
	p.drawFooter(pdf, data.CertID, pageHeight)

	// Generate PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *PDFGeneratorService) drawBorder(pdf *gofpdf.Fpdf, width, height float64) {
	// Thick Outer Border
	pdf.SetLineWidth(2.5)
	pdf.SetDrawColor(25, 99, 235) // Deep Blue
	pdf.Rect(10, 10, width-20, height-20, "D")

	// Thin Inner Border
	pdf.SetLineWidth(0.6)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Rect(13, 13, width-26, height-26, "D")
}

func (p *PDFGeneratorService) drawHeader(pdf *gofpdf.Fpdf, university string, width float64) {
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(25, 99, 235)
	pdf.SetXY(0, 25)
	pdf.CellFormat(width, 15, university, "", 0, "C", false, 0, "")

	// Decorative Underline
	pdf.SetDrawColor(124, 58, 237) // Purple accents
	pdf.SetLineWidth(0.8)
	pdf.Line(width/2-50, 42, width/2+50, 42)
}

func (p *PDFGeneratorService) drawTitle(pdf *gofpdf.Fpdf, width float64) {
	pdf.SetFont("Arial", "B", 32)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(0, 55)
	pdf.CellFormat(width, 20, "CERTIFICATE OF ACHIEVEMENT", "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "I", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(0, 75)
	pdf.CellFormat(width, 10, "This is to certify that", "", 0, "C", false, 0, "")
}

func (p *PDFGeneratorService) drawBody(pdf *gofpdf.Fpdf, data CertificateData, width float64) {
	centerX := width / 2
	currentY := 90.0

	// Student Name (Prominent center-piece)
	pdf.SetFont("Arial", "B", 42)
	pdf.SetTextColor(25, 99, 235)
	pdf.SetXY(0, currentY)
	pdf.CellFormat(width, 25, data.StudentName, "", 0, "C", false, 0, "")
	currentY += 28

	// Simple decorative line
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(1.0)
	pdf.Line(centerX-70, currentY, centerX+70, currentY)
	currentY += 15

	// completion text
	pdf.SetFont("Arial", "", 16)
	pdf.SetTextColor(70, 70, 70)
	pdf.SetXY(0, currentY)
	pdf.CellFormat(width, 10, "has successfully fulfilled the requirements for the degree of", "", 0, "C", false, 0, "")
	currentY += 18

	// Degree Name
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(40, 40, 40)
	pdf.SetXY(0, currentY)
	pdf.CellFormat(width, 15, data.Degree, "", 0, "C", false, 0, "")
	currentY += 22

	// GPA and Date
	pdf.SetFont("Arial", "", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(0, currentY)
	gpaStr := ""
	if data.GPA > 0 {
		gpaStr = fmt.Sprintf(" with a GPA of %.2f,", data.GPA)
	}
	infoText := fmt.Sprintf("awarded%s on this day, %s", gpaStr, data.IssueDate.Format("January 2, 2006"))
	pdf.CellFormat(width, 10, infoText, "", 0, "C", false, 0, "")
}

func (p *PDFGeneratorService) drawFooter(pdf *gofpdf.Fpdf, certID string, height float64) {
	footerY := height - 25
	leftMargin := 20.0

	// Certificate ID
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetXY(leftMargin, footerY)
	pdf.CellFormat(100, 5, fmt.Sprintf("Verification ID: %s", certID), "", 0, "L", false, 0, "")

	// Instruction
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(140, 140, 140)
	pdf.SetXY(leftMargin, footerY+7)
	pdf.CellFormat(150, 5, "Authenticity can be verified via the university portal or QR scan.", "", 0, "L", false, 0, "")
}

func (p *PDFGeneratorService) generateQRCode(certID string) ([]byte, error) {
	// Update with your specific redirect URL
	verifyURL := fmt.Sprintf("http://localhost:5173/verify/%s", certID)
	qr, err := qrcode.New(verifyURL, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = true
	return qr.PNG(256)
}

func (p *PDFGeneratorService) addQRCode(pdf *gofpdf.Fpdf, qrBytes []byte, width, height float64) {
	img, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		return
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	pdf.RegisterImageReader("qr", "PNG", &buf)

	qrSize := 40.0
	qrX := width - qrSize - 20
	qrY := height - qrSize - 20

	pdf.ImageOptions("qr", qrX, qrY, qrSize, qrSize, false, gofpdf.ImageOptions{}, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(qrX, qrY+qrSize+2)
	pdf.CellFormat(qrSize, 3, "SCAN TO VERIFY", "", 0, "C", false, 0, "")
}

func (p *PDFGeneratorService) ComputePDFHash(pdfBytes []byte) string {
	hash := sha256.Sum256(pdfBytes)
	return hex.EncodeToString(hash[:])
}
