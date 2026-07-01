package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	Clipboard       string        // clipboard compartilhado entre todas as janelas de edição
	ClipboardWindow *editorWindow // janela "Show clipboard" aberta (nil se fechada)

	OpenMenu func(index int) // abre a barra de menu (Ctrl+K D em qualquer janela flutuante)

	LastFind    FindParams // últimas opções usadas no diálogo Find (Search menu)
	FindHistory []string   // histórico de textos buscados (compartilhado com Replace)

	LastReplace       ReplaceParams // últimas opções usadas no diálogo Replace
	ReplaceNewHistory []string      // histórico do campo "New text" do Replace

	LastGotoLine    GotoLineParams // últimas opções usadas no diálogo Go to Line Number
	GotoLineHistory []string       // histórico de números de linha
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
	a.OpenMenu = menu.open

	// Barra de Status inferior
	a.StatusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [red]F1[-] Help  [red]F2[-] Save  [red]F3[-] Open  [red]Alt+[F9][-] Compile  [red]F9[-] Make  [red]Alt+[F10][-] Local menu")
	a.StatusBar.SetBackgroundColor(a.Theme.StatusBarBg)
	a.StatusBar.SetTextColor(a.Theme.StatusBarFg)

	// Layout principal: desktop quadriculado ao fundo; janelas de edição/ajuda/
	// clipboard flutuam por cima como páginas independentes de a.Pages, então
	// fechar todas elas apenas revela o desktop — o app continua rodando.
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(desktop, 0, 1, true).
		AddItem(a.StatusBar, 1, 0, false)

	a.Pages.AddPage("main", layout, true, true)

	name := "Sem Nome"
	if filePath != "" {
		name = filePath
	}
	editorWin := a.createEditorWindow(name)
	a.Editors = []*editorWindow{editorWin}
	a.ActiveEditor = 0
	a.Editor = editorWin.editor
	a.Pages.AddPage(editorPageName(editorWin.number), editorWin, true, true)

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
		if event.Key() == tcell.KeyF3 && event.Modifiers()&tcell.ModAlt != 0 {
			a.closeActiveWindow()
			return nil
		}
		if event.Key() == tcell.KeyF1 && event.Modifiers()&tcell.ModCtrl != 0 {
			a.showLanguageHelp()
			return nil
		}
		if event.Key() == tcell.KeyCtrlL {
			a.searchAgain()
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

// showFindDialog abre o diálogo "Find" (Search > Find...), pré-preenchido
// com as opções e o histórico da última busca.
func (a *App) showFindDialog() {
	dialog := newFindDialog(a, a.LastFind, a.FindHistory)
	dialog.onFind = func(params FindParams) {
		a.LastFind = params
		a.FindHistory = dialog.field.history
		a.performFind(params)
	}
	showFindDialogCentered(dialog)
}

// performFind executa uma busca no editor ativo e atualiza a barra de status
// de acordo com o resultado (encontrado, não encontrado, busca reiniciada ou
// expressão regular inválida).
func (a *App) performFind(p FindParams) {
	ed := a.activeEditor()
	if ed == nil {
		return
	}
	found, wrapped, err := ed.findNext(p)
	if a.StatusBar == nil {
		return
	}
	switch {
	case err != nil:
		a.StatusBar.SetText(fmt.Sprintf(" [red]Expressão regular inválida: %v[-]", err))
	case !found:
		a.StatusBar.SetText(fmt.Sprintf(" [red]Texto não encontrado: %s[-]", p.Text))
	case wrapped:
		a.StatusBar.SetText(" Busca reiniciada (fim do texto alcançado).")
	default:
		ed.restoreStatusBar()
	}
}

// searchAgain implementa Search > Search again: repete a última busca
// (mesmo texto, mesmas opções, mesmo sentido) a partir da posição atual do
// cursor. Sem busca anterior, abre o diálogo Find.
func (a *App) searchAgain() {
	if strings.TrimSpace(a.LastFind.Text) == "" {
		a.showFindDialog()
		return
	}
	a.performFind(a.LastFind)
}

// showGotoLineDialog abre o diálogo "Go to Line Number" (Search > Go to
// line number...), pré-preenchido com o último número/opção usados.
func (a *App) showGotoLineDialog() {
	dialog := newGotoLineDialog(a, a.LastGotoLine, a.GotoLineHistory)
	dialog.onGoto = func(params GotoLineParams) {
		a.LastGotoLine = params
		a.GotoLineHistory = dialog.field.history
		a.performGotoLine(params)
	}
	showGotoLineDialogCentered(dialog)
}

// performGotoLine move o cursor do editor ativo para a linha pedida — de
// texto ou do MSX-Basic, conforme params.MSXBasic — e atualiza a barra de
// status quando a linha não é encontrada.
func (a *App) performGotoLine(params GotoLineParams) {
	ed := a.activeEditor()
	if ed == nil {
		return
	}
	var ok bool
	if params.MSXBasic {
		ok = ed.gotoBasicLine(params.Line)
	} else {
		ok = ed.gotoTextLine(params.Line)
	}
	if a.StatusBar == nil {
		return
	}
	if !ok {
		a.StatusBar.SetText(fmt.Sprintf(" [red]Linha não encontrada: %d[-]", params.Line))
		return
	}
	ed.restoreStatusBar()
}

// showReplaceDialog abre o diálogo "Replace" (Search > Replace...),
// pré-preenchido com a última busca (texto/opções) e o histórico do campo
// "New text".
func (a *App) showReplaceDialog() {
	seed := a.LastReplace
	seed.FindParams = a.LastFind
	dialog := newReplaceDialog(a, seed, a.FindHistory, a.ReplaceNewHistory)
	dialog.onReplace = func(params ReplaceParams) {
		a.LastReplace = params
		a.LastFind = params.FindParams
		a.FindHistory = dialog.findField.history
		a.ReplaceNewHistory = dialog.newField.history
	}
	showReplaceDialogCentered(dialog)
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
	ed.resetPerFileState()
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
		a.focusActiveEditor()
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

// showLanguageHelp abre o Help já navegado até o tópico "Reserved Words"
// (Ctrl+F1 — "Language help"), a referência mais próxima de ajuda contextual
// sobre a linguagem MSX-BASIC disponível hoje.
func (a *App) showLanguageHelp() {
	a.showHelpWindow()
	if helpWin := a.currentHelpWindow(); helpWin != nil {
		helpWin.content.NavigateToTopic("reserved_words")
	}
}

// closeActiveWindow fecha a janela de edição ativa (Alt+F3), equivalente a
// clicar no botão [■] dela.
func (a *App) closeActiveWindow() {
	ed := a.activeEditor()
	if ed == nil || ed.onClose == nil {
		return
	}
	ed.onClose()
}

func (a *App) currentHelpWindow() *helpWindow {
	if len(a.HelpWindows) == 0 {
		return nil
	}
	return a.HelpWindows[len(a.HelpWindows)-1]
}

// activeEditor retorna a janela de edição ativa, ou nil se não houver.
func (a *App) activeEditor() *editorWindow {
	if len(a.Editors) == 0 || a.ActiveEditor < 0 || a.ActiveEditor >= len(a.Editors) {
		return nil
	}
	return a.Editors[a.ActiveEditor]
}

// focusActiveEditor devolve o foco para a janela de edição ativa. Se não
// houver nenhuma (todas foram fechadas), o foco vai para o desktop — o app
// continua rodando normalmente, pronto para File > New ou File > Open.
func (a *App) focusActiveEditor() {
	if ed := a.activeEditor(); ed != nil {
		a.Application.SetFocus(ed)
		return
	}
	a.Application.SetFocus(a.Pages)
}

// editorPageName retorna o nome da página usada para uma janela de edição
// com o ID informado dentro de a.Pages.
func editorPageName(number int) string {
	return fmt.Sprintf("editor_%d", number)
}

// createEditorWindow cria uma nova janela de edição já ligada à aplicação
// (clipboard compartilhado, atalhos, fechar/ativar), mas ainda não adicionada
// a a.Pages nem a a.Editors — isso fica a cargo de quem a criar (Run,
// showNewEditor), que também decide a posição inicial.
func (a *App) createEditorWindow(name string) *editorWindow {
	win := newEditorWindow(a.Theme, name, a.getNextWindowID())
	win.app = a
	win.highlightEnabled = true
	win.onClose = func() {
		a.closeEditorWindow(win)
	}
	win.onExitToMenu = func() {
		if a.OpenMenu != nil {
			a.OpenMenu(0)
		}
	}
	win.onFocus = func() {
		for i, ed := range a.Editors {
			if ed == win {
				a.ActiveEditor = i
				return
			}
		}
	}
	return win
}

// closeEditorWindow remove uma janela de edição da tela e da lista a.Editors.
// Fechar a última janela aberta NÃO encerra o programa: o desktop fica visível
// e o usuário pode continuar usando File > New ou File > Open normalmente.
func (a *App) closeEditorWindow(win *editorWindow) {
	a.Pages.RemovePage(editorPageName(win.number))
	a.releaseWindowID(win.number)

	for i, ed := range a.Editors {
		if ed == win {
			a.Editors = append(a.Editors[:i], a.Editors[i+1:]...)
			break
		}
	}
	if a.ActiveEditor >= len(a.Editors) {
		a.ActiveEditor = len(a.Editors) - 1
	}
	if ed := a.activeEditor(); ed != nil {
		a.Editor = ed.editor
	} else {
		a.Editor = nil
	}
	a.focusActiveEditor()
}

// showNewEditor implementa File > New: cria uma janela de edição em branco,
// em cascata a partir da janela ativa atual.
func (a *App) showNewEditor() {
	from := a.activeEditor()

	win := a.createEditorWindow("Sem Nome")
	width, height := 0, 0
	if from != nil {
		width, height = from.winW, from.winH
	}
	x, y, w, h := a.cascadePosition(from, width, height)
	win.winX, win.winY, win.winW, win.winH = x, y, w, h
	win.savedX, win.savedY, win.savedW, win.savedH = x, y, w, h
	win.positioned = true

	a.Editors = append(a.Editors, win)
	a.ActiveEditor = len(a.Editors) - 1

	a.Pages.AddPage(editorPageName(win.number), win, true, true)
	a.Application.SetFocus(win)
}

// cascadePosition calcula a posição/tamanho de uma nova janela flutuante a
// partir de uma janela de referência ("from"), deslocando-a em cascata e
// dando a volta quando não há mais espaço na tela. Se width/height forem 0,
// usa um tamanho que preenche a maior parte da tela.
func (a *App) cascadePosition(from *editorWindow, width, height int) (x, y, w, h int) {
	sw, sh := 80, 25
	if from != nil && from.lastScreenW > 0 && from.lastScreenH > 0 {
		sw, sh = from.lastScreenW, from.lastScreenH
	}

	w, h = width, height
	if w <= 0 {
		w = sw - 4
	}
	if h <= 0 {
		h = sh - 3
	}
	if w > sw-2 {
		w = sw - 2
	}
	if h > sh-2 {
		h = sh - 2
	}

	const dx, dy = 3, 2
	x, y = 2, 1
	if from != nil {
		x, y = from.winX+dx, from.winY+dy
	}
	if x+w > sw || y+h > sh-1 {
		x, y = 2, 1
	}
	return
}

// syncClipboardWindow atualiza o conteúdo exibido na janela "Show clipboard"
// (se aberta) para refletir o clipboard compartilhado atual.
func (a *App) syncClipboardWindow() {
	if a.ClipboardWindow == nil {
		return
	}
	if a.ClipboardWindow.editor.GetText() != a.Clipboard {
		a.ClipboardWindow.editor.SetText(a.Clipboard, true)
	}
}

// showClipboardWindow abre (ou foca, se já aberta) a janela especial que
// exibe e edita o clipboard compartilhado entre todas as janelas de edição.
func (a *App) showClipboardWindow() {
	if a.ClipboardWindow != nil {
		a.Application.SetFocus(a.ClipboardWindow)
		return
	}

	win := newEditorWindow(a.Theme, "Clipboard", a.getNextWindowID())
	win.app = a
	win.isClipboard = true
	win.editor.SetText(a.Clipboard, true)
	win.editor.SetChangedFunc(func() {
		a.Clipboard = win.editor.GetText()
	})

	// Janela menor que uma janela de edição normal, em cascata a partir da
	// janela ativa (ou centralizada, se não houver nenhuma aberta).
	const clipboardW, clipboardH = 56, 18
	from := a.activeEditor()
	x, y, w, h := a.cascadePosition(from, clipboardW, clipboardH)
	win.winX, win.winY, win.winW, win.winH = x, y, w, h
	win.savedX, win.savedY, win.savedW, win.savedH = x, y, w, h
	win.positioned = true

	pageName := fmt.Sprintf("clipboard_%d", win.number)
	win.onClose = func() {
		a.Pages.RemovePage(pageName)
		a.releaseWindowID(win.number)
		a.ClipboardWindow = nil
		a.focusActiveEditor()
	}
	win.onExitToMenu = func() {
		if a.OpenMenu != nil {
			a.OpenMenu(0)
		}
	}

	a.ClipboardWindow = win
	a.Pages.AddPage(pageName, win, true, true)
	a.Application.SetFocus(win)
}

