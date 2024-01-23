package main

import (
	"flag"
	"fmt"
	"gmaps-x/internal"
	"gmaps-x/types"
	"log"
	"os"
	"sort"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
)

const HOST = "localhost"
const PORT = "3030"
const API_MODE = true

func main() {
	var jobs map[string]*types.ApiReponses = make(map[string]*types.ApiReponses, 0)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	var headless bool
	var port string

	const DEFAULT_HEADLESS = false
	flag.BoolVar(&headless, "headless", DEFAULT_HEADLESS, "run browser in headless mode")
	flag.StringVar(&port, "port", PORT, "port to listen on")
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

	app.Post("/search", func(c *fiber.Ctx) error {

		request := new(types.CreateJobRequest)
		if err := c.BodyParser(request); err != nil {
			fmt.Println("error = ", err)
			return c.SendStatus(400)
		}

		search := internal.NewSearch(&context, request.Query, request.Lang, request.Limit, request.Timeout, request.Cooldown, API_MODE)

		uid := uuid.NewString()

		response := &types.ApiReponses{
			Timestamp:   time.Now().Local().Unix(),
			JobId:       uid,
			JobStatus:   "running",
			JobProgress: 0,
			Query:       request.Query,
			Lang:        request.Lang,
			Places:      make([]*types.Place, 0),
		}

		go func(uid string, search *internal.Search) {
			err := search.GetPlaces()
			if err != nil {
				jobs[uid].Success = false
				jobs[uid].JobStatus = "stopeed on places extraction"
			}

			err = search.Extract(request.Concurrency, nil)
			if err != nil {
				jobs[uid].Success = false
				jobs[uid].JobStatus = "stopped on details extraction"
			}

			err = search.ExtractEmails(request.Emails, request.Concurrency, nil)
			if err != nil {
				jobs[uid].Success = false
				jobs[uid].JobStatus = "stopped on emails extraction"
			}

			jobs[uid].Success = true
			jobs[uid].JobStatus = "done"
			jobs[uid].Places = search.Places
			// responses = append(responses, response)
		}(uid, search)
		jobs[uid] = response
		return c.JSON(response)
	})

	app.Get("/search/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		return c.JSON(jobs[id])
	})

	app.Get("/search", func(c *fiber.Ctx) error {

		type JobTemp struct {
			Timestamp  int64  `json:"timestamp"`
			JobId      string `json:"job_id"`
			JobStatus  string `json:"job_status"`
			JobSuccess bool   `json:"job_success"`
			JobQuery   string `json:"job_query"`
			JobLang    string `json:"job_lang"`
			Limit      int    `json:"limit"`
		}

		jobstemp := make([]*JobTemp, 0)

		for k, v := range jobs {
			jobstemp = append(jobstemp, &JobTemp{
				JobId:      k,
				Timestamp:  v.Timestamp,
				JobStatus:  v.JobStatus,
				JobSuccess: v.Success,
				JobQuery:   v.Query,
				JobLang:    v.Lang,
				Limit:      len(v.Places),
			})
		}

		sort.Slice(jobstemp, func(i, j int) bool {
			return jobstemp[i].Timestamp < jobstemp[j].Timestamp
		})
		return c.JSON(jobstemp)
	})

	app.Delete("/search", func(c *fiber.Ctx) error {
		jobs = make(map[string]*types.ApiReponses, 0)
		return c.JSON(jobs)
	})

	app.Get("/csv/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if _, ok := jobs[id]; !ok {
			return c.SendStatus(404)
		}

		filepath := fmt.Sprintf("data/results-%s.csv", id)
		outputFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return c.SendStatus(500)
		}
		err = gocsv.MarshalFile(jobs[id].Places, outputFile)

		if err != nil {
			return c.SendStatus(500)
		}
		return c.SendFile(filepath)
	})

	log.Fatal(app.Listen(":" + port))
}
