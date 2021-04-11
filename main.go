package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/leejames00/learngo/scrapper"
)

const fileName string = "jobs.csv"

func main() {
	e := echo.New()

	// Routes
	e.GET("/", hello)
	e.POST("/scrape", handleScrape)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
	// scrapper.Scrape("python")
}

func hello(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := scrapper.CleanString(c.FormValue("term"))
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)
}
