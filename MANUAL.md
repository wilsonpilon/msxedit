# Manual de Operação e Compilação - MSXEdit

Este documento descreve o estado operacional atual do MSXEdit na release `4.1.9`.

## Compilação

O MSXEdit é escrito em Go e requer **Go 1.26 ou superior**.

### Usando o script `build.ps1` (recomendado no Windows)

O script compila um ou ambos os executáveis do pacote (`msxedit` e `msxread`), gerando um
**Build ID** hexadecimal em tempo de compilação, comum a ambos os binários.

```powershell
# Windows / release — compila msxedit e msxread
.\build.ps1

# Compilar apenas o editor
.\build.ps1 -Editor

# Compilar apenas o visualizador
.\build.ps1 -Reader

# Linux / release (cross-compile)
.\build.ps1 -Linux

# Build de desenvolvimento (sem stripping de símbolos)
.\build.ps1 -Dev

# Compilar e abrir o editor imediatamente
.\build.ps1 -Dev -Run

# Compilar e abrir o visualizador com um arquivo
.\build.ps1 -Dev -View MANUAL.md
```

Parâmetros disponíveis:

- `-Windows`: força alvo Windows (comportamento padrão)
- `-Linux`: gera binário para Linux via cross-compile
- `-Release`: flags de release — comportamento padrão
- `-Dev`: compila sem stripping de símbolos de debug
- `-Editor`: compila apenas `msxedit`
- `-Reader`: compila apenas `msxread`
- `-Run`: executa `msxedit` após compilar (apenas Windows nativo)
- `-View`: executa `msxread` após compilar (apenas Windows nativo)

Argumentos extras após as flags são repassados ao binário executado por `-Run` ou `-View`.

### Compilação manual

```bash
go mod tidy
go build -o msxedit.exe ./cmd/msxedit/main.go
go build -o msxread.exe ./cmd/msxread/main.go
```

Nesse caso o `BuildID` fica como `dev`, indicando build manual sem o script.

## Inicialização — msxedit

O programa pode receber um caminho de arquivo opcional na linha de comando:

```powershell
.\msxedit.exe
.\msxedit.exe meu_arquivo.txt
.\msxedit.exe --theme blue
```

Ao iniciar, a aplicação sempre cria a **janela de edição 1**, mesmo quando nenhum arquivo é informado.

`File → New` cria janelas de edição adicionais em cascata a partir da janela ativa, cada uma com
seu próprio número e nome (`Sem Nome` até ser salva). Todas as janelas compartilham o mesmo
clipboard. Fechar uma janela (`Alt+F3` ou clique em `[■]`) não encerra o programa — mesmo fechando
a última janela aberta, o desktop fica visível e `File → New`/`Open…` continuam disponíveis.

## Atalhos e navegação — msxedit

### Atalhos atualmente funcionais

- **F1**: abre a janela de `Help`
- **F3**: abre o diálogo `Open File`
- **Alt+F3**: fecha a janela de edição ativa
- **Alt+F1**: volta ao tópico anterior dentro do `Help`
- **Alt+Q**: fallback para voltar no `Help` em terminais que não repassam `Alt+F1`
- **Ctrl+F1**: "Language help" — abre o `Help` já navegado até `Reserved Words`
- **Ctrl+L**: repete a última busca (`Search again`)
- **F10**: abre o menu `File`
- **Alt+X**: encerra a aplicação
- **Alt + letra do menu**:
  - `Alt+F` = `File`
  - `Alt+E` = `Edit`
  - `Alt+S` = `Search`
  - `Alt+R` = `Run`
  - `Alt+C` = `Compile`
  - `Alt+D` = `Debug`
  - `Alt+T` = `Tools`
  - `Alt+O` = `Options`
  - `Alt+W` = `Window`
  - `Alt+H` = `Help`

### Edit — atalhos de clipboard e histórico

| Tecla | Ação |
|-------|------|
| `Alt+BkSp` | Undo (nativo do editor) |
| — (só pelo menu `Edit → Redo`) | Redo — sem atalho direto, pois `Ctrl+Y` já é "Delete line" |
| `Shift+Del` | Cut para o clipboard compartilhado |
| `Ctrl+Ins` | Copy para o clipboard compartilhado |
| `Shift+Ins` | Paste do clipboard compartilhado |
| `Ctrl+Del` | Clear (apaga o bloco selecionado sem copiar) |

