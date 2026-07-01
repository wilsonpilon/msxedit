package tui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// historyField é um campo de texto de uma linha com histórico (seta ↓),
// compartilhado pelos diálogos Find e Replace.
type historyField struct {
	text   []rune
	cursor int

	history     []string
	showHistory bool
	historyIdx  int
}

func newHistoryField(initial string, history []string) *historyField {
	f := &historyField{text: []rune(initial), history: history}
	f.cursor = len(f.text)
	return f
}

func (f *historyField) value() string {
	return string(f.text)
}

// handleKey trata teclas quando o campo está com foco. Retorna false para
// teclas que o campo não trata (o chamador decide o que fazer, ex.: Escape
// fecha o diálogo, Tab/Backtab movem o foco).
func (f *historyField) handleKey(event *tcell.EventKey) bool {
	if f.showHistory {
		f.handleHistoryKey(event)
		return true
	}
	switch event.Key() {
	case tcell.KeyLeft:
		if f.cursor > 0 {
			f.cursor--
		}
	case tcell.KeyRight:
		if f.cursor < len(f.text) {
			f.cursor++
		}
	case tcell.KeyHome:
		f.cursor = 0
	case tcell.KeyEnd:
		f.cursor = len(f.text)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if f.cursor > 0 {
			f.text = append(f.text[:f.cursor-1], f.text[f.cursor:]...)
			f.cursor--
		}
	case tcell.KeyDelete:
		if f.cursor < len(f.text) {
			f.text = append(f.text[:f.cursor], f.text[f.cursor+1:]...)
		}
	case tcell.KeyDown:
		if len(f.history) > 0 {
			f.showHistory = true
			f.historyIdx = 0
		}
	case tcell.KeyRune:
		r := event.Rune()
		if unicode.IsPrint(r) {
			tail := append([]rune{r}, f.text[f.cursor:]...)
			f.text = append(f.text[:f.cursor], tail...)
			f.cursor++
		}
	default:
		return false
	}
	return true
}

func (f *historyField) handleHistoryKey(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyEscape:
		f.showHistory = false
	case tcell.KeyUp:
		if f.historyIdx > 0 {
			f.historyIdx--
		}
	case tcell.KeyDown:
		if f.historyIdx < len(f.history)-1 {
			f.historyIdx++
		}
	case tcell.KeyEnter:
		if f.historyIdx >= 0 && f.historyIdx < len(f.history) {
			f.text = []rune(f.history[f.historyIdx])
			f.cursor = len(f.text)
		}
		f.showHistory = false
	}
}

func (f *historyField) addHistory(text string) {
	if text == "" {
		return
	}
	for i, h := range f.history {
		if h == text {
			f.history = append(f.history[:i], f.history[i+1:]...)
			break
		}
	}
	f.history = append([]string{text}, f.history...)
	if len(f.history) > 16 {
		f.history = f.history[:16]
	}
}

// clickInInput informa se (mx,my) cai dentro do campo de entrada desenhado
// por drawHistoryFieldRow, atualizando o cursor de acordo com a posição
// clicada. Retorna false se o clique não foi nesse campo.
func (f *historyField) clickInInput(mx, my, row, inputX, inputEndX int) bool {
	if my != row || mx < inputX || mx > inputEndX {
		return false
	}
	inputW := inputEndX - inputX + 1
	offset := 0
	if f.cursor >= inputW {
		offset = f.cursor - inputW + 1
	}
	pos := (mx - inputX) + offset
	if pos > len(f.text) {
		pos = len(f.text)
	}
	f.cursor = pos
	return true
}

// drawHistoryFieldRow desenha "label" seguido do campo azul de edição e da
// seta ↓ de histórico, tudo na mesma linha. labelX é onde o rótulo começa,
// inputX é onde o campo azul começa (permite alinhar vários campos com
// rótulos de tamanhos diferentes) e arrowX é a coluna absoluta da seta.
// Retorna inputEndX para uso no mouse handler do diálogo.
func drawHistoryFieldRow(screen tcell.Screen, maxX, maxY, labelX, y int, label string, field *historyField, focused bool, inputX, arrowX int, labelStyle, accelStyle tcell.Style) (inputEndX int) {
	plain, accel := stripAccelerator(label)
	runes := []rune(plain)
	for i, r := range runes {
		style := labelStyle
		if i == accel {
			style = accelStyle
		}
		putRune(screen, maxX, maxY, labelX+i, y, r, style)
	}

	inputEndX = arrowX - 2
	inputW := inputEndX - inputX + 1

	inputBg := vgaBlue
	if focused {
		inputBg = vgaLightBlue
	}
	inputStyle := tcell.StyleDefault.Foreground(vgaLightCyan).Background(inputBg)
	for col := 0; col < inputW; col++ {
		putRune(screen, maxX, maxY, inputX+col, y, ' ', inputStyle)
	}

	offset := 0
	if field.cursor >= inputW {
		offset = field.cursor - inputW + 1
	}
	for i, r := range field.text {
		col := i - offset
		if col < 0 || col >= inputW {
			continue
		}
		putRune(screen, maxX, maxY, inputX+col, y, r, inputStyle)
	}

	if focused {
		curCol := field.cursor - offset
		if curCol >= 0 && curCol < inputW {
			ch := rune(' ')
			if field.cursor < len(field.text) {
				ch = field.text[field.cursor]
			}
			putRune(screen, maxX, maxY, inputX+curCol, y, ch, tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}

	arrowStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaGreen)
	putRune(screen, maxX, maxY, arrowX, y, '↓', arrowStyle)
	return inputEndX
}

// drawFieldHistoryDropdown desenha a caixa de histórico logo abaixo do campo,
// alinhada em (dX,dY) com largura dW.
func drawFieldHistoryDropdown(screen tcell.Screen, maxX, maxY, dX, dY, dW int, field *historyField) {
	if len(field.history) == 0 {
		return
	}
	bgStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)
	selStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	border := bgStyle

	dH := len(field.history) + 2
	if dH > 10 {
		dH = 10
	}

	for row := 0; row < dH; row++ {
		for col := 0; col < dW; col++ {
			putRune(screen, maxX, maxY, dX+col, dY+row, ' ', bgStyle)
		}
	}
	for col := 1; col < dW-1; col++ {
		putRune(screen, maxX, maxY, dX+col, dY, '─', border)
		putRune(screen, maxX, maxY, dX+col, dY+dH-1, '─', border)
	}
	for row := 1; row < dH-1; row++ {
		putRune(screen, maxX, maxY, dX, dY+row, '│', border)
		putRune(screen, maxX, maxY, dX+dW-1, dY+row, '│', border)
	}
	putRune(screen, maxX, maxY, dX, dY, '┌', border)
	putRune(screen, maxX, maxY, dX+dW-1, dY, '┐', border)
	putRune(screen, maxX, maxY, dX, dY+dH-1, '└', border)
	putRune(screen, maxX, maxY, dX+dW-1, dY+dH-1, '┘', border)

	for i, entry := range field.history {
		if i >= dH-2 {
			break
		}
		style := bgStyle
		if i == field.historyIdx {
			style = selStyle
		}
		putString(screen, maxX, maxY, dX+1, dY+1+i, fdPad(entry, dW-2), style)
	}
}

func fdPad(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width])
	}
	for len(runes) < width {
		runes = append(runes, ' ')
	}
	return string(runes)
}
