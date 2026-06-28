# Changelog

Todas as mudancas relevantes deste projeto serao documentadas neste arquivo.

Formato baseado em categorias (`Added`, `Changed`, `Fixed`, `Docs`).

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

