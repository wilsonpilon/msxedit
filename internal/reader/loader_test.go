package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return p
}

func TestLoadDocumentText(t *testing.T) {
	p := writeTemp(t, "nota.txt", []byte("linha 1\nlinha 2\n"))
	doc, err := LoadDocument(p, TypeAuto, 4, 0)
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if doc.Type != TypeText {
		t.Errorf("tipo = %s, esperado txt", doc.Type)
	}
	if len(doc.Lines) != 2 || doc.Lines[0] != "linha 1" {
		t.Errorf("linhas inesperadas: %#v", doc.Lines)
	}
	if doc.FileName != "nota.txt" {
		t.Errorf("FileName = %q", doc.FileName)
	}
}

func TestLoadDocumentMarkdown(t *testing.T) {
	md := "# Titulo\n\nVeja [o manual](MANUAL.md) aqui.\n- item\n"
	p := writeTemp(t, "ajuda.md", []byte(md))
	doc, err := LoadDocument(p, TypeAuto, 4, 0)
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if doc.Type != TypeMD {
		t.Errorf("tipo = %s, esperado md", doc.Type)
	}
	if doc.Lines[0] != "Titulo" || !doc.Headings[0] {
		t.Errorf("título não detectado: %#v / headings=%v", doc.Lines[0], doc.Headings)
	}
	// link reduzido ao texto
	if want := "Veja o manual aqui."; doc.Lines[2] != want {
		t.Errorf("linha de link = %q, esperado %q", doc.Lines[2], want)
	}
	if doc.Lines[3] != "• item" {
		t.Errorf("item de lista = %q", doc.Lines[3])
	}
}

func TestLoadDocumentTokenizedBas(t *testing.T) {
	// Programa tokenizado: 10 PRINT "HI"
	data := []byte{
		0xFF,             // cabeçalho
		0x01, 0x80,       // ponteiro fictício
		0x0A, 0x00,       // linha 10
		0x91, 0x22, 'H', 'I', 0x22, // PRINT "HI"
		0x00,             // fim da linha
		0x00, 0x00,       // fim do programa
	}
	p := writeTemp(t, "prog.bas", data)
	doc, err := LoadDocument(p, TypeAuto, 4, 0)
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if doc.Type != TypeBas {
		t.Fatalf("tipo = %s, esperado bas", doc.Type)
	}
	if doc.Lines[0] != `10 PRINT"HI"` {
		t.Errorf("linha detokenizada = %q", doc.Lines[0])
	}
}

func TestExpandTabs(t *testing.T) {
	if got := expandTabs("a\tb", 4); got != "a   b" {
		t.Errorf("expandTabs = %q", got)
	}
}
