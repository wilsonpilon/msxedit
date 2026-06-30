package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ── Layout ───────────────────────────────────────────────────────────────────
//
//  y+0   ╔══════════ Change Directory ═══════════╗
//  y+1   ║                                       ║
//  y+2   ║  Directory name                       ║
//  y+3   ║  [input..............................↓]║  [ &Ok   ]
//  y+4   ║                                       ║  (shadow)
//  y+5   ║  Directory tree                       ║  [&Chdir ]
//  y+6   ║   C:\                             ▲  ║  (shadow)
//  y+7   ║    ├─ dos                         ▒  ║  [&Revert]
//  y+8   ║    │  └─ msxedit                  ▒  ║  (shadow)
//  y+9   ║    └─ Windows                     ▒  ║  [ &Help ]
//  y+10  ║   D:\                             ▒  ║  (shadow)
//  ...   ║   ...                             ▒  ║
//  y+17  ║                                   ▼  ║
//  y+18  ║                                       ║
//  y+19  ╚═══════════════════════════════════════╝

const (
	chdirDlgW = 68
	chdirDlgH = 20

	chdirInputX = 2
	chdirInputW = 44
	chdirArrowX = 46

	chdirListX   = 2
	chdirListW   = 46 // x+2..x+47 inclusive
	chdirScrollX = 48 // vertical scrollbar column

	chdirListRows = 12 // y+6..y+17

	chdirBtnX          = 52
	chdirBtnRowOk      = 3
	chdirBtnRowChdir   = 5
	chdirBtnRowRevert  = 7
	chdirBtnRowHelp    = 9

	chdirRowNameLabel = 2
	chdirRowInput     = 3
	chdirRowTreeLabel = 5
	chdirRowListTop   = 6
)

// ── Focus ─────────────────────────────────────────────────────────────────────

const (
	chdirFocusName   = 0
	chdirFocusTree   = 1
	chdirFocusOk     = 2
	chdirFocusChdir  = 3
	chdirFocusRevert = 4
	chdirFocusHelp   = 5
	chdirFocusMax    = 5
)

// ── Tree structures ───────────────────────────────────────────────────────────

type chdirNode struct {
	name     string
	fullPath string
	expanded bool
	children []*chdirNode
	parent   *chdirNode
	loaded   bool
	isRoot   bool
}

type chdirFlatEntry struct {
	node   *chdirNode
	prefix string // pre-computed connection-line prefix
}

// ── Dialog struct ─────────────────────────────────────────────────────────────

type changeDirDialog struct {
	*tview.Box
	app      *App
	pageName string

	nameText    []rune
	nameCursor  int
	history     []string
	showHistory bool

	roots      []*chdirNode
	flat       []chdirFlatEntry
	selected   int
	treeOffset int

	origDir string

	focusField int

	btnOk     *turboButton
	btnChdir  *turboButton
	btnRevert *turboButton
	btnHelp   *turboButton

	onClose func()
}

func newChangeDirDialog(app *App) *changeDirDialog {
	origDir, _ := os.Getwd()
	d := &changeDirDialog{
		Box:        tview.NewBox().SetBackgroundColor(app.Theme.PopupBg),
		app:        app,
		pageName:   "change_dir",
		nameText:   []rune(origDir),
		nameCursor: len([]rune(origDir)),
		history:    []string{origDir},
		origDir:    origDir,
		focusField: chdirFocusTree,
		btnOk:      newTurboButton("  &Ok  ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnChdir:   newTurboButton("&Chdir ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnRevert:  newTurboButton("&Revert", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
		btnHelp:    newTurboButton(" &Help ", vgaWhite, vgaGreen, vgaYellow, vgaBlack),
	}
	d.buildRoots()
	d.expandToPath(origDir)
	return d
}

// ── Tree building ─────────────────────────────────────────────────────────────

func (d *changeDirDialog) buildRoots() {
	d.roots = nil
	if runtime.GOOS == "windows" {
		for c := 'A'; c <= 'Z'; c++ {
			drive := string(c) + ":\\"
			if _, err := os.Stat(drive); err == nil {
				d.roots = append(d.roots, &chdirNode{
					name:     string(c) + ":",
					fullPath: drive,
					isRoot:   true,
				})
			}
		}
	} else {
		d.roots = []*chdirNode{{
			name:     "/",
			fullPath: "/",
			isRoot:   true,
		}}
	}
}

func (d *changeDirDialog) loadChildren(node *chdirNode) {
	if node.loaded {
		return
	}
	node.loaded = true
	entries, err := os.ReadDir(node.fullPath)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "$") {
			continue
		}
		node.children = append(node.children, &chdirNode{
			name:     name,
			fullPath: filepath.Join(node.fullPath, name),
			parent:   node,
		})
	}
	sort.Slice(node.children, func(i, j int) bool {
		return strings.ToLower(node.children[i].name) < strings.ToLower(node.children[j].name)
	})
}

