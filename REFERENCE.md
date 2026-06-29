# Referência de Opções - MSXEdit

Este arquivo resume as opções de linha de comando, chaves de configuração e comportamentos visíveis da release `4.1.0`.

## Opções de Linha de Comando

| Opção | Descrição |
|-------|-----------|
| `--help` | Exibe a mensagem de ajuda com todas as opções disponíveis. |
| `--version` | Exibe a versão corrente do programa (atualmente `4.0.7`) com o `Build ID` gerado na execução. |
| `--local` | Força o uso do arquivo `msxedit.json` no diretório atual em vez do diretório global. |
| `--theme <nome>` | Define o tema de cores da interface (`default` ou `blue`). |
| `--tabsize <n>` | Define o tamanho do caractere Tab (ex: 4 ou 8). |
| `--no-highlight` | Desativa o realce de sintaxe MSX-BASIC no editor. |

### Argumento posicional

Além das flags, o programa aceita um argumento opcional com caminho de arquivo:

```text
msxedit [opções] [arquivo]
```

Quando informado, o nome do arquivo é usado no título da primeira janela de edição.

## Configurações (msxedit.json)

As seguintes chaves podem ser configuradas no arquivo JSON:

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
- **show_line_numbers**: Boolean. Preferência já persistida em configuração, mas a margem visual com números de linha ainda não é desenhada na UI atual.
- **highlight**: Boolean. Ativa/desativa o realce de sintaxe MSX-BASIC no editor.

### Ordem de precedência

1. Flags de linha de comando
2. Configuração local (`--local`)
3. Configuração global do usuário
4. Padrões internos

## Temas de Cores Disponíveis

- **`default`**: VGA Borland blue (estilo MS-DOS/Turbo).
- **`blue`**: VGA NC-style (Norton Commander), com barra superior e status em ciano.

Ambos os temas aplicam paleta explícita para:

- desktop quadriculado
- barra de menus
- barra de status
- janela do editor
- popups e diálogos
- janela de `Help`

## Componentes Internos de UI

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

## Menus e Atalhos Atuais

### Menus superiores

- `Options` -> `Compiler/Interpreter...`: abre a janela `Compiler/Interpreter Options`
  - radio buttons: `MSX-BASIC`, `Basic Dignified`, `MSX Bas2Rom`, `Turbo Basic`, `NBasic`, `MSXgl/SDCC`, `N80/LK80`, `ASCII-C`, `Turbo Pascal 3.3f`
  - checkboxes: `Extended syntax`, `Overflow checking`, `Strict vars`
  - área `Conditional defines:` para texto livre
  - botões `OK`, `Cancel` e `Help`

- `File`
- `Edit`
- `Search`
- `Run`
- `Compile`
- `Debug`
- `Tools`
- `Options`
- `Window`
- `Help`

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

## Janela de Help

O sistema de ajuda:

- carrega tópicos a partir de `HELP.md`
- usa fallback interno se o markdown não estiver disponível
- suporta links entre tópicos
- mostra breadcrumb de navegação
- suporta teclado e mouse

Controles principais dentro do `Help`:

- `Tab / Shift+Tab`: navegar entre links
- `Enter`: seguir link
- `Esc`: fechar
- `Setas`, `PgUp`, `PgDn`, `Home`, `End`: rolagem e navegação

## Syntax Highlighting MSX-BASIC

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

Zonas literais tratadas corretamente: `REM` (tudo comentário), `DATA` (tudo default, exceto strings), strings `"…"`, comentário por apóstrofo `'`.

## Recursos em andamento

- fluxo de `Open` / `Save`
- ações de `Compile` / `Make`
- parser de arquivo BASIC tokenizado (`.BAS` binário) — referência completa em `TOKEN.md`
- renderização de números de linha na margem do editor
- submenus de `Environment`, `Window` e restantes de `Options`
