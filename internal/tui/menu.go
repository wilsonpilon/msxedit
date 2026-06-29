package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MenuEntry struct {
	Label     string
	Shortcut  string
	Separator bool
	Action    func()
}

type MenuDefinition struct {
	Title string
	Items []MenuEntry
}

type menuController struct {
	app      *App
	bar      *tview.TextView
	menus    []MenuDefinition
	active   bool
	index    int
	popup    *tview.List
	pageName string
}

type shadowedList struct {
	*tview.Box
	list  *tview.List
	theme Theme
}

func newShadowedList(list *tview.List, theme Theme) *shadowedList {
	return &shadowedList{
		Box:   tview.NewBox(),
		list:  list,
		theme: theme,
	}
}

func (s *shadowedList) Draw(screen tcell.Screen) {
	x, y, width, height := s.GetRect()
	shadowStyle := tcell.StyleDefault.Background(s.theme.ShadowBg)
	maxX, maxY := screen.Size()

	// sombra à direita (2 colunas)
	for row := 1; row <= height && y+row < maxY; row++ {
		for col := 0; col < 2 && x+width+col < maxX; col++ {
			screen.SetContent(x+width+col, y+row, ' ', nil, shadowStyle)
		}
	}
	// sombra inferior (1 linha)
	for col := 2; col < width+2 && x+col < maxX; col++ {
		screen.SetContent(x+col, y+height, ' ', nil, shadowStyle)
	}

	s.list.SetRect(x, y, width, height)
	s.list.Draw(screen)
}

func (s *shadowedList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return s.list.InputHandler()
}

func (s *shadowedList) Focus(delegate func(p tview.Primitive)) {
	s.list.Focus(delegate)
}

func (s *shadowedList) HasFocus() bool {
	return s.list.HasFocus()
}

func (s *shadowedList) Blur() {
	s.list.Blur()
}

func (s *shadowedList) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return s.list.MouseHandler()
}

func (s *shadowedList) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return s.list.PasteHandler()
}

func newMenuController(app *App, bar *tview.TextView) *menuController {
	return &menuController{
		app:      app,
		bar:      bar,
		menus:    defaultMenus(app),
		pageName: "menu_popup",
	}
}

func defaultMenus(app *App) []MenuDefinition {
	return []MenuDefinition{
		{
			Title: "&File",
			Items: []MenuEntry{
				{Label: "&New"},
				{Label: "&Open...", Shortcut: "[F3]"},
				{Label: "&Save", Shortcut: "[F2]"},
				{Label: "Save &as..."},
				{Label: "Save a&ll"},
				{Separator: true},
				{Label: "&Change dir..."},
				{Label: "&Print"},
				{Label: "P&rinter setup..."},
				{Label: "&Powershell"},
				{Label: "E&xit", Shortcut: "[Alt]+[X]", Action: func() { app.Application.Stop() }},
			},
		},
		{Title: "&Edit", Items: []MenuEntry{{Label: "No options yet"}}},
		{Title: "&Search", Items: []MenuEntry{{Label: "No options yet"}}},
		{Title: "&Run", Items: []MenuEntry{{Label: "No options yet"}}},
		{Title: "&Compile", Items: []MenuEntry{{Label: "No options yet"}}},
		{Title: "&Debug", Items: []MenuEntry{{Label: "No options yet"}}},
		{Title: "&Tools", Items: []MenuEntry{{Label: "No options yet"}}},
		{
			Title: "&Options",
			Items: []MenuEntry{
				{Label: "&Compiler/Interpreter...", Action: func() { app.showCompilerInterpreterOptions() }},
				{Label: "&Memory Sizes..."},
				{Label: "&Linker..."},
				{Label: "De&bugger..."},
				{Label: "&Directories"},
				{Label: "&Tools"},
				{Separator: true},
				{Label: "&Environment", Shortcut: "►"},
				{Separator: true},
				{Label: "&Open..."},
				{Label: `&Save ...\msxedit.cfg`},
				{Label: "Save &as..."},
			},
		},
		{Title: "&Window", Items: []MenuEntry{{Label: "No options yet"}}},
		{
			Title: "&Help",
			Items: []MenuEntry{
				{Label: "&Contents", Action: func() { app.showHelpWindow() }},
				{Label: "&Index", Shortcut: "[Shift]+[F1]"},
				{Label: "&Topc search", Shortcut: "[Ctrl]+[F1]"},
				{Label: "&Previous topic", Shortcut: "[Alt]+[F1]"},
				{Label: "Using &help"},
				{Label: "&Files..."},
				{Separator: true},
				{Label: "Compiler &directives"},
				{Label: "&Reserved words"},
				{Label: "Standard &units"},
				{Label: "MSX Basic &Language"},
				{Label: "&Error messages"},
				{Separator: true},
				{Label: "&About...", Action: func() { app.showAbout() }},
			},
		},
	}
}

