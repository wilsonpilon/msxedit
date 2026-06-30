package basic

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// msxBase é o endereço de memória da primeira linha BASIC no MSX.
const msxBase = 0x8001

// jumpKeywords são instruções cujos números seguintes são referências de linha.
var jumpKeywords = map[string]bool{
	"RESTORE": true, "AUTO": true, "RENUM": true, "DELETE": true,
	"RESUME": true, "ERL": true, "ELSE": true, "RUN": true,
	"LIST": true, "LLIST": true, "GOTO": true, "RETURN": true,
	"THEN": true, "GOSUB": true,
}

// literalKeywords: após o token, o restante da linha é ASCII literal.
var literalKeywords = map[string]bool{
	"DATA": true, "REM": true, "'": true, "CALL": true, "_": true,
}

type kwEntry struct {
	upper string
	tok   []byte
}

var sortedKW []kwEntry

func init() {
	type pair struct{ kw, hex string }
	raw := []pair{
		{"END", "81"}, {"FOR", "82"}, {"NEXT", "83"}, {"DATA", "84"},
		{"INPUT", "85"}, {"DIM", "86"}, {"READ", "87"}, {"LET", "88"},
		{"GOTO", "89"}, {"RUN", "8a"}, {"IF", "8b"}, {"RESTORE", "8c"},
		{"GOSUB", "8d"}, {"RETURN", "8e"}, {"REM", "8f"}, {"STOP", "90"},
		{"PRINT", "91"}, {"?", "91"}, {"CLEAR", "92"}, {"LIST", "93"},
		{"NEW", "94"}, {"ON", "95"}, {"WAIT", "96"}, {"DEF", "97"},
		{"POKE", "98"}, {"CONT", "99"}, {"CSAVE", "9a"}, {"CLOAD", "9b"},
		{"OUT", "9c"}, {"LPRINT", "9d"}, {"LLIST", "9e"}, {"CLS", "9f"},
		{"WIDTH", "a0"}, {"TRON", "a2"}, {"TROFF", "a3"}, {"SWAP", "a4"},
		{"ERASE", "a5"}, {"ERROR", "a6"}, {"RESUME", "a7"}, {"DELETE", "a8"},
		{"AUTO", "a9"}, {"RENUM", "aa"}, {"DEFSTR", "ab"}, {"DEFINT", "ac"},
		{"DEFSNG", "ad"}, {"DEFDBL", "ae"}, {"LINE", "af"}, {"OPEN", "b0"},
		{"FIELD", "b1"}, {"GET", "b2"}, {"PUT", "b3"}, {"CLOSE", "b4"},
		{"LOAD", "b5"}, {"MERGE", "b6"}, {"FILES", "b7"}, {"LSET", "b8"},
		{"RSET", "b9"}, {"SAVE", "ba"}, {"LFILES", "bb"}, {"CIRCLE", "bc"},
		{"COLOR", "bd"}, {"DRAW", "be"}, {"PAINT", "bf"}, {"BEEP", "c0"},
		{"PLAY", "c1"}, {"PSET", "c2"}, {"PRESET", "c3"}, {"SOUND", "c4"},
		{"SCREEN", "c5"}, {"VPOKE", "c6"}, {"SPRITE", "c7"}, {"VDP", "c8"},
		{"BASE", "c9"}, {"CALL", "ca"}, {"_", "5f"}, {"TIME", "cb"},
		{"KEY", "cc"}, {"MAX", "cd"}, {"MOTOR", "ce"}, {"BLOAD", "cf"},
		{"BSAVE", "d0"}, {"DSKO$", "d1"}, {"SET", "d2"}, {"NAME", "d3"},
		{"KILL", "d4"}, {"IPL", "d5"}, {"COPY", "d6"}, {"CMD", "d7"},
		{"LOCATE", "d8"}, {"TO", "d9"}, {"THEN", "da"}, {"TAB(", "db"},
		{"STEP", "dc"}, {"USR", "dd"}, {"FN", "de"}, {"SPC(", "df"},
		{"NOT", "e0"}, {"ERL", "e1"}, {"ERR", "e2"}, {"STRING$", "e3"},
		{"USING", "e4"}, {"INSTR", "e5"}, {"VARPTR", "e7"}, {"CSRLIN", "e8"},
		{"ATTR$", "e9"}, {"DSKI$", "ea"}, {"OFF", "eb"}, {"INKEY$", "ec"},
		{"POINT", "ed"}, {">", "ee"}, {"=", "ef"}, {"<", "f0"},
		{"+", "f1"}, {"-", "f2"}, {"*", "f3"}, {"/", "f4"},
		{"^", "f5"}, {"AND", "f6"}, {"OR", "f7"}, {"XOR", "f8"},
		{"EQV", "f9"}, {"IMP", "fa"}, {"MOD", "fb"}, {"\\", "fc"},
		// Tokens estendidos (prefixo FF)
		{"PDL", "ffa4"}, {"EXP", "ff8b"}, {"PEEK", "ff97"}, {"FIX", "ffa1"},
		{"POS", "ff91"}, {"FPOS", "ffa7"}, {"ABS", "ff86"}, {"FRE", "ff8f"},
		{"ASC", "ff95"}, {"ATN", "ff8e"}, {"HEX$", "ff9b"}, {"BIN$", "ff9d"},
		{"INP", "ff90"}, {"RIGHT$", "ff82"}, {"RND", "ff88"}, {"INT", "ff85"},
		{"CDBL", "ffa0"}, {"CHR$", "ff96"}, {"CINT", "ff9e"}, {"LEFT$", "ff81"},
		{"SGN", "ff84"}, {"LEN", "ff92"}, {"SIN", "ff89"}, {"SPACE$", "ff99"},
		{"SQR", "ff87"}, {"LOC(", "ffac28"}, {"STICK", "ffa2"}, {"COS", "ff8c"},
		{"LOF", "ffad"}, {"STR$", "ff93"}, {"CSNG", "ff9f"}, {"LOG", "ff8a"},
		{"STRIG", "ffa3"}, {"LPOS", "ff9c"}, {"CVD", "ffaa"}, {"CVI", "ffa8"},
		{"CVS", "ffa9"}, {"TAN", "ff8d"}, {"MID$", "ff83"}, {"MKD$", "ffb0"},
		{"MKI$", "ffae"}, {"MKS$", "ffaf"}, {"VAL", "ff94"}, {"DSKF", "ffa6"},
		{"VPEEK", "ff98"}, {"OCT$", "ff9a"}, {"EOF", "ffab"}, {"PAD", "ffa5"},
		// Tokens especiais multi-byte
		{"'", "3a8fe6"}, {"ELSE", "3aa1"}, {"AS", "4153"},
	}

	for _, p := range raw {
		sortedKW = append(sortedKW, kwEntry{upper: p.kw, tok: hexDecodeStr(p.hex)})
	}
	// Longest-match first
	sort.SliceStable(sortedKW, func(i, j int) bool {
		return len(sortedKW[i].upper) > len(sortedKW[j].upper)
	})
}

