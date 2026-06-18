package rag

import (
	"fmt"
	"regexp"
	"strings"
)

type PaperChunk struct {
	ID           string
	PaperID      string
	Title        string
	Section      string
	Content      string
	Page         int
	FigureRefs   []string
	CitationRefs []string
	Metadata     map[string]string
}

var sectionPattern = regexp.MustCompile(`(?im)^\s*(abstract|introduction|related work|method|methods|methodology|experiment|experiments|results?|discussion|conclusion|references)\s*$`)
var figurePattern = regexp.MustCompile(`(?i)\b(fig\.?|figure)\s*\d+`)
var citationPattern = regexp.MustCompile(`\[[0-9,\-\s]+\]`)

func ChunkBySection(text string) []PaperChunk {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	matches := sectionPattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return ChunkWithSlidingWindow("Unknown", text)
	}

	var chunks []PaperChunk
	for i, match := range matches {
		section := strings.TrimSpace(text[match[0]:match[1]])
		start := match[1]
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		content := strings.TrimSpace(text[start:end])
		chunks = append(chunks, ChunkWithSlidingWindow(normalizeSection(section), content)...)
	}
	return chunks
}

func ChunkWithSlidingWindow(section string, content string) []PaperChunk {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	const windowSize = 1200
	const overlap = 180

	var chunks []PaperChunk
	runes := []rune(content)
	for start := 0; start < len(runes); {
		end := start + windowSize
		if end > len(runes) {
			end = len(runes)
		}
		part := strings.TrimSpace(string(runes[start:end]))
		if part != "" {
			chunks = append(chunks, PaperChunk{
				ID:           fmt.Sprintf("%s-%03d", slug(section), len(chunks)+1),
				Section:      section,
				Content:      part,
				FigureRefs:   figurePattern.FindAllString(part, -1),
				CitationRefs: citationPattern.FindAllString(part, -1),
				Metadata: map[string]string{
					"chunk_strategy": "section_first_sliding_window",
					"note":           "后续可扩展 parent-child chunking，保留章节级父节点和细粒度子块。",
				},
			})
		}
		if end == len(runes) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

func normalizeSection(section string) string {
	section = strings.ToLower(strings.TrimSpace(section))
	switch section {
	case "methods", "methodology":
		return "Method"
	case "experiments":
		return "Experiment"
	case "result":
		return "Result"
	default:
		if section == "" {
			return "Unknown"
		}
		return strings.ToUpper(section[:1]) + section[1:]
	}
}

func slug(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")
	if text == "" {
		return "chunk"
	}
	return text
}
