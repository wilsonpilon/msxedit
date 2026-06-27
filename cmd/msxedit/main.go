package main

import (
	"fmt"
	"log"
	"msxedit/internal/cli"
	"msxedit/internal/config"
	"msxedit/internal/tui"
	"time"
)

const Version = "4.0.3"

func main() {
	// Gerar Build ID (Hexadecimal do Unix UTC)
	buildID := fmt.Sprintf("%X", time.Now().UTC().Unix())
	fullVersion := fmt.Sprintf("%s (%s)", Version, buildID)

	// 1. Interpretar opções de linha de comando
	opts := cli.Parse(fullVersion)

	// 2. Determinar caminho da configuração e carregar
	configPath, err := config.GetConfigPath(opts.LocalConfig)
	if err != nil {
		log.Fatalf("Erro ao determinar caminho da configuração: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("Aviso: Erro ao carregar configuração: %v. Usando padrões.", err)
	}

	// 3. Sobrescrever configurações com opções de CLI (se fornecidas)
	if opts.Theme != "" {
		cfg.Theme = opts.Theme
	}
	if opts.TabSize > 0 {
		cfg.TabSize = opts.TabSize
	}
	if opts.NoHighlight {
		cfg.Highlight = false
	}

	// 4. Iniciar Interface TUI
	app := tui.NewApp(Version, buildID, cfg.Theme)
	if err := app.Run(opts.FilePath); err != nil {
		log.Fatalf("Erro ao iniciar interface: %v", err)
	}
}
