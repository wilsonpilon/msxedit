package tui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ── Insert mode (Ctrl+V / Ins) ───────────────────────────────────────────────

// cmdToggleOverwrite alterna entre modo Insert (padrão) e Overwrite: com
// Overwrite ativo, digitar um caractere substitui o que está sob o cursor em
// vez de deslocá-lo para a direita.
func (w *editorWindow) cmdToggleOverwrite() {
	w.overwriteMode = !w.overwriteMode
	if w.app != nil && w.app.StatusBar != nil {
		if w.overwriteMode {
			w.app.StatusBar.SetText(" Modo: Overwrite")
		} else {
			w.app.StatusBar.SetText(" Modo: Insert")
		}
	}
}

// applyOverwrite, quando o modo Overwrite está ativo, remove o caractere sob
// o cursor antes de uma tecla imprimível ser inserida pelo TextArea — dando o
// efeito de substituição em vez de inserção. Não afeta Enter/Tab.
func (w *editorWindow) applyOverwrite(event *tcell.EventKey) {
	if !w.overwriteMode || event.Key() != tcell.KeyRune {
		return
	}
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	row, col, _, _ := w.editor.GetCursor()
	if row < 0 || row >= len(lines) {
		return
	}
	lineRunes := []rune(lines[row])
	if col >= len(lineRunes) {
		return // fim da linha: apenas insere, como o Insert mode
	}
	off := w.cursorByteOffset()
	next := string(lineRunes[col])
	w.editor.Replace(off, off+len(next), "")
}

// ── Insert/Delete line (Ctrl+N / Ctrl+Y) ─────────────────────────────────────

// cmdInsertLine insere uma linha em branco na posição da linha atual,
// deixando o cursor no início da nova linha (o texto original desce uma linha).
func (w *editorWindow) cmdInsertLine() {
	row, _, _, _ := w.editor.GetCursor()
	text := w.editor.GetText()
	off := rowColToByteOffset(text, row, 0)
	w.editor.Replace(off, off, "\n")
	newText := w.editor.GetText()
	newOff := rowColToByteOffset(newText, row, 0)
	w.editor.Select(newOff, newOff)
}

// cmdDeleteLine remove a linha inteira em que o cursor está (conteúdo e a
// quebra de linha correspondente), deixando o cursor na linha que ocupar a
// mesma posição em seguida.
func (w *editorWindow) cmdDeleteLine() {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	row, _, _, _ := w.editor.GetCursor()
	if row < 0 || row >= len(lines) {
		return
	}

	start := rowColToByteOffset(text, row, 0)
	var end int
	if row+1 < len(lines) {
		end = rowColToByteOffset(text, row+1, 0)
	} else {
		end = len(text)
		if start > 0 {
			start-- // última linha: remove também o '\n' anterior
		}
	}
	w.editor.Replace(start, end, "")

	newText := w.editor.GetText()
	newLines := strings.Split(newText, "\n")
	newRow := row
	if newRow >= len(newLines) {
		newRow = len(newLines) - 1
	}
	if newRow < 0 {
		newRow = 0
	}
	off := rowColToByteOffset(newText, newRow, 0)
	w.editor.Select(off, off)
}

// ── Delete to end of line (Ctrl+Q Y) ─────────────────────────────────────────

// cmdDeleteToEndOfLine remove do cursor até o final da linha atual.
func (w *editorWindow) cmdDeleteToEndOfLine() {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	row, col, _, _ := w.editor.GetCursor()
	if row < 0 || row >= len(lines) {
		return
	}
	lineLen := len([]rune(lines[row]))
	if col >= lineLen {
		return // já está no fim da linha
	}
	start := w.cursorByteOffset()
	end := rowColToByteOffset(text, row, lineLen)
	w.editor.Replace(start, end, "")
}

// ── Delete character (Ctrl+G) ────────────────────────────────────────────────

// cmdDeleteCharRight apaga o caractere à direita do cursor (equivalente a Del).
func (w *editorWindow) cmdDeleteCharRight(setFocus func(tview.Primitive)) {
	if handler := w.editor.InputHandler(); handler != nil {
		handler(tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone), setFocus)
	}
}

// ── Delete word right (Ctrl+T) ───────────────────────────────────────────────

// isWordChar usa a mesma definição de "caractere de palavra" do Ctrl+K T
// (mark word), incluindo os sufixos de tipo do MSX-BASIC ($ % ! #).
func isWordChar(r rune) bool {
	return r == '_' || r == '$' || r == '%' || r == '!' || r == '#' ||
		unicode.IsLetter(r) || unicode.IsDigit(r)
}

// cmdDeleteWordRight apaga a palavra à direita do cursor, incluindo os
// espaços/pontuação até o início da próxima palavra (ou fim da linha).
func (w *editorWindow) cmdDeleteWordRight() {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	row, col, _, _ := w.editor.GetCursor()
	if row < 0 || row >= len(lines) {
		return
	}
	lineRunes := []rune(lines[row])

	end := col
	for end < len(lineRunes) && isWordChar(lineRunes[end]) {
		end++
	}
	for end < len(lineRunes) && !isWordChar(lineRunes[end]) {
		end++
	}
	if end == col {
		return
	}

	start := w.cursorByteOffset()
	finish := rowColToByteOffset(text, row, end)
	w.editor.Replace(start, finish, "")
}
