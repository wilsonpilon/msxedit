package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"msxedit/internal/basic"
)

// Índices de foco
const (
	saveFileFocusName   = 0
	saveFileFocusFiles  = 1
	saveFileFocusDirs   = 2
	saveFileFocusOk     = 3 // "&Ok" (não-MSX) ou "&ASCII" (MSX-BASIC)
	saveFileFocusToken  = 4 // "&Token" — só em modo MSX-BASIC
	saveFileFocusCancel = 5
	saveFileFocusHelp   = 6
	saveFileFocusMax    = 6
)

// Layout idêntico ao Open File dialog
const (
	saveFileDlgW = 68
	saveFileDlgH = 19

	saveFileInputX = 2
	saveFileInputW = 44
	saveFileArrowX = 46

	saveFileListX    = 2
	saveFileListXEnd = 47
	saveFileListRows = 8
	saveFileFileColW = 14
	saveFileDivOff   = 28

	saveFileBtnX          = 52
	saveFileBtnRowOk      = 3
	saveFileBtnRowToken   = 5
	saveFileBtnRowCancel  = 11
	saveFileBtnRowHelp    = 13

	saveFileRowLabel   = 2
	saveFileRowInput   = 3
	saveFileRowFiles   = 5
	saveFileRowListTop = 6
	saveFileRowScroll  = 14
	saveFileRowStat1   = 16
	saveFileRowStat2   = 17
)

type saveFileDialog struct {
	*tview.Box
	app      *App
	pageName string

	nameText   []rune
	nameCursor int

	history     []string
	showHistory bool
	historyIdx  int

	currentDir   string
	mask         string
	files        []openFileFEntry
	dirs         []string
	fileOffset   int
	selectedFile int
	selectedDir  int

	focusField int
	msxMode    bool // true quando radio MSX-BASIC está selecionado

	btnOk     *turboButton // "&Ok" ou "&ASCII"
	btnToken  *turboButton // "&Token" (só MSX-BASIC)
	btnCancel *turboButton
	btnHelp   *turboButton

	onSave  func(path string, tokenized bool)
	onClose func()
}