// expandToPath expands the tree to reveal targetPath and selects it.
func (d *changeDirDialog) expandToPath(targetPath string) {
	targetPath = filepath.Clean(targetPath)
	vol := filepath.VolumeName(targetPath)

	for _, root := range d.roots {
		rootVol := filepath.VolumeName(root.fullPath)
		if !strings.EqualFold(rootVol, vol) {
			continue
		}
		d.loadChildren(root)
		root.expanded = true

		// Walk each path segment
		rest := targetPath[len(vol):]
		parts := strings.FieldsFunc(rest, func(r rune) bool { return r == '/' || r == '\\' })
		cur := root
		for _, part := range parts {
			for _, child := range cur.children {
				if strings.EqualFold(child.name, part) {
					d.loadChildren(child)
					child.expanded = true
					cur = child
					break
				}
			}
		}

		d.rebuildFlat()
		for i, e := range d.flat {
			if strings.EqualFold(filepath.Clean(e.node.fullPath), targetPath) {
				d.selected = i
				break
			}
		}
		d.ensureVisible()
		return
	}
	d.rebuildFlat()
}

// addToFlat recursively fills d.flat with depth-first entries.
// connector is the inherited prefix string from parent.
func (d *changeDirDialog) addToFlat(nodes []*chdirNode, connector string, depth int) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		var prefix string
		if depth == 0 {
			prefix = " " // drive roots: simple leading space
		} else if isLast {
			prefix = connector + "└─"
		} else {
			prefix = connector + "├─"
		}
		d.flat = append(d.flat, chdirFlatEntry{node: node, prefix: prefix})

		if node.expanded && len(node.children) > 0 {
			var childConn string
			if depth == 0 {
				childConn = "  "
			} else if isLast {
				childConn = connector + "  "
			} else {
				childConn = connector + "│ "
			}
			d.addToFlat(node.children, childConn, depth+1)
		}
	}
}

func (d *changeDirDialog) rebuildFlat() {
	d.flat = nil
	d.addToFlat(d.roots, "", 0)
	if d.selected >= len(d.flat) {
		d.selected = len(d.flat) - 1
	}
	if d.selected < 0 {
		d.selected = 0
	}
	d.ensureVisible()
}

func (d *changeDirDialog) ensureVisible() {
	if d.selected < d.treeOffset {
		d.treeOffset = d.selected
	}
	if d.selected >= d.treeOffset+chdirListRows {
		d.treeOffset = d.selected - chdirListRows + 1
	}
	maxOff := len(d.flat) - chdirListRows
	if maxOff < 0 {
		maxOff = 0
	}
	if d.treeOffset > maxOff {
		d.treeOffset = maxOff
	}
	if d.treeOffset < 0 {
		d.treeOffset = 0
	}
}

