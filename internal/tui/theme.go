package tui

import "github.com/gdamore/tcell/v2"

// Theme contém todas as cores da interface gráfica.
type Theme struct {
	Name string

	// Fundo quadriculado (desktop)
	DesktopBg   tcell.Color
	DesktopFg   tcell.Color
	DesktopChar rune

	// Barra de menu superior
	MenuBarBg    tcell.Color
	MenuBarFg    tcell.Color
	MenuCursorBg tcell.Color
	MenuCursorFg tcell.Color

	// Barra de status inferior
	StatusBarBg tcell.Color
	StatusBarFg tcell.Color

	// Janela do editor
	EditorBg       tcell.Color
	EditorFg       tcell.Color
	EditorBorderFg tcell.Color

	// Menus popup (File, Help, About)
	PopupBg         tcell.Color
	PopupFg         tcell.Color
	PopupBorderFg   tcell.Color
	PopupSelectedBg tcell.Color
	PopupSelectedFg tcell.Color
	ShadowBg        tcell.Color
}

// GetTheme retorna o tema pelo nome. Caso não encontrado, retorna o tema padrão.
func GetTheme(name string) Theme {
	switch name {
	case "blue":
		return blueTheme()
	default:
		return defaultTheme()
	}
}

// defaultTheme – estilo monocromático cinza, inspirado em Norton Commander / Turbo Vision.
func defaultTheme() Theme {
	return Theme{
		Name: "default",

		DesktopBg:   tcell.ColorNavy,
		DesktopFg:   tcell.ColorGray,
		DesktopChar: '░',

		MenuBarBg:    tcell.ColorSilver,
		MenuBarFg:    tcell.ColorBlack,
		MenuCursorBg: tcell.ColorGreen,
		MenuCursorFg: tcell.ColorBlack,

		StatusBarBg: tcell.ColorSilver,
		StatusBarFg: tcell.ColorBlack,

		EditorBg:       tcell.ColorBlack,
		EditorFg:       tcell.ColorSilver,
		EditorBorderFg: tcell.ColorWhite,

		PopupBg:         tcell.ColorSilver,
		PopupFg:         tcell.ColorBlack,
		PopupBorderFg:   tcell.ColorBlack,
		PopupSelectedBg: tcell.ColorBlack,
		PopupSelectedFg: tcell.ColorSilver,
		ShadowBg:        tcell.ColorGray,
	}
}

// blueTheme – estilo azul clássico tipo Norton Commander.
func blueTheme() Theme {
	return Theme{
		Name: "blue",

		DesktopBg:   tcell.ColorNavy,
		DesktopFg:   tcell.ColorBlue,
		DesktopChar: '░',

		MenuBarBg:    tcell.ColorTeal,
		MenuBarFg:    tcell.ColorWhite,
		MenuCursorBg: tcell.ColorGreen,
		MenuCursorFg: tcell.ColorBlack,

		StatusBarBg: tcell.ColorTeal,
		StatusBarFg: tcell.ColorWhite,

		EditorBg:       tcell.ColorNavy,
		EditorFg:       tcell.ColorWhite,
		EditorBorderFg: tcell.ColorWhite,

		PopupBg:         tcell.ColorSilver,
		PopupFg:         tcell.ColorBlack,
		PopupBorderFg:   tcell.ColorBlack,
		PopupSelectedBg: tcell.ColorNavy,
		PopupSelectedFg: tcell.ColorWhite,
		ShadowBg:        tcell.ColorBlack,
	}
}
