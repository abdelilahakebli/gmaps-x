package main

import (
	"flag"
	"fmt"
	"gmaps-x/internal"
	"log"
	"os"
	"runtime"

	"github.com/gocarina/gocsv"
	"github.com/playwright-community/playwright-go"
	"github.com/schollz/progressbar/v3"
)

func main() {

	var query string
	var emails bool
	var lang string
	var limit int
	var concurrency int
	var headless bool
	var numCpus int = (runtime.NumCPU()) / 2
	var timeout int
	var cooldown int
	var output string

	const DEFAULT_LANG = "en"
	const DEFAULT_LIMIT = 10
	const DEFAULT_QUERY = "restaurants in new york"
	const DEFAULT_HEADLESS = false
	const DEFAULT_TIMEOUT = 10
	const DEFAULT_COULDOWN = 5
	const DEFAULT_OUTPUT = "results.csv"
	const DEFAULT_EMAILS = false

	flag.IntVar(&concurrency, "c", numCpus, "number of concurrent jobs")
	flag.BoolVar(&emails, "emails", DEFAULT_EMAILS, "extract emails")
	flag.StringVar(&lang, "lang", DEFAULT_LANG, "language of the search")
	flag.IntVar(&limit, "limit", DEFAULT_LIMIT, "depth of the search")
	flag.StringVar(&query, "query", DEFAULT_QUERY, "query to search")
	flag.BoolVar(&headless, "headless", DEFAULT_HEADLESS, "run browser in headless mode")
	flag.IntVar(&timeout, "timeout", DEFAULT_TIMEOUT, "timeout in seconds")
	flag.IntVar(&cooldown, "cooldown", DEFAULT_COULDOWN, "couldown in seconds")
	flag.StringVar(&output, "output", DEFAULT_OUTPUT, "output file")

	flag.Parse()

	// Init Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	context.Browser().Contexts()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	defer context.Close()

	// Init search
	search := internal.NewSearch(&context, query, lang, limit, timeout, cooldown, false)
	// search.Browser = &context
	// search.Cooldown = cooldown
	fmt.Println("Getting items...")

	search.Progress = progressbar.Default(int64(limit))

	if err := search.GetPlaces(); err != nil {
		log.Fatalf("could not get places: %v", err)
	}

	fmt.Println("Fetching details...")
	_ = search.Extract(concurrency, progressbar.Default(int64(len(search.Jobs))))

	// if emails {
	// 	fmt.Println("Extracting emails...")
	// 	_ = search.ExtractEmails(concurrency, progressbar.Default(int64(len(search.Jobs))))
	// }

	fmt.Println("Extracting emails...")
	_ = search.ExtractEmails(emails, concurrency, progressbar.Default(int64(len(search.Jobs))))

	log.Printf("Saving results to %v", output)
	outputFile, _ := os.OpenFile(output, os.O_RDWR|os.O_CREATE, os.ModePerm)

	_ = gocsv.MarshalFile(search.Places, outputFile)
}
