package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type editorWindow struct {
	*tview.Box
	editor           *tview.TextArea
	app              *App   // referência à aplicação (para diálogos e status)
	theme            Theme
	fileName         string // base name (exibido na barra de título)
	filePath         string // caminho completo (vazio = arquivo não salvo ainda)
	isTokenized      bool   // true = arquivo original era tokenizado MSX-BASIC
	number           int
	onClose          func()
	onExitToMenu     func() // Ctrl+K D — foca a barra de menu
	highlightEnabled bool

	// Comandos de bloco (estilo Turbo Pascal)
	blkBeginRow int    // linha do início do bloco (-1 = não marcado)
	blkBeginCol int    // coluna do início do bloco (rune index)
	blkEndRow   int    // linha do fim do bloco (-1 = não marcado)
	blkEndCol   int    // coluna do fim do bloco
	blkVisible  bool   // exibir highlight do bloco
	blkClip     string // clipboard interno para operações de bloco
	waitingK    bool   // recebeu Ctrl+K, aguardando segunda tecla
	waitingQ    bool   // recebeu Ctrl+Q, aguardando segunda tecla

	// Posição e tamanho da janela flutuante (gerenciados por nós, não pelo tview layout)
	winX, winY int
	winW, winH int
	positioned bool // false até o primeiro Draw inicializar com base na tela

	// Maximize/restore
	savedX, savedY  int
	savedW, savedH  int
	isMaximized     bool

	// Drag da barra de título
	isDragging   bool
	dragOffX     int
	dragOffY     int

	// Resize pelo canto inferior direito
	isResizing bool

	// Tamanho da tela cacheado para clamping durante drag/resize
	lastScreenW, lastScreenH int
}

func newEditorWindow(theme Theme, fileName string, number int) *editorWindow {
	editor := tview.NewTextArea()
	editor.SetPlaceholder("Digite seu codigo aqui...")
	editor.SetWrap(false)
	editor.SetWordWrap(false)
	editor.SetTextStyle(tcell.StyleDefault.Foreground(theme.EditorFg).Background(theme.EditorBg))
	editor.SetPlaceholderStyle(tcell.StyleDefault.Foreground(theme.EditorFg).Background(theme.EditorBg))
	editor.SetBackgroundColor(theme.EditorBg)

	ew := &editorWindow{
		Box:         tview.NewBox().SetBackgroundColor(theme.EditorBg),
		editor:      editor,
		theme:       theme,
		fileName:    fileName,
		number:      number,
		blkBeginRow: -1,
		blkEndRow:   -1,
	}
	// Integra clipboard do editor com o clipboard de bloco
	editor.SetClipboard(
		func(text string) { ew.blkClip = text },
		func() string { return ew.blkClip },
	)
	return ew
}

// zoomSymbol retorna o símbolo do botão de maximizar/restaurar.
func (w *editorWindow) zoomSymbol() string {
	if w.isMaximized {
		return "[▼]"
	}
	return "[▲]"
}