// syncNameToSelection copies the selected node's fullPath into the name field.
func (d *changeDirDialog) syncNameToSelection() {
	if d.selected >= 0 && d.selected < len(d.flat) {
		path := d.flat[d.selected].node.fullPath
		d.nameText = []rune(path)
		d.nameCursor = len(d.nameText)
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (d *changeDirDialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	if width < chdirDlgW || height < chdirDlgH {
		return
	}
	maxX, maxY := screen.Size()

	bgStyle     := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	borderStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(d.app.Theme.PopupBg)

	// Background fill
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
	title := " Change Directory "
	titleX := x + (width-len([]rune(title)))/2
	putString(screen, maxX, maxY, titleX, y, title, borderStyle)

	// [■] close button
	putRune(screen, maxX, maxY, x+2, y, '[', borderStyle)
	putRune(screen, maxX, maxY, x+3, y, '■', tcell.StyleDefault.Foreground(vgaGreen).Background(d.app.Theme.PopupBg))
	putRune(screen, maxX, maxY, x+4, y, ']', borderStyle)

	d.drawNameSection(screen, maxX, maxY, x, y)
	d.drawTreeSection(screen, maxX, maxY, x, y)
	d.drawScrollbar(screen, maxX, maxY, x, y)
	d.drawButtons(screen, maxX, maxY, x, y)

	if d.showHistory {
		d.drawHistoryDropdown(screen, maxX, maxY, x, y)
	}
}

func (d *changeDirDialog) drawNameSection(screen tcell.Screen, maxX, maxY, x, y int) {
	labelStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)

	// "Directory &name" — 'n' highlighted
	putString(screen, maxX, maxY, x+2, y+chdirRowNameLabel, "Directory ", labelStyle)
	putRune(screen, maxX, maxY, x+12, y+chdirRowNameLabel, 'n', hotKeyStyle)
	putString(screen, maxX, maxY, x+13, y+chdirRowNameLabel, "ame", labelStyle)

	// Input field
	inputBg := vgaBlue
	if d.focusField == chdirFocusName {
		inputBg = vgaLightBlue
	}
	inputStyle := tcell.StyleDefault.Foreground(vgaLightCyan).Background(inputBg)
	for col := 0; col < chdirInputW; col++ {
		putRune(screen, maxX, maxY, x+chdirInputX+col, y+chdirRowInput, ' ', inputStyle)
	}

	offset := 0
	if d.nameCursor >= chdirInputW {
		offset = d.nameCursor - chdirInputW + 1
	}
	for i, r := range d.nameText {
		col := i - offset
		if col < 0 || col >= chdirInputW {
			continue
		}
		putRune(screen, maxX, maxY, x+chdirInputX+col, y+chdirRowInput, r, inputStyle)
	}

	if d.focusField == chdirFocusName {
		curCol := d.nameCursor - offset
		if curCol >= 0 && curCol < chdirInputW {
			ch := rune(' ')
			if d.nameCursor < len(d.nameText) {
				ch = d.nameText[d.nameCursor]
			}
			putRune(screen, maxX, maxY, x+chdirInputX+curCol, y+chdirRowInput, ch,
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaYellow))
		}
	}

	// ↓ arrow
	arrowStyle := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaGreen)
	if d.showHistory {
		arrowStyle = tcell.StyleDefault.Foreground(vgaBlack).Background(vgaYellow)
	}
	putRune(screen, maxX, maxY, x+chdirArrowX, y+chdirRowInput, '↓', arrowStyle)
}

func (d *changeDirDialog) drawTreeSection(screen tcell.Screen, maxX, maxY, x, y int) {
	labelStyle  := tcell.StyleDefault.Foreground(d.app.Theme.PopupFg).Background(d.app.Theme.PopupBg)
	hotKeyStyle := tcell.StyleDefault.Foreground(vgaYellow).Background(d.app.Theme.PopupBg)
	listStyle   := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaCyan)
	selStyle    := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)

	// "Directory &tree" — 't' highlighted
	putString(screen, maxX, maxY, x+2, y+chdirRowTreeLabel, "Directory ", labelStyle)
	putRune(screen, maxX, maxY, x+12, y+chdirRowTreeLabel, 't', hotKeyStyle)
	putString(screen, maxX, maxY, x+13, y+chdirRowTreeLabel, "ree", labelStyle)

	// Tree rows
	for row := 0; row < chdirListRows; row++ {
		rowY := y + chdirRowListTop + row
		entryIdx := d.treeOffset + row

		style := listStyle
		if entryIdx == d.selected && d.focusField == chdirFocusTree {
			style = selStyle
		}

		// Fill row background
		for col := 0; col < chdirListW; col++ {
			putRune(screen, maxX, maxY, x+chdirListX+col, rowY, ' ', style)
		}

		if entryIdx >= len(d.flat) {
			continue
		}
		e := d.flat[entryIdx]
		line := e.prefix + " " + e.node.name

		// Truncate to fit (leave room for a trailing ellipsis)
		runes := []rune(line)
		if len(runes) > chdirListW {
			runes = append(runes[:chdirListW-1], '…')
		}
		putString(screen, maxX, maxY, x+chdirListX, rowY, string(runes), style)

		// Selected item without tree focus — show a faint marker
		if entryIdx == d.selected && d.focusField != chdirFocusTree {
			putString(screen, maxX, maxY, x+chdirListX, rowY, string(runes),
				tcell.StyleDefault.Foreground(vgaBlue).Background(vgaCyan))
		}
	}
}

