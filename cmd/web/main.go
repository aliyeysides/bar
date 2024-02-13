package main

import (
	"html/template"
	"io"
	"log"
	"time"

	"database/sql"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate() *Template {
	return &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Form struct {
	Value string
	Error string
	URL   string
}

func NewForm() *Form {
	return &Form{}
}

func ResetForm(form *Form) {
	form.Value = ""
	form.Error = ""
	form.URL = ""
}

func cleanupViewedUrls(db *sql.DB) {
	for {
		<-time.After(time.Hour) // Wait for 1 hour before each cleanup run

		_, err := db.Exec("DELETE FROM urls WHERE viewed_at IS NOT NULL AND viewed_at < datetime('now', '-1 hour')")
		if err != nil {
			log.Printf("Error cleaning up urls: %v", err)
		} else {
			log.Println("Old viewed urls cleaned up successfully.")
		}
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE urls (id TEXT PRIMARY KEY, url TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, value TEXT, viewed_at TIMESTAMP)")

	if err != nil {
		log.Println(err)
	}

	log.Println("Database and table created")

	e := echo.New()
	e.Use(middleware.Logger())

	e.Renderer = NewTemplate()
	e.Static("/public/css", "public/css")

	form := NewForm()

	e.GET("/", func(c echo.Context) error {
		ResetForm(form)
		return c.Render(200, "index", form)
	})

	e.POST("/generate", func(c echo.Context) error {
		input := c.FormValue("input")

		if input == "" {
			form.URL = ""
			form.Error = "Please enter a value"
			return c.Render(422, "form", form)
		}

		id, err := gonanoid.New(7)
		if err != nil {
			form.Error = "Error generating id"
			return c.Render(500, "form", form)
		}

		newUUID := uuid.New().String()
		_, err = db.Exec("INSERT INTO urls (id, url, value) VALUES (?, ?, ?)", newUUID, id, input)
		if err != nil {
			form.Error = "Error saving to database"
			return c.Render(500, "form", form)
		}

		url := "localhost:8080" + "/" + id
		form.URL = url

		form.Value = ""
		form.Error = ""
		return c.Render(200, "form", form)
	})

	e.GET("/:id", func(c echo.Context) error {
		id := c.Param("id")
		err := db.QueryRow("SELECT value FROM urls WHERE url = ?", id).Scan(&form.Value)
		if err != nil {
			log.Println(err)
			form.Error = "URL not found"
			return c.Render(404, "form", form)
		}
		// mark as viewed
		_, err = db.Exec("UPDATE urls SET viewed_at = CURRENT_TIMESTAMP WHERE url = ?", id)
		return c.Render(200, "secret", form)
	})

	go cleanupViewedUrls(db)

	e.Logger.Fatal(e.Start("localhost:8080"))
}
