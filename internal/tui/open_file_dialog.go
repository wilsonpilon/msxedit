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
)

// Focus indices
const (
	openFileFocusName    = 0
	openFileFocusFiles   = 1
	openFileFocusDirs    = 2
	openFileFocusOpen    = 3
	openFileFocusReplace = 4
	openFileFocusCancel  = 5
	openFileFocusHelp    = 6
	openFileFocusMax     = 6
)

// Layout — all x/y offsets are relative to the dialog's own (x,y) origin.
//
// Row map (dialog height = 19, border at y+0 and y+18):
//   y+1   blank
//   y+2   &Name label              [Open   ] right col (aligned with input)
//   y+3   input field + ↓          [shadow ]
//   y+4   blank (gap Name→Files)
//   y+5   &Files label             [Replace] right col (aligned with &Files)
//   y+6..y+13  file list (8 rows)  [shadow ] at y+6
//   y+11                           [Cancel ] right col
//   y+12                           [shadow ]
//   y+13                           [Help   ] right col
//   y+14  scrollbar                [shadow ]
//   y+15  blank (gap Files→Status)
//   y+16  status line 1  (full-width, no border)
//   y+17  status line 2
//   y+18  bottom border
const (
	openFileDlgW = 68
	openFileDlgH = 19 // ends right after the status block

	// Name input field: x+2..x+45 (44 chars), ↓ at x+46
	openFileInputX = 2
	openFileInputW = 44
	openFileArrowX = 46

	// File list (no border): x+2..x+47, rows y+6..y+13 (8 visible rows)
	openFileListX    = 2
	openFileListXEnd = 47
	openFileListRows = 8
	openFileFileColW = 14 // each file sub-column (two = 28 chars)
	openFileDivOff   = 28 // divider offset: x+2+28 = x+30

	// Buttons: all padded so plain text = 7 chars → width = 11
	openFileBtnX          = 52
	openFileBtnRowOpen    = 3  // aligned with the Name input field
	openFileBtnRowReplace = 5  // aligned with the &Files label
	openFileBtnRowCancel  = 11 // a few rows above status
	openFileBtnRowHelp    = 13

	// Row positions
	openFileRowName    = 2
	openFileRowInput   = 3
	openFileRowFiles   = 5  // shifted down by 1 blank line
	openFileRowListTop = 6  // shifted down by 1 blank line
	openFileRowScroll  = 14 // openFileRowListTop + openFileListRows
	openFileRowStat1   = 16 // +1 blank line after scrollbar
	openFileRowStat2   = 17
)

type openFileFEntry struct {
	name    string
	size    int64
	modTime time.Time
}

type openFileDialog struct {
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

	btnOpen    *turboButton
	btnReplace *turboButton
	btnCancel  *turboButton
	btnHelp    *turboButton

	onOpen    func(path string)
	onReplace func(path string)
	onClose   func()
}

