# MSXEdit

<figure>
  <img src="images/MSX-Edit-01.png" alt="Banner do MSXEdit">
  <figcaption>Banner do MSXEdit</figcaption>
</figure>

MSXEdit é um editor TUI com estética retrô, pensado para desenvolvimento em plataformas clássicas como MSX, mas executado em terminais modernos no Windows e Linux.

O projeto combina uma base visual inspirada em Turbo Vision, Norton Editor e ferramentas Borland com uma arquitetura Go moderna, priorizando janelas customizadas, temas VGA explícitos e navegação por teclado/mouse.

## Release atual

- **Versão**: `4.1.7`
- **Build ID**: gerado dinamicamente em hexadecimal UTC durante a execução/build.

## Novidades da 4.1.7

- **Diálogo Open File (F3)** no MSXEdit: janela estilo Turbo Pascal com campo `&Name` (fundo azul escuro), lista `&Files` bicoluna em ciano (sem moldura), barra de rolagem horizontal azul, área de status completa (caminho+máscara, item selecionado) e quatro botões 3D de largura igual: `Open`, `Replace`, `Cancel`, `Help`.
- **Tela de ajuda do MSX-Read** aprimorada: cabeçalho com "Welcome to MSX-Read v:(versão)" centralizado, seguido de "MSX-Read Help Screen" e linha de copyright.
- Layout do diálogo Open File refinado: espaçamento entre blocos Name/Files/Status, todos os botões com largura idêntica (11 colunas), Open e Replace alinhados com o campo de entrada.

## O que já está implementado

- **Janela de edição retrô no startup**: a aplicação abre automaticamente a primeira janela de edição, com moldura dupla, botão `[■]`, título centralizado, identificador numérico e barras de rolagem desenhadas manualmente.
- **Desktop quadriculado estilo DOS**: o fundo da aplicação usa renderização dedicada com padrão VGA clássico.
- **Barra de menus Turbo-like**: menus superiores com navegação por `Alt+Letra`, `F10`, setas, `Enter` e clique do mouse.
- **Estrutura atual de menus**:
  - Ativos: `File` (Open F3, Exit), `Options` (Compiler/Interpreter…), `Help` (Contents, About)
  - Estruturais / placeholder visual: `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Tools`, `Window`
- **Diálogo Open File (F3)**:
  - Ativado por `F3` globalmente ou pelo menu `File → Open… F3`
  - Campo `&Name` com fundo azul escuro e seta `↓` verde para histórico de máscaras
  - Lista `&Files` bicolunar em ciano, sem moldura, com separador de diretórios
  - Barra de rolagem horizontal em azul com controles `◄▒■►`
  - Área de status completa: caminho+máscara na linha 1; nome, tipo, data, hora e tamanho na linha 2
  - Quatro botões Turbo Vision: `Open`, `Replace`, `Cancel`, `Help` — todos com largura 11
- **Sistema de Help navegável**:
  - carregamento automático do arquivo externo [`HELP.md`](HELP.md)
  - fallback para tópicos internos quando o markdown não estiver disponível
  - links entre tópicos, breadcrumb de navegação
  - retorno por `Alt+F1` (com fallback `Alt+Q`)
  - suporte a teclado e mouse
- **Diálogos reutilizáveis**: componente `dialogoOK` com centralização automática, botão configurável e fechamento por teclado/mouse.
- **Botões estilo Turbo Vision**: componente `turboButton` com hotkey destacada e modos de sombra `shadowModeTurboClassic` e `shadowModeFlat`.
- **Diálogo de opções do compilador/interpretador**: janela `Compiler/Interpreter Options` com 9 radio buttons, 3 checkboxes com marcador em bolinha, área `Conditional defines:` e botões `OK`/`Cancel`/`Help`.
- **Syntax highlighting MSX-BASIC** no editor: mais de 100 keywords, 11 categorias de token, números em todas as bases, zonas literais `REM`/`DATA`/`string`/apóstrofo.
- **Windowing flutuante**: arrastar pela barra de título, redimensionar pelo canto `◢`, maximizar/restaurar com `[▲]`/`[▼]`, scrollbars clicáveis.
- **Tema VGA padronizado**:
  - `default`: Borland blue / MS-DOS clássico
  - `blue`: NC-style / Norton Commander, com menu e status em ciano
- **CLI e configuração persistente**:
  - `--theme`, `--tabsize`, `--no-highlight`, `--local`, `--version`
- **Build automatizado**: script [`build.ps1`](build.ps1) extrai a versão automaticamente de `cmd/msxedit/main.go`, gera `Build ID` e compila para Windows ou Linux.

## Estado atual do projeto

### Em funcionamento hoje

- edição de texto em `TextArea`
- janelas e diálogos customizados
- menu superior com dropdown
- diálogo Open File (F3) — navegação de arquivos e diretórios
- janela de `Help`
- `About`
- temas e configuração base
- suporte a mouse em áreas principais da UI

### Em evolução / ainda não finalizado

- leitura efetiva do arquivo selecionado no diálogo Open (integração com editor)
- fluxo completo de `Save`
- ações reais de `Compile` / `Make`
- renderização de números de linha usando `show_line_numbers`

<figure>
  <img src="images/MSX-Edit-02.png" alt="Tela do MSXEdit">
  <figcaption>Tela principal do MSXEdit em execução</figcaption>
</figure>

## msxread — visualizador companheiro

O projeto inclui um segundo executável, **`msxread`**, um visualizador de textos no espírito do
leitor de `README` do Turbo Pascal. Funciona de forma independente do `msxedit`.

<figure>
  <img src="images/MSX-Read-01.png" alt="Tela do MSX-Read">
  <figcaption>MSX-Read em execução</figcaption>
</figure>

- **Tipos suportados**: `.txt` (texto puro), `.bas` **tokenizado** (detokenizado para listagem
  BASIC legível) e `.md` (ajuda, com render leve de títulos e links).
- **Tela de ajuda (F1)**: overlay com cabeçalho "Welcome to MSX-Read v:(versão)" centralizado,
  copyright Cybernostra e lista completa de atalhos.
- **Layout**: barra de topo cinza (data `◆` hora `◆` nome do arquivo), corpo com fundo cyan e
  barra de status `Command►` com indicador de posição (`*** Top of File ***`) e mini-help de teclas.
- **Navegação**: setas, `PgUp`/`PgDn`, `Home`/`End`, roda do mouse, `F1` (ajuda) e `ESC` (sair).
- **Busca**: `F` inicia, digitar filtra em tempo real, `N` próxima ocorrência, `C` case-sensitive.
- **Cores**: `F5`/`F6` cor do texto, `F7`/`F8` cor do fundo — 16 cores VGA.
- **CLI** (via `cobra`):

```powershell
.\msxread.exe MANUAL.md
.\msxread.exe programa.bas
.\msxread.exe --type txt arquivo.dat
.\msxread.exe --version
```

O detokenizador MSX-BASIC vive em `internal/basic` e segue a referência de [`TOKEN.md`](TOKEN.md).

## Stack do projeto

- **Go 1.26+**
- **tview** e **tcell**
- **PowerShell** para automação de build
- **GPL 3.0**

## Documentação

- [`MANUAL.md`](MANUAL.md): compilação, operação, atalhos e limitações atuais
- [`REFERENCE.md`](REFERENCE.md): opções de CLI, configuração e comportamento da UI
- [`HELP.md`](HELP.md): conteúdo do sistema de ajuda carregado em runtime
- [`OUTLINE.md`](OUTLINE.md): histórico conceitual e decisões de arquitetura
- [`TOKEN.md`](TOKEN.md): referência do formato binário MSX-BASIC
