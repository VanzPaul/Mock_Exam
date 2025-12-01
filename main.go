package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	// Add API endpoint to serve JSON files from the json directory
	http.HandleFunc("/api/exams", serveExamFiles)

	fmt.Println("Server starting on port 9090...")
	log.Println("Application started on port 9090")

	// Start the server on port 9090
	log.Fatal(http.ListenAndServe(":9090", nil))
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
