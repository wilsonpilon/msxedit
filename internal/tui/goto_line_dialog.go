package tui

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GotoLineParams reúne as opções escolhidas no diálogo Go to Line Number.
type GotoLineParams struct {
	Line     int
	MSXBasic bool // true = número de linha do MSX-Basic; false = linha do texto
}

// ── Layout (largura 48, altura 9) ───────────────────────────────────────────
//
//  y+0  ╔══════════ Go to Line Number ═══════════╗
//  y+1  ║                                        ║
//  y+2  ║ Enter new line number [_______]  ↓     ║
//  y+3  ║                                        ║
//  y+4  ║ [ ] Linha do MSX-Basic                 ║
//  y+5  ║                                        ║
//  y+6  ║        [ &Ok ]  [Cancel]  [Help]       ║
//  y+7  ║         (shadow)  (shadow)  (shadow)   ║
//  y+8  ╚════════════════════════════════════════╝

const (
	gotoDlgW = 48
	gotoDlgH = 9

	gotoRowInput    = 2
	gotoRowCheckbox = 4

	gotoLabel = "&Enter new line number "
)

const (
	gotoFocusInput    = 0
	gotoFocusCheckbox = 1
	gotoFocusOK       = 2
	gotoFocusCancel   = 3
	gotoFocusHelp     = 4
	gotoFocusMax      = 4
)

type gotoLineDialog struct {
	*tview.Box
	app      *App
	pageName string

	field    *historyField
	msxBasic bool

	focusIndex int

	okButton     *turboButton
	cancelButton *turboButton
	helpButton   *turboButton

	onGoto  func(params GotoLineParams)
	onClose func()
}

