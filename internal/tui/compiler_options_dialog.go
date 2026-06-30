package tui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type compilerOptionsDialog struct {
	*tview.Box
	app      *App
	pageName string

	// Radio "Basic Code": 0=MSX-BASIC  1=MSX Bas2Rom  2=Basic Dignified
	radioOptions []string
	radioIndex   int

	// Checkboxes "Others": 0=NBasic  1=Turbo Basic
	othersLabels []string
	others       []bool

	// Checkbox "Token" — visível apenas quando MSX-BASIC está selecionado
	tokenCheckbox bool

	// Campo de defines condicionais
	conditional []string
	cursorRow   int
	cursorCol   int

	focusIndex int

	okButton     *turboButton
	cancelButton *turboButton
	helpButton   *turboButton

	onClose func()
}

// Layout (altura 18):
//   y+0  borda superior
//   y+1  vazio
//   y+2  "Basic Code" (label)
//   y+3  (x) MSX-BASIC        (x) MSX Bas2Rom
//   y+4  (x) Basic Dignified
//   y+5  vazio
//   y+6  "Others" (label)
//   y+7  [•] NBasic            [•] Turbo Basic
//   y+8  [•] Token  (só em modo MSX-BASIC)
//   y+9  "Conditional defines:"
//   y+10..y+13  text area (height=4)
//   y+14 vazio
//   y+15 botões (height-3)
//   y+16 sombra dos botões
//   y+17 borda inferior

const (
	compilerOptionsTextAreaHeight = 4

	// Foco: 0-2 radio, 3-4 checkboxes Others, 5 Token (só MSX-BASIC), 6 textarea, 7 OK, 8 Cancel, 9 Help
	compilerOptionsFocusToken    = 5
	compilerOptionsFocusTextArea = 6
	compilerOptionsFocusOK       = 7
	compilerOptionsFocusCancel   = 8
	compilerOptionsFocusHelp     = 9
	compilerOptionsFocusMax      = 9
)

