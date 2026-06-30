package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"msxedit/internal/basic"
)

// PrinterConfig armazena as preferências de impressão.
type PrinterConfig struct {
	WrapColumn   int    // coluna de quebra de linha (0 = desativado)
	LinesPerPage int    // linhas por página
	LineNumbers  bool   // imprimir números de linha
	FontSize     int    // tamanho da fonte em pontos (6, 8, 10, 12)
	FontName     string // nome da fonte
}

type App struct {
	Application  *tview.Application
	Pages        *tview.Pages
	Editor       *tview.TextArea
	Editors      []*editorWindow
	HelpWindows  []*helpWindow
	ActiveEditor int
	Menu         *tview.List
	StatusBar    *tview.TextView
	Version      string
	Build        string
	Theme        Theme
	NextWindowID int
	UsedWindowIDs map[int]bool
	CompilerMode  int          // 0 = MSX-BASIC (default), espelha radioIndex do CompilerOptionsDialog
	Printer       PrinterConfig // configurações de impressão
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
		Application:   tview.NewApplication(),
		Pages:         tview.NewPages(),
		Version:       version,
		Build:         build,
		Theme:         GetTheme(themeName),
		NextWindowID:  2,
		UsedWindowIDs: map[int]bool{1: true}, // ID 1 reservado para o primeiro editor
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
	editorWin.app = a
	editorWin.highlightEnabled = true
	editorWin.onClose = func() {
		a.Application.Stop()
	}
	editorWin.onExitToMenu = func() {
		menu.open(0) // foca primeiro item do menu (File)
	}
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

	pages.AddPage("editor", editorWin, true, true)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(a.StatusBar, 1, 0, false)

	a.Pages.AddPage("main", layout, true, true)

	// Captura de Teclas (Hotkeys para Menus)
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF1 && event.Modifiers()&tcell.ModAlt != 0 {
			if helpWin := a.currentHelpWindow(); helpWin != nil {
				if helpWin.goBackBreadcrumb() {
					return nil
				}
			}
		}
		if event.Key() == tcell.KeyF1 && event.Modifiers() == 0 {
			a.showHelpWindow()
			return nil
		}
		if event.Key() == tcell.KeyF2 && event.Modifiers() == 0 {
			a.showSave()
			return nil
		}
		if event.Key() == tcell.KeyF3 && event.Modifiers() == 0 {
			showOpenFileDialog(a, "*.BAS",
				func(path string) { a.openFile(path) },
				func(path string) { a.openFile(path) })
			return nil
		}
		if menu.handleHotkey(event) {
			return nil
		}
		return event
	})

	// Habilitar suporte a mouse
	a.Application.EnableMouse(true)

	// Interceptar cliques na barra de menus (linha 0 da tela)
	a.Application.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if action == tview.MouseLeftClick {
			mx, my := event.Position()
			if my == 0 {
				if idx := menuClickIndex(menu.menus, mx); idx >= 0 {
					menu.toggleFromHotkey(idx)
					return nil, action
				}
			}
		}
		return event, action
	})

	a.Application.SetFocus(editorWin)
	return a.Application.SetRoot(a.Pages, true).Run()
}

// showChangeDir abre o diálogo "Change Directory".
func (a *App) showChangeDir() {
	showChangeDirDialog(a)
}

// showPrinterSetup abre o diálogo "Printer Setup".
func (a *App) showPrinterSetup() {
	showPrinterSetupDialog(a)
}

// showSaveAll salva todos os editores abertos.
// Editores com caminho definido são salvos direto; sem nome abrem Save As sequencialmente.
func (a *App) showSaveAll() {
	savedCount := 0
	for _, ed := range a.Editors {
		if ed.filePath == "" {
			continue
		}
		text := ed.editor.GetText()
		var data []byte
		if ed.isTokenized {
			var tokErr error
			data, tokErr = basic.Tokenize(text)
			if tokErr != nil {
				if a.StatusBar != nil {
					a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao tokenizar %s: %v[-]", ed.fileName, tokErr))
				}
				continue
			}
		} else {
			data = []byte(text)
		}
		if err := os.WriteFile(ed.filePath, data, 0644); err != nil {
			if a.StatusBar != nil {
				a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao salvar %s: %v[-]", ed.fileName, err))
			}
		} else {
			savedCount++
		}
	}

	var queue []*editorWindow
	for _, ed := range a.Editors {
		if ed.filePath == "" {
			queue = append(queue, ed)
		}
	}

	if len(queue) == 0 {
		if a.StatusBar != nil {
			a.StatusBar.SetText(fmt.Sprintf(" %d arquivo(s) salvo(s).", savedCount))
		}
		return
	}
	a.saveNextUnnamed(queue, 0, savedCount)
}

// saveNextUnnamed exibe o diálogo Save As para cada editor sem nome da fila, em sequência.
func (a *App) saveNextUnnamed(queue []*editorWindow, idx int, alreadySaved int) {
	if idx >= len(queue) {
		if a.StatusBar != nil {
			a.StatusBar.SetText(fmt.Sprintf(" %d arquivo(s) salvo(s).", alreadySaved))
		}
		return
	}
	ed := queue[idx]
	showSaveFileDialogForEditor(a, ed, func() {
		extra := 0
		if ed.filePath != "" {
			extra = 1
		}
		a.saveNextUnnamed(queue, idx+1, alreadySaved+extra)
	})
}

func (a *App) showAbout() {
	dialog := newAboutDialog(a)
	const dlgW = 40
	const dlgH = 12
	showDialogoOKCentered(dialog.dialogoOK, dlgW, dlgH)
}

