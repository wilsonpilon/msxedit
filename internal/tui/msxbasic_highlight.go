package tui

import (
	"strings"
	"unicode"
)

// basicTokenKind classifica cada span de texto para syntax highlighting.
type basicTokenKind int

const (
	basicKindDefault    basicTokenKind = iota
	basicKindLineNumber                // número de linha inicial: 10, 20, 100
	basicKindStatement                 // comandos: PRINT, GOTO, IF, FOR, SCREEN...
	basicKindModifier                  // palavras estruturais: TO, THEN, ELSE, STEP, AS...
	basicKindFunction                  // funções built-in: ABS(), LEFT$(), PEEK(), RND...
	basicKindString                    // literais string: "HELLO WORLD"
	basicKindNumber                    // literais numéricos: 123, 3.14, &HFFFF, &B1010
	basicKindComment                   // REM ... ou ' ...
	basicKindVariable                  // variáveis do usuário: A, COUNT$, X1, FLAG%
	basicKindOperator                  // operadores: AND, OR, NOT, +, -, =, <, >, ^, \
	basicKindSymbol                    // símbolos: (, ), ,, ;, :, #, @
)

// basicSpan representa um trecho colorido em uma linha.
type basicSpan struct {
	col  int
	ln   int // comprimento em runes
	kind basicTokenKind
}

