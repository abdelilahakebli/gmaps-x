package helpers

import (
	"github.com/playwright-community/playwright-go"
)

func WaitForLoadStateIdle(page *playwright.Page, timeout int) error {
	return (*page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(float64(timeout) * 1000),
	})
}
