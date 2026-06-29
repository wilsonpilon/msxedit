package main

import (
	"fmt"
	"os"

	"msxedit/internal/reader"

	"github.com/spf13/cobra"
)

const Version = "4.1.5"

// BuildID é injetado em tempo de compilação pelo build.ps1
// via -ldflags "-X main.BuildID=<hex>". Valor "dev" indica build manual.
var BuildID = "dev"

func main() {
	fullVersion := fmt.Sprintf("%s (%s)", Version, BuildID)

	var (
		fileType string
		tabSize  int
		width    int
	)

	root := &cobra.Command{
		Use:           "msxread [opções] <arquivo>",
		Short:         "Visualizador de textos do MSXEdit (TXT, BAS tokenizado e MD)",
		Long: "msxread é o visualizador que faz par com o MSXEdit. Exibe arquivos de\n" +
			"texto (.txt), programas MSX-BASIC tokenizados (.bas binário, mostrados\n" +
			"de forma legível) e arquivos de ajuda em markdown (.md), no estilo do\n" +
			"leitor de README do Turbo Pascal.",
		Args:          cobra.ExactArgs(1),
		Version:       fullVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return reader.Run(reader.Options{
				FilePath: args[0],
				Type:     reader.DocType(fileType),
				TabSize:  tabSize,
				Width:    width,
			})
		},
	}

	root.Flags().StringVarP(&fileType, "type", "t", "auto",
		"Tipo do arquivo: auto, txt, bas, md")
	root.Flags().IntVar(&tabSize, "tabsize", 4,
		"Número de espaços por tabulação")
	root.Flags().IntVarP(&width, "width", "w", 0,
		"Largura máxima para quebra de linha em markdown (0 = sem limite)")
	root.SetVersionTemplate("msxread versão {{.Version}}\n")

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "msxread: %v\n", err)
		os.Exit(1)
	}
}