// msxKeywordMap mapeia keywords MSX-BASIC (uppercase) para seu tipo de token.
// Baseado na tabela de tokens de TOKEN.md e do basic-dignified.
var msxKeywordMap = map[string]basicTokenKind{
	// ── Statements / Comandos ──────────────────────────────────────────────
	"AUTO":    basicKindStatement,
	"BASE":    basicKindStatement,
	"BEEP":    basicKindStatement,
	"BLOAD":   basicKindStatement,
	"BSAVE":   basicKindStatement,
	"CALL":    basicKindStatement,
	"CIRCLE":  basicKindStatement,
	"CLEAR":   basicKindStatement,
	"CLOAD":   basicKindStatement,
	"CLOSE":   basicKindStatement,
	"CLS":     basicKindStatement,
	"CMD":     basicKindStatement,
	"COLOR":   basicKindStatement,
	"CONT":    basicKindStatement,
	"COPY":    basicKindStatement,
	"CSAVE":   basicKindStatement,
	"DATA":    basicKindStatement,
	"DEFDBL":  basicKindStatement,
	"DEFINT":  basicKindStatement,
	"DEFSNG":  basicKindStatement,
	"DEFSTR":  basicKindStatement,
	"DEF":     basicKindStatement,
	"DELETE":  basicKindStatement,
	"DIM":     basicKindStatement,
	"DRAW":    basicKindStatement,
	"DSKO$":   basicKindStatement,
	"END":     basicKindStatement,
	"ERASE":   basicKindStatement,
	"ERROR":   basicKindStatement,
	"FIELD":   basicKindStatement,
	"FILES":   basicKindStatement,
	"FOR":     basicKindStatement,
	"GET":     basicKindStatement,
	"GOSUB":   basicKindStatement,
	"GOTO":    basicKindStatement,
	"IF":      basicKindStatement,
	"INPUT":   basicKindStatement,
	"IPL":     basicKindStatement,
	"KEY":     basicKindStatement,
	"KILL":    basicKindStatement,
	"LET":     basicKindStatement,
	"LFILES":  basicKindStatement,
	"LINE":    basicKindStatement,
	"LIST":    basicKindStatement,
	"LLIST":   basicKindStatement,
	"LOAD":    basicKindStatement,
	"LOCATE":  basicKindStatement,
	"LPRINT":  basicKindStatement,
	"LSET":    basicKindStatement,
	"MAX":     basicKindStatement,
	"MERGE":   basicKindStatement,
	"MOTOR":   basicKindStatement,
	"NAME":    basicKindStatement,
	"NEW":     basicKindStatement,
	"NEXT":    basicKindStatement,
	"ON":      basicKindStatement,
	"OPEN":    basicKindStatement,
	"OUT":     basicKindStatement,
	"PAINT":   basicKindStatement,
	"PLAY":    basicKindStatement,
	"POKE":    basicKindStatement,
	"PRESET":  basicKindStatement,
	"PRINT":   basicKindStatement,
	"PSET":    basicKindStatement,
	"PUT":     basicKindStatement,
	"READ":    basicKindStatement,
	"REM":     basicKindStatement,
	"RENUM":   basicKindStatement,
	"RESTORE": basicKindStatement,
	"RESUME":  basicKindStatement,
	"RETURN":  basicKindStatement,
	"RSET":    basicKindStatement,
	"RUN":     basicKindStatement,
	"SAVE":    basicKindStatement,
	"SCREEN":  basicKindStatement,
	"SET":     basicKindStatement,
	"SOUND":   basicKindStatement,
	"SPRITE":  basicKindStatement,
	"STOP":    basicKindStatement,
	"SWAP":    basicKindStatement,
	"TIME":    basicKindStatement,
	"TRON":    basicKindStatement,
	"TROFF":   basicKindStatement,
	"VDP":     basicKindStatement,
	"VPOKE":   basicKindStatement,
	"WAIT":    basicKindStatement,
	"WIDTH":   basicKindStatement,

	// ── Modificadores / Palavras estruturais ──────────────────────────────
	"AS":    basicKindModifier,
	"ELSE":  basicKindModifier,
	"FN":    basicKindModifier,
	"OFF":   basicKindModifier,
	"SPC":   basicKindModifier,
	"STEP":  basicKindModifier,
	"TAB":   basicKindModifier,
	"THEN":  basicKindModifier,
	"TO":    basicKindModifier,
	"USR":   basicKindModifier,
	"USING": basicKindModifier,

	// ── Operadores por palavra ─────────────────────────────────────────────
	"AND": basicKindOperator,
	"EQV": basicKindOperator,
	"IMP": basicKindOperator,
	"MOD": basicKindOperator,
	"NOT": basicKindOperator,
	"OR":  basicKindOperator,
	"XOR": basicKindOperator,

	// ── Funções built-in (tabela estendida 0xFF + byte) ───────────────────
	"ABS":     basicKindFunction,
	"ASC":     basicKindFunction,
	"ATN":     basicKindFunction,
	"ATTR$":   basicKindFunction,
	"BIN$":    basicKindFunction,
	"CDBL":    basicKindFunction,
	"CHR$":    basicKindFunction,
	"CINT":    basicKindFunction,
	"COS":     basicKindFunction,
	"CSNG":    basicKindFunction,
	"CSRLIN":  basicKindFunction,
	"CVD":     basicKindFunction,
	"CVI":     basicKindFunction,
	"CVS":     basicKindFunction,
	"DSKI$":   basicKindFunction,
	"DSKF":    basicKindFunction,
	"EOF":     basicKindFunction,
	"ERL":     basicKindFunction,
	"ERR":     basicKindFunction,
	"EXP":     basicKindFunction,
	"FIX":     basicKindFunction,
	"FPOS":    basicKindFunction,
	"FRE":     basicKindFunction,
	"HEX$":    basicKindFunction,
	"INKEY$":  basicKindFunction,
	"INP":     basicKindFunction,
	"INPUT$":  basicKindFunction,
	"INSTR":   basicKindFunction,
	"INT":     basicKindFunction,
	"LEFT$":   basicKindFunction,
	"LEN":     basicKindFunction,
	"LOC":     basicKindFunction,
	"LOF":     basicKindFunction,
	"LOG":     basicKindFunction,
	"LPOS":    basicKindFunction,
	"MID$":    basicKindFunction,
	"MKD$":    basicKindFunction,
	"MKI$":    basicKindFunction,
	"MKS$":    basicKindFunction,
	"OCT$":    basicKindFunction,
	"PAD":     basicKindFunction,
	"PDL":     basicKindFunction,
	"PEEK":    basicKindFunction,
	"POINT":   basicKindFunction,
	"POS":     basicKindFunction,
	"RIGHT$":  basicKindFunction,
	"RND":     basicKindFunction,
	"SGN":     basicKindFunction,
	"SIN":     basicKindFunction,
	"SPACE$":  basicKindFunction,
	"SQR":     basicKindFunction,
	"STR$":    basicKindFunction,
	"STICK":   basicKindFunction,
	"STRING$": basicKindFunction,
	"STRIG":   basicKindFunction,
	"TAN":     basicKindFunction,
	"VAL":     basicKindFunction,
	"VARPTR":  basicKindFunction,
	"VPEEK":   basicKindFunction,
}

// isOperatorChar retorna true para caracteres que formam operadores simbólicos.
func isOperatorChar(r rune) bool {
	switch r {
	case '+', '-', '*', '/', '^', '=', '<', '>', '\\':
		return true
	}
	return false
}

// isSymbolChar retorna true para símbolos de pontuação BASIC.
func isSymbolChar(r rune) bool {
	switch r {
	case '(', ')', ',', ';', ':', '#', '@', '?', '!', '%':
		return true
	}
	return false
}

// classifyWord retorna o tipo de token para um identificador (já em uppercase).
func classifyWord(word string) basicTokenKind {
	if kind, ok := msxKeywordMap[word]; ok {
		return kind
	}
	return basicKindVariable
}

