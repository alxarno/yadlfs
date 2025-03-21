package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
)

var (
	ErrCreateRequest          = errors.New("failed to create request")
	ErrRequestUploadURL       = errors.New("failed to request upload URL")
	ErrDecodeUploadResponse   = errors.New("failed to decode upload URL response")
	ErrUploadFile             = errors.New("failed to upload file")
	ErrRequestDownloadURL     = errors.New("failed to request download URL")
	ErrDecodeDownloadResponse = errors.New("failed to decode download URL response")
	ErrDownloadFile           = errors.New("failed to download file")
)

type yandexDiskClientResponse struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

// YandexDiskClient represents a client for interacting with Yandex Disk API.
type YandexDiskClient struct {
	OAuthToken string
	DiskFolder string
	BaseURL    string
}

// NewYandexDiskClient creates a new YandexDiskClient with the provided OAuth token.
func NewYandexDiskClient(oauthToken string, diskFolder string) *YandexDiskClient {
	return &YandexDiskClient{
		OAuthToken: oauthToken,
		DiskFolder: diskFolder,
		BaseURL:    "https://cloud-api.yandex.net/v1/disk",
	}
}

// Upload uploads a file to Yandex Disk.
func (c *YandexDiskClient) Upload(ctx context.Context, filePath string, file io.Reader, overwrite bool) error {
	// Step 1: Request upload URL
	filePath = filepath.Join(c.DiskFolder, filePath)
	uploadURL := fmt.Sprintf("%s/resources/upload?path=%s&overwrite=%t", c.BaseURL, url.QueryEscape(filePath), overwrite)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uploadURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header.Set("Authorization", "OAuth "+c.OAuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestUploadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrRequestUploadURL, resp.Status)
	}

	uploadResponse := yandexDiskClientResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&uploadResponse); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeUploadResponse, err)
	}

	// Step 2: Upload the file
	req, err = http.NewRequestWithContext(ctx, uploadResponse.Method, uploadResponse.Href, file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUploadFile, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w: %s", ErrUploadFile, resp.Status)
	}

	return nil
}

// Download downloads a file from Yandex Disk.
func (c *YandexDiskClient) Download(ctx context.Context, filePath string) (io.ReadCloser, error) {
	// Step 1: Request download URL
	filePath = filepath.Join(c.DiskFolder, filePath)
	downloadURL := fmt.Sprintf("%s/resources/download?path=%s", c.BaseURL, url.QueryEscape(filePath))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header.Set("Authorization", "OAuth "+c.OAuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestDownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrRequestDownloadURL, resp.Status)
	}

	downloadResponse := yandexDiskClientResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&downloadResponse); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeDownloadResponse, err)
	}

	// Step 2: Download the file
	req, err = http.NewRequestWithContext(ctx, downloadResponse.Method, downloadResponse.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header.Set("Authorization", "OAuth "+c.OAuthToken)

	//nolint:bodyclose //resp.Body is io.ReadCloser, and will be closed by the caller
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDownloadFile, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDownloadFile, resp.Status)
	}

	return resp.Body, nil
}