func newSaveFileDialog(app *App, initialName string) *saveFileDialog {
	if initialName == "" {
		initialName = "*.BAS"
	}
	cwd, _ := os.Getwd()
	msxMode := app.CompilerMode == 0 // índice 0 = MSX-BASIC

	var btnOk *turboButton
	if msxMode {
		btnOk = newTurboButton(" &ASCII ", vgaWhite, vgaGreen, vgaYellow, vgaBlack)
	} else {
		btnOk = newTurboButton("  &Ok   ", vgaWhite, vgaGreen, vgaYellow, vgaBlack)
	}

	d := &saveFileDialog{
		Box:          tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:          app,
		pageName:     "save_file",
		nameText:     []rune(initialName),
		nameCursor:   len([]rune(initialName)),
		history:      []string{"*.BAS", "*.bas", "*.TXT"},
		currentDir:   cwd,
		mask:         initialName,
		selectedFile: -1,
		selectedDir:  -1,
		focusField:   saveFileFocusFiles,
		msxMode:      msxMode,
		btnOk:        btnOk,
		btnToken:     newTurboButton(" &Token ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnCancel:    newTurboButton("Cancel ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnHelp:      newTurboButton(" &Help  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	d.refreshFiles()
	return d
}

// ── File listing ─────────────────────────────────────────────────────────────

func (d *saveFileDialog) refreshFiles() {
	d.files = nil
	d.dirs = nil
	d.selectedFile = -1
	d.selectedDir = -1
	d.fileOffset = 0

	entries, err := os.ReadDir(d.currentDir)
	if err != nil {
		return
	}
	d.dirs = append(d.dirs, "..")
	for _, e := range entries {
		if e.IsDir() {
			d.dirs = append(d.dirs, e.Name())
			continue
		}
		matched, _ := filepath.Match(strings.ToUpper(d.mask), strings.ToUpper(e.Name()))
		if !matched {
			matched, _ = filepath.Match(d.mask, e.Name())
		}
		if matched {
			info, err2 := e.Info()
			var sz int64
			var mt time.Time
			if err2 == nil {
				sz = info.Size()
				mt = info.ModTime()
			}
			d.files = append(d.files, openFileFEntry{name: e.Name(), size: sz, modTime: mt})
		}
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (d *saveFileDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < saveFileDlgW || height < saveFileDlgH {
		return
	}
	maxX, maxY := screen.Size()

	bgStyle     := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
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

	title := " Save a File As "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	d.drawLabelSection(screen, maxX, maxY, x, y)
	d.drawFilesLabel(screen, maxX, maxY, x, y)
	d.drawFileList(screen, maxX, maxY, x, y)
	d.drawScrollbar(screen, maxX, maxY, x, y)
	d.drawStatus(screen, maxX, maxY, x, y)
	d.drawButtons(screen, maxX, maxY, x, y)

	if d.showHistory {
		d.drawHistoryDropdown(screen, maxX, maxY, x, y)
	}
}

func (d *saveFileDialog) drawLabelSection(screen tcell.Screen, maxX, maxY, x, y int) {
	labelStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)

	// "&Save file as" — S em destaque
	putRune(screen, maxX, maxY, x+2, y+saveFileRowLabel, 'S', hotKeyStyle)
	putString(screen, maxX, maxY, x+3, y+saveFileRowLabel, "ave file as", labelStyle)

	inputBg := vgaBlue
	if d.focusField == saveFileFocusName {
		inputBg = vgaLightBlue
	}
	inputStyle := tcell.StyleDefault.Foreground(vgaLightCyan).Background(inputBg)

	for col := 0; col < saveFileInputW; col++ {
		putRune(screen, maxX, maxY, x+saveFileInputX+col, y+saveFileRowInput, ' ', inputStyle)
	}

	offset := 0
	if d.nameCursor >= saveFileInputW {
		offset = d.nameCursor - saveFileInputW + 1
	}
	for i, r := range d.nameText {
		col := i - offset
		if col < 0 || col >= saveFileInputW {
			continue
		}
		putRune(screen, maxX, maxY, x+saveFileInputX+col, y+saveFileRowInput, r, inputStyle)
	}

	if d.focusField == saveFileFocusName {
		curCol := d.nameCursor - offset
		if curCol >= 0 && curCol < saveFileInputW {
			ch := rune(' ')
			if d.nameCursor < len(d.nameText) {
				ch = d.nameText[d.nameCursor]
			}
			putRune(screen, maxX, maxY, x+saveFileInputX+curCol, y+saveFileRowInput, ch,
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}

	arrowStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaGreen)
	putRune(screen, maxX, maxY, x+saveFileArrowX, y+saveFileRowInput, '↓', arrowStyle)
}

func (d *saveFileDialog) drawFilesLabel(screen tcell.Screen, maxX, maxY, x, y int) {
	plainStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)
	putRune(screen, maxX, maxY, x+2, y+saveFileRowFiles, 'F', hotKeyStyle)
	putString(screen, maxX, maxY, x+3, y+saveFileRowFiles, "iles", plainStyle)
}

func (d *saveFileDialog) drawFileList(screen tcell.Screen, maxX, maxY, x, y int) {
	listStyle    := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)
	selFileStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	selDirStyle  := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	divStyle     := tcell.StyleDefault.Foreground(vgaDarkGray).Background(vgaCyan)

	listW := saveFileListXEnd - saveFileListX + 1
	divX  := x + saveFileListX + saveFileDivOff

	for row := 0; row < saveFileListRows; row++ {
		rowY := y + saveFileRowListTop + row

		for col := 0; col < listW; col++ {
			putRune(screen, maxX, maxY, x+saveFileListX+col, rowY, ' ', listStyle)
		}
		putRune(screen, maxX, maxY, divX, rowY, '│', divStyle)

		leftIdx := d.fileOffset + row
		if leftIdx < len(d.files) {
			style := listStyle
			if leftIdx == d.selectedFile && d.focusField == saveFileFocusFiles {
				style = selFileStyle
			}
			putString(screen, maxX, maxY, x+saveFileListX, rowY,
				ofdPadName(d.files[leftIdx].name, saveFileFileColW), style)
		}

		rightIdx := d.fileOffset + saveFileListRows + row
		if rightIdx < len(d.files) {
			style := listStyle
			if rightIdx == d.selectedFile && d.focusField == saveFileFocusFiles {
				style = selFileStyle
			}
			putString(screen, maxX, maxY, x+saveFileListX+saveFileFileColW, rowY,
				ofdPadName(d.files[rightIdx].name, saveFileFileColW), style)
		}

		dirColW := saveFileListXEnd - saveFileListX - saveFileDivOff - 1
		if row < len(d.dirs) {
			style := listStyle
			if row == d.selectedDir && d.focusField == saveFileFocusDirs {
				style = selDirStyle
			}
			putString(screen, maxX, maxY, divX+1, rowY,
				ofdFmtDir(d.dirs[row], dirColW), style)
		}
	}
}

func (d *saveFileDialog) drawScrollbar(screen tcell.Screen, maxX, maxY, x, y int) {
	style   := tcell.StyleDefault.Foreground(vgaLightCyan).Background(vgaBlue)
	arrowSt := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	scrollY := y + saveFileRowScroll
	scrollL := x + saveFileListX
	scrollR := x + saveFileListXEnd

	for col := scrollL; col <= scrollR; col++ {
		putRune(screen, maxX, maxY, col, scrollY, '▒', style)
	}
	putRune(screen, maxX, maxY, scrollL, scrollY, '◄', arrowSt)
	putRune(screen, maxX, maxY, scrollR, scrollY, '►', arrowSt)

	filesPerPage := saveFileListRows * 2
	totalFiles   := len(d.files)
	maxOffset    := totalFiles - filesPerPage
	if maxOffset < 0 {
		maxOffset = 0
	}
	trackLen := scrollR - scrollL - 2
	if trackLen > 0 && maxOffset > 0 {
		pos := scrollL + 1 + (d.fileOffset*trackLen)/maxOffset
		if pos > scrollR-1 {
			pos = scrollR - 1
		}
		putRune(screen, maxX, maxY, pos, scrollY, '■', arrowSt)
	}
}

func (d *saveFileDialog) drawStatus(screen tcell.Screen, maxX, maxY, x, y int) {
	statusStyle := tcell.StyleDefault.Foreground(vgaCyan).Background(vgaBlue)
	innerW      := saveFileDlgW - 2

	for col := 0; col < innerW; col++ {
		putRune(screen, maxX, maxY, x+1+col, y+saveFileRowStat1, ' ', statusStyle)
		putRune(screen, maxX, maxY, x+1+col, y+saveFileRowStat2, ' ', statusStyle)
	}

	line1 := filepath.Join(d.currentDir, d.mask)
	putString(screen, maxX, maxY, x+2, y+saveFileRowStat1, ofdTrunc(line1, innerW-2), statusStyle)
	putString(screen, maxX, maxY, x+2, y+saveFileRowStat2, ofdTrunc(d.selectedInfo(), innerW-2), statusStyle)
}

func (d *saveFileDialog) selectedInfo() string {
	if d.focusField == saveFileFocusFiles && d.selectedFile >= 0 && d.selectedFile < len(d.files) {
		f := d.files[d.selectedFile]
		return fmt.Sprintf("%-16s  file    %s  %s  %d",
			ofdTrunc(f.name, 16),
			f.modTime.Format("02-01-2006"),
			f.modTime.Format("15:04"),
			f.size)
	}
	if d.focusField == saveFileFocusDirs && d.selectedDir >= 0 && d.selectedDir < len(d.dirs) {
		return fmt.Sprintf("%-16s  directory", ofdTrunc(d.dirs[d.selectedDir], 16))
	}
	return ""
}

func (d *saveFileDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y int) {
	btnX := x + saveFileBtnX

	draw := func(btn *turboButton, focused bool, btnY int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
	}

	draw(d.btnOk, d.focusField == saveFileFocusOk, y+saveFileBtnRowOk)
	if d.msxMode {
		draw(d.btnToken, d.focusField == saveFileFocusToken, y+saveFileBtnRowToken)
	}
	draw(d.btnCancel, d.focusField == saveFileFocusCancel, y+saveFileBtnRowCancel)
	draw(d.btnHelp, d.focusField == saveFileFocusHelp, y+saveFileBtnRowHelp)
}

func (d *saveFileDialog) drawHistoryDropdown(screen tcell.Screen, maxX, maxY, x, y int) {
	if len(d.history) == 0 {
		return
	}
	bgStyle  := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)
	selStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	border   := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)

	dX, dY := x+saveFileInputX, y+saveFileRowInput+1
	dW := saveFileInputW + 2
	dH := len(d.history) + 2

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

	for i, entry := range d.history {
		style := bgStyle
		if i == d.historyIdx {
			style = selStyle
		}
		putString(screen, maxX, maxY, dX+1, dY+1+i, ofdPadName(entry, dW-2), style)
	}
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *saveFileDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event == nil {
			return
		}

		if d.showHistory {
			switch event.Key() {
			case tcell.KeyEscape:
				d.showHistory = false
			case tcell.KeyUp:
				if d.historyIdx > 0 {
					d.historyIdx--
				}
			case tcell.KeyDown:
				if d.historyIdx < len(d.history)-1 {
					d.historyIdx++
				}
			case tcell.KeyEnter:
				if d.historyIdx >= 0 && d.historyIdx < len(d.history) {
					d.nameText = []rune(d.history[d.historyIdx])
					d.nameCursor = len(d.nameText)
					d.mask = string(d.nameText)
					d.refreshFiles()
				}
				d.showHistory = false
			}
			return
		}

		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
			return
		case tcell.KeyTab:
			d.advanceFocus(1)
			return
		case tcell.KeyBacktab:
			d.advanceFocus(-1)
			return
		case tcell.KeyEnter:
			d.activateFocused()
			return
		}

		switch d.focusField {
		case saveFileFocusName:
			d.handleNameKey(event)
		case saveFileFocusFiles:
			d.handleFilesKey(event)
		case saveFileFocusDirs:
			d.handleDirsKey(event)
		default:
			if event.Key() == tcell.KeyRune {
				switch unicode.ToLower(event.Rune()) {
				case 'a':
					if d.msxMode {
						d.focusField = saveFileFocusOk
						d.activateFocused()
					}
				case 't':
					if d.msxMode {
						d.focusField = saveFileFocusToken
						d.activateFocused()
					}
				case 'o':
					if !d.msxMode {
						d.focusField = saveFileFocusOk
						d.activateFocused()
					}
				case 'h':
					d.focusField = saveFileFocusHelp
				}
			}
		}
	})
}

