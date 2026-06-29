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

	radioOptions []string
	radioIndex   int

	checkboxLabels []string
	checkboxes     []bool

	conditional []string
	cursorRow   int
	cursorCol   int

	focusIndex int

	okButton     *turboButton
	cancelButton *turboButton
	helpButton   *turboButton

	onClose func()
}

const (
	compilerOptionsTextAreaHeight = 4
	compilerOptionsFocusTextArea  = 12
	compilerOptionsFocusOK        = 13
	compilerOptionsFocusCancel    = 14
	compilerOptionsFocusHelp      = 15
	compilerOptionsFocusMax       = 15
)

func newCompilerOptionsDialog(app *App) *compilerOptionsDialog {
	return &compilerOptionsDialog{
		Box:      tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:      app,
		pageName: "compiler_options",
		radioOptions: []string{
			"MSX-BASIC",
			"Basic Dignified",
			"MSX Bas2Rom",
			"Turbo Basic",
			"NBasic",
			"MSXgl/SDCC",
			"N80/LK80",
			"ASCII-C",
			"Turbo Pascal 3.3f",
		},
		radioIndex: 0,
		checkboxLabels: []string{
			"Extended syntax",
			"Overflow checking",
			"Strict vars",
		},
		checkboxes: []bool{false, false, false},
		conditional: []string{
			"",
		},
		okButton:     newTurboButton("&OK", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		cancelButton: newTurboButton("&Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		helpButton:   newTurboButton("&Help", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
}

func (d *compilerOptionsDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < 64 || height < 22 {
		return
	}

	maxX, maxY := screen.Size()
	bgStyle := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)
	focusStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	labelStyle := tcell.StyleDefault.Foreground(vgaBlue).Background(d.app.Theme.PopupBg)

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

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

	title := " Compiler/Interpreter Options "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	leftCol := x + 4
	rightCol := x + width/2

	putString(screen, maxX, maxY, x+2, y+2, "Basic Code", labelStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+3, 0, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, rightCol, y+3, 1, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+4, 2, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, rightCol, y+4, 3, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+5, 4, bgStyle, focusStyle)

	putString(screen, maxX, maxY, x+2, y+7, "Others", labelStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+8, 5, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, rightCol, y+8, 6, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, leftCol, y+9, 7, bgStyle, focusStyle)
	d.drawRadio(screen, maxX, maxY, rightCol, y+9, 8, bgStyle, focusStyle)

	d.drawCheckbox(screen, maxX, maxY, x+2, y+11, 0, bgStyle, focusStyle)
	d.drawCheckbox(screen, maxX, maxY, x+2, y+12, 1, bgStyle, focusStyle)
	d.drawCheckbox(screen, maxX, maxY, x+2, y+13, 2, bgStyle, focusStyle)

	putString(screen, maxX, maxY, x+2, y+15, "Conditional defines:", labelStyle)
	d.drawTextArea(screen, maxX, maxY, x+3, y+16, width-6, compilerOptionsTextAreaHeight, bgStyle, focusStyle)

	d.drawButtons(screen, maxX, maxY, x, y, width, height)
}

func (d *compilerOptionsDialog) drawRadio(screen tcell.Screen, maxX, maxY, x, y, index int, baseStyle, focusStyle tcell.Style) {
	mark := ' '
	if d.radioIndex == index {
		mark = 'x'
	}
	text := "(" + string(mark) + ") " + d.radioOptions[index]
	style := baseStyle
	if d.focusIndex == index {
		style = focusStyle
	}
	putString(screen, maxX, maxY, x, y, text, style)
}

func (d *compilerOptionsDialog) drawCheckbox(screen tcell.Screen, maxX, maxY, x, y, index int, baseStyle, focusStyle tcell.Style) {
	mark := ' '
	if d.checkboxes[index] {
		mark = '•'
	}
	text := "[" + string(mark) + "] " + d.checkboxLabels[index]
	style := baseStyle
	if d.focusIndex == 9+index {
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
		cursorStyle := focusStyle
		if d.cursorRow < len(d.conditional) {
			lineRunes := []rune(d.conditional[d.cursorRow])
			if d.cursorCol < len(lineRunes) {
				putRune(screen, maxX, maxY, cursorX, cursorY, lineRunes[d.cursorCol], cursorStyle)
			} else {
				putRune(screen, maxX, maxY, cursorX, cursorY, ' ', cursorStyle)
			}
		}
	}
}

func (d *compilerOptionsDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	btnY := y + height - 3
	okW := d.okButton.width()
	cancelW := d.cancelButton.width()
	helpW := d.helpButton.width()
	total := okW + cancelW + helpW + 4
	btnX := x + (width-total)/2

	drawBtn := func(button *turboButton, focus bool, atX int) {
		if focus {
			focused := newTurboButton(button.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			focused.SetShadowMode(shadowModeFlat)
			focused.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
			return
		}
		button.SetShadowMode(shadowModeFlat)
		button.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
	}

	drawBtn(d.okButton, d.focusIndex == compilerOptionsFocusOK, btnX)
	drawBtn(d.cancelButton, d.focusIndex == compilerOptionsFocusCancel, btnX+okW+2)
	drawBtn(d.helpButton, d.focusIndex == compilerOptionsFocusHelp, btnX+okW+2+cancelW+2)
}

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
			d.focusIndex = (d.focusIndex + 1) % (compilerOptionsFocusMax + 1)
			return
		case tcell.KeyBacktab:
			d.focusIndex = (d.focusIndex - 1 + compilerOptionsFocusMax + 1) % (compilerOptionsFocusMax + 1)
			return
		case tcell.KeyUp:
			d.focusIndex = (d.focusIndex - 1 + compilerOptionsFocusMax + 1) % (compilerOptionsFocusMax + 1)
			return
		case tcell.KeyDown:
			d.focusIndex = (d.focusIndex + 1) % (compilerOptionsFocusMax + 1)
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
			r := unicode.ToLower(event.Rune())
			switch r {
			case 'o':
				d.focusIndex = compilerOptionsFocusOK
				d.close()
				return
			case 'c':
				d.focusIndex = compilerOptionsFocusCancel
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

func (d *compilerOptionsDialog) handleTextAreaInput(event *tcell.EventKey) bool {
	if d.focusIndex != compilerOptionsFocusTextArea {
		return false
	}
	if len(d.conditional) == 0 {
		d.conditional = []string{""}
	}
	maxLines := compilerOptionsTextAreaHeight - 2
	maxCols := d.GetInnerRectWidth() - 8
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
			return true
		}
		return true
	case tcell.KeyEnter:
		if len(d.conditional) >= maxLines {
			return true
		}
		left := string(line[:d.cursorCol])
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

func (d *compilerOptionsDialog) activateFocused() {
	switch {
	case d.focusIndex >= 0 && d.focusIndex <= 8:
		d.radioIndex = d.focusIndex
	case d.focusIndex >= 9 && d.focusIndex <= 11:
		d.checkboxes[d.focusIndex-9] = !d.checkboxes[d.focusIndex-9]
	case d.focusIndex == compilerOptionsFocusOK:
		d.close()
	case d.focusIndex == compilerOptionsFocusCancel:
		d.close()
	case d.focusIndex == compilerOptionsFocusHelp:
		d.app.showCompilerOptionsHelp()
	}
}

func (d *compilerOptionsDialog) toggleFocusedOption() {
	if d.focusIndex >= 0 && d.focusIndex <= 8 {
		d.radioIndex = d.focusIndex
		return
	}
	if d.focusIndex >= 9 && d.focusIndex <= 11 {
		d.checkboxes[d.focusIndex-9] = !d.checkboxes[d.focusIndex-9]
	}
}

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
		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		btnY := y + height - 3
		okW := d.okButton.width()
		cancelW := d.cancelButton.width()
		helpW := d.helpButton.width()
		total := okW + cancelW + helpW + 4
		btnX := x + (width-total)/2

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


