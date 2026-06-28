package tui

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type markdownTopicBuilder struct {
	id    string
	title string
	lines []string
	links []LinkInfo
}

func (hc *HelpContent) loadTopicsFromMarkdown() error {
	path, err := findHelpMarkdownPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	topics, rootID, err := parseHelpMarkdown(string(data))
	if err != nil {
		return err
	}
	if len(topics) == 0 {
		return os.ErrInvalid
	}

	hc.topics = topics
	hc.currentTopicID = rootID
	if _, ok := hc.topics[hc.currentTopicID]; !ok {
		for id := range hc.topics {
			hc.currentTopicID = id
			break
		}
	}
	return nil
}

func findHelpMarkdownPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "HELP.md")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func parseHelpMarkdown(src string) (map[string]*HelpTopic, string, error) {
	lines := strings.Split(src, "\n")
	topics := make(map[string]*HelpTopic)
	var current *markdownTopicBuilder
	rootID := ""
	rootTitle := ""

	finalize := func() {
		if current == nil {
			return
		}
		topics[current.id] = &HelpTopic{
			ID:    current.id,
			Title: current.title,
			Lines: trimTrailingBlankLines(current.lines),
			Links: current.links,
		}
		if rootID == "" {
			rootID = current.id
			rootTitle = current.title
		}
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if isMarkdownTopicHeading(trimmed) {
			finalize()
			title := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			id := slugifyAnchor(title)
			if strings.EqualFold(title, "Help Contents") || strings.EqualFold(title, "Contents") {
				id = "contents"
				title = "Help"
			}
			current = &markdownTopicBuilder{id: id, title: title}
			if rootID == "" {
				rootID = id
				rootTitle = title
			}
			continue
		}
		if current == nil {
			continue
		}
		plain, links := parseMarkdownLine(line, len(current.lines))
		current.lines = append(current.lines, plain)
		current.links = append(current.links, links...)
	}
	finalize()

	if rootID == "" {
		return nil, "", os.ErrInvalid
	}
	if rootTitle == "" {
		rootTitle = "Help"
	}
	if topic, ok := topics[rootID]; ok && topic.Title == "" {
		topic.Title = rootTitle
	}
	return topics, rootID, nil
}

func isMarkdownTopicHeading(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	count := 0
	for _, r := range line {
		if r != '#' {
			break
		}
		count++
	}
	return count == 1 || count == 2
}

func parseMarkdownLine(line string, row int) (string, []LinkInfo) {
	runes := []rune(line)
	out := make([]rune, 0, len(runes))
	links := make([]LinkInfo, 0, 2)

	for i := 0; i < len(runes); {
		if runes[i] == '[' {
			if labelEnd := findRune(runes, ']', i+1); labelEnd >= 0 && labelEnd+1 < len(runes) && runes[labelEnd+1] == '(' {
				if targetEnd := findRune(runes, ')', labelEnd+2); targetEnd >= 0 {
					label := string(runes[i+1 : labelEnd])
					target := strings.TrimSpace(string(runes[labelEnd+2 : targetEnd]))
					if strings.HasPrefix(target, "#") {
						target = target[1:]
					}
					target = slugifyAnchor(target)
					colStart := len(out)
					labelRunes := []rune(label)
					out = append(out, labelRunes...)
					links = append(links, LinkInfo{
						Text:     label,
						TopicID:  target,
						Row:      row,
						ColStart: colStart,
						ColEnd:   colStart + len(labelRunes),
					})
					i = targetEnd + 1
					continue
				}
			}
		}
		out = append(out, runes[i])
		i++
	}
	return string(out), links
}

func findRune(runes []rune, target rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
}

func slugifyAnchor(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastSep := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastSep = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '&':
			if b.Len() > 0 && !lastSep {
				b.WriteRune('_')
				lastSep = true
			}
		default:
			if b.Len() > 0 && !lastSep {
				b.WriteRune('_')
				lastSep = true
			}
		}
	}
	result := strings.Trim(b.String(), "_")
	if result == "" {
		return "topic"
	}
	return result
}

func trimTrailingBlankLines(lines []string) []string {
	end := len(lines)
	for end > 0 {
		if strings.TrimSpace(lines[end-1]) != "" {
			break
		}
		end--
	}
	if end < 0 {
		end = 0
	}
	return append([]string(nil), lines[:end]...)
}