func hexDecodeStr(s string) []byte {
	var out []byte
	for i := 0; i+1 < len(s); i += 2 {
		v, _ := strconv.ParseUint(s[i:i+2], 16, 8)
		out = append(out, byte(v))
	}
	return out
}

// Tokenize converte texto ASCII MSX BASIC para o formato tokenizado binário.
func Tokenize(text string) ([]byte, error) {
	out := []byte{MSXHeader}
	addr := uint16(msxBase)

	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	var lastNum uint16
	for i, raw := range rawLines {
		line := strings.TrimRight(raw, " \r\n")
		if line == "" {
			continue
		}
		if len(line) > 0 && line[0] == 26 {
			continue
		}

		// Extrai número da linha
		j := 0
		for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
			j++
		}
		if j >= len(line) || line[j] < '0' || line[j] > '9' {
			return nil, fmt.Errorf("linha %d: não começa com número", i+1)
		}
		start := j
		for j < len(line) && line[j] >= '0' && line[j] <= '9' {
			j++
		}
		lineNum64, _ := strconv.ParseUint(strings.TrimSpace(line[start:j]), 10, 32)
		lineNum := uint16(lineNum64)
		if lineNum > 65529 {
			return nil, fmt.Errorf("linha %d: número %d muito alto", i+1, lineNum)
		}
		if lastNum > 0 && lineNum <= lastNum {
			return nil, fmt.Errorf("linha %d: número %d fora de ordem", i+1, lineNum)
		}
		lastNum = lineNum

		// Pula espaço opcional após número
		if j < len(line) && line[j] == ' ' {
			j++
		}

		content, err := tokLine(line[j:])
		if err != nil {
			return nil, fmt.Errorf("linha %d: %w", i+1, err)
		}

		// Monta: [ptr_prox LE] [num_linha LE] [conteúdo] [0x00]
		lineSize := uint16(2 + 2 + len(content) + 1)
		nextAddr := addr + lineSize

		lineBuf := []byte{byte(nextAddr), byte(nextAddr >> 8), byte(lineNum), byte(lineNum >> 8)}
		lineBuf = append(lineBuf, content...)
		lineBuf = append(lineBuf, 0x00)
		out = append(out, lineBuf...)
		addr = nextAddr
	}

	out = append(out, 0x00, 0x00)
	return out, nil
}