// tokenizeBasicLine tokeniza uma linha de MSX-BASIC e retorna spans coloridos.
// Todos os caracteres da linha são cobertos por algum span.
func tokenizeBasicLine(line string) []basicSpan {
	spans := make([]basicSpan, 0, 16)
	runes := []rune(line)
	n := len(runes)
	i := 0

	// ── Número de linha inicial ──────────────────────────────────────────
	if i < n && unicode.IsDigit(runes[i]) {
		start := i
		for i < n && unicode.IsDigit(runes[i]) {
			i++
		}
		spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindLineNumber})
	}

	// ── Corpo da linha ───────────────────────────────────────────────────
	inComment := false
	inData := false

	for i < n {
		if inComment {
			spans = append(spans, basicSpan{col: i, ln: n - i, kind: basicKindComment})
			break
		}

		if inData {
			// Após DATA: tudo é literal (default), exceto strings que mantêm sua cor
			start := i
			for i < n && runes[i] != ':' {
				if runes[i] == '"' {
					// Emitir o que acumulou antes da string
					if i > start {
						spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindDefault})
					}
					// Emitir a string
					strStart := i
					i++
					for i < n && runes[i] != '"' {
						i++
					}
					if i < n {
						i++ // fecha aspas
					}
					spans = append(spans, basicSpan{col: strStart, ln: i - strStart, kind: basicKindString})
					start = i
					continue
				}
				i++
			}
			if i > start {
				spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindDefault})
			}
			// Se encontrou ':', DATA termina — volta ao modo normal
			if i < n && runes[i] == ':' {
				inData = false
			}
			continue
		}

		r := runes[i]

		// ── Comentário com apóstrofo ─────────────────────────────────────
		if r == '\'' {
			spans = append(spans, basicSpan{col: i, ln: n - i, kind: basicKindComment})
			break
		}

		// ── String literal ───────────────────────────────────────────────
		if r == '"' {
			start := i
			i++
			for i < n && runes[i] != '"' {
				i++
			}
			if i < n {
				i++ // fecha aspas
			}
			spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindString})
			continue
		}

		// ── Números em bases especiais: &H, &O, &B ───────────────────────
		if r == '&' && i+1 < n {
			next := unicode.ToUpper(runes[i+1])
			if next == 'H' || next == 'O' || next == 'B' {
				start := i
				i += 2
				for i < n && isHexDigit(runes[i]) {
					i++
				}
				spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindNumber})
				continue
			}
		}

		// ── Número decimal ───────────────────────────────────────────────
		if unicode.IsDigit(r) || (r == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			hasE := false
			for i < n {
				c := runes[i]
				if unicode.IsDigit(c) || c == '.' {
					i++
					continue
				}
				if (c == 'E' || c == 'e') && !hasE {
					hasE = true
					i++
					if i < n && (runes[i] == '+' || runes[i] == '-') {
						i++
					}
					continue
				}
				break
			}
			// sufixos de tipo: !, #, %
			if i < n && (runes[i] == '!' || runes[i] == '#' || runes[i] == '%') {
				i++
			}
			spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindNumber})
			continue
		}

		// ── Identificador / Keyword ──────────────────────────────────────
		if unicode.IsLetter(r) {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
				i++
			}
			// Sufixo de tipo de variável ou nome de função: $, %, !, #
			if i < n && (runes[i] == '$' || runes[i] == '%' || runes[i] == '!' || runes[i] == '#') {
				i++
			}
			word := strings.ToUpper(string(runes[start:i]))
			kind := classifyWord(word)
			spans = append(spans, basicSpan{col: start, ln: i - start, kind: kind})

			switch word {
			case "REM":
				inComment = true
			case "DATA":
				inData = true
			}
			continue
		}

		// ── Operadores simbólicos ────────────────────────────────────────
		if isOperatorChar(r) {
			start := i
			for i < n && isOperatorChar(runes[i]) {
				i++
			}
			spans = append(spans, basicSpan{col: start, ln: i - start, kind: basicKindOperator})
			continue
		}

		// ── Símbolos de pontuação ────────────────────────────────────────
		if isSymbolChar(r) {
			spans = append(spans, basicSpan{col: i, ln: 1, kind: basicKindSymbol})
			i++
			continue
		}

		// ── Default (espaços, outros) ────────────────────────────────────
		spans = append(spans, basicSpan{col: i, ln: 1, kind: basicKindDefault})
		i++
	}

	return spans
}

func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) ||
		(r >= 'A' && r <= 'F') ||
		(r >= 'a' && r <= 'f') ||
		(r >= '0' && r <= '9')
}
