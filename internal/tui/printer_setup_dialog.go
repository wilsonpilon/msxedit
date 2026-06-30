package tui

import (
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ── Layout ───────────────────────────────────────────────────────────────────
//
//  y+0   ╔══════════════ Printer Setup ═══════════════╗
//  y+1   ║                                            ║
//  y+2   ║  Paper size:  A4                           ║
//  y+3   ║                                            ║
//  y+4   ║  Word wrap column:     [   80]             ║
//  y+5   ║  Lines per page:       [   66]             ║
//  y+6   ║                                            ║
//  y+7   ║  [■] Print line numbers                    ║
//  y+8   ║                                            ║
//  y+9   ║  Font size:                                ║
//  y+10  ║  (●) 6pt   (●) 8pt   (●) 10pt  (●) 12pt  ║
//  y+11  ║                                            ║
//  y+12  ║  Font name: [Consolas                   ]▼ ║
//  y+13  ║             ┌─────────────────────────────┐  ← dropdown (when open)
//  y+14  ║        [  &Ok  ]  [&Cancel]  [ &Help ]    ║
//  y+15  ║        (shadow)   (shadow)   (shadow)      ║
//  y+16  ║                                            ║
//  y+17  ╚════════════════════════════════════════════╝

const (
	psDlgW = 58
	psDlgH = 18

	psRowPaper     = 2
	psRowWrapCol   = 4
	psRowLinesPage = 5
	psRowLineNums  = 7
	psRowFontLabel = 9
	psRowFontRadio = 10
	psRowFontName  = 12
	psRowButtons   = 14

	psNumInputX = 26
	psNumInputW = 6

	psFontNameLabelEnd = 14 // field starts at dialog_x+14
	psFontNameW        = 35 // inner width of font name input
	psFontArrowX       = 50 // psFontNameLabelEnd + psFontNameW + 1

	fontListMaxVis = 6 // max visible rows in the font dropdown
)

var (
	psFontLabels = [4]string{"6pt", "8pt", "10pt", "12pt"}
	psFontSizes  = [4]int{6, 8, 10, 12}
	psFontRadioX = [4]int{2, 12, 22, 32}
)

// monospacedFonts is the preset list for the font picker (Windows-centric).
var monospacedFonts = []string{
	"Cascadia Code",
	"Cascadia Mono",
	"Consolas",
	"Courier New",
	"Fixedsys",
	"Lucida Console",
	"OCR A Extended",
	"Terminal",
}

// psFilteredFonts returns fonts whose names contain query (case-insensitive).
// Returns the full list when query is empty or produces no match.
func psFilteredFonts(query string) []string {
	if query == "" {
		return monospacedFonts
	}
	q := strings.ToLower(query)
	var result []string
	for _, f := range monospacedFonts {
		if strings.Contains(strings.ToLower(f), q) {
			result = append(result, f)
		}
	}
	if len(result) == 0 {
		return monospacedFonts
	}
	return result
}

// ── Focus ─────────────────────────────────────────────────────────────────────

const (
	psFocusWrapCol   = 0
	psFocusLinesPage = 1
	psFocusLineNums  = 2
	psFocusFont6     = 3
	psFocusFont8     = 4
	psFocusFont10    = 5
	psFocusFont12    = 6
	psFocusFontName  = 7
	psFocusOk        = 8
	psFocusCancel    = 9
	psFocusHelp      = 10
	psFocusMax       = 10
)

// ── Struct ────────────────────────────────────────────────────────────────────

type printerSetupDialog struct {
	*tview.Box
	app      *App
	pageName string

	wrapColText   []rune
	wrapColCursor int

	linesPageText   []rune
	linesPageCursor int

	lineNumbers bool

	fontSizeIdx int // 0..3 → 6, 8, 10, 12 pt

	fontNameText   []rune
	fontNameCursor int

	focusField   int
	fontListOpen bool // true while font dropdown is visible

	btnOk     *turboButton
	btnCancel *turboButton
	btnHelp   *turboButton

	onClose func()
}

func newPrinterSetupDialog(app *App) *printerSetupDialog {
	defaultFont := "Consolas"
	switch runtime.GOOS {
	case "linux":
		defaultFont = "Monospace"
	case "darwin":
		defaultFont = "Menlo"
	}

	p := app.Printer

	wrapCol := []rune(strconv.Itoa(p.WrapColumn))
	if p.WrapColumn <= 0 {
		wrapCol = []rune("80")
	}
	linesPage := []rune(strconv.Itoa(p.LinesPerPage))
	if p.LinesPerPage <= 0 {
		linesPage = []rune("66")
	}
	fontSizeIdx := 2 // default 10pt
	for i, sz := range psFontSizes {
		if sz == p.FontSize {
			fontSizeIdx = i
			break
		}
	}
	fontName := p.FontName
	if fontName == "" {
		fontName = defaultFont
	}

	fnRunes := []rune(fontName)
	return &printerSetupDialog{
		Box:             tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:             app,
		pageName:        "printer_setup",
		wrapColText:     wrapCol,
		wrapColCursor:   len(wrapCol),
		linesPageText:   linesPage,
		linesPageCursor: len(linesPage),
		lineNumbers:     p.LineNumbers,
		fontSizeIdx:     fontSizeIdx,
		fontNameText:    fnRunes,
		fontNameCursor:  len(fnRunes),
		focusField:      psFocusWrapCol,
		btnOk:           newTurboButton("  &Ok  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnCancel:       newTurboButton("&Cancel", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnHelp:         newTurboButton(" &Help ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (d *printerSetupDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < psDlgW || height < psDlgH {
		return
	}
	maxX, maxY := screen.Size()

	bgStyle     := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)
	labelStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotStyle    := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)

	// Background
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	// Double border
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

	// Title
	title := " Printer Setup "
	putString(screen, maxX, maxY, x+(width-len([]rune(title)))/2, y, title, borderStyle)

	// [■] close button
	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	// ── Paper size ─────────────────────────────────────────────────────────────
	putString(screen, maxX, maxY, x+2, y+psRowPaper, "Paper size:  ", labelStyle)
	putString(screen, maxX, maxY, x+15, y+psRowPaper, "A4", tcell.StyleDefault.Foreground(vgaCyan).Background(d.app.Theme.PopupBg))

	// ── Numeric inputs ─────────────────────────────────────────────────────────
	putString(screen, maxX, maxY, x+2, y+psRowWrapCol, "Word wrap column:", labelStyle)
	d.drawNumInput(screen, maxX, maxY, x+psNumInputX, y+psRowWrapCol,
		d.wrapColText, d.wrapColCursor, d.focusField == psFocusWrapCol)

	putString(screen, maxX, maxY, x+2, y+psRowLinesPage, "Lines per page:", labelStyle)
	d.drawNumInput(screen, maxX, maxY, x+psNumInputX, y+psRowLinesPage,
		d.linesPageText, d.linesPageCursor, d.focusField == psFocusLinesPage)

	// ── Line numbers checkbox ──────────────────────────────────────────────────
	chMark := ' '
	if d.lineNumbers {
		chMark = '■'
	}
	chStyle := labelStyle
	if d.focusField == psFocusLineNums {
		chStyle = tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	}
	putString(screen, maxX, maxY, x+2, y+psRowLineNums, "["+string(chMark)+"] ", chStyle)
	putString(screen, maxX, maxY, x+6, y+psRowLineNums, "Print ", labelStyle)
	putRune(screen, maxX, maxY, x+12, y+psRowLineNums, 'l', hotStyle)
	putString(screen, maxX, maxY, x+13, y+psRowLineNums, "ine numbers", labelStyle)

	// ── Font size label + radios ───────────────────────────────────────────────
	putString(screen, maxX, maxY, x+2, y+psRowFontLabel, "Font ", labelStyle)
	putRune(screen, maxX, maxY, x+7, y+psRowFontLabel, 's', hotStyle)
	putString(screen, maxX, maxY, x+8, y+psRowFontLabel, "ize:", labelStyle)

	for i, lbl := range psFontLabels {
		rx := x + psFontRadioX[i]
		focused := d.focusField == psFocusFont6+i
		selected := d.fontSizeIdx == i
		mark := ' '
		if selected {
			mark = '●'
		}
		rs := labelStyle
		if focused {
			rs = tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
		}
		putString(screen, maxX, maxY, rx, y+psRowFontRadio, "("+string(mark)+") "+lbl, rs)
	}

	// ── Font name input + ▼ arrow ──────────────────────────────────────────────
	putString(screen, maxX, maxY, x+2, y+psRowFontName, "Font ", labelStyle)
	putRune(screen, maxX, maxY, x+7, y+psRowFontName, 'n', hotStyle)
	putString(screen, maxX, maxY, x+8, y+psRowFontName, "ame: ", labelStyle)

	d.drawTextInput(screen, maxX, maxY, x+psFontNameLabelEnd, y+psRowFontName,
		d.fontNameText, d.fontNameCursor, psFontNameW, d.focusField == psFocusFontName)

	// ▼ arrow button (vgaBlack on vgaGreen, same style as history arrows)
	arrowSt := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaGreen)
	putRune(screen, maxX, maxY, x+psFontArrowX, y+psRowFontName, '▼', arrowSt)

	// ── Buttons ────────────────────────────────────────────────────────────────
	d.drawButtons(screen, maxX, maxY, x, y, width)
}

func (d *printerSetupDialog) drawNumInput(screen tcell.Screen, maxX, maxY, x, y int, text []rune, cursor int, focused bool) {
	bg := vgaBlue
	if focused {
		bg = vgaLightBlue
	}
	st := tcell.StyleDefault.Foreground(vgaLightCyan).Background(bg)

	for col := 0; col < psNumInputW; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', st)
	}

	// Right-align the digits
	padded := make([]rune, psNumInputW)
	for i := range padded {
		padded[i] = ' '
	}
	start := psNumInputW - len(text)
	if start < 0 {
		start = 0
	}
	for i, r := range text {
		if start+i < psNumInputW {
			padded[start+i] = r
		}
	}
	for i, r := range padded {
		putRune(screen, maxX, maxY, x+i, y, r, st)
	}

	if focused {
		curCol := start + cursor
		if curCol >= 0 && curCol < psNumInputW {
			ch := padded[curCol]
			putRune(screen, maxX, maxY, x+curCol, y, ch,
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}
}

func (d *printerSetupDialog) drawTextInput(screen tcell.Screen, maxX, maxY, x, y int, text []rune, cursor, fieldW int, focused bool) {
	bg := vgaBlue
	if focused {
		bg = vgaLightBlue
	}
	st := tcell.StyleDefault.Foreground(vgaLightCyan).Background(bg)

	for col := 0; col < fieldW; col++ {
		putRune(screen, maxX, maxY, x+col, y, ' ', st)
	}

	offset := 0
	if cursor >= fieldW {
		offset = cursor - fieldW + 1
	}
	for i, r := range text {
		col := i - offset
		if col < 0 || col >= fieldW {
			continue
		}
		putRune(screen, maxX, maxY, x+col, y, r, st)
	}

	if focused {
		curCol := cursor - offset
		if curCol >= 0 && curCol < fieldW {
			ch := rune(' ')
			if cursor < len(text) {
				ch = text[cursor]
			}
			putRune(screen, maxX, maxY, x+curCol, y, ch,
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}
}

func (d *printerSetupDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y, width int) {
	btnW   := d.btnOk.width()
	total  := 3*btnW + 2*2
	startX := x + (width-total)/2
	btnY   := y + psRowButtons

	draw := func(btn *turboButton, focused bool, bx int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, bx, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, bx, btnY, d.app.Theme.PopupBg)
	}
	draw(d.btnOk, d.focusField == psFocusOk, startX)
	draw(d.btnCancel, d.focusField == psFocusCancel, startX+btnW+2)
	draw(d.btnHelp, d.focusField == psFocusHelp, startX+2*btnW+4)
}

// ── Font list popup ───────────────────────────────────────────────────────────

const psFontListPage = "printer_font_list"

type psFontListPopup struct {
	*tview.Box
	app      *App
	fonts    []string
	selected int
	scroll   int
	maxVis   int
	onSelect func(string)
	onCancel func()
}

func newPsFontListPopup(app *App, fonts []string, x, y int) *psFontListPopup {
	vis := fontListMaxVis
	if len(fonts) < vis {
		vis = len(fonts)
	}
	p := &psFontListPopup{
		Box:    tview.NewBox(),
		app:    app,
		fonts:  fonts,
		maxVis: vis,
	}
	// width = psFontNameW + 2 border cols; height = vis + 2 border rows
	p.Box.SetRect(x, y, psFontNameW+2, vis+2)
	return p
}

func (p *psFontListPopup) Draw(screen tcell.Screen) {
	x, y, w, h := p.GetRect()
	maxX, maxY := screen.Size()

	bgSt  := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)
	selSt := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	borSt := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)
	shSt  := tcell.StyleDefault.Background(vgaBlack)

	// Shadow (2 cols right, 1 row below)
	for row := 1; row < h+1; row++ {
		for col := 0; col < 2; col++ {
			putRune(screen, maxX, maxY, x+w+col, y+row, ' ', shSt)
		}
	}
	for col := 2; col < w+2; col++ {
		putRune(screen, maxX, maxY, x+col, y+h, ' ', shSt)
	}

	// Single border
	for col := 1; col < w-1; col++ {
		putRune(screen, maxX, maxY, x+col, y, '─', borSt)
		putRune(screen, maxX, maxY, x+col, y+h-1, '─', borSt)
	}
	for row := 1; row < h-1; row++ {
		putRune(screen, maxX, maxY, x, y+row, '│', borSt)
		putRune(screen, maxX, maxY, x+w-1, y+row, '│', borSt)
	}
	putRune(screen, maxX, maxY, x, y, '┌', borSt)
	putRune(screen, maxX, maxY, x+w-1, y, '┐', borSt)
	putRune(screen, maxX, maxY, x, y+h-1, '└', borSt)
	putRune(screen, maxX, maxY, x+w-1, y+h-1, '┘', borSt)

	// Items
	innerW := w - 2
	for i := 0; i < p.maxVis; i++ {
		idx := p.scroll + i
		if idx >= len(p.fonts) {
			break
		}
		st := bgSt
		if idx == p.selected {
			st = selSt
		}
		itemY := y + 1 + i
		for col := 0; col < innerW; col++ {
			putRune(screen, maxX, maxY, x+1+col, itemY, ' ', st)
		}
		runes := []rune(p.fonts[idx])
		for col, r := range runes {
			if col >= innerW {
				break
			}
			putRune(screen, maxX, maxY, x+1+col, itemY, r, st)
		}
	}
}

func (p *psFontListPopup) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if p.selected > 0 {
				p.selected--
				if p.selected < p.scroll {
					p.scroll = p.selected
				}
			}
		case tcell.KeyDown:
			if p.selected < len(p.fonts)-1 {
				p.selected++
				if p.selected >= p.scroll+p.maxVis {
					p.scroll = p.selected - p.maxVis + 1
				}
			}
		case tcell.KeyEnter:
			if p.onSelect != nil && p.selected < len(p.fonts) {
				p.onSelect(p.fonts[p.selected])
			}
		case tcell.KeyEscape:
			if p.onCancel != nil {
				p.onCancel()
			}
		case tcell.KeyRune:
			// Any printable char closes the list and returns focus to the dialog
			if p.onCancel != nil {
				p.onCancel()
			}
		}
	})
}