func tokLine(line string) ([]byte, error) {
	var out []byte
	up := strings.ToUpper(line)
	pos := 0

	for pos < len(line) {
		// Tenta casar keyword (longest-match, case-insensitive)
		matched := false
		for _, kw := range sortedKW {
			if !strings.HasPrefix(up[pos:], kw.upper) {
				continue
			}
			// Evita casamento parcial de palavras (ex: "FORT" não é "FOR")
			endPos := pos + len(kw.upper)
			last := kw.upper[len(kw.upper)-1]
			if (last >= 'A' && last <= 'Z') && endPos < len(up) {
				nc := up[endPos]
				if (nc >= 'A' && nc <= 'Z') || (nc >= '0' && nc <= '9') {
					continue
				}
			}

			out = append(out, kw.tok...)
			pos += len(kw.upper)

			// Zona literal: restante da linha como ASCII
			if literalKeywords[kw.upper] {
				for pos < len(line) {
					out = append(out, line[pos])
					pos++
				}
			} else if jumpKeywords[kw.upper] {
				// Números seguintes viram referências de linha 0x0E + uint16 LE
				for pos < len(line) {
					// Consome espaços
					for pos < len(line) && line[pos] == ' ' {
						out = append(out, ' ')
						pos++
					}
					// Tenta número
					j := pos
					for j < len(line) && line[j] >= '0' && line[j] <= '9' {
						j++
					}
					if j > pos {
						n, _ := strconv.ParseUint(line[pos:j], 10, 16)
						out = append(out, 0x0E, byte(n), byte(n>>8))
						pos = j
					} else if pos < len(line) && line[pos] == ',' {
						out = append(out, ',')
						pos++
					} else {
						break
					}
				}
			}

			matched = true
			break
		}
		if matched {
			continue
		}

		c := line[pos]

		// String entre aspas
		if c == '"' {
			out = append(out, '"')
			pos++
			for pos < len(line) && line[pos] != '"' {
				out = append(out, line[pos])
				pos++
			}
			if pos < len(line) {
				out = append(out, '"')
				pos++
			}
			continue
		}

		// Bases numéricas &H, &O, &B
		if c == '&' && pos+1 < len(line) {
			base := up[pos+1]
			switch base {
			case 'H':
				pos += 2
				j := pos
				for j < len(line) && isHexDigit(up[j]) {
					j++
				}
				v := uint64(0)
				if j > pos {
					v, _ = strconv.ParseUint(line[pos:j], 16, 64)
				}
				out = append(out, 0x0C, byte(v), byte(v>>8))
				pos = j
				continue
			case 'O':
				pos += 2
				j := pos
				for j < len(line) && line[j] >= '0' && line[j] <= '7' {
					j++
				}
				v := uint64(0)
				if j > pos {
					v, _ = strconv.ParseUint(line[pos:j], 8, 64)
				}
				out = append(out, 0x0B, byte(v), byte(v>>8))
				pos = j
				continue
			case 'B':
				pos += 2
				out = append(out, 0x26, 0x42)
				for pos < len(line) && (line[pos] == '0' || line[pos] == '1') {
					out = append(out, line[pos])
					pos++
				}
				continue
			}
		}

		// Número
		cu := up[pos]
		if cu >= '0' && cu <= '9' || (c == '.' && pos+1 < len(line) && up[pos+1] >= '0' && up[pos+1] <= '9') {
			tok, n := parseNumber(line, up, pos)
			out = append(out, tok...)
			pos = n
			continue
		}

		// ASCII literal
		out = append(out, c)
		pos++
	}

	return out, nil
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

// parseNumber analisa um número a partir de pos e retorna os bytes tokenizados e a nova posição.
func parseNumber(line, up string, pos int) ([]byte, int) {
	j := pos
	hasDot := false

	if line[j] == '.' {
		hasDot = true
		j++
	}
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}
	if !hasDot && j < len(line) && line[j] == '.' {
		hasDot = true
		j++
		for j < len(line) && line[j] >= '0' && line[j] <= '9' {
			j++
		}
	}

	// Notação científica E/D
	hasExp := false
	isDouble := false
	if j < len(line) && (up[j] == 'E' || up[j] == 'D') {
		peek := j + 1
		if peek < len(line) && (line[peek] == '+' || line[peek] == '-') {
			isDouble = up[j] == 'D'
			hasExp = true
			j += 2
			for j < len(line) && line[j] >= '0' && line[j] <= '9' {
				j++
			}
		}
	}

	// Sufixo de tipo
	typeSuffix := byte(0)
	if j < len(line) {
		switch line[j] {
		case '%', '!', '#':
			typeSuffix = line[j]
			j++
		}
	}

	numStr := line[pos:j]
	if typeSuffix != 0 {
		numStr = line[pos : j-1]
	}

	// Float se tiver ponto decimal, exp ou sufixo ! ou #
	if hasDot || hasExp || typeSuffix == '!' || typeSuffix == '#' || isDouble {
		f, _ := strconv.ParseFloat(strings.TrimRight(numStr, "%!#"), 64)
		if isDouble || typeSuffix == '#' {
			return encMSXFloat(f, 0x1F, 14, 7), j
		}
		return encMSXFloat(f, 0x1D, 6, 3), j
	}

	// Inteiro
	n, _ := strconv.ParseInt(strings.TrimRight(numStr, "%!#"), 10, 64)
	switch {
	case n >= 0 && n <= 9:
		return []byte{byte(0x11 + n)}, j
	case n >= 10 && n <= 255:
		return []byte{0x0F, byte(n)}, j
	case n >= 256 && n <= 32767:
		return []byte{0x1C, byte(n), byte(n >> 8)}, j
	default:
		// Inteiro grande → single precision
		return encMSXFloat(float64(n), 0x1D, 6, 3), j
	}
}

