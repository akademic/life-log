package main

import (
	"bytes"
	"github.com/erikstmartin/go-testdb"
	"github.com/jinzhu/gorm"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupDB() {
	db, _ = gorm.Open("testdb", "")
	db.LogMode(true)
}

func TestListEvents(t *testing.T) {

	setupDB()

	sql := `SELECT * FROM "events"  WHERE "events"."deleted_at" IS NULL ORDER BY id desc LIMIT 10`
	columns := []string{"id", "title", "description"}
	result := `
			3,title1,desc1
			2,title2,desc2
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, listEvents(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		JSON := "{\"Events\":[{\"ID\":3,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Title\":\"title1\",\"Description\":\"desc1\",\"Files\":null},{\"ID\":2,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Title\":\"title2\",\"Description\":\"desc2\",\"Files\":null}]}"
		assert.Equal(t, JSON, rec.Body.String())
	}

}

func TestGetEvent(t *testing.T) {

	setupDB()

	sql := `SELECT * FROM "events"  WHERE "events"."deleted_at" IS NULL AND (("events"."id" = ?)) ORDER BY "events"."id" ASC LIMIT 1`
	columns := []string{"id", "title", "description"}
	result := `
			1,title2,desc2
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	sql_file := `SELECT * FROM "files"  WHERE "files"."deleted_at" IS NULL AND (("event_id" = ?))`
	columns_file := []string{"id", "name", "storage_path", "event_id"}
	result_files := `1,file.jpg,data/file.jpg,1`

	testdb.StubQuery(sql_file, testdb.RowsFromCSVString(columns_file, result_files))

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/events/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	if assert.NoError(t, getEvent(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		JSON := "{\"ID\":1,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Title\":\"title2\",\"Description\":\"desc2\",\"Files\":[{\"ID\":1,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Name\":\"file.jpg\",\"StoragePath\":\"data/file.jpg\",\"EventID\":1}]}"
		assert.Equal(t, JSON, rec.Body.String())
	}

}

func TestAddEvent(t *testing.T) {
	setupDB()

	Body := strings.NewReader(`------WebKitFormBoundaryJMDX924v2WrPw8Wl
Content-Disposition: form-data; name="title"

test
------WebKitFormBoundaryJMDX924v2WrPw8Wl
Content-Disposition: form-data; name="description"

test2
------WebKitFormBoundaryJMDX924v2WrPw8Wl
Content-Disposition: form-data; name="files"; filename=""
Content-Type: application/octet-stream

1234
------WebKitFormBoundaryJMDX924v2WrPw8Wl--`)

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/events", Body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm+"; boundary=----WebKitFormBoundaryJMDX924v2WrPw8Wl")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, addEvent(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Add worked", rec.Body.String())
	}

}

func TestUpdateEvent(t *testing.T) {
	setupDB()

	sql := `SELECT * FROM "events"  WHERE "events"."deleted_at" IS NULL ORDER BY "events"."id" ASC LIMIT 1`
	columns := []string{"id", "title", "description"}
	result := `
			1,title2,desc2
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(columns, result))

	Body := strings.NewReader(`------WebKitFormBoundaryJMDX924v2WrPw8Wl
Content-Disposition: form-data; name="title"

test
------WebKitFormBoundaryJMDX924v2WrPw8Wl
Content-Disposition: form-data; name="description"

test2
------WebKitFormBoundaryJMDX924v2WrPw8Wl--`)

	e := echo.New()
	req := httptest.NewRequest(echo.PUT, "/events/1", Body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm+"; boundary=----WebKitFormBoundaryJMDX924v2WrPw8Wl")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/events/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	if assert.NoError(t, updateEvent(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"Result\":\"ok\"}", rec.Body.String())
	}

}

func TestDeleteFile(t *testing.T) {
	setupDB()

	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/events/1/files/133", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/events/:id/files/:file_id")
	c.SetParamNames("id", "file_id")
	c.SetParamValues("1", "1")

	if assert.NoError(t, deleteFile(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"Result\":\"ok\"}", rec.Body.String())
	}
}

func TestAddFile(t *testing.T) {
	setupDB()

	var b bytes.Buffer

	mpw := multipart.NewWriter(&b)

	fh, _ := mpw.CreateFormFile("file", "test_file.jpg")
	fh.Write([]byte("1234"))
	mpw.Close()

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/events/1/files", &b)
	//req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm+"; boundary=----WebKitFormBoundaryJMDX924v2WrPw8Wl")
	req.Header.Set(echo.HeaderContentType, mpw.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	conf = &Config{}
	conf.Data_dir = "./data"

	if assert.NoError(t, addFile(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"Result\":\"ok\"}", rec.Body.String())
	}
}