`Edit → Show clipboard` abre uma janela dedicada (borda amarela) que exibe e permite editar o
clipboard compartilhado; qualquer Copy/Cut/Paste em qualquer janela de edição a atualiza ao vivo.

### Search — Find, Replace e Go to line number

- **`Ctrl+Q F`** ou **`Search → Find...`**: abre o diálogo `Find`, com opções `Case sensitive`,
  `Whole words only`, `Regular expression`, `Direction` (Forward/Backward), `Scope`
  (Global/Selected text) e `Origin` (From cursor/Entire scope). Confirma a busca real no editor
  ativo e avisa na barra de status quando o texto não é encontrado ou a busca dá a volta (wrap).
- **`Ctrl+L`** ou **`Search → Search again`**: repete a última busca a partir do cursor; sem busca
  anterior, abre o `Find`.
- **`Ctrl+Q A`** ou **`Search → Replace...`**: abre o diálogo `Replace` (texto a buscar, novo
  texto, `Prompt on replace`, mesmas opções/grupos do `Find`, botões `Ok`/`Change all`). O diálogo
  já captura e memoriza todas as opções e o histórico dos dois campos — **a substituição efetiva
  do texto no editor ainda não está implementada**.
- **`Search → Go to line number...`**: move o cursor para uma linha de texto (1-based) ou, com a
  opção "linha do MSX-Basic" marcada, para a linha cujo número BASIC (dígitos no início da linha,
  ex.: `10 PRINT...`) seja o informado. Avisa na barra de status se a linha não existir.
- `Search → Show last compiler error`, `Find error...` e `Find procedure...` permanecem
  placeholders sem ação.

### Navegação em menus dropdown

- **Seta esquerda/direita**: alterna entre menus
- **Seta cima/baixo**: percorre itens do menu aberto
- **Enter**: ativa o item selecionado
- **Esc**: fecha o menu aberto

### Navegação na janela de Help

- **Tab / Shift+Tab**: alterna entre links
- **Enter**: abre o link selecionado
- **Setas**: rolagem ou seleção, dependendo do tópico
- **Page Up / Page Down / Home / End**: navegação rápida
- **Esc**: fecha o `Help`

## Edição — atalhos estilo WordStar/Turbo

### Movimentação de cursor

| Ação | Tecla |
|------|-------|
| Caractere esquerda/direita | `Ctrl+S`/`Ctrl+D` ou setas |
| Palavra esquerda/direita | `Ctrl+A`/`Ctrl+F` ou `Ctrl+`setas |
| Linha cima/baixo | `Ctrl+E`/`Ctrl+X` ou setas |
| Rolar tela cima/baixo (sem mover cursor) | `Ctrl+W`/`Ctrl+Z` |
| Página cima/baixo | `Ctrl+R`/`Ctrl+C` ou `PgUp`/`PgDn` |
| Meia tela cima/baixo | `Ctrl+Up`/`Ctrl+Down` |
| Início/fim do arquivo | `Ctrl+Home`/`Ctrl+End` |

### Seleção de texto

`Shift` + qualquer tecla de movimentação (setas, `Home`, `End`, `PgUp`, `PgDn`) inicia/estende uma
seleção a partir do ponto onde o cursor estava. Arrastar o mouse com o botão esquerdo pressionado
faz o mesmo. Em ambos os casos, a seleção reaproveita o destaque e os comandos de bloco/clipboard
descritos em `Ctrl+K`/`Ctrl+Q` (Copy, Cut, Delete, Paste).

### Inserção e remoção

| Ação | Tecla |
|------|-------|
| Alternar modo Insert/Overwrite | `Ins` (sem modificador) ou `Ctrl+V` |
| Inserir linha em branco | `Ctrl+N` |
| Apagar a linha inteira | `Ctrl+Y` |
| Apagar até o fim da linha | `Ctrl+Q Y` |
| Apagar caractere à esquerda | `Backspace` ou `Ctrl+H` |
| Apagar caractere à direita | `Del` ou `Ctrl+G` |
| Apagar palavra à direita | `Ctrl+T` |

