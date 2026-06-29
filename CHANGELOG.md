# Changelog

Todas as mudancas relevantes deste projeto serao documentadas neste arquivo.

Formato baseado em categorias (`Added`, `Changed`, `Fixed`, `Docs`).

## [4.1.0] - 2026-06-29

### Added
- **Syntax highlighting** para MSX-BASIC no editor:
  - Novo arquivo `msxbasic_highlight.go` com tokenizador completo (`tokenizeBasicLine`)
  - Categorias: número de linha, statement, modifier, função built-in, string, número, comentário, variável, operador, símbolo
  - Mais de 100 keywords mapeadas na tabela `msxKeywordMap`
  - Números em todas as bases: decimal, `&H` (hex), `&O` (octal), `&B` (binário), com sufixos de tipo
  - Zonas literais: `REM …`, `DATA …`, strings `"…"`, comentário `'…`
  - Campo `highlightEnabled` em `editorWindow` controla ativação por janela
- **Janela de edição flutuante** com windowing completo:
  - Arrastar pela barra de título (clique e movimento)
  - Redimensionar pelo canto inferior direito (`◢`)
  - Botão de zoom `[▲]`/`[▼]` para maximizar/restaurar
  - Barras de rolagem horizontal e vertical clicáveis
  - `MouseScrollUp`/`Down` rola o conteúdo
  - Captura de mouse durante drag/resize via `return (true, w)` no `MouseHandler`
  - Posicionamento e tamanho armazenados em `winX/winY/winW/winH` (independentes do layout flex)
- **Diálogo Compiler/Interpreter Options** (`compiler_options_dialog.go`):
  - 9 opções de linguagem em dois grupos: Basic Code (MSX-BASIC, Basic Dignified, MSX Bas2Rom, Turbo Basic, NBasic) e Others (MSXgl/SDCC, N80/LK80, ASCII-C, Turbo Pascal 3.3f)
  - 3 checkboxes: Extended syntax, Overflow checking, Strict vars
  - Área de texto "Conditional defines:" com cursor e edição completa (insert, delete, Enter para nova linha)
  - 3 botões estilo Turbo Vision: OK, Cancel, Help (com destaque de foco)
  - Navegação por Tab, Shift+Tab, setas, Space, hotkeys O/C/H
  - Atalho `Help` abre diálogo de ajuda secundário (`showCompilerOptionsHelp`)
- **Menu Options** expandido com 12 itens:
  - Compiler/Interpreter… (ativo — abre diálogo)
  - Memory Sizes…, Linker…, Debugger…, Directories, Tools (scaffold)
  - Environment ► (submenu futuro)
  - Open…, Save …\msxedit.cfg, Save as… (scaffold)
- **TOKEN.md**: referência completa do formato binário MSX-BASIC (tabela de tokens simples 0x81–0xFC, tokens estendidos 0xFF+byte, encoding de números, estrutura de linha, algoritmos de tokenização/detokenização com exemplos hexadecimais anotados)

### Changed
- `editorWindow` reescrita para windowing flutuante: ignora rect do layout flex, usa `winX/winY/winW/winH` próprios; inicializa posição na primeira chamada a `Draw()`
- `tui.go`: janela do editor adicionada diretamente ao `pages` (sem wrapper `editorHost` Flex), permitindo posicionamento livre na tela
- Canto inferior direito da janela de edição mudou de `╝` para `◢` (handle de resize)
- Botão de zoom mudou de `[↕]` para `[▲]`/`[▼]` (indica estado atual)

### Docs
- `TOKEN.md` criado como referência definitiva de tokenização MSX-BASIC
- `CHANGELOG.md`, `OUTLINE.md`, `REFERENCE.md` atualizados para refletir o estado da release 4.1.0

## [4.0.7] - 2026-06-27

### Added
- Janela de `Help` navegavel com conteudo carregado de `HELP.md`, fallback interno e topicos interligados por links.
- Navegacao por breadcrumb no `Help`, com retorno por `Alt+F1` e fallback `Alt+Q` para terminais que nao entregam a combinacao original.
- Suporte a mouse na interface principal para:
  - clique na barra de menus
  - clique no controle `[■]` para fechar janelas/dialogos
  - clique em links e barras de rolagem da janela de `Help`
- Subjanela centralizada para `Editor Commands` dentro do `Help`, com cabecalho estilo Turbo e lista selecionavel.

### Changed
- Documentacao principal sincronizada com o estado real do projeto:
  - `README.md`
  - `MANUAL.md`
  - `HELP.md`
  - `REFERENCE.md`
  - `OUTLINE.md`
- Menus documentados de forma mais precisa: `File` e `Help` possuem fluxos ativos; os demais itens permanecem como scaffolding visual para iteracoes futuras.
- A release-base do projeto passa a ser identificada como `4.0.7`.

### Fixed
- Referencias desatualizadas de versao (`4.0.3`) substituidas por `4.0.7` onde representavam a versao corrente.
- Documentacao corrigida para separar recursos implementados de recursos ainda planejados (como `Open/Save`, `Compile/Make`, `syntax highlight` completo e parser tokenizado de BASIC).

### Docs
- Ajuda embarcada reescrita para refletir o comportamento atual da UI, atalhos, temas, mouse, menus e limitacoes conhecidas.
- Manual e referencia atualizados com parametros reais de build/configuracao e observacoes sobre funcionalidades em andamento.

## [4.0.3] - 2026-06-27

### Added
- Janela de edicao inicial customizada no startup, com suporte a multiplas janelas na estrutura interna (`Editors` e `ActiveEditor`).
- Renderizacao de janela estilo Turbo Vision com moldura dupla, titulo centralizado, identificador de janela e controles no topo.
- Barras de rolagem customizadas na moldura (horizontal e vertical), com setas, trilho e cursor visual.
- Componente reutilizavel `dialogoOK` para dialogos de confirmacao/informacao.
- Helper de layout `showDialogoOKCentered(...)` para centralizar qualquer `dialogoOK`.
- Componente reutilizavel `turboButton` para botao estilo Turbo Vision.
- Configuracao de sombra do botao por modo:
  - `shadowModeTurboClassic`
  - `shadowModeFlat`

### Changed
- Tema `default` migrado para paleta VGA explicita com foco em Borland blue classico (`RGB 0,0,170`).
- Tema `blue` padronizado para usar apenas a mesma tabela VGA interna (sem mistura com `tcell.Color*`).
- Dialogo `About` refeito como dialogo custom:
  - fundo cinza
  - moldura dupla branca
  - controle `[■]` no topo esquerdo com quadrado verde
  - botao `O&K` verde com texto branco e hotkey em amarelo
- `About` agora usa o componente `dialogoOK` e abertura centralizada.
- Menus `Help` atualizados com novos subitens e separadores conforme especificacao.

### Fixed
- Ajustes progressivos de alinhamento e orientacao da sombra 3D do botao para o padrao visual esperado do estilo Turbo.
- Correcao de import nao utilizado apos refatoracoes do dialogo `About`.

### Docs
- Atualizacao de `README.md`, `MANUAL.md`, `REFERENCE.md` e `OUTLINE.md` para refletir:
  - temas VGA (`default`, `blue`)
  - novos componentes de UI reutilizaveis
  - evolucao da arquitetura e padroes visuais
- Versao de referencia atualizada para `4.0.3` na documentacao da epoca.

## [4.0.1] - 2026-06-xx

### Added
- Base de menus interativos (`File/Help`) e hotkeys principais.
- Estrutura inicial de build/versionamento documentada.

### Notes
- Esta secao representa a linha-base historica antes das mudancas detalhadas em `4.0.3`.

