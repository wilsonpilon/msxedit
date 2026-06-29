package reader

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// keysHelp: apenas navegação essencial, alinhado à direita na barra de status.
const keysHelp = "Keys: ↑↓←→ PgUp PgDn  ESC=Exit  F1=Help"

// visualRow mapeia uma linha visual (modo wrap) para uma posição no documento.
type visualRow struct {
	line int
	col  int
	len  int // número de runes a exibir nesta linha visual
}

// Viewer é a primitiva de tela cheia do msxread.
type Viewer struct {
	*tview.Box
	theme      Theme
	doc        *Document
	offsetY    int
	offsetX    int
	showHelp   bool
	onExit     func()
	bodyHeight int
	bodyWidth  int

	// Cores dinâmicas do corpo (índices VGA 0-15)
	bodyFgIdx int
	bodyBgIdx int

	// Modo wrap
	wrapMode   bool
	visualRows []visualRow
	prevBodyW  int // última largura usada para construir visualRows

	// Hi-bit: exibir bytes 128-255 como-estão (true) ou como '·' (false)
	hiBit bool

	// Mensagem temporária na barra de status
	statusMsg     string
	statusClearAt time.Time

	// Busca (F / N / C)
	findMode          bool
	findQuery         string
	findCaseSensitive bool
	matchLine         int // -1 = sem match
	matchCol          int
	matchLen          int // largura em runas
}

// NewViewer cria o visualizador com as configurações iniciais.
func NewViewer(theme Theme, doc *Document, s Settings) *Viewer {
	return &Viewer{
		Box:       tview.NewBox(),
		theme:     theme,
		doc:       doc,
		bodyFgIdx: s.BodyFg,
		bodyBgIdx: s.BodyBg,
		wrapMode:  s.WrapMode,
		hiBit:     s.HiBit,
		matchLine: -1,
	}
}

func (v *Viewer) SetOnExit(fn func()) { v.onExit = fn }

