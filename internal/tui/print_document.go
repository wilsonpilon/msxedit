package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// A4 page in PDF points (1 pt = 1/72 inch).
const (
	pdfA4W     = 595.276
	pdfA4H     = 841.890
	pdfMarginL = 50.0
	pdfMarginR = 50.0
	pdfMarginT = 50.0
	pdfMarginB = 50.0
)

// showPrint generates a PDF from the active editor using app.Printer settings
// and saves it in the current working directory.
func (a *App) showPrint() {
	if len(a.Editors) == 0 || a.ActiveEditor < 0 || a.ActiveEditor >= len(a.Editors) {
		if a.StatusBar != nil {
			a.StatusBar.SetText(" Nenhum editor ativo.")
		}
		return
	}
	ed := a.Editors[a.ActiveEditor]
	content := ed.editor.GetText()

	// Build output path: <name>.pdf in the current working directory.
	name := ed.fileName
	if name == "" || name == "Sem Nome" {
		name = "documento"
	}
	if ext := filepath.Ext(name); ext != "" {
		name = name[:len(name)-len(ext)]
	}
	outPath := name + ".pdf"

	cfg := a.Printer
	if cfg.WrapColumn <= 0 {
		cfg.WrapColumn = 80
	}
	if cfg.FontSize == 0 {
		cfg.FontSize = 10
	}

	if err := buildPDF(outPath, content, cfg); err != nil {
		if a.StatusBar != nil {
			a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao gerar PDF: %v[-]", err))
		}
		return
	}
	abs, _ := filepath.Abs(outPath)
	if a.StatusBar != nil {
		a.StatusBar.SetText(fmt.Sprintf(" PDF gerado: %s", abs))
	}
}

