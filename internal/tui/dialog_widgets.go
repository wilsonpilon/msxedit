package tui

import (
	"github.com/gdamore/tcell/v2"
)

// drawGroupBox desenha uma caixa com borda simples e título embutido na
// borda superior, no estilo "┌─Title────┐" usado pelos diálogos do Borland.
func drawGroupBox(screen tcell.Screen, maxX, maxY, x, y, w, h int, title string, borderStyle, titleStyle tcell.Style) {
	for col := 1; col < w-1; col++ {
		putRune(screen, maxX, maxY, x+col, y, '─', borderStyle)
		putRune(screen, maxX, maxY, x+col, y+h-1, '─', borderStyle)
	}
	for row := 1; row < h-1; row++ {
		putRune(screen, maxX, maxY, x, y+row, '│', borderStyle)
		putRune(screen, maxX, maxY, x+w-1, y+row, '│', borderStyle)
	}
	putRune(screen, maxX, maxY, x, y, '┌', borderStyle)
	putRune(screen, maxX, maxY, x+w-1, y, '┐', borderStyle)
	putRune(screen, maxX, maxY, x, y+h-1, '└', borderStyle)
	putRune(screen, maxX, maxY, x+w-1, y+h-1, '┘', borderStyle)
	putString(screen, maxX, maxY, x+2, y, title, titleStyle)
}

// putAccelLabel desenha um rótulo com "&acelerador" destacado.
func putAccelLabel(screen tcell.Screen, maxX, maxY, x, y int, label string, baseStyle, accelStyle tcell.Style) {
	plain, accel := stripAccelerator(label)
	for i, r := range []rune(plain) {
		style := baseStyle
		if i == accel {
			style = accelStyle
		}
		putRune(screen, maxX, maxY, x+i, y, r, style)
	}
}

// drawCheckbox desenha "[x] Label" com o acelerador destacado.
func drawCheckbox(screen tcell.Screen, maxX, maxY, x, y int, label string, checked, focused bool, baseStyle, focusStyle, accelStyle tcell.Style) {
	mark := ' '
	if checked {
		mark = '■'
	}
	markStyle := baseStyle
	if focused {
		markStyle = focusStyle
	}
	putString(screen, maxX, maxY, x, y, "["+string(mark)+"] ", markStyle)
	labelBase := baseStyle
	if focused {
		labelBase = focusStyle
		accelStyle = focusStyle
	}
	putAccelLabel(screen, maxX, maxY, x+3, y, label, labelBase, accelStyle)
}

// drawRadio desenha "(x) Label" com o acelerador destacado.
func drawRadio(screen tcell.Screen, maxX, maxY, x, y int, label string, selected, focused bool, baseStyle, focusStyle, accelStyle tcell.Style) {
	mark := ' '
	if selected {
		mark = '●'
	}
	markStyle := baseStyle
	if focused {
		markStyle = focusStyle
	}
	putString(screen, maxX, maxY, x, y, "("+string(mark)+") ", markStyle)
	labelBase := baseStyle
	if focused {
		labelBase = focusStyle
		accelStyle = focusStyle
	}
	putAccelLabel(screen, maxX, maxY, x+3, y, label, labelBase, accelStyle)
}
