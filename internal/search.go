package internal

import (
	"gmaps-x/helpers"
	"gmaps-x/types"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/schollz/progressbar/v3"
)

// type Place types.Place
// type Job types.Job
type Search types.Search

func NewSearch(browser *playwright.BrowserContext, query string, lang string, limit int, timeout int, cooldown int, apimode bool) *Search {
	return &Search{
		Query:    query,
		Lang:     lang,
		Browser:  browser,
		Timeout:  timeout,
		ApiMode:  apimode,
		Cooldown: cooldown,
		Limit:    limit,
		Jobs:     make([]*types.Job, 0),
	}
}

func (search *Search) GetPlaces() (err error) {
	search.Page, err = (*search.Browser).NewPage()

	if err != nil {
		return err
	}
	defer search.Page.Close()

	_, err = search.Page.Goto(helpers.GoogleMapsFormatQueryToURLSearchFormat(search.Query, search.Lang))

	if err != nil {
		return err
	}
	// search.Page.SetViewportSize(1920, 1080)

	helpers.WaitForLoadStateIdle(&search.Page, search.Timeout)
	search.getResultFeedItems()
	// Get all places

	return nil
}
func (search *Search) getResultFeedItems() (err error) {

	const RESULT_FEED_LOCATOR_SELECTOR = "div[role='feed']"
	const RESULT_FEED_ITEMS_LOCATOR_SELECTOR = "div:not([class]) >>> div[jsaction]"

	resultFeed := search.Page.Locator(RESULT_FEED_LOCATOR_SELECTOR).First()
	resultFeedItems, _ := resultFeed.Locator(RESULT_FEED_ITEMS_LOCATOR_SELECTOR).All()
	lastResultFeedItemsLength := 0
	stop := false

	lastLenMultiplexer := 0

	for !stop {

		if len(resultFeedItems) >= search.Limit {
			resultFeedItems = resultFeedItems[:search.Limit]
			stop = true
		}

		if len(resultFeedItems) == lastResultFeedItemsLength {
			if lastLenMultiplexer > 2 {
				break
			}
			lastLenMultiplexer++
		} else {
			lastLenMultiplexer = 0
		}

		timeFisrt := time.Now()
		for idx, item := range resultFeedItems[lastResultFeedItemsLength:] {
			job := search.extractResultItemData(&item)
			job.Index = idx + lastResultFeedItemsLength
			job.Place.Index = job.Index
			if !search.ApiMode {
				search.Progress.Add(1)
			}

			search.Jobs = append(search.Jobs, job)
		}
		timeSecond := time.Now()

		lastLink := resultFeedItems[len(resultFeedItems)-1]

		if !stop {

			lastLink.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{
				Timeout: playwright.Float(float64(search.Timeout) * 1000),
			})
			timeSleep := (time.Duration(search.Cooldown) * time.Second) - timeSecond.Sub(timeFisrt)

			if timeSleep < time.Duration(3)*time.Second {
				timeSleep = time.Duration(3) * time.Second

			}
			if err := helpers.WaitForLoadStateIdle(&search.Page, search.Timeout); err != nil {
				break
			}

			time.Sleep(timeSleep)
		}

		lastResultFeedItemsLength = len(resultFeedItems)
		resultFeedItems, _ = resultFeed.Locator(RESULT_FEED_ITEMS_LOCATOR_SELECTOR).All()

	}

	if len(search.Jobs) < search.Limit {
		search.Jobs = search.Jobs[:len(search.Jobs)]
	} else {
		search.Jobs = search.Jobs[:search.Limit]
	}
	return nil
}

func (search *Search) extractResultItemData(listItem *playwright.Locator) *types.Job {
	link := (*listItem).Locator("a[jsaction]").First()
	title, _ := link.GetAttribute("aria-label")
	url, _ := link.GetAttribute("href")
	dataContainer := (*listItem).Locator("div").Last()
	sponsored, _ := dataContainer.Locator("[aria-label^='Sponsor']").IsVisible()

	textData, _ := (*listItem).InnerText()
	var dataArray = strings.Split(textData, "\n")
	if sponsored {
		dataArray = dataArray[1:]
	}

	lineData := strings.Split(dataArray[1], "·")
	var reviewCount int
	var reviewRating float64
	var priceRange string

	reviewRating, _ = strconv.ParseFloat(lineData[0][:3], 64)
	reviewCountTmp := strings.Trim(lineData[0][3:], " ")
	pattern := `^\(([^)]+)\)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(reviewCountTmp)

	if len(match) > 1 {
		reviewCountTmp = match[1]
		reviewCount, _ = strconv.Atoi(reviewCountTmp)
	} else {
		reviewCount = 0
	}

	if len(lineData) > 1 {
		priceRange = lineData[1]
	}

	job := &types.Job{}
	job.Title = title
	job.Link = url
	job.Sponsored = sponsored
	job.Done = false
	job.Place = &types.Place{
		Title:        title,
		Link:         url,
		ReviewsCount: reviewCount,
		ReviewRating: reviewRating,
		PriceRange:   priceRange,
	}
	search.Places = append(search.Places, job.Place)
	// log.Printf("Review count: %s\nReviews rating: %s\nPrice range: %s", reviewCount, reviewRating, priceRange)

	return job

}

func (search *Search) Extract(concurrency int, progress *progressbar.ProgressBar) error {

	trigger := make(chan struct{})
	wg := sync.WaitGroup{}
	wgcount := 0
	for _, job := range search.Jobs {

		if wgcount >= concurrency {
			<-trigger
			wgcount--
		}

		wg.Add(1)
		wgcount++
		go func(job *types.Job, w *sync.WaitGroup) {

			job.Place.Export(search.Browser, search.Timeout, w)
			// log.Printf("Exported: %v\n", job.Place.Title)
			(*job.Place.Page).Close()

			if !search.ApiMode {
				progress.Add(1)
			}
			job.Place.Index = job.Index + 1
			// log.Printf("Exported %v: %v\n", index, job.Place.Title)
			// search.Places = append(search.Places, job.Place)
			// place.Done = true

			trigger <- struct{}{}

		}(job, &wg)

	}

	wg.Wait()
	time.Sleep(time.Duration(500) * time.Millisecond)

	sort.Slice(search.Places, func(i, j int) bool {
		return search.Places[i].Index < search.Places[j].Index
	})

	return nil
}

func (search *Search) ExtractEmails(extract bool, concurrency int, progress *progressbar.ProgressBar) error {

	trigger := make(chan struct{})
	wg := sync.WaitGroup{}
	wgcount := 0

	for _, job := range search.Jobs {

		if !extract {
			job.Place.Emails = "disabled"
			continue
		}

		if wgcount >= concurrency {
			<-trigger
			wgcount--
		}

		wg.Add(1)
		wgcount++
		go func(index int, job *types.Job, w *sync.WaitGroup) {

			job.Place.ExtractEmails(search.Browser, search.Timeout, w)
			if !search.ApiMode {

				progress.Add(1)
			}
			trigger <- struct{}{}

		}(job.Index, job, &wg)

	}

	wg.Wait()

	if extract {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	return nil
}