// advanceFocus move o foco, pulando Token se não for modo MSX-BASIC.
func (d *saveFileDialog) advanceFocus(delta int) {
	for i := 0; i <= saveFileFocusMax; i++ {
		d.focusField = (d.focusField + delta + saveFileFocusMax + 1) % (saveFileFocusMax + 1)
		if d.focusField == saveFileFocusToken && !d.msxMode {
			continue
		}
		break
	}
}

func (d *saveFileDialog) handleNameKey(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyLeft:
		if d.nameCursor > 0 {
			d.nameCursor--
		}
	case tcell.KeyRight:
		if d.nameCursor < len(d.nameText) {
			d.nameCursor++
		}
	case tcell.KeyHome:
		d.nameCursor = 0
	case tcell.KeyEnd:
		d.nameCursor = len(d.nameText)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.nameCursor > 0 {
			d.nameText = append(d.nameText[:d.nameCursor-1], d.nameText[d.nameCursor:]...)
			d.nameCursor--
		}
	case tcell.KeyDelete:
		if d.nameCursor < len(d.nameText) {
			d.nameText = append(d.nameText[:d.nameCursor], d.nameText[d.nameCursor+1:]...)
		}
	case tcell.KeyEnter:
		d.applyMask()
	case tcell.KeyRune:
		r := event.Rune()
		if unicode.IsPrint(r) {
			tail := append([]rune{r}, d.nameText[d.nameCursor:]...)
			d.nameText = append(d.nameText[:d.nameCursor], tail...)
			d.nameCursor++
		}
	}
}