func (a *App) showCompilerInterpreterOptions() {
	dialog := newCompilerOptionsDialog(a)
	const dlgW = 72
	const dlgH = 18
	showCompilerOptionsDialogCentered(dialog, dlgW, dlgH)
}

func (a *App) showCompilerOptionsHelp() {
	text := []string{
		"Compiler/Interpreter Options",
		"",
		"Use TAB to move between controls.",
		"Use SPACE to toggle options.",
		"",
		"Detailed help will be added in a future update.",
	}
	dlg := newDialogoOK(a, "compiler_options_help", text, nil)
	dlg.SetButton("O&K", 'k', nil)
	dlg.SetButtonShadowMode(shadowModeTurboClassic)
	showDialogoOKCentered(dlg, 52, 13)
}

// getNextWindowID retorna o próximo ID de janela disponível
func (a *App) getNextWindowID() int {
	// Procura por um ID reutilizável
	for i := 1; i < a.NextWindowID; i++ {
		if !a.UsedWindowIDs[i] {
			a.UsedWindowIDs[i] = true
			return i
		}
	}
	// Se nenhum ID reutilizável, usa o próximo
	id := a.NextWindowID
	a.NextWindowID++
	a.UsedWindowIDs[id] = true
	return id
}

// releaseWindowID marca um ID como disponível para reutilização
func (a *App) releaseWindowID(id int) {
	delete(a.UsedWindowIDs, id)
}

// showSaveAs abre o diálogo "Save File As" para o editor ativo.
func (a *App) showSaveAs() {
	name := ""
	if len(a.Editors) > 0 && a.ActiveEditor >= 0 && a.ActiveEditor < len(a.Editors) {
		fn := a.Editors[a.ActiveEditor].fileName
		if fn != "" && fn != "Sem Nome" {
			name = fn
		}
	}
	showSaveFileDialog(a, name)
}

// openFile carrega um arquivo no editor ativo.
// Detecta arquivos tokenizados MSX-BASIC (0xFF) e detokeniza em memória.
func (a *App) openFile(path string) {
	if len(a.Editors) == 0 || a.ActiveEditor < 0 || a.ActiveEditor >= len(a.Editors) {
		return
	}
	ed := a.Editors[a.ActiveEditor]
	data, err := os.ReadFile(path)
	if err != nil {
		if a.StatusBar != nil {
			a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao abrir: %v[-]", err))
		}
		return
	}
	if basic.IsTokenized(data) {
		text, err2 := basic.DetokenizeToText(data)
		if err2 != nil {
			if a.StatusBar != nil {
				a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao detokenizar: %v[-]", err2))
			}
			return
		}
		ed.editor.SetText(text, true)
		ed.isTokenized = true
		a.CompilerMode = 0 // MSX-BASIC
	} else {
		ed.editor.SetText(string(data), true)
		ed.isTokenized = false
	}
	ed.filePath = path
	ed.fileName = filepath.Base(path)
	a.Application.SetFocus(ed)
}

// showOpenFile abre o diálogo "Open a File".
func (a *App) showOpenFile() {
	showOpenFileDialog(a, "*.BAS",
		func(path string) { a.openFile(path) },
		func(path string) { a.openFile(path) })
}

// showSave salva o arquivo ativo diretamente se já tem caminho; caso contrário abre "Save As".
// Re-tokeniza arquivos MSX-BASIC tokenizados antes de gravar.
func (a *App) showSave() {
	if len(a.Editors) == 0 || a.ActiveEditor < 0 || a.ActiveEditor >= len(a.Editors) {
		return
	}
	ed := a.Editors[a.ActiveEditor]
	if ed.filePath == "" {
		a.showSaveAs()
		return
	}
	text := ed.editor.GetText()
	var data []byte
	if ed.isTokenized {
		var err error
		data, err = basic.Tokenize(text)
		if err != nil {
			if a.StatusBar != nil {
				a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao tokenizar: %v[-]", err))
			}
			return
		}
	} else {
		data = []byte(text)
	}
	if err := os.WriteFile(ed.filePath, data, 0644); err != nil {
		if a.StatusBar != nil {
			a.StatusBar.SetText(fmt.Sprintf(" [red]Erro ao salvar: %v[-]", err))
		}
		return
	}
	if a.StatusBar != nil {
		a.StatusBar.SetText(fmt.Sprintf(" Salvo: %s", ed.filePath))
	}
}

// showHelpWindow abre uma janela de Help
func (a *App) showHelpWindow() {
	helpWin := newHelpWindow(a.Theme, a.getNextWindowID())
	a.HelpWindows = append(a.HelpWindows, helpWin)

	const dlgW = 72
	const dlgH = 22

	pageName := fmt.Sprintf("help_%d", helpWin.number)

	// Callback para fechar a janela
	helpWin.onClose = func() {
		a.Pages.RemovePage(pageName)
		a.releaseWindowID(helpWin.number)
		for i, hw := range a.HelpWindows {
			if hw == helpWin {
				a.HelpWindows = append(a.HelpWindows[:i], a.HelpWindows[i+1:]...)
				break
			}
		}
		a.Application.SetFocus(a.Editor)
	}

	// Layout centralizado (mesmo padrão do showDialogoOKCentered)
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(helpWin, dlgW, 0, true).
			AddItem(nil, 0, 1, false), dlgH, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage(pageName, container, true, true)
	a.Application.SetFocus(helpWin)
}

func (a *App) currentHelpWindow() *helpWindow {
	if len(a.HelpWindows) == 0 {
		return nil
	}
	return a.HelpWindows[len(a.HelpWindows)-1]
}

