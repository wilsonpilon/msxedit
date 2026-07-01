package tui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ReplaceParams reúne as opções escolhidas no diálogo Replace (Search > Replace...).
type ReplaceParams struct {
	FindParams
	NewText         string
	PromptOnReplace bool
	ReplaceAll      bool // true quando confirmado via "Change all", false via "Ok"
}

// ── Layout (largura 62, altura 21) ──────────────────────────────────────────
//
//  y+0   ╔═════════════════════════ Replace ══════════════════════════╗
//  y+1   ║                                                            ║
//  y+2   ║ Text to find [___________________________________]  ↓     ║
//  y+3   ║                                                            ║
//  y+4   ║ New text     [___________________________________]  ↓     ║
//  y+5   ║                                                            ║
//  y+6   ║ ┌─Options───────────────────┐ ┌─Direction────────────────┐ ║
//  y+7   ║ │ [ ] Case sensitive         │ │ ( ) Forward               │ ║
//  y+8   ║ │ [ ] Whole words only       │ │ ( ) Backward              │ ║
//  y+9   ║ │ [ ] Regular expression     │ │                           │ ║
//  y+10  ║ │ [ ] Prompt on replace      │ │                           │ ║
//  y+11  ║ └────────────────────────────┘ └───────────────────────────┘ ║
//  y+12  ║                                                            ║
//  y+13  ║ ┌─Scope─────────────────────┐ ┌─Origin───────────────────┐ ║
//  y+14  ║ │ ( ) Global                  │ │ ( ) From cursor           │ ║
//  y+15  ║ │ ( ) Selected text           │ │ ( ) Entire scope          │ ║
//  y+16  ║ └────────────────────────────┘ └───────────────────────────┘ ║
//  y+17  ║                                                            ║
//  y+18  ║      [ &Ok ]  [Change &all]  [Cancel]  [Help]              ║
//  y+19  ║       (shadow)  (shadow)      (shadow)   (shadow)           ║
//  y+20  ╚════════════════════════════════════════════════════════════╝

const (
	replaceDlgW = 62
	replaceDlgH = 21

	replaceRowText1 = 2
	replaceRowText2 = 4
	replaceRowBox1  = 6
	replaceRowBox2  = 13

	replaceLeftBoxX = 2
	replaceLeftBoxW = 30
	replaceGap      = 2
	replaceBoxH1    = 6 // Options (4 itens) / Direction (2 itens + 2 linhas em branco)
	replaceBoxH2    = 4 // Scope (2 itens) / Origin (2 itens)

	replaceFindLabel = "&Text to find "
	replaceNewLabel  = "&New text "
)

// Foco: 0=find, 1=new text, 2-5=Options, 6-7=Direction, 8-9=Scope, 10-11=Origin,
// 12=Ok, 13=Change all, 14=Cancel, 15=Help
const (
	replaceFocusFindText   = 0
	replaceFocusNewText    = 1
	replaceFocusCase       = 2
	replaceFocusWholeWords = 3
	replaceFocusRegex      = 4
	replaceFocusPrompt     = 5
	replaceFocusForward    = 6
	replaceFocusBackward   = 7
	replaceFocusGlobal     = 8
	replaceFocusSelected   = 9
	replaceFocusFromCursor = 10
	replaceFocusEntireScop = 11
	replaceFocusOK         = 12
	replaceFocusChangeAll  = 13
	replaceFocusCancel     = 14
	replaceFocusHelp       = 15
	replaceFocusMax        = 15
)

type replaceDialog struct {
	*tview.Box
	app      *App
	pageName string

	findField *historyField
	newField  *historyField

	caseSensitive   bool
	wholeWords      bool
	regex           bool
	promptOnReplace bool

	directionIdx int // 0=Forward 1=Backward
	scopeIdx     int // 0=Global  1=Selected text
	originIdx    int // 0=From cursor 1=Entire scope

	focusIndex int

	okButton        *turboButton
	changeAllButton *turboButton
	cancelButton    *turboButton
	helpButton      *turboButton

	onReplace func(params ReplaceParams)
	onClose   func()
}

