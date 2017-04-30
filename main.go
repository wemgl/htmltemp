package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var templates = make(map[string]*template.Template)

type Note struct {
	Title       string
	Description string
	CreatedOn   time.Time
}

// ViewModel for editNote
type EditNote struct {
	Note
	Id string
}

// Store for the Notes collection
var noteStore = make(map[string]Note)

// Variable to generate key for the collection
var id int = 0

// Compile view templates
func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	templates["index"] = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
	templates["add"] = template.Must(template.ParseFiles("templates/add.html", "templates/base.html"))
	templates["edit"] = template.Must(template.ParseFiles("templates/edit.html", "templates/base.html"))

	// Seed the noteStore with a default note
	noteStore = map[string]Note{
		"0": {"text/template", "Template generates textual output", time.Now()},
	}
}

// Render templates for the given name, template definition, and data object
func renderTemplate(w http.ResponseWriter, name string, template string, viewModel interface{}) {
	// Ensure the template exists in the map
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "The template does not exist.", http.StatusInternalServerError)
	}
	err := tmpl.ExecuteTemplate(w, template, viewModel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", "base", noteStore)
}

func addNote(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "add", "base", nil)
}

// Handler for "/notes/ave" for saving a new item to the data store
func saveNote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.PostFormValue("title")
	desc := r.PostFormValue("description")
	note := Note{title, desc, time.Now()}
	// increment the value of id for generating key for the map
	id++
	// convert id value to string
	k := strconv.Itoa(id)
	noteStore[k] = note
	http.Redirect(w, r, "/", http.StatusFound)
}

func editNote(w http.ResponseWriter, r *http.Request) {
	var viewModel EditNote
	// Read value from route variable
	vars := mux.Vars(r)
	k := vars["id"]

	if note, ok := noteStore[k]; ok {
		viewModel = EditNote{note, k}
	} else {
		http.Error(w, "Could not find the resource to edit.", http.StatusBadRequest)
	}

	renderTemplate(w, "edit", "base", viewModel)

}

// Handler for "/notes/update/{id}" which updates an item into the datastore
func updateNote(w http.ResponseWriter, r *http.Request) {
	// Read value from rout variable
	vars := mux.Vars(r)
	k := vars["id"]
	var noteToUpdate Note
	if _, ok := noteStore[k]; ok {
		r.ParseForm()
		noteToUpdate.Title = r.PostFormValue("title")
		noteToUpdate.Description = r.PostFormValue("description")
		noteToUpdate.CreatedOn = time.Now()

		// delete existing item and add updated item
		delete(noteStore, k)
		noteStore[k] = noteToUpdate
	} else {
		http.Error(w, "Could not find the resource to update", http.StatusBadRequest)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Handler for "/notes/delete/{id}" which deletes an item from the store
func deleteNote(w http.ResponseWriter, r *http.Request) {
	// Read value from route variable
	vars := mux.Vars(r)
	k := vars["id"]
	// Remove from store
	if _, ok := noteStore[k]; ok {
		// delete existing item
		delete(noteStore, k)
	} else {
		http.Error(w, "Could not fine the resource to delete", http.StatusBadRequest)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Entry point of the program
func main() {

	r := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("public"))
	r.Handle("/public/", fs)
	r.HandleFunc("/", getNotes)
	r.HandleFunc("/notes/add", addNote)
	r.HandleFunc("/notes/save", saveNote)
	r.HandleFunc("/notes/edit/{id}", editNote)
	r.HandleFunc("/notes/update/{id}", updateNote)
	r.HandleFunc("/notes/delete/{id}", deleteNote)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	log.Println("Listening...")
	server.ListenAndServe()
}
