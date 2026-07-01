package tui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// resetPerFileState limpa o estado de "restore line" e os marcadores de
// posição sempre que o conteúdo do editor é substituído por completo (ex.:
// abrir um novo arquivo) — caso contrário eles continuariam apontando para
// posições do texto anterior.
func (w *editorWindow) resetPerFileState() {
	w.restoreLineRow = -1
	w.restoreLineText = ""
	for i := range w.placeMarkers {
		w.placeMarkers[i].row = -1
	}
}

// ── Restore line (Ctrl+Q L) ──────────────────────────────────────────────────

// snapshotRestoreLine memoriza o conteúdo da linha atual assim que o cursor
// chega nela, permitindo que Ctrl+Q L desfaça todas as edições feitas nela
// desde então (mesmo que Undo já tenha sido usado para outra coisa depois).
func (w *editorWindow) snapshotRestoreLine() {
	row, _, _, _ := w.editor.GetCursor()
	if row == w.restoreLineRow {
		return
	}
	lines := strings.Split(w.editor.GetText(), "\n")
	if row < 0 || row >= len(lines) {
		return
	}
	w.restoreLineRow = row
	w.restoreLineText = lines[row]
}

// cmdRestoreLine restaura a linha atual ao conteúdo salvo por snapshotRestoreLine.
func (w *editorWindow) cmdRestoreLine() {
	if w.restoreLineRow < 0 {
		w.restoreStatusBar()
		return
	}
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	row := w.restoreLineRow
	if row >= len(lines) {
		w.restoreStatusBar()
		return
	}

	_, col, _, _ := w.editor.GetCursor()
	start := rowColToByteOffset(text, row, 0)
	end := rowColToByteOffset(text, row, len([]rune(lines[row])))
	w.editor.Replace(start, end, w.restoreLineText)

	newText := w.editor.GetText()
	restoredLen := len([]rune(w.restoreLineText))
	if col > restoredLen {
		col = restoredLen
	}
	off := rowColToByteOffset(newText, row, col)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

// ── Place markers (Ctrl+K 0-9 / Ctrl+Q 0-9) ──────────────────────────────────

// placeMarker é uma posição memorizada no texto (linha/coluna).
type placeMarker struct {
	row, col int
}

// cmdSetMarker grava a posição atual do cursor no marcador n (0-9).
func (w *editorWindow) cmdSetMarker(n int) {
	if n < 0 || n > 9 {
		return
	}
	row, col, _, _ := w.editor.GetCursor()
	w.placeMarkers[n] = placeMarker{row: row, col: col}
	w.restoreStatusBar()
}

// cmdGotoMarker move o cursor para o marcador n (0-9), se definido.
func (w *editorWindow) cmdGotoMarker(n int) {
	if n < 0 || n > 9 || w.placeMarkers[n].row < 0 {
		w.restoreStatusBar()
		return
	}
	m := w.placeMarkers[n]
	off := rowColToByteOffset(w.editor.GetText(), m.row, m.col)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

// ── Tab mode (Ctrl+O T) ───────────────────────────────────────────────────────

const tabStopSize = 8

// cmdToggleTabMode alterna entre inserir um caractere de tabulação (padrão) ou
// espaços até a próxima parada de tabulação.
func (w *editorWindow) cmdToggleTabMode() {
	w.tabInsertsSpaces = !w.tabInsertsSpaces
	if w.app != nil && w.app.StatusBar != nil {
		if w.tabInsertsSpaces {
			w.app.StatusBar.SetText(" Tab mode: Spaces")
		} else {
			w.app.StatusBar.SetText(" Tab mode: Character")
		}
	}
}

// applyTabMode, quando tabInsertsSpaces está ativo, substitui o Tab por
// espaços até a próxima parada de tabulação. Retorna true se tratou o evento.
func (w *editorWindow) applyTabMode(event *tcell.EventKey) bool {
	if !w.tabInsertsSpaces || event.Key() != tcell.KeyTab {
		return false
	}
	_, col, _, _ := w.editor.GetCursor()
	spaces := tabStopSize - (col % tabStopSize)
	off := w.cursorByteOffset()
	w.editor.Replace(off, off, strings.Repeat(" ", spaces))
	return true
}

// ── Auto indent (Ctrl+O I) ───────────────────────────────────────────────────

// cmdToggleAutoIndent alterna a indentação automática ao pressionar Enter.
func (w *editorWindow) cmdToggleAutoIndent() {
	w.autoIndent = !w.autoIndent
	if w.app != nil && w.app.StatusBar != nil {
		if w.autoIndent {
			w.app.StatusBar.SetText(" Auto indent: On")
		} else {
			w.app.StatusBar.SetText(" Auto indent: Off")
		}
	}
}

// applyAutoIndent, quando autoIndent está ativo, insere no início da nova
// linha (após Enter) o mesmo espaço em branco inicial da linha anterior.
// Retorna true se tratou o evento (delegando a quebra de linha em si ao
// TextArea antes de aplicar a indentação).
func (w *editorWindow) applyAutoIndent(event *tcell.EventKey, setFocus func(tview.Primitive)) bool {
	if !w.autoIndent || event.Key() != tcell.KeyEnter {
		return false
	}
	lines := strings.Split(w.editor.GetText(), "\n")
	row, _, _, _ := w.editor.GetCursor()
	indent := ""
	if row >= 0 && row < len(lines) {
		line := lines[row]
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		indent = line[:i]
	}

	if handler := w.editor.InputHandler(); handler != nil {
		handler(event, setFocus)
	}
	if indent != "" {
		off := w.cursorByteOffset()
		w.editor.Replace(off, off, indent)
	}
	return true
}

// ── Ctrl+character prefix (Ctrl+P) ──────────────────────────────────────────

// ctrlCodeForKey calcula o código de controle ASCII (1-26) correspondente a
// uma tecla Ctrl+letra (ou a uma letra simples), para uso por Ctrl+P.
func ctrlCodeForKey(event *tcell.EventKey) (rune, bool) {
	key := event.Key()
	if key >= tcell.KeyCtrlA && key <= tcell.KeyCtrlZ {
		return rune(key-tcell.KeyCtrlA) + 1, true
	}
	if key == tcell.KeyRune {
		r := unicode.ToUpper(event.Rune())
		if r >= 'A' && r <= 'Z' {
			return r - 'A' + 1, true
		}
	}
	return 0, false
}

// cmdInsertLiteralControl insere o byte de controle correspondente à próxima
// tecla pressionada após Ctrl+P (o "Ctrl+character prefix" do WordStar/Turbo).
func (w *editorWindow) cmdInsertLiteralControl(event *tcell.EventKey) {
	code, ok := ctrlCodeForKey(event)
	if !ok {
		w.restoreStatusBar()
		return
	}
	off := w.cursorByteOffset()
	w.editor.Replace(off, off, string(code))
	w.restoreStatusBar()
}
