# Referência de Opções - MSXEdit

Este arquivo resume as opções de linha de comando, chaves de configuração e comportamentos
visíveis da release `4.1.5`.

## msxedit — Opções de Linha de Comando

| Opção | Descrição |
|-------|-----------|
| `--help` | Exibe a mensagem de ajuda com todas as opções disponíveis. |
| `--version` | Exibe a versão e o `Build ID` gerado em tempo de compilação. |
| `--local` | Força o uso do arquivo `msxedit.json` no diretório atual em vez do diretório global. |
| `--theme <nome>` | Define o tema de cores da interface (`default` ou `blue`). |
| `--tabsize <n>` | Define o tamanho do caractere Tab (ex: 4 ou 8). |
| `--no-highlight` | Desativa o realce de sintaxe MSX-BASIC no editor. |

### Argumento posicional

```text
msxedit [opções] [arquivo]
```

Quando informado, o nome do arquivo é usado no título da primeira janela de edição.

## msxread — Opções de Linha de Comando

```text
msxread [opções] <arquivo>
```

| Opção | Descrição |
|-------|-----------|
| `--type`, `-t <auto\|txt\|bas\|md>` | Força o tipo de conteúdo. Padrão `auto` (detecta por cabeçalho `0xFF` / extensão). |
| `--tabsize <n>` | Espaços por tabulação ao expandir tabs (padrão `4`). |
| `--width`, `-w <n>` | Largura máxima de linha no pré-processamento markdown (0 = sem limite). |
| `--version`, `-v` | Exibe a versão e o `Build ID`. |
| `--help`, `-h` | Exibe a ajuda gerada pelo cobra. |

## msxread — Teclas do Visualizador

### Busca

| Tecla | Modo | Ação |
|-------|------|------|
| `F` | normal | Inicia busca — status muda para `Find►  _` |
| (rune) | busca | Adiciona caractere à query; primeiro match destacado ao vivo |
| `Backspace` | busca | Remove último caractere da query |
| `Enter` | busca | Confirma busca e volta ao modo normal |
| `ESC` | busca | Cancela e remove destaque |
| `N` | normal | Próxima ocorrência (wrap ao final do arquivo) |
| `C` | normal | Alterna busca sensível a maiúsculas (`[Aa]` na barra quando ativo) |

O match atual é destacado em **amarelo sobre preto**, com prioridade sobre syntax highlighting.

### Navegação

| Tecla | Ação |
|-------|------|
| `↑` / `↓` | Uma linha acima / abaixo |
| `←` / `→` | Oito colunas à esquerda / direita (desativado no modo wrap) |
| `PgUp` / `PgDn` | Página anterior / próxima |
| `Home` | Início do arquivo |
| `End` | Final do arquivo |
| Roda do mouse | Rolagem vertical |

### Cores

| Tecla | Ação |
|-------|------|
| `F5` | Cor do texto: recua na paleta VGA |
| `F6` | Cor do texto: avança na paleta VGA |
| `F7` | Cor do fundo: recua na paleta VGA |
| `F8` | Cor do fundo: avança na paleta VGA |

As 16 cores seguem a paleta VGA clássica (mesma de `internal/tui/theme.go`).

### Quebra de linha (wrap)

| Tecla | Ação |
|-------|------|
| `W` | Ativa / desativa word wrap |

Quando ativo: quebra na última palavra que caiba na linha; palavras maiores que a largura
recebem quebra forçada no limite. A porção da linha exibida é limitada por `visualRow.len`,
garantindo que nenhum caractere da palavra seguinte apareça na linha anterior.

### Hi-bit

| Tecla | Ação |
|-------|------|
| `7` | Desativa hi-bit — bytes 128–255 substituídos por `·` |
| `8` | Ativa hi-bit — bytes 128–255 exibidos como-estão (padrão) |

Aplica-se apenas a bytes no intervalo 128–255 (runes Go `>= 128 && <= 255`).
Caracteres Unicode acima de U+00FF não são afetados.

### Impressão e configurações

| Tecla | Ação |
|-------|------|
| `P` | Imprime o arquivo (Windows: `notepad /p`; Linux: `lpr`) |
| `S` | Salva cor, wrap e hi-bit em `msxread.json` (mesmo diretório do executável) |

### Ajuda e saída

| Tecla | Ação |
|-------|------|
| `F1` | Abre / fecha overlay de ajuda com todas as teclas |
| `ESC` / `Q` | Sai do programa |