// encMSXFloat codifica um float no formato BCD do MSX-BASIC.
// header = 0x1D (single, mantBytes=3, precision=6) ou 0x1F (double, mantBytes=7, precision=14).
func encMSXFloat(f float64, header byte, precision, mantBytes int) []byte {
	absF := math.Abs(f)

	// Obtém string do número com dígitos significativos suficientes
	fmtStr := fmt.Sprintf("%.*f", precision+2, absF)
	dotIdx := strings.Index(fmtStr, ".")
	var intPart, fracPart string
	if dotIdx < 0 {
		intPart = fmtStr
	} else {
		intPart = fmtStr[:dotIdx]
		fracPart = fmtStr[dotIdx+1:]
	}

	// Calcula expoente: quantos dígitos significativos há antes do ponto
	stripped := strings.TrimLeft(intPart, "0")
	var expByte byte
	if stripped == "" {
		// Número puramente fracionário: conta zeros iniciais após o ponto
		leadZ := len(fracPart) - len(strings.TrimLeft(fracPart, "0"))
		if leadZ >= 63 {
			leadZ = 63
		}
		expByte = byte(64 - leadZ)
	} else {
		e := len(stripped)
		if e > 63 {
			e = 63
		}
		expByte = byte(e + 64)
	}

	// Junta todos os dígitos (sem ponto), remove zeros à esquerda, completa à direita
	allDigits := strings.TrimLeft(intPart+fracPart, "0")
	if allDigits == "" {
		allDigits = "0"
	}
	for len(allDigits) < precision {
		allDigits += "0"
	}
	allDigits = allDigits[:precision]

	// Monta bytes: cada par de dígitos decimais → 1 byte (nibble alto, nibble baixo)
	out := []byte{header, expByte}
	for i := 0; i < mantBytes; i++ {
		var hi, lo byte
		if i*2 < len(allDigits) {
			hi = allDigits[i*2] - '0'
		}
		if i*2+1 < len(allDigits) {
			lo = allDigits[i*2+1] - '0'
		}
		out = append(out, (hi<<4)|lo)
	}
	return out
}
