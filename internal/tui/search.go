package tui

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// searchMatch é uma ocorrência encontrada no texto, como offsets em bytes.
type searchMatch struct {
	start, end int
}

// findMatches retorna todas as ocorrências de p.Text dentro de text,
// respeitando Case sensitive / Whole words only / Regular expression.
func findMatches(text string, p FindParams) ([]searchMatch, error) {
	if p.Text == "" {
		return nil, nil
	}

	if p.Regex {
		pattern := p.Text
		if !p.CaseSensitive {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		idxs := re.FindAllStringIndex(text, -1)
		matches := make([]searchMatch, 0, len(idxs))
		for _, idx := range idxs {
			if idx[0] == idx[1] {
				continue // ignora ocorrências vazias (ex.: "a*")
			}
			matches = append(matches, searchMatch{idx[0], idx[1]})
		}
		return matches, nil
	}

	haystack := text
	needle := p.Text
	if !p.CaseSensitive {
		haystack = strings.ToLower(haystack)
		needle = strings.ToLower(needle)
	}

	var matches []searchMatch
	pos := 0
	for pos <= len(haystack) {
		idx := strings.Index(haystack[pos:], needle)
		if idx < 0 {
			break
		}
		mStart := pos + idx
		mEnd := mStart + len(needle)
		if !p.WholeWords || isWholeWord(text, mStart, mEnd) {
			matches = append(matches, searchMatch{mStart, mEnd})
		}
		pos = mStart + 1 // avança 1 byte para permitir ocorrências adjacentes
	}
	return matches, nil
}

// isWholeWord informa se o trecho text[start:end] não está colado a
// caracteres de palavra nas bordas.
func isWholeWord(text string, start, end int) bool {
	isWordChar := func(r rune) bool {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:start])
		if isWordChar(r) {
			return false
		}
	}
	if end < len(text) {
		r, _ := utf8.DecodeRuneInString(text[end:])
		if isWordChar(r) {
			return false
		}
	}
	return true
}

// findNext localiza a próxima ocorrência de p.Text a partir da posição do
// cursor (ou do início/fim do escopo, se p.EntireScope), respeitando
// Direction, Scope e Origin, e seleciona o trecho encontrado. Retorna se
// algo foi encontrado, se a busca deu a volta (wrap) e um erro (regex inválida).
func (w *editorWindow) findNext(p FindParams) (found bool, wrapped bool, err error) {
	if strings.TrimSpace(p.Text) == "" {
		return false, false, nil
	}

	fullText := w.editor.GetText()
	scopeStart, scopeEnd := 0, len(fullText)
	if p.SelectedOnly {
		if !w.blockValid() {
			return false, false, nil
		}
		scopeStart, scopeEnd = w.blockByteRange()
	}

	matches, err := findMatches(fullText[scopeStart:scopeEnd], p)
	if err != nil {
		return false, false, err
	}
	if len(matches) == 0 {
		return false, false, nil
	}
	for i := range matches {
		matches[i].start += scopeStart
		matches[i].end += scopeStart
	}

	var target *searchMatch
	if p.EntireScope {
		if p.Backward {
			target = &matches[len(matches)-1]
		} else {
			target = &matches[0]
		}
	} else {
		curOff := w.cursorByteOffset()
		if p.Backward {
			for i := len(matches) - 1; i >= 0; i-- {
				if matches[i].start < curOff {
					target = &matches[i]
					break
				}
			}
			if target == nil {
				target = &matches[len(matches)-1]
				wrapped = true
			}
		} else {
			for i := range matches {
				if matches[i].start > curOff {
					target = &matches[i]
					break
				}
			}
			if target == nil {
				target = &matches[0]
				wrapped = true
			}
		}
	}

	w.editor.Select(target.start, target.end)
	w.blkBeginRow, w.blkBeginCol = byteOffsetToRowCol(fullText, target.start)
	w.blkEndRow, w.blkEndCol = byteOffsetToRowCol(fullText, target.end)
	w.blkVisible = true
	return true, wrapped, nil
}
