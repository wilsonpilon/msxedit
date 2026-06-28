# Manual de Operação e Compilação - MSXEdit

Este documento descreve o estado operacional atual do MSXEdit na release `4.0.7`.

## Compilação

O MSXEdit é escrito em Go e requer **Go 1.26 ou superior**.

### Usando o script `build.ps1` (recomendado no Windows)

O script:

- lê a versão diretamente de `cmd/msxedit/main.go`
- gera um `Build ID` hexadecimal em UTC
- executa `go mod tidy`
- compila para Windows ou Linux

Exemplos:

```powershell
# Windows / release (padrão)
.\build.ps1

# Linux / release
.\build.ps1 -Linux

# Build de desenvolvimento
.\build.ps1 -Dev

# Build de desenvolvimento e execução imediata
.\build.ps1 -Dev -Run
```

Parâmetros relevantes do script:

- `-Windows`: força alvo Windows
- `-Linux`: gera binário Linux
- `-Release`: usa flags de release (comportamento padrão)
- `-Dev`: compila sem flags de stripping
- `-Run`: executa o binário gerado quando o alvo for Windows

### Compilação manual

```bash
go mod tidy
go build -o msxedit.exe ./cmd/msxedit/main.go
```

## Inicialização

O programa pode receber um caminho de arquivo opcional na linha de comando:

```powershell
.\msxedit.exe
.\msxedit.exe meu_arquivo.txt
.\msxedit.exe --theme blue
```

Ao iniciar, a aplicação sempre cria a **janela de edição 1**, mesmo quando nenhum arquivo é informado.

## Atalhos e navegação

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

## Interface

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
- Demais menus: ainda atuam como scaffolding visual e mostram `No options yet`

### 4. Diálogos e componentes reutilizáveis

- **`dialogoOK`**: diálogo reutilizável para confirmação/informação
- **`showDialogoOKCentered(...)`**: helper para centralização automática
- **`turboButton`**: botão reutilizável com hotkey destacada
- **Modos de sombra do botão**:
  - `shadowModeTurboClassic`
  - `shadowModeFlat`

## Sistema de Help

O `Help` carrega seu conteúdo a partir do arquivo [`HELP.md`](HELP.md), buscando-o a partir do diretório atual e subindo na árvore de diretórios. Se o arquivo não for encontrado ou não puder ser interpretado, a aplicação usa um conjunto de tópicos internos como fallback.

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

## Limitações atuais

Os itens abaixo já possuem estrutura de configuração, UI ou roadmap, mas **ainda não estão concluídos**:

- `Open` / `Save` completos
- `Compile` / `Make`
- syntax highlighting efetivo no editor
- parser de BASIC tokenizado (`.BAS` binário)
- números de linha visíveis usando `show_line_numbers`

Para detalhes objetivos de opções e comportamento, consulte [`REFERENCE.md`](REFERENCE.md).
