package helpers

import "strings"

func GoogleMapsFormatQueryToURLSearchFormat(query string, lang string) string {

	const BASE_URL = "https://www.google.com/maps/search/"

	query = strings.ToLower(query)
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, " ", "+")

	return BASE_URL + query + "?authuser=0&hl=" + lang + "&entry=ttu"

}