func newReplaceDialog(app *App, params ReplaceParams, findHistory, newHistory []string) *replaceDialog {
	d := &replaceDialog{
		Box:             tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:             app,
		pageName:        "replace_dialog",
		findField:       newHistoryField(params.Text, findHistory),
		newField:        newHistoryField(params.NewText, newHistory),
		caseSensitive:   params.CaseSensitive,
		wholeWords:      params.WholeWords,
		regex:           params.Regex,
		promptOnReplace: params.PromptOnReplace,
		focusIndex:      replaceFocusFindText,
		okButton:        newTurboButton("&Ok", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		changeAllButton: newTurboButton("Change &all", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		cancelButton:    newTurboButton("Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		helpButton:      newTurboButton("Help", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
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

func (d *replaceDialog) currentParams(all bool) ReplaceParams {
	return ReplaceParams{
		FindParams: FindParams{
			Text:          d.findField.value(),
			CaseSensitive: d.caseSensitive,
			WholeWords:    d.wholeWords,
			Regex:         d.regex,
			Backward:      d.directionIdx == 1,
			SelectedOnly:  d.scopeIdx == 1,
			EntireScope:   d.originIdx == 1,
		},
		NewText:         d.newField.value(),
		PromptOnReplace: d.promptOnReplace,
		ReplaceAll:      all,
	}
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (d *replaceDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < replaceDlgW || height < replaceDlgH {
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

	title := " Replace "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	arrowX := x + replaceDlgW - 3
	textInputX := x + 2 + len([]rune(stripAcc(replaceFindLabel)))
	drawHistoryFieldRow(screen, maxX, maxY, x+2, y+replaceRowText1, replaceFindLabel, d.findField, d.focusIndex == replaceFocusFindText, textInputX, arrowX, labelStyle, accelStyle)
	drawHistoryFieldRow(screen, maxX, maxY, x+2, y+replaceRowText2, replaceNewLabel, d.newField, d.focusIndex == replaceFocusNewText, textInputX, arrowX, labelStyle, accelStyle)

	leftX := x + replaceLeftBoxX
	rightX := x + replaceLeftBoxX + replaceLeftBoxW + replaceGap
	rightW := width - 2 - replaceLeftBoxX - replaceLeftBoxW - replaceGap

	drawGroupBox(screen, maxX, maxY, leftX, y+replaceRowBox1, replaceLeftBoxW, replaceBoxH1, "Options", borderStyle, labelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+replaceRowBox1+1, "&Case sensitive", d.caseSensitive, d.focusIndex == replaceFocusCase, bgStyle, focusStyle, accelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+replaceRowBox1+2, "&Whole words only", d.wholeWords, d.focusIndex == replaceFocusWholeWords, bgStyle, focusStyle, accelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+replaceRowBox1+3, "&Regular expression", d.regex, d.focusIndex == replaceFocusRegex, bgStyle, focusStyle, accelStyle)
	drawCheckbox(screen, maxX, maxY, leftX+2, y+replaceRowBox1+4, "&Prompt on replace", d.promptOnReplace, d.focusIndex == replaceFocusPrompt, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, rightX, y+replaceRowBox1, rightW, replaceBoxH1, "Direction", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+replaceRowBox1+1, "Forwar&d", d.directionIdx == 0, d.focusIndex == replaceFocusForward, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+replaceRowBox1+2, "&Backward", d.directionIdx == 1, d.focusIndex == replaceFocusBackward, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, leftX, y+replaceRowBox2, replaceLeftBoxW, replaceBoxH2, "Scope", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, leftX+2, y+replaceRowBox2+1, "&Global", d.scopeIdx == 0, d.focusIndex == replaceFocusGlobal, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, leftX+2, y+replaceRowBox2+2, "&Selected text", d.scopeIdx == 1, d.focusIndex == replaceFocusSelected, bgStyle, focusStyle, accelStyle)

	drawGroupBox(screen, maxX, maxY, rightX, y+replaceRowBox2, rightW, replaceBoxH2, "Origin", borderStyle, labelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+replaceRowBox2+1, "&From cursor", d.originIdx == 0, d.focusIndex == replaceFocusFromCursor, bgStyle, focusStyle, accelStyle)
	drawRadio(screen, maxX, maxY, rightX+2, y+replaceRowBox2+2, "&Entire scope", d.originIdx == 1, d.focusIndex == replaceFocusEntireScop, bgStyle, focusStyle, accelStyle)

	d.drawButtons(screen, maxX, maxY, x, y, width, height)

	if d.findField.showHistory {
		dY := y + replaceRowText1 + 1
		dW := arrowX - textInputX - 2
		drawFieldHistoryDropdown(screen, maxX, maxY, textInputX, dY, dW, d.findField)
	}
	if d.newField.showHistory {
		dY := y + replaceRowText2 + 1
		dW := arrowX - textInputX - 2
		drawFieldHistoryDropdown(screen, maxX, maxY, textInputX, dY, dW, d.newField)
	}
}

func (d *replaceDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	btnY := y + height - 3
	okW := d.okButton.width()
	allW := d.changeAllButton.width()
	cancelW := d.cancelButton.width()
	helpW := d.helpButton.width()
	total := okW + allW + cancelW + helpW + 6
	btnX := x + (width-total)/2

	draw := func(btn *turboButton, focused bool, atX int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, atX, btnY, d.app.Theme.PopupBg)
	}

	draw(d.okButton, d.focusIndex == replaceFocusOK, btnX)
	draw(d.changeAllButton, d.focusIndex == replaceFocusChangeAll, btnX+okW+2)
	draw(d.cancelButton, d.focusIndex == replaceFocusCancel, btnX+okW+2+allW+2)
	draw(d.helpButton, d.focusIndex == replaceFocusHelp, btnX+okW+2+allW+2+cancelW+2)
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *replaceDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}

		if d.focusIndex == replaceFocusFindText || d.focusIndex == replaceFocusNewText {
			d.handleFieldFocusKey(event)
			return
		}

		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			return
		case tcell.KeyTab, tcell.KeyDown:
			d.focusIndex = (d.focusIndex + 1) % (replaceFocusMax + 1)
			return
		case tcell.KeyBacktab, tcell.KeyUp:
			d.focusIndex = (d.focusIndex - 1 + replaceFocusMax + 1) % (replaceFocusMax + 1)
			return
		case tcell.KeyLeft:
			if d.focusIndex >= replaceFocusOK && d.focusIndex <= replaceFocusHelp {
				d.focusIndex--
				if d.focusIndex < replaceFocusOK {
					d.focusIndex = replaceFocusHelp
				}
			}
			return
		case tcell.KeyRight:
			if d.focusIndex >= replaceFocusOK && d.focusIndex <= replaceFocusHelp {
				d.focusIndex++
				if d.focusIndex > replaceFocusHelp {
					d.focusIndex = replaceFocusOK
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
				d.focusIndex = replaceFocusFindText
			case 'n':
				d.focusIndex = replaceFocusNewText
			case 'c':
				d.focusIndex = replaceFocusCase
				d.caseSensitive = !d.caseSensitive
			case 'w':
				d.focusIndex = replaceFocusWholeWords
				d.wholeWords = !d.wholeWords
			case 'r':
				d.focusIndex = replaceFocusRegex
				d.regex = !d.regex
			case 'p':
				d.focusIndex = replaceFocusPrompt
				d.promptOnReplace = !d.promptOnReplace
			case 'd':
				d.focusIndex = replaceFocusForward
				d.directionIdx = 0
			case 'b':
				d.focusIndex = replaceFocusBackward
				d.directionIdx = 1
			case 'g':
				d.focusIndex = replaceFocusGlobal
				d.scopeIdx = 0
			case 's':
				d.focusIndex = replaceFocusSelected
				d.scopeIdx = 1
			case 'f':
				d.focusIndex = replaceFocusFromCursor
				d.originIdx = 0
			case 'e':
				d.focusIndex = replaceFocusEntireScop
				d.originIdx = 1
			case 'o':
				d.commit(false)
			case 'a':
				d.commit(true)
			}
			return
		}
	})
}

// handleFieldFocusKey trata teclas quando "Text to find" ou "New text" está
// com foco: primeiro as ações do diálogo (Escape/Tab/Backtab/Enter), depois
// delega ao historyField correspondente.
func (d *replaceDialog) handleFieldFocusKey(event *tcell.EventKey) {
	field := d.findField
	if d.focusIndex == replaceFocusNewText {
		field = d.newField
	}

	if field.showHistory {
		field.handleKey(event)
		return
	}
	switch event.Key() {
	case tcell.KeyEscape:
		d.close()
		return
	case tcell.KeyTab:
		if d.focusIndex == replaceFocusFindText {
			d.focusIndex = replaceFocusNewText
		} else {
			d.focusIndex = replaceFocusCase
		}
		return
	case tcell.KeyBacktab:
		if d.focusIndex == replaceFocusNewText {
			d.focusIndex = replaceFocusFindText
		} else {
			d.focusIndex = replaceFocusHelp
		}
		return
	case tcell.KeyEnter:
		d.commit(false)
		return
	}
	field.handleKey(event)
}

func (d *replaceDialog) activateFocused() {
	switch {
	case d.focusIndex >= replaceFocusCase && d.focusIndex <= replaceFocusPrompt:
		d.toggleFocused()
	case d.focusIndex == replaceFocusForward:
		d.directionIdx = 0
	case d.focusIndex == replaceFocusBackward:
		d.directionIdx = 1
	case d.focusIndex == replaceFocusGlobal:
		d.scopeIdx = 0
	case d.focusIndex == replaceFocusSelected:
		d.scopeIdx = 1
	case d.focusIndex == replaceFocusFromCursor:
		d.originIdx = 0
	case d.focusIndex == replaceFocusEntireScop:
		d.originIdx = 1
	case d.focusIndex == replaceFocusOK:
		d.commit(false)
	case d.focusIndex == replaceFocusChangeAll:
		d.commit(true)
	case d.focusIndex == replaceFocusCancel:
		d.close()
	case d.focusIndex == replaceFocusHelp:
		d.showHelp()
	}
}

func (d *replaceDialog) toggleFocused() {
	switch d.focusIndex {
	case replaceFocusCase:
		d.caseSensitive = !d.caseSensitive
	case replaceFocusWholeWords:
		d.wholeWords = !d.wholeWords
	case replaceFocusRegex:
		d.regex = !d.regex
	case replaceFocusPrompt:
		d.promptOnReplace = !d.promptOnReplace
	case replaceFocusForward:
		d.directionIdx = 0
	case replaceFocusBackward:
		d.directionIdx = 1
	case replaceFocusGlobal:
		d.scopeIdx = 0
	case replaceFocusSelected:
		d.scopeIdx = 1
	case replaceFocusFromCursor:
		d.originIdx = 0
	case replaceFocusEntireScop:
		d.originIdx = 1
	}
}

func (d *replaceDialog) showHelp() {
	text := []string{
		"Replace",
		"",
		"Type the text to search for and the",
		"replacement text, then press Ok to",
		"replace the next match or Change all",
		"to replace every match.",
	}
	dlg := newDialogoOK(d.app, "replace_help", text, nil)
	dlg.SetButton("O&K", 'k', nil)
	dlg.SetButtonShadowMode(shadowModeTurboClassic)
	showDialogoOKCentered(dlg, 46, 13)
}

func (d *replaceDialog) commit(all bool) {
	d.findField.addHistory(d.findField.value())
	d.newField.addHistory(d.newField.value())
	params := d.currentParams(all)
	if d.onReplace != nil {
		d.onReplace(params)
	}
	d.close()
}

func (d *replaceDialog) close() {
	d.app.Pages.RemovePage(d.pageName)
	if d.onClose != nil {
		d.onClose()
		return
	}
	d.app.focusActiveEditor()
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *replaceDialog) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
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

		arrowX := x + replaceDlgW - 3
		textInputX := x + 2 + len([]rune(stripAcc(replaceFindLabel)))
		inputEndX := arrowX - 2

		row1 := y + replaceRowText1
		row2 := y + replaceRowText2

		if my == row1 && mx == arrowX {
			d.findField.showHistory = !d.findField.showHistory
			d.findField.historyIdx = 0
			return true, nil
		}
		if d.findField.clickInInput(mx, my, row1, textInputX, inputEndX) {
			d.focusIndex = replaceFocusFindText
			return true, nil
		}
		if my == row2 && mx == arrowX {
			d.newField.showHistory = !d.newField.showHistory
			d.newField.historyIdx = 0
			return true, nil
		}
		if d.newField.clickInInput(mx, my, row2, textInputX, inputEndX) {
			d.focusIndex = replaceFocusNewText
			return true, nil
		}

		leftX := x + replaceLeftBoxX
		rightX := x + replaceLeftBoxX + replaceLeftBoxW + replaceGap

		clickIn := func(bx, by, bw, bh int) bool {
			return mx >= bx && mx < bx+bw && my >= by && my < by+bh
		}

		if clickIn(leftX+2, y+replaceRowBox1+1, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusCase
			d.caseSensitive = !d.caseSensitive
			return true, nil
		}
		if clickIn(leftX+2, y+replaceRowBox1+2, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusWholeWords
			d.wholeWords = !d.wholeWords
			return true, nil
		}
		if clickIn(leftX+2, y+replaceRowBox1+3, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusRegex
			d.regex = !d.regex
			return true, nil
		}
		if clickIn(leftX+2, y+replaceRowBox1+4, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusPrompt
			d.promptOnReplace = !d.promptOnReplace
			return true, nil
		}

		rightW := width - 2 - replaceLeftBoxX - replaceLeftBoxW - replaceGap
		if clickIn(rightX+2, y+replaceRowBox1+1, rightW-4, 1) {
			d.focusIndex = replaceFocusForward
			d.directionIdx = 0
			return true, nil
		}
		if clickIn(rightX+2, y+replaceRowBox1+2, rightW-4, 1) {
			d.focusIndex = replaceFocusBackward
			d.directionIdx = 1
			return true, nil
		}

		if clickIn(leftX+2, y+replaceRowBox2+1, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusGlobal
			d.scopeIdx = 0
			return true, nil
		}
		if clickIn(leftX+2, y+replaceRowBox2+2, replaceLeftBoxW-4, 1) {
			d.focusIndex = replaceFocusSelected
			d.scopeIdx = 1
			return true, nil
		}
		if clickIn(rightX+2, y+replaceRowBox2+1, rightW-4, 1) {
			d.focusIndex = replaceFocusFromCursor
			d.originIdx = 0
			return true, nil
		}
		if clickIn(rightX+2, y+replaceRowBox2+2, rightW-4, 1) {
			d.focusIndex = replaceFocusEntireScop
			d.originIdx = 1
			return true, nil
		}

		btnY := y + height - 3
		okW := d.okButton.width()
		allW := d.changeAllButton.width()
		cancelW := d.cancelButton.width()
		helpW := d.helpButton.width()
		total := okW + allW + cancelW + helpW + 6
		btnX := x + (width-total)/2

		if my == btnY {
			switch {
			case mx >= btnX && mx < btnX+okW:
				d.focusIndex = replaceFocusOK
				d.activateFocused()
			case mx >= btnX+okW+2 && mx < btnX+okW+2+allW:
				d.focusIndex = replaceFocusChangeAll
				d.activateFocused()
			case mx >= btnX+okW+2+allW+2 && mx < btnX+okW+2+allW+2+cancelW:
				d.focusIndex = replaceFocusCancel
				d.activateFocused()
			case mx >= btnX+okW+2+allW+2+cancelW+2 && mx < btnX+okW+2+allW+2+cancelW+2+helpW:
				d.focusIndex = replaceFocusHelp
				d.activateFocused()
			}
			return true, nil
		}

		return true, nil
	}
}

func showReplaceDialogCentered(dialog *replaceDialog) {
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dialog, replaceDlgW, 0, true).
			AddItem(nil, 0, 1, false), replaceDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	dialog.app.Pages.AddPage(dialog.pageName, container, true, true)
	dialog.app.Application.SetFocus(dialog)
}