func newGotoLineDialog(app *App, initial GotoLineParams, history []string) *gotoLineDialog {
	initialText := ""
	if initial.Line > 0 {
		initialText = strconv.Itoa(initial.Line)
	}
	d := &gotoLineDialog{
		Box:          tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:          app,
		pageName:     "goto_line_dialog",
		field:        newHistoryField(initialText, history),
		msxBasic:     initial.MSXBasic,
		focusIndex:   gotoFocusInput,
		okButton:     newTurboButton("&Ok", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		cancelButton: newTurboButton("Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		helpButton:   newTurboButton("Help", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	return d
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (d *gotoLineDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < gotoDlgW || height < gotoDlgH {
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

	title := " Go to Line Number "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	arrowX := x + gotoDlgW - 3
	inputX := x + 2 + len([]rune(stripAcc(gotoLabel)))
	drawHistoryFieldRow(screen, maxX, maxY, x+2, y+gotoRowInput, gotoLabel, d.field, d.focusIndex == gotoFocusInput, inputX, arrowX, labelStyle, accelStyle)

	drawCheckbox(screen, maxX, maxY, x+2, y+gotoRowCheckbox, "&Linha do MSX-Basic", d.msxBasic, d.focusIndex == gotoFocusCheckbox, bgStyle, focusStyle, accelStyle)

	d.drawButtons(screen, maxX, maxY, x, y, width, height)

	if d.field.showHistory {
		dY := y + gotoRowInput + 1
		dW := arrowX - inputX - 2
		drawFieldHistoryDropdown(screen, maxX, maxY, inputX, dY, dW, d.field)
	}
}

func (d *gotoLineDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
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

	draw(d.okButton, d.focusIndex == gotoFocusOK, btnX)
	draw(d.cancelButton, d.focusIndex == gotoFocusCancel, btnX+okW+2)
	draw(d.helpButton, d.focusIndex == gotoFocusHelp, btnX+okW+2+cancelW+2)
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *gotoLineDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}

		if d.focusIndex == gotoFocusInput {
			d.handleFieldFocusKey(event)
			return
		}

		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			return
		case tcell.KeyTab, tcell.KeyDown:
			d.focusIndex = (d.focusIndex + 1) % (gotoFocusMax + 1)
			return
		case tcell.KeyBacktab, tcell.KeyUp:
			d.focusIndex = (d.focusIndex - 1 + gotoFocusMax + 1) % (gotoFocusMax + 1)
			return
		case tcell.KeyLeft:
			if d.focusIndex >= gotoFocusOK && d.focusIndex <= gotoFocusHelp {
				d.focusIndex--
				if d.focusIndex < gotoFocusOK {
					d.focusIndex = gotoFocusHelp
				}
			}
			return
		case tcell.KeyRight:
			if d.focusIndex >= gotoFocusOK && d.focusIndex <= gotoFocusHelp {
				d.focusIndex++
				if d.focusIndex > gotoFocusHelp {
					d.focusIndex = gotoFocusOK
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
			case 'e':
				d.focusIndex = gotoFocusInput
			case 'l':
				d.focusIndex = gotoFocusCheckbox
				d.msxBasic = !d.msxBasic
			case 'o':
				d.commit()
			}
			return
		}
	})
}

// handleFieldFocusKey trata teclas quando o campo de número está com foco:
// primeiro as ações do diálogo (Escape/Tab/Backtab/Enter), depois delega ao
// historyField — mas só aceita dígitos (é um campo numérico).
func (d *gotoLineDialog) handleFieldFocusKey(event *tcell.EventKey) {
	if d.field.showHistory {
		d.field.handleKey(event)
		return
	}
	switch event.Key() {
	case tcell.KeyEscape:
		d.close()
		return
	case tcell.KeyTab:
		d.focusIndex = gotoFocusCheckbox
		return
	case tcell.KeyBacktab:
		d.focusIndex = gotoFocusHelp
		return
	case tcell.KeyEnter:
		d.commit()
		return
	case tcell.KeyRune:
		if !unicode.IsDigit(event.Rune()) {
			return // campo numérico: ignora não-dígitos
		}
	}
	d.field.handleKey(event)
}

func (d *gotoLineDialog) activateFocused() {
	switch d.focusIndex {
	case gotoFocusCheckbox:
		d.toggleFocused()
	case gotoFocusOK:
		d.commit()
	case gotoFocusCancel:
		d.close()
	case gotoFocusHelp:
		d.showHelp()
	}
}

func (d *gotoLineDialog) toggleFocused() {
	if d.focusIndex == gotoFocusCheckbox {
		d.msxBasic = !d.msxBasic
	}
}

func (d *gotoLineDialog) showHelp() {
	text := []string{
		"Go to Line Number",
		"",
		"Type a line number and press Ok.",
		"",
		"When \"Linha do MSX-Basic\" is checked,",
		"the number is treated as a MSX-Basic",
		"line number instead of a text line.",
	}
	dlg := newDialogoOK(d.app, "goto_line_help", text, nil)
	dlg.SetButton("O&K", 'k', nil)
	dlg.SetButtonShadowMode(shadowModeTurboClassic)
	showDialogoOKCentered(dlg, 46, 13)
}

// commit valida o número digitado e, se válido, dispara onGoto e fecha o
// diálogo. Com valor inválido/vazio, o diálogo permanece aberto.
func (d *gotoLineDialog) commit() {
	text := strings.TrimSpace(d.field.value())
	n, err := strconv.Atoi(text)
	if err != nil || n <= 0 {
		return
	}
	d.field.addHistory(text)
	if d.onGoto != nil {
		d.onGoto(GotoLineParams{Line: n, MSXBasic: d.msxBasic})
	}
	d.close()
}

func (d *gotoLineDialog) close() {
	d.app.Pages.RemovePage(d.pageName)
	if d.onClose != nil {
		d.onClose()
		return
	}
	d.app.focusActiveEditor()
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *gotoLineDialog) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
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

		row := y + gotoRowInput
		arrowX := x + gotoDlgW - 3
		inputX := x + 2 + len([]rune(stripAcc(gotoLabel)))
		inputEndX := arrowX - 2

		if my == row && mx == arrowX {
			d.field.showHistory = !d.field.showHistory
			d.field.historyIdx = 0
			return true, nil
		}
		if d.field.clickInInput(mx, my, row, inputX, inputEndX) {
			d.focusIndex = gotoFocusInput
			return true, nil
		}

		if my == y+gotoRowCheckbox && mx >= x+2 && mx < x+width-2 {
			d.focusIndex = gotoFocusCheckbox
			d.msxBasic = !d.msxBasic
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
				d.focusIndex = gotoFocusOK
				d.activateFocused()
			case mx >= btnX+okW+2 && mx < btnX+okW+2+cancelW:
				d.focusIndex = gotoFocusCancel
				d.activateFocused()
			case mx >= btnX+okW+2+cancelW+2 && mx < btnX+okW+2+cancelW+2+helpW:
				d.focusIndex = gotoFocusHelp
				d.activateFocused()
			}
			return true, nil
		}

		return true, nil
	}
}

func showGotoLineDialogCentered(dialog *gotoLineDialog) {
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dialog, gotoDlgW, 0, true).
			AddItem(nil, 0, 1, false), gotoDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	dialog.app.Pages.AddPage(dialog.pageName, container, true, true)
	dialog.app.Application.SetFocus(dialog)
}
