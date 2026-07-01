package tui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FindParams reúne as opções escolhidas no diálogo Find (Search > Find...).
type FindParams struct {
	Text          string
	CaseSensitive bool
	WholeWords    bool
	Regex         bool
	Backward      bool // Direction: false=Forward true=Backward
	SelectedOnly  bool // Scope: false=Global true=Selected text
	EntireScope   bool // Origin: false=From cursor true=Entire scope
}

// ── Layout (largura 62, altura 18) ──────────────────────────────────────────
//
//  y+0   ╔══════════════════════════ Find ═══════════════════════════╗
//  y+1   ║                                                            ║
//  y+2   ║ Text to find [___________________________________]  ↓     ║
//  y+3   ║                                                            ║
//  y+4   ║ ┌─Options───────────────────┐ ┌─Direction────────────────┐ ║
//  y+5   ║ │ [ ] Case sensitive         │ │ ( ) Forward               │ ║
//  y+6   ║ │ [ ] Whole words only       │ │ ( ) Backward              │ ║
//  y+7   ║ │ [ ] Regular expression     │ │                           │ ║
//  y+8   ║ └────────────────────────────┘ └───────────────────────────┘ ║
//  y+9   ║                                                            ║
//  y+10  ║ ┌─Scope─────────────────────┐ ┌─Origin───────────────────┐ ║
//  y+11  ║ │ ( ) Global                  │ │ ( ) From cursor           │ ║
//  y+12  ║ │ ( ) Selected text           │ │ ( ) Entire scope          │ ║
//  y+13  ║ └────────────────────────────┘ └───────────────────────────┘ ║
//  y+14  ║                                                            ║
//  y+15  ║              [ &Ok ]    [Cancel]    [Help]                ║
//  y+16  ║               (shadow)   (shadow)    (shadow)              ║
//  y+17  ╚════════════════════════════════════════════════════════════╝

const (
	findDlgW = 62
	findDlgH = 18

	findRowText    = 2
	findRowBoxTop1 = 4
	findRowBoxTop2 = 10

	findLeftBoxX = 2
	findLeftBoxW = 30
	findGap      = 2
	findBoxH1    = 5 // Options (3 itens) / Direction (2 itens + 1 linha em branco)
	findBoxH2    = 4 // Scope (2 itens) / Origin (2 itens)

	findTextLabel = "&Text to find "
)

// Foco: 0=campo de texto, 1-3=Options, 4-5=Direction, 6-7=Scope, 8-9=Origin,
// 10=OK, 11=Cancel, 12=Help
const (
	findFocusText       = 0
	findFocusCase       = 1
	findFocusWholeWords = 2
	findFocusRegex      = 3
	findFocusForward    = 4
	findFocusBackward   = 5
	findFocusGlobal     = 6
	findFocusSelected   = 7
	findFocusFromCursor = 8
	findFocusEntireScop = 9
	findFocusOK         = 10
	findFocusCancel     = 11
	findFocusHelp       = 12
	findFocusMax        = 12
)

type findDialog struct {
	*tview.Box
	app      *App
	pageName string

	field *historyField

	caseSensitive bool
	wholeWords    bool
	regex         bool

	directionIdx int // 0=Forward 1=Backward
	scopeIdx     int // 0=Global  1=Selected text
	originIdx    int // 0=From cursor 1=Entire scope

	focusIndex int

	okButton     *turboButton
	cancelButton *turboButton
	helpButton   *turboButton

	onFind  func(params FindParams)
	onClose func()
}

