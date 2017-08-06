package main

import (
	"github.com/erikstmartin/go-testdb"
	"github.com/jinzhu/gorm"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEvent(t *testing.T) {

	db, _ = gorm.Open("testdb", "")
	//db.LogMode(true)

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
	c.Set("db", db)

	if assert.NoError(t, listEvents(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		JSON := "{\"Events\":[{\"ID\":3,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Title\":\"title1\",\"Description\":\"desc1\",\"Files\":null},{\"ID\":2,\"CreatedAt\":\"0001-01-01T00:00:00Z\",\"UpdatedAt\":\"0001-01-01T00:00:00Z\",\"DeletedAt\":null,\"Title\":\"title2\",\"Description\":\"desc2\",\"Files\":null}]}"
		assert.Equal(t, JSON, rec.Body.String())
	}

}
