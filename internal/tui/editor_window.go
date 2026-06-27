package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type editorWindow struct {
	*tview.Box
	editor   *tview.TextArea
	theme    Theme
	fileName string
	number   int
}

func newEditorWindow(theme Theme, fileName string, number int) *editorWindow {
	editor := tview.NewTextArea()
	editor.SetPlaceholder("Digite seu codigo aqui...")
	editor.SetWrap(false)
	editor.SetWordWrap(false)
	editor.SetTextStyle(tcell.StyleDefault.Foreground(theme.EditorFg).Background(theme.EditorBg))
	editor.SetPlaceholderStyle(tcell.StyleDefault.Foreground(theme.EditorFg).Background(theme.EditorBg))
	editor.SetBackgroundColor(theme.EditorBg)

	return &editorWindow{
		Box:      tview.NewBox().SetBackgroundColor(theme.EditorBg),
		editor:   editor,
		theme:    theme,
		fileName: fileName,
		number:   number,
	}
}

func (w *editorWindow) Draw(screen tcell.Screen) {
	x, y, width, height := w.GetRect()
	if width < 8 || height < 6 {
		return
	}

	maxX, maxY := screen.Size()
	frameStyle := tcell.StyleDefault.Foreground(w.theme.EditorBorderFg).Background(w.theme.EditorBg)
	bgStyle := tcell.StyleDefault.Foreground(w.theme.EditorFg).Background(w.theme.EditorBg)

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	for col := 1; col < width-1; col++ {
		putRune(screen, maxX, maxY, x+col, y, '═', frameStyle)
		putRune(screen, maxX, maxY, x+col, y+height-1, '═', frameStyle)
	}
	for row := 1; row < height-1; row++ {
		putRune(screen, maxX, maxY, x, y+row, '║', frameStyle)
		putRune(screen, maxX, maxY, x+width-1, y+row, '║', frameStyle)
	}
	putRune(screen, maxX, maxY, x, y, '╔', frameStyle)
	putRune(screen, maxX, maxY, x+width-1, y, '╗', frameStyle)
	putRune(screen, maxX, maxY, x, y+height-1, '╚', frameStyle)
	putRune(screen, maxX, maxY, x+width-1, y+height-1, '╝', frameStyle)

	innerX, innerY := x+1, y+1
	innerWidth, innerHeight := width-2, height-2
	w.editor.SetRect(innerX, innerY, innerWidth, innerHeight)
	w.editor.Draw(screen)

	putString(screen, maxX, maxY, x+2, y, "[■]", frameStyle)
	title := " " + w.fileName + " "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, frameStyle)

	scrollMarkX := x + width - 7
	putString(screen, maxX, maxY, scrollMarkX, y, "[↕]", frameStyle)
	numberText := fmt.Sprintf("%d", w.number)
	putString(screen, maxX, maxY, scrollMarkX-2-len([]rune(numberText)), y, numberText, frameStyle)

	fromRow, fromCol, _, _ := w.editor.GetCursor()
	statusText := fmt.Sprintf(" %d:%d ", fromRow+1, fromCol+1)
	statusWidth := innerWidth / 4
	if statusWidth < len([]rune(statusText))+1 {
		statusWidth = len([]rune(statusText)) + 1
	}
	putString(screen, maxX, maxY, x+1, y+height-1, statusText, frameStyle)

	trackStart := x + 1 + statusWidth
	trackEnd := x + width - 2
	leftArrowX := trackStart
	rightArrowX := x + width - 4
	if rightArrowX > trackEnd {
		rightArrowX = trackEnd
	}
	if leftArrowX+1 < rightArrowX {
		putRune(screen, maxX, maxY, leftArrowX, y+height-1, '◄', frameStyle)
		putRune(screen, maxX, maxY, rightArrowX, y+height-1, '►', frameStyle)
		for col := leftArrowX + 1; col < rightArrowX; col++ {
			putRune(screen, maxX, maxY, col, y+height-1, '▒', frameStyle)
		}
		cursorX := w.horizontalCursorPosition(leftArrowX+1, rightArrowX-1)
		putRune(screen, maxX, maxY, cursorX, y+height-1, '■', frameStyle)
	}

	vTop, vBottom := y+1, y+height-2
	if vTop+1 < vBottom {
		putRune(screen, maxX, maxY, x+width-1, vTop, '▲', frameStyle)
		putRune(screen, maxX, maxY, x+width-1, vBottom, '▼', frameStyle)
		for row := vTop + 1; row < vBottom; row++ {
			putRune(screen, maxX, maxY, x+width-1, row, '▒', frameStyle)
		}
		cursorY := w.verticalCursorPosition(vTop+1, vBottom-1)
		putRune(screen, maxX, maxY, x+width-1, cursorY, '■', frameStyle)
	}
}

func (w *editorWindow) horizontalCursorPosition(start, end int) int {
	if end <= start {
		return start
	}
	_, offsetCol := w.editor.GetOffset()
	fieldWidth := w.editor.GetFieldWidth()
	if fieldWidth <= 0 {
		return start
	}

	maxLine := 0
	for _, line := range strings.Split(w.editor.GetText(), "\n") {
		if l := len([]rune(line)); l > maxLine {
			maxLine = l
		}
	}
	rangeCols := maxLine - fieldWidth
	if rangeCols <= 0 {
		return start
	}
	if offsetCol > rangeCols {
		offsetCol = rangeCols
	}
	trackLen := end - start
	return start + (offsetCol*trackLen)/rangeCols
}

func (w *editorWindow) verticalCursorPosition(start, end int) int {
	if end <= start {
		return start
	}
	offsetRow, _ := w.editor.GetOffset()
	fieldHeight := w.editor.GetFieldHeight()
	if fieldHeight <= 0 {
		return start
	}

	totalRows := len(strings.Split(w.editor.GetText(), "\n"))
	rangeRows := totalRows - fieldHeight
	if rangeRows <= 0 {
		return start
	}
	if offsetRow > rangeRows {
		offsetRow = rangeRows
	}
	trackLen := end - start
	return start + (offsetRow*trackLen)/rangeRows
}

func (w *editorWindow) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return w.editor.InputHandler()
}

func (w *editorWindow) Focus(delegate func(p tview.Primitive)) {
	w.editor.Focus(delegate)
}

func (w *editorWindow) HasFocus() bool {
	return w.editor.HasFocus()
}

func (w *editorWindow) Blur() {
	w.editor.Blur()
}

func (w *editorWindow) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return w.editor.MouseHandler()
}

func (w *editorWindow) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return w.editor.PasteHandler()
}

func putRune(screen tcell.Screen, maxX, maxY, x, y int, r rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= maxX || y >= maxY {
		return
	}
	screen.SetContent(x, y, r, nil, style)
}

func putString(screen tcell.Screen, maxX, maxY, x, y int, text string, style tcell.Style) {
	for i, r := range []rune(text) {
		putRune(screen, maxX, maxY, x+i, y, r, style)
	}
}
