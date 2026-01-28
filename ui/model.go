package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/mtix28/noteme/model"
	"github.com/mtix28/noteme/storage"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type sessionState int

const (
	DashboardView sessionState = iota
	NoteListView
	NoteEditView
	TodoListView
	TodoAddView
    DeleteConfirmView
)

type MainModel struct {
	state  sessionState
	store  *storage.Storage
	width  int
	height int

	// Data
	notes []model.Note
	todos []model.Todo

	// Components
	noteList list.Model
	todoList list.Model
    keys     KeyMap
    help     help.Model

	// Editor components
	noteTitleInput   textinput.Model
	noteFolderInput  textinput.Model
	noteContentInput textarea.Model
	currentNoteID    string

	// Todo Input
	todoInput textinput.Model
    
    // Deletion State
    itemToDeleteID   string
    itemToDeleteType string // "note" or "todo"
}

func NewModel() (MainModel, error) {
	store, err := storage.NewStorage()
	if err != nil {
		return MainModel{}, err
	}

	// Notes List
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Notes"
    l.SetShowHelp(false) // We'll use our own help
    l.DisableQuitKeybindings()

	// Todo List
	tl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tl.Title = "Todos"
    tl.SetShowHelp(false)
    tl.DisableQuitKeybindings()

	// Editor
	ti := textinput.New()
	ti.Placeholder = "Note Title"
	ti.Focus()

	fi := textinput.New()
	fi.Placeholder = "Folder (e.g. daily, work)"

	ta := textarea.New()
	ta.Placeholder = "Start typing your note..."

	// Todo Input
	tdi := textinput.New()
	tdi.Placeholder = "New Todo... (ends with /daily, /weekly for frequency)"

	return MainModel{
		state:            DashboardView,
		store:            store,
		noteList:         l,
		todoList:         tl,
        keys:             NewKeyMap(),
        help:             help.New(),
		noteTitleInput:   ti,
		noteFolderInput:  fi,
		noteContentInput: ta,
		todoInput:        tdi,
	}, nil
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadNotesCmd,
		m.loadTodosCmd,
	)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
        // Always allow quitting
        if key.Matches(msg, m.keys.Quit) {
            return m, tea.Quit
        }

		// State specific handling
		switch m.state {
        case DashboardView:
            switch {
            case key.Matches(msg, m.keys.Tab):
                m.state = NoteListView
            case key.Matches(msg, m.keys.NewNote): // Uppercase N
                return m.startNewNote()
            case key.Matches(msg, m.keys.NewTodo): // Uppercase T
                return m.startNewTodo()
            case msg.String() == "n": // Lowercase n for convenience if they just want a note
                 return m.startNewNote()
            case msg.String() == "t": // Lowercase t
                 return m.startNewTodo()
            }

		case NoteListView:
			switch {
            case key.Matches(msg, m.keys.Tab):
                m.state = TodoListView
			case key.Matches(msg, m.keys.New):
                return m.startNewNote()
            case key.Matches(msg, m.keys.Delete):
                if m.noteList.SelectedItem() != nil {
                    item := m.noteList.SelectedItem().(noteItem)
                    m.itemToDeleteID = item.note.ID
                    m.itemToDeleteType = "note"
                    m.state = DeleteConfirmView
                    return m, nil
                }
            case key.Matches(msg, m.keys.Enter):
				if m.noteList.SelectedItem() != nil {
					item := m.noteList.SelectedItem().(noteItem)
					m.state = NoteEditView
					m.currentNoteID = item.note.ID
					m.noteTitleInput.SetValue(item.note.Title)
					m.noteFolderInput.SetValue(item.note.Folder)
					m.noteContentInput.SetValue(item.note.Content)
					m.noteTitleInput.Focus()
				}
            }

		case NoteEditView:
			switch {
			case key.Matches(msg, m.keys.Back):
				m.state = NoteListView
			case key.Matches(msg, m.keys.Save):
				return m, m.saveNoteCmd()
			case key.Matches(msg, m.keys.Tab):
				if m.noteTitleInput.Focused() {
					m.noteTitleInput.Blur()
					m.noteFolderInput.Focus()
				} else if m.noteFolderInput.Focused() {
					m.noteFolderInput.Blur()
					m.noteContentInput.Focus()
				} else {
					m.noteContentInput.Blur()
					m.noteTitleInput.Focus()
				}
            }

		case TodoListView:
			switch {
            case key.Matches(msg, m.keys.Tab):
                m.state = DashboardView
			case key.Matches(msg, m.keys.New):
                return m.startNewTodo()
            case key.Matches(msg, m.keys.Delete):
                if m.todoList.SelectedItem() != nil {
                     // Get the actual todo item - since we filter or sort, using index might be unsafe if we add filters later,
                     // but currently m.todoList items match m.todos unless filtered. 
                     // Ideally we should use the SelectedItem()
                     item, ok := m.todoList.SelectedItem().(todoItem)
                     if ok {
                         m.itemToDeleteID = item.todo.ID
                         m.itemToDeleteType = "todo"
                         m.state = DeleteConfirmView
                         return m, nil
                     }
                }
            case key.Matches(msg, m.keys.Toggle), key.Matches(msg, m.keys.Enter):
				if m.todoList.SelectedItem() != nil {
					idx := m.todoList.Index()
					if idx >= 0 && idx < len(m.todos) {
						m.todos[idx].Done = !m.todos[idx].Done
						return m, m.saveTodosCmd()
					}
				}
            }

		case TodoAddView:
			switch {
			case key.Matches(msg, m.keys.Back):
				m.state = TodoListView
			case key.Matches(msg, m.keys.Enter):
				text := m.todoInput.Value()
				if text != "" {
					freq := model.Once
					content := text
					if strings.HasSuffix(text, "/daily") {
						freq = model.Daily
						content = strings.TrimSpace(strings.TrimSuffix(text, "/daily"))
					} else if strings.HasSuffix(text, "/weekly") {
						freq = model.Weekly
						content = strings.TrimSpace(strings.TrimSuffix(text, "/weekly"))
					} else if strings.HasSuffix(text, "/monthly") {
						freq = model.Monthly
						content = strings.TrimSpace(strings.TrimSuffix(text, "/monthly"))
					}

					newTodo := model.Todo{
						ID:        uuid.New().String(),
						Content:   content,
						Done:      false,
						CreatedAt: time.Now(),
						Frequency: freq,
					}
					m.todos = append([]model.Todo{newTodo}, m.todos...)
					m.state = TodoListView
					return m, m.saveTodosCmd()
				}
			}

		case DeleteConfirmView:
            switch {
            case key.Matches(msg, m.keys.Enter) || msg.String() == "y":
                return m, m.deleteItemCmd()
            case key.Matches(msg, m.keys.Back) || msg.String() == "n":
                 // Return to previous view
                 if m.itemToDeleteType == "note" {
                     m.state = NoteListView
                 } else {
                     m.state = TodoListView
                 }
                 return m, nil
            }
        }

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
        m.help.Width = msg.Width
		
        // Calculate dynamic sizes
        h, v := appStyle.GetFrameSize()
        availableWidth := msg.Width - h
        availableHeight := msg.Height - v
        
        m.noteList.SetSize(availableWidth, availableHeight - 3) // leave room for help
		m.todoList.SetSize(availableWidth, availableHeight - 3)
		m.noteContentInput.SetWidth(availableWidth)
		m.noteContentInput.SetHeight(availableHeight - 10)

	case notesLoadedMsg:
		m.notes = msg.notes
		m.updateNoteListItems()

	case todosLoadedMsg:
		m.todos = msg.todos
		m.updateTodoListItems()

	case noteSavedMsg:
		return m, m.loadNotesCmd

	case todosSavedMsg:
		return m, m.loadTodosCmd
        
    case itemDeletedMsg:
        // Reload everything to be safe
        if m.itemToDeleteType == "note" {
            m.state = NoteListView
            return m, m.loadNotesCmd
        } else {
            m.state = TodoListView
            return m, m.loadTodosCmd
        }
	}

	// Update components
	switch m.state {
	case NoteListView:
		m.noteList, cmd = m.noteList.Update(msg)
		cmds = append(cmds, cmd)
	case TodoListView:
		m.todoList, cmd = m.todoList.Update(msg)
		cmds = append(cmds, cmd)
	case NoteEditView:
		m.noteTitleInput, cmd = m.noteTitleInput.Update(msg)
		cmds = append(cmds, cmd)
		m.noteFolderInput, cmd = m.noteFolderInput.Update(msg)
		cmds = append(cmds, cmd)
		m.noteContentInput, cmd = m.noteContentInput.Update(msg)
		cmds = append(cmds, cmd)
	case TodoAddView:
		m.todoInput, cmd = m.todoInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
    var content string
    var helpKeys []key.Binding

	switch m.state {
	case DashboardView:
        content = m.renderDashboard()
        helpKeys = []key.Binding{m.keys.Tab, m.keys.NewNote, m.keys.NewTodo, m.keys.Quit}

	case NoteListView:
		content = m.noteList.View()
        helpKeys = []key.Binding{m.keys.Tab, m.keys.New, m.keys.Enter, m.keys.Delete, m.keys.Up, m.keys.Down, m.keys.Quit}

	case TodoListView:
		content = m.todoList.View()
        helpKeys = []key.Binding{m.keys.Tab, m.keys.New, m.keys.Toggle, m.keys.Delete, m.keys.Up, m.keys.Down, m.keys.Quit}

	case NoteEditView:
        content = lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Edit Note"),
				"Title:",
				m.noteTitleInput.View(),
				"Folder:",
				m.noteFolderInput.View(),
				"Content:",
				m.noteContentInput.View(),
		)
        helpKeys = []key.Binding{m.keys.Tab, m.keys.Save, m.keys.Back}

	case TodoAddView:
		content = lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("New Todo"),
				"Description (append /daily, /weekly, etc.):",
				m.todoInput.View(),
		)
        helpKeys = []key.Binding{m.keys.Enter, m.keys.Back}
    
    case DeleteConfirmView:
        content = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("196")). // Red
            Padding(1, 2).
            Render(fmt.Sprintf(
                "Are you sure you want to delete this %s?\n\n(y/Enter) Yes    (n/Esc) No",
                m.itemToDeleteType,
            ))
        // Center it roughly (simple way)
        content = lipgloss.Place(m.width, m.height-5, lipgloss.Center, lipgloss.Center, content)
        helpKeys = []key.Binding{m.keys.Enter, m.keys.Back}
	}
    
    // Combine content and help
    helpView := m.help.ShortHelpView(helpKeys)
    return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content, "\n", helpView))
}

