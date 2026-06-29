package reader

import "github.com/gdamore/tcell/v2"

// Paleta VGA clássica (mesma referência usada em internal/tui/theme.go),
// replicada aqui para manter o pacote reader autônomo.
var (
	vgaBlack     = tcell.NewRGBColor(0, 0, 0)
	vgaCyan      = tcell.NewRGBColor(0, 170, 170)
	vgaDarkGray  = tcell.NewRGBColor(85, 85, 85)
	vgaLightGray = tcell.NewRGBColor(170, 170, 170)
	vgaWhite     = tcell.NewRGBColor(255, 255, 255)
	vgaYellow    = tcell.NewRGBColor(255, 255, 85)
)

// Theme reúne as cores do visualizador. Conforme especificação:
// topo cinza com letras pretas, corpo cyan com letras pretas, status cinza.
type Theme struct {
	TopBg    tcell.Color
	TopFg    tcell.Color
	BodyBg   tcell.Color
	BodyFg   tcell.Color
	StatusBg tcell.Color
	StatusFg tcell.Color

	HeadingFg tcell.Color // títulos markdown
	KeywordFg tcell.Color // palavras-chave BASIC
	StringFg  tcell.Color // literais string BASIC
	CommentFg tcell.Color // comentários REM / '
}

// DefaultTheme devolve o tema padrão do msxread.
func DefaultTheme() Theme {
	return Theme{
		TopBg:     vgaLightGray,
		TopFg:     vgaBlack,
		BodyBg:    vgaCyan,
		BodyFg:    vgaBlack,
		StatusBg:  vgaLightGray,
		StatusFg:  vgaBlack,
		HeadingFg: vgaWhite,
		KeywordFg: vgaYellow,
		StringFg:  vgaWhite,
		CommentFg: vgaDarkGray,
	}
}

// sepGlyph é o losango pequeno usado como separador na barra de topo.
const sepGlyph = '◆'

// vgaPalette contém as 16 cores VGA clássicas, indexadas de 0 a 15.
// Usadas pelas teclas F5/F6 (texto) e F7/F8 (fundo).
var vgaPalette = [16]tcell.Color{
	tcell.NewRGBColor(0, 0, 0),       // 0  Preto
	tcell.NewRGBColor(0, 0, 170),     // 1  Azul escuro
	tcell.NewRGBColor(0, 170, 0),     // 2  Verde escuro
	tcell.NewRGBColor(0, 170, 170),   // 3  Ciano escuro
	tcell.NewRGBColor(170, 0, 0),     // 4  Vermelho escuro
	tcell.NewRGBColor(170, 0, 170),   // 5  Magenta escuro
	tcell.NewRGBColor(170, 85, 0),    // 6  Marrom
	tcell.NewRGBColor(170, 170, 170), // 7  Cinza claro
	tcell.NewRGBColor(85, 85, 85),    // 8  Cinza escuro
	tcell.NewRGBColor(85, 85, 255),   // 9  Azul brilhante
	tcell.NewRGBColor(85, 255, 85),   // 10 Verde brilhante
	tcell.NewRGBColor(85, 255, 255),  // 11 Ciano brilhante
	tcell.NewRGBColor(255, 85, 85),   // 12 Vermelho brilhante
	tcell.NewRGBColor(255, 85, 255),  // 13 Magenta brilhante
	tcell.NewRGBColor(255, 255, 85),  // 14 Amarelo
	tcell.NewRGBColor(255, 255, 255), // 15 Branco
}

var vgaColorNames = [16]string{
	"Preto", "Azul", "Verde", "Ciano",
	"Vermelho", "Magenta", "Marrom", "Cinza Cl.",
	"Cinza Esc.", "Azul Cl.", "Verde Cl.", "Ciano Cl.",
	"Verm. Cl.", "Magenta Cl.", "Amarelo", "Branco",
}

func vgaColorName(i int) string {
	if i >= 0 && i < 16 {
		return vgaColorNames[i]
	}
	return "?"
}