## build.ps1 — Parâmetros

| Parâmetro | Descrição |
|-----------|-----------|
| `-Windows` | Força alvo Windows (comportamento padrão) |
| `-Linux` | Cross-compila para Linux (`GOOS=linux GOARCH=amd64`) |
| `-Release` | Flags de release (comportamento padrão) |
| `-Dev` | Compila sem stripping de símbolos de debug |
| `-Editor` | Compila apenas `msxedit` |
| `-Reader` | Compila apenas `msxread` |
| `-Run` | Executa `msxedit` após compilar (apenas Windows nativo) |
| `-View` | Executa `msxread` após compilar (apenas Windows nativo) |

Sem `-Editor` nem `-Reader`, ambos os binários são compilados.
`-Run` e `-View` podem ser combinados.

O Build ID é gerado **uma única vez** no `build.ps1` e injetado em ambos os binários via
`-ldflags "-X main.BuildID=<hex>"`, de modo que `msxedit --version` e `msxread --version`
exibam exatamente o mesmo ID de build.

## Configurações (msxedit.json)

```json
{
  "theme": "default",
  "tab_size": 4,
  "show_line_numbers": true,
  "highlight": true
}
```

- **theme**: String. Nome do tema de cores (`default` ou `blue`).
- **tab_size**: Integer. Espaços por Tab.
- **show_line_numbers**: Boolean. Preferência persistida; margem visual ainda não renderizada.
- **highlight**: Boolean. Ativa/desativa o realce de sintaxe MSX-BASIC no editor.

### Ordem de precedência

1. Flags de linha de comando
2. Configuração local (`--local`)
3. Configuração global do usuário
4. Padrões internos

## Configurações (msxread.json)

Salvo automaticamente no mesmo diretório do executável ao pressionar `S`.

```json
{
  "body_fg": 0,
  "body_bg": 3,
  "wrap_mode": false,
  "hi_bit": true
}
```

- **body_fg**: Índice VGA 0–15 para cor do texto (padrão `0` = preto).
- **body_bg**: Índice VGA 0–15 para cor do fundo (padrão `3` = ciano).
- **wrap_mode**: Boolean. Word wrap ativo ao iniciar.
- **hi_bit**: Boolean. Exibir bytes 128–255 como-estão (padrão `true`).

## Temas de Cores Disponíveis (msxedit)

- **`default`**: VGA Borland blue (estilo MS-DOS/Turbo).
- **`blue`**: VGA NC-style (Norton Commander), com barra superior e status em ciano.

Ambos os temas aplicam paleta explícita para desktop, barra de menus, barra de status,
janela do editor, popups, diálogos e janela de `Help`.

## Paleta VGA Clássica (msxread)

| Índice | Nome | RGB |
|--------|------|-----|
| 0 | Preto | 0, 0, 0 |
| 1 | Azul | 0, 0, 170 |
| 2 | Verde | 0, 170, 0 |
| 3 | Ciano | 0, 170, 170 |
| 4 | Vermelho | 170, 0, 0 |
| 5 | Magenta | 170, 0, 170 |
| 6 | Marrom | 170, 85, 0 |
| 7 | Cinza Cl. | 170, 170, 170 |
| 8 | Cinza Esc. | 85, 85, 85 |
| 9 | Azul Cl. | 85, 85, 255 |
| 10 | Verde Cl. | 85, 255, 85 |
| 11 | Ciano Cl. | 85, 255, 255 |
| 12 | Verm. Cl. | 255, 85, 85 |
| 13 | Magenta Cl. | 255, 85, 255 |
| 14 | Amarelo | 255, 255, 85 |
| 15 | Branco | 255, 255, 255 |

## Componentes Internos de UI (msxedit)

- **`dialogoOK`**: Diálogo reutilizável com botão configurável.
  - `SetButton(label, hotkey, callback)`
  - `SetButtonShadowMode(mode)`
  - `showDialogoOKCentered(dialog, width, height)`
- **`turboButton`**: Botão visual estilo Turbo Vision.
  - Modos de sombra: `shadowModeTurboClassic`, `shadowModeFlat`
- **`compilerOptionsDialog`**: Diálogo completo de opções do compilador/interpretador.
  - 9 radio buttons em dois grupos (Basic Code / Others)
  - 3 checkboxes desenhados com marcador em bolinha (`[•]`)
  - Área de texto editável para defines condicionais
  - Botões OK / Cancel / Help com foco e hotkeys
  - `showCompilerOptionsDialogCentered(dialog, width, height)`
