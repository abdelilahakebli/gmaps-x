package types

import (
	"fmt"
	"gmaps-x/helpers"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

type Place struct {
	Index       int `json:"index" csv:"index"`
	Page        *playwright.Page
	Title       string  `json:"title" csv:"title"`
	Category    string  `json:"category" csv:"category"`
	Link        string  `json:"link" csv:"link"`
	Phone       string  `json:"phone" csv:"phone"`
	Coordinates string  `json:"coordinates" csv:"coordinates"`
	Latitude    float64 `json:"latitude" csv:"latitude"`
	Longtitude  float64 `json:"longtitude" csv:"longtitude"`
	Timezone    string  `json:"timezone" csv:"timezone"`
	// Description     string  `json:"description" csv:"description"`
	// Query string `json:"query" csv:"query"`
	// CreatedDate     string  `json:"created_date" csv:"created_date"`
	Owner     bool   `json:"owner" csv:"owner"`
	Website   string `json:"website" csv:"website"`
	Image     string `json:"image" csv:"image"`
	Sponsored bool   `json:"sponsored" csv:"sponsored"`
	Address   string `json:"address" csv:"address"`
	// CompleteAddress string  `json:"complete_address" csv:"complete_address"`
	// About           string  `json:"about" csv:"about"`
	Emails       string  `json:"emails" csv:"emails"`
	ReviewsCount int     `json:"reviews_count" csv:"reviews_count"`
	PlusCode     string  `json:"plus_code" csv:"plus_code"`
	ReviewRating float64 `json:"review_rating" csv:"review_rating"`
	// ReviewsPerStar  []int16 `json:"reviews_per_star" csv:"reviews_per_star"`
	// ReviewsLink     string  `json:"reviews_link" csv:"reviews_link"`

	// Cid    string `json:"cid" csv:"cid"`
	// Status string `json:"status" csv:"status"`

	PriceRange string `json:"price_range" csv:"price_range"`
	// DataId       string `json:"data_id" csv:"data_id"`
	// Reservations string `json:"reservations" csv:"reservations"`
	// OrderOnline  string `json:"order_online" csv:"order_online"`
	Menu string `json:"menu" csv:"menu"`
	// UsersReviews    string  `json:"users_reviews" csv:"users_reviews"`
	// OpenHours       string  `json:"open_hours" csv:"open_hours"`
}

func (place *Place) Export(browser *playwright.BrowserContext, timeout int, wg *sync.WaitGroup) error {
	defer (*wg).Done()
	page, err := (*browser).NewPage()

	if err != nil {
		return err
	}
	place.Page = &page
	defer page.Close()

	if _, err := page.Goto(place.Link); err != nil {
		return err
	}
	if err := helpers.WaitForLoadStateIdle(&page, timeout); err != nil {
		return err
	}

	if err := place.extractInsights(); err != nil {
		return err
	}

	return nil

}

func (place *Place) extractInsights() error {
	// Extracting data
	resultContainer := (*place.Page).Locator("div[role='main']").First()

	image, _ := resultContainer.Locator("img").First().GetAttribute("src")
	// title, _ := resultContainer.Locator("div.TIHn2 > div.tAiQdd > div.lMbq3e > div:not([class]) > h1").First().InnerText()
	category, _ := resultContainer.Locator("div.TIHn2 > div.tAiQdd > div.lMbq3e > div.LBgpqf > div.skqShb > div.fontBodyMedium > span > span > button[jsaction]").First().InnerHTML()

	// informations, _ := resultContainer.Locator("div[role='region'] >>> div:not([jslog])").All()

	addressLocatorQuery := "div[role='region'] >>> div:not([jslog]) > button[data-item-id='address']"
	websiteLocatorQuery := "div[role='region'] >>> div:not([jslog]) > [data-item-id='authority']"
	phoneLocatorQuery := "div[role='region'] >>> div:not([jslog]) > [data-item-id^='phone']"
	plusCodeLocatorQuery := "div[role='region'] >>> div:not([jslog]) > [data-item-id='oloc']"
	// reviewsLocatorQuery := "div[jsaction='pane.reviewChart.moreReviews'] >>> div"
	// reviewsCountLocatorQuery := "div[jsaction='pane.reviewChart.moreReviews'] >>> div > button >>> span"
	ownerLocatorQuery := "div[role='region'] >>> div:not([jslog]) > [data-item-id='merchant']"
	menuLocatorQuery := "div[role='region'] >>> div:not([jslog]) > [data-item-id='menu']"

	var address string
	var website string
	var phone string
	var pluscode string
	var owner bool
	var menu string

	addressLocator := resultContainer.Locator(addressLocatorQuery).First()
	websiteLocator := resultContainer.Locator(websiteLocatorQuery).First()
	phoneLocator := resultContainer.Locator(phoneLocatorQuery).First()
	plusCodeLocator := resultContainer.Locator(plusCodeLocatorQuery).First()
	ownerLocator := resultContainer.Locator(ownerLocatorQuery).First()
	menuLocator := resultContainer.Locator(menuLocatorQuery).First()

	resultContainer.Locator("div[role='region']").First().ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{
		Timeout: playwright.Float(500),
	})
	if visible, err := addressLocator.IsVisible(); err == nil && visible {
		address, _ = addressLocator.TextContent()
	}

	if visible, err := websiteLocator.IsVisible(); err == nil && visible {
		website, _ = websiteLocator.GetAttribute("href")
	}

	if visible, err := phoneLocator.IsVisible(); err == nil && visible {
		phone, _ = phoneLocator.TextContent()
	}

	if visible, err := plusCodeLocator.IsVisible(); err == nil && visible {
		pluscode, _ = plusCodeLocator.TextContent()
	}

	owner, _ = ownerLocator.IsVisible()
	owner = !owner

	if visible, err := menuLocator.IsVisible(); err == nil && visible {
		menu, _ = menuLocator.GetAttribute("href")
	}

	place.Category = category
	place.Address = address
	place.Website = website
	place.Phone = phone
	place.PlusCode = pluscode
	place.Owner = owner
	place.Menu = menu
	place.Image = image

	data := (*place.Page).URL()
	// time.Sleep(time.Millisecond * 500)
	data = data[strings.Index(data, "data="):]
	data = data[:strings.Index(data, "?")]
	data = strings.Split(data, ":")[1]

	lat := data[strings.Index(data, "!3d"):strings.Index(data, "!4d")][3:]
	long := data[strings.Index(data, "!4d"):][3:]
	long = long[:strings.Index(long, "!")]
	// println(coordinates)

	// coords := strings.Split(coordinates[1:], ",")
	place.Latitude, _ = strconv.ParseFloat(lat, 64)
	place.Longtitude, _ = strconv.ParseFloat(long, 64)
	place.Coordinates = fmt.Sprintf("@%s,%s", lat, long)
	// place.Coordinates = fmt.Sprintf("@%f,%f", place.Latitude, place.Longtitude)
	zone, err := tz.GetZone(tz.Point{
		Lat: place.Latitude,
		Lon: place.Longtitude,
	})

	if err == nil {
		place.Timezone = zone[0]
	}

	return nil

}

func (place *Place) ExtractEmails(browser *playwright.BrowserContext, timeout int, wg *sync.WaitGroup) error {

	defer (*wg).Done()
	if len(place.Website) == 0 {

		return nil
	}

	page, err := (*browser).NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err != nil {
		return err
	}

	if _, err := page.Goto(place.Website); err != nil {
		return err
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(float64(timeout) * 1000),
	})

	content, _ := page.Locator("html").First().InnerText()
	// fmt.Println(content)
	emailRegex := `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`
	emails := regexp.MustCompile(emailRegex).FindAllString(content, -1)
	ee := make([]string, 0)
	for _, email := range emails {
		if !slices.Contains(ee, email) {
			ee = append(ee, email)
		}
	}

	// Process the extracted emails as needed
	place.Emails = strings.Join(ee, ",")

	return nil
	// Rest of your code...
}