func (p *psFontListPopup) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, w, h := p.GetRect()
		// Click outside → cancel
		if mx < x || mx >= x+w || my < y || my >= y+h {
			if action == tview.MouseLeftClick {
				if p.onCancel != nil {
					p.onCancel()
				}
			}
			return false, nil
		}
		if action == tview.MouseLeftClick {
			row := my - y - 1
			idx := p.scroll + row
			if idx >= 0 && idx < len(p.fonts) {
				p.selected = idx
				if p.onSelect != nil {
					p.onSelect(p.fonts[p.selected])
				}
			}
		}
		return true, nil
	}
}

func (p *psFontListPopup) PasteHandler() func(string, func(tview.Primitive)) {
	return func(string, func(tview.Primitive)) {}
}

// ── Open font list popup ──────────────────────────────────────────────────────

func (d *printerSetupDialog) openFontListPopup() {
	if d.fontListOpen {
		return
	}
	d.fontListOpen = true

	dx, dy, _, _ := d.GetRect()
	popX := dx + psFontNameLabelEnd
	popY := dy + psRowFontName + 1

	query := strings.TrimSpace(string(d.fontNameText))
	fonts := psFilteredFonts(query)

	// Pre-select the exact match if any, otherwise index 0
	presel := 0
	for i, f := range fonts {
		if strings.EqualFold(f, query) {
			presel = i
			break
		}
	}

	popup := newPsFontListPopup(d.app, fonts, popX, popY)
	popup.selected = presel
	if presel >= fontListMaxVis {
		popup.scroll = presel - fontListMaxVis + 1
	}

	popup.onSelect = func(name string) {
		d.fontListOpen = false
		d.fontNameText = []rune(name)
		d.fontNameCursor = len(d.fontNameText)
		d.app.Pages.RemovePage(psFontListPage)
		d.app.Application.SetFocus(d)
	}
	popup.onCancel = func() {
		d.fontListOpen = false
		d.app.Pages.RemovePage(psFontListPage)
		d.app.Application.SetFocus(d)
	}

	d.app.Pages.AddPage(psFontListPage, popup, false, true)
	d.app.Application.SetFocus(popup)
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *printerSetupDialog) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if event == nil {
			return
		}

		// Font name field: Up/Down/Enter opens the dropdown list
		if d.focusField == psFocusFontName {
			switch event.Key() {
			case tcell.KeyUp, tcell.KeyDown, tcell.KeyEnter:
				d.openFontListPopup()
				return
			}
		}

		// Field-specific key handling
		switch d.focusField {
		case psFocusWrapCol:
			if d.handleNumKey(event, &d.wrapColText, &d.wrapColCursor) {
				return
			}
		case psFocusLinesPage:
			if d.handleNumKey(event, &d.linesPageText, &d.linesPageCursor) {
				return
			}
		case psFocusFontName:
			if d.handleTextKey(event, &d.fontNameText, &d.fontNameCursor) {
				return
			}
		}

		// Global navigation
		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
		case tcell.KeyTab:
			d.focusField = (d.focusField + 1) % (psFocusMax + 1)
		case tcell.KeyBacktab:
			d.focusField = (d.focusField - 1 + psFocusMax + 1) % (psFocusMax + 1)
		case tcell.KeyLeft:
			switch {
			case d.focusField >= psFocusFont6 && d.focusField <= psFocusFont12:
				if d.focusField > psFocusFont6 {
					d.focusField--
				}
			case d.focusField >= psFocusOk && d.focusField <= psFocusHelp:
				if d.focusField > psFocusOk {
					d.focusField--
				}
			}
		case tcell.KeyRight:
			switch {
			case d.focusField >= psFocusFont6 && d.focusField <= psFocusFont12:
				if d.focusField < psFocusFont12 {
					d.focusField++
				}
			case d.focusField >= psFocusOk && d.focusField <= psFocusHelp:
				if d.focusField < psFocusHelp {
					d.focusField++
				}
			}
		case tcell.KeyEnter:
			d.activateFocused()
		case tcell.KeyRune:
			switch unicode.ToLower(event.Rune()) {
			case 'o':
				d.applySettings()
				d.close()
			case ' ':
				d.activateFocused()
			}
		}
	})
}