func (w *editorWindow) Draw(screen tcell.Screen) {
	sw, sh := screen.Size()
	w.lastScreenW, w.lastScreenH = sw, sh

	// Inicializa posição no primeiro Draw, usando o rect fornecido pelo tview layout
	if !w.positioned {
		px, py, pw, ph := w.GetRect()
		w.winX = px + 2
		w.winY = py
		w.winW = pw - 4
		w.winH = ph - 1
		w.savedX, w.savedY = w.winX, w.winY
		w.savedW, w.savedH = w.winW, w.winH
		w.positioned = true
	}

	x, y, width, height := w.winX, w.winY, w.winW, w.winH
	if width < 20 || height < 6 {
		return
	}

	maxX, maxY := sw, sh
	frameStyle := tcell.StyleDefault.Foreground(w.theme.EditorBorderFg).Background(w.theme.EditorBg)
	bgStyle := tcell.StyleDefault.Foreground(w.theme.EditorFg).Background(w.theme.EditorBg)

	// Preencher fundo da janela
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	// Moldura dupla
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

	// Handle de resize: canto inferior direito substituído por ◢
	putRune(screen, maxX, maxY, x+width-1, y+height-1, '◢', frameStyle)

	// Área interna do editor
	innerX, innerY := x+1, y+1
	innerWidth, innerHeight := width-2, height-2
	w.editor.SetRect(innerX, innerY, innerWidth, innerHeight)
	w.editor.Draw(screen)

	if w.highlightEnabled {
		w.applySyntaxHighlight(screen, innerX, innerY, innerWidth, innerHeight)
	}
	if w.blkVisible && w.blkBeginRow >= 0 && w.blkEndRow >= 0 {
		w.applyBlockHighlight(screen, innerX, innerY, innerWidth, innerHeight)
	}

	// Botão fechar [■] — canto esquerdo do título
	putString(screen, maxX, maxY, x+2, y, "[■]", frameStyle)

	// Título centralizado
	title := " " + w.fileName + " "
	titleX := x + (width-len([]rune(title)))/2
	if titleX < x+6 {
		titleX = x + 6
	}
	putString(screen, maxX, maxY, titleX, y, title, frameStyle)

	// Botão zoom [▲]/[▼] e número da janela — lado direito do título
	zoomX := x + width - 5
	putString(screen, maxX, maxY, zoomX, y, w.zoomSymbol(), frameStyle)
	numberText := fmt.Sprintf("%d", w.number)
	numX := zoomX - 1 - len([]rune(numberText))
	if numX > x+5 {
		putString(screen, maxX, maxY, numX, y, numberText, frameStyle)
	}

	// Texto de posição do cursor na borda inferior (status)
	fromRow, fromCol, _, _ := w.editor.GetCursor()
	statusText := fmt.Sprintf(" %d:%d ", fromRow+1, fromCol+1)
	putString(screen, maxX, maxY, x+1, y+height-1, statusText, frameStyle)
	statusWidth := len([]rune(statusText))

	// Scrollbar horizontal (borda inferior)
	hLeft := x + 1 + statusWidth
	hRight := x + width - 2
	if hLeft+2 < hRight {
		putRune(screen, maxX, maxY, hLeft, y+height-1, '◄', frameStyle)
		putRune(screen, maxX, maxY, hRight, y+height-1, '►', frameStyle)
		for col := hLeft + 1; col < hRight; col++ {
			putRune(screen, maxX, maxY, col, y+height-1, '▒', frameStyle)
		}
		thumbX := w.hThumbPos(hLeft+1, hRight-1)
		putRune(screen, maxX, maxY, thumbX, y+height-1, '■', frameStyle)
	}

	// Scrollbar vertical (borda direita)
	vTop := y + 1
	vBottom := y + height - 2
	if vTop+1 < vBottom {
		putRune(screen, maxX, maxY, x+width-1, vTop, '▲', frameStyle)
		putRune(screen, maxX, maxY, x+width-1, vBottom, '▼', frameStyle)
		for row := vTop + 1; row < vBottom; row++ {
			putRune(screen, maxX, maxY, x+width-1, row, '▒', frameStyle)
		}
		thumbY := w.vThumbPos(vTop+1, vBottom-1)
		putRune(screen, maxX, maxY, x+width-1, thumbY, '■', frameStyle)
	}
}

// hThumbPos calcula a posição do thumb da scrollbar horizontal.
func (w *editorWindow) hThumbPos(start, end int) int {
	if end <= start {
		return start
	}
	_, offsetCol := w.editor.GetOffset()
	fieldWidth := w.editor.GetFieldWidth()
	if fieldWidth <= 0 {
		return start
	}
	maxCol := 0
	for _, line := range strings.Split(w.editor.GetText(), "\n") {
		if l := len([]rune(line)); l > maxCol {
			maxCol = l
		}
	}
	rangeCols := maxCol - fieldWidth
	if rangeCols <= 0 {
		return start
	}
	if offsetCol > rangeCols {
		offsetCol = rangeCols
	}
	trackLen := end - start
	return start + (offsetCol*trackLen)/rangeCols
}