func (m MainModel) renderDashboard() string {
    noteCount := len(m.notes)
    todoCount := len(m.todos)
    doneCount := 0
    for _, t := range m.todos {
        if t.Done {
            doneCount++
        }
    }

    // Header
    header := headerStyle.Width(m.width - 6).Render(
        lipgloss.JoinHorizontal(lipgloss.Center,
            titleStyle.Render(fmt.Sprintf("Good %s, User!", timeOfDay())),
            lipgloss.NewStyle().MarginLeft(4).Render(time.Now().Format("Monday, Jan 02")),
        ),
    )

    // Empty State
    if noteCount == 0 && todoCount == 0 {
        return lipgloss.JoinVertical(lipgloss.Center,
            header,
            "\n\n\n",
            emptyStateStyle.Render("No notes or todos yet.\nPress 'n' to create a note or 't' for a todo to get started!"),
        )
    }

    // Status / Stats
    stats := fmt.Sprintf(
        "%s %s    %s %s    %s %s",
        statLabel.Render("Notes:"), statValue.Render(fmt.Sprintf("%d", noteCount)),
        statLabel.Render("Active Todos:"), statValue.Render(fmt.Sprintf("%d", todoCount-doneCount)),
        statLabel.Render("Done:"), statValue.Render(fmt.Sprintf("%d", doneCount)),
    )
    
    statusSection := cardStyle.Width(m.width - 6).Render(
        lipgloss.JoinVertical(lipgloss.Center,
            lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Status Overview"),
            "\n",
            stats,
        ),
    )
    
    // Navigation Hint
    navHint := lipgloss.NewStyle().
        Foreground(textColor).
        Background(subtleColor).
        Padding(0, 1).
        MarginTop(1).
        Render("PRESS [TAB] TO SWITCH LISTS")

    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        statusSection,
        lipgloss.PlaceHorizontal(m.width - 6, lipgloss.Center, navHint),
    )
}