func newCompilerOptionsDialog(app *App) *compilerOptionsDialog {
	tokenVal := false
	if len(app.Editors) > 0 && app.ActiveEditor >= 0 && app.ActiveEditor < len(app.Editors) {
		tokenVal = app.Editors[app.ActiveEditor].isTokenized
	}
	d := &compilerOptionsDialog{
		Box:      tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:      app,
		pageName: "compiler_options",
		radioOptions: []string{
			"MSX-BASIC",
			"MSX Bas2Rom",
			"Basic Dignified",
		},
		radioIndex:    app.CompilerMode,
		othersLabels:  []string{"NBasic", "Turbo Basic"},
		others:        []bool{false, false},
		tokenCheckbox: tokenVal,
		conditional:   []string{""},
		okButton:      newTurboButton("  &OK  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		cancelButton:  newTurboButton("&Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		helpButton:    newTurboButton(" &Help ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	return d
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (d *compilerOptionsDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < 52 || height < 16 {
		return
	}

	maxX, maxY := screen.Size()
	bgStyle     := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)
	focusStyle  := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	labelStyle  := tcell.StyleDefault.Foreground(vgaBlue).Background(d.app.Theme.PopupBg)

	// Preenche fundo
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	// Borda dupla
	for col := 1; col < width-1; col++ {
		putRune(screen, maxX, maxY, x+col, y, '═', borderStyle)
		putRune(screen, maxX, maxY, x+col, y+height-1, '═', borderStyle)
	}
	for row := 1; row < height-1; row++ {
		putRune(screen, maxX, maxY, x, y+row, '║', borderStyle)
		putRune(screen, maxX, maxY, x+width-1, y+row, '║', borderStyle)
	}
	putRune(screen, maxX, maxY, x, y, '╔', borderStyle)
	putRune(screen, maxX, maxY, x+width-1, y, '╗', borderStyle)
	putRune(screen, maxX, maxY, x, y+height-1, '╚', borderStyle)
	putRune(screen, maxX, maxY, x+width-1, y+height-1, '╝', borderStyle)

	// Título
	title := " Compiler/Interpreter Options "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	// [■] fechar
	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	leftCol  := x + 4
	rightCol := x + width/2

	// Seção "Basic Code" — radio buttons
	putString(screen, maxX, maxY, x+2, y+2, "Basic Code", labelStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+3, 0, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, rightCol, y+3, 1, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+4, 2, bgStyle, focusStyle)

	// Seção "Others" — checkboxes lado a lado
	putString(screen, maxX, maxY, x+2, y+6, "Others", labelStyle)
	d.drawOthers(screen, maxX, maxY, leftCol, y+7, 0, bgStyle, focusStyle)
	d.drawOthers(screen, maxX, maxY, rightCol, y+7, 1, bgStyle, focusStyle)

	// Checkbox "Token" — apenas quando MSX-BASIC selecionado
	if d.radioIndex == 0 {
		d.drawToken(screen, maxX, maxY, leftCol, y+8, bgStyle, focusStyle)
	}

	// Defines condicionais
	putString(screen, maxX, maxY, x+2, y+9, "Conditional defines:", labelStyle)
	d.drawTextArea(screen, maxX, maxY, x+3, y+10, width-6, compilerOptionsTextAreaHeight, bgStyle, focusStyle)

	d.drawButtons(screen, maxX, maxY, x, y, width, height)
}

func (d *compilerOptionsDialog) drawRadio(screen tcell.Screen, maxX, maxY, x, y, index int, baseStyle, focusStyle tcell.Style) {
	mark := ' '
	if d.radioIndex == index {
		mark = '●'
	}
	text := "(" + string(mark) + ") " + d.radioOptions[index]
	style := baseStyle
	if d.focusIndex == index {
		style = focusStyle
	}
	putString(screen, maxX, maxY, x, y, text, style)
}

// drawOthers desenha um checkbox da seção "Others".
// O foco para Others começa em index 3 (3+0=NBasic, 3+1=Turbo Basic).
func (d *compilerOptionsDialog) drawOthers(screen tcell.Screen, maxX, maxY, x, y, index int, baseStyle, focusStyle tcell.Style) {
	mark := ' '
	if d.others[index] {
		mark = '■'
	}
	text := "[" + string(mark) + "] " + d.othersLabels[index]
	style := baseStyle
	if d.focusIndex == 3+index {
		style = focusStyle
	}
	putString(screen, maxX, maxY, x, y, text, style)
}

func (d *compilerOptionsDialog) drawToken(screen tcell.Screen, maxX, maxY, x, y int, baseStyle, focusStyle tcell.Style) {
	mark := ' '
	if d.tokenCheckbox {
		mark = '■'
	}
	text := "[" + string(mark) + "] Token"
	style := baseStyle
	if d.focusIndex == compilerOptionsFocusToken {
		style = focusStyle
	}
	putString(screen, maxX, maxY, x, y, text, style)
}

func (d *compilerOptionsDialog) drawTextArea(screen tcell.Screen, maxX, maxY, x, y, width, height int, baseStyle, focusStyle tcell.Style) {
	for col := 0; col < width; col++ {
		putRune(screen, maxX, maxY, x+col, y, '─', baseStyle)
		putRune(screen, maxX, maxY, x+col, y+height-1, '─', baseStyle)
	}
	for row := 1; row < height-1; row++ {
		putRune(screen, maxX, maxY, x, y+row, '│', baseStyle)
		putRune(screen, maxX, maxY, x+width-1, y+row, '│', baseStyle)
		for col := 1; col < width-1; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', baseStyle)
		}
	}
	putRune(screen, maxX, maxY, x, y, '┌', baseStyle)
	putRune(screen, maxX, maxY, x+width-1, y, '┐', baseStyle)
	putRune(screen, maxX, maxY, x, y+height-1, '└', baseStyle)
	putRune(screen, maxX, maxY, x+width-1, y+height-1, '┘', baseStyle)

	maxLines := height - 2
	for row := 0; row < maxLines && row < len(d.conditional); row++ {
		putString(screen, maxX, maxY, x+1, y+1+row, d.conditional[row], baseStyle)
	}

	if d.focusIndex == compilerOptionsFocusTextArea {
		cursorX := x + 1 + d.cursorCol
		cursorY := y + 1 + d.cursorRow
		if d.cursorRow < len(d.conditional) {
			lineRunes := []rune(d.conditional[d.cursorRow])
			if d.cursorCol < len(lineRunes) {
				putRune(screen, maxX, maxY, cursorX, cursorY, lineRunes[d.cursorCol], focusStyle)
			} else {
				putRune(screen, maxX, maxY, cursorX, cursorY, ' ', focusStyle)
			}
		}
	}
}

func (d *compilerOptionsDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	btnY := y + height - 3
	okW     := d.okButton.width()
	cancelW := d.cancelButton.width()
	helpW   := d.helpButton.width()
	total   := okW + cancelW + helpW + 4
	btnX    := x + (width-total)/2

	drawBtn := func(button *turboButton, focus bool, atX int) {
		if focus {
			focused := newTurboButton(button.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			focused.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
			return
		}
		button.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
	}

	drawBtn(d.okButton, d.focusIndex == compilerOptionsFocusOK, btnX)
	drawBtn(d.cancelButton, d.focusIndex == compilerOptionsFocusCancel, btnX+okW+2)
	drawBtn(d.helpButton, d.focusIndex == compilerOptionsFocusHelp, btnX+okW+2+cancelW+2)
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *compilerOptionsDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}

		if d.focusIndex == compilerOptionsFocusTextArea && d.handleTextAreaInput(event) {
			return
		}

		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			return
		case tcell.KeyTab:
			d.focusIndex = d.advanceFocus(1)
			return
		case tcell.KeyBacktab:
			d.focusIndex = d.advanceFocus(-1)
			return
		case tcell.KeyUp:
			d.focusIndex = d.advanceFocus(-1)
			return
		case tcell.KeyDown:
			d.focusIndex = d.advanceFocus(1)
			return
		case tcell.KeyLeft:
			if d.focusIndex >= compilerOptionsFocusOK && d.focusIndex <= compilerOptionsFocusHelp {
				d.focusIndex--
				if d.focusIndex < compilerOptionsFocusOK {
					d.focusIndex = compilerOptionsFocusHelp
				}
				return
			}
		case tcell.KeyRight:
			if d.focusIndex >= compilerOptionsFocusOK && d.focusIndex <= compilerOptionsFocusHelp {
				d.focusIndex++
				if d.focusIndex > compilerOptionsFocusHelp {
					d.focusIndex = compilerOptionsFocusOK
				}
				return
			}
		case tcell.KeyEnter:
			d.activateFocused()
			return
		case tcell.KeyRune:
			switch unicode.ToLower(event.Rune()) {
			case 'o':
				d.applyAndClose()
				return
			case 'c':
				d.close()
				return
			case 'h':
				d.focusIndex = compilerOptionsFocusHelp
				d.app.showCompilerOptionsHelp()
				return
			case ' ':
				d.toggleFocusedOption()
				return
			}
		default:
			return
		}
	})
}

