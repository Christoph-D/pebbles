package peb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrInvalidFormat = errors.New("invalid peb file format")

func WriteFile(pebblesDir string, peb *Peb) error {
	filename := Filename(peb)
	filepath := filepath.Join(pebblesDir, filename)

	var buf bytes.Buffer

	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	if err := encoder.Encode(peb); err != nil {
		return fmt.Errorf("failed to encode peb: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	buf.WriteString("---\n")
	buf.WriteString(peb.Content)

	if err := os.WriteFile(filepath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write peb file: %w", err)
	}

	return nil
}

func ReadFile(path string) (*Peb, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read peb file: %w", err)
	}

	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		return nil, ErrInvalidFormat
	}

	endMarker := strings.Index(content[len("---\n"):], "---\n")
	if endMarker == -1 {
		return nil, ErrInvalidFormat
	}

	frontmatterEnd := len("---\n") + endMarker
	frontmatter := content[len("---\n"):frontmatterEnd]
	bodyContent := content[frontmatterEnd+len("---\n"):]

	peb := &Peb{}
	if err := yaml.Unmarshal([]byte(frontmatter), peb); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	peb.Content = bodyContent

	return peb, nil
}