### Comandos de bloco (`Ctrl+K` / `Ctrl+Q`)

`Ctrl+K` e `Ctrl+Q` são prefixos: pressione a combinação e, em seguida, a segunda tecla.

| Prefixo | Tecla | Ação |
|---------|-------|------|
| `Ctrl+K` | `B` | Marca início do bloco |
| `Ctrl+K` | `K` | Marca fim do bloco |
| `Ctrl+K` | `T` | Marca uma palavra |
| `Ctrl+K` | `L` | Marca a linha |
| `Ctrl+K` | `C` | Copia bloco |
| `Ctrl+K` | `V` | Move bloco |
| `Ctrl+K` | `Y` | Apaga bloco |
| `Ctrl+K` | `R` / `W` | Lê/grava bloco em disco |
| `Ctrl+K` | `H` | Mostra/esconde bloco |
| `Ctrl+K` | `P` | Imprime bloco |
| `Ctrl+K` | `I` / `U` | Indenta/desindenta bloco |
| `Ctrl+K` | `D` | Sai para a barra de menu |
| `Ctrl+K` | `S` | Salva o arquivo (igual a `F2`) |
| `Ctrl+K` | `0`–`9` | Grava a posição do cursor no marcador *n* |
| `Ctrl+Q` | `B` / `K` | Vai ao início/fim do bloco |
| `Ctrl+Q` | `Y` | Apaga até o fim da linha |
| `Ctrl+Q` | `L` | Restaura a linha atual (desfaz todas as edições feitas nela) |
| `Ctrl+Q` | `F` | Abre `Find` |
| `Ctrl+Q` | `A` | Abre `Replace` |
| `Ctrl+Q` | `0`–`9` | Vai até a posição gravada no marcador *n* |
| `Ctrl+Ins` / `Shift+Del` / `Shift+Ins` / `Ctrl+Del` | — | Copy / Cut / Paste / Clear no clipboard compartilhado |

### Outros prefixos e comandos

| Tecla | Ação |
|-------|------|
| `Ctrl+O T` | Alterna modo Tab: caractere de tabulação (padrão) ou espaços até a próxima parada (8 colunas) |
| `Ctrl+O I` | Alterna Auto indent: repete a indentação da linha anterior ao pressionar Enter |
| `Ctrl+P` + tecla | Insere o código de controle (1–26) da próxima tecla `Ctrl+letra`, estilo WordStar |
| `Esc` | Cancela um prefixo `Ctrl+K`/`Ctrl+Q`/`Ctrl+O`/`Ctrl+P` pendente |

## Interface — msxedit

### 1. Desktop e layout principal

- fundo quadriculado estilo DOS
- barra de menu no topo
- área central com janela de edição customizada
- barra de status na base

### 2. Janela de edição

A janela principal de edição possui:

- moldura dupla branca
- controle visual `[■]` no topo esquerdo
- título centralizado com nome do arquivo
- número da janela no topo
- indicador `[↕]` na moldura superior
- coordenadas de cursor na barra inferior da própria moldura
- barras de rolagem horizontal e vertical desenhadas manualmente

### 3. Menus

O layout visual já inclui os menus:

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

Estado atual:

- `File`: `New` (nova janela em cascata), `Open…`, `Exit` implementados
- `Edit`: `Undo`, `Redo`, `Cut`, `Copy`, `Paste`, `Clear`, `Show clipboard` implementados
- `Search`: `Find...`, `Replace...`, `Search again`, `Go to line number...` implementados;
  `Show last compiler error`, `Find error...` e `Find procedure...` ainda são placeholders
- `Help`: possui fluxo real para `Contents` e `About`
- `Options`: possui o diálogo `Compiler/Interpreter Options`
- Demais menus (`Run`, `Compile`, `Debug`, `Tools`, `Window`): ainda atuam como scaffolding visual
  e mostram `No options yet`

### 4. Diálogos e componentes reutilizáveis