// buildPDF creates a minimal PDF/1.4 file with the listing.
func buildPDF(outPath, content string, cfg PrinterConfig) error {
	fs    := float64(cfg.FontSize) // font size in points
	charW := fs * 0.6             // Courier: 600 units per 1000 em → 0.6× size
	lineH := fs * 1.2             // standard leading (120% of font size)

	usableW := pdfA4W - pdfMarginL - pdfMarginR
	usableH := pdfA4H - pdfMarginT - pdfMarginB

	physW := int(usableW / charW) // physical columns that fit on the page
	lpp   := int(usableH / lineH) // physical lines per page

	// ── Prepare source lines ───────────────────────────────────────────────────
	raw := strings.TrimRight(content, "\r\n")
	rawLines := strings.Split(raw, "\n")
	for i, l := range rawLines {
		rawLines[i] = strings.TrimRight(l, "\r")
	}

	numPfxW := 0
	numFmt := ""
	if cfg.LineNumbers {
		numPfxW = len(fmt.Sprintf("%d", len(rawLines))) + 2 // e.g. "  1: " = 5
		numFmt = fmt.Sprintf("%%%dd: ", numPfxW-2)
	}

	// Content width respects WrapColumn (but always bounded by physical page width).
	contentW := physW - numPfxW
	if cfg.WrapColumn > 0 && cfg.WrapColumn < contentW {
		contentW = cfg.WrapColumn
	}
	if contentW < 10 {
		contentW = 10
	}

	// Build output lines with word wrap and optional line numbers.
	var outLines []string
	for i, line := range rawLines {
		pfx := ""
		if cfg.LineNumbers {
			pfx = fmt.Sprintf(numFmt, i+1)
		}
		runes := []rune(line)
		if len(runes) == 0 {
			outLines = append(outLines, pfx)
			continue
		}
		cont := strings.Repeat(" ", numPfxW)
		first := true
		for len(runes) > 0 {
			w := contentW
			if len(runes) < w {
				w = len(runes)
			}
			if first {
				outLines = append(outLines, pfx+string(runes[:w]))
				first = false
			} else {
				outLines = append(outLines, cont+string(runes[:w]))
			}
			runes = runes[w:]
		}
	}

	// ── Paginate ───────────────────────────────────────────────────────────────
	var pages [][]string
	for len(outLines) > 0 {
		end := lpp
		if end > len(outLines) {
			end = len(outLines)
		}
		pages = append(pages, outLines[:end])
		outLines = outLines[end:]
	}
	if len(pages) == 0 {
		pages = [][]string{{}} // at minimum one blank page
	}

	// ── Build PDF binary ───────────────────────────────────────────────────────
	//
	// Object layout:
	//   1        Catalog
	//   2        Pages (parent of all page dicts)
	//   3        Font  (Courier Type1, built-in – no embedding required)
	//   4+i*2   Page[i]
	//   5+i*2   Content stream[i]
	//
	N := len(pages)
	totalObjs := 3 + N*2
	offsets := make([]int, totalObjs)

	var buf bytes.Buffer

	put := func(s string) { buf.WriteString(s) }
	putf := func(f string, a ...interface{}) { fmt.Fprintf(&buf, f, a...) }

	beginObj := func(n int) {
		offsets[n-1] = buf.Len()
		putf("%d 0 obj\n", n)
	}
	endObj := func() { put("endobj\n") }

	// PDF header (binary comment hints to mail/ftp tools that this is binary)
	put("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")

	// 1: Catalog
	beginObj(1)
	put("<< /Type /Catalog /Pages 2 0 R >>\n")
	endObj()

	// 2: Pages
	beginObj(2)
	put("<< /Type /Pages /Kids [")
	for i := 0; i < N; i++ {
		putf("%d 0 R ", 4+i*2)
	}
	putf("] /Count %d >>\n", N)
	endObj()

	// 3: Font – Courier is a standard Type1 font; no file embedding needed.
	beginObj(3)
	put("<< /Type /Font /Subtype /Type1 /BaseFont /Courier /Encoding /WinAnsiEncoding >>\n")
	endObj()

	// Pages + content stream pairs
	startY := pdfA4H - pdfMarginT - fs // Y coordinate of first text baseline

	for i, pageLines := range pages {
		var cs bytes.Buffer
		fmt.Fprintf(&cs, "BT\n")
		fmt.Fprintf(&cs, "/F1 %.4f Tf\n", fs)
		fmt.Fprintf(&cs, "%.4f TL\n", lineH)                        // set text leading
		fmt.Fprintf(&cs, "%.4f %.4f Td\n", pdfMarginL, startY)      // move to first line
		for _, line := range pageLines {
			fmt.Fprintf(&cs, "(%s) Tj T*\n", pdfEscape(line))        // Tj + T* = show + next line
		}
		cs.WriteString("ET\n")
		stream := cs.Bytes()

		pageObj    := 4 + i*2
		contentObj := 5 + i*2

		// Page dictionary
		beginObj(pageObj)
		putf("<< /Type /Page /Parent 2 0 R\n"+
			"   /MediaBox [0 0 %.3f %.3f]\n"+
			"   /Contents %d 0 R\n"+
			"   /Resources << /Font << /F1 3 0 R >> >>\n"+
			">>\n",
			pdfA4W, pdfA4H, contentObj)
		endObj()

		// Content stream
		beginObj(contentObj)
		putf("<< /Length %d >>\n", len(stream))
		put("stream\n")
		buf.Write(stream)
		put("endstream\n")
		endObj()
	}

	// ── Cross-reference table ──────────────────────────────────────────────────
	// Each xref entry must be exactly 20 bytes: "nnnnnnnnnn ggggg x\r\n"
	xrefOff := buf.Len()
	putf("xref\n0 %d\n", totalObjs+1)
	put("0000000000 65535 f\r\n") // object 0: always free
	for _, off := range offsets {
		putf("%010d 00000 n\r\n", off)
	}

	// Trailer
	putf("trailer\n<< /Size %d /Root 1 0 R >>\n", totalObjs+1)
	putf("startxref\n%d\n", xrefOff)
	put("%%EOF\n")

	return os.WriteFile(outPath, buf.Bytes(), 0644)
}

// pdfEscape escapes a string for use in a PDF literal string (parenthesis-delimited).
// Handles printable ASCII and Latin-1 supplement (WinAnsiEncoding).
func pdfEscape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch {
		case r == '(':
			b.WriteString(`\(`)
		case r == ')':
			b.WriteString(`\)`)
		case r == '\\':
			b.WriteString(`\\`)
		case r >= 0x20 && r <= 0x7E: // printable ASCII
			b.WriteRune(r)
		case r >= 0xA0 && r <= 0xFF: // Latin-1 supplement (covered by WinAnsiEncoding)
			fmt.Fprintf(&b, "\\%03o", r)
		default:
			b.WriteByte('?')
		}
	}
	return b.String()
}