func newFindDialog(app *App, params FindParams, history []string) *findDialog {
	d := &findDialog{
		Box:           tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:           app,
		pageName:      "find_dialog",
		field:         newHistoryField(params.Text, history),
		caseSensitive: params.CaseSensitive,
		wholeWords:    params.WholeWords,
		regex:         params.Regex,
		focusIndex:    findFocusText,
		okButton:      newTurboButton("&Ok", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		cancelButton:  newTurboButton("Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		helpButton:    newTurboButton("Help", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	if params.Backward {
		d.directionIdx = 1
	}
	if params.SelectedOnly {
		d.scopeIdx = 1
	}
	if params.EntireScope {
		d.originIdx = 1
	}
	return d
}

func (d *findDialog) currentParams() FindParams {
	return FindParams{
		Text:          d.field.value(),
		CaseSensitive: d.caseSensitive,
		WholeWords:    d.wholeWords,
		Regex:         d.regex,
		Backward:      d.directionIdx == 1,
		SelectedOnly:  d.scopeIdx == 1,
		EntireScope:   d.originIdx == 1,
	}
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (d *findDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < findDlgW || height < findDlgH {
		return
	}

	maxX, maxY := screen.Size()
	bgStyle := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)
	labelStyle := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	focusStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	accelStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)

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

	title := " Find "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	arrowX := x + findDlgW - 3
	textInputX := x + 2 + len([]rune(stripAcc(findTextLabel)))
	drawHistoryFieldRow(screen, maxX, maxY, x+2, y+findRowText, findTextLabel, d.field, d.focusIndex == findFocusText, textInputX, arrowX, labelStyle, accelStyle)

	leftX := x + findLeftBoxX
	rightX := x + findLeftBoxX + findLeftBoxW + findGap
	rightW := width - 2 - findLeftBoxX - findLeftBoxW - findGap

	drawGroupBox(screen, maxX, maxY, leftX, y+findRowBoxTop1, findLeftBoxW, findBoxH1, "Options", borderStyle, labelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+findRowBoxTop1+1, "&Case sensitive", d.caseSensitive, d.focusIndex == findFocusCase, bgStyle, focusStyle, accelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+findRowBoxTop1+2, "&Whole words only", d.wholeWords, d.focusIndex == findFocusWholeWords, bgStyle, focusStyle, accelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+findRowBoxTop1+3, "&Regular expression", d.regex, d.focusIndex == findFocusRegex, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, rightX, y+findRowBoxTop1, rightW, findBoxH1, "Direction", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+findRowBoxTop1+1, "Forwar&d", d.directionIdx == 0, d.focusIndex == findFocusForward, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+findRowBoxTop1+2, "&Backward", d.directionIdx == 1, d.focusIndex == findFocusBackward, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, leftX, y+findRowBoxTop2, findLeftBoxW, findBoxH2, "Scope", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, leftX+2, y+findRowBoxTop2+1, "&Global", d.scopeIdx == 0, d.focusIndex == findFocusGlobal, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, leftX+2, y+findRowBoxTop2+2, "&Selected text", d.scopeIdx == 1, d.focusIndex == findFocusSelected, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, rightX, y+findRowBoxTop2, rightW, findBoxH2, "Origin", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+findRowBoxTop2+1, "&From cursor", d.originIdx == 0, d.focusIndex == findFocusFromCursor, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+findRowBoxTop2+2, "&Entire scope", d.originIdx == 1, d.focusIndex == findFocusEntireScop, bgStyle, focusStyle, accelStyle)

	d.drawButtons(screen, maxX, maxY, x, y, width, height)

	if d.field.showHistory {
		dY := y + findRowText + 1
		dX := x + 2 + len([]rune(stripAcc(findTextLabel)))
		dW := arrowX - dX - 2
		drawFieldHistoryDropdown(screen, maxX, maxY, dX, dY, dW, d.field)
	}
}

// stripAcc é um atalho para obter o texto puro (sem "&") de um rótulo.
func stripAcc(label string) string {
	plain, _ := stripAccelerator(label)
	return plain
}

func (d *findDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	btnY := y + height - 3
	okW := d.okButton.width()
	cancelW := d.cancelButton.width()
	helpW := d.helpButton.width()
	total := okW + cancelW + helpW + 4
	btnX := x + (width-total)/2

	draw := func(btn *turboButton, focused bool, atX int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
	}

	draw(d.okButton, d.focusIndex == findFocusOK, btnX)
	draw(d.cancelButton, d.focusIndex == findFocusCancel, btnX+okW+2)
	draw(d.helpButton, d.focusIndex == findFocusHelp, btnX+okW+2+cancelW+2)
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *findDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}

		if d.focusIndex == findFocusText {
			d.handleTextFocusKey(event)
			return
		}

		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			return
		case tcell.KeyTab, tcell.KeyDown:
			d.focusIndex = (d.focusIndex + 1) % (findFocusMax + 1)
			return
		case tcell.KeyBacktab, tcell.KeyUp:
			d.focusIndex = (d.focusIndex - 1 + findFocusMax + 1) % (findFocusMax + 1)
			return
		case tcell.KeyLeft:
			if d.focusIndex >= findFocusOK && d.focusIndex <= findFocusHelp {
				d.focusIndex--
				if d.focusIndex < findFocusOK {
					d.focusIndex = findFocusHelp
				}
			}
			return
		case tcell.KeyRight:
			if d.focusIndex >= findFocusOK && d.focusIndex <= findFocusHelp {
				d.focusIndex++
				if d.focusIndex > findFocusHelp {
					d.focusIndex = findFocusOK
				}
			}
			return
		case tcell.KeyEnter:
			d.activateFocused()
			return
		case tcell.KeyRune:
			switch unicode.ToLower(event.Rune()) {
			case ' ':
				d.toggleFocused()
			case 't':
				d.focusIndex = findFocusText
			case 'c':
				d.focusIndex = findFocusCase
				d.caseSensitive = !d.caseSensitive
			case 'w':
				d.focusIndex = findFocusWholeWords
				d.wholeWords = !d.wholeWords
			case 'r':
				d.focusIndex = findFocusRegex
				d.regex = !d.regex
			case 'd':
				d.focusIndex = findFocusForward
				d.directionIdx = 0
			case 'b':
				d.focusIndex = findFocusBackward
				d.directionIdx = 1
			case 'g':
				d.focusIndex = findFocusGlobal
				d.scopeIdx = 0
			case 's':
				d.focusIndex = findFocusSelected
				d.scopeIdx = 1
			case 'f':
				d.focusIndex = findFocusFromCursor
				d.originIdx = 0
			case 'e':
				d.focusIndex = findFocusEntireScop
				d.originIdx = 1
			case 'o':
				d.commit()
			}
			return
		}
	})
}