func newOpenFileDialog(app *App, initialMask string) *openFileDialog {
	if initialMask == "" {
		initialMask = "*.BAS"
	}
	cwd, _ := os.Getwd()
	d := &openFileDialog{
		Box:          tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:          app,
		pageName:     "open_file",
		nameText:     []rune(initialMask),
		nameCursor:   len([]rune(initialMask)),
		history:      []string{"*.BAS", "*.bas", "*.TXT", "*.MD"},
		currentDir:   cwd,
		mask:         initialMask,
		selectedFile: -1,
		selectedDir:  -1,
		focusField:   openFileFocusFiles,
		// All buttons padded so plain text = 7 chars → width = 11
		btnOpen:    newTurboButton(" &Open  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnReplace: newTurboButton("&Replace", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnCancel:  newTurboButton("Cancel ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnHelp:    newTurboButton(" &Help  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	d.refreshFiles()
	return d
}

// ── File listing ─────────────────────────────────────────────────────────────

func (d *openFileDialog) refreshFiles() {
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

func (d *openFileDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < openFileDlgW || height < openFileDlgH {
		return
	}
	maxX, maxY := screen.Size()

	bgStyle     := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)

	// Fill background
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			putRune(screen, maxX, maxY, x+col, y+row, ' ', bgStyle)
		}
	}

	// Outer border ╔═...═╗
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

	// Title centred
	title := " Open a File "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	// [■] close
	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	d.drawNameSection(screen, maxX, maxY, x, y)
	d.drawFilesLabel(screen, maxX, maxY, x, y)
	d.drawFileList(screen, maxX, maxY, x, y)
	d.drawScrollbar(screen, maxX, maxY, x, y)
	d.drawStatus(screen, maxX, maxY, x, y)
	d.drawButtons(screen, maxX, maxY, x, y)

	if d.showHistory {
		d.drawHistoryDropdown(screen, maxX, maxY, x, y)
	}
}

// Row y+2: &Name label  |  Row y+3: input (dark blue) + ↓ (green)
func (d *openFileDialog) drawNameSection(screen tcell.Screen, maxX, maxY, x, y int) {
	labelStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)

	// &Name label — N highlighted
	putRune(screen, maxX, maxY, x+2, y+openFileRowName, 'N', hotKeyStyle)
	putString(screen, maxX, maxY, x+3, y+openFileRowName, "ame", labelStyle)

	// Input field: dark blue background
	inputBg := vgaBlue
	if d.focusField == openFileFocusName {
		inputBg = vgaLightBlue
	}
	inputStyle := tcell.StyleDefault.Foreground(vgaLightCyan).Background(inputBg)

	for col := 0; col < openFileInputW; col++ {
		putRune(screen, maxX, maxY, x+openFileInputX+col, y+openFileRowInput, ' ', inputStyle)
	}

	// Draw text, scrolled so cursor stays visible
	offset := 0
	if d.nameCursor >= openFileInputW {
		offset = d.nameCursor - openFileInputW + 1
	}
	for i, r := range d.nameText {
		col := i - offset
		if col < 0 || col >= openFileInputW {
			continue
		}
		putRune(screen, maxX, maxY, x+openFileInputX+col, y+openFileRowInput, r, inputStyle)
	}

	// Cursor block
	if d.focusField == openFileFocusName {
		curCol := d.nameCursor - offset
		if curCol >= 0 && curCol < openFileInputW {
			ch := rune(' ')
			if d.nameCursor < len(d.nameText) {
				ch = d.nameText[d.nameCursor]
			}
			putRune(screen, maxX, maxY, x+openFileInputX+curCol, y+openFileRowInput, ch,
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}

	// ↓ arrow — green background, no brackets, immediately right of field
	arrowStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaGreen)
	putRune(screen, maxX, maxY, x+openFileArrowX, y+openFileRowInput, '↓', arrowStyle)
}

// Row y+4: &Files label
func (d *openFileDialog) drawFilesLabel(screen tcell.Screen, maxX, maxY, x, y int) {
	plainStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)
	putRune(screen, maxX, maxY, x+2, y+openFileRowFiles, 'F', hotKeyStyle)
	putString(screen, maxX, maxY, x+3, y+openFileRowFiles, "iles", plainStyle)
}

