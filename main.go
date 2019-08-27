package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	host     = os.Getenv("DB_HOST")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	port     = os.Getenv("DB_PORT")
	dbname   = os.Getenv("DB_DATABASE")
)

type Note struct {
	Id   int
	Date time.Time
	Text string
}

type PageNote struct {
	Id   int
	Date string
	Text string
}

func replaceChars(s string) string {
	s = strings.Replace(s, `\n`, "<br>", -1)
	s = strings.Replace(s, `\t`, "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	return s
}

func root(w http.ResponseWriter, r *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	fmt.Println("Connecting to", psqlInfo)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	delete := r.URL.Query().Get("delete")
	if delete != "" {
		fmt.Println("Delete note", delete)
		query := `
		DELETE FROM notes
	    WHERE id=$1`
		_, err = db.Exec(query, delete)
		if err != nil {
			panic(err)
		}
		w.Write([]byte("Deleted"))

		return
	}

	entry := r.FormValue("text")
	if len(entry) > 0 {
		query := `
		INSERT INTO notes (date, text)
		VALUES ($1, $2)`
		_, err = db.Exec(query,
			time.Now().Format(time.RFC3339), entry)
		if err != nil {
			panic(err)
		}
	}

	rows, err := db.Query(
		`SELECT id, date, text
		FROM notes
		ORDER BY date DESC`)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var note Note
	var notes []PageNote
	for rows.Next() {
		err = rows.Scan(
			&note.Id,
			&note.Date,
			&note.Text)

		if err != nil {
			panic(err)
		}

		pageNote := PageNote{
			Id:   note.Id,
			Date: note.Date.Format("Mon, 02 Jan 2006 15:04:05 MST"),
			Text: note.Text}

		notes = append(notes, pageNote)
	}

	t := template.Must(template.ParseFiles("index.html"))

	t.Execute(w, notes)

}

func main() {
	http.HandleFunc("/", root)

	fmt.Printf("Started server on port 8080...\n")

	http.ListenAndServe(":8080", nil)
}
