package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"msxedit/internal/basic"
)

// ── Utilidades de posição ───────────────────────────────────────────────────

// rowColToByteOffset converte posição (row, col) em rune-index para byte offset no texto.
func rowColToByteOffset(text string, row, col int) int {
	lines := strings.Split(text, "\n")
	off := 0
	for i := 0; i < row && i < len(lines); i++ {
		off += len(lines[i]) + 1 // +1 pelo '\n'
	}
	if row < len(lines) {
		runes := []rune(lines[row])
		if col < 0 {
			col = 0
		}
		if col > len(runes) {
			col = len(runes)
		}
		off += len(string(runes[:col]))
	}
	return off
}

// byteOffsetToRowCol converte byte offset para (row, col) em rune-index.
func byteOffsetToRowCol(text string, offset int) (row, col int) {
	for i, r := range text {
		if i >= offset {
			break
		}
		if r == '\n' {
			row++
			col = 0
		} else {
			col++
		}
	}
	return
}

// ── Helpers de estado do bloco ──────────────────────────────────────────────

func (w *editorWindow) clearBlock() {
	w.blkBeginRow = -1
	w.blkBeginCol = 0
	w.blkEndRow = -1
	w.blkEndCol = 0
	w.blkVisible = false
}

func (w *editorWindow) blockValid() bool {
	if w.blkBeginRow < 0 || w.blkEndRow < 0 {
		return false
	}
	text := w.editor.GetText()
	s := rowColToByteOffset(text, w.blkBeginRow, w.blkBeginCol)
	e := rowColToByteOffset(text, w.blkEndRow, w.blkEndCol)
	return s < e
}

// blockByteRange retorna (start, end) em byte offsets, normalizado start <= end.
func (w *editorWindow) blockByteRange() (int, int) {
	text := w.editor.GetText()
	s := rowColToByteOffset(text, w.blkBeginRow, w.blkBeginCol)
	e := rowColToByteOffset(text, w.blkEndRow, w.blkEndCol)
	if s > e {
		s, e = e, s
	}
	return s, e
}

func (w *editorWindow) normalizeBlock() {
	if w.blkBeginRow < 0 || w.blkEndRow < 0 {
		return
	}
	text := w.editor.GetText()
	bs := rowColToByteOffset(text, w.blkBeginRow, w.blkBeginCol)
	es := rowColToByteOffset(text, w.blkEndRow, w.blkEndCol)
	if bs > es {
		w.blkBeginRow, w.blkBeginCol, w.blkEndRow, w.blkEndCol =
			w.blkEndRow, w.blkEndCol, w.blkBeginRow, w.blkBeginCol
	}
}

// cursorByteOffset retorna o byte offset do cursor no texto.
func (w *editorWindow) cursorByteOffset() int {
	row, col, _, _ := w.editor.GetCursor()
	return rowColToByteOffset(w.editor.GetText(), row, col)
}

// inBlock informa se a posição (textRow, textCol) está dentro do bloco marcado.
// lineLen é o número de runes na linha textRow.
func (w *editorWindow) inBlock(textRow, textCol, lineLen int) bool {
	br, bc := w.blkBeginRow, w.blkBeginCol
	er, ec := w.blkEndRow, w.blkEndCol
	if textRow < br || textRow > er {
		return false
	}
	if br == er {
		return textCol >= bc && textCol < ec
	}
	if textRow == br {
		return textCol >= bc
	}
	if textRow == er {
		return textCol < ec
	}
	return true // linha intermediária: toda ela
}

// ── Highlight do bloco ──────────────────────────────────────────────────────