func (d *changeDirDialog) drawScrollbar(screen tcell.Screen, maxX, maxY, x, y int) {
	total   := len(d.flat)
	sx      := x + chdirScrollX
	topY    := y + chdirRowListTop
	bottomY := topY + chdirListRows - 1

	trackStyle := tcell.StyleDefault.Foreground(vgaLightCyan).Background(vgaBlue)
	arrowStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)

	putRune(screen, maxX, maxY, sx, topY, '▲', arrowStyle)
	putRune(screen, maxX, maxY, sx, bottomY, '▼', arrowStyle)

	trackH := chdirListRows - 2
	if trackH < 1 {
		trackH = 1
	}
	for i := 0; i < trackH; i++ {
		putRune(screen, maxX, maxY, sx, topY+1+i, '▒', trackStyle)
	}

	if total > chdirListRows && trackH > 0 {
		thumbPos := d.treeOffset * trackH / (total - chdirListRows)
		if thumbPos >= trackH {
			thumbPos = trackH - 1
		}
		putRune(screen, maxX, maxY, sx, topY+1+thumbPos, '■', arrowStyle)
	}
}

func (d *changeDirDialog) drawButtons(screen tcell.Screen, maxX, maxY, x, y int) {
	btnX := x + chdirBtnX
	draw := func(btn *turboButton, focused bool, btnY int) {
		if focused {
			f := newTurboButton(btn.label, vgaBlack, vgaYellow, vgaRed, vgaBlack)
			f.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
			return
		}
		btn.draw(screen, maxX, maxY, btnX, btnY, d.app.Theme.PopupBg)
	}
	draw(d.btnOk, d.focusField == chdirFocusOk, y+chdirBtnRowOk)
	draw(d.btnChdir, d.focusField == chdirFocusChdir, y+chdirBtnRowChdir)
	draw(d.btnRevert, d.focusField == chdirFocusRevert, y+chdirBtnRowRevert)
	draw(d.btnHelp, d.focusField == chdirFocusHelp, y+chdirBtnRowHelp)
}

func (d *changeDirDialog) drawHistoryDropdown(screen tcell.Screen, maxX, maxY, x, y int) {
	if len(d.history) == 0 {
		return
	}
	bgStyle  := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)
	selStyle := tcell.StyleDefault.Foreground(vgaWhite).Background(vgaBlue)
	border   := tcell.StyleDefault.Foreground(vgaBlack).Background(vgaLightGray)

	dX, dY := x+chdirInputX, y+chdirRowInput+1
	dW := chdirInputW + 2
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
		if i == 0 { // highlight first (most recent)
			style = selStyle
		}
		runes := []rune(entry)
		if len(runes) > dW-2 {
			runes = runes[:dW-2]
		}
		putString(screen, maxX, maxY, dX+1, dY+1+i, string(runes), style)
	}
}

// ── Input handler ─────────────────────────────────────────────────────────────

func (d *changeDirDialog) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if event == nil {
			return
		}
		if d.showHistory {
			d.handleHistoryKey(event)
			return
		}
		switch d.focusField {
		case chdirFocusName:
			if d.handleNameKey(event) {
				return
			}
		case chdirFocusTree:
			if d.handleTreeKey(event) {
				return
			}
		}
		// Global keys
		switch event.Key() {
		case tcell.KeyEscape:
			d.close()
		case tcell.KeyTab:
			d.focusField = (d.focusField + 1) % (chdirFocusMax + 1)
		case tcell.KeyBacktab:
			d.focusField = (d.focusField - 1 + chdirFocusMax + 1) % (chdirFocusMax + 1)
		case tcell.KeyEnter:
			d.activateFocused()
		case tcell.KeyRune:
			switch unicode.ToLower(event.Rune()) {
			case 'o':
				d.doChdir()
			case 'c':
				if event.Modifiers() == 0 {
					d.doChdir()
				}
			case 'r':
				d.doRevert()
			case 'h':
				d.focusField = chdirFocusHelp
			}
		}
	})
}

