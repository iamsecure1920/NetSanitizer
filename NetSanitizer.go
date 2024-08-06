package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

func normalizeURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedURL.Fragment = "" // Remove fragment
	queryParams := parsedURL.Query()
	keys := make([]string, 0, len(queryParams))

	for key := range queryParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	sortedQueryParams := url.Values{}
	for _, key := range keys {
		sortedQueryParams[key] = queryParams[key]
	}

	parsedURL.RawQuery = sortedQueryParams.Encode()
	return parsedURL.String()
}

func getURLPath(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsedURL.Path
}

func hasIgnoredSuffix(urlPath string) bool {
	ignoredSuffixes := []string{"css", "js", "gif", "jpg", "png", "jpeg", "svg", "xml", "txt", "json", "ico", "webp", "otf", "ttf", "woff", "woff2", "eot", "swf", "zip", "pdf", "doc", "ppt", "docx", "xls", "xlsx", "ogg", "mp4", "mp3", "mov"}
	for _, suffix := range ignoredSuffixes {
		if strings.HasSuffix(urlPath, "."+suffix) {
			return true
		}
	}
	return false
}

func hasWebSuffix(urlPath string) bool {
	webSuffixes := []string{"htm", "html", "xhtml", "shtml", "jhtml", "cfm", "jsp", "jsf", "jspf", "wss", "action", "php", "php4", "php5", "py", "rb", "pl", "do", "xml", "rss", "cgi", "axd", "asx", "asmx", "ashx", "asp", "aspx", "dll"}
	for _, suffix := range webSuffixes {
		if strings.HasSuffix(urlPath, "."+suffix) {
			return true
		}
	}
	return false
}

func deduplicateURLs(inputFile string) {
	urlMap := make(map[string]string)
	basePathMap := make(map[string]string)

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("File not found: %s\n", inputFile)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}

		normalizedURL := normalizeURL(url)
		urlPath := getURLPath(normalizedURL)

		if hasIgnoredSuffix(urlPath) {
			continue
		}

		if hasWebSuffix(urlPath) {
			basePathMap[urlPath] = url
			continue
		}

		basePath := path.Dir(urlPath)

		if existingURL, exists := urlMap[basePath]; !exists {
			urlMap[basePath] = url
		} else {
			existingParams := strings.Split(existingURL, "?")
			currentParams := strings.Split(normalizedURL, "?")
			if len(existingParams) == 1 {
				urlMap[basePath] = url
			} else if len(currentParams) > 1 && len(strings.Split(currentParams[1], "&")) > len(strings.Split(existingParams[1], "&")) {
				urlMap[basePath] = url
			}
		}
	}

	for _, url := range basePathMap {
		fmt.Println(url)
	}

	for _, url := range urlMap {
		fmt.Println(url)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: <input_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	deduplicateURLs(inputFile)
}
