package provider

import (
	"os"
	"path/filepath"
	"strings"
)

// FileProvider отвечает за получение списка текстовых файлов
type FileProvider interface {
	ListTextFiles() ([]string, error)
}

// DiskFileProvider реализует FileProvider для работы с файловой системой
type DiskFileProvider struct {
	path string
}

// NewDiskFileProvider создает новый провайдер файлов
func NewDiskFileProvider(path string) *DiskFileProvider {
	return &DiskFileProvider{
		path: path,
	}
}

// ListTextFiles возвращает все .txt файлы в указанной директории
func (p *DiskFileProvider) ListTextFiles() ([]string, error) {
	var textFiles []string

	err := filepath.Walk(p.path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(filePath, ".txt") {
			textFiles = append(textFiles, filePath)
		}

		return nil
	})

	return textFiles, err
}
