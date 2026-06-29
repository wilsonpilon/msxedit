package reader

import "strings"

// renderMarkdown faz um render leve de markdown para exibição: devolve as linhas
// já limpas e o conjunto de índices que são títulos (para realce no viewer).
// wrapWidth > 0 quebra linhas longas nessa largura (0 = sem limite).
//
// Tratamento aplicado:
//   - títulos `#`..`######`  → texto sem os `#`, marcados como heading
//   - links `[texto](url)`   → apenas `texto`
//   - itens de lista `- `/`* ` → `• `
//   - cercas de código ```` ``` ```` → removidas
func renderMarkdown(data []byte, wrapWidth int) ([]string, map[int]bool) {
	raw := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(raw))
	headings := make(map[int]bool)

	for _, line := range raw {
		trimmed := strings.TrimSpace(line)

		// Cercas de código: descartar a linha de marcação.
		if strings.HasPrefix(trimmed, "```") {
			continue
		}

		// Títulos.
		if strings.HasPrefix(trimmed, "#") {
			title := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			headings[len(lines)] = true
			lines = append(lines, title)
			continue
		}

		// Itens de lista.
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		body := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(body, "- ") || strings.HasPrefix(body, "* ") {
			body = "• " + body[2:]
		}

		processed := indent + stripInlineMarkdown(body)
		if wrapWidth > 0 {
			wrapped := wordWrap(processed, wrapWidth)
			// Marcar linha de título também nas linhas resultantes do wrap.
			for _, wl := range wrapped {
				lines = append(lines, wl)
			}
		} else {
			lines = append(lines, processed)
		}
	}

	return lines, headings
}

// wordWrap quebra line em múltiplas linhas de no máximo width runas,
// preservando a indentação original nas linhas de continuação.
func wordWrap(line string, width int) []string {
	runes := []rune(line)
	if len(runes) <= width {
		return []string{line}
	}

	// Calcula indentação para linhas de continuação.
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	indentLen := len([]rune(indent))

	var result []string
	for len(runes) > width {
		cut := width
		// Recua até encontrar um espaço onde cortar.
		for cut > indentLen && runes[cut] != ' ' {
			cut--
		}
		if cut <= indentLen {
			// Nenhum espaço dentro do limite: avança até o próximo espaço.
			cut = width
			for cut < len(runes) && runes[cut] != ' ' {
				cut++
			}
		}
		result = append(result, strings.TrimRight(string(runes[:cut]), " "))
		rest := strings.TrimLeft(string(runes[cut:]), " ")
		runes = []rune(indent + rest)
	}
	result = append(result, string(runes))
	return result
}

// stripInlineMarkdown remove a marcação de links, mantendo apenas o texto visível.
func stripInlineMarkdown(s string) string {
	runes := []rune(s)
	var out strings.Builder

	for i := 0; i < len(runes); {
		if runes[i] == '[' {
			if labelEnd := indexRune(runes, ']', i+1); labelEnd >= 0 &&
				labelEnd+1 < len(runes) && runes[labelEnd+1] == '(' {
				if targetEnd := indexRune(runes, ')', labelEnd+2); targetEnd >= 0 {
					out.WriteString(string(runes[i+1 : labelEnd]))
					i = targetEnd + 1
					continue
				}
			}
		}
		out.WriteRune(runes[i])
		i++
	}
	return out.String()
}

func indexRune(runes []rune, target rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
}
