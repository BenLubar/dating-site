package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dateFormat = "2006-01-02"

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Ben Lubar’s Dating Site</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="go-import" content="ben.lubar.me/dating-site git https://github.com/BenLubar/dating-site.git">
<style>
body {
font-family: sans-serif;
}
.error {
background: #f77;
color: #000;
}
</style>
</head>
<body>
<h1>Ben Lubar’s Dating Site</h1>
<form method="post" action="submit">
{{with .Error -}}
<div class="error">{{.}}</div>
{{end -}}
<label for="date">Date:</label> <input id="date" name="date" type="date" required autofocus>
<br><input type="submit" value="Submit">
</form>
<ul id="dates">
{{range .Dates -}}
<li>{{.}}</li>
{{end -}}
</ul>
</body>
</html>
`))

func main() {
	db, err := sql.Open("sqlite3", "data/dates.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS dates ( year INT, month INT, day INT );`)
	if err != nil {
		panic(err)
	}

	getAll, err := db.Prepare(`SELECT year, month, day FROM dates ORDER BY rowid DESC LIMIT 10000;`)
	if err != nil {
		panic(err)
	}
	insert, err := db.Prepare(`INSERT INTO dates (year, month, day) VALUES (?, ?, ?);`)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=30")

		dates, err := getAll.Query()
		if err != nil {
			panic(err)
		}
		defer dates.Close()

		var data struct {
			Error string
			Dates []string
		}
		data.Error = r.URL.Query().Get("error")

		for dates.Next() {
			var year, month, day int
			if err = dates.Scan(&year, &month, &day); err != nil {
				panic(err)
			}

			data.Dates = append(data.Dates, time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format(dateFormat))
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if data.Error != "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-cache")

		var err error
		if r.Method == http.MethodPost {
			var d time.Time
			d, err = time.Parse(dateFormat, r.FormValue("date"))
			if err == nil {
				year, month, day := d.Date()
				if year >= 1480 && year < 1490 {
					// nah, let's not.
				} else {
					_, err = insert.Exec(year, int(month), day)
				}
			}
		}

		path := "/dating-site/"
		if err != nil {
			path += "?error=" + url.QueryEscape(err.Error())
		}
		http.Redirect(w, r, path, http.StatusFound)
	})

	panic(http.ListenAndServe(":80", nil))
}
