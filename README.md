# MSXEdit

MSXEdit é um editor de textos moderno com alma retrô, projetado para desenvolvedores que trabalham com plataformas clássicas como o MSX, mas preferem o conforto de terminais modernos no Windows ou Linux.

O editor oferece uma interface baseada em caracteres (TUI) inspirada no clássico Norton Editor e nos ambientes Turbo C++/Turbo Vision da Borland, proporcionando uma experiência produtiva e nostálgica.

## Funcionalidades Principais

- **Syntax Highlighting**: Suporte para MSX-BASIC, Turbo Pascal 3, MSX-C 1.2 e SDCC 4 com MSXgl.
- **Suporte MSX-BASIC**: Capacidade de carregar e salvar programas em formato ASCII e BASIC tokenizado (.BAS).
- **Interface TUI**: Menu superior "Top Down", janela de edição central e barra de status inferior.
- **Janela de Edição Retro**: Primeira janela aberta automaticamente no startup, com moldura dupla branca, título centralizado, identificador de janela e barras de rolagem horizontal/vertical no estilo clássico.
- **Diálogo OK Reutilizável**: Componente `Dialogo OK` centralizável, com moldura dupla, botão de controle `[■]` no topo e botão principal configurável por label/hotkey/callback.
- **Botão Turbo Reutilizável**: Componente de botão verde estilo Turbo Vision, com tecla de destaque e sombra configurável (`shadowModeTurboClassic` e `shadowModeFlat`).
- **Temas VGA**: `default` (VGA Borland blue, estilo MS-DOS/Turbo) e `blue` (VGA NC-style/Norton Commander, com menu e status em ciano).
- **Multiplataforma**: Executa nativamente em consoles Windows e Linux com suporte a cores e caracteres gráficos.
- **Configuração Flexível**: Suporte a arquivos de configuração locais ou globais (%APPDATA%) e argumentos de linha de comando.

## Ferramentas Utilizadas

O projeto é desenvolvido e mantido utilizando as seguintes tecnologias:

- **Linguagem Go (Golang)**: Versão 1.26, aproveitando recursos modernos de concorrência e tipagem.
- **Tview & Tcell**: Bibliotecas robustas para criação de interfaces de terminal ricas.
- **PowerShell**: Para scripts de build e automação de tarefas.
- **Git**: Controle de versão.
- **GPL 3.0**: Licenciamento de código aberto.

## Como Começar

Consulte o arquivo [MANUAL.md](MANUAL.md) para instruções detalhadas de compilação e operação.