func (w *editorWindow) applyBlockHighlight(screen tcell.Screen, innerX, innerY, innerWidth, innerHeight int) {
	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	offsetRow, offsetCol := w.editor.GetOffset()
	cursorRow, cursorCol, _, _ := w.editor.GetCursor()
	maxX, maxY := screen.Size()

	blkStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)

	for visRow := 0; visRow < innerHeight; visRow++ {
		textRow := offsetRow + visRow
		if textRow >= len(lines) {
			break
		}
		if textRow < w.blkBeginRow || textRow > w.blkEndRow {
			continue
		}

		lineRunes := []rune(lines[textRow])
		lineLen := len(lineRunes)

		for visCol := 0; visCol < innerWidth; visCol++ {
			textCol := visCol + offsetCol

			if textCol >= lineLen {
				break // sem mais caracteres nesta linha
			}
			if !w.inBlock(textRow, textCol, lineLen) {
				continue
			}
			if textRow == cursorRow && textCol == cursorCol {
				continue // preserva renderização do cursor
			}
			putRune(screen, maxX, maxY, innerX+visCol, innerY+visRow, lineRunes[textCol], blkStyle)
		}
	}
}

// ── Dispatcher principal de teclas de bloco ─────────────────────────────────

func (w *editorWindow) handleBlockKey(event *tcell.EventKey, setFocus func(tview.Primitive)) bool {
	key := event.Key()
	mod := event.Modifiers()

	// Esc cancela o prefixo Ctrl+K / Ctrl+Q / Ctrl+O / Ctrl+P pendente
	if (w.waitingK || w.waitingQ || w.waitingO || w.waitingP) && key == tcell.KeyEscape {
		w.waitingK = false
		w.waitingQ = false
		w.waitingO = false
		w.waitingP = false
		return true
	}

	// ── Sequência Ctrl+K ────────────────────────────────────────────────────
	if w.waitingK {
		w.waitingK = false
		if key == tcell.KeyRune {
			r := unicode.ToLower(event.Rune())
			if r >= '0' && r <= '9' {
				w.cmdSetMarker(int(r - '0'))
				return true
			}
			switch r {
			case 'b':
				w.cmdMarkBegin()
				return true
			case 'k':
				w.cmdMarkEnd()
				return true
			case 't':
				w.cmdMarkWord()
				return true
			case 'c':
				w.cmdCopyBlock()
				return true
			case 'v':
				w.cmdMoveBlock()
				return true
			case 'y':
				w.cmdDeleteBlock()
				return true
			case 'r':
				w.cmdReadBlock()
				return true
			case 'w':
				w.cmdWriteBlock()
				return true
			case 'h':
				w.cmdHideBlock()
				return true
			case 'p':
				w.cmdPrintBlock()
				return true
			case 'i':
				w.cmdIndentBlock()
				return true
			case 'u':
				w.cmdUnindentBlock()
				return true
			case 'd':
				w.cmdExitToMenu()
				return true
			case 'l':
				w.cmdMarkLine()
				return true
			case 's':
				if w.app != nil {
					w.app.showSave()
				}
				return true
			}
		}
		return false // segunda tecla desconhecida: passa ao editor
	}

	// ── Sequência Ctrl+Q ────────────────────────────────────────────────────
	if w.waitingQ {
		w.waitingQ = false
		if key == tcell.KeyRune {
			r := unicode.ToLower(event.Rune())
			if r >= '0' && r <= '9' {
				w.cmdGotoMarker(int(r - '0'))
				return true
			}
			switch r {
			case 'b':
				w.cmdGotoBegin()
				return true
			case 'k':
				w.cmdGotoEnd()
				return true
			case 'y':
				w.cmdDeleteToEndOfLine()
				return true
			case 'l':
				w.cmdRestoreLine()
				return true
			case 'f':
				if w.app != nil {
					w.app.showFindDialog()
				}
				return true
			case 'a':
				if w.app != nil {
					w.app.showReplaceDialog()
				}
				return true
			}
		}
		return false
	}

	// ── Sequência Ctrl+O ────────────────────────────────────────────────────
	if w.waitingO {
		w.waitingO = false
		if key == tcell.KeyRune {
			switch unicode.ToLower(event.Rune()) {
			case 't':
				w.cmdToggleTabMode()
				return true
			case 'i':
				w.cmdToggleAutoIndent()
				return true
			}
		}
		return false
	}

	// ── Sequência Ctrl+P (insere o código de controle da próxima tecla) ──────
	if w.waitingP {
		w.waitingP = false
		w.cmdInsertLiteralControl(event)
		return true
	}

	// ── Prefixos ─────────────────────────────────────────────────────────────
	if key == tcell.KeyCtrlK {
		w.waitingK = true
		if w.app != nil && w.app.StatusBar != nil {
			w.app.StatusBar.SetText("^K B:Begin K:End T:Word C:Copy V:Move Y:Del R:Read W:Write H:Hide P:Print I:Indent U:Unind D:Menu L:Line S:Save 0-9:Marker")
		}
		return true
	}
	if key == tcell.KeyCtrlQ {
		w.waitingQ = true
		if w.app != nil && w.app.StatusBar != nil {
			w.app.StatusBar.SetText("^Q B:GoBegin K:GoEnd Y:DelToEOL L:Restore F:Find A:Replace 0-9:GoMarker")
		}
		return true
	}
	if key == tcell.KeyCtrlO {
		w.waitingO = true
		if w.app != nil && w.app.StatusBar != nil {
			w.app.StatusBar.SetText("^O T:TabMode I:AutoIndent")
		}
		return true
	}
	if key == tcell.KeyCtrlP {
		w.waitingP = true
		if w.app != nil && w.app.StatusBar != nil {
			w.app.StatusBar.SetText("^P: pressione Ctrl+letra para inserir o codigo de controle")
		}
		return true
	}

	// ── Atalhos de clipboard ─────────────────────────────────────────────────
	if key == tcell.KeyInsert {
		if mod&tcell.ModCtrl != 0 {
			w.cmdCopyToClipboard()
			return true
		}
		if mod&tcell.ModShift != 0 {
			w.cmdPasteFromClipboard()
			return true
		}
		// Insert simples (sem modificador): alterna modo Insert/Overwrite.
		w.cmdToggleOverwrite()
		return true
	}
	if key == tcell.KeyDelete {
		if mod&tcell.ModShift != 0 {
			w.cmdCutToClipboard()
			return true
		}
		if mod&tcell.ModCtrl != 0 {
			w.cmdDeleteBlock()
			return true
		}
	}

	// ── Atalho de undo (Alt+BkSp) ────────────────────────────────────────────
	if (key == tcell.KeyBackspace || key == tcell.KeyBackspace2) && mod&tcell.ModAlt != 0 {
		w.cmdUndo()
		return true
	}

	// ── Insert & Delete commands (Help > Editor Commands) ───────────────────
	switch key {
	case tcell.KeyCtrlV: // Insert mode on/off
		w.cmdToggleOverwrite()
		return true
	case tcell.KeyCtrlN: // Insert line
		w.cmdInsertLine()
		return true
	case tcell.KeyCtrlY: // Delete line
		w.cmdDeleteLine()
		return true
	case tcell.KeyCtrlG: // Delete character (equivalente a Del)
		w.cmdDeleteCharRight(setFocus)
		return true
	case tcell.KeyCtrlT: // Delete word right
		w.cmdDeleteWordRight()
		return true
	}

	return false
}

