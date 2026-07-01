package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type helpWindow struct {
	*tview.Box
	theme           Theme
	number          int
	content         *HelpContent
	offsetRow       int
	offsetCol       int
	selectedLinkIdx int
	focused         bool
	onClose         func() // Callback chamado ao fechar a janela
}

func newHelpWindow(theme Theme, number int) *helpWindow {
	return &helpWindow{
		Box:             tview.NewBox().SetBackgroundColor(theme.HelpBg),
		theme:           theme,
		number:          number,
		content:         NewHelpContent(),
		offsetRow:       0,
		offsetCol:       0,
		selectedLinkIdx: -1,
		onClose:         nil,
	}
}

func (w *helpWindow) Draw(screen tcell.Screen) {
	x, y, width, height := w.GetRect()
	if width < 8 || height < 6 {
		return
	}

	maxX, maxY := screen.Size()
	frameStyle := tcell.StyleDefault.Foreground(w.theme.HelpBorderFg).Background(w.theme.HelpBg)
	bgStyle := tcell.StyleDefault.Foreground(w.theme.HelpFg).Background(w.theme.HelpBg)

	// Preencher fundo
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	// Desenhar moldura
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

	// Desenhar conteúdo
	innerX, innerY := x+1, y+1
	innerWidth, innerHeight := width-2, height-2

	w.drawContent(screen, maxX, maxY, innerX, innerY, innerWidth, innerHeight)

	// Título e número
	putString(screen, maxX, maxY, x+2, y, "[■]", frameStyle)
	title := " " + w.content.GetCurrentTopic().Title + " "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, frameStyle)

	scrollMarkX := x + width - 7
	putString(screen, maxX, maxY, scrollMarkX, y, "[↕]", frameStyle)
	numberText := fmt.Sprintf("%d", w.number)
	putString(screen, maxX, maxY, scrollMarkX-2-len([]rune(numberText)), y, numberText, frameStyle)

	// Barra de status
	statusText := fmt.Sprintf(" L:%d C:%d ", w.offsetRow+1, w.offsetCol+1)
	statusWidth := innerWidth / 4
	if statusWidth < len([]rune(statusText))+1 {
		statusWidth = len([]rune(statusText)) + 1
	}
	putString(screen, maxX, maxY, x+1, y+height-1, statusText, frameStyle)

	// Barra de rolagem horizontal
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

	// Barra de rolagem vertical
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

func (w *helpWindow) drawContent(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	topic := w.content.GetCurrentTopic()
	if topic.ID == "editor_commands" {
		w.drawEditorCommandsSubWindow(screen, maxX, maxY, x, y, width, height, topic)
		return
	}
	if headerTitle, ok := helpHeaderButtonTitles[topic.ID]; ok {
		w.drawCommandsHeaderPage(screen, maxX, maxY, x, y, width, height, topic, headerTitle)
		return
	}

	startY := y
	contentHeight := height
	if len(w.content.BreadcrumbTrail()) > 1 {
		crumb := w.content.BreadcrumbText()
		putString(screen, maxX, maxY, x, y, crumb, tcell.StyleDefault.Foreground(vgaBlack).Background(w.theme.HelpBg))
		startY = y + 1
		contentHeight--
		if contentHeight < 0 {
			contentHeight = 0
		}
	}

	lines := w.content.GetPagedContent(w.offsetRow, contentHeight)

	for i, line := range lines {
		actualRow := w.offsetRow + i
		w.drawLine(screen, maxX, maxY, x, startY+i, width, line, actualRow, topic)
	}
}

