# Distraction Writer

A minimal terminal writing app in Go with Vim-style editing.

## Features

- Full-screen, distraction-free writing surface.
- Centered writing column (`72` chars wide).
- Dark, uniform background with configurable colors in one place ([`appearance.go`](appearance.go)).
- Vim-style modes with bottom-left indicator (`N`, `I`, `V`, `L`).
- Live word counter on the bottom-right (`<count>w`) using the same style color as the mode indicator.
- Live Markdown rendering via Glamour for non-active lines.
- Top-of-file frontmatter support (`---` YAML / `+++` TOML) with syntax-aware rendering.
- Vim-style command line in Normal mode: `:` opens a centered one-line prompt (`>`), `Esc` closes it.
- Save support via command line: `:w` saves current file, `:w <path>` saves to a new `.md`/`.txt` file.
- Active line is always raw in Insert/Normal mode.
- In Visual/Visual Line mode, all selected lines render raw.
- Markdown tables render in both edit and preview flows; selecting any table line makes the whole table raw.
- Fenced code blocks render with Glamour syntax highlighting on non-raw lines.
- Headings keep distinct per-level colors in both raw and rendered states.
- Links render as label-only in preview, while raw syntax stays visible (with label/url color accents).
- Visual selection highlight and `yy` copy flash feedback.
- Undo/redo, line/word motions, and clipboard-aware yank/paste.
- Cross-platform clipboard support (macOS, Linux Wayland/X11, Windows).

## Run

```bash
go run .
go run . notes.md
go run . /path/to/notes.txt
```

Startup file loading accepts a single `.md` or `.txt` file path.

On startup the editor opens in **Normal mode** (`N`).
If no file is passed, it starts with an empty buffer.
Only one file can be opened at startup (directories and other extensions are rejected).

## Vim Commands

| Mode                 | Command                       | Action                                              |
| -------------------- | ----------------------------- | --------------------------------------------------- |
| Global               | `Ctrl-c`                      | Quit                                                |
| Insert               | text input                    | Insert characters                                   |
| Insert               | `Enter`                       | New line                                            |
| Insert               | `Backspace`                   | Delete backward                                     |
| Insert               | `←/→/↑/↓`                     | Move cursor                                         |
| Insert               | `Esc`                         | Switch to Normal (cursor shifts left when possible) |
| Normal               | `h j k l` or `←/→/↑/↓`        | Move cursor                                         |
| Normal               | `w`                           | Word forward                                        |
| Normal               | `b`                           | Word backward                                       |
| Normal               | `0`                           | Start of line                                       |
| Normal               | `$`                           | End of line                                         |
| Normal               | `gg`                          | First line                                          |
| Normal               | `G`                           | Last line                                           |
| Normal               | `i a I A`                     | Enter Insert (cursor variants)                      |
| Normal               | `o`                           | Open line below + Insert                            |
| Normal               | `v`                           | Enter/toggle Visual (charwise)                      |
| Normal               | `V`                           | Enter/toggle Visual Line                            |
| Normal               | `x`                           | Delete character                                    |
| Normal               | `dd`                          | Delete current line                                 |
| Normal               | `D`                           | Delete to end of line                               |
| Normal               | `C`                           | Change to end of line (delete + Insert)             |
| Normal               | `r<char>`                     | Replace character under cursor                      |
| Normal               | `yy`                          | Yank current line (with flash)                      |
| Normal               | `p`                           | Paste after cursor                                  |
| Normal               | `u`                           | Undo                                                |
| Normal               | `Ctrl-r`                      | Redo                                                |
| Normal               | `:`                           | Open command line                                   |
| Command line         | text input                    | Enter command text                                  |
| Command line         | `Backspace`                   | Delete backward                                     |
| Command line         | `←/→`                         | Move command cursor                                 |
| Command line         | `Enter`                       | Execute command                                     |
| Command line         | `Esc`                         | Close command line                                  |
| Visual / Visual Line | `h j k l` or `←/→/↑/↓`        | Move selection                                      |
| Visual / Visual Line | `w`, `b`, `0`, `$`, `gg`, `G` | Selection motions                                   |
| Visual / Visual Line | `v` / `V`                     | Toggle visual mode type / exit                      |
| Visual / Visual Line | `y`                           | Yank selection                                      |
| Visual / Visual Line | `x`                           | Delete selection                                    |
| Visual / Visual Line | `Esc`                         | Exit to Normal                                      |

## Command-Line Commands

Supported `:` commands:

- `:q`
- `:q!`
- `:quit`
- `:qa`
- `:qa!`
- `:w` (save to current file name)
- `:w <path>` (save to provided `.md`/`.txt` path and set it as current file)

Save notes:

- `:w` on an unnamed buffer shows an error until you provide a file name with `:w <path>`.
- Save errors are shown in red in the command bar.