func timeOfDay() string {
    h := time.Now().Hour()
    if h < 12 { return "Morning" }
    if h < 18 { return "Afternoon" }
    return "Evening"
}

// Actions

func (m MainModel) startNewNote() (tea.Model, tea.Cmd) {
    m.state = NoteEditView
    m.currentNoteID = ""
    m.noteTitleInput.SetValue("")
    m.noteFolderInput.SetValue("general")
    m.noteContentInput.SetValue("")
    m.noteTitleInput.Focus()
    return m, nil
}

func (m MainModel) startNewTodo() (tea.Model, tea.Cmd) {
    m.state = TodoAddView
    m.todoInput.SetValue("")
    m.todoInput.Focus()
    return m, nil
}


// Helpers

func (m *MainModel) updateNoteListItems() {
	items := make([]list.Item, len(m.notes))
	for i, n := range m.notes {
		items[i] = noteItem{n}
	}
	m.noteList.SetItems(items)
}

func (m *MainModel) updateTodoListItems() {
	items := make([]list.Item, len(m.todos))
	for i, t := range m.todos {
		items[i] = todoItem{t}
	}
	m.todoList.SetItems(items)
}

// Commands & Messages

type notesLoadedMsg struct{ notes []model.Note }
type todosLoadedMsg struct{ todos []model.Todo }
type noteSavedMsg struct{}
type todosSavedMsg struct{}
type itemDeletedMsg struct{}

