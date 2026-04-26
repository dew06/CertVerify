package ipfs

import (
	"bytes"
	"errors"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

type Client struct {
	sh *shell.Shell
}

// NewClient creates IPFS client
func NewClient(apiURL string) *Client {
	return &Client{
		sh: shell.NewShell(apiURL),
	}
}

// UploadPDF uploads PDF to IPFS and pins it
func (c *Client) UploadPDF(pdfBytes []byte) (string, error) {
	// Upload
	cid, err := c.sh.Add(bytes.NewReader(pdfBytes))
	if err != nil {
		return "", err
	}

	// Pin (prevents garbage collection)
	err = c.sh.Pin(cid)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// DownloadPDF retrieves PDF from IPFS
func (c *Client) DownloadPDF(cid string) ([]byte, error) {
	reader, err := c.sh.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// CheckConnection verifies IPFS is running
func (c *Client) CheckConnection() error {
	_, err := c.sh.ID()
	if err != nil {
		return errors.New("IPFS not running! Open IPFS Desktop")
	}
	return nil
}