func (d *printerSetupDialog) handleNumKey(event *tcell.EventKey, text *[]rune, cursor *int) bool {
	t := *text
	c := *cursor
	switch event.Key() {
	case tcell.KeyLeft:
		if c > 0 {
			*cursor = c - 1
		}
		return true
	case tcell.KeyRight:
		if c < len(t) {
			*cursor = c + 1
		}
		return true
	case tcell.KeyHome:
		*cursor = 0
		return true
	case tcell.KeyEnd:
		*cursor = len(t)
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if c > 0 {
			*text = append(t[:c-1], t[c:]...)
			*cursor = c - 1
		}
		return true
	case tcell.KeyDelete:
		if c < len(t) {
			*text = append(t[:c], t[c+1:]...)
		}
		return true
	case tcell.KeyRune:
		r := event.Rune()
		if r >= '0' && r <= '9' && len(t) < psNumInputW {
			*text = append(t[:c], append([]rune{r}, t[c:]...)...)
			*cursor = c + 1
			return true
		}
		return true // swallow non-digits
	}
	return false
}

func (d *printerSetupDialog) handleTextKey(event *tcell.EventKey, text *[]rune, cursor *int) bool {
	t := *text
	c := *cursor
	switch event.Key() {
	case tcell.KeyLeft:
		if c > 0 {
			*cursor = c - 1
		}
		return true
	case tcell.KeyRight:
		if c < len(t) {
			*cursor = c + 1
		}
		return true
	case tcell.KeyHome:
		*cursor = 0
		return true
	case tcell.KeyEnd:
		*cursor = len(t)
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if c > 0 {
			*text = append(t[:c-1], t[c:]...)
			*cursor = c - 1
		}
		return true
	case tcell.KeyDelete:
		if c < len(t) {
			*text = append(t[:c], t[c+1:]...)
		}
		return true
	case tcell.KeyRune:
		r := event.Rune()
		if unicode.IsPrint(r) && len(t) < psFontNameW {
			*text = append(t[:c], append([]rune{r}, t[c:]...)...)
			*cursor = c + 1
			return true
		}
	}
	return false
}

