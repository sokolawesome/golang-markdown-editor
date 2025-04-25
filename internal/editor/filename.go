package editor

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func generateNewFilename() string {
	timestamp := time.Now().Format("20250101-100105")
	return fmt.Sprintf("note-%s.md", timestamp)
}

func sanitizeFilename(input string) string {
	cleaned := regexp.MustCompile(`^#+\s*`).ReplaceAllString(input, "")
	cleaned = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return "untitled"
	}
	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}

	return strings.ToLower(cleaned)
}