// vThumbPos calcula a posição do thumb da scrollbar vertical.
func (w *editorWindow) vThumbPos(start, end int) int {
	if end <= start {
		return start
	}
	offsetRow, _ := w.editor.GetOffset()
	fieldHeight := w.editor.GetFieldHeight()
	if fieldHeight <= 0 {
		return start
	}
	totalRows := len(strings.Split(w.editor.GetText(), "\n"))
	rangeRows := totalRows - fieldHeight
	if rangeRows <= 0 {
		return start
	}
	if offsetRow > rangeRows {
		offsetRow = rangeRows
	}
	trackLen := end - start
	return start + (offsetRow*trackLen)/rangeRows
}

// clampPosition mantém a janela com pelo menos parte do título visível.
func (w *editorWindow) clampPosition() {
	if w.lastScreenH > 0 {
		if w.winY < 1 {
			w.winY = 1
		}
		if w.winY > w.lastScreenH-3 {
			w.winY = w.lastScreenH - 3
		}
	}
	if w.lastScreenW > 0 {
		minVisible := 10
		if w.winX+w.winW < minVisible {
			w.winX = minVisible - w.winW
		}
		if w.winX > w.lastScreenW-minVisible {
			w.winX = w.lastScreenW - minVisible
		}
	}
}

// clampSize garante tamanho mínimo e que a janela não ultrapasse a tela.
func (w *editorWindow) clampSize() {
	if w.winW < 20 {
		w.winW = 20
	}
	if w.winH < 6 {
		w.winH = 6
	}
	if w.lastScreenW > 0 && w.winX+w.winW > w.lastScreenW {
		w.winW = w.lastScreenW - w.winX
	}
	if w.lastScreenH > 0 && w.winY+w.winH > w.lastScreenH-1 {
		w.winH = w.lastScreenH - 1 - w.winY
	}
}

// toggleMaximize maximiza ou restaura a janela.
func (w *editorWindow) toggleMaximize() {
	if w.isMaximized {
		w.winX, w.winY = w.savedX, w.savedY
		w.winW, w.winH = w.savedW, w.savedH
		w.isMaximized = false
	} else {
		w.savedX, w.savedY = w.winX, w.winY
		w.savedW, w.savedH = w.winW, w.winH
		w.winX = 0
		w.winY = 1 // linha 0 = menu bar
		w.winW = w.lastScreenW
		w.winH = w.lastScreenH - 2 // -2 = menu + status bar
		w.isMaximized = true
	}
}

// scrollVertical aciona o scroll vertical pelo clique na scrollbar.
func (w *editorWindow) scrollVertical(my, vTop, vBottom int) {
	offsetRow, offsetCol := w.editor.GetOffset()
	switch my {
	case vTop: // seta ▲
		if offsetRow > 0 {
			w.editor.SetOffset(offsetRow-1, offsetCol)
		}
	case vBottom: // seta ▼
		w.editor.SetOffset(offsetRow+1, offsetCol)
	default: // clique na trilha
		totalRows := len(strings.Split(w.editor.GetText(), "\n"))
		trackLen := vBottom - vTop - 1
		if trackLen > 0 {
			row := (my - vTop - 1) * totalRows / trackLen
			w.editor.SetOffset(row, offsetCol)
		}
	}
}

// scrollHorizontal aciona o scroll horizontal pelo clique na scrollbar.
func (w *editorWindow) scrollHorizontal(mx, hLeft, hRight int) {
	offsetRow, offsetCol := w.editor.GetOffset()
	switch mx {
	case hLeft: // seta ◄
		if offsetCol > 0 {
			w.editor.SetOffset(offsetRow, offsetCol-1)
		}
	case hRight: // seta ►
		w.editor.SetOffset(offsetRow, offsetCol+1)
	default: // clique na trilha
		maxCol := 0
		for _, line := range strings.Split(w.editor.GetText(), "\n") {
			if l := len([]rune(line)); l > maxCol {
				maxCol = l
			}
		}
		trackLen := hRight - hLeft - 1
		if trackLen > 0 {
			col := (mx - hLeft - 1) * maxCol / trackLen
			w.editor.SetOffset(offsetRow, col)
		}
	}
}