func (mc *menuController) renderBar() {
	parts := make([]string, 0, len(mc.menus))
	for i, menu := range mc.menus {
		parts = append(parts, renderMenuTitle(menu.Title, i == mc.index && mc.active, mc.app.Theme))
	}
	mc.bar.SetDynamicColors(true)
	mc.bar.SetBackgroundColor(mc.app.Theme.MenuBarBg)
	mc.bar.SetTextColor(mc.app.Theme.MenuBarFg)
	mc.bar.SetText("  " + strings.Join(parts, "  "))
}

func (mc *menuController) open(index int) {
	if index < 0 {
		index = len(mc.menus) - 1
	}
	if index >= len(mc.menus) {
		index = 0
	}
	mc.index = index
	mc.active = true
	mc.renderBar()
	mc.showPopup()
}

func (mc *menuController) close() {
	mc.active = false
	mc.removePopup()
	mc.renderBar()
	mc.app.Application.SetFocus(mc.app.Editor)
}

func (mc *menuController) toggleFromHotkey(index int) {
	if mc.active && mc.index == index {
		mc.close()
		return
	}
	mc.open(index)
}

func (mc *menuController) moveMenu(delta int) {
	if !mc.active {
		return
	}
	mc.index = (mc.index + delta + len(mc.menus)) % len(mc.menus)
	mc.showPopup()
	mc.renderBar()
}

func (mc *menuController) moveItem(delta int) {
	if mc.popup == nil {
		return
	}
	count := mc.popup.GetItemCount()
	if count == 0 {
		return
	}
	current := mc.popup.GetCurrentItem()
	for i := 0; i < count; i++ {
		current = (current + delta + count) % count
		if !mc.menus[mc.index].Items[current].Separator {
			mc.popup.SetCurrentItem(current)
			return
		}
	}
}

func (mc *menuController) activateCurrent() {
	if mc.popup == nil {
		return
	}
	idx := mc.popup.GetCurrentItem()
	if idx < 0 || idx >= len(mc.menus[mc.index].Items) {
		return
	}
	entry := mc.menus[mc.index].Items[idx]
	if entry.Action != nil {
		mc.close()
		entry.Action()
	}
}

func (mc *menuController) showPopup() {
	mc.removePopup()
	menu := mc.menus[mc.index]

	popup := tview.NewList()
	popup.SetBorder(true)
	popup.SetBorderStyle(tcell.StyleDefault.Foreground(mc.app.Theme.PopupBorderFg).Background(mc.app.Theme.PopupBg))
	popup.SetBorderColor(mc.app.Theme.PopupBorderFg)
	popup.SetBackgroundColor(mc.app.Theme.PopupBg)
	popup.SetMainTextColor(mc.app.Theme.PopupFg)
	popup.SetSelectedBackgroundColor(mc.app.Theme.MenuCursorBg)
	popup.SetSelectedTextColor(mc.app.Theme.MenuCursorFg)
	popup.SetHighlightFullLine(true)
	popup.ShowSecondaryText(false)
	popup.SetSelectedStyle(tcell.StyleDefault.Background(mc.app.Theme.MenuCursorBg).Foreground(mc.app.Theme.MenuCursorFg))
	popup.SetMainTextStyle(tcell.StyleDefault.Background(mc.app.Theme.PopupBg).Foreground(mc.app.Theme.PopupFg))

	width := popupWidth(menu)
	for _, item := range menu.Items {
		if item.Separator {
			popup.AddItem(strings.Repeat("─", maxInt(1, width-2)), "", 0, nil)
			continue
		}
		popup.AddItem(formatMenuItemWithShortcut(item.Label, item.Shortcut, width-2), "", 0, item.Action)
	}
	popup.SetCurrentItem(firstSelectableIndex(menu.Items))
	popup.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		mc.activateCurrent()
	})

	popup.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			mc.close()
			return nil
		}
		if event.Key() == tcell.KeyLeft {
			mc.moveMenu(-1)
			return nil
		}
		if event.Key() == tcell.KeyRight {
			mc.moveMenu(1)
			return nil
		}
		if event.Key() == tcell.KeyUp {
			mc.moveItem(-1)
			return nil
		}
		if event.Key() == tcell.KeyDown {
			mc.moveItem(1)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			mc.activateCurrent()
			return nil
		}
		return event
	})

	mc.popup = popup
	popupHeight := len(menu.Items) + 2
	popupRect := newShadowedList(popup, mc.app.Theme)
	popupRect.SetRect(menuX(mc.menus, mc.index), 1, width, popupHeight)
	mc.app.Pages.AddPage(mc.pageName, popupRect, false, true)
	mc.app.Application.SetFocus(popup)
}

func (mc *menuController) removePopup() {
	if mc.popup == nil {
		return
	}
	mc.app.Pages.RemovePage(mc.pageName)
	mc.popup = nil
}