// handleTextFocusKey trata teclas enquanto o campo "Text to find" está com foco:
// primeiro as ações do próprio diálogo (Escape/Tab/Backtab/Enter), depois
// delega ao historyField.
func (d *findDialog) handleTextFocusKey(event *tcell.EventKey) {
	if d.field.showHistory {
		d.field.handleKey(event)
		return
	}
	switch event.Key() {
	case tcell.KeyEscape:
		d.close()
		return
	case tcell.KeyTab:
		d.focusIndex = findFocusCase
		return
	case tcell.KeyBacktab:
		d.focusIndex = findFocusHelp
		return
	case tcell.KeyEnter:
		d.commit()
		return
	}
	d.field.handleKey(event)
}

func (d *findDialog) activateFocused() {
	switch {
	case d.focusIndex >= findFocusCase && d.focusIndex <= findFocusRegex:
		d.toggleFocused()
	case d.focusIndex == findFocusForward:
		d.directionIdx = 0
	case d.focusIndex == findFocusBackward:
		d.directionIdx = 1
	case d.focusIndex == findFocusGlobal:
		d.scopeIdx = 0
	case d.focusIndex == findFocusSelected:
		d.scopeIdx = 1
	case d.focusIndex == findFocusFromCursor:
		d.originIdx = 0
	case d.focusIndex == findFocusEntireScop:
		d.originIdx = 1
	case d.focusIndex == findFocusOK:
		d.commit()
	case d.focusIndex == findFocusCancel:
		d.close()
	case d.focusIndex == findFocusHelp:
		d.showHelp()
	}
}

func (d *findDialog) toggleFocused() {
	switch d.focusIndex {
	case findFocusCase:
		d.caseSensitive = !d.caseSensitive
	case findFocusWholeWords:
		d.wholeWords = !d.wholeWords
	case findFocusRegex:
		d.regex = !d.regex
	case findFocusForward:
		d.directionIdx = 0
	case findFocusBackward:
		d.directionIdx = 1
	case findFocusGlobal:
		d.scopeIdx = 0
	case findFocusSelected:
		d.scopeIdx = 1
	case findFocusFromCursor:
		d.originIdx = 0
	case findFocusEntireScop:
		d.originIdx = 1
	}
}

func (d *findDialog) showHelp() {
	text := []string{
		"Find",
		"",
		"Type the text to search for and press",
		"Ok (or Enter) to confirm.",
		"",
		"Options, Direction, Scope and Origin",
		"control how the search is performed.",
	}
	dlg := newDialogoOK(d.app, "find_help", text, nil)
	dlg.SetButton("O&K", 'k', nil)
	dlg.SetButtonShadowMode(shadowModeTurboClassic)
	showDialogoOKCentered(dlg, 46, 13)
}

