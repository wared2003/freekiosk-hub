package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Sound info structure for the UI
type SoundFileInfo struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}

// Service definition
type MediaService interface {
	Upload(name string, content io.ReadSeeker) (*SoundFileInfo, error)
	List() ([]SoundFileInfo, error)
	Delete(name string) error
}

type mediaService struct {
	storageDir string
	baseURL    string
	soundDir   string
}

func NewMediaService(storageDir, baseURL string) MediaService {
	// Create storage dir if not exists
	_ = os.MkdirAll(storageDir, 0755)
	_ = os.MkdirAll(filepath.Join(storageDir, "sounds"), 0755)

	return &mediaService{
		storageDir: storageDir,
		baseURL:    baseURL,
		soundDir:   filepath.Join(storageDir, "sounds"),
	}
}

func (s *mediaService) Upload(name string, content io.ReadSeeker) (*SoundFileInfo, error) {
	buffer := make([]byte, 512)
	if _, err := content.Read(buffer); err != nil {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	content.Seek(0, 0)

	contentType := http.DetectContentType(buffer)
	if contentType != "audio/mpeg" && contentType != "audio/wav" && contentType != "audio/ogg" {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	safeName := filepath.Base(name)
	dstPath := filepath.Join(s.soundDir, safeName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, content); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &SoundFileInfo{
		Name:      safeName,
		Extension: filepath.Ext(safeName),
		URL:       fmt.Sprintf("%s/media/sounds%s", s.baseURL, name),
	}, nil
}

func (s *mediaService) List() ([]SoundFileInfo, error) {
	files, err := os.ReadDir(s.soundDir)
	if err != nil {
		return nil, err
	}

	var list []SoundFileInfo
	for _, f := range files {
		if !f.IsDir() {
			name := f.Name()
			list = append(list, SoundFileInfo{
				Name:      name,
				Extension: filepath.Ext(name),
				URL:       s.getSoundURL(name),
			})
		}
	}
	return list, nil
}

func (s *mediaService) Delete(name string) error {
	safeName := filepath.Base(name)
	return os.Remove(filepath.Join(s.storageDir, safeName))
}

func (s *mediaService) getSoundURL(filename string) string {
	base := strings.TrimSuffix(s.baseURL, "/")
	safeName := url.PathEscape(filename)
	return fmt.Sprintf("http://%s/media/sounds/%s", base, safeName)
}