func (w *helpWindow) drawEditorCommandsSubWindow(screen tcell.Screen, maxX, maxY, x, y, width, height int, topic *HelpTopic) {
	bgStyle := tcell.StyleDefault.Foreground(w.theme.HelpFg).Background(w.theme.HelpBg)
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	panelW := 46
	if panelW > width-2 {
		panelW = width - 2
	}
	panelH := 16
	if panelH > height-1 {
		panelH = height - 1
	}
	if panelW < 24 || panelH < 8 {
		return
	}

	panelX := x + (width-panelW)/2
	panelY := y + (height-panelH)/2

	borderStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(w.theme.HelpBg)
	for col := 1; col < panelW-1; col++ {
		putRune(screen, maxX, maxY, panelX+col, panelY, '─', borderStyle)
		putRune(screen, maxX, maxY, panelX+col, panelY+panelH-1, '─', borderStyle)
	}
	for row := 1; row < panelH-1; row++ {
		putRune(screen, maxX, maxY, panelX, panelY+row, '│', borderStyle)
		putRune(screen, maxX, maxY, panelX+panelW-1, panelY+row, '│', borderStyle)
	}
	putRune(screen, maxX, maxY, panelX, panelY, '┌', borderStyle)
	putRune(screen, maxX, maxY, panelX+panelW-1, panelY, '┐', borderStyle)
	putRune(screen, maxX, maxY, panelX, panelY+panelH-1, '└', borderStyle)
	putRune(screen, maxX, maxY, panelX+panelW-1, panelY+panelH-1, '┘', borderStyle)

	header := newHelpHeaderButton("Editor Commands", vgaBlack, w.theme.HelpBg, vgaDarkGray)
	header.draw(screen, maxX, maxY, panelX+2, panelY+1, w.theme.HelpBg)
	crumb := w.content.BreadcrumbText()
	if crumb != "" {
		putString(screen, maxX, maxY, panelX+2, panelY+2, crumb, tcell.StyleDefault.Foreground(vgaBlack).Background(w.theme.HelpBg))
	}

	optionsStartY := panelY + 4
	for i, link := range topic.Links {
		if i >= 5 {
			break
		}
		style := tcell.StyleDefault.Foreground(w.theme.HelpLinkFg).Background(w.theme.HelpBg)
		if i == w.selectedLinkIdx {
			style = tcell.StyleDefault.Foreground(w.theme.HelpSelectedFg).Background(w.theme.HelpBg)
		}
		putString(screen, maxX, maxY, panelX+4, optionsStartY+i, link.Text, style)
	}

	if len(topic.Links) > 5 {
		idx := len(topic.Links) - 1
		style := tcell.StyleDefault.Foreground(w.theme.HelpLinkFg).Background(w.theme.HelpBg)
		if idx == w.selectedLinkIdx {
			style = tcell.StyleDefault.Foreground(w.theme.HelpSelectedFg).Background(w.theme.HelpBg)
		}
		putString(screen, maxX, maxY, panelX+4, panelY+panelH-2, topic.Links[idx].Text, style)
	}
}

// helpHeaderButtonTitles mapeia tópicos cuja linha 1 deve ser substituída por
// um botão 3D de título (estilo "Block Commands"), para o texto do botão.
var helpHeaderButtonTitles = map[string]string{
	"block_commands":           "Block Commands",
	"cursor_movement_commands": "Cursor Movement Commands",
	"insert_delete_commands":   "Insert & Delete Commands",
	"miscelaneous_commands":    "Miscellaneous Editor Commands",
}

// drawCommandsHeaderPage renderiza uma página de comandos (ex.: "Block
// commands", "Cursor-movement commands") com o título em botão 3D. A linha 1
// do tópico é substituída visualmente por um helpHeaderButton (cyan/preto/
// sombra), mantendo o estilo Turbo Pascal. As demais linhas usam drawLine
// normalmente.
func (w *helpWindow) drawCommandsHeaderPage(screen tcell.Screen, maxX, maxY, x, y, width, height int, topic *HelpTopic, headerTitle string) {
	startY := y
	contentHeight := height
	if len(w.content.BreadcrumbTrail()) > 1 {
		crumb := w.content.BreadcrumbText()
		putString(screen, maxX, maxY, x, y, crumb, tcell.StyleDefault.Foreground(vgaBlack).Background(w.theme.HelpBg))
		startY = y + 1
		contentHeight--
		if contentHeight < 0 {
			contentHeight = 0
		}
	}

	const titleRow = 1 // a linha 1 do tópico contém o texto do título

	lines := w.content.GetPagedContent(w.offsetRow, contentHeight)
	titleScreenRow := -1
	for i, line := range lines {
		actualRow := w.offsetRow + i
		screenRow := startY + i
		if actualRow == titleRow {
			titleScreenRow = screenRow // marca para desenhar o botão depois
		} else {
			w.drawLine(screen, maxX, maxY, x, screenRow, width, line, actualRow, topic)
		}
	}
	// Desenha o botão 3D por último para que sua sombra não seja sobrescrita por drawLine.
	if titleScreenRow >= 0 {
		btn := newHelpHeaderButton(headerTitle, vgaBlack, w.theme.HelpBg, vgaDarkGray)
		btn.draw(screen, maxX, maxY, x+2, titleScreenRow, w.theme.HelpBg)
	}
}

func (w *helpWindow) drawLine(screen tcell.Screen, maxX, maxY, x, y, width int, line string, rowIdx int, topic *HelpTopic) {
	runes := []rune(line)
	bgStyle := tcell.StyleDefault.Foreground(w.theme.HelpFg).Background(w.theme.HelpBg)
	drawnUpto := 0

	for colIdx, r := range runes {
		displayCol := colIdx - w.offsetCol
		if displayCol < 0 {
			continue
		}
		if displayCol >= width {
			break
		}

		style := bgStyle
		for i, link := range topic.Links {
			if link.Row == rowIdx && colIdx >= link.ColStart && colIdx < link.ColEnd {
				if i == w.selectedLinkIdx {
					style = tcell.StyleDefault.Foreground(w.theme.HelpSelectedFg).Background(w.theme.HelpBg)
				} else {
					style = tcell.StyleDefault.Foreground(w.theme.HelpLinkFg).Background(w.theme.HelpBg)
				}
				break
			}
		}

		putRune(screen, maxX, maxY, x+displayCol, y, r, style)
		drawnUpto = displayCol + 1
	}

	for col := drawnUpto; col < width; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', bgStyle)
	}
}

