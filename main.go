package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	tmpl = template.Must(template.ParseFiles("templates/index.html"))
)

type Todo struct {
	ID          int
	Title       string
	Description string
	Done        bool
}

func main() {
	var err error
	// Open the SQLite database
	db, err = sql.Open("sqlite", "./database/todo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Set up HTTP handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/export", exportHandler)
	http.HandleFunc("/import", importHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the HTTP server
	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Get the search query parameter
	search := r.URL.Query().Get("search")
	query := "SELECT id, title, description, done FROM todos WHERE title LIKE ? OR description LIKE ? ORDER BY done ASC, id DESC"
	rows, err := db.Query(query, "%"+search+"%", "%"+search+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	// Iterate over the rows and scan them into the todos slice
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Done); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	// Execute the template with the todos data
	if err := tmpl.Execute(w, todos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// Get form values
	title := r.FormValue("title")
	description := r.FormValue("description")

	// Redirect if title or description is empty
	if title == "" || description == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Insert the new todo into the database
	_, err := db.Exec("INSERT INTO todos (title, description, done) VALUES (?, ?, false)", title, description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	// Get form values
	id := r.FormValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")

	// Redirect if title or description is empty
	if title == "" || description == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Update the todo in the database
	_, err := db.Exec("UPDATE todos SET title = ?, description = ? WHERE id = ?", title, description, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	// Get the todo ID from the form
	id := r.FormValue("id")

	// Toggle the done status of the todo
	_, err := db.Exec("UPDATE todos SET done = NOT done WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Get the todo ID from the form
	id := r.FormValue("id")

	// Delete the todo from the database
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	// Query all todos from the database
	rows, err := db.Query("SELECT id, title, description, done FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create the CSV file
	file, err := os.Create("todos.csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the CSV header
	if err := writer.Write([]string{"ID", "Title", "Description", "Done"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write each todo to the CSV file
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Done); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		doneStr := "0"
		if todo.Done {
			doneStr = "1"
		}
		if err := writer.Write([]string{strconv.Itoa(todo.ID), todo.Title, todo.Description, doneStr}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Exporting task: ID=%d, Title=%s, Description=%s, Done=%s\n", todo.ID, todo.Title, todo.Description, doneStr)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the CSV file to the client
	http.ServeFile(w, r, "todos.csv")
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	// Get the uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert or update each todo in the database
	for _, record := range records[1:] {
		id, _ := strconv.Atoi(record[0])
		title := record[1]
		description := record[2]
		done, _ := strconv.ParseBool(record[3])

		if _, err := db.Exec("INSERT INTO todos (id, title, description, done) VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, description=excluded.description, done=excluded.done", id, title, description, done); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect to the index page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