func (d *saveFileDialog) handleFilesKey(event *tcell.EventKey) {
	filesPerPage := saveFileListRows * 2
	switch event.Key() {
	case tcell.KeyUp:
		if d.selectedFile > 0 {
			d.selectedFile--
			if d.selectedFile < d.fileOffset {
				d.fileOffset -= filesPerPage
				if d.fileOffset < 0 {
					d.fileOffset = 0
				}
			}
		}
	case tcell.KeyDown:
		if d.selectedFile == -1 && len(d.files) > 0 {
			d.selectedFile = 0
		} else if d.selectedFile < len(d.files)-1 {
			d.selectedFile++
			if d.selectedFile >= d.fileOffset+filesPerPage {
				d.fileOffset += filesPerPage
			}
		}
	case tcell.KeyLeft:
		if d.fileOffset >= filesPerPage {
			d.fileOffset -= filesPerPage
		}
	case tcell.KeyRight:
		maxOff := len(d.files) - filesPerPage
		if maxOff < 0 {
			maxOff = 0
		}
		if d.fileOffset < maxOff {
			d.fileOffset += filesPerPage
		}
	case tcell.KeyEnter:
		if d.selectedFile >= 0 && d.selectedFile < len(d.files) {
			d.nameText = []rune(d.files[d.selectedFile].name)
			d.nameCursor = len(d.nameText)
		}
	}
}