func (d *changeDirDialog) handleNameKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyLeft:
		if d.nameCursor > 0 {
			d.nameCursor--
		}
		return true
	case tcell.KeyRight:
		if d.nameCursor < len(d.nameText) {
			d.nameCursor++
		}
		return true
	case tcell.KeyHome:
		d.nameCursor = 0
		return true
	case tcell.KeyEnd:
		d.nameCursor = len(d.nameText)
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.nameCursor > 0 {
			d.nameText = append(d.nameText[:d.nameCursor-1], d.nameText[d.nameCursor:]...)
			d.nameCursor--
		}
		return true
	case tcell.KeyDelete:
		if d.nameCursor < len(d.nameText) {
			d.nameText = append(d.nameText[:d.nameCursor], d.nameText[d.nameCursor+1:]...)
		}
		return true
	case tcell.KeyEnter:
		target := strings.TrimSpace(string(d.nameText))
		if target != "" {
			d.expandToPath(target)
			d.addToHistory(target)
		}
		return true
	case tcell.KeyDown:
		d.showHistory = true
		return true
	case tcell.KeyRune:
		r := event.Rune()
		if unicode.IsPrint(r) {
			d.nameText = append(d.nameText[:d.nameCursor], append([]rune{r}, d.nameText[d.nameCursor:]...)...)
			d.nameCursor++
			return true
		}
	}
	return false
}

func (d *changeDirDialog) handleTreeKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyUp:
		if d.selected > 0 {
			d.selected--
			d.ensureVisible()
			d.syncNameToSelection()
		}
		return true
	case tcell.KeyDown:
		if d.selected < len(d.flat)-1 {
			d.selected++
			d.ensureVisible()
			d.syncNameToSelection()
		}
		return true
	case tcell.KeyPgUp:
		d.selected -= chdirListRows
		if d.selected < 0 {
			d.selected = 0
		}
		d.ensureVisible()
		d.syncNameToSelection()
		return true
	case tcell.KeyPgDn:
		d.selected += chdirListRows
		if d.selected >= len(d.flat) {
			d.selected = len(d.flat) - 1
		}
		d.ensureVisible()
		d.syncNameToSelection()
		return true
	case tcell.KeyHome:
		d.selected = 0
		d.treeOffset = 0
		d.syncNameToSelection()
		return true
	case tcell.KeyEnd:
		d.selected = len(d.flat) - 1
		d.ensureVisible()
		d.syncNameToSelection()
		return true
	case tcell.KeyRight, tcell.KeyEnter:
		d.toggleExpand()
		return true
	case tcell.KeyLeft:
		d.collapseOrGoParent()
		return true
	}
	return false
}

func (d *changeDirDialog) handleHistoryKey(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyEscape:
		d.showHistory = false
	case tcell.KeyEnter:
		if len(d.history) > 0 {
			path := d.history[0]
			d.nameText = []rune(path)
			d.nameCursor = len(d.nameText)
			d.expandToPath(path)
		}
		d.showHistory = false
	case tcell.KeyDown:
		d.showHistory = false
	}
}

func (d *changeDirDialog) toggleExpand() {
	if d.selected < 0 || d.selected >= len(d.flat) {
		return
	}
	node := d.flat[d.selected].node
	if node.expanded {
		node.expanded = false
	} else {
		d.loadChildren(node)
		node.expanded = len(node.children) > 0
	}
	d.rebuildFlat()
	d.syncNameToSelection()
}

func (d *changeDirDialog) collapseOrGoParent() {
	if d.selected < 0 || d.selected >= len(d.flat) {
		return
	}
	node := d.flat[d.selected].node
	if node.expanded {
		node.expanded = false
		d.rebuildFlat()
		return
	}
	// Navigate to parent in the flat list
	if node.parent != nil {
		for i, e := range d.flat {
			if e.node == node.parent {
				d.selected = i
				d.ensureVisible()
				d.syncNameToSelection()
				return
			}
		}
	}
}

func (d *changeDirDialog) activateFocused() {
	switch d.focusField {
	case chdirFocusOk, chdirFocusChdir:
		d.doChdir()
	case chdirFocusRevert:
		d.doRevert()
	case chdirFocusHelp:
		// placeholder
	case chdirFocusTree:
		d.toggleExpand()
	}
}

// ── Mouse handler ─────────────────────────────────────────────────────────────