func (v *Viewer) Draw(screen tcell.Screen) {
	x, y, width, height := v.GetRect()
	if width < 4 || height < 3 {
		return
	}
	maxX, maxY := screen.Size()

	bodyFg := vgaPalette[v.bodyFgIdx]
	bodyBg := vgaPalette[v.bodyBgIdx]

	topStyle    := tcell.StyleDefault.Foreground(v.theme.TopFg).Background(v.theme.TopBg)
	bodyStyle   := tcell.StyleDefault.Foreground(bodyFg).Background(bodyBg)
	headStyle   := tcell.StyleDefault.Foreground(v.theme.HeadingFg).Background(bodyBg).Bold(true)
	statusStyle := tcell.StyleDefault.Foreground(v.theme.StatusFg).Background(v.theme.StatusBg)

	v.bodyHeight = height - 2
	v.bodyWidth  = width

	// Wrap: recalcula a cada Draw com a largura atual da janela.
	// Usa width-2 (1 col margem esq. + 1 col de segurança à dir.)
	if v.wrapMode {
		wrapW := width - 2
		if wrapW < 10 {
			wrapW = 10
		}
		if v.prevBodyW != wrapW || len(v.visualRows) == 0 {
			v.visualRows = buildVisualRows(v.doc.Lines, wrapW)
			v.prevBodyW  = wrapW
		}
	}
	v.clampOffsets()

	// ── Barra de topo ────────────────────────────────────────────────────────
	fillRow(screen, maxX, maxY, x, y, width, topStyle)
	now := time.Now()
	header := fmt.Sprintf(" %s  %c  %s  %c  %s",
		now.Format("02-01-2006"), sepGlyph,
		now.Format("15:04:05"), sepGlyph,
		v.doc.FileName)
	putString(screen, maxX, maxY, x, y, header, topStyle)

	// ── Corpo ────────────────────────────────────────────────────────────────
	for row := 0; row < v.bodyHeight; row++ {
		bodyY := y + 1 + row
		fillRow(screen, maxX, maxY, x, bodyY, width, bodyStyle)

		var textRow, startCol, rowLen int
		rowLen = -1 // -1 = sem limite (só restringe pelo fim da linha)
		if v.wrapMode && len(v.visualRows) > 0 {
			vIdx := v.offsetY + row
			if vIdx >= len(v.visualRows) {
				continue
			}
			vr       := v.visualRows[vIdx]
			textRow   = vr.line
			startCol  = vr.col
			rowLen    = vr.len
		} else {
			textRow  = v.offsetY + row
			startCol = v.offsetX
		}
		if textRow >= len(v.doc.Lines) {
			continue
		}

		lineStyle := bodyStyle
		if v.doc.Headings[textRow] {
			lineStyle = headStyle
		}
		var lineSpans []Span
		if textRow < len(v.doc.Spans) {
			lineSpans = v.doc.Spans[textRow]
		}

		runes := []rune(v.doc.Lines[textRow])
		for col := 0; col < width-1; col++ {
			if rowLen >= 0 && col >= rowLen {
				break
			}
			srcCol := startCol + col
			if srcCol < 0 || srcCol >= len(runes) {
				continue
			}
			r := runes[srcCol]
			if !v.hiBit && r >= 128 && r <= 255 {
				r = '·'
			}
			style := v.spanStyleAt(lineStyle, lineSpans, textRow, srcCol)
			putRune(screen, maxX, maxY, x+1+col, bodyY, r, style)
		}
	}

	// ── Barra de status ──────────────────────────────────────────────────────
	statusY := y + height - 1
	fillRow(screen, maxX, maxY, x, statusY, width, statusStyle)

	var statusText string
	switch {
	case v.statusMsg != "" && time.Now().Before(v.statusClearAt):
		statusText = " " + v.statusMsg

	case v.findMode:
		caseFlag := ""
		if v.findCaseSensitive {
			caseFlag = " [Aa]"
		}
		statusText = fmt.Sprintf(" Find►  %s_%s", v.findQuery, caseFlag)

	default:
		left  := fmt.Sprintf(" Command►  %s", v.positionLabel())
		right := keysHelp
		pad   := width - len([]rune(left)) - len([]rune(right)) - 1
		if pad < 1 {
			pad = 1
		}
		statusText = left + strings.Repeat(" ", pad) + right
	}
	putString(screen, maxX, maxY, x, statusY, statusText, statusStyle)

	// ── Overlay F1 ───────────────────────────────────────────────────────────
	if v.showHelp {
		v.drawHelpOverlay(screen, maxX, maxY, x, y, width, height)
	}
}

// ── Posição e limites ────────────────────────────────────────────────────────

func (v *Viewer) positionLabel() string {
	if v.offsetY <= 0 {
		return "*** Top of File ***"
	}
	if v.offsetY >= v.maxOffsetY() {
		return "*** End of File ***"
	}
	total := len(v.doc.Lines)
	docLine := v.offsetY
	if v.wrapMode && len(v.visualRows) > 0 && v.offsetY < len(v.visualRows) {
		docLine = v.visualRows[v.offsetY].line
	}
	return fmt.Sprintf("Linha %d de %d", docLine+1, total)
}

func (v *Viewer) maxOffsetY() int {
	total := len(v.doc.Lines)
	if v.wrapMode && len(v.visualRows) > 0 {
		total = len(v.visualRows)
	}
	m := total - v.bodyHeight
	if m < 0 {
		return 0
	}
	return m
}

func (v *Viewer) clampOffsets() {
	if v.offsetY > v.maxOffsetY() {
		v.offsetY = v.maxOffsetY()
	}
	if v.offsetY < 0 {
		v.offsetY = 0
	}
	if v.offsetX < 0 {
		v.offsetX = 0
	}
}

func buildVisualRows(lines []string, bodyWidth int) []visualRow {
	if bodyWidth <= 0 {
		bodyWidth = 79
	}
	var rows []visualRow
	for i, l := range lines {
		runes := []rune(l)
		if len(runes) == 0 {
			rows = append(rows, visualRow{i, 0, 0})
			continue
		}
		start := 0
		for start < len(runes) {
			remaining := runes[start:]
			if len(remaining) <= bodyWidth {
				rows = append(rows, visualRow{i, start, len(remaining)})
				break
			}
			// Procura o último espaço dentro do limite
			breakAt := -1
			for k := bodyWidth - 1; k > 0; k-- {
				if remaining[k] == ' ' {
					breakAt = k
					break
				}
			}
			if breakAt < 0 {
				// Palavra maior que a linha: quebra forçada no limite
				rows = append(rows, visualRow{i, start, bodyWidth})
				start += bodyWidth
			} else {
				// Exibe até o espaço (exclusive); próxima linha começa após os espaços
				rows = append(rows, visualRow{i, start, breakAt})
				start += breakAt
				for start < len(runes) && runes[start] == ' ' {
					start++
				}
			}
		}
	}
	return rows
}

