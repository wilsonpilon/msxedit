package reader

import (
	"sort"
	"strings"
)

// SpanKind classifica o papel de um trecho de texto na linha BASIC.
type SpanKind int

const (
	SpanKeyword SpanKind = iota + 1
	SpanString
	SpanComment
)

// Span marca o intervalo [Start, End) de runas de uma linha com um SpanKind.
type Span struct {
	Start, End int
	Kind       SpanKind
}

// basicKeywordList contém todas as palavras-chave do MSX-BASIC, ordenadas do
// mais longo para o mais curto para garantir correspondência gananciosa.
var basicKeywordList []string

func init() {
	raw := []string{
		// Comandos simpleTokens
		"END", "FOR", "NEXT", "DATA", "INPUT", "DIM", "READ", "LET",
		"GOTO", "RUN", "IF", "RESTORE", "GOSUB", "RETURN", "REM", "STOP",
		"PRINT", "CLEAR", "LIST", "NEW", "ON", "WAIT", "DEF", "POKE",
		"CONT", "CSAVE", "CLOAD", "OUT", "LPRINT", "LLIST", "CLS", "WIDTH",
		"TRON", "TROFF", "SWAP", "ERASE", "ERROR", "RESUME", "DELETE", "AUTO",
		"RENUM", "DEFSTR", "DEFINT", "DEFSNG", "DEFDBL", "LINE", "OPEN", "FIELD",
		"GET", "PUT", "CLOSE", "LOAD", "MERGE", "FILES", "LSET", "RSET",
		"SAVE", "LFILES", "CIRCLE", "COLOR", "DRAW", "PAINT", "BEEP", "PLAY",
		"PSET", "PRESET", "SOUND", "SCREEN", "VPOKE", "SPRITE", "VDP", "BASE",
		"CALL", "TIME", "KEY", "MAX", "MOTOR", "BLOAD", "BSAVE", "DSKO$",
		"SET", "NAME", "KILL", "IPL", "COPY", "CMD", "LOCATE",
		// Palavras auxiliares e operadores textuais
		"TO", "THEN", "TAB(", "STEP", "USR", "FN", "SPC(", "NOT",
		"ERL", "ERR", "STRING$", "USING", "INSTR", "VARPTR", "CSRLIN", "ATTR$",
		"DSKI$", "OFF", "INKEY$", "POINT", "AND", "OR", "XOR", "EQV", "IMP",
		"MOD", "ELSE",
		// Funções extendedTokens
		"LEFT$", "RIGHT$", "MID$", "SGN", "INT", "ABS", "SQR", "RND",
		"SIN", "LOG", "EXP", "COS", "TAN", "ATN", "FRE", "INP",
		"POS", "LEN", "STR$", "VAL", "ASC", "CHR$", "PEEK", "VPEEK",
		"SPACE$", "OCT$", "HEX$", "LPOS", "BIN$", "CINT", "CSNG", "CDBL",
		"FIX", "STICK", "STRIG", "PDL", "PAD", "DSKF", "FPOS", "CVI",
		"CVS", "CVD", "EOF", "LOC", "LOF", "MKI$", "MKS$", "MKD$",
	}
	sort.Slice(raw, func(i, j int) bool { return len(raw[i]) > len(raw[j]) })
	basicKeywordList = raw
}

// buildBasicSpans percorre as linhas detokenizadas e devolve os spans de cada uma.
func buildBasicSpans(lines []string) [][]Span {
	out := make([][]Span, len(lines))
	for i, l := range lines {
		out[i] = scanBasicLine(l)
	}
	return out
}

// scanBasicLine classifica os trechos de uma única linha BASIC detokenizada.
func scanBasicLine(line string) []Span {
	runes := []rune(line)
	n := len(runes)
	var spans []Span
	i := 0

	// Número de linha: dígitos iniciais (sem span especial — usa estilo padrão).
	for i < n && runes[i] >= '0' && runes[i] <= '9' {
		i++
	}
	if i < n && runes[i] == ' ' {
		i++
	}

	for i < n {
		switch runes[i] {
		case '"':
			start := i
			i++
			for i < n && runes[i] != '"' {
				i++
			}
			if i < n {
				i++ // fecha aspas
			}
			spans = append(spans, Span{Start: start, End: i, Kind: SpanString})

		case '\'':
			spans = append(spans, Span{Start: i, End: n, Kind: SpanComment})
			return spans

		default:
			if kw, end := matchBasicKeyword(runes, i); kw != "" {
				spans = append(spans, Span{Start: i, End: end, Kind: SpanKeyword})
				if kw == "REM" && end < n {
					spans = append(spans, Span{Start: end, End: n, Kind: SpanComment})
					return spans
				}
				i = end
			} else {
				i++
			}
		}
	}
	return spans
}

// matchBasicKeyword tenta reconhecer uma palavra-chave em runes[pos:].
// Devolve a palavra-chave e a posição logo após ela, ou ("", pos) se não houver.
func matchBasicKeyword(runes []rune, pos int) (string, int) {
	n := len(runes)
	upper := strings.ToUpper(string(runes[pos:]))
	for _, kw := range basicKeywordList {
		kwLen := len(kw) // todos são ASCII
		if n-pos < kwLen {
			continue
		}
		if !strings.HasPrefix(upper, kw) {
			continue
		}
		end := pos + kwLen
		// Verificação de fronteira de palavra para palavras-chave que não
		// terminam em '$' ou '(' (esses tipos já incluem o delimitador).
		if !strings.HasSuffix(kw, "$") && !strings.HasSuffix(kw, "(") {
			if end < n {
				next := runes[end]
				if isBasicAlphaNum(next) {
					continue
				}
			}
		}
		return kw, end
	}
	return "", pos
}

func isBasicAlphaNum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') || r == '_' || r == '$' || r == '%' ||
		r == '!' || r == '#'
}