// Rows y+5..y+12: file list (no border), cyan background
func (d *openFileDialog) drawFileList(screen tcell.Screen, maxX, maxY, x, y int) {
	listStyle    := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)
	selFileStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	selDirStyle  := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	divStyle     := tcell.StyleDefault.Foreground(vgaDarkGray).Background(vgaCyan)

	listW := openFileListXEnd - openFileListX + 1
	divX  := x + openFileListX + openFileDivOff // = x+30

	for row := 0; row < openFileListRows; row++ {
		rowY := y + openFileRowListTop + row

		// Fill entire row with cyan
		for col := 0; col < listW; col++ {
			putRune(screen, maxX, maxY, x+openFileListX+col, rowY, ' ', listStyle)
		}

		// Vertical divider
		putRune(screen, maxX, maxY, divX, rowY, '│', divStyle)

		// Left file sub-column
		leftIdx := d.fileOffset + row
		if leftIdx < len(d.files) {
			style := listStyle
			if leftIdx == d.selectedFile && d.focusField == openFileFocusFiles {
				style = selFileStyle
			}
			putString(screen, maxX, maxY, x+openFileListX, rowY,
				ofdPadName(d.files[leftIdx].name, openFileFileColW), style)
		}

		// Right file sub-column
		rightIdx := d.fileOffset + openFileListRows + row
		if rightIdx < len(d.files) {
			style := listStyle
			if rightIdx == d.selectedFile && d.focusField == openFileFocusFiles {
				style = selFileStyle
			}
			putString(screen, maxX, maxY, x+openFileListX+openFileFileColW, rowY,
				ofdPadName(d.files[rightIdx].name, openFileFileColW), style)
		}

		// Dirs column (right of divider)
		dirColW := openFileListXEnd - openFileListX - openFileDivOff - 1 // 47-2-28-1 = 16
		if row < len(d.dirs) {
			style := listStyle
			if row == d.selectedDir && d.focusField == openFileFocusDirs {
				style = selDirStyle
			}
			putString(screen, maxX, maxY, divX+1, rowY,
				ofdFmtDir(d.dirs[row], dirColW), style)
		}
	}
}

// Row y+13: horizontal scrollbar with blue background
func (d *openFileDialog) drawScrollbar(screen tcell.Screen, maxX, maxY, x, y int) {
	style    := tcell.StyleDefault.Foreground(vgaLightCyan).Background(vgaBlue)
	arrowSt  := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	scrollY  := y + openFileRowScroll
	scrollL  := x + openFileListX
	scrollR  := x + openFileListXEnd

	for col := scrollL; col <= scrollR; col++ {
		putRune(screen, maxX, maxY, col, scrollY, '▒', style)
	}
	putRune(screen, maxX, maxY, scrollL, scrollY, '◄', arrowSt)
	putRune(screen, maxX, maxY, scrollR, scrollY, '►', arrowSt)

	filesPerPage := openFileListRows * 2
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

// Rows y+14..y+15: status — no border, full inner width (x+1..x+66), dark blue/cyan
func (d *openFileDialog) drawStatus(screen tcell.Screen, maxX, maxY, x, y int) {
	statusStyle := tcell.StyleDefault.Foreground(vgaCyan).Background(vgaBlue)
	innerW      := openFileDlgW - 2 // 66 chars from x+1 to x+66

	// Fill both rows
	for col := 0; col < innerW; col++ {
		putRune(screen, maxX, maxY, x+1+col, y+openFileRowStat1, ' ', statusStyle)
		putRune(screen, maxX, maxY, x+1+col, y+openFileRowStat2, ' ', statusStyle)
	}

	line1 := filepath.Join(d.currentDir, d.mask)
	putString(screen, maxX, maxY, x+2, y+openFileRowStat1, ofdTrunc(line1, innerW-2), statusStyle)
	putString(screen, maxX, maxY, x+2, y+openFileRowStat2, ofdTrunc(d.selectedInfo(), innerW-2), statusStyle)
}

func (d *openFileDialog) selectedInfo() string {
	if d.focusField == openFileFocusFiles && d.selectedFile >= 0 && d.selectedFile < len(d.files) {
		f := d.files[d.selectedFile]
		return fmt.Sprintf("%-16s  file    %s  %s  %d",
			ofdTrunc(f.name, 16),
			f.modTime.Format("02-01-2006"),
			f.modTime.Format("15:04"),
			f.size)
	}
	if d.focusField == openFileFocusDirs && d.selectedDir >= 0 && d.selectedDir < len(d.dirs) {
		return fmt.Sprintf("%-16s  directory", ofdTrunc(d.dirs[d.selectedDir], 16))
	}
	return ""
}

// Buttons: Open/Replace at top, Cancel/Help at bottom, all width 11
func (d *openFileDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y int) {
	btnX := x + openFileBtnX

	draw := func(btn *turboButton, focused bool, btnY int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
	}

	draw(d.btnOpen, d.focusField == openFileFocusOpen, y+openFileBtnRowOpen)
	draw(d.btnReplace, d.focusField == openFileFocusReplace, y+openFileBtnRowReplace)
	draw(d.btnCancel, d.focusField == openFileFocusCancel, y+openFileBtnRowCancel)
	draw(d.btnHelp, d.focusField == openFileFocusHelp, y+openFileBtnRowHelp)
}

// History dropdown overlaid below input row
func (d *openFileDialog) drawHistoryDropdown(screen tcell.Screen, maxX, maxY, x, y int) {
	if len(d.history) == 0 {
		return
	}
	bgStyle  := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)
	selStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	border   := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)

	dX, dY := x+openFileInputX, y+openFileRowInput+1
	dW := openFileInputW + 2
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

