package tui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type dialogoOK struct {
	*tview.Box
	app          *App
	pageName     string
	text         []string
	okButton     *turboButton
	buttonHotkey rune
	buttonAction func()
	onClose      func()
}

func newDialogoOK(app *App, pageName string, text []string, onClose func()) *dialogoOK {
	return &dialogoOK{
		Box:          tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:          app,
		pageName:     pageName,
		text:         text,
		okButton:     newTurboButton("O&K", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		buttonHotkey: 'k',
		onClose:      onClose,
	}
}

// SetButton configura o botao principal do Dialogo OK com label, hotkey e callback.
// Se callback for nil, a acao padrao e fechar o dialogo.
func (d *dialogoOK) SetButton(label string, hotkey rune, callback func()) *dialogoOK {
	d.okButton = newTurboButton(label, vgaWhite, vgaGreen, vgaYellow, vgaBlack)
	if hotkey == 0 {
		plain, accel := stripAccelerator(label)
		runes := []rune(plain)
		if accel >= 0 && accel < len(runes) {
			hotkey = runes[accel]
		}
	}
	d.buttonHotkey = unicode.ToLower(hotkey)
	d.buttonAction = callback
	return d
}

func (d *dialogoOK) SetButtonShadowMode(mode turboButtonShadowMode) *dialogoOK {
	d.okButton.SetShadowMode(mode)
	return d
}

func (d *dialogoOK) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < 20 || height < 8 {
		return
	}

	maxX, maxY := screen.Size()
	bgStyle := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)

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

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	contentTop := y + 2
	buttonY := y + height - 3
	maxTextY := buttonY - 2 // uma linha de folga entre texto e botao
	for i, line := range d.text {
		lineY := contentTop + i
		if lineY > maxTextY {
			break
		}
		lineWidth := len([]rune(line))
		lineX := x + (width-lineWidth)/2
		putString(screen, maxX, maxY, lineX, lineY, line, bgStyle)
	}

	d.drawOKButton(screen, maxX, maxY, x, width, buttonY)
}

func (d *dialogoOK) drawOKButton(screen tcell.Screen, maxX, maxY, x, width, btnY int) {
	btnWidth := d.okButton.width()
	btnX := x + (width-btnWidth)/2
	d.okButton.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
}

func (d *dialogoOK) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}
		if event.Key() == tcell.KeyEscape {
			d.close()
			return
		}
		if event.Key() == tcell.KeyEnter {
			d.activateButton()
			return
		}
		if event.Key() == tcell.KeyRune {
			if unicode.ToLower(event.Rune()) == d.buttonHotkey {
				d.activateButton()
				return
			}
		}
	})
}

func (d *dialogoOK) activateButton() {
	if d.buttonAction != nil {
		d.buttonAction()
		return
	}
	d.close()
}

func (d *dialogoOK) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := d.GetRect()

		// Fora dos limites
		if mx < x || mx >= x+width || my < y || my >= y+height {
			return false, nil
		}

		if action == tview.MouseLeftClick {
			setFocus(d)
			if my == y && mx >= x+2 && mx <= x+4 {
				d.close()
				return true, nil
			}
			btnWidth := d.okButton.width()
			btnX := x + (width-btnWidth)/2
			btnY := y + height - 3
			// +1 because the button label starts one col to the right of btnX
			if my == btnY && mx >= btnX && mx < btnX+btnWidth {
				d.activateButton()
			}
			return true, nil
		}
		return false, nil
	}
}

func (d *dialogoOK) close() {
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

// showDialogoOKCentered mostra um Dialogo OK centralizado na tela inteira.
func showDialogoOKCentered(dialog *dialogoOK, width, height int) {
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
