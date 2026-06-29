# Manual de Operação e Compilação - MSXEdit

Este documento descreve o estado operacional atual do MSXEdit na release `4.1.5`.

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

## Atalhos e navegação — msxedit

### Atalhos atualmente funcionais

- **F1**: abre a janela de `Help`
- **Alt+F1**: volta ao tópico anterior dentro do `Help`
- **Alt+Q**: fallback para voltar no `Help` em terminais que não repassam `Alt+F1`
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

- `File`: possui fluxo de abertura do dropdown e ação real de `Exit`
- `Help`: possui fluxo real para `Contents` e `About`
- `Options`: possui o diálogo `Compiler/Interpreter Options`
- Demais menus: ainda atuam como scaffolding visual e mostram `No options yet`

### 4. Diálogos e componentes reutilizáveis

- **`dialogoOK`**: diálogo reutilizável para confirmação/informação
- **`showDialogoOKCentered(...)`**: helper para centralização automática
- **`turboButton`**: botão reutilizável com hotkey destacada
- **`compilerOptionsDialog`**: diálogo de opções com 9 radio buttons, 3 checkboxes em bolinha,
  área `Conditional defines:` e botões `OK`/`Cancel`/`Help`
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

Pressione `F1` para abrir o overlay de ajuda com todas as teclas listadas.
Qualquer tecla fecha o overlay.

## Limitações atuais

Os itens abaixo já possuem estrutura de configuração, UI ou roadmap, mas **ainda não estão concluídos**:

- `Open` / `Save` completos no msxedit
- `Compile` / `Make`
- syntax highlighting efetivo no editor
- números de linha visíveis usando `show_line_numbers`

Para detalhes objetivos de opções e comportamento, consulte [`REFERENCE.md`](REFERENCE.md).