// ── Navegação ────────────────────────────────────────────────────────────────

func (v *Viewer) ScrollLines(d int) { v.offsetY += d; v.clampOffsets() }

func (v *Viewer) ScrollCols(d int) {
	if v.wrapMode {
		return
	}
	v.offsetX += d
	if v.offsetX < 0 {
		v.offsetX = 0
	}
}

func (v *Viewer) PageDown() { v.offsetY += v.pageStep(); v.clampOffsets() }
func (v *Viewer) PageUp()   { v.offsetY -= v.pageStep(); v.clampOffsets() }
func (v *Viewer) Home()     { v.offsetY = 0; v.offsetX = 0 }
func (v *Viewer) End()      { v.offsetY = v.maxOffsetY() }

func (v *Viewer) pageStep() int {
	if v.bodyHeight > 1 {
		return v.bodyHeight - 1
	}
	return 1
}

// ── Teclas de função ─────────────────────────────────────────────────────────

func (v *Viewer) CycleFg(d int) {
	v.bodyFgIdx = (v.bodyFgIdx + d + 16) % 16
	v.ShowStatus(fmt.Sprintf("Texto: %d – %s", v.bodyFgIdx, vgaColorName(v.bodyFgIdx)), 2*time.Second)
}

func (v *Viewer) CycleBg(d int) {
	v.bodyBgIdx = (v.bodyBgIdx + d + 16) % 16
	v.ShowStatus(fmt.Sprintf("Fundo: %d – %s", v.bodyBgIdx, vgaColorName(v.bodyBgIdx)), 2*time.Second)
}

func (v *Viewer) ToggleWrap() {
	v.wrapMode = !v.wrapMode
	if v.wrapMode {
		v.offsetX    = 0
		v.prevBodyW  = 0 // força reconstrução no próximo Draw
		v.ShowStatus("Wrap: Ativo", 2*time.Second)
	} else {
		v.visualRows = nil
		v.ShowStatus("Wrap: Inativo", 2*time.Second)
	}
	v.clampOffsets()
}

func (v *Viewer) SetHiBit(on bool) {
	v.hiBit = on
	if on {
		v.ShowStatus("Hi-bit: Ativo (8 bits)", 2*time.Second)
	} else {
		v.ShowStatus("Hi-bit: Inativo (7 bits)", 2*time.Second)
	}
}

func (v *Viewer) SaveCurrentSettings() {
	s := Settings{
		BodyFg:   v.bodyFgIdx,
		BodyBg:   v.bodyBgIdx,
		WrapMode: v.wrapMode,
		HiBit:    v.hiBit,
	}
	err := SaveSettings(s)
	if err != nil {
		v.ShowStatus("Erro ao salvar: "+err.Error(), 4*time.Second)
	} else {
		v.ShowStatus("Configurações salvas em "+settingsPath(), 4*time.Second)
	}
}

func (v *Viewer) PrintDoc(onDone func(err error)) {
	v.ShowStatus("Imprimindo…", 2*time.Second)
	go func() {
		_, err := PrintLines(v.doc.FileName, v.doc.Lines)
		onDone(err)
	}()
}

func (v *Viewer) ShowStatus(msg string, d time.Duration) {
	v.statusMsg     = msg
	v.statusClearAt = time.Now().Add(d)
}

// ── Busca (F / N / C) ────────────────────────────────────────────────────────

func (v *Viewer) InFindMode() bool { return v.findMode }

func (v *Viewer) StartFind() {
	v.findMode  = true
	v.findQuery = ""
}

func (v *Viewer) AddFindChar(r rune) {
	v.findQuery += string(r)
	v.findFirst()
}

func (v *Viewer) FindBackspace() {
	runes := []rune(v.findQuery)
	if len(runes) > 0 {
		v.findQuery = string(runes[:len(runes)-1])
		v.findFirst()
	}
}