func (d *saveFileDialog) handleDirsKey(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyUp:
		if d.selectedDir > 0 {
			d.selectedDir--
		}
	case tcell.KeyDown:
		if d.selectedDir == -1 && len(d.dirs) > 0 {
			d.selectedDir = 0
		} else if d.selectedDir < len(d.dirs)-1 {
			d.selectedDir++
		}
	case tcell.KeyEnter:
		d.enterDir()
	}
}

func (d *saveFileDialog) enterDir() {
	if d.selectedDir < 0 || d.selectedDir >= len(d.dirs) {
		return
	}
	name := d.dirs[d.selectedDir]
	var newDir string
	if name == ".." {
		newDir = filepath.Dir(d.currentDir)
	} else {
		newDir = filepath.Join(d.currentDir, name)
	}
	d.currentDir = newDir
	d.refreshFiles()
}

func (d *saveFileDialog) applyMask() {
	mask := strings.TrimSpace(string(d.nameText))
	if mask == "" {
		mask = "*.BAS"
	}
	info, err := os.Stat(mask)
	if err == nil && info.IsDir() {
		d.currentDir = mask
		d.mask = "*.BAS"
		d.nameText = []rune(d.mask)
		d.nameCursor = len(d.nameText)
	} else {
		d.mask = mask
	}
	newHist := []string{mask}
	for _, h := range d.history {
		if h != mask {
			newHist = append(newHist, h)
		}
	}
	if len(newHist) > 10 {
		newHist = newHist[:10]
	}
	d.history = newHist
	d.refreshFiles()
	d.focusField = saveFileFocusFiles
}

func (d *saveFileDialog) activateFocused() {
	switch d.focusField {
	case saveFileFocusName:
		d.applyMask()
	case saveFileFocusFiles:
		if d.selectedFile >= 0 && d.selectedFile < len(d.files) {
			d.nameText = []rune(d.files[d.selectedFile].name)
			d.nameCursor = len(d.nameText)
		}
	case saveFileFocusDirs:
		d.enterDir()
	case saveFileFocusOk:
		d.doSave(false)
	case saveFileFocusToken:
		d.doSave(true)
	case saveFileFocusCancel:
		d.close()
	}
}

func (d *saveFileDialog) doSave(tokenized bool) {
	name := strings.TrimSpace(string(d.nameText))
	if name == "" || strings.ContainsAny(name, "*?") {
		return
	}
	path := name
	if !filepath.IsAbs(name) {
		path = filepath.Join(d.currentDir, name)
	}
	if d.onSave != nil {
		d.onSave(path, tokenized)
	}
	d.close()
}

