package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const serverAddr string = "127.0.0.1:8081"

type Note struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	notes = make(map[string]Note)
	mutex = &sync.Mutex{}
)

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func createNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var note Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, "failed to unmarshal", http.StatusBadRequest)
		return
	}

	note.ID = generateID()
	note.CreatedAt = time.Now()

	mutex.Lock()
	notes[note.ID] = note
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"id": note.ID})
	if err != nil {
		http.Error(w, "failed to encode", http.StatusInternalServerError)
	}
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	var updatedNote Note
	err := json.NewDecoder(r.Body).Decode(&updatedNote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	note, exists := notes[id]
	if !exists {
		mutex.Unlock()
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	note.Name = updatedNote.Name
	note.Text = updatedNote.Text
	notes[id] = note
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(note)
	if err != nil {
		http.Error(w, "failed to encode", http.StatusInternalServerError)
	}
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "path parameter not found", http.StatusNotFound)
		return
	}

	mutex.Lock()
	_, exists := notes[id]
	if !exists {
		mutex.Unlock()
		http.Error(w, "note does not exists", http.StatusNotFound)
		return
	}
	delete(notes, id)
	mutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func getNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "path parameter not found", http.StatusNotFound)
		return
	}

	mutex.Lock()
	note, exists := notes[id]
	if !exists {
		mutex.Unlock()
		http.Error(w, "note does not exists", http.StatusNotFound)
		return
	}
	mutex.Unlock()

	w.Header().Set("Content-type", "application/json")
	err := json.NewEncoder(w).Encode(note)
	if err != nil {
		http.Error(w, "failed to encode", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/note", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createNote(w, r)
		case http.MethodPut:
			updateNote(w, r)
		case http.MethodGet:
			getNote(w, r)
		case http.MethodDelete:
			deleteNote(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	slog.Info("starting server", "address", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		slog.Info("server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server")
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

	slog.Info("shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server properly", "error", err)
	}

	slog.Info("server stopped")
}