func (w *editorWindow) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	editorHandler := w.editor.InputHandler()
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}
		if w.handleBlockKey(event, setFocus) {
			return
		}
		if editorHandler != nil {
			editorHandler(event, setFocus)
		}
	}
}

func (w *editorWindow) Focus(delegate func(p tview.Primitive)) {
	w.editor.Focus(delegate)
}

func (w *editorWindow) HasFocus() bool {
	return w.editor.HasFocus()
}

func (w *editorWindow) Blur() {
	w.editor.Blur()
}

func (w *editorWindow) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	base := w.editor.MouseHandler()
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := w.winX, w.winY, w.winW, w.winH
		inBounds := mx >= x && mx < x+width && my >= y && my < y+height

		// ── Drag e Resize: consumidos independente de posição (mouse capturado) ──

		if action == tview.MouseMove {
			if w.isDragging {
				w.winX = mx - w.dragOffX
				w.winY = my - w.dragOffY
				w.clampPosition()
				return true, w
			}
			if w.isResizing {
				newW := mx - w.winX + 1
				newH := my - w.winY + 1
				if newW >= 20 {
					w.winW = newW
				}
				if newH >= 6 {
					w.winH = newH
				}
				w.clampSize()
				return true, w
			}
			return false, nil
		}

		if action == tview.MouseLeftUp {
			if w.isDragging || w.isResizing {
				w.isDragging = false
				w.isResizing = false
				return true, nil // libera captura
			}
			return false, nil
		}

		// Fora dos limites da janela: ignora
		if !inBounds {
			return false, nil
		}

		// ── Posições dos elementos da borda ──────────────────────────────────
		zoomX := x + width - 5
		fromRow, fromCol, _, _ := w.editor.GetCursor()
		statusText := fmt.Sprintf(" %d:%d ", fromRow+1, fromCol+1)
		statusWidth := len([]rune(statusText))
		hLeft := x + 1 + statusWidth
		hRight := x + width - 2
		vTop := y + 1
		vBottom := y + height - 2

		switch action {
		case tview.MouseLeftDown:
			setFocus(w)

			// Handle de resize: canto inferior direito ◢
			if mx == x+width-1 && my == y+height-1 {
				w.isResizing = true
				return true, w
			}

			// Drag pelo título (excluindo botões [■] e zoom)
			if my == y && mx > x+4 && mx < zoomX-1 {
				w.isDragging = true
				w.dragOffX = mx - x
				w.dragOffY = my - y
				return true, w
			}

			// Clique na janela em geral: só transfere foco
			return true, w

		case tview.MouseLeftClick:
			// Botão fechar [■]
			if my == y && mx >= x+2 && mx <= x+4 {
				if w.onClose != nil {
					w.onClose()
				}
				return true, nil
			}

			// Botão zoom [▲]/[▼]
			if my == y && mx >= zoomX && mx <= zoomX+2 {
				w.toggleMaximize()
				return true, nil
			}

			// Scrollbar vertical (borda direita)
			if mx == x+width-1 && my >= vTop && my <= vBottom {
				w.scrollVertical(my, vTop, vBottom)
				return true, nil
			}

			// Scrollbar horizontal (borda inferior)
			if my == y+height-1 && mx >= hLeft && mx <= hRight {
				w.scrollHorizontal(mx, hLeft, hRight)
				return true, nil
			}

		case tview.MouseScrollUp:
			offsetRow, offsetCol := w.editor.GetOffset()
			if offsetRow > 0 {
				w.editor.SetOffset(offsetRow-1, offsetCol)
			}
			return true, nil

		case tview.MouseScrollDown:
			offsetRow, offsetCol := w.editor.GetOffset()
			w.editor.SetOffset(offsetRow+1, offsetCol)
			return true, nil
		}

		// Delega eventos de conteúdo ao TextArea
		return base(action, event, setFocus)
	}
}

