package tui

import "github.com/gdamore/tcell/v2"

var (
	vgaBlack      = tcell.NewRGBColor(0, 0, 0)
	vgaBlue       = tcell.NewRGBColor(0, 0, 170)
	vgaGreen      = tcell.NewRGBColor(0, 170, 0)
	vgaCyan       = tcell.NewRGBColor(0, 170, 170)
	vgaRed        = tcell.NewRGBColor(170, 0, 0)
	vgaMagenta    = tcell.NewRGBColor(170, 0, 170)
	vgaBrown      = tcell.NewRGBColor(170, 85, 0)
	vgaLightGray  = tcell.NewRGBColor(170, 170, 170)
	vgaDarkGray   = tcell.NewRGBColor(85, 85, 85)
	vgaLightBlue  = tcell.NewRGBColor(85, 85, 255)
	vgaLightGreen = tcell.NewRGBColor(85, 255, 85)
	vgaLightCyan  = tcell.NewRGBColor(85, 255, 255)
	vgaLightRed   = tcell.NewRGBColor(255, 85, 85)
	vgaLightMag   = tcell.NewRGBColor(255, 85, 255)
	vgaYellow     = tcell.NewRGBColor(255, 255, 85)
	vgaWhite      = tcell.NewRGBColor(255, 255, 255)
)

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

	// Janela de Help
	HelpBg        tcell.Color
	HelpFg        tcell.Color
	HelpBorderFg  tcell.Color
	HelpLinkFg    tcell.Color
	HelpSelectedFg tcell.Color
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

// defaultTheme - VGA Borland blue (MS-DOS/Turbo style) como tema padrao.
func defaultTheme() Theme {
	return Theme{
		Name: "default",

		DesktopBg:   vgaBlue,
		DesktopFg:   vgaLightBlue,
		DesktopChar: '░',

		MenuBarBg:    vgaLightGray,
		MenuBarFg:    vgaBlack,
		MenuCursorBg: vgaGreen,
		MenuCursorFg: vgaBlack,

		StatusBarBg: vgaLightGray,
		StatusBarFg: vgaBlack,

		EditorBg:       vgaBlue,
		EditorFg:       vgaWhite,
		EditorBorderFg: vgaWhite,

		PopupBg:         vgaLightGray,
		PopupFg:         vgaBlack,
		PopupBorderFg:   vgaBlack,
		PopupSelectedBg: vgaBlack,
		PopupSelectedFg: vgaLightGray,
		ShadowBg:        vgaDarkGray,

		HelpBg:         vgaCyan,
		HelpFg:         vgaBlack,
		HelpBorderFg:   vgaWhite,
		HelpLinkFg:     vgaYellow,
		HelpSelectedFg: vgaRed,
	}
}

// blueTheme - VGA NC-style (Norton Commander), com barra superior e status em ciano.
func blueTheme() Theme {
	return Theme{
		Name: "blue",

		DesktopBg:   vgaBlue,
		DesktopFg:   vgaLightBlue,
		DesktopChar: '░',

		MenuBarBg:    vgaCyan,
		MenuBarFg:    vgaWhite,
		MenuCursorBg: vgaGreen,
		MenuCursorFg: vgaBlack,

		StatusBarBg: vgaCyan,
		StatusBarFg: vgaWhite,

		EditorBg:       vgaBlue,
		EditorFg:       vgaWhite,
		EditorBorderFg: vgaWhite,

		PopupBg:         vgaLightGray,
		PopupFg:         vgaBlack,
		PopupBorderFg:   vgaBlack,
		PopupSelectedBg: vgaBlue,
		PopupSelectedFg: vgaWhite,
		ShadowBg:        vgaBlack,

		HelpBg:         vgaCyan,
		HelpFg:         vgaBlack,
		HelpBorderFg:   vgaWhite,
		HelpLinkFg:     vgaYellow,
		HelpSelectedFg: vgaRed,
	}
}
