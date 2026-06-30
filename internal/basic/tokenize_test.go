package basic

import (
	"strings"
	"testing"
)

func TestTokenizeRoundtrip(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			"print string",
			"10 PRINT \"HELLO\"\n",
		},
		{
			"for loop",
			"10 FOR I=1 TO 10\n20 NEXT I\n",
		},
		{
			"goto",
			"10 GOTO 10\n",
		},
		{
			"rem comment",
			"10 REM this is a comment\n",
		},
		{
			"integers",
			"10 A=5\n20 B=100\n30 C=1000\n",
		},
		{
			"gosub return",
			"10 GOSUB 100\n20 END\n100 PRINT \"SUB\"\n110 RETURN\n",
		},
		{
			"if then",
			"10 IF A=1 THEN 100\n100 PRINT \"YES\"\n",
		},
		{
			"data",
			"10 DATA 1,2,3,HELLO\n20 READ A\n",
		},
		{
			"hex octal",
			"10 A=&HFF\n20 B=&O77\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := Tokenize(tc.src)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}
			if len(tok) == 0 || tok[0] != MSXHeader {
				t.Fatal("missing 0xFF header")
			}

			text, err := DetokenizeToText(tok)
			if err != nil {
				t.Fatalf("Detokenize error: %v", err)
			}

			// Normaliza: remove espaços extras, uppercase
			normSrc := normalizeBasic(tc.src)
			normOut := normalizeBasic(text)
			if normSrc != normOut {
				t.Errorf("roundtrip mismatch\n  src: %q\n  got: %q", normSrc, normOut)
			}
		})
	}
}

// normalizeBasic uppercases e normaliza espaços para comparação.
func normalizeBasic(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.ToUpper(strings.TrimRight(l, " "))
	}
	return strings.Join(lines, "\n")
}
