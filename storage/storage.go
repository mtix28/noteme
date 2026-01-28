package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
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
        return []model.Note{}, nil
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
        return []model.Todo{}, nil
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