- **`dialogoOK`**: diálogo reutilizável para confirmação/informação
- **`showDialogoOKCentered(...)`**: helper para centralização automática
- **`turboButton`**: botão reutilizável com hotkey destacada
- **`compilerOptionsDialog`**: diálogo de opções com 9 radio buttons, 3 checkboxes em bolinha,
  área `Conditional defines:` e botões `OK`/`Cancel`/`Help`
- **`openFileDialog`**: diálogo Open File estilo Turbo Pascal, ativado por `F3`:
  - Campo `&Name` (fundo azul escuro) com seta `↓` verde para histórico de máscaras
  - Lista `&Files` bicolunar em ciano, sem moldura; scrollbar horizontal azul
  - Área de status: caminho+máscara / nome, tipo, data, hora, tamanho
  - Botões `Open`, `Replace`, `Cancel`, `Help` — todos largura 11
- **`findDialog`** (`find_dialog.go`): diálogo `Find`, campo com histórico, checkboxes `Case
  sensitive`/`Whole words only`/`Regular expression`, grupos `Direction`/`Scope`/`Origin`.
- **`replaceDialog`** (`replace_dialog.go`): diálogo `Replace`, mesma base do `findDialog` mais
  campo `New text`, checkbox `Prompt on replace` e botões `Ok`/`Change all`.
- **`gotoLineDialog`** (`goto_line_dialog.go`): diálogo `Go to Line Number`, campo numérico e
  checkbox "linha do MSX-Basic".
- **`historyField`** (`text_field.go`): campo de texto de uma linha com histórico (seta `↓`),
  compartilhado pelos diálogos `Find`, `Replace` e `Go to Line Number`.
- **`drawGroupBox` / `drawCheckbox` / `drawRadio`** (`dialog_widgets.go`): helpers de desenho
  reutilizados pelos novos diálogos (caixa com título embutido, checkbox `[x]`, radio `(x)`).
- **Modos de sombra do botão**:
  - `shadowModeTurboClassic`
  - `shadowModeFlat`

## Sistema de Help

O `Help` carrega seu conteúdo a partir do arquivo [`HELP.md`](HELP.md), buscando-o a partir do
diretório atual e subindo na árvore de diretórios. Se o arquivo não for encontrado ou não puder
ser interpretado, a aplicação usa um conjunto de tópicos internos como fallback.

Recursos atuais:

- links entre tópicos
- breadcrumb no topo do conteúdo quando aplicável
- subpainel especial para `Editor Commands`
- rolagem horizontal e vertical
- suporte a clique em links, barras de rolagem e botão `[■]`

## Mouse

O suporte a mouse está ativo para:

- clique na barra de menus
- foco na janela do editor
- fechamento de janelas/dialogos pelo `[■]`
- rolagem e clique dentro da janela de `Help`

## Temas

- **`default`**: VGA Borland blue, visual clássico MS-DOS/Turbo
- **`blue`**: VGA NC-style, com barra de menu/status em ciano

## Configuração

Prioridade de configuração:

1. argumentos de linha de comando
2. `msxedit.json` local, quando `--local` for usado
3. `msxedit.json` no diretório de configuração do usuário
4. valores padrão internos

Localização do arquivo global:

- **Windows**: `%APPDATA%/msxedit/msxedit.json`
- **Linux**: `~/.config/msxedit/msxedit.json`

## msxread — visualizador companheiro

O `msxread` é o segundo executável do pacote MSXEdit, inspirado no leitor de `README` do
Turbo Pascal. Exibe arquivos `.txt`, `.bas` MSX-BASIC tokenizados e `.md` diretamente no terminal.

### Uso

```powershell
.\msxread.exe MANUAL.md
.\msxread.exe programa.bas
.\msxread.exe --type txt arquivo.dat
.\msxread.exe --tabsize 8 codigo.bas
```

### Tipos de arquivo suportados

| Extensão / tipo | Comportamento |
|-----------------|---------------|
| `.txt` | Texto puro com tabulações expandidas |
| `.bas` | Se tokenizado (cabeçalho `0xFF`): detokenizado para listagem BASIC legível com syntax highlighting; em ASCII: tratado como texto |
| `.md` | Render leve: títulos realçados, parágrafos com quebra de linha opcional |

A detecção é automática por extensão e conteúdo. Use `--type` para forçar.

