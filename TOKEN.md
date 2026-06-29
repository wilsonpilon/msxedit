# TOKEN.md — Referência do Tokenizador MSX-BASIC

Este documento descreve o formato binário dos arquivos `.BAS` tokenizados do MSX-BASIC, servindo de referência para a implementação do tokenizador/detokenizador em Go no MSXEdit.

Fonte: análise do projeto **basic-dignified** (`msx/msxbatoken/msxbatoken.py`, `documentation/BATOKEN.md`).

---

## 1. Estrutura do Arquivo Binário

Um arquivo `.BAS` tokenizado começa com um byte de cabeçalho `0xFF` no endereço de memória `0x8000`, seguido por uma sequência de linhas e terminado por dois bytes nulos.

```
Offset 0x8000: FF                        ← cabeçalho obrigatório
Offset 0x8001: [linha 1]
               [linha 2]
               ...
               00 00                     ← marcador de fim de programa
```

### Estrutura de cada linha

```
[2 bytes LE] ptr_próxima_linha
[2 bytes LE] número_da_linha
[N bytes   ] conteúdo tokenizado
[1 byte    ] 00  ← terminador de linha
```

- Todos os ponteiros e números de linha usam **little-endian** (`uint16`).
- O ponteiro da próxima linha aponta para o endereço absoluto da linha seguinte na memória do MSX.
- A última linha aponta para `0x0000` (junto com o marcador de fim do programa).

### Cálculo do ponteiro de próxima linha

```
endereço_atual + 2 (ptr) + 2 (num_linha) + len(conteúdo_tokenizado) + 1 (terminador)
```

---

## 2. Tabela de Tokens Simples (1 byte)

Bytes de `0x80` a `0xFC` representam palavras-chave e operadores.

### Comandos (0x81–0xD8)

| Byte | Keyword   | Byte | Keyword   | Byte | Keyword   | Byte | Keyword   |
|------|-----------|------|-----------|------|-----------|------|-----------|
| 0x81 | END       | 0x82 | FOR       | 0x83 | NEXT      | 0x84 | DATA      |
| 0x85 | INPUT     | 0x86 | DIM       | 0x87 | READ      | 0x88 | LET       |
| 0x89 | GOTO      | 0x8A | RUN       | 0x8B | IF        | 0x8C | RESTORE   |
| 0x8D | GOSUB     | 0x8E | RETURN    | 0x8F | REM       | 0x90 | STOP      |
| 0x91 | PRINT     | 0x92 | CLEAR     | 0x93 | LIST      | 0x94 | NEW       |
| 0x95 | ON        | 0x96 | WAIT      | 0x97 | DEF       | 0x98 | POKE      |
| 0x99 | CONT      | 0x9A | CSAVE     | 0x9B | CLOAD     | 0x9C | OUT       |
| 0x9D | LPRINT    | 0x9E | LLIST     | 0x9F | CLS       | 0xA0 | WIDTH     |
| 0xA2 | TRON      | 0xA3 | TROFF     | 0xA4 | SWAP      | 0xA5 | ERASE     |
| 0xA6 | ERROR     | 0xA7 | RESUME    | 0xA8 | DELETE    | 0xA9 | AUTO      |
| 0xAA | RENUM     | 0xAB | DEFSTR    | 0xAC | DEFINT    | 0xAD | DEFSNG    |
| 0xAE | DEFDBL    | 0xAF | LINE      | 0xB0 | OPEN      | 0xB1 | FIELD     |
| 0xB2 | GET       | 0xB3 | PUT       | 0xB4 | CLOSE     | 0xB5 | LOAD      |
| 0xB6 | MERGE     | 0xB7 | FILES     | 0xB8 | LSET      | 0xB9 | RSET      |
| 0xBA | SAVE      | 0xBB | LFILES    | 0xBC | CIRCLE    | 0xBD | COLOR     |
| 0xBE | DRAW      | 0xBF | PAINT     | 0xC0 | BEEP      | 0xC1 | PLAY      |
| 0xC2 | PSET      | 0xC3 | PRESET    | 0xC4 | SOUND     | 0xC5 | SCREEN    |
| 0xC6 | VPOKE     | 0xC7 | SPRITE    | 0xC8 | VDP       | 0xC9 | BASE      |
| 0xCA | CALL      | 0xCB | TIME      | 0xCC | KEY       | 0xCD | MAX       |
| 0xCE | MOTOR     | 0xCF | BLOAD     | 0xD0 | BSAVE     | 0xD1 | DSKO$     |
| 0xD2 | SET       | 0xD3 | NAME      | 0xD4 | KILL      | 0xD5 | IPL       |
| 0xD6 | COPY      | 0xD7 | CMD       | 0xD8 | LOCATE    |       |           |

