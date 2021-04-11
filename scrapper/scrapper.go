package scrapper

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Scrape ..
func Scrape(term string) {
	var jobs []extractedJob
	c := make(chan []extractedJob)

	baseURL := "https://kr.indeed.com/jobs?q=" + term + "&limit=50"

	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkError(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"Link", "Title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkError(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkError(jwErr)
	}
}

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&start=" + strconv.Itoa(page*50)

	res, err := http.Get(pageURL)
	checkError(err)
	checkCode(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")

	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find(".salaryText ").Text())
	summary := CleanString(card.Find("summary").Text())

	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0

	res, err := http.Get(url)

	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(resp *http.Response) {
	if resp.StatusCode != 200 {
		log.Fatalln(resp)
	}
}