func (v *Viewer) CommitFind() {
	v.findMode = false
	if v.matchLine < 0 && v.findQuery != "" {
		v.ShowStatus("Não encontrado: "+v.findQuery, 2*time.Second)
	}
}

func (v *Viewer) CancelFind() {
	v.findMode  = false
	v.matchLine = -1
	v.findQuery = ""
}

// FindNext busca a próxima ocorrência após o match atual.
func (v *Viewer) FindNext() {
	if v.findQuery == "" {
		return
	}
	if v.matchLine < 0 {
		v.findFirst()
		return
	}
	qLen := len([]rune(v.findQuery))

	// Mesma linha, após o match atual
	if col := v.searchInLine(v.matchLine, v.matchCol+1); col >= 0 {
		v.matchCol = col
		v.matchLen = qLen
		v.scrollToMatch()
		return
	}
	// Linhas seguintes
	for i := v.matchLine + 1; i < len(v.doc.Lines); i++ {
		if col := v.searchInLine(i, 0); col >= 0 {
			v.matchLine = i
			v.matchCol  = col
			v.matchLen  = qLen
			v.scrollToMatch()
			return
		}
	}
	// Volta ao início (wrap around)
	for i := 0; i <= v.matchLine; i++ {
		start := 0
		if i == v.matchLine {
			start = v.matchCol + 1
		}
		if col := v.searchInLine(i, start); col >= 0 {
			v.matchLine = i
			v.matchCol  = col
			v.matchLen  = qLen
			v.scrollToMatch()
			v.ShowStatus("Busca recomeçou do início", 1500*time.Millisecond)
			return
		}
	}
	v.ShowStatus("Não encontrado: "+v.findQuery, 2*time.Second)
}

// ToggleCaseSensitive alterna entre busca sensível/insensível a maiúsculas.
func (v *Viewer) ToggleCaseSensitive() {
	v.findCaseSensitive = !v.findCaseSensitive
	if v.findCaseSensitive {
		v.ShowStatus("Busca: diferencia maiúsculas/minúsculas", 2*time.Second)
	} else {
		v.ShowStatus("Busca: ignora maiúsculas/minúsculas", 2*time.Second)
	}
	if v.findQuery != "" {
		v.findFirst()
	}
}

// findFirst busca a primeira ocorrência desde o início do documento.
func (v *Viewer) findFirst() {
	if v.findQuery == "" {
		v.matchLine = -1
		return
	}
	qLen := len([]rune(v.findQuery))
	for i := range v.doc.Lines {
		if col := v.searchInLine(i, 0); col >= 0 {
			v.matchLine = i
			v.matchCol  = col
			v.matchLen  = qLen
			v.scrollToMatch()
			return
		}
	}
	v.matchLine = -1
}