func (w *editorWindow) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return w.editor.PasteHandler()
}

// ── Syntax Highlighting ─────────────────────────────────────────────────────

// styleForKind retorna o tcell.Style para cada tipo de token MSX-BASIC.
// Paleta inspirada no Turbo Pascal 7 sobre fundo azul VGA.
func (w *editorWindow) styleForKind(kind basicTokenKind) tcell.Style {
	bg := w.theme.EditorBg
	switch kind {
	case basicKindLineNumber:
		return tcell.StyleDefault.Foreground(vgaWhite).Background(bg).Bold(true)
	case basicKindStatement:
		return tcell.StyleDefault.Foreground(vgaLightCyan).Background(bg)
	case basicKindModifier:
		return tcell.StyleDefault.Foreground(vgaCyan).Background(bg)
	case basicKindFunction:
		return tcell.StyleDefault.Foreground(vgaYellow).Background(bg)
	case basicKindString:
		return tcell.StyleDefault.Foreground(vgaLightGreen).Background(bg)
	case basicKindNumber:
		return tcell.StyleDefault.Foreground(vgaLightMag).Background(bg)
	case basicKindComment:
		return tcell.StyleDefault.Foreground(vgaDarkGray).Background(bg)
	case basicKindOperator:
		return tcell.StyleDefault.Foreground(vgaLightGray).Background(bg)
	case basicKindSymbol:
		return tcell.StyleDefault.Foreground(vgaLightGray).Background(bg)
	case basicKindVariable:
		return tcell.StyleDefault.Foreground(vgaWhite).Background(bg)
	default:
		return tcell.StyleDefault.Foreground(w.theme.EditorFg).Background(bg)
	}
}

// applySyntaxHighlight re-renderiza o conteúdo visível do editor com cores de
// syntax highlighting por cima do que o tview.TextArea já desenhou.
// A posição do cursor é preservada.
func (w *editorWindow) applySyntaxHighlight(screen tcell.Screen, innerX, innerY, innerWidth, innerHeight int) {
	text := w.editor.GetText()
	if text == "" {
		return
	}

	lines := strings.Split(text, "\n")
	offsetRow, offsetCol := w.editor.GetOffset()
	cursorRow, cursorCol, _, _ := w.editor.GetCursor()
	maxX, maxY := screen.Size()

	for visRow := 0; visRow < innerHeight; visRow++ {
		textRow := offsetRow + visRow
		if textRow >= len(lines) {
			break
		}

		line := lines[textRow]
		lineRunes := []rune(line)
		spans := tokenizeBasicLine(line)

		for _, span := range spans {
			style := w.styleForKind(span.kind)

			for k := 0; k < span.ln; k++ {
				textCol := span.col + k

				// Preserva cursor — o TextArea já o renderizou
				if textRow == cursorRow && textCol == cursorCol {
					continue
				}

				visCol := textCol - offsetCol
				if visCol < 0 || visCol >= innerWidth {
					continue
				}
				if textCol >= len(lineRunes) {
					continue
				}

				putRune(screen, maxX, maxY, innerX+visCol, innerY+visRow, lineRunes[textCol], style)
			}
		}
	}
}

// ── Helpers de renderização ─────────────────────────────────────────────────

func putRune(screen tcell.Screen, maxX, maxY, x, y int, r rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= maxX || y >= maxY {
		return
	}
	screen.SetContent(x, y, r, nil, style)
}

func putString(screen tcell.Screen, maxX, maxY, x, y int, text string, style tcell.Style) {
	for i, r := range []rune(text) {
		putRune(screen, maxX, maxY, x+i, y, r, style)
	}
}