- **`editorWindow`** (windowing flutuante):
  - Arrastar pela barra de título
  - Redimensionar pelo canto `◢`
  - Maximizar/restaurar com botão `[▲]`/`[▼]`
  - Scrollbars H/V clicáveis
  - `highlightEnabled bool` — ativa syntax highlighting MSX-BASIC

## Menus e Atalhos (msxedit)

### Menus superiores

- `Options` -> `Compiler/Interpreter...`: abre a janela `Compiler/Interpreter Options`
  - radio buttons: `MSX-BASIC`, `Basic Dignified`, `MSX Bas2Rom`, `Turbo Basic`, `NBasic`,
    `MSXgl/SDCC`, `N80/LK80`, `ASCII-C`, `Turbo Pascal 3.3f`
  - checkboxes: `Extended syntax`, `Overflow checking`, `Strict vars`
  - área `Conditional defines:` para texto livre
  - botões `OK`, `Cancel` e `Help`

### Menus com ação efetiva hoje

- **`File`**: `Exit`
- **`Options`**: `Compiler/Interpreter…` — abre diálogo de opções completo
- **`Help`**: `Contents` (janela Help navegável) e `About`

Os demais itens de menu existem como estrutura visual/scaffold.

### Hotkeys implementadas

- `F1`: abre o `Help`
- `Alt+F1`: volta ao tópico anterior no `Help`
- `Alt+Q`: fallback para voltar no `Help`
- `F10`: abre o menu `File`
- `Alt+X`: sai da aplicação
- `Alt+F/E/S/R/C/D/T/O/W/H`: abre o menu correspondente

## Janela de Help (msxedit)

- carrega tópicos a partir de `HELP.md`
- usa fallback interno se o markdown não estiver disponível
- suporta links entre tópicos
- mostra breadcrumb de navegação
- suporta teclado e mouse

Controles dentro do `Help`:

- `Tab / Shift+Tab`: navegar entre links
- `Enter`: seguir link
- `Esc`: fechar
- `Setas`, `PgUp`, `PgDn`, `Home`, `End`: rolagem e navegação

## Syntax Highlighting MSX-BASIC (msxedit)

Implementado em `msxbasic_highlight.go`. Categorias de token:

| Categoria | Cor (tema default) | Exemplos |
|-----------|-------------------|---------|
| `basicKindLineNumber` | amarelo | `10`, `100` |
| `basicKindStatement` | ciano brilhante | `PRINT`, `GOTO`, `IF`, `FOR` |
| `basicKindModifier` | verde brilhante | `THEN`, `ELSE`, `TO`, `STEP` |
| `basicKindFunction` | azul brilhante | `LEFT$()`, `PEEK()`, `RND` |
| `basicKindString` | magenta | `"HELLO"` |
| `basicKindNumber` | verde | `42`, `3.14`, `&HFFFF` |
| `basicKindComment` | cinza | `REM …`, `' …` |
| `basicKindVariable` | branco | `A`, `COUNT$`, `X1%` |
| `basicKindOperator` | amarelo | `AND`, `OR`, `+`, `-`, `=` |
| `basicKindSymbol` | branco | `(`, `)`, `,`, `;`, `:` |

Zonas literais tratadas corretamente: `REM` (tudo comentário), `DATA` (tudo default, exceto
strings), strings `"…"`, comentário por apóstrofo `'`.

## Syntax Highlighting MSX-BASIC (msxread)

Implementado em `internal/reader/highlight_basic.go`. Ativo para arquivos `.bas`.

| Categoria | Cor | Exemplos |
|-----------|-----|---------|
| `SpanKeyword` | Amarelo | `PRINT`, `FOR`, `GOTO`, `IF` |
| `SpanString` | Branco | `"HELLO WORLD"` |
| `SpanComment` | Cinza escuro | `REM nota`, `' comentário` |

Detecção de fronteira de palavra garante que variáveis como `FORA` não sejam confundidas com
`FOR`. Keywords são verificadas com matching longest-first.

## Recursos em andamento

- fluxo de `Open` / `Save` no msxedit
- ações de `Compile` / `Make`
- renderização de números de linha na margem do editor
- submenus de `Environment`, `Window` e restantes de `Options`