// searchInLine retorna a posição de runa da query em doc.Lines[lineIdx]
// a partir de startRuneCol, ou -1 se não encontrada.
func (v *Viewer) searchInLine(lineIdx, startRuneCol int) int {
	if lineIdx < 0 || lineIdx >= len(v.doc.Lines) {
		return -1
	}
	line  := []rune(v.doc.Lines[lineIdx])
	query := []rune(v.findQuery)
	if !v.findCaseSensitive {
		line  = []rune(strings.ToLower(string(line)))
		query = []rune(strings.ToLower(string(query)))
	}
	qLen := len(query)
	if qLen == 0 || startRuneCol > len(line)-qLen {
		return -1
	}
	for i := startRuneCol; i <= len(line)-qLen; i++ {
		match := true
		for j := 0; j < qLen; j++ {
			if line[i+j] != query[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// scrollToMatch ajusta offsetY para que o match fique visível.
func (v *Viewer) scrollToMatch() {
	if v.matchLine < 0 {
		return
	}
	target := v.matchLine
	if v.wrapMode && len(v.visualRows) > 0 {
		for i, vr := range v.visualRows {
			if vr.line == v.matchLine && vr.col <= v.matchCol {
				target = i
			} else if vr.line > v.matchLine {
				break
			}
		}
	}
	if target < v.offsetY || target >= v.offsetY+v.bodyHeight {
		v.offsetY = target - v.bodyHeight/3
	}
	v.clampOffsets()
}

// ── Help ─────────────────────────────────────────────────────────────────────

func (v *Viewer) ToggleHelp() { v.showHelp = !v.showHelp }
func (v *Viewer) IsHelpOpen() bool { return v.showHelp }
func (v *Viewer) CloseHelp()  { v.showHelp = false }

func (v *Viewer) drawHelpOverlay(screen tcell.Screen, maxX, maxY, x, y, width, height int) {
	lines := []string{
		"  msxread  –  Teclas de atalho   ",
		"",
		"  F             Buscar texto",
		"  N             Próxima ocorrência",
		"  C             Alternar maiúsc./minúsc.",
		"",
		"  ↑ ↓          Linha acima / abaixo",
		"  ← →          Coluna  (±8 cols)",
		"  PgUp  PgDn   Página anterior / próxima",
		"  Home          Início do arquivo",
		"  End           Final do arquivo",
		"",
		"  W             Quebra de linha (wrap)",
		"  7             Hi-bit inativo (7 bits)",
		"  8             Hi-bit ativo   (8 bits)",
		"  P             Imprimir",
		"  S             Salvar configurações",
		"",
		"  F5 / F6       Cor do texto  ◄ ►",
		"  F7 / F8       Cor do fundo  ◄ ►",
		"",
		"  F1            Esta ajuda",
		"  ESC / Q       Sair",
		"",
		"  Qualquer tecla para fechar",
	}
	bw := 42
	bh := len(lines) + 2
	bx := x + (width-bw)/2
	by := y + (height-bh)/2

	boxStyle    := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)
	borderStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)

	for row := 0; row < bh; row++ {
		fillRow(screen, maxX, maxY, bx, by+row, bw, boxStyle)
	}
	for col := 0; col < bw; col++ {
		putRune(screen, maxX, maxY, bx+col, by, '─', borderStyle)
		putRune(screen, maxX, maxY, bx+col, by+bh-1, '─', borderStyle)
	}
	for row := 0; row < bh; row++ {
		putRune(screen, maxX, maxY, bx, by+row, '│', borderStyle)
		putRune(screen, maxX, maxY, bx+bw-1, by+row, '│', borderStyle)
	}
	putRune(screen, maxX, maxY, bx, by, '┌', borderStyle)
	putRune(screen, maxX, maxY, bx+bw-1, by, '┐', borderStyle)
	putRune(screen, maxX, maxY, bx, by+bh-1, '└', borderStyle)
	putRune(screen, maxX, maxY, bx+bw-1, by+bh-1, '┘', borderStyle)

	for i, l := range lines {
		putString(screen, maxX, maxY, bx+1, by+1+i, l, boxStyle)
	}
}

// ── spanStyleAt ──────────────────────────────────────────────────────────────

// spanStyleAt devolve o estilo para a coluna col na linha docLine,
// dando prioridade ao highlight de match sobre os spans de sintaxe.
func (v *Viewer) spanStyleAt(base tcell.Style, spans []Span, docLine, col int) tcell.Style {
	if v.matchLine >= 0 && docLine == v.matchLine &&
		col >= v.matchCol && col < v.matchCol+v.matchLen {
		return tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	}
	for _, sp := range spans {
		if col >= sp.Start && col < sp.End {
			switch sp.Kind {
			case SpanKeyword:
				return base.Foreground(v.theme.KeywordFg)
			case SpanString:
				return base.Foreground(v.theme.StringFg)
			case SpanComment:
				return base.Foreground(v.theme.CommentFg)
			}
		}
	}
	return base
}

// ── MouseHandler ─────────────────────────────────────────────────────────────

func (v *Viewer) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return v.WrapMouseHandler(func(action tview.MouseAction, _ *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		switch action {
		case tview.MouseScrollUp:
			v.ScrollLines(-1)
			return true, nil
		case tview.MouseScrollDown:
			v.ScrollLines(1)
			return true, nil
		}
		return false, nil
	})
}

// ── Helpers de renderização ──────────────────────────────────────────────────

func fillRow(screen tcell.Screen, maxX, maxY, x, y, width int, style tcell.Style) {
	for col := 0; col < width; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', style)
	}
}

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
