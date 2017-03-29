package main

import (
	"crypto/sha256"
	rice "github.com/GeertJohan/go.rice"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type Event struct {
	gorm.Model
	Title       string
	Description string
	Files       string
}

var db *gorm.DB

func main() {

	initData()
	initDb()

	e := echo.New()

	e.Use(middleware.Logger())

	publicHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())

	e.GET("/", echo.WrapHandler(publicHandler))
	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", publicHandler)))

	e.POST("/events", addEvent)

	e.Logger.Fatal(e.Start(":1323"))
}

func initData() {
	var mode os.FileMode = 0777
	path := "./data"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, mode)
	}

}

func initDb() {

	db, err := gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Event{})
}

func listEvents(c echo.Context) error {
	return c.String(http.StatusOK, "List events")
}

func addEvent(c echo.Context) error {
	title := c.FormValue("title")
	title = title

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	files := form.File["files"]
	var file_paths []string

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			panic("fail reading file from form")
		}

		h := sha256.New()
		if _, err := io.Copy(h, src); err != nil {
			panic("fail coping file to hash")
		}

		hash := h.Sum(nil)

		file_paths = append(file_paths, saveFile(src, hash))

		defer src.Close()
	}

	return c.String(http.StatusOK, "Add worked")
}

func saveFile(src multipart.File, hash []byte) string {
	return "123"
}
