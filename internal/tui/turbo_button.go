package tui

import (
	"github.com/gdamore/tcell/v2"
)

type turboButtonShadowMode int

const (
	shadowModeTurboClassic turboButtonShadowMode = iota
	shadowModeFlat
)

type turboButton struct {
	label       string
	fg          tcell.Color
	bg          tcell.Color
	hotkeyColor tcell.Color
	shadowColor tcell.Color
	shadowMode  turboButtonShadowMode
}

func newTurboButton(label string, fg, bg, hotkeyColor, shadowColor tcell.Color) *turboButton {
	return &turboButton{
		label:       label,
		fg:          fg,
		bg:          bg,
		hotkeyColor: hotkeyColor,
		shadowColor: shadowColor,
		shadowMode:  shadowModeTurboClassic,
	}
}

func (b *turboButton) SetShadowMode(mode turboButtonShadowMode) *turboButton {
	b.shadowMode = mode
	return b
}

func (b *turboButton) width() int {
	plain, _ := stripAccelerator(b.label)
	return len([]rune(" "+plain+" ")) + 2
}

func (b *turboButton) draw(screen tcell.Screen, maxX, maxY, x, y int, baseBg tcell.Color) {
	w := b.width()
	buttonStyle := tcell.StyleDefault.Foreground(b.fg).Background(b.bg)
	for col := 0; col < w; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', buttonStyle)
	}

	plain, accel := stripAccelerator(b.label)
	runes := []rune(" " + plain + " ")
	for i, r := range runes {
		style := buttonStyle
		if i == accel+1 {
			style = tcell.StyleDefault.Foreground(b.hotkeyColor).Background(b.bg)
		}
		putRune(screen, maxX, maxY, x+1+i, y, r, style)
	}

	shadowStyle := tcell.StyleDefault.Foreground(b.shadowColor).Background(baseBg)
	switch b.shadowMode {
	case shadowModeFlat:
		for col := 0; col < w; col++ {
			putRune(screen, maxX, maxY, x+col, y+1, '▄', shadowStyle)
		}
	default:
		// Turbo classic: [BOTAO]▄ /  ▀▀▀▀
		putRune(screen, maxX, maxY, x+w, y, '▄', shadowStyle)
		for col := 0; col < w; col++ {
			putRune(screen, maxX, maxY, x+col+1, y+1, '▀', shadowStyle)
		}
	}
}
