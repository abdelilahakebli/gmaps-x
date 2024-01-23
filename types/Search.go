package types

import (
	"github.com/playwright-community/playwright-go"
	"github.com/schollz/progressbar/v3"
)

type Search struct {
	Query    string `json:"query" csv:"query"`
	Lang     string `json:"lang" csv:"lang"`
	Browser  *playwright.BrowserContext
	Page     playwright.Page
	Timeout  int
	ApiMode  bool
	Cooldown int
	Progress *progressbar.ProgressBar
	Limit    int
	Jobs     []*Job `json:"places" csv:"places"`
	Places   []*Place
}