func (d *printerSetupDialog) activateFocused() {
	switch {
	case d.focusField == psFocusLineNums:
		d.lineNumbers = !d.lineNumbers
	case d.focusField >= psFocusFont6 && d.focusField <= psFocusFont12:
		d.fontSizeIdx = d.focusField - psFocusFont6
	case d.focusField == psFocusFontName:
		d.openFontListPopup()
	case d.focusField == psFocusOk:
		d.applySettings()
		d.close()
	case d.focusField == psFocusCancel:
		d.close()
	}
}

func (d *printerSetupDialog) applySettings() {
	wrapStr := strings.TrimSpace(string(d.wrapColText))
	wrapCol, _ := strconv.Atoi(wrapStr)
	if wrapCol <= 0 {
		wrapCol = 80
	}

	linesStr := strings.TrimSpace(string(d.linesPageText))
	linesPage, _ := strconv.Atoi(linesStr)
	if linesPage <= 0 {
		linesPage = 66
	}

	d.app.Printer = PrinterConfig{
		WrapColumn:   wrapCol,
		LinesPerPage: linesPage,
		LineNumbers:  d.lineNumbers,
		FontSize:     psFontSizes[d.fontSizeIdx],
		FontName:     strings.TrimSpace(string(d.fontNameText)),
	}
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *printerSetupDialog) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := d.GetRect()
		if mx < x || mx >= x+width || my < y || my >= y+height {
			return false, nil
		}
		if action != tview.MouseLeftClick {
			return true, nil
		}
		setFocus(d)

		// [■] close
		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		// Numeric inputs
		if my == y+psRowWrapCol && mx >= x+psNumInputX && mx < x+psNumInputX+psNumInputW {
			d.focusField = psFocusWrapCol
			return true, nil
		}
		if my == y+psRowLinesPage && mx >= x+psNumInputX && mx < x+psNumInputX+psNumInputW {
			d.focusField = psFocusLinesPage
			return true, nil
		}

		// Checkbox
		if my == y+psRowLineNums && mx >= x+2 && mx <= x+5 {
			d.focusField = psFocusLineNums
			d.lineNumbers = !d.lineNumbers
			return true, nil
		}

		// Font size radios
		if my == y+psRowFontRadio {
			for i, rx := range psFontRadioX {
				end := x + rx + 4 + len([]rune(psFontLabels[i]))
				if mx >= x+rx && mx < end {
					d.focusField = psFocusFont6 + i
					d.fontSizeIdx = i
					return true, nil
				}
			}
		}

		// Font name input
		if my == y+psRowFontName && mx >= x+psFontNameLabelEnd && mx < x+psFontNameLabelEnd+psFontNameW {
			d.focusField = psFocusFontName
			offset := 0
			if d.fontNameCursor >= psFontNameW {
				offset = d.fontNameCursor - psFontNameW + 1
			}
			pos := (mx - (x + psFontNameLabelEnd)) + offset
			if pos < 0 {
				pos = 0
			}
			if pos > len(d.fontNameText) {
				pos = len(d.fontNameText)
			}
			d.fontNameCursor = pos
			return true, nil
		}

		// ▼ arrow button → open font list
		if my == y+psRowFontName && mx == x+psFontArrowX {
			d.focusField = psFocusFontName
			d.openFontListPopup()
			return true, nil
		}

		// Buttons
		btnW   := d.btnOk.width()
		total  := 3*btnW + 4
		startX := x + (width-total)/2
		btnY   := y + psRowButtons
		if my == btnY {
			switch {
			case mx >= startX && mx < startX+btnW:
				d.focusField = psFocusOk
				d.activateFocused()
			case mx >= startX+btnW+2 && mx < startX+2*btnW+2:
				d.focusField = psFocusCancel
				d.activateFocused()
			case mx >= startX+2*btnW+4 && mx < startX+3*btnW+4:
				d.focusField = psFocusHelp
				d.activateFocused()
			}
		}
		return true, nil
	}
}