func (d *openFileDialog) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
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
			d.focusField = (d.focusField + 1) % (openFileFocusMax + 1)
			return
		case tcell.KeyBacktab:
			d.focusField = (d.focusField - 1 + openFileFocusMax + 1) % (openFileFocusMax + 1)
			return
		case tcell.KeyEnter:
			d.activateFocused()
			return
		}

		switch d.focusField {
		case openFileFocusName:
			d.handleNameKey(event)
		case openFileFocusFiles:
			d.handleFilesKey(event)
		case openFileFocusDirs:
			d.handleDirsKey(event)
		default:
			if event.Key() == tcell.KeyRune {
				switch unicode.ToLower(event.Rune()) {
				case 'o':
					d.focusField = openFileFocusOpen
					d.activateFocused()
				case 'r':
					d.focusField = openFileFocusReplace
					d.activateFocused()
				case 'h':
					d.focusField = openFileFocusHelp
				}
			}
		}
	})
}

func (d *openFileDialog) handleNameKey(event *tcell.EventKey) {
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

func (d *openFileDialog) handleFilesKey(event *tcell.EventKey) {
	filesPerPage := openFileListRows * 2
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
			if d.selectedFile >= d.fileOffset+filesPerPage {
				d.selectedFile = d.fileOffset + filesPerPage - 1
			}
		}
	case tcell.KeyRight:
		maxOff := len(d.files) - filesPerPage
		if maxOff < 0 {
			maxOff = 0
		}
		if d.fileOffset < maxOff {
			d.fileOffset += filesPerPage
			if d.fileOffset > maxOff {
				d.fileOffset = maxOff
			}
		}
	case tcell.KeyEnter:
		if d.selectedFile >= 0 && d.selectedFile < len(d.files) {
			d.nameText = []rune(d.files[d.selectedFile].name)
			d.nameCursor = len(d.nameText)
		}
	}
}

