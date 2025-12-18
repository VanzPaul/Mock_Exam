package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcozac/go-jsonc"
)

// ExamFile represents a JSON file with its name and content
type ExamFile struct {
	Name    string      `json:"name"`
	Content interface{} `json:"content"`
}

// Subject represents a subject with its name and associated exams
type Subject struct {
	Name  string     `json:"name"`
	Exams []ExamFile `json:"exams"`
}

func main() {
	// Serve static files from the current directory
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Add API endpoint to serve JSON files from the json directory with gzip compression
	http.HandleFunc("/api/exams", gzipMiddleware(serveExamFiles))

	fmt.Printf("Server starting on port %s...\n", port)
	log.Printf("Application started on port %s", port)

	// Start the server on the specified port
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// serveExamFiles reads all JSON files from the "json" directory organized by subjects and returns subjects with their exams
func serveExamFiles(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Read all files from the json directory organized by subjects
	subjects, err := readExamFiles()
	if err != nil {
		http.Error(w, "Failed to read exam files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(subjects); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// readExamFiles reads all JSON files from the "json" directory organized by subjects and returns subjects with their exams
func readExamFiles() ([]Subject, error) {
	subjectsMap := make(map[string][]ExamFile)

	// Read files from the json directory
	err := filepath.Walk("json", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a file and has a .json or .jsonc extension
		ext := filepath.Ext(path)
		if !info.IsDir() && (ext == ".json" || ext == ".jsonc") {
			// Extract subject name from the directory path
			dir := filepath.Dir(path)
			subjectName := filepath.Base(dir)

			// Read the file content
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			// Skip empty files
			if len(content) == 0 {
				fmt.Printf("Warning: Skipping empty file %s\n", path)
				return nil // This continues with other files in filepath.Walk
			}

			// Parse JSON content to interface{}
			var parsedContent interface{}
			if ext == ".jsonc" {
				// Use jsonc package for JSONC files
				err = jsonc.Unmarshal(content, &parsedContent)
				if err != nil {
					return fmt.Errorf("failed to parse JSONC in file %s: %w", path, err)
				}
			} else {
				// Use standard json package for regular JSON files
				if err := json.Unmarshal(content, &parsedContent); err != nil {
					return fmt.Errorf("failed to parse JSON in file %s: %w", path, err)
				}
			}

			// Add to the appropriate subject's exams
			examFile := ExamFile{
				Name:    info.Name(),
				Content: parsedContent,
			}
			subjectsMap[subjectName] = append(subjectsMap[subjectName], examFile)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice of subjects
	var subjects []Subject
	for subjectName, exams := range subjectsMap {
		subject := Subject{
			Name:  subjectName,
			Exams: exams,
		}
		subjects = append(subjects, subject)
	}

	return subjects, nil
}

// gzipMiddleware wraps an HTTP handler to add gzip compression support
func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Client supports gzip, so we'll compress the response
			w.Header().Set("Content-Encoding", "gzip")

			// Create a gzip writer
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Wrap the response writer to use gzip
			gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
			next(gzw, r)
		} else {
			// Client doesn't support gzip, serve normally
			next(w, r)
		}
	}
}

// gzipResponseWriter wraps the HTTP response writer to support gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// WriteHeader sends a protocol-specific header prior to sending the body
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Del("Content-Length") // Remove Content-Length header as it's invalid for gzipped content
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes the data to the connection as part of an HTTP reply
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