func (w *helpWindow) horizontalCursorPosition(start, end int) int {
	if end <= start {
		return start
	}

	maxCol := w.content.GetMaxCol()
	if maxCol <= 0 {
		return start
	}

	rangeCols := maxCol
	if rangeCols <= 0 {
		return start
	}
	if w.offsetCol > rangeCols {
		w.offsetCol = rangeCols
	}
	trackLen := end - start
	if rangeCols == 0 {
		return start
	}
	return start + (w.offsetCol*trackLen)/rangeCols
}

func (w *helpWindow) verticalCursorPosition(start, end int) int {
	if end <= start {
		return start
	}

	maxLine := w.content.GetMaxLine()
	if maxLine <= 0 {
		return start
	}

	rangeRows := maxLine
	if rangeRows <= 0 {
		return start
	}
	if w.offsetRow > rangeRows {
		w.offsetRow = rangeRows
	}
	trackLen := end - start
	if rangeRows == 0 {
		return start
	}
	return start + (w.offsetRow*trackLen)/rangeRows
}

func (w *helpWindow) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return w.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}
		currentTopic := w.content.GetCurrentTopic()

		switch event.Key() {
		case tcell.KeyEscape:
			if w.onClose != nil {
				w.onClose()
			}
		case tcell.KeyUp:
			if currentTopic.ID == "editor_commands" {
				w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, false)
			} else {
				w.scrollUp()
			}
		case tcell.KeyDown:
			if currentTopic.ID == "editor_commands" {
				w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, true)
			} else {
				w.scrollDown()
			}
		case tcell.KeyLeft:
			w.scrollLeft()
		case tcell.KeyRight:
			w.scrollRight()
		case tcell.KeyPgUp:
			w.pageUp()
		case tcell.KeyPgDn:
			w.pageDown()
		case tcell.KeyHome:
			w.offsetRow = 0
			w.offsetCol = 0
		case tcell.KeyEnd:
			maxLine := w.content.GetMaxLine()
			w.offsetRow = maxLine - 1
			if w.offsetRow < 0 {
				w.offsetRow = 0
			}
		case tcell.KeyTab:
			w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, true)
		case tcell.KeyBacktab:
			w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, false)
		case tcell.KeyEnter:
			if w.selectedLinkIdx >= 0 {
				link := w.content.GetLinkAtIndex(w.selectedLinkIdx)
				if link != nil && w.content.NavigateToTopic(link.TopicID) {
					w.resetAfterTopicChange()
				}
			}
		case tcell.KeyF1:
			// Alt+F1 voltar ao tópico anterior
			if event.Modifiers()&tcell.ModAlt != 0 {
				w.goBackBreadcrumb()
			}
		case tcell.KeyRune:
			if event.Modifiers()&tcell.ModAlt != 0 && (event.Rune() == 'q' || event.Rune() == 'Q') {
				// fallback opcional em terminais que não entregam Alt+F1 corretamente
				w.goBackBreadcrumb()
			}
		}
	})
}

func (w *helpWindow) resetAfterTopicChange() {
	w.offsetRow = 0
	w.offsetCol = 0
	if w.content.GetCurrentTopic().ID == "editor_commands" {
		w.selectedLinkIdx = 0
	} else {
		w.selectedLinkIdx = -1
	}
}

func (w *helpWindow) goBackBreadcrumb() bool {
	if !w.content.GoBack() {
		return false
	}
	w.resetAfterTopicChange()
	return true
}

func (w *helpWindow) scrollUp() {
	if w.offsetRow > 0 {
		w.offsetRow--
	}
}

func (w *helpWindow) scrollDown() {
	maxLine := w.content.GetMaxLine()
	if w.offsetRow < maxLine-1 {
		w.offsetRow++
	}
}

func (w *helpWindow) scrollLeft() {
	if w.offsetCol > 0 {
		w.offsetCol--
	}
}

func (w *helpWindow) scrollRight() {
	maxCol := w.content.GetMaxCol()
	_, _, width, _ := w.GetRect()
	innerWidth := width - 2
	if w.offsetCol < maxCol-innerWidth {
		w.offsetCol++
	}
}

func (w *helpWindow) pageUp() {
	_, _, _, height := w.GetRect()
	pageSize := height - 3
	if w.offsetRow > pageSize {
		w.offsetRow -= pageSize
	} else {
		w.offsetRow = 0
	}
}