func (mc *menuController) handleHotkey(event *tcell.EventKey) bool {
	if event == nil {
		return false
	}
	if event.Key() == tcell.KeyF10 {
		mc.toggleFromHotkey(0)
		return true
	}
	if event.Key() != tcell.KeyRune || event.Modifiers()&tcell.ModAlt == 0 {
		return false
	}
	switch event.Rune() {
	case 'x', 'X':
		mc.app.Application.Stop()
		return true
	case 'f', 'F':
		mc.toggleFromHotkey(0)
		return true
	case 'e', 'E':
		mc.toggleFromHotkey(1)
		return true
	case 's', 'S':
		mc.toggleFromHotkey(2)
		return true
	case 'r', 'R':
		mc.toggleFromHotkey(3)
		return true
	case 'c', 'C':
		mc.toggleFromHotkey(4)
		return true
	case 'd', 'D':
		mc.toggleFromHotkey(5)
		return true
	case 't', 'T':
		mc.toggleFromHotkey(6)
		return true
	case 'o', 'O':
		mc.toggleFromHotkey(7)
		return true
	case 'w', 'W':
		mc.toggleFromHotkey(8)
		return true
	case 'h', 'H':
		mc.toggleFromHotkey(9)
		return true
	}
	return false
}

func renderMenuTitle(label string, selected bool, theme Theme) string {
	plain, accel := stripAccelerator(label)
	runes := []rune(plain)
	if len(runes) == 0 {
		return ""
	}
	baseTag := fmt.Sprintf("[%s:%s]", theme.MenuBarFg.String(), theme.MenuBarBg.String())
	selectedTag := fmt.Sprintf("[%s:%s]", theme.MenuCursorFg.String(), theme.MenuCursorBg.String())
	accentBaseTag := fmt.Sprintf("[#ff0000:%s:b]", theme.MenuBarBg.String())
	accentSelectedTag := fmt.Sprintf("[#ff0000:%s:b]", theme.MenuCursorBg.String())
	activeTag := baseTag
	activeAccentTag := accentBaseTag
	if selected {
		activeTag = selectedTag
		activeAccentTag = accentSelectedTag
	}

	if accel < 0 || accel >= len(runes) {
		return fmt.Sprintf("%s %s[-]", activeTag, plain)
	}
	prefix := string(runes[:accel])
	letter := string(runes[accel])
	suffix := string(runes[accel+1:])
	return fmt.Sprintf("%s %s%s%s%s[-]", activeTag, prefix, activeAccentTag, letter, activeTag+suffix)
}

func stripAccelerator(label string) (string, int) {
	runes := []rune(label)
	out := make([]rune, 0, len(runes))
	accel := -1
	for i := 0; i < len(runes); i++ {
		if runes[i] == '&' && i+1 < len(runes) {
			accel = len(out)
			i++
			out = append(out, runes[i])
			continue
		}
		out = append(out, runes[i])
	}
	return string(out), accel
}

func popupWidth(menu MenuDefinition) int {
	width := 0
	for _, item := range menu.Items {
		if item.Separator {
			continue
		}
		plain, _ := stripAccelerator(item.Label)
		w := len([]rune(plain))
		if item.Shortcut != "" {
			w += 2 + len([]rune(item.Shortcut))
		}
		if w > width {
			width = w
		}
	}
	if width < 18 {
		width = 18
	}
	return width + 2
}

func firstSelectableIndex(items []MenuEntry) int {
	for i, item := range items {
		if !item.Separator {
			return i
		}
	}
	return 0
}

func menuX(menus []MenuDefinition, active int) int {
	x := 2
	for i := 0; i < active && i < len(menus); i++ {
		plain, _ := stripAccelerator(menus[i].Title)
		x += len([]rune(plain)) + 3
	}
	return x
}

// menuClickIndex retorna o índice do menu clicado na barra ou -1 se não há menu nessa posição.
func menuClickIndex(menus []MenuDefinition, clickX int) int {
	for i := range menus {
		startX := menuX(menus, i)
		plain, _ := stripAccelerator(menus[i].Title)
		// ocupa: espaço-antes + título (sem espaço-depois que vai para o separador)
		endX := startX + len([]rune(plain))
		if clickX >= startX && clickX <= endX {
			return i
		}
	}
	return -1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatMenuItemWithShortcut(label string, shortcut string, width int) string {
	plain, accel := stripAccelerator(label)
	runes := []rune(plain)

	// Renderizar label com acelerador em vermelho
	var labelText string
	if accel >= 0 && accel < len(runes) {
		prefix := string(runes[:accel])
		letter := string(runes[accel])
		suffix := string(runes[accel+1:])
		labelText = prefix + "[#ff0000::b]" + letter + "[-]" + suffix
	} else {
		labelText = plain
	}

	// Se não há atalho, apenas retorna o label
	if shortcut == "" {
		return labelText
	}

	// Calcular espaço disponível
	shortcutClean := tview.Escape(shortcut)
	labelLen := len(runes) // comprimento real sem tags
	shortcutLen := len([]rune(shortcutClean))
	availableSpace := width - labelLen - shortcutLen - 2 // -2 para espaçamento

	if availableSpace < 1 {
		availableSpace = 1
	}

	// Preencher com espaços e colocar atalho alinhado à direita na mesma cor do item.
	return labelText + strings.Repeat(" ", availableSpace) + shortcutClean
}
