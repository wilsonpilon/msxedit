# Help Contents

MSXEDIT HELP CONTENTS

Use `Tab` or `Shift+Tab` to move between links, `Enter` to open a topic, `Alt+F1` to go back, and `Esc` to close Help.

- [How to Use Help](#how-to-use-help)
- [Menus and Hot Keys](#menus-and-hot-keys)
- [Editor Commands](#editor-commands)
- [Built-in](#built-in)
- [Command Line](#command-line)
- [Debugging](#debugging)
- [Directives](#directives)
- [Error Messages](#error-messages)
- [Browser](#browser)
- [Reserved Words](#reserved-words)
- [Start-Up Options](#start-up-options)
- [MSXgl](#msxgl)
- [Library](#library)

## How to Use Help

The Help window in MSXEdit is a full custom TUI window with its own frame, scrollbars, topic title and window number.

Keyboard navigation:

- `Tab` / `Shift+Tab`: select next or previous link
- `Enter`: open selected link
- `Alt+F1`: return to previous topic
- `Alt+Q`: fallback back-navigation for terminals that do not forward `Alt+F1`
- `PgUp`, `PgDn`, `Home`, `End`: fast scrolling
- `Esc`: close Help

Mouse navigation:

- click a link to open it
- click `[■]` to close the Help window
- click the vertical scrollbar to scroll by position
- click the bottom scrollbar area to move horizontally

References:
- [Help Contents](#contents)
- [Menus and Hot Keys](#menus-and-hot-keys)

## Menus and Hot Keys

The top menu bar follows the classic Turbo style and can be opened by keyboard or mouse.

Implemented hot keys:

- `F1`: open Help
- `F3`: open the `Open File` dialog
- `Alt+F3`: close the active editor window
- `Ctrl+F1`: "Language help" — open Help already navigated to `Reserved Words`
- `Ctrl+L`: repeat the last search (`Search again`)
- `F10`: open the `File` menu
- `Alt+X`: exit MSXEdit
- `Alt+F`, `Alt+E`, `Alt+S`, `Alt+R`, `Alt+C`, `Alt+D`, `Alt+T`, `Alt+O`, `Alt+W`, `Alt+H`: open the matching menu

Current menu status:

- `File`: active menu, with `New` (cascading editor window), `Open…` and `Exit` implemented
- `Edit`: active menu, with `Undo`, `Redo`, `Cut`, `Copy`, `Paste`, `Clear` and `Show clipboard`
  implemented — clipboard is now shared across all editor windows
- `Search`: active menu, with `Find...`, `Replace...`, `Search again` and `Go to line number...`
  implemented (`Replace...` still does not perform the actual substitution); `Show last compiler
  error`, `Find error...` and `Find procedure...` remain placeholders
- `Help`: active menu, with `Contents` and `About` implemented
- `Run`, `Compile`, `Debug`, `Tools`, `Window`: visible and navigable, still scaffolded with placeholder content
- `Options`: now opens `Compiler/Interpreter Options` with radio buttons, checkboxes, `Conditional defines:`, and `OK`/`Cancel`/`Help`

The status bar also displays legacy labels for `F2 Save`, `F3 Open`, `Alt+F9 Compile` and `F9 Make`, but those workflows are not yet wired to concrete actions.

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)
- [Command Line](#command-line)

## Editor Commands

This topic groups the current editing-related areas.

- [Block commands](#block-commands)
- [Cursor-movement commands](#cursor-movement-commands)
- [Insert & Delete commands](#insert-delete-commands)
- [Miscelaneous commands](#miscelaneous-commands)
- [Syntax highlighting](#syntax-highlighting)

References:
- [Help Contents](#contents)
- [How to Use Help](#how-to-use-help)

## Built-in

Built-in components already used by the application:

- custom editor window with double border, supporting multiple simultaneous windows
- `dialogoOK` reusable dialog
- `turboButton` reusable Turbo-style button
- `compilerOptionsDialog` for `Compiler/Interpreter Options`
- `findDialog`, `replaceDialog` and `gotoLineDialog` for the `Search` menu
- shared clipboard window (`Show clipboard`)
- `About` dialog
- `Help` window with markdown topic loading and fallback internal content

References:
- [Help Contents](#contents)
- [Library](#library)

## Command Line

MSXEdit currently supports the following command line options:

- `--theme <default|blue>`
- `--tabsize <n>`
- `--no-highlight`
- `--local`
- `--version`

You can also pass an optional file name as the last argument. When present, that path is used as the title of the first editor window.

References:
- [Help Contents](#contents)
- [Start-Up Options](#start-up-options)

## Debugging

The current focus of the project is UI structure and navigation. Debugging tools such as source-level stepping, watches and runtime inspectors are not yet available from the `Debug` menu.

At this stage, the most visible diagnostics are:

- CLI usage and version output
- config loading warnings in terminal mode
- markdown Help fallback when `HELP.md` cannot be read

References:
- [Help Contents](#contents)
- [Error Messages](#error-messages)

## Directives

Compiler or project directives are not yet interpreted by the editor UI. This topic is reserved for future language-aware behavior.

At the moment, the nearest related pieces are the planned syntax/highlight pipeline and the language-specific roadmap tracked in the project documentation.

References:
- [Help Contents](#contents)
- [Syntax highlighting](#syntax-highlighting)

## Error Messages

Known runtime/user-visible situations include:

- configuration file cannot be decoded
- configuration file does not exist and defaults are used
- `HELP.md` cannot be found, so built-in topics are loaded instead
- tokenized BASIC loader is still not implemented

References:
- [Help Contents](#contents)
- [Debugging](#debugging)

## Browser

There is no full file browser dialog yet.

What already exists:

- editor startup with optional filename argument
- menu scaffolding for future file actions
- mouse-aware top menu interaction

What is still pending:

- `Open...`
- directory changes from the UI
- save dialogs and file browsing panels

References:
- [Help Contents](#contents)
- [Menus and Hot Keys](#menus-and-hot-keys)

## Reserved Words

Language keyword databases are still in progress.

The repository already contains an initial MSX-BASIC token table that will support future language-aware features, but no keyword browser is exposed in the UI yet.

References:
- [Help Contents](#contents)
- [MSXgl](#msxgl)

## Start-Up Options

Startup behavior summary:

- no argument: open editor window `1` with title `Sem Nome`
- file argument: open editor window `1` using the provided path as title
- `--theme`: choose the VGA palette profile
- `--local`: load `msxedit.json` from the current directory

The application always opens the main editor window automatically.

References:
- [Help Contents](#contents)
- [Command Line](#command-line)

## MSXgl

MSXgl remains part of the long-term language support roadmap. The current release does not yet provide dedicated editing assistance, compile integration or contextual help for MSXgl symbols.

References:
- [Help Contents](#contents)
- [Library](#library)

## Library

Current reusable UI building blocks include:

- `dialogoOK`
- `turboButton`
- `helpHeaderButton`
- theme palette helpers
- markdown Help loader and parser

These pieces are intended to keep future dialogs and tool windows visually consistent.

References:
- [Help Contents](#contents)
- [Built-in](#built-in)

## Block commands

  Block Commands

  Ctrl+K B    Mark block begin
  Ctrl+K K    Mark block end
  Ctrl+K T    Mark single word
  Ctrl+K L    Mark line
  Ctrl+K C    Copy block
  Ctrl+K V    Move block
  Ctrl+K Y    Delete block
  Ctrl+K R    Read block from disk
  Ctrl+K W    Write block to disk
  Ctrl+K H    Hide/display block
  Ctrl+K P    Print block
  Ctrl+K I    Indent block
  Ctrl+K U    Unindent block
  Ctrl+K D    Exit to menu bar
  Ctrl+Q B    Move to begin of block
  Ctrl+Q K    Move to end of block
  Ctrl+Ins    Copy to clipboard
  Shift+Del   Cut to clipboard
  Ctrl+Del    Delete block
  Shift+Ins   Paste from clipboard

  See Also:
    [Extending Selected Blocks](#extending-selected-blocks)

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)

## Cursor-movement commands

  Cursor Movement Commands

  Character left      Ctrl+S or Left arrow
  Character right     Ctrl+D or Right arrow
  Word left           Ctrl+A or Ctrl+Left arrow
  Word right          Ctrl+F or Ctrl+Right arrow
  Line up             Ctrl+E or Up arrow
  Line down           Ctrl+X or Down arrow
  Scroll up           Ctrl+W
  Scroll down         Ctrl+Z
  Page up             Ctrl+R or PgUp
  Page down           Ctrl+C or PgDn

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)

## Insert & Delete commands

  Insert & Delete Commands

  Insert mode on/off       Ctrl+V or Ins
  Insert line              Ctrl+N
  Delete line               Ctrl+Y
  Delete to end of line    Ctrl+Q Y
  Delete character left    Ctrl+H or Backspace
  Delete character         Ctrl+G or Del
  Delete word right         Ctrl+T

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)

## Miscelaneous commands

  Miscellaneous Editor Commands

  Menu bar                    F10
  Save and edit                Ctrl+K S or F2
  Open file                   F3
  Close active window          Alt+F3

  Tab                          Ctrl+I or Tab
  Tab mode                     Ctrl+O T
  Auto indent on/off            Ctrl+O I
  Restore line                 Ctrl+Q L

  Set place marker(0-9)         Ctrl+K n (n = 0..9)
  Find plc. marker(0-9)         Ctrl+Q n (n = 0..9)
  Language help                Ctrl+F1
  Ctrl+character prefix         Ctrl+P

  Find                         Ctrl+Q F
  Find & replace                Ctrl+Q A
  Repeat last find              Ctrl+L
  Abort operation               Esc

References:
- [Help Contents](#contents)
- [Editor Commands](#editor-commands)

## Syntax highlighting

The project already exposes the configuration flag `highlight` and the CLI switch `--no-highlight`, but the actual syntax coloring pipeline is not fully connected to the editor surface yet.

This topic is therefore a status note rather than a completed feature description.

References:
- [Editor Commands](#editor-commands)
- [Directives](#directives)

## Extending Selected Blocks

  EXTENDING SELECTED BLOCKS

  [Content to be implemented]

References:
- [Help Contents](#contents)
- [Block commands](#block-commands)

