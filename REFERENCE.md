# Referência de Opções - MSXEdit

Este arquivo contém um resumo das opções de linha de comando e chaves de configuração suportadas pelo MSXEdit.

## Opções de Linha de Comando

| Opção | Descrição |
|-------|-----------|
| `--help` | Exibe a mensagem de ajuda com todas as opções disponíveis. |
| `--version` | Exibe a versão corrente do programa (atualmente 4.0.1). |
| `--local` | Força o uso do arquivo `msxedit.json` no diretório atual em vez do diretório global. |
| `--theme <nome>` | Define o tema de cores da interface (ex: `default`, `blue`). |
| `--tabsize <n>` | Define o tamanho do caractere Tab (ex: 4 ou 8). |
| `--no-highlight` | Desativa o realce de sintaxe (Syntax Highlighting). |

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
- **show_line_numbers**: Boolean. Se verdadeiro, exibe números de linha na margem esquerda.
- **highlight**: Boolean. Se verdadeiro, habilita syntax highlight para linguagens suportadas.

## Temas de Cores Disponíveis

- **`default`**: VGA Borland blue (estilo MS-DOS/Turbo).
- **`blue`**: VGA NC-style (Norton Commander), com barra superior e status em ciano.

## Linguagens Suportadas para Syntax Highlight

- **MSX-BASIC** (.BAS)
- **Turbo Pascal 3** (.PAS)
- **MSX-C 1.2** (.C, .H)
- **SDCC 4 (MSXgl)** (.C, .H)
