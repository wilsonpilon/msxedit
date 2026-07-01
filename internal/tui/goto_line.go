package tui

import (
	"strconv"
	"strings"
)

// gotoTextLine move o cursor para o início da linha N (1-based) do texto.
// Retorna false se o editor estiver vazio.
func (w *editorWindow) gotoTextLine(lineNum int) bool {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return false
	}
	row := lineNum - 1
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	off := rowColToByteOffset(text, row, 0)
	w.editor.Select(off, off)
	return true
}

// gotoBasicLine procura a linha de código cujo número MSX-Basic (os dígitos
// no início da linha, ex.: "10 PRINT...") seja igual a basicLineNum e move o
// cursor até ela. Retorna false se nenhuma linha com esse número existir.
func (w *editorWindow) gotoBasicLine(basicLineNum int) bool {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		j := 0
		for j < len(trimmed) && trimmed[j] >= '0' && trimmed[j] <= '9' {
			j++
		}
		if j == 0 {
			continue
		}
		n, err := strconv.Atoi(trimmed[:j])
		if err != nil {
			continue
		}
		if n == basicLineNum {
			off := rowColToByteOffset(text, i, 0)
			w.editor.Select(off, off)
			return true
		}
	}
	return false
}
