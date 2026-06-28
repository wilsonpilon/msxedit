package tui

import "github.com/gdamore/tcell/v2"

// helpHeaderButton desenha um botao visual (desativado) para cabecalhos em sub-janelas de Help.
type helpHeaderButton struct {
	label       string
	fg          tcell.Color
	bg          tcell.Color
	shadowColor tcell.Color
}

func newHelpHeaderButton(label string, fg, bg, shadowColor tcell.Color) *helpHeaderButton {
	return &helpHeaderButton{
		label:       label,
		fg:          fg,
		bg:          bg,
		shadowColor: shadowColor,
	}
}

func (b *helpHeaderButton) width() int {
	return len([]rune(" "+b.label+" ")) + 2
}

func (b *helpHeaderButton) draw(screen tcell.Screen, maxX, maxY, x, y int, baseBg tcell.Color) {
	w := b.width()
	buttonStyle := tcell.StyleDefault.Foreground(b.fg).Background(b.bg)
	for col := 0; col < w; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', buttonStyle)
	}

	runes := []rune(" " + b.label + " ")
	for i, r := range runes {
		putRune(screen, maxX, maxY, x+1+i, y, r, buttonStyle)
	}

	// Sombra 3D no estilo Turbo classic.
	shadowStyle := tcell.StyleDefault.Foreground(b.shadowColor).Background(baseBg)
	putRune(screen, maxX, maxY, x+w, y, '▄', shadowStyle)
	for col := 0; col < w; col++ {
		putRune(screen, maxX, maxY, x+col+1, y+1, '▀', shadowStyle)
	}
}