func (d *changeDirDialog) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		x, y, width, height := d.GetRect()
		if mx < x || mx >= x+width || my < y || my >= y+height {
			return false, nil
		}
		if action == tview.MouseScrollUp {
			if d.treeOffset > 0 {
				d.treeOffset--
			}
			return true, nil
		}
		if action == tview.MouseScrollDown {
			if d.treeOffset < len(d.flat)-chdirListRows {
				d.treeOffset++
			}
			return true, nil
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

		// ↓ history arrow
		if my == y+chdirRowInput && mx == x+chdirArrowX {
			d.showHistory = !d.showHistory
			return true, nil
		}

		// Input field
		if my == y+chdirRowInput && mx >= x+chdirInputX && mx < x+chdirInputX+chdirInputW {
			d.focusField = chdirFocusName
			offset := 0
			if d.nameCursor >= chdirInputW {
				offset = d.nameCursor - chdirInputW + 1
			}
			pos := (mx - (x + chdirInputX)) + offset
			if pos < 0 {
				pos = 0
			}
			if pos > len(d.nameText) {
				pos = len(d.nameText)
			}
			d.nameCursor = pos
			return true, nil
		}

		// Vertical scrollbar
		if mx == x+chdirScrollX {
			topY    := y + chdirRowListTop
			bottomY := topY + chdirListRows - 1
			if my == topY {
				if d.treeOffset > 0 {
					d.treeOffset--
				}
			} else if my == bottomY {
				if d.treeOffset < len(d.flat)-chdirListRows {
					d.treeOffset++
				}
			} else if my > topY && my < bottomY {
				trackH := chdirListRows - 2
				total  := len(d.flat)
				if total > chdirListRows && trackH > 0 {
					d.treeOffset = (my - topY - 1) * (total - chdirListRows) / trackH
					if d.treeOffset < 0 {
						d.treeOffset = 0
					}
				}
			}
			return true, nil
		}

		// Tree list rows
		if my >= y+chdirRowListTop && my < y+chdirRowListTop+chdirListRows {
			if mx >= x+chdirListX && mx < x+chdirListX+chdirListW {
				row := my - (y + chdirRowListTop)
				idx := d.treeOffset + row
				if idx >= 0 && idx < len(d.flat) {
					d.focusField = chdirFocusTree
					if idx == d.selected {
						// Second click on same item: toggle expand
						d.toggleExpand()
					} else {
						d.selected = idx
						d.syncNameToSelection()
					}
				}
				return true, nil
			}
		}

		// Buttons
		btnX := x + chdirBtnX
		for i, row := range []int{chdirBtnRowOk, chdirBtnRowChdir, chdirBtnRowRevert, chdirBtnRowHelp} {
			if my == y+row && mx >= btnX && mx < btnX+d.btnOk.width() {
				d.focusField = chdirFocusOk + i
				d.activateFocused()
				return true, nil
			}
		}
		return true, nil
	}
}

func (d *changeDirDialog) PasteHandler() func(string, func(tview.Primitive)) {
	return func(text string, setFocus func(tview.Primitive)) {
		if d.focusField != chdirFocusName {
			return
		}
		for _, r := range text {
			if unicode.IsPrint(r) {
				d.nameText = append(d.nameText[:d.nameCursor], append([]rune{r}, d.nameText[d.nameCursor:]...)...)
				d.nameCursor++
			}
		}
	}
}

// ── Actions ───────────────────────────────────────────────────────────────────

func (d *changeDirDialog) doChdir() {
	target := strings.TrimSpace(string(d.nameText))
	if target == "" {
		return
	}
	if err := os.Chdir(target); err != nil {
		if d.app.StatusBar != nil {
			d.app.StatusBar.SetText(" [red]Erro ao mudar diretório: " + err.Error() + "[-]")
		}
		return
	}
	d.addToHistory(target)
	d.close()
}

func (d *changeDirDialog) doRevert() {
	d.nameText  = []rune(d.origDir)
	d.nameCursor = len(d.nameText)
	d.expandToPath(d.origDir)
}

func (d *changeDirDialog) addToHistory(path string) {
	// Remove duplicate if present
	for i, h := range d.history {
		if strings.EqualFold(h, path) {
			d.history = append(d.history[:i], d.history[i+1:]...)
			break
		}
	}
	// Prepend
	d.history = append([]string{path}, d.history...)
	if len(d.history) > 16 {
		d.history = d.history[:16]
	}
}

func (d *changeDirDialog) close() {
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

func showChangeDirDialog(app *App) {
	dlg := newChangeDirDialog(app)
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
			AddItem(dlg, chdirDlgW, 0, true).
			AddItem(nil, 0, 1, false), chdirDlgH, 0, true).
		AddItem(nil, 0, 1, false)

	app.Pages.AddPage(dlg.pageName, container, true, true)
	app.Application.SetFocus(dlg)
}
