package basic

import (
	"errors"
	"strconv"
	"strings"
)

// IsTokenized informa se os dados representam um programa BASIC tokenizado,
// identificado pelo byte de cabeçalho 0xFF.
func IsTokenized(data []byte) bool {
	return len(data) > 0 && data[0] == MSXHeader
}

// Detokenize converte um programa BASIC tokenizado (.BAS binário) na lista de
// linhas em ASCII. O algoritmo segue TOKEN.md seção 8.
func Detokenize(data []byte) ([]Line, error) {
	if !IsTokenized(data) {
		return nil, errors.New("basic: cabeçalho 0xFF ausente (arquivo não tokenizado)")
	}

	var lines []Line
	pos := 1 // pula o cabeçalho 0xFF

	for {
		// Ponteiro da próxima linha (2 bytes LE). 0x0000 marca o fim do programa.
		if pos+1 >= len(data) {
			break
		}
		nextAddr := uint16(data[pos]) | uint16(data[pos+1])<<8
		pos += 2
		if nextAddr == 0 {
			break
		}

		// Número da linha (2 bytes LE).
		if pos+1 >= len(data) {
			break
		}
		lineNum := uint16(data[pos]) | uint16(data[pos+1])<<8
		pos += 2

		content, next := decodeLine(data, pos)
		pos = next
		lines = append(lines, Line{Number: lineNum, Content: content})
	}

	return lines, nil
}

// DetokenizeToText devolve o programa detokenizado como texto pronto para exibição.
func DetokenizeToText(data []byte) (string, error) {
	lines, err := Detokenize(data)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(strconv.Itoa(int(l.Number)))
		b.WriteByte(' ')
		b.WriteString(l.Content)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// decodeLine decodifica o conteúdo de uma única linha a partir de pos, parando
// no terminador 0x00. Devolve o texto e a posição logo após o terminador.
func decodeLine(data []byte, pos int) (string, int) {
	var b strings.Builder

	emitLiteralToEOL := func() {
		for pos < len(data) && data[pos] != 0x00 {
			b.WriteByte(data[pos])
			pos++
		}
	}

	for pos < len(data) {
		c := data[pos]

		switch {
		case c == 0x00:
			pos++ // consome terminador
			return b.String(), pos

		case c == 0xFF: // token estendido
			pos++
			if pos < len(data) {
				if kw, ok := extendedTokens[data[pos]]; ok {
					b.WriteString(kw)
				}
				pos++
			}

		case c == 0x3A: // ':' — pode iniciar ELSE ou comentário "'"
			if pos+1 < len(data) && data[pos+1] == 0xA1 {
				b.WriteString("ELSE")
				pos += 2
			} else if pos+2 < len(data) && data[pos+1] == 0x8F && data[pos+2] == 0xE6 {
				b.WriteByte('\'')
				pos += 3
				emitLiteralToEOL()
			} else {
				b.WriteByte(':')
				pos++
			}

		case c == 0x8F: // REM — restante da linha é literal
			b.WriteString("REM")
			pos++
			emitLiteralToEOL()

		case c == 0x84: // DATA — restante da linha é literal
			b.WriteString("DATA")
			pos++
			emitLiteralToEOL()

		case c == 0xCA: // CALL — restante da linha é literal (assembly inline)
			b.WriteString("CALL")
			pos++
			emitLiteralToEOL()

		case c == 0x22: // string entre aspas
			b.WriteByte('"')
			pos++
			for pos < len(data) && data[pos] != 0x22 && data[pos] != 0x00 {
				b.WriteByte(data[pos])
				pos++
			}
			if pos < len(data) && data[pos] == 0x22 {
				b.WriteByte('"')
				pos++
			}

		case c == 0x0B: // octal &O
			v, n := readWord(data, pos+1)
			b.WriteString("&O")
			b.WriteString(strconv.FormatUint(uint64(v), 8))
			pos = n

		case c == 0x0C: // hexadecimal &H
			v, n := readWord(data, pos+1)
			b.WriteString("&H")
			b.WriteString(strings.ToUpper(strconv.FormatUint(uint64(v), 16)))
			pos = n

		case c == 0x0D || c == 0x0E: // referência de número de linha
			v, n := readWord(data, pos+1)
			b.WriteString(strconv.Itoa(int(v)))
			pos = n

		case c == 0x0F: // inteiro byte (10–255)
			if pos+1 < len(data) {
				b.WriteString(strconv.Itoa(int(data[pos+1])))
				pos += 2
			} else {
				pos++
			}

		case c >= 0x11 && c <= 0x1A: // dígito imediato 0–9
			b.WriteByte('0' + (c - 0x11))
			pos++

		case c == 0x1C: // inteiro word (256–32767)
			v, n := readWord(data, pos+1)
			b.WriteString(strconv.Itoa(int(v)))
			pos = n

		case c == 0x1D: // float single (exp + 3 mantissa)
			s, n := decodeFloat(data, pos+1, 3)
			b.WriteString(s)
			pos = n

		case c == 0x1F: // float double (exp + 7 mantissa)
			s, n := decodeFloat(data, pos+1, 7)
			b.WriteString(s)
			pos = n

		case c >= 0x81: // token simples
			if kw, ok := simpleTokens[c]; ok {
				b.WriteString(kw)
			}
			pos++

		default: // ASCII literal (letras, espaço, símbolos, &B...)
			b.WriteByte(c)
			pos++
		}
	}

	return b.String(), pos
}

// readWord lê um uint16 little-endian em pos e devolve o valor e a posição seguinte.
func readWord(data []byte, pos int) (uint16, int) {
	if pos+1 >= len(data) {
		if pos < len(data) {
			return uint16(data[pos]), pos + 1
		}
		return 0, pos
	}
	return uint16(data[pos]) | uint16(data[pos+1])<<8, pos + 2
}

// decodeFloat decodifica um número de ponto flutuante BCD do MSX-BASIC.
// mantBytes = 3 (single) ou 7 (double). Decodificação best-effort conforme
// TOKEN.md seção 5.8/5.9 (1 byte de expoente + mantissa BCD).
func decodeFloat(data []byte, pos, mantBytes int) (string, int) {
	end := pos + 1 + mantBytes
	if end > len(data) {
		end = len(data)
	}
	if pos >= len(data) {
		return "0", pos
	}

	exp := int(data[pos]&0x7f) - 0x40 // dígitos antes do ponto decimal
	var digits strings.Builder
	for i := pos + 1; i < end; i++ {
		digits.WriteByte('0' + (data[i] >> 4))
		digits.WriteByte('0' + (data[i] & 0x0f))
	}
	ds := digits.String()

	var out string
	switch {
	case exp <= 0:
		out = "0." + strings.Repeat("0", -exp) + ds
	case exp >= len(ds):
		out = ds + strings.Repeat("0", exp-len(ds))
	default:
		out = ds[:exp] + "." + ds[exp:]
	}

	if strings.Contains(out, ".") {
		out = strings.TrimRight(out, "0")
		out = strings.TrimRight(out, ".")
	}
	if out == "" {
		out = "0"
	}
	return out, end
}
