package main

import (
	"github.com/labstack/echo"
	"net/http"
)

func main() {
	e := echo.New()

	e.POST("/events", addEvent)

	e.Logger.Fatal(e.Start(":1323"))
}

func addEvent(c echo.Context) error {
	return c.String(http.StatusOK, "Add worked")
}
