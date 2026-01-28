# NoteMe

**NoteMe** is a sleek, keyboard-centric terminal user interface (TUI) application for managing notes and todos. Built with Go and Bubble Tea, it brings the "The Elm Architecture" to your CLI.

## Features

*   **Notes:** Create rich text notes with titles and folders.
*   **Todos:** Manage tasks with recurrence (Daily, Weekly, Monthly).
*   **Dashboard:** Visual heatmap of your activity and quick stats.
*   **Keyboard First:** Vim-like navigation (`j`/`k`) and efficient shortcuts.
*   **Local Storage:** Data is safely stored in `~/.noteme/` as JSON.

## Installation

Ensure you have [Go](https://go.dev/dl/) installed (1.19+ recommended).

### Using go install (Recommended)

You can install `noteme` directly from GitHub:

```bash
go install github.com/yourusername/noteme@latest
```

This will compile the binary and place it in your `$GOPATH/bin` (usually `~/go/bin`). Ensure this directory is in your system's `PATH`.

### From Source

```bash
git clone https://github.com/yourusername/noteme.git
cd noteme
go build -o noteme main.go
mv noteme /usr/local/bin/ # Optional
```

## Usage

Run the app:

```bash
noteme
```

### Controls

| Context | Key | Action |
| :--- | :--- | :--- |
| **Global** | `Tab` | Switch Views (Dashboard -> Notes -> Todos) |
| | `q` / `Ctrl+C` | Quit |
| **Dashboard** | `n` | Create New Note |
| | `t` | Create New Todo |
| **Lists** | `j` / `k` | Navigate Up/Down |
| | `Enter` | Edit Note / Toggle Todo |
| | `d` | Delete Item |
| **Editor** | `Tab` | Switch Fields |
| | `Ctrl+S` | Save |
| | `Esc` | Cancel / Back |

## Data Location

Your data is stored in standard JSON files, making it easy to backup or edit manually if needed:

*   `~/.noteme/notes.json`
*   `~/.noteme/todos.json`

## Built With

*   [Bubble Tea](https://github.com/charmbracelet/bubbletea)
*   [Lip Gloss](https://github.com/charmbracelet/lipgloss)
*   [Bubbles](https://github.com/charmbracelet/bubbles)

## License

MIT