// restoreStatusBar restaura a barra de status padrão após um comando de bloco.
func (w *editorWindow) restoreStatusBar() {
	if w.app == nil || w.app.StatusBar == nil {
		return
	}
	w.app.StatusBar.SetText(" [red]F1[-] Help  [red]F2[-] Save  [red]F3[-] Open  [red]Alt+[F9][-] Compile  [red]F9[-] Make  [red]Alt+[F10][-] Local menu")
}

// ── Comandos de bloco ───────────────────────────────────────────────────────

func (w *editorWindow) cmdMarkBegin() {
	row, col, _, _ := w.editor.GetCursor()
	w.blkBeginRow, w.blkBeginCol = row, col
	if w.blkEndRow >= 0 {
		w.normalizeBlock()
		w.blkVisible = true
	}
	w.restoreStatusBar()
}

func (w *editorWindow) cmdMarkEnd() {
	row, col, _, _ := w.editor.GetCursor()
	w.blkEndRow, w.blkEndCol = row, col
	if w.blkBeginRow >= 0 {
		w.normalizeBlock()
		w.blkVisible = true
	}
	w.restoreStatusBar()
}

func (w *editorWindow) cmdMarkWord() {
	text := w.editor.GetText()
	row, col, _, _ := w.editor.GetCursor()
	lines := strings.Split(text, "\n")
	if row >= len(lines) {
		return
	}
	lineRunes := []rune(lines[row])
	if col >= len(lineRunes) {
		return
	}
	isWordChar := func(r rune) bool {
		return r == '_' || r == '$' || r == '%' || r == '!' || r == '#' ||
			unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	if !isWordChar(lineRunes[col]) {
		return
	}
	start := col
	for start > 0 && isWordChar(lineRunes[start-1]) {
		start--
	}
	end := col + 1
	for end < len(lineRunes) && isWordChar(lineRunes[end]) {
		end++
	}
	w.blkBeginRow, w.blkBeginCol = row, start
	w.blkEndRow, w.blkEndCol = row, end
	w.blkVisible = true
	w.restoreStatusBar()
}

func (w *editorWindow) cmdMarkLine() {
	text := w.editor.GetText()
	row, _, _, _ := w.editor.GetCursor()
	lines := strings.Split(text, "\n")
	lineLen := 0
	if row < len(lines) {
		lineLen = len([]rune(lines[row]))
	}
	w.blkBeginRow, w.blkBeginCol = row, 0
	w.blkEndRow, w.blkEndCol = row, lineLen
	w.blkVisible = true
	w.restoreStatusBar()
}

func (w *editorWindow) cmdCopyBlock() {
	if !w.blockValid() {
		w.restoreStatusBar()
		return
	}
	text := w.editor.GetText()
	bStart, bEnd := w.blockByteRange()
	blockText := text[bStart:bEnd]

	curOff := w.cursorByteOffset()
	if curOff > bStart && curOff < bEnd {
		w.restoreStatusBar()
		return // cursor dentro do bloco
	}

	w.editor.Replace(curOff, curOff, blockText)
	newText := w.editor.GetText()
	pasteEnd := curOff + len(blockText)
	w.blkBeginRow, w.blkBeginCol = byteOffsetToRowCol(newText, curOff)
	w.blkEndRow, w.blkEndCol = byteOffsetToRowCol(newText, pasteEnd)
	w.blkVisible = true
	w.restoreStatusBar()
}

func (w *editorWindow) cmdMoveBlock() {
	if !w.blockValid() {
		w.restoreStatusBar()
		return
	}
	text := w.editor.GetText()
	bStart, bEnd := w.blockByteRange()
	blockText := text[bStart:bEnd]

	curOff := w.cursorByteOffset()
	if curOff >= bStart && curOff <= bEnd {
		w.restoreStatusBar()
		return // cursor dentro ou na borda do bloco
	}

	var newText string
	var pasteStart int

	if curOff < bStart {
		newText = text[:curOff] + blockText + text[curOff:bStart] + text[bEnd:]
		pasteStart = curOff
	} else {
		newText = text[:bStart] + text[bEnd:curOff] + blockText + text[curOff:]
		pasteStart = bStart + (curOff - bEnd)
	}
	pasteEnd := pasteStart + len(blockText)

	w.editor.SetText(newText, false)
	w.editor.Select(pasteEnd, pasteEnd)

	newText2 := w.editor.GetText()
	w.blkBeginRow, w.blkBeginCol = byteOffsetToRowCol(newText2, pasteStart)
	w.blkEndRow, w.blkEndCol = byteOffsetToRowCol(newText2, pasteEnd)
	w.blkVisible = true
	w.restoreStatusBar()
}

func (w *editorWindow) cmdDeleteBlock() {
	if !w.blockValid() {
		w.restoreStatusBar()
		return
	}
	bStart, bEnd := w.blockByteRange()
	w.editor.Replace(bStart, bEnd, "")
	w.clearBlock()
	w.restoreStatusBar()
}

func (w *editorWindow) cmdReadBlock() {
	if w.app == nil {
		return
	}
	w.restoreStatusBar()
	showOpenFileDialog(w.app, "*.*",
		func(path string) {
			data, err := os.ReadFile(path)
			if err != nil {
				return
			}
			content := string(data)
			if basic.IsTokenized(data) {
				if t, err2 := basic.DetokenizeToText(data); err2 == nil {
					content = t
				}
			}
			curOff := w.cursorByteOffset()
			w.editor.Replace(curOff, curOff, content)
			newText := w.editor.GetText()
			pasteEnd := curOff + len(content)
			w.blkBeginRow, w.blkBeginCol = byteOffsetToRowCol(newText, curOff)
			w.blkEndRow, w.blkEndCol = byteOffsetToRowCol(newText, pasteEnd)
			w.blkVisible = true
		},
		func(path string) {
			// double-click: mesmo comportamento
		},
	)
}

func (w *editorWindow) cmdWriteBlock() {
	if !w.blockValid() || w.app == nil {
		w.restoreStatusBar()
		return
	}
	text := w.editor.GetText()
	bStart, bEnd := w.blockByteRange()
	blockText := text[bStart:bEnd]

	w.restoreStatusBar()
	dlg := newSaveFileDialog(w.app, "")
	dlg.onSave = func(path string, tokenized bool) {
		data := []byte(blockText)
		if tokenized {
			if tok, err := basic.Tokenize(blockText); err == nil {
				data = tok
			}
		}
		os.WriteFile(path, data, 0644) //nolint:errcheck
	}
	dlg.onClose = func() {
		w.app.Application.SetFocus(w)
	}
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dlg, saveFileDlgW, 0, true).
			AddItem(nil, 0, 1, false), saveFileDlgH, 0, true).
		AddItem(nil, 0, 1, false)
	w.app.Pages.AddPage("save_file", container, true, true)
	w.app.Application.SetFocus(dlg)
}

