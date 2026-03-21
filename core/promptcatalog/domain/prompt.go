package domain

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultCategoryName = "Default"
	FileName            = "system.md"
	MetaFileName        = "prompt.json"
	MaxImageURLs        = 3
)

var (
	ErrPromptNotFound   = errors.New("prompt not found")
	ErrPromptExists     = errors.New("prompt already exists")
	ErrEmptyContent     = errors.New("prompt content is empty")
	ErrInvalidName      = errors.New("invalid prompt name")
	ErrCategoryNotEmpty = errors.New("category not empty")
	ErrTooManyImages    = errors.New("prompt has too many image urls")
	ErrInvalidImageURL  = errors.New("prompt image url is invalid")
	ErrInvalidWebLink   = errors.New("prompt web link is invalid")
)

var promptMarkdownLinkPattern = regexp.MustCompile(`^\[(.+?)\]\((.+?)\)$`)

type PromptLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Prompt struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Category    string       `json:"category"`
	Path        string       `json:"path"`
	FilePath    string       `json:"filePath"`
	Content     string       `json:"content"`
	ImageURLs   []string     `json:"imageURLs,omitempty"`
	WebLinks    []PromptLink `json:"webLinks,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

func NormalizeCategoryName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, DefaultCategoryName) {
		return DefaultCategoryName, nil
	}
	if err := validatePathSegment(trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

func NormalizePromptName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidName
	}
	if err := validatePathSegment(trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

func ParseWebLinksMarkdown(raw string) ([]PromptLink, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	links := make([]PromptLink, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		matches := promptMarkdownLinkPattern.FindStringSubmatch(trimmed)
		if len(matches) != 3 {
			return nil, ErrInvalidWebLink
		}
		link, err := normalizePromptLink(PromptLink{
			Label: matches[1],
			URL:   matches[2],
		})
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func NormalizePromptImageURLs(imageURLs []string) ([]string, error) {
	if len(imageURLs) == 0 {
		return nil, nil
	}
	normalized := make([]string, 0, len(imageURLs))
	for _, rawURL := range imageURLs {
		trimmed := strings.TrimSpace(rawURL)
		if trimmed == "" {
			continue
		}
		validURL, err := normalizePromptURL(trimmed)
		if err != nil {
			return nil, ErrInvalidImageURL
		}
		normalized = append(normalized, validURL)
		if len(normalized) > MaxImageURLs {
			return nil, ErrTooManyImages
		}
	}
	return normalized, nil
}

func NormalizePromptLinks(links []PromptLink) ([]PromptLink, error) {
	if len(links) == 0 {
		return nil, nil
	}
	normalized := make([]PromptLink, 0, len(links))
	for _, link := range links {
		if strings.TrimSpace(link.Label) == "" && strings.TrimSpace(link.URL) == "" {
			continue
		}
		validLink, err := normalizePromptLink(link)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, validLink)
	}
	return normalized, nil
}

func NextAvailableName(base string, existing map[string]struct{}) string {
	candidate := base
	index := 2
	for {
		candidate = fmt.Sprintf("%s-%d", base, index)
		if _, exists := existing[candidate]; !exists {
			return candidate
		}
		index++
	}
}

func CompareFilesystemEntries(currentPath, targetPath string) (bool, bool, error) {
	currentInfo, err := osStat(currentPath)
	if err != nil {
		return false, false, err
	}
	targetInfo, err := osStat(targetPath)
	if err != nil {
		if isNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, sameFile(currentInfo, targetInfo), nil
}

func PromptDir(root, category, name string) string {
	return filepath.Join(root, category, name)
}

func validatePathSegment(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return ErrInvalidName
	}
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, string(filepath.Separator)) {
		return ErrInvalidName
	}
	invalidChars := `<>:"\\|?*`
	if strings.ContainsAny(trimmed, invalidChars) {
		return ErrInvalidName
	}
	if strings.HasSuffix(trimmed, ".") || strings.HasSuffix(trimmed, " ") {
		return ErrInvalidName
	}
	switch strings.ToUpper(trimmed) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return ErrInvalidName
	}
	return nil
}

func normalizePromptLink(link PromptLink) (PromptLink, error) {
	label := strings.TrimSpace(link.Label)
	rawURL := strings.TrimSpace(link.URL)
	if label == "" || rawURL == "" {
		return PromptLink{}, ErrInvalidWebLink
	}
	validURL, err := normalizePromptURL(rawURL)
	if err != nil {
		return PromptLink{}, ErrInvalidWebLink
	}
	return PromptLink{Label: label, URL: validURL}, nil
}

func normalizePromptURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported prompt url scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("prompt url host is empty")
	}
	return parsed.String(), nil
}
