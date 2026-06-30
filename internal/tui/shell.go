package tui

import (
	"os"
	"os/exec"
	"path/filepath"
)

// showPowerShell suspende o TUI, abre um terminal PowerShell no diretório do
// arquivo ativo e retorna ao editor quando o usuário digitar "exit".
func (a *App) showPowerShell() {
	dir := ""
	if len(a.Editors) > 0 && a.ActiveEditor >= 0 && a.ActiveEditor < len(a.Editors) {
		ed := a.Editors[a.ActiveEditor]
		if ed.filePath != "" {
			dir = filepath.Dir(ed.filePath)
		}
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}

	a.Application.Suspend(func() {
		shell := "pwsh.exe"
		if _, err := exec.LookPath(shell); err != nil {
			shell = "powershell.exe"
		}
		cmd := exec.Command(shell, "-NoLogo")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		cmd.Run() //nolint:errcheck — saída normal via exit não é erro
	})
}
