package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	rice "github.com/GeertJohan/go.rice"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

type Event struct {
	gorm.Model
	Title       string
	Description string
	Files       []File
}

type File struct {
	gorm.Model
	Name        string
	StoragePath string
	EventID     uint
}

type ResponseEventsList struct {
	Events []Event
}

type Config struct {
	Db_name  string
	Data_dir string
}

var db *gorm.DB
var conf *Config

func main() {

	conf = &Config{}
	conf.Db_name = "database.db"
	conf.Data_dir = "./data"

	initData()
	initDb()

	e := echo.New()

	e.Use(middleware.Logger())

	publicHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())

	e.GET("/", echo.WrapHandler(publicHandler))
	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", publicHandler)))

	e.POST("/events", addEvent)
	e.GET("/events", listEvents)
	e.GET("/events/:id", getEvent)

	e.Logger.Fatal(e.Start(":1323"))
}

func initData() {
	var mode os.FileMode = 0755
	path := conf.Data_dir
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, mode)
	}

}

func initDb() {

	var err error
	db, err = gorm.Open("sqlite3", conf.Db_name)
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Event{}, &File{})
}

func getEvent(c echo.Context) error {
	id := c.Param("id")
	var event Event

	db.First(&event, id)

	return c.JSON(http.StatusOK, event)
}

func listEvents(c echo.Context) error {
	var events []Event
	db.Order("id desc").Limit(10).Find(&events)

	resp_events := ResponseEventsList{Events: events}
	return c.JSON(http.StatusOK, resp_events)
}

func addEvent(c echo.Context) error {
	title := c.FormValue("title")
	desc := c.FormValue("description")

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	files := form.File["files"]
	var store_files []File

	for _, file := range files {
		path, name := saveFile(file)
		store_files = append(store_files, File{Name: name, StoragePath: path})
	}

	fmt.Println(store_files)

	event := &Event{
		Title:       title,
		Description: desc,
		Files:       store_files,
	}

	fmt.Println(event)

	db.Create(event)

	return c.String(http.StatusOK, "Add worked")
}

func saveFile(file_header *multipart.FileHeader) (string, string) {

	hpath, subdirs := getPathForSaving(file_header)

	prepareSubdirs(subdirs)

	full_hpath := path.Join(conf.Data_dir, hpath)

	file, err := file_header.Open()
	if err != nil {
		panic("fail reading file from form second time")
	}

	defer file.Close()

	file_save, err := os.Create(full_hpath)
	if err != nil {
		panic("Open file for saving failed")
	}
	defer file_save.Close()

	if _, err := io.Copy(file_save, file); err != nil {
		panic("Save file failed")
	}

	return hpath, file_header.Filename
}

func prepareSubdirs(subdirs []string) string {
	dir_path := conf.Data_dir
	var mode os.FileMode = 0755

	for _, sd := range subdirs {
		dir_path = path.Join(dir_path, sd)
		if _, err := os.Stat(dir_path); os.IsNotExist(err) {
			os.Mkdir(dir_path, mode)
		}
	}

	return dir_path
}

func getPathForSaving(file_header *multipart.FileHeader) (string, []string) {

	file, err := file_header.Open()
	if err != nil {
		panic("fail reading file from form")
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		panic("fail coping file to hash")
	}

	str_hash := hex.EncodeToString(h.Sum(nil))

	name_parts := strings.Split(file_header.Filename, ".")
	ext := name_parts[len(name_parts)-1]

	fst := str_hash[0:3]
	sec := str_hash[3:6]
	last := str_hash[6:len(str_hash)] + "." + ext

	return path.Join(fst, sec, last), []string{fst, sec}
}