func (w *helpWindow) pageDown() {
	_, _, _, height := w.GetRect()
	pageSize := height - 3
	maxLine := w.content.GetMaxLine()
	w.offsetRow += pageSize
	if w.offsetRow+pageSize > maxLine {
		w.offsetRow = maxLine - pageSize
		if w.offsetRow < 0 {
			w.offsetRow = 0
		}
	}
}

func (w *helpWindow) Focus(delegate func(p tview.Primitive)) {
	w.focused = true
}

func (w *helpWindow) HasFocus() bool {
	return w.focused
}

func (w *helpWindow) Blur() {
	w.focused = false
}

func (w *helpWindow) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		wx, wy, ww, wh := w.GetRect()

		// Só trata eventos dentro dos limites da janela
		if mx < wx || mx >= wx+ww || my < wy || my >= wy+wh {
			return false, nil
		}

		rx, ry := mx-wx, my-wy
		topic := w.content.GetCurrentTopic()

		switch action {
		case tview.MouseScrollUp:
			if topic.ID == "editor_commands" {
				w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, false)
			} else {
				w.scrollUp()
			}
			return true, nil

		case tview.MouseScrollDown:
			if topic.ID == "editor_commands" {
				w.selectedLinkIdx = w.content.FindNextLink(w.selectedLinkIdx, true)
			} else {
				w.scrollDown()
			}
			return true, nil

		case tview.MouseLeftClick:
			setFocus(w)

			// ── Botão fechar [■] no canto esquerdo da borda superior ──────────
			if ry == 0 && rx >= 2 && rx <= 4 {
				if w.onClose != nil {
					w.onClose()
				}
				return true, nil
			}

			// ── Scrollbar vertical (borda direita) ────────────────────────────
			if rx == ww-1 && ry >= 1 && ry < wh-1 {
				vTop, vBottom := 1, wh-2
				switch {
				case ry == vTop:
					w.scrollUp()
				case ry == vBottom:
					w.scrollDown()
				default:
					trackLen := vBottom - vTop - 1
					if trackLen > 0 {
						maxLine := w.content.GetMaxLine()
						w.offsetRow = ((ry - vTop - 1) * maxLine) / trackLen
						if w.offsetRow < 0 {
							w.offsetRow = 0
						}
					}
				}
				return true, nil
			}

			// ── Scrollbar horizontal (borda inferior) ─────────────────────────
			if ry == wh-1 && rx >= 1 && rx < ww-1 {
				mid := ww / 2
				if rx < mid {
					w.scrollLeft()
				} else {
					w.scrollRight()
				}
				return true, nil
			}

			// ── Área de conteúdo ──────────────────────────────────────────────
			if rx < 1 || rx >= ww-1 || ry < 1 || ry >= wh-1 {
				return true, nil
			}

			if topic.ID == "editor_commands" {
				// Calcular limites do painel (idêntico a drawEditorCommandsSubWindow)
				innerX, innerY, innerW, innerH := wx+1, wy+1, ww-2, wh-2
				panelW := 46
				if panelW > innerW-2 {
					panelW = innerW - 2
				}
				panelH := 16
				if panelH > innerH-1 {
					panelH = innerH - 1
				}
				if panelW < 24 || panelH < 8 {
					return true, nil
				}
				panelX := innerX + (innerW-panelW)/2
				panelY := innerY + (innerH-panelH)/2
				bodyTop := panelY + 3
				bodyHeight := panelH - 4

				if mx >= panelX+2 && mx < panelX+panelW-2 && my >= bodyTop && my < bodyTop+bodyHeight {
					lineIdx := my - bodyTop
					colIdx := mx - (panelX + 2) + w.offsetCol
					for i, link := range topic.Links {
						if link.Row == lineIdx && colIdx >= link.ColStart && colIdx < link.ColEnd {
							w.selectedLinkIdx = i
							if w.content.NavigateToTopic(link.TopicID) {
								w.resetAfterTopicChange()
							}
							return true, nil
						}
					}
				}
			} else {
				// Calcular posição no conteúdo (ajustar para o breadcrumb se presente)
				contentRowBase := wy + 1
				if len(w.content.BreadcrumbTrail()) > 1 {
					contentRowBase++
				}
				contentY := (my - contentRowBase) + w.offsetRow
				contentX := (mx - (wx + 1)) + w.offsetCol

				if contentY >= 0 && contentX >= 0 {
					for i, link := range topic.Links {
						if link.Row == contentY && contentX >= link.ColStart && contentX < link.ColEnd {
							w.selectedLinkIdx = i
							if w.content.NavigateToTopic(link.TopicID) {
								w.resetAfterTopicChange()
							}
							return true, nil
						}
					}
				}
			}
			return true, nil
		}

		return false, nil
	}
}

func (w *helpWindow) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return func(text string, setFocus func(p tview.Primitive)) {
	}
}

