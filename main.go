package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type News struct {
	ID    int
	Title string
	Anons string
	Text  string
}

func dbConnect() (*sql.DB, error) {
	connStr := "postgresql://postgres:GcuN2WTUM3bhb9LM@db.eindaciqxixgylvblosq.supabase.co:5432/postgres"
	return sql.Open("postgres", connStr)
}

func index(w http.ResponseWriter, r *http.Request) {
	db, err := dbConnect()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, anons, text FROM news ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var newsList []News
	for rows.Next() {
		var n News
		err := rows.Scan(&n.ID, &n.Title, &n.Anons, &n.Text)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		newsList = append(newsList, n)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmpl.Execute(w, newsList)
}

func create(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func save(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	anons := r.FormValue("anons")
	text := r.FormValue("text")

	db, err := dbConnect()
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO news (title, anons, text) VALUES ($1, $2, $3)", title, anons, text)
	if err != nil {
		http.Error(w, "Save Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func article(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}

	db, err := dbConnect()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()

	var n News
	row := db.QueryRow("SELECT id, title, anons, text FROM news WHERE id = $1", id)
	err = row.Scan(&n.ID, &n.Title, &n.Anons, &n.Text)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmpl, err := template.ParseFiles("templates/article.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmpl.Execute(w, n)
}

func handleFunc() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/article", article)
	r.HandleFunc("/create", create).Methods("GET")
	r.HandleFunc("/save", save).Methods("POST")
	http.ListenAndServe(":8080", r)
}

func main() {
	handleFunc()
}
