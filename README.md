# Distraction Writer

A minimal terminal writing app in Go with Vim-style editing.

## Features

- Full-screen, distraction-free writing surface.
- Centered writing column (`72` chars wide).
- Dark, uniform background with configurable colors in one place ([`appearance.go`](appearance.go)).
- Vim-style modes with bottom-left indicator (`N`, `I`, `V`, `L`).
- Visual selection highlight and `yy` copy flash feedback.
- Undo/redo, line/word motions, and clipboard-aware yank/paste.
- Cross-platform clipboard support (macOS, Linux Wayland/X11, Windows).

## Run

```bash
go run .
```

## Vim Commands

| Mode | Command | Action |
| --- | --- | --- |
| Global | `Ctrl-c` | Quit |
| Insert | text input | Insert characters |
| Insert | `Enter` | New line |
| Insert | `Backspace` | Delete backward |
| Insert | `←/→/↑/↓` | Move cursor |
| Insert | `Esc` | Switch to Normal |
| Normal | `h j k l` or `←/→/↑/↓` | Move cursor |
| Normal | `w` | Word forward |
| Normal | `b` | Word backward |
| Normal | `0` | Start of line |
| Normal | `$` | End of line |
| Normal | `gg` | First line |
| Normal | `G` | Last line |
| Normal | `i a I A` | Enter Insert (cursor variants) |
| Normal | `o` | Open line below + Insert |
| Normal | `v` | Enter/toggle Visual (charwise) |
| Normal | `V` | Enter/toggle Visual Line |
| Normal | `x` | Delete character |
| Normal | `dd` | Delete current line |
| Normal | `D` | Delete to end of line |
| Normal | `C` | Change to end of line (delete + Insert) |
| Normal | `r<char>` | Replace character under cursor |
| Normal | `yy` | Yank current line (with flash) |
| Normal | `p` | Paste after cursor |
| Normal | `u` | Undo |
| Normal | `Ctrl-r` | Redo |
| Visual / Visual Line | `h j k l`, `w`, `b`, `0`, `$`, `gg`, `G` | Move selection |
| Visual / Visual Line | `v` / `V` | Toggle visual mode type / exit |
| Visual / Visual Line | `y` | Yank selection |
| Visual / Visual Line | `x` | Delete selection |
| Visual / Visual Line | `Esc` | Exit to Normal |
