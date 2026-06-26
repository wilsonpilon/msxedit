package cli

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	FilePath    string
	LocalConfig bool
	Theme       string
	TabSize     int
	NoHighlight bool
	Version     bool
}

func Parse(version string) Options {
	opts := Options{}

	flag.StringVar(&opts.Theme, "theme", "", "Define o tema de cores")
	flag.IntVar(&opts.TabSize, "tabsize", 0, "Define o tamanho do tab")
	flag.BoolVar(&opts.NoHighlight, "no-highlight", false, "Desativa syntax highlighting")
	flag.BoolVar(&opts.LocalConfig, "local", false, "Usa arquivo de configuração local")
	flag.BoolVar(&opts.Version, "version", false, "Exibe a versão do programa")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MSXEdit versão %s\n", version)
		fmt.Fprintf(os.Stderr, "Uso: %s [opções] [arquivo]\n", os.Args[0])
		fmt.Println("\nOpções:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if opts.Version {
		fmt.Printf("MSXEdit versão %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		opts.FilePath = flag.Arg(0)
	}

	return opts
}
