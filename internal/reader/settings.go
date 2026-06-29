package reader

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Settings guarda as preferências persistentes do msxread.
type Settings struct {
	BodyFg   int  `json:"body_fg"`   // índice VGA 0-15 para cor do texto
	BodyBg   int  `json:"body_bg"`   // índice VGA 0-15 para cor do fundo
	WrapMode bool `json:"wrap_mode"` // quebra de linha ativa
	HiBit    bool `json:"hi_bit"`    // exibir caracteres acima de 127
}

func settingsPath() string {
	// Usa o diretório do executável: portátil e funciona em qualquer
	// contexto de terminal (incluindo MSYS/tmux sem USERPROFILE definido).
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "msxread.json")
	}
	return "msxread.json"
}

// LoadSettings carrega as configurações salvas; usa defaults se não existir.
func LoadSettings() Settings {
	s := Settings{BodyFg: 0, BodyBg: 3, HiBit: true} // padrão: preto sobre ciano, hi-bit ativo
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	if s.BodyFg < 0 || s.BodyFg > 15 {
		s.BodyFg = 0
	}
	if s.BodyBg < 0 || s.BodyBg > 15 {
		s.BodyBg = 3
	}
	return s
}

// SaveSettings grava as configurações em ~/.msxread.json.
func SaveSettings(s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(), data, 0o644)
}

// PrintLines grava as linhas em um arquivo temporário e envia para a impressora
// padrão do sistema operacional. Devolve o caminho do arquivo temporário gerado.
func PrintLines(fileName string, lines []string) (string, error) {
	tmp := filepath.Join(os.TempDir(), fileName+".print.txt")
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(tmp, []byte(sb.String()), 0o644); err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// notepad /p envia para a impressora padrão e fecha
		cmd = exec.Command("cmd", "/c", "notepad", "/p", tmp)
	default:
		// lpr é o padrão POSIX
		cmd = exec.Command("lpr", tmp)
	}
	return tmp, cmd.Start()
}