func (d *openFileDialog) handleDirsKey(event *tcell.EventKey) {
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

func (d *openFileDialog) enterDir() {
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

func (d *openFileDialog) applyMask() {
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
	d.focusField = openFileFocusFiles
}

func (d *openFileDialog) activateFocused() {
	switch d.focusField {
	case openFileFocusName:
		d.applyMask()
	case openFileFocusFiles:
		if d.selectedFile >= 0 && d.selectedFile < len(d.files) {
			d.nameText = []rune(d.files[d.selectedFile].name)
			d.nameCursor = len(d.nameText)
		}
	case openFileFocusDirs:
		d.enterDir()
	case openFileFocusOpen:
		if d.onOpen != nil {
			d.onOpen(d.resolvedPath())
		}
		d.close()
	case openFileFocusReplace:
		if d.onReplace != nil {
			d.onReplace(d.resolvedPath())
		}
		d.close()
	case openFileFocusCancel:
		d.close()
	}
}

func (d *openFileDialog) resolvedPath() string {
	name := strings.TrimSpace(string(d.nameText))
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(d.currentDir, name)
}

func (d *openFileDialog) close() {
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

func (d *openFileDialog) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
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

		// [■] close
		if my == y && mx >= x+2 && mx <= x+4 {
			d.close()
			return true, nil
		}

		// ↓ arrow
		if my == y+openFileRowInput && mx == x+openFileArrowX {
			d.showHistory = !d.showHistory
			d.historyIdx = 0
			return true, nil
		}

		// Name input
		if my == y+openFileRowInput && mx >= x+openFileInputX && mx <= x+openFileInputX+openFileInputW-1 {
			d.focusField = openFileFocusName
			offset := 0
			if d.nameCursor >= openFileInputW {
				offset = d.nameCursor - openFileInputW + 1
			}
			pos := (mx - (x + openFileInputX)) + offset
			if pos > len(d.nameText) {
				pos = len(d.nameText)
			}
			d.nameCursor = pos
			return true, nil
		}

		// File list rows
		if my >= y+openFileRowListTop && my < y+openFileRowListTop+openFileListRows {
			row := my - (y + openFileRowListTop)
			divX := x + openFileListX + openFileDivOff
			if mx >= x+openFileListX && mx < divX {
				sub := (mx - (x + openFileListX)) / openFileFileColW
				idx := d.fileOffset + sub*openFileListRows + row
				if idx < len(d.files) {
					d.focusField = openFileFocusFiles
					d.selectedFile = idx
				}
			} else if mx > divX && mx <= x+openFileListXEnd {
				if row < len(d.dirs) {
					d.focusField = openFileFocusDirs
					d.selectedDir = row
				}
			}
			return true, nil
		}

		// Scrollbar
		if my == y+openFileRowScroll {
			filesPerPage := openFileListRows * 2
			maxOff := len(d.files) - filesPerPage
			if maxOff < 0 {
				maxOff = 0
			}
			scrollL := x + openFileListX
			scrollR := x + openFileListXEnd
			if mx == scrollL {
				if d.fileOffset >= filesPerPage {
					d.fileOffset -= filesPerPage
				}
			} else if mx == scrollR {
				if d.fileOffset < maxOff {
					d.fileOffset += filesPerPage
					if d.fileOffset > maxOff {
						d.fileOffset = maxOff
					}
				}
			} else {
				trackLen := scrollR - scrollL - 2
				if trackLen > 0 && maxOff > 0 {
					ratio := float64(mx-(scrollL+1)) / float64(trackLen)
					d.fileOffset = int(ratio*float64(maxOff+1)) / filesPerPage * filesPerPage
					if d.fileOffset > maxOff {
						d.fileOffset = maxOff
					}
				}
			}
			return true, nil
		}

		// Buttons
		btnX := x + openFileBtnX
		for i, btnRow := range []int{openFileBtnRowOpen, openFileBtnRowReplace, openFileBtnRowCancel, openFileBtnRowHelp} {
			if my == y+btnRow && mx >= btnX {
				d.focusField = openFileFocusOpen + i
				d.activateFocused()
				return true, nil
			}
		}

		return true, nil
	}
}

func (d *openFileDialog) PasteHandler() func(string, func(tview.Primitive)) {
	return func(_ string, _ func(tview.Primitive)) {}
}

// ── Show helper ───────────────────────────────────────────────────────────────

func showOpenFileDialog(app *App, mask string, onOpen func(string), onReplace func(string)) {
	dlg := newOpenFileDialog(app, mask)
	dlg.onOpen = onOpen
	dlg.onReplace = onReplace
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
			AddItem(dlg, openFileDlgW, 0, true).
			AddItem(nil, 0, 1, false), openFileDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	app.Pages.AddPage("open_file", container, true, true)
	app.Application.SetFocus(dlg)
}

// ── String helpers ────────────────────────────────────────────────────────────

func ofdPadName(name string, width int) string {
	runes := []rune(name)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return name + strings.Repeat(" ", width-len(runes))
}

func ofdFmtDir(name string, width int) string {
	if name == ".." {
		return ofdPadName("[..]", width)
	}
	return ofdPadName("["+strings.ToUpper(name)+"]", width)
}

func ofdTrunc(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
