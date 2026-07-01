package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// isNavigationKey informa se a tecla é uma tecla de movimentação de cursor
// que participa da seleção de texto com Shift.
func isNavigationKey(key tcell.Key) bool {
	switch key {
	case tcell.KeyLeft, tcell.KeyRight, tcell.KeyUp, tcell.KeyDown,
		tcell.KeyHome, tcell.KeyEnd, tcell.KeyPgUp, tcell.KeyPgDn:
		return true
	}
	return false
}

// translateWordStarKey traduz os atalhos estilo WordStar/Turbo (Ctrl+letra,
// ver Help > Editor Commands > Cursor-movement commands) para a tecla de
// seta/página equivalente, permitindo reaproveitar toda a lógica de
// navegação/seleção já existente para as setas. ok=false se key não for uma
// dessas combinações.
func translateWordStarKey(key tcell.Key) (newKey tcell.Key, mod tcell.ModMask, ok bool) {
	switch key {
	case tcell.KeyCtrlS: // Character left
		return tcell.KeyLeft, tcell.ModNone, true
	case tcell.KeyCtrlD: // Character right
		return tcell.KeyRight, tcell.ModNone, true
	case tcell.KeyCtrlA: // Word left
		return tcell.KeyLeft, tcell.ModCtrl, true
	case tcell.KeyCtrlF: // Word right
		return tcell.KeyRight, tcell.ModCtrl, true
	case tcell.KeyCtrlE: // Line up
		return tcell.KeyUp, tcell.ModNone, true
	case tcell.KeyCtrlX: // Line down
		return tcell.KeyDown, tcell.ModNone, true
	case tcell.KeyCtrlW: // Scroll up (rola a tela sem mover o cursor)
		return tcell.KeyUp, tcell.ModAlt, true
	case tcell.KeyCtrlZ: // Scroll down (rola a tela sem mover o cursor)
		return tcell.KeyDown, tcell.ModAlt, true
	case tcell.KeyCtrlR: // Page up
		return tcell.KeyPgUp, tcell.ModNone, true
	case tcell.KeyCtrlC: // Page down
		return tcell.KeyPgDn, tcell.ModNone, true
	}
	return 0, 0, false
}

// handleNavigationKey trata movimentação de cursor (setas, Home/End, PgUp/PgDn,
// Ctrl+setas, Ctrl+Home/End, os atalhos WordStar/Turbo Ctrl+S/D/A/F/E/X/W/Z/R/C)
// e a seleção de texto com Shift, reaproveitando o destaque de bloco (blk*) já
// usado pelos comandos Ctrl-K/Clipboard. Retorna true se o evento foi tratado.
func (w *editorWindow) handleNavigationKey(event *tcell.EventKey, setFocus func(tview.Primitive)) bool {
	key := event.Key()
	mod := event.Modifiers()

	if translatedKey, translatedMod, translated := translateWordStarKey(key); translated {
		key, mod = translatedKey, translatedMod
		event = tcell.NewEventKey(key, 0, mod)
	} else if !isNavigationKey(key) {
		return false
	}

	shift := mod&tcell.ModShift != 0
	if shift {
		if !w.shiftSelecting {
			row, col, _, _ := w.editor.GetCursor()
			w.selAnchorRow, w.selAnchorCol = row, col
			w.shiftSelecting = true
		}
	} else if w.shiftSelecting {
		w.shiftSelecting = false
		w.clearBlock()
	}

	switch {
	case key == tcell.KeyHome && mod&tcell.ModCtrl != 0:
		// Ctrl+Home: início do arquivo.
		w.editor.Select(0, 0)
	case key == tcell.KeyEnd && mod&tcell.ModCtrl != 0:
		// Ctrl+End: final do arquivo.
		n := len(w.editor.GetText())
		w.editor.Select(n, n)
	case (key == tcell.KeyUp || key == tcell.KeyDown) && mod&tcell.ModCtrl != 0:
		// Ctrl+Up/Down: meia tela para cima/baixo (o TextArea não tem isso nativamente).
		w.moveHalfScreen(key == tcell.KeyDown)
	default:
		// Demais teclas (setas, Ctrl+setas para palavra, Home/End, PgUp/PgDn):
		// delega ao TextArea, mas sem o bit de Shift — a seleção nativa dele
		// não é usada (o destaque é feito por nós via blk*), então retiramos
		// o Shift para manter o rastreamento interno do TextArea sempre em
		// "sem seleção", evitando ambiguidade sobre a posição real do cursor.
		if handler := w.editor.InputHandler(); handler != nil {
			stripped := tcell.NewEventKey(key, event.Rune(), mod&^tcell.ModShift)
			handler(stripped, setFocus)
		}
	}

	if w.shiftSelecting {
		row, col, _, _ := w.editor.GetCursor()
		w.blkBeginRow, w.blkBeginCol = w.selAnchorRow, w.selAnchorCol
		w.blkEndRow, w.blkEndCol = row, col
		w.normalizeBlock()
		w.blkVisible = true
	}

	return true
}

// syncMouseSelection espelha a seleção nativa do TextArea (usada para
// clique/arraste do mouse) nos campos blk*, reaproveitando o mesmo destaque
// visual e os comandos de bloco/clipboard já existentes.
func (w *editorWindow) syncMouseSelection() {
	fromRow, fromCol, toRow, toCol := w.editor.GetCursor()
	w.blkBeginRow, w.blkBeginCol = fromRow, fromCol
	w.blkEndRow, w.blkEndCol = toRow, toCol
	w.blkVisible = true
}

// moveHalfScreen move o cursor meia tela para cima (Ctrl+Up) ou para baixo
// (Ctrl+Down), preservando a coluna quando possível.
func (w *editorWindow) moveHalfScreen(down bool) {
	row, col, _, _ := w.editor.GetCursor()

	half := w.editor.GetFieldHeight() / 2
	if half < 1 {
		half = 1
	}
	if down {
		row += half
	} else {
		row -= half
	}
	if row < 0 {
		row = 0
	}

	text := w.editor.GetText()
	lines := strings.Split(text, "\n")
	if row >= len(lines) {
		row = len(lines) - 1
	}

	off := rowColToByteOffset(text, row, col)
	w.editor.Select(off, off)
}
