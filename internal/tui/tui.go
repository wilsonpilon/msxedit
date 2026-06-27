package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	Application  *tview.Application
	Pages        *tview.Pages
	Editor       *tview.TextArea
	Editors      []*editorWindow
	ActiveEditor int
	Menu         *tview.List
	StatusBar    *tview.TextView
	Version      string
	Build        string
	Theme        Theme
}

type checkerboardDesktop struct {
	*tview.Box
	theme Theme
}

func newCheckerboardDesktop(t Theme) *checkerboardDesktop {
	return &checkerboardDesktop{
		Box:   tview.NewBox().SetBackgroundColor(t.DesktopBg),
		theme: t,
	}
}

func (d *checkerboardDesktop) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetRect()
	style := tcell.StyleDefault.Foreground(d.theme.DesktopFg).Background(d.theme.DesktopBg)

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			screen.SetContent(x+col, y+row, d.theme.DesktopChar, nil, style)
		}
	}
}

func NewApp(version, build string, themeName string) *App {
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		Version:     version,
		Build:       build,
		Theme:       GetTheme(themeName),
	}
}

func (a *App) Run(filePath string) error {
	// Desktop quadriculado estilo DOS/Norton, preenchendo a área inteira.
	desktop := newCheckerboardDesktop(a.Theme)

	// Menu Superior com drop-down navegável.
	menuBar := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	menuBar.SetBackgroundColor(a.Theme.MenuBarBg)
	menuBar.SetTextColor(a.Theme.MenuBarFg)
	menu := newMenuController(a, menuBar)
	menu.renderBar()

	name := "Sem Nome"
	if filePath != "" {
		name = filePath
	}
	editorWin := newEditorWindow(a.Theme, name, 1)
	a.Editors = []*editorWindow{editorWin}
	a.ActiveEditor = 0
	a.Editor = editorWin.editor

	// Barra de Status inferior
	a.StatusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [red]F1[-] Help  [red]F2[-] Save  [red]F3[-] Open  [red]Alt+[F9][-] Compile  [red]F9[-] Make  [red]Alt+[F10][-] Local menu")
	a.StatusBar.SetBackgroundColor(a.Theme.StatusBarBg)
	a.StatusBar.SetTextColor(a.Theme.StatusBarFg)

	// Layout principal
	pages := tview.NewPages().
		AddPage("desktop", desktop, true, true)

	editorHost := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 2, 0, false).
			AddItem(editorWin, 0, 1, true).
			AddItem(nil, 2, 0, false), 0, 1, true).
		AddItem(nil, 1, 0, false)

	pages.AddPage("editor", editorHost, true, true)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(a.StatusBar, 1, 0, false)

	a.Pages.AddPage("main", layout, true, true)

	// Captura de Teclas (Hotkeys para Menus)
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if menu.handleHotkey(event) {
			return nil
		}
		return event
	})

	a.Application.SetFocus(editorWin)
	return a.Application.SetRoot(a.Pages, true).Run()
}

func (a *App) showFileMenu() {
	list := tview.NewList().
		AddItem(" ──────────────────", "", 0, nil).
		AddItem(" E[red]x[-]it       Alt-X", "", 0, func() {
			a.Application.Stop()
		})

	list.SetBorder(true).
		SetBorderStyle(tcell.StyleDefault.Foreground(a.Theme.PopupBorderFg).Background(a.Theme.PopupBg)).
		SetBorderColor(a.Theme.PopupBorderFg)

	list.SetBackgroundColor(a.Theme.PopupBg)
	list.SetMainTextColor(a.Theme.PopupFg)
	list.SetSelectedBackgroundColor(a.Theme.PopupSelectedBg)
	list.SetSelectedTextColor(a.Theme.PopupSelectedFg)
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	// Habilitar cores dinâmicas para o destaque da letra
	list.SetSelectedStyle(tcell.StyleDefault.Background(a.Theme.PopupSelectedBg).Foreground(a.Theme.PopupSelectedFg))
	list.SetMainTextStyle(tcell.StyleDefault.Background(a.Theme.PopupBg).Foreground(a.Theme.PopupFg))

	// Captura de teclas interna do menu para atalhos sem Alt
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'x', 'X':
				a.Application.Stop()
				return nil
			}
		}
		return event
	})

	// Sombras e Posicionamento
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		shadow := tview.NewBox().SetBackgroundColor(a.Theme.ShadowBg)

		return tview.NewFlex().
			AddItem(nil, 2, 0, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 1, 0, false).
				AddItem(tview.NewFlex().
					AddItem(p, width, 0, true).
					AddItem(shadow, 1, 0, false), height, 0, true).
				AddItem(tview.NewFlex().
					AddItem(nil, 1, 0, false).
					AddItem(shadow, width, 0, false), 1, 0, false).
				AddItem(nil, 0, 1, false), width+1, 1, true).
			AddItem(nil, 0, 1, false)
	}

	a.Pages.AddPage("file_menu", modal(list, 24, 4), true, true)
	a.Application.SetFocus(list)

	list.SetDoneFunc(func() {
		a.Pages.RemovePage("file_menu")
		a.Application.SetFocus(a.Editor)
	})
}

func (a *App) showHelpMenu() {
	list := tview.NewList().
		AddItem(" ──────────────────", "", 0, nil).
		AddItem(" [red]A[-]bout", "", 0, func() {
			a.Pages.RemovePage("help_menu")
			a.showAbout()
		})

	list.SetBorder(true).
		SetBorderStyle(tcell.StyleDefault.Foreground(a.Theme.PopupBorderFg).Background(a.Theme.PopupBg)).
		SetBorderColor(a.Theme.PopupBorderFg)

	list.SetBackgroundColor(a.Theme.PopupBg)
	list.SetMainTextColor(a.Theme.PopupFg)
	list.SetSelectedBackgroundColor(a.Theme.PopupSelectedBg)
	list.SetSelectedTextColor(a.Theme.PopupSelectedFg)
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	// Habilitar cores dinâmicas para o destaque da letra
	list.SetSelectedStyle(tcell.StyleDefault.Background(a.Theme.PopupSelectedBg).Foreground(a.Theme.PopupSelectedFg))
	list.SetMainTextStyle(tcell.StyleDefault.Background(a.Theme.PopupBg).Foreground(a.Theme.PopupFg))

	// Captura de teclas interna do menu para atalhos sem Alt
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'a', 'A':
				a.Pages.RemovePage("help_menu")
				a.showAbout()
				return nil
			}
		}
		return event
	})

	// Sombras e Posicionamento (ajustado para ficar sob Help)
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		shadow := tview.NewBox().SetBackgroundColor(a.Theme.ShadowBg)

		return tview.NewFlex().
			AddItem(nil, 32, 0, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 1, 0, false).
				AddItem(tview.NewFlex().
					AddItem(p, width, 0, true).
					AddItem(shadow, 1, 0, false), height, 0, true).
				AddItem(tview.NewFlex().
					AddItem(nil, 1, 0, false).
					AddItem(shadow, width, 0, false), 1, 0, false).
				AddItem(nil, 0, 1, false), width+1, 1, true).
			AddItem(nil, 0, 1, false)
	}

	a.Pages.AddPage("help_menu", modal(list, 24, 4), true, true)
	a.Application.SetFocus(list)

	list.SetDoneFunc(func() {
		a.Pages.RemovePage("help_menu")
		a.Application.SetFocus(a.Editor)
	})
}

func (a *App) showAbout() {
	dialog := newAboutDialog(a)
	const dlgW = 40
	const dlgH = 12
	showDialogoOKCentered(dialog.dialogoOK, dlgW, dlgH)
}