### Layout de tela

- **Topo** (cinza/preto): `data ◆ hora ◆ nome do arquivo` — relógio ao vivo
- **Corpo** (cor configurável, padrão cyan/preto): conteúdo rolável
- **Status** (cinza/preto): `Command►  posição` à esquerda · `Keys: ↑↓←→ PgUp PgDn  ESC=Exit  F1=Help` à direita

### Teclas de navegação

| Tecla | Ação |
|-------|------|
| `↑` / `↓` | Uma linha acima / abaixo |
| `←` / `→` | Oito colunas à esquerda / direita |
| `PgUp` / `PgDn` | Página anterior / próxima |
| `Home` | Início do arquivo |
| `End` | Final do arquivo |
| Roda do mouse | Rolagem vertical |

### Busca

| Tecla | Ação |
|-------|------|
| `F` | Inicia busca — o prompt muda para `Find►  _` |
| (digitar) | Adiciona caracteres à query; o primeiro match é destacado em tempo real |
| `Backspace` | Remove último caractere da query |
| `Enter` | Confirma e volta ao modo normal |
| `ESC` | Cancela a busca e remove o destaque |
| `N` | Próxima ocorrência (com wrap ao final) |
| `C` | Alterna busca sensível a maiúsculas/minúsculas (`[Aa]` aparece na barra quando ativo) |

### Cores e aparência

| Tecla | Ação |
|-------|------|
| `F5` / `F6` | Cor do texto: cicla pelas 16 cores VGA |
| `F7` / `F8` | Cor do fundo: cicla pelas 16 cores VGA |

### Quebra de linha

| Tecla | Ação |
|-------|------|
| `W` | Ativa/desativa quebra de linha (word wrap) |

O wrap respeita fronteiras de palavras: nunca quebra uma palavra no meio.
Se uma palavra for maior que a largura da janela, a quebra forçada ocorre no limite.
O scroll horizontal (`←`/`→`) é desativado automaticamente no modo wrap.

### Hi-bit (caracteres acima de 127)

| Tecla | Ação |
|-------|------|
| `7` | Desativa hi-bit — bytes 128–255 são substituídos por `·` (modo 7 bits) |
| `8` | Ativa hi-bit — bytes 128–255 são exibidos como-estão (modo 8 bits, padrão) |

### Impressão e configurações

| Tecla | Ação |
|-------|------|
| `P` | Imprime o arquivo (Windows: `notepad /p`; Linux: `lpr`) |
| `S` | Salva as configurações atuais (cor, wrap, hi-bit) em `msxread.json` |

As configurações são salvas no mesmo diretório do executável (`msxread.json`).

### Flags de linha de comando

| Flag | Descrição |
|------|-----------|
| `--type`, `-t` | Força tipo: `auto`, `txt`, `bas`, `md` |
| `--tabsize` | Espaços por tabulação (padrão `4`) |
| `--width`, `-w` | Largura máxima para quebra markdown em pré-processamento (0 = sem limite) |
| `--version`, `-v` | Exibe versão e Build ID |
| `--help`, `-h` | Exibe ajuda do cobra |

### F1 — Overlay de ajuda

Pressione `F1` para abrir o overlay de ajuda. O overlay exibe:

- Cabeçalho "Welcome to MSX-Read v:(versão)" centralizado.
- "MSX-Read Help Screen" e "Copyright (c) 1972,2026 Cybernostra, Inc.".
- Lista completa de teclas disponíveis.

Qualquer tecla fecha o overlay.

## Limitações atuais

Os itens abaixo já possuem estrutura de configuração, UI ou roadmap, mas **ainda não estão concluídos**:

- leitura efetiva do arquivo selecionado no diálogo `Open File` (integração com editor)
- substituição efetiva de texto no diálogo `Replace` (opções, histórico e o `Find` embutido nele
  já funcionam; falta a ação de gravar o `New text` no lugar das ocorrências)
- fluxo de `Save` completo no msxedit
- `Compile` / `Make`
- números de linha visíveis usando `show_line_numbers`

Para detalhes objetivos de opções e comportamento, consulte [`REFERENCE.md`](REFERENCE.md).
