package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"msxedit/internal/basic"
)

// DocType identifica o tipo de conteúdo exibido.
type DocType string

const (
	TypeAuto DocType = "auto"
	TypeText DocType = "txt"
	TypeBas  DocType = "bas"
	TypeMD   DocType = "md"
)

// Document é o conteúdo carregado e pronto para o viewer.
type Document struct {
	FileName string
	Type     DocType
	Lines    []string
	Headings map[int]bool // índices de linha que são títulos (markdown)
	Spans    [][]Span     // spans de realce por linha; preenchido apenas para TypeBas
}

// LoadDocument lê o arquivo em path, detecta o tipo (ou usa forceType quando
// diferente de TypeAuto) e devolve o Document com as linhas prontas.
// wrapWidth define a largura máxima para quebra de linha em markdown (0 = sem limite).
func LoadDocument(path string, forceType DocType, tabSize, wrapWidth int) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dt := forceType
	if dt == "" || dt == TypeAuto {
		dt = detectType(path, data)
	}

	doc := &Document{
		FileName: filepath.Base(path),
		Type:     dt,
		Headings: map[int]bool{},
	}

	switch dt {
	case TypeBas:
		text, err := basic.DetokenizeToText(data)
		if err != nil {
			return nil, fmt.Errorf("falha ao detokenizar BASIC: %w", err)
		}
		doc.Lines = splitLines(text, tabSize)
		doc.Spans = buildBasicSpans(doc.Lines)

	case TypeMD:
		lines, headings := renderMarkdown(data, wrapWidth)
		doc.Lines = expandTabsAll(lines, tabSize)
		doc.Headings = headings

	default: // TypeText
		doc.Lines = splitLines(string(data), tabSize)
	}

	if len(doc.Lines) == 0 {
		doc.Lines = []string{""}
	}
	return doc, nil
}

// detectType decide o tipo a partir do conteúdo e da extensão.
func detectType(path string, data []byte) DocType {
	if basic.IsTokenized(data) {
		return TypeBas
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return TypeMD
	case ".bas":
		return TypeText // .bas em ASCII: exibe como texto
	default:
		return TypeText
	}
}

// splitLines divide o texto em linhas (normalizando CRLF) e expande tabs.
func splitLines(text string, tabSize int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	raw := strings.Split(text, "\n")
	return expandTabsAll(raw, tabSize)
}

func expandTabsAll(lines []string, tabSize int) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = expandTabs(l, tabSize)
	}
	return out
}

// expandTabs substitui tabs por espaços respeitando as colunas de tabulação.
func expandTabs(s string, tabSize int) string {
	if tabSize <= 0 {
		tabSize = 8
	}
	if !strings.ContainsRune(s, '\t') {
		return s
	}
	var b strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			n := tabSize - (col % tabSize)
			b.WriteString(strings.Repeat(" ", n))
			col += n
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String()
}
