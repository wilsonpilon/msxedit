package basic

import "testing"

// buildProgram monta um programa tokenizado completo (cabeçalho 0xFF, linhas com
// ponteiro fictício + número + conteúdo + 0x00, e marcador final 0x00 0x00).
func buildProgram(lines ...testLine) []byte {
	out := []byte{MSXHeader}
	for _, l := range lines {
		out = append(out, 0x01, 0x80) // ponteiro de próxima linha (fictício, != 0)
		out = append(out, byte(l.num), byte(l.num>>8))
		out = append(out, l.content...)
		out = append(out, 0x00)
	}
	out = append(out, 0x00, 0x00) // fim do programa (próximo ponteiro == 0)
	return out
}

type testLine struct {
	num     uint16
	content []byte
}

func TestDetokenizeExamples(t *testing.T) {
	cases := []struct {
		name string
		line testLine
		want string
	}{
		{
			// Exemplo 1 — PRINT "HELLO" (TOKEN.md §9)
			name:    "print string",
			line:    testLine{10, []byte{0x91, 0x22, 'H', 'E', 'L', 'L', 'O', 0x22}},
			want:    `PRINT"HELLO"`,
		},
		{
			// Exemplo 2 — GOTO com referência de linha 100 (0x0E 64 00)
			name: "goto line ref",
			line: testLine{50, []byte{0x89, 0x0E, 0x64, 0x00}},
			want: "GOTO100",
		},
		{
			// Exemplo 3 — REM com texto literal
			name: "rem literal",
			line: testLine{30, []byte{0x8F, 'O', 'l', 'a', ' ', 'm', 'u', 'n', 'd', 'o'}},
			want: "REMOla mundo",
		},
		{
			// Exemplo 4 — LET A=10 (0x0F 0x0A = byte 10), com espaço literal
			name: "let assign byte",
			line: testLine{100, []byte{0x88, ' ', 'A', 0xEF, 0x0F, 0x0A}},
			want: "LET A=10",
		},
		{
			// Exemplo 5 — comentário curto com apóstrofo (3A 8F E6)
			name: "apostrophe comment",
			line: testLine{70, []byte{0x3A, 0x8F, 0xE6, 'O', 'i'}},
			want: "'Oi",
		},
		{
			// Hexadecimal &H100 (0x0C 00 01)
			name: "hex literal",
			line: testLine{10, []byte{0x91, ' ', 0x0C, 0x00, 0x01}},
			want: "PRINT &H100",
		},
		{
			// Inteiro word 1000 (0x1C E8 03)
			name: "word integer",
			line: testLine{10, []byte{0x88, ' ', 'A', 0xEF, 0x1C, 0xE8, 0x03}},
			want: "LET A=1000",
		},
		{
			// Octal &O777 (0x0B FF 01)
			name: "octal literal",
			line: testLine{10, []byte{0x91, ' ', 0x0B, 0xFF, 0x01}},
			want: "PRINT &O777",
		},
		{
			// Dígito imediato 5 (0x16)
			name: "immediate digit",
			line: testLine{10, []byte{0x88, ' ', 'A', 0xEF, 0x16}},
			want: "LET A=5",
		},
		{
			// Token estendido PEEK (FF 97) + parêntese literal e dígito 0 (0x11)
			name: "extended function",
			line: testLine{10, []byte{0x91, ' ', 0xFF, 0x97, '(', 0x11, ')'}},
			want: "PRINT PEEK(0)",
		},
		{
			// ELSE multi-byte (3A A1)
			name: "else token",
			line: testLine{10, []byte{0x8B, ' ', 'A', ' ', 0xDA, ' ', 'B', ' ', 0x3A, 0xA1, ' ', 'C'}},
			want: "IF A THEN B ELSE C",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := buildProgram(tc.line)
			lines, err := Detokenize(data)
			if err != nil {
				t.Fatalf("Detokenize: %v", err)
			}
			if len(lines) != 1 {
				t.Fatalf("esperava 1 linha, obtive %d", len(lines))
			}
			if lines[0].Number != tc.line.num {
				t.Errorf("número da linha = %d, esperado %d", lines[0].Number, tc.line.num)
			}
			if lines[0].Content != tc.want {
				t.Errorf("conteúdo = %q, esperado %q", lines[0].Content, tc.want)
			}
		})
	}
}

func TestDetokenizeToText(t *testing.T) {
	data := buildProgram(
		testLine{10, []byte{0x91, 0x22, 'H', 'I', 0x22}},
		testLine{20, []byte{0x81}}, // END
	)
	got, err := DetokenizeToText(data)
	if err != nil {
		t.Fatalf("DetokenizeToText: %v", err)
	}
	want := "10 PRINT\"HI\"\n20 END\n"
	if got != want {
		t.Errorf("texto = %q, esperado %q", got, want)
	}
}

func TestIsTokenized(t *testing.T) {
	if !IsTokenized([]byte{0xFF, 0x00}) {
		t.Error("0xFF deveria ser reconhecido como tokenizado")
	}
	if IsTokenized([]byte("10 PRINT")) {
		t.Error("texto ASCII não deveria ser reconhecido como tokenizado")
	}
	if IsTokenized(nil) {
		t.Error("vazio não deveria ser tokenizado")
	}
}