func (d *findDialog) commit() {
	d.field.addHistory(d.field.value())
	params := d.currentParams()
	if d.onFind != nil {
		d.onFind(params)
	}
	d.close()
}

func (d *findDialog) close() {
	d.app.Pages.RemovePage(d.pageName)
	if d.onClose != nil {
		d.onClose()
		return
	}
	d.app.focusActiveEditor()
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *findDialog) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := d.GetRect()
		if mx < x || mx >= x+width || my < y || my >= y+height {
			return false, nil
		}
		if action != tview.MouseLeftClick {
			return true, nil
		}
		setFocus(d)

		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		row := y + findRowText
		inputX := x + 2 + len([]rune(stripAcc(findTextLabel)))
		arrowX := x + findDlgW - 3
		inputEndX := arrowX - 2

		if my == row && mx == arrowX {
			d.field.showHistory = !d.field.showHistory
			d.field.historyIdx = 0
			return true, nil
		}
		if d.field.clickInInput(mx, my, row, inputX, inputEndX) {
			d.focusIndex = findFocusText
			return true, nil
		}

		leftX := x + findLeftBoxX
		rightX := x + findLeftBoxX + findLeftBoxW + findGap

		clickIn := func(bx, by, bw, bh int) bool {
			return mx >= bx && mx < bx+bw && my >= by && my < by+bh
		}

		if clickIn(leftX+2, y+findRowBoxTop1+1, findLeftBoxW-4, 1) {
			d.focusIndex = findFocusCase
			d.caseSensitive = !d.caseSensitive
			return true, nil
		}
		if clickIn(leftX+2, y+findRowBoxTop1+2, findLeftBoxW-4, 1) {
			d.focusIndex = findFocusWholeWords
			d.wholeWords = !d.wholeWords
			return true, nil
		}
		if clickIn(leftX+2, y+findRowBoxTop1+3, findLeftBoxW-4, 1) {
			d.focusIndex = findFocusRegex
			d.regex = !d.regex
			return true, nil
		}

		rightW := width - 2 - findLeftBoxX - findLeftBoxW - findGap
		if clickIn(rightX+2, y+findRowBoxTop1+1, rightW-4, 1) {
			d.focusIndex = findFocusForward
			d.directionIdx = 0
			return true, nil
		}
		if clickIn(rightX+2, y+findRowBoxTop1+2, rightW-4, 1) {
			d.focusIndex = findFocusBackward
			d.directionIdx = 1
			return true, nil
		}

		if clickIn(leftX+2, y+findRowBoxTop2+1, findLeftBoxW-4, 1) {
			d.focusIndex = findFocusGlobal
			d.scopeIdx = 0
			return true, nil
		}
		if clickIn(leftX+2, y+findRowBoxTop2+2, findLeftBoxW-4, 1) {
			d.focusIndex = findFocusSelected
			d.scopeIdx = 1
			return true, nil
		}
		if clickIn(rightX+2, y+findRowBoxTop2+1, rightW-4, 1) {
			d.focusIndex = findFocusFromCursor
			d.originIdx = 0
			return true, nil
		}
		if clickIn(rightX+2, y+findRowBoxTop2+2, rightW-4, 1) {
			d.focusIndex = findFocusEntireScop
			d.originIdx = 1
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
				d.focusIndex = findFocusOK
				d.activateFocused()
			case mx >= btnX+okW+2 && mx < btnX+okW+2+cancelW:
				d.focusIndex = findFocusCancel
				d.activateFocused()
			case mx >= btnX+okW+2+cancelW+2 && mx < btnX+okW+2+cancelW+2+helpW:
				d.focusIndex = findFocusHelp
				d.activateFocused()
			}
			return true, nil
		}

		return true, nil
	}
}

func showFindDialogCentered(dialog *findDialog) {
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dialog, findDlgW, 0, true).
			AddItem(nil, 0, 1, false), findDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	dialog.app.Pages.AddPage(dialog.pageName, container, true, true)
	dialog.app.Application.SetFocus(dialog)
}
