package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
	"noteme/model"
)

const (
    DirName = ".noteme"
    NotesFile = "notes.json"
    TodosFile = "todos.json"
)

type Storage struct {
	basePath string
}

func NewStorage() (*Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(home, DirName)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &Storage{basePath: basePath}, nil
}

func (s *Storage) LoadNotes() ([]model.Note, error) {
	path := filepath.Join(s.basePath, NotesFile)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Seed default note
        defaultNotes := []model.Note{
            {
                ID:        "welcome-note",
                Title:     "Welcome to NoteMe!",
                Content:   "This is your first note.\n\n- Press 'Enter' to edit this note.\n- Press 'n' to create a new one.\n- Press 'd' to delete.\n\nEnjoy using NoteMe!",
                Folder:    "General",
                CreatedAt: time.Now(),
            },
        }
        if err := s.SaveNotes(defaultNotes); err != nil {
            return nil, err
        }
        return defaultNotes, nil
    }
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var notes []model.Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *Storage) SaveNotes(notes []model.Note) error {
	path := filepath.Join(s.basePath, NotesFile)
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) DeleteNote(id string) error {
    notes, err := s.LoadNotes()
    if err != nil {
        return err
    }
    newNotes := []model.Note{}
    found := false
    for _, n := range notes {
        if n.ID != id {
            newNotes = append(newNotes, n)
        } else {
            found = true
        }
    }
    
    // Optimization: if not found, don't write
    if !found {
        return nil
    }

    return s.SaveNotes(newNotes)
}

func (s *Storage) LoadTodos() ([]model.Todo, error) {
	path := filepath.Join(s.basePath, TodosFile)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Seed default todo
        defaultTodos := []model.Todo{
            {
                ID:        "welcome-todo",
                Content:   "Try creating a new todo (Press 't')",
                Done:      false,
                CreatedAt: time.Now(),
                Frequency: model.Once,
            },
        }
        if err := s.SaveTodos(defaultTodos); err != nil {
            return nil, err
        }
        return defaultTodos, nil
    }
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var todos []model.Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *Storage) SaveTodos(todos []model.Todo) error {
	path := filepath.Join(s.basePath, TodosFile)
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) DeleteTodo(id string) error {
    todos, err := s.LoadTodos()
    if err != nil {
        return err
    }
    newTodos := []model.Todo{}
    found := false
    for _, t := range todos {
        if t.ID != id {
            newTodos = append(newTodos, t)
        } else {
            found = true
        }
    }

    if !found {
        return nil
    }

    return s.SaveTodos(newTodos)
}