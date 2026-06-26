package basic

import (
	"errors"
	"io"
)

// Tokens comuns do MSX-BASIC (exemplos iniciais de 0x80 em diante)
var Tokens = map[byte]string{
	0x81: "END",
	0x82: "FOR",
	0x83: "NEXT",
	0x84: "DATA",
	0x85: "INPUT",
	0x86: "DIM",
	0x87: "READ",
	0x88: "LET",
	0x89: "GOTO",
	0x8A: "RUN",
	0x8B: "IF",
	0x8C: "RESTORE",
	0x8D: "GOSUB",
	0x8E: "RETURN",
	0x8F: "REM",
	0x90: "STOP",
	0x91: "PRINT",
	0x92: "CLEAR",
	0x93: "LIST",
	0x94: "NEW",
	0x95: "ON",
	0x96: "WAIT",
	0x97: "DEF",
	0x98: "POKE",
	0x99: "CONT",
	// ... mais tokens seriam adicionados aqui
}

// MSXHeader é o byte inicial de arquivos BASIC tokenizados (0xFF)
const MSXHeader = 0xFF

// Line representa uma linha do BASIC
type Line struct {
	Number  uint16
	Content string
}

// Tokenizer para converter ASCII -> Tokens e vice-versa
type Processor struct {
	// Poderíamos ter tabelas de lookup otimizadas aqui
}

func NewProcessor() *Processor {
	return &Processor{}
}

// LoadTokenized lê um arquivo .BAS tokenizado
func (p *Processor) LoadTokenized(r io.Reader) ([]Line, error) {
	// Implementação futura do parser de formato binário do MSX
	// [Header 0xFF]
	// [Ptr Prox Linha (2 bytes)][Num Linha (2 bytes)][Tokens/Chars...][0x00]
	// Final do arquivo: [Ptr Prox Linha == 0x0000]
	return nil, errors.New("não implementado: carregamento de binário BASIC")
}