func (d *printerSetupDialog) PasteHandler() func(string, func(tview.Primitive)) {
	return func(text string, setFocus func(tview.Primitive)) {
		if d.focusField != psFocusFontName {
			return
		}
		for _, r := range text {
			if unicode.IsPrint(r) && len(d.fontNameText) < psFontNameW {
				d.fontNameText = append(d.fontNameText[:d.fontNameCursor],
					append([]rune{r}, d.fontNameText[d.fontNameCursor:]...)...)
				d.fontNameCursor++
			}
		}
	}
}

func (d *printerSetupDialog) close() {
	if d.fontListOpen {
		d.fontListOpen = false
		d.app.Pages.RemovePage(psFontListPage)
	}
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

// ── Show ──────────────────────────────────────────────────────────────────────

func showPrinterSetupDialog(app *App) {
	dlg := newPrinterSetupDialog(app)
	dlg.onClose = func() {
		if len(app.Editors) > 0 && app.ActiveEditor >= 0 && app.ActiveEditor < len(app.Editors) {
			app.Application.SetFocus(app.Editors[app.ActiveEditor])
			return
		}
		if app.Editor != nil {
			app.Application.SetFocus(app.Editor)
		}
	}

	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(dlg, psDlgW, 0, true).
			AddItem(nil, 0, 1, false), psDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	app.Pages.AddPage(dlg.pageName, container, true, true)
	app.Application.SetFocus(dlg)
}
