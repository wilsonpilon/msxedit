package reader

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Options reúne os parâmetros de execução do msxread.
type Options struct {
	FilePath string
	Type     DocType
	TabSize  int
	Width    int // largura máxima para quebra de linha em markdown (0 = sem limite)
}

// Run carrega o documento e executa a interface do visualizador.
func Run(opts Options) error {
	doc, err := LoadDocument(opts.FilePath, opts.Type, opts.TabSize, opts.Width)
	if err != nil {
		return err
	}

	app      := tview.NewApplication()
	settings := LoadSettings()
	theme    := DefaultTheme()
	viewer   := NewViewer(theme, doc, settings)
	viewer.SetOnExit(func() { app.Stop() })

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Overlay de ajuda: qualquer tecla fecha.
		if viewer.IsHelpOpen() {
			viewer.CloseHelp()
			return nil
		}

		// ── Modo busca: captura toda entrada do teclado ──────────────────────
		if viewer.InFindMode() {
			switch event.Key() {
			case tcell.KeyEscape:
				viewer.CancelFind()
				return nil
			case tcell.KeyEnter:
				viewer.CommitFind()
				return nil
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				viewer.FindBackspace()
				return nil
			case tcell.KeyRune:
				viewer.AddFindChar(event.Rune())
				return nil
			}
			return nil
		}

		// ── Teclas especiais ─────────────────────────────────────────────────
		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyF1:
			viewer.ToggleHelp()
			return nil

		// Navegação
		case tcell.KeyUp:
			viewer.ScrollLines(-1)
			return nil
		case tcell.KeyDown:
			viewer.ScrollLines(1)
			return nil
		case tcell.KeyLeft:
			viewer.ScrollCols(-8)
			return nil
		case tcell.KeyRight:
			viewer.ScrollCols(8)
			return nil
		case tcell.KeyPgUp:
			viewer.PageUp()
			return nil
		case tcell.KeyPgDn:
			viewer.PageDown()
			return nil
		case tcell.KeyHome:
			viewer.Home()
			return nil
		case tcell.KeyEnd:
			viewer.End()
			return nil

		// Cores (F5/F6 = texto, F7/F8 = fundo)
		case tcell.KeyF5:
			viewer.CycleFg(-1)
			return nil
		case tcell.KeyF6:
			viewer.CycleFg(1)
			return nil
		case tcell.KeyF7:
			viewer.CycleBg(-1)
			return nil
		case tcell.KeyF8:
			viewer.CycleBg(1)
			return nil
		}

		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q', 'Q':
				app.Stop()
				return nil
			case 'w', 'W':
				viewer.ToggleWrap()
				return nil
			case '7':
				viewer.SetHiBit(false)
				return nil
			case '8':
				viewer.SetHiBit(true)
				return nil
			case 's', 'S':
				viewer.SaveCurrentSettings()
				return nil
			case 'p', 'P':
				viewer.PrintDoc(func(err error) {
					app.QueueUpdateDraw(func() {
						if err != nil {
							viewer.ShowStatus("Erro ao imprimir: "+err.Error(), 4*time.Second)
						} else {
							viewer.ShowStatus("Enviado para a impressora", 3*time.Second)
						}
					})
				})
				return nil
			case 'f', 'F':
				viewer.StartFind()
				return nil
			case 'n', 'N':
				viewer.FindNext()
				return nil
			case 'c', 'C':
				viewer.ToggleCaseSensitive()
				return nil
			}
		}
		return event
	})

	app.EnableMouse(true)

	// Relógio ao vivo: redesenha a cada segundo (atualiza hora + expira statusMsg).
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				app.QueueUpdateDraw(func() {})
			}
		}
	}()
	defer close(stop)

	return app.SetRoot(viewer, true).Run()
}
