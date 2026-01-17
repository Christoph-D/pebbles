package peb

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
)

func GenerateID(prefix string, length int) (string, error) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return fmt.Sprintf("%s-%s", prefix, string(result)), nil
}

func Filename(peb *Peb) string {
	slug := slugifyTitle(peb.Title)
	return fmt.Sprintf("peb-%s--%s.md", peb.ID, slug)
}

func slugifyTitle(title string) string {
	slug := strings.ToLower(title)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if r == '-' {
			return r
		}
		return r
	}, slug)
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, "-", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "untitled"
	}
	return slug
}

func ParseID(filename string) (string, error) {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, ".md")
	if !strings.HasPrefix(base, "peb-") {
		return "", errors.New("invalid peb filename format")
	}
	parts := strings.SplitN(base, "-", 3)
	if len(parts) < 3 {
		return "", errors.New("invalid peb filename format")
	}
	return fmt.Sprintf("peb-%s", parts[1]), nil
}