// advanceFocus avança o foco em delta (+1/-1), saltando Token quando não é MSX-BASIC.
func (d *compilerOptionsDialog) advanceFocus(delta int) int {
	f := d.focusIndex
	for i := 0; i <= compilerOptionsFocusMax; i++ {
		f = (f + delta + compilerOptionsFocusMax + 1) % (compilerOptionsFocusMax + 1)
		if f == compilerOptionsFocusToken && d.radioIndex != 0 {
			continue
		}
		break
	}
	return f
}

func (d *compilerOptionsDialog) handleTextAreaInput(event *tcell.EventKey) bool {
	if d.focusIndex != compilerOptionsFocusTextArea {
		return false
	}
	if len(d.conditional) == 0 {
		d.conditional = []string{""}
	}
	maxLines := compilerOptionsTextAreaHeight - 2
	maxCols  := d.GetInnerRectWidth() - 8
	if maxCols < 1 {
		maxCols = 1
	}

	line := []rune(d.conditional[d.cursorRow])
	switch event.Key() {
	case tcell.KeyLeft:
		if d.cursorCol > 0 {
			d.cursorCol--
		} else if d.cursorRow > 0 {
			d.cursorRow--
			d.cursorCol = len([]rune(d.conditional[d.cursorRow]))
		}
		return true
	case tcell.KeyRight:
		if d.cursorCol < len(line) {
			d.cursorCol++
		} else if d.cursorRow+1 < len(d.conditional) {
			d.cursorRow++
			d.cursorCol = 0
		}
		return true
	case tcell.KeyUp:
		if d.cursorRow > 0 {
			d.cursorRow--
			if d.cursorCol > len([]rune(d.conditional[d.cursorRow])) {
				d.cursorCol = len([]rune(d.conditional[d.cursorRow]))
			}
		}
		return true
	case tcell.KeyDown:
		if d.cursorRow+1 < len(d.conditional) {
			d.cursorRow++
			if d.cursorCol > len([]rune(d.conditional[d.cursorRow])) {
				d.cursorCol = len([]rune(d.conditional[d.cursorRow]))
			}
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.cursorCol > 0 {
			left := append([]rune{}, line[:d.cursorCol-1]...)
			left = append(left, line[d.cursorCol:]...)
			d.conditional[d.cursorRow] = string(left)
			d.cursorCol--
			return true
		}
		if d.cursorRow > 0 {
			prev := []rune(d.conditional[d.cursorRow-1])
			curr := []rune(d.conditional[d.cursorRow])
			if len(prev)+len(curr) <= maxCols {
				d.cursorCol = len(prev)
				d.conditional[d.cursorRow-1] = string(append(prev, curr...))
				d.conditional = append(d.conditional[:d.cursorRow], d.conditional[d.cursorRow+1:]...)
				d.cursorRow--
			}
		}
		return true
	case tcell.KeyEnter:
		if len(d.conditional) >= maxLines {
			return true
		}
		left  := string(line[:d.cursorCol])
		right := string(line[d.cursorCol:])
		d.conditional[d.cursorRow] = left
		d.conditional = append(d.conditional[:d.cursorRow+1], append([]string{right}, d.conditional[d.cursorRow+1:]...)...)
		d.cursorRow++
		d.cursorCol = 0
		return true
	case tcell.KeyRune:
		r := event.Rune()
		if unicode.IsPrint(r) {
			if len(line) >= maxCols {
				return true
			}
			updated := append([]rune{}, line[:d.cursorCol]...)
			updated = append(updated, r)
			updated = append(updated, line[d.cursorCol:]...)
			d.conditional[d.cursorRow] = string(updated)
			d.cursorCol++
			return true
		}
	default:
		return false
	}
	return false
}

func (d *compilerOptionsDialog) GetInnerRectWidth() int {
	_, _, width, _ := d.GetRect()
	return width
}

func (d *compilerOptionsDialog) applyAndClose() {
	d.app.CompilerMode = d.radioIndex
	if d.radioIndex == 0 && len(d.app.Editors) > 0 && d.app.ActiveEditor >= 0 && d.app.ActiveEditor < len(d.app.Editors) {
		d.app.Editors[d.app.ActiveEditor].isTokenized = d.tokenCheckbox
	}
	d.close()
}

func (d *compilerOptionsDialog) activateFocused() {
	switch {
	case d.focusIndex >= 0 && d.focusIndex <= 2: // radio buttons
		d.radioIndex = d.focusIndex
	case d.focusIndex >= 3 && d.focusIndex <= 4: // Others checkboxes
		d.others[d.focusIndex-3] = !d.others[d.focusIndex-3]
	case d.focusIndex == compilerOptionsFocusToken:
		d.tokenCheckbox = !d.tokenCheckbox
	case d.focusIndex == compilerOptionsFocusOK:
		d.applyAndClose()
	case d.focusIndex == compilerOptionsFocusCancel:
		d.close()
	case d.focusIndex == compilerOptionsFocusHelp:
		d.app.showCompilerOptionsHelp()
	}
}

func (d *compilerOptionsDialog) toggleFocusedOption() {
	switch {
	case d.focusIndex >= 0 && d.focusIndex <= 2:
		d.radioIndex = d.focusIndex
	case d.focusIndex >= 3 && d.focusIndex <= 4:
		d.others[d.focusIndex-3] = !d.others[d.focusIndex-3]
	case d.focusIndex == compilerOptionsFocusToken:
		d.tokenCheckbox = !d.tokenCheckbox
	}
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *compilerOptionsDialog) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		if action != tview.MouseLeftClick {
			return false, nil
		}
		mx, my := event.Position()
		x, y, width, height := d.GetRect()
		if mx < x || mx >= x+width || my < y || my >= y+height {
			return false, nil
		}

		setFocus(d)

		// [■] fechar
		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		leftCol  := x + 4
		rightCol := x + width/2

		// Clique nos radio buttons "Basic Code"
		if my == y+3 {
			if mx >= leftCol {
				if mx < rightCol {
					d.focusIndex = 0
					d.radioIndex = 0
				} else {
					d.focusIndex = 1
					d.radioIndex = 1
				}
			}
			return true, nil
		}
		if my == y+4 && mx >= leftCol && mx < rightCol {
			d.focusIndex = 2
			d.radioIndex = 2
			return true, nil
		}

		// Clique nos checkboxes "Others"
		if my == y+7 {
			if mx >= leftCol && mx < rightCol {
				d.focusIndex = 3
				d.others[0] = !d.others[0]
			} else if mx >= rightCol {
				d.focusIndex = 4
				d.others[1] = !d.others[1]
			}
			return true, nil
		}

		// Clique no checkbox "Token" (só visível em modo MSX-BASIC)
		if my == y+8 && d.radioIndex == 0 && mx >= leftCol {
			d.focusIndex = compilerOptionsFocusToken
			d.tokenCheckbox = !d.tokenCheckbox
			return true, nil
		}

		// Botões
		btnY    := y + height - 3
		okW     := d.okButton.width()
		cancelW := d.cancelButton.width()
		helpW   := d.helpButton.width()
		total   := okW + cancelW + helpW + 4
		btnX    := x + (width-total)/2

		if my == btnY {
			switch {
			case mx >= btnX && mx < btnX+okW:
				d.focusIndex = compilerOptionsFocusOK
				d.activateFocused()
			case mx >= btnX+okW+2 && mx < btnX+okW+2+cancelW:
				d.focusIndex = compilerOptionsFocusCancel
				d.activateFocused()
			case mx >= btnX+okW+2+cancelW+2 && mx < btnX+okW+2+cancelW+2+helpW:
				d.focusIndex = compilerOptionsFocusHelp
				d.activateFocused()
			}
			return true, nil
		}

		return true, nil
	}
}

func (d *compilerOptionsDialog) close() {
	d.app.Pages.RemovePage(d.pageName)
	if d.onClose != nil {
		d.onClose()
		return
	}
	if len(d.app.Editors) > 0 && d.app.ActiveEditor >= 0 && d.app.ActiveEditor < len(d.app.Editors) {
		d.app.Application.SetFocus(d.app.Editors[d.app.ActiveEditor])
		return
	}
	if d.app.Editor != nil {
		d.app.Application.SetFocus(d.app.Editor)
	}
}

func showCompilerOptionsDialogCentered(dialog *compilerOptionsDialog, width, height int) {
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dialog, width, 0, true).
			AddItem(nil, 0, 1, false), height, 0, true).
		AddItem(nil, 0, 1, false)

	dialog.app.Pages.AddPage(dialog.pageName, container, true, true)
	dialog.app.Application.SetFocus(dialog)
}