func (m MainModel) loadNotesCmd() tea.Msg {
	notes, _ := m.store.LoadNotes()
	return notesLoadedMsg{notes}
}

func (m MainModel) loadTodosCmd() tea.Msg {
	todos, _ := m.store.LoadTodos()
	return todosLoadedMsg{todos}
}

func (m MainModel) saveNoteCmd() tea.Cmd {
	return func() tea.Msg {
		// Construct note
		note := model.Note{
			ID:        m.currentNoteID,
			Title:     m.noteTitleInput.Value(),
			Content:   m.noteContentInput.Value(),
			CreatedAt: time.Now(),
			Folder:    m.noteFolderInput.Value(),
		}

		if note.ID == "" {
			note.ID = uuid.New().String()
			m.notes = append([]model.Note{note}, m.notes...)
		} else {
			// Update existing
			for i, n := range m.notes {
				if n.ID == note.ID {
					note.CreatedAt = n.CreatedAt // Keep original creation time
					m.notes[i] = note
					break
				}
			}
		}

		m.store.SaveNotes(m.notes)
		return noteSavedMsg{}
	}
}

func (m MainModel) saveTodosCmd() tea.Cmd {
	return func() tea.Msg {
		m.store.SaveTodos(m.todos)
		return todosSavedMsg{}
	}
}

func (m MainModel) deleteItemCmd() tea.Cmd {
    return func() tea.Msg {
        if m.itemToDeleteType == "note" {
            m.store.DeleteNote(m.itemToDeleteID)
        } else {
            m.store.DeleteTodo(m.itemToDeleteID)
        }
        return itemDeletedMsg{}
    }
}

// List Items Adapters
type noteItem struct{ note model.Note }

func (n noteItem) FilterValue() string { return n.note.Title }
func (n noteItem) Title() string       { return n.note.Title }
func (n noteItem) Description() string {
	return fmt.Sprintf("[%s] %s", n.note.Folder, n.note.CreatedAt.Format("2006-01-02"))
}

type todoItem struct{ todo model.Todo }

func (t todoItem) FilterValue() string { return t.todo.Content }
func (t todoItem) Title() string {
	prefix := "[ ] "
	if t.todo.Done {
		prefix = "[x] "
	}
	return prefix + t.todo.Content
}
func (t todoItem) Description() string {
	return string(t.todo.Frequency) + " | " + t.todo.CreatedAt.Format("2006-01-02")
}