### Palavras auxiliares e operadores (0xD9–0xFC)

| Byte | Keyword   | Byte | Keyword   | Byte | Keyword   | Byte | Keyword   |
|------|-----------|------|-----------|------|-----------|------|-----------|
| 0xD9 | TO        | 0xDA | THEN      | 0xDB | TAB(      | 0xDC | STEP      |
| 0xDD | USR       | 0xDE | FN        | 0xDF | SPC(      | 0xE0 | NOT       |
| 0xE1 | ERL       | 0xE2 | ERR       | 0xE3 | STRING$   | 0xE4 | USING     |
| 0xE5 | INSTR     | 0xE7 | VARPTR    | 0xE8 | CSRLIN    | 0xE9 | ATTR$     |
| 0xEA | DSKI$     | 0xEB | OFF       | 0xEC | INKEY$    | 0xED | POINT     |
| 0xEE | >         | 0xEF | =         | 0xF0 | <         | 0xF1 | +         |
| 0xF2 | -         | 0xF3 | *         | 0xF4 | /         | 0xF5 | ^         |
| 0xF6 | AND       | 0xF7 | OR        | 0xF8 | XOR       | 0xF9 | EQV       |
| 0xFA | IMP       | 0xFB | MOD       | 0xFC | \         |       |           |

> **Nota:** `0xA1` é usado como parte do token multi-byte `ELSE`. `0xE6` é parte de `'` (comentário). Veja seção 4.

---

## 3. Tabela de Tokens Estendidos (prefixo `0xFF`)

Quando o byte `0xFF` é encontrado, o byte seguinte completa o token (funções e variáveis do sistema).

| Bytes    | Keyword  | Bytes    | Keyword  | Bytes    | Keyword  |
|----------|----------|----------|----------|----------|----------|
| FF 81    | LEFT$    | FF 82    | RIGHT$   | FF 83    | MID$     |
| FF 84    | SGN      | FF 85    | INT      | FF 86    | ABS      |
| FF 87    | SQR      | FF 88    | RND      | FF 89    | SIN      |
| FF 8A    | LOG      | FF 8B    | EXP      | FF 8C    | COS      |
| FF 8D    | TAN      | FF 8E    | ATN      | FF 8F    | FRE      |
| FF 90    | INP      | FF 91    | POS      | FF 92    | LEN      |
| FF 93    | STR$     | FF 94    | VAL      | FF 95    | ASC      |
| FF 96    | CHR$     | FF 97    | PEEK     | FF 98    | VPEEK    |
| FF 99    | SPACE$   | FF 9A    | OCT$     | FF 9B    | HEX$     |
| FF 9C    | LPOS     | FF 9D    | BIN$     | FF 9E    | CINT     |
| FF 9F    | CSNG     | FF A0    | CDBL     | FF A1    | FIX      |
| FF A2    | STICK    | FF A3    | STRIG    | FF A4    | PDL      |
| FF A5    | PAD      | FF A6    | DSKF     | FF A7    | FPOS     |
| FF A8    | CVI      | FF A9    | CVS      | FF AA    | CVD      |
| FF AB    | EOF      | FF AC    | LOC(     | FF AD    | LOF      |
| FF AE    | MKI$     | FF AF    | MKS$     | FF B0    | MKD$     |

> **Token especial de 3 bytes:** `LOC(` = `FF AC 28` (o `0x28` é o parêntese `(` em ASCII).

---

## 4. Tokens Multi-byte Especiais

Alguns tokens não se encaixam nas tabelas acima e requerem sequências específicas.

| Token ASCII | Sequência binária | Descrição |
|-------------|-------------------|-----------|
| `'`         | `3A 8F E6`        | Comentário curto (equivale a `: REM`) |
| `ELSE`      | `3A A1`           | Sempre precedido de `:` |
| `AS`        | `41 53`           | Letras ASCII 'A' e 'S' literais |

---

## 5. Encoding de Números

### 5.1 Dígitos imediatos (0 a 9)

Armazenados como um único byte sem prefixo:

```
valor = byte - 0x11
```

| Valor | Byte |
|-------|------|
| 0     | 0x11 |
| 1     | 0x12 |
| 5     | 0x16 |
| 9     | 0x1A |

### 5.2 Inteiros byte (10 a 255)

```
[0x0F] [valor: 1 byte]
```

Exemplo: `255` → `0F FF`

### 5.3 Inteiros word (256 a 32767)

```
[0x1C] [low byte] [high byte]   (little-endian)
```

Exemplo: `1000` → `1C E8 03`

### 5.4 Referência de número de linha (GOTO, GOSUB, THEN, etc.)

```
[0x0E] [low byte] [high byte]   (little-endian)
```

Usado sempre que um número de linha aparece como destino de desvio.

Exemplo: `GOTO 100` → `89 0E 64 00`

### 5.5 Hexadecimal (&Hxxxx)

```
[0x0C] [low byte] [high byte]   (little-endian)
```

Exemplo: `&H100` → `0C 00 01`

### 5.6 Octal (&Oxxxx)

```
[0x0B] [low byte] [high byte]   (little-endian)
```

Exemplo: `&O777` → `0B FF 01`

### 5.7 Binário (&Bxxxxxxxx)

```
[0x26] [0x42] [chars ASCII '0' e '1'...]
```

Exemplo: `&B1010` → `26 42 31 30 31 30`

### 5.8 Float single-precision

```
[0x1D] [expoente: 1 byte] [mantissa: 3 bytes]   = 5 bytes total
```

- Expoente: `len(dígitos_significativos) + 64`
- Mantissa: 6 dígitos decimais significativos empacotados em 3 bytes

### 5.9 Float double-precision

```
[0x1F] [expoente: 1 byte] [mantissa: 7 bytes]   = 9 bytes total
```

- Expoente: mesmo esquema do single
- Mantissa: 14 dígitos decimais significativos empacotados em 7 bytes

---

## 6. Strings e Zonas Literais

### 6.1 Strings entre aspas

Todo o conteúdo entre aspas duplas é armazenado como bytes ASCII literais, incluindo as próprias aspas:

```
[0x22] [bytes ASCII...] [0x22]
```

Exemplo: `"HELLO"` → `22 48 45 4C 4C 4F 22`

### 6.2 REM e comentários

Após o token `0x8F` (REM) ou a sequência `3A 8F E6` (`'`), **todos os bytes até o terminador `0x00`** são ASCII literal — nenhum processamento de tokens é feito.

### 6.3 DATA

Após o token `0x84` (DATA), **todos os bytes até o terminador `0x00`** são ASCII literal — incluindo vírgulas separadoras.

### 6.4 CALL

Após o token `0xCA` (CALL), **todos os bytes até o terminador `0x00`** são ASCII literal (permite sintaxe de assembly inline).

---

## 7. Algoritmo de Tokenização (ASCII → Binário)

```
1. Ler arquivo ASCII linha por linha
2. Para cada linha não vazia:
   a. Extrair número de linha inicial (uint16)
   b. Escanear o restante da linha:
      - Tentar casar keyword (case-insensitive) na tabela de tokens
        → se token = REM ou DATA ou CALL: emitir token + restante como ASCII literal
      - Tentar casar número:
        → 0–9 imediato: byte 0x11+valor
        → 10–255: 0x0F + byte
        → 256–32767: 0x1C + uint16 LE
        → float: 0x1D ou 0x1F + mantissa
        → &H: 0x0C + uint16 LE
        → &O: 0x0B + uint16 LE
        → &B: 0x26 0x42 + chars ASCII
      - Se o contexto exige número de linha (GOTO/GOSUB/THEN/RESTORE/ON/etc.):
        → emitir 0x0E + uint16 LE
      - String: emitir 0x22 + bytes ASCII + 0x22
      - Outros caracteres: ASCII literal
   c. Appendar terminador 0x00
   d. Calcular endereço da próxima linha
   e. Montar header: [ptr_prox LE uint16] [num_linha LE uint16] + conteúdo + 0x00
3. Appendar marcador de fim: 0x00 0x00
4. Prefixar arquivo com 0xFF
```

---

## 8. Algoritmo de Detokenização (Binário → ASCII)

```
1. Verificar byte 0xFF no início (offset 0x8000)
2. Loop:
   a. Ler ptr_prox (uint16 LE)
   b. Se ptr_prox == 0x0000 → fim do programa
   c. Ler num_linha (uint16 LE)
   d. Emitir: strconv.Itoa(num_linha) + " "
   e. Ler bytes até 0x00:
      byte < 0x80  → emitir como caractere ASCII
      0x80–0xFE    → lookup na tabela simples → emitir keyword
      0xFF         → ler próximo byte → lookup na tabela estendida → emitir keyword
      0x0E         → ler uint16 LE → emitir número de linha
      0x0F         → ler 1 byte → emitir inteiro (10–255)
      0x11–0x1A    → emitir dígito (byte - 0x11)
      0x1C         → ler uint16 LE → emitir inteiro (256–32767)
      0x1D         → ler 4 bytes → decodificar float single → emitir
      0x1F         → ler 8 bytes → decodificar float double → emitir
      0x0B         → ler uint16 LE → emitir "&O" + fmt octal
      0x0C         → ler uint16 LE → emitir "&H" + fmt hex
      0x0D         → ler uint16 LE → emitir número de linha (THEN/ELSE interno)
      0x22 (")     → copiar bytes até próximo 0x22 → emitir string literal
   f. Emitir newline
3. Fim
```

---

## 9. Exemplos Práticos

### Exemplo 1 — PRINT com string

```
ASCII:   10 PRINT "HELLO"
Hex:     XX XX  0A 00  91  22 48 45 4C 4C 4F 22  00
         ↑ ptr  ↑ ln#  ↑   ← string "HELLO"  →  ↑
                      PRINT                      end
```

### Exemplo 2 — GOTO com número de linha

```
ASCII:   50 GOTO 100
Hex:     XX XX  32 00  89  0E 64 00  00
         ↑ ptr  ↑ ln#  ↑   ↑ ↑  →   ↑
                      GOTO ref 100   end
```

### Exemplo 3 — REM (comentário)

```
ASCII:   30 REM Ola mundo
Hex:     XX XX  1E 00  8F  4F 6C 61 20 6D 75 6E 64 6F  00
         ↑ ptr  ↑ ln#  ↑   ← ASCII literal          →  ↑
                      REM                               end
```

### Exemplo 4 — Cálculo com inteiros

```
ASCII:   100 LET A=10
Hex:     XX XX  64 00  88  20 41  EF  0F 0A  00
         ↑ ptr  ↑ ln#  ↑   ↑  ↑   ↑   ↑ ↑   ↑
                      LET sp  A   =  byte 10  end
```

### Exemplo 5 — Comentário curto com apóstrofo

```
ASCII:   70 ' Isso e um comentario
Hex:     XX XX  46 00  3A 8F E6  49 73 73 6F ...  00
         ↑ ptr  ↑ ln#  ↑ ↑  ↑   ← ASCII literal → ↑
                      ' (token 3 bytes)             end
```

---

## 10. Notas de Implementação em Go

### Estruturas recomendadas

```go
// Tokens simples: byte → string (detokenização)
var SimpleTokens = map[byte]string{
    0x81: "END", 0x82: "FOR", 0x83: "NEXT", /* ... */
}

// Tokens simples: string → byte (tokenização), case-insensitive
var SimpleTokensReverse = map[string]byte{
    "END": 0x81, "FOR": 0x82, "NEXT": 0x83, /* ... */
}

// Tokens estendidos: byte → string (lidos após 0xFF)
var ExtendedTokens = map[byte]string{
    0x81: "LEFT$", 0x82: "RIGHT$", /* ... */
}

// Estrutura de uma linha tokenizada
type TokenizedLine struct {
    NextAddr   uint16
    LineNumber uint16
    Content    []byte
}
```

### Pontos críticos

1. **Endianness:** usar `binary.LittleEndian` em todas as leituras/escritas de `uint16`.
2. **Zonas literais:** ao tokenizar, após REM / DATA / CALL, parar o scan de tokens e copiar ASCII até o fim da linha.
3. **Busca de keywords:** ao tokenizar, tentar casar a keyword mais longa primeiro (ex.: `STRING$` antes de `STR$`).
4. **Números de linha em desvios:** ao tokenizar `GOTO 100`, o `100` vira `0x0E 0x64 0x00`, não um número comum.
5. **Atualização de ponteiros:** calcular `NextAddr` somente depois de serializar o conteúdo da linha.
6. **Strings:** o byte `0x22` (`"`) abre e fecha zonas literais — nenhum token é processado dentro.
7. **Token `'`:** a sequência `3A 8F E6` é um token de 3 bytes e deve ser reconhecida como unidade na detokenização.

---

## 11. Referências

- `E:/msxedit/basic-dignified/msx/msxbatoken/msxbatoken.py` — implementação completa do tokenizador (Python)
- `E:/msxedit/basic-dignified/msx/msxbader.py` — detokenizador (Python)
- `E:/msxedit/basic-dignified/documentation/BATOKEN.md` — documentação oficial do BATOKEN
- `E:/msxedit/internal/basic/tokens.go` — stub atual em Go (tabela parcial, LoadTokenized não implementado)
