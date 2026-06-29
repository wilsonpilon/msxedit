package reader

import "testing"

func TestScanBasicLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		wants []Span
	}{
		{
			name:  "keyword simples",
			line:  "10 PRINT",
			wants: []Span{{Start: 3, End: 8, Kind: SpanKeyword}},
		},
		{
			name: "keyword + string",
			line: `30 PRINT"Hello"`,
			wants: []Span{
				{Start: 3, End: 8, Kind: SpanKeyword},
				{Start: 8, End: 15, Kind: SpanString},
			},
		},
		{
			name: "REM marca resto como comentário",
			line: "10 REM Este é um comentário",
			wants: []Span{
				{Start: 3, End: 6, Kind: SpanKeyword},
				{Start: 6, End: len([]rune("10 REM Este é um comentário")), Kind: SpanComment},
			},
		},
		{
			name: "apóstrofo como comentário",
			line: "20 A=1 'nota",
			wants: []Span{
				{Start: 7, End: 12, Kind: SpanComment},
			},
		},
		{
			name: "FOR TO NEXT",
			line: "20 FOR I=1 TO 10",
			wants: []Span{
				{Start: 3, End: 6, Kind: SpanKeyword},
				{Start: 11, End: 13, Kind: SpanKeyword},
			},
		},
		{
			name: "keyword não confunde variável",
			line: "10 FORA=1",
			// FORA não é keyword (FOR seguido de A, não é fronteira)
			wants: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanBasicLine(tt.line)
			if len(got) != len(tt.wants) {
				t.Fatalf("spans = %v, quer %v", got, tt.wants)
			}
			for i, sp := range got {
				w := tt.wants[i]
				if sp.Start != w.Start || sp.End != w.End || sp.Kind != w.Kind {
					t.Errorf("span[%d] = %+v, quer %+v", i, sp, w)
				}
			}
		})
	}
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		line  string
		width int
		want  []string
	}{
		{
			line:  "Linha curta",
			width: 40,
			want:  []string{"Linha curta"},
		},
		{
			line:  "Um dois tres quatro cinco seis sete oito",
			width: 20,
			want:  []string{"Um dois tres quatro", "cinco seis sete oito"},
		},
		{
			line:  "  • item com texto longo que precisa de quebra aqui",
			width: 30,
			want:  []string{"  • item com texto longo que", "  precisa de quebra aqui"},
		},
	}

	for _, tt := range tests {
		got := wordWrap(tt.line, tt.width)
		if len(got) != len(tt.want) {
			t.Errorf("wordWrap(%q, %d) = %v, quer %v", tt.line, tt.width, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("wordWrap linha[%d] = %q, quer %q", i, got[i], tt.want[i])
			}
		}
	}
}