func (d *saveFileDialog) close() {
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

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *saveFileDialog) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, w, h := d.GetRect()
		if mx < x || mx >= x+w || my < y || my >= y+h {
			return false, nil
		}
		if action != tview.MouseLeftClick {
			return true, nil
		}
		setFocus(d)

		// [■] fechar
		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		// ↓ histórico
		if my == y+saveFileRowInput && mx == x+saveFileArrowX {
			d.showHistory = !d.showHistory
			d.historyIdx = 0
			return true, nil
		}

		// Campo de entrada
		if my == y+saveFileRowInput && mx >= x+saveFileInputX && mx <= x+saveFileInputX+saveFileInputW-1 {
			d.focusField = saveFileFocusName
			offset := 0
			if d.nameCursor >= saveFileInputW {
				offset = d.nameCursor - saveFileInputW + 1
			}
			pos := (mx - (x + saveFileInputX)) + offset
			if pos > len(d.nameText) {
				pos = len(d.nameText)
			}
			d.nameCursor = pos
			return true, nil
		}

		// Lista de arquivos
		if my >= y+saveFileRowListTop && my < y+saveFileRowListTop+saveFileListRows {
			row := my - (y + saveFileRowListTop)
			divX := x + saveFileListX + saveFileDivOff
			if mx >= x+saveFileListX && mx < divX {
				sub := (mx - (x + saveFileListX)) / saveFileFileColW
				idx := d.fileOffset + sub*saveFileListRows + row
				if idx < len(d.files) {
					d.focusField = saveFileFocusFiles
					d.selectedFile = idx
				}
			} else if mx > divX && mx <= x+saveFileListXEnd {
				if row < len(d.dirs) {
					d.focusField = saveFileFocusDirs
					d.selectedDir = row
				}
			}
			return true, nil
		}

		// Botões
		btnX := x + saveFileBtnX
		type btnDef struct {
			row   int
			focus int
		}
		btns := []btnDef{
			{saveFileBtnRowOk, saveFileFocusOk},
			{saveFileBtnRowCancel, saveFileFocusCancel},
			{saveFileBtnRowHelp, saveFileFocusHelp},
		}
		if d.msxMode {
			btns = append(btns, btnDef{saveFileBtnRowToken, saveFileFocusToken})
		}
		for _, b := range btns {
			if my == y+b.row && mx >= btnX {
				d.focusField = b.focus
				d.activateFocused()
				return true, nil
			}
		}

		return true, nil
	}
}

func (d *saveFileDialog) PasteHandler() func(string, func(tview.Primitive)) {
	return func(_ string, _ func(tview.Primitive)) {}
}

// ── Show helper ───────────────────────────────────────────────────────────────

func showSaveFileDialog(app *App, initialName string) {
	dlg := newSaveFileDialog(app, initialName)

	dlg.onSave = func(path string, tokenized bool) {
		var text string
		if len(app.Editors) > 0 && app.ActiveEditor >= 0 && app.ActiveEditor < len(app.Editors) {
			text = app.Editors[app.ActiveEditor].editor.GetText()
		} else if app.Editor != nil {
			text = app.Editor.GetText()
		}

		var data []byte
		if tokenized {
			var err error
			data, err = basic.Tokenize(text)
			if err != nil {
				// Mostra erro simples na barra de status
				if app.StatusBar != nil {
					app.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao tokenizar: %v[-]", err))
				}
				return
			}
		} else {
			data = []byte(text)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			if app.StatusBar != nil {
				app.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao salvar: %v[-]", err))
			}
			return
		}

		// Atualiza o nome do arquivo no editor ativo
		if len(app.Editors) > 0 && app.ActiveEditor >= 0 && app.ActiveEditor < len(app.Editors) {
			app.Editors[app.ActiveEditor].fileName = filepath.Base(path)
		}
	}

	dlg.onClose = func() {
		if len(app.Editors) > 0 && app.ActiveEditor >= 0 {
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
			AddItem(dlg, saveFileDlgW, 0, true).
			AddItem(nil, 0, 1, false), saveFileDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	app.Pages.AddPage("save_file", container, true, true)
	app.Application.SetFocus(dlg)
}