func (w *editorWindow) cmdPrintBlock() {
	if !w.blockValid() || w.app == nil {
		w.restoreStatusBar()
		return
	}
	text := w.editor.GetText()
	bStart, bEnd := w.blockByteRange()
	blockText := text[bStart:bEnd]

	name := w.fileName
	if name == "" || name == "Sem Nome" {
		name = "bloco"
	}
	if ext := filepath.Ext(name); ext != "" {
		name = name[:len(name)-len(ext)]
	}
	outPath := name + "_block.pdf"

	cfg := w.app.Printer
	if cfg.WrapColumn <= 0 {
		cfg.WrapColumn = 80
	}
	if cfg.FontSize == 0 {
		cfg.FontSize = 10
	}
	if err := buildPDF(outPath, blockText, cfg); err != nil {
		w.restoreStatusBar()
		return
	}
	abs, _ := filepath.Abs(outPath)
	if w.app.StatusBar != nil {
		w.app.StatusBar.SetText(fmt.Sprintf(" PDF do bloco: %s", abs))
	}
}

func (w *editorWindow) cmdHideBlock() {
	w.blkVisible = !w.blkVisible
	w.restoreStatusBar()
}

func (w *editorWindow) cmdIndentBlock() {
	if !w.blockValid() {
		w.restoreStatusBar()
		return
	}
	const indentStr = "  " // 2 espaços
	const indentLen = 2

	text := w.editor.GetText()
	lines := strings.Split(text, "\n")

	for i := w.blkBeginRow; i <= w.blkEndRow && i < len(lines); i++ {
		lines[i] = indentStr + lines[i]
	}
	newText := strings.Join(lines, "\n")

	w.blkBeginCol += indentLen
	w.blkEndCol += indentLen

	w.editor.SetText(newText, false)
	off := rowColToByteOffset(newText, w.blkBeginRow, w.blkBeginCol)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

func (w *editorWindow) cmdUnindentBlock() {
	if !w.blockValid() {
		w.restoreStatusBar()
		return
	}
	const indentLen = 2

	text := w.editor.GetText()
	lines := strings.Split(text, "\n")

	removedBegin, removedEnd := 0, 0
	for i := w.blkBeginRow; i <= w.blkEndRow && i < len(lines); i++ {
		runes := []rune(lines[i])
		removed := 0
		for removed < indentLen && removed < len(runes) && runes[removed] == ' ' {
			removed++
		}
		lines[i] = string(runes[removed:])
		if i == w.blkBeginRow {
			removedBegin = removed
		}
		if i == w.blkEndRow {
			removedEnd = removed
		}
	}
	newText := strings.Join(lines, "\n")

	w.blkBeginCol -= removedBegin
	if w.blkBeginCol < 0 {
		w.blkBeginCol = 0
	}
	w.blkEndCol -= removedEnd
	if w.blkEndCol < 0 {
		w.blkEndCol = 0
	}

	w.editor.SetText(newText, false)
	off := rowColToByteOffset(newText, w.blkBeginRow, w.blkBeginCol)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

func (w *editorWindow) cmdGotoBegin() {
	if w.blkBeginRow < 0 {
		return
	}
	text := w.editor.GetText()
	off := rowColToByteOffset(text, w.blkBeginRow, w.blkBeginCol)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

func (w *editorWindow) cmdGotoEnd() {
	if w.blkEndRow < 0 {
		return
	}
	text := w.editor.GetText()
	off := rowColToByteOffset(text, w.blkEndRow, w.blkEndCol)
	w.editor.Select(off, off)
	w.restoreStatusBar()
}

func (w *editorWindow) cmdExitToMenu() {
	w.restoreStatusBar()
	if w.onExitToMenu != nil {
		w.onExitToMenu()
	}
}

// ── Clipboard ────────────────────────────────────────────────────────────────

// setClipboard atualiza o clipboard compartilhado da aplicação (visível para
// todas as janelas de edição e refletido na janela "Show clipboard").
// Sem app associado (ex.: uso isolado/testes), cai para o clipboard local.
func (w *editorWindow) setClipboard(text string) {
	w.blkClip = text
	if w.app != nil {
		w.app.Clipboard = text
		w.app.syncClipboardWindow()
	}
}

// getClipboard lê o clipboard compartilhado da aplicação, com fallback local.
func (w *editorWindow) getClipboard() string {
	if w.app != nil {
		return w.app.Clipboard
	}
	return w.blkClip
}

func (w *editorWindow) cmdCopyToClipboard() {
	if !w.blockValid() {
		return
	}
	text := w.editor.GetText()
	s, e := w.blockByteRange()
	w.setClipboard(text[s:e])
	w.restoreStatusBar()
}

func (w *editorWindow) cmdCutToClipboard() {
	w.cmdCopyToClipboard()
	if w.getClipboard() != "" {
		w.cmdDeleteBlock()
	}
	w.restoreStatusBar()
}

func (w *editorWindow) cmdPasteFromClipboard() {
	clip := w.getClipboard()
	if clip == "" {
		return
	}
	curOff := w.cursorByteOffset()
	w.editor.Replace(curOff, curOff, clip)
	newText := w.editor.GetText()
	pasteEnd := curOff + len(clip)
	w.blkBeginRow, w.blkBeginCol = byteOffsetToRowCol(newText, curOff)
	w.blkEndRow, w.blkEndCol = byteOffsetToRowCol(newText, pasteEnd)
	w.blkVisible = true
	w.restoreStatusBar()
}

// ── Undo/Redo ────────────────────────────────────────────────────────────────
// O tview.TextArea já implementa undo/redo internamente (Ctrl+Z / Ctrl+Y);
// aqui apenas disparamos esses mesmos eventos a partir de outros atalhos/menu.

func (w *editorWindow) cmdUndo() {
	if handler := w.editor.InputHandler(); handler != nil {
		handler(tcell.NewEventKey(tcell.KeyCtrlZ, 0, tcell.ModCtrl), func(tview.Primitive) {})
	}
	w.restoreStatusBar()
}

func (w *editorWindow) cmdRedo() {
	if handler := w.editor.InputHandler(); handler != nil {
		handler(tcell.NewEventKey(tcell.KeyCtrlY, 0, tcell.ModCtrl), func(tview.Primitive) {})
	}
	w.restoreStatusBar()
}
