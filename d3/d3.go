package d3

import (
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

// Whole sdfjhb
type Whole struct {
	URL   string
	Links []string
	IMGS  []string
	JSS   []string
	CSSS  []string
}

// Response sdfjkhn
type Response struct {
	URL    string   `json:"url"`
	Links  []string `json:"links"`
	Static []string `json:"static"`
}

// Static sdfkjn
func (w Whole) Static() []string {
	return append(w.IMGS, append(w.CSSS, w.JSS...)...)
}

// CrawlURL sdfkjn
func CrawlURL(urlToCrawl string) (map[string]Response, error) {
	whole, err := CrawlURLSingle(urlToCrawl)
	if err != nil {
		panic(err)
	}
	result := make(map[string]Response)
	var mutex = &sync.Mutex{}
	result[urlToCrawl] = Response{
		URL:    whole.URL,
		Links:  whole.Links,
		Static: whole.Static(),
	}
	var wg sync.WaitGroup
	var anyerror error
	wg.Add(len(whole.Links))
	for index, link := range whole.Links {
		go func(index int, url string) {
			defer wg.Done()
			wh, err := CrawlURLSingle(url)
			if err != nil {
				anyerror = err
				return
			}
			mutex.Lock()
			result[url] = Response{
				URL:    wh.URL,
				Links:  wh.Links,
				Static: wh.Static(),
			}
			mutex.Unlock()
		}(index, link)
	}
	wg.Wait()
	if anyerror != nil {
		return result, anyerror
	}
	return result, nil
}

// CrawlURLSingle sdfkjn
func CrawlURLSingle(urlToCrawl string) (Whole, error) {
	whole := Whole{URL: urlToCrawl}
	urlToGet, err := url.Parse(urlToCrawl)
	if err != nil {
		return whole, err
	}
	content, err := getURLContent(urlToGet.String())
	if err != nil {
		return whole, err

	}
	var wg sync.WaitGroup
	var anyerror error
	wg.Add(4)
	go func() {
		defer wg.Done()
		links, err := parseByRegex(urlToGet, content, "<a.*?href=\"(.*?)\"")
		if err != nil {
			anyerror = err
			return
		}
		result := []string{}
		for _, link := range links {
			if !containsTwitterORForumORIRC(link) {
				if !Contains(result, link) {
					result = append(result, link)
				}
			}
		}
		whole.Links = result
	}()
	go func() {
		defer wg.Done()
		jss, err := parseByRegex(urlToGet, content, "<script .*?src=\"(.*?)\"")
		if err != nil {
			anyerror = err
			return
		}
		result := []string{}
		for _, link := range jss {
			if !Contains(result, link) {
				result = append(result, link)
			}
		}
		whole.JSS = result
	}()
	go func() {
		defer wg.Done()
		csss, err := parseByRegex(urlToGet, content, "<link .*?href=\"(.*?)\"")
		if err != nil {
			anyerror = err
			return
		}
		result := []string{}
		for _, link := range csss {
			if !Contains(result, link) {
				result = append(result, link)
			}
		}
		whole.CSSS = result
	}()
	go func() {
		defer wg.Done()
		images, err := parseByRegex(urlToGet, content, "<img.*?src=\"(.*?)\"")
		if err != nil {
			anyerror = err
			return
		}
		result := []string{}
		for _, link := range images {
			if !Contains(result, link) {
				result = append(result, link)
			}
		}
		whole.IMGS = result
	}()
	wg.Wait()
	if anyerror != nil {
		return whole, anyerror
	}
	return whole, nil
}
func containsTwitterORForumORIRC(url string) bool {
	if strings.Contains(url, "twitter") || strings.Contains(url, "forum") || strings.Contains(url, "irc://") {
		return true
	}
	return false
}

// Contains kjsdnf
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func getURLContent(urlToGet string) (string, error) {
	resp, err := http.Get(urlToGet)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return html.UnescapeString(string(content)), err
}

func parseByRegex(urlToGet *url.URL, content, regex string) ([]string, error) {
	imgs := []string{}
	matches := regexp.MustCompile(regex).FindAllStringSubmatch(content, -1)
	for _, val := range matches {
		imgURL, err := url.Parse(val[1])
		if err != nil {
			return imgs, err
		}
		if imgURL.IsAbs() {
			imgs = append(imgs, imgURL.String())
		} else {
			imgs = append(imgs, urlToGet.Scheme+"://"+urlToGet.Host+imgURL.String())
		}
	}
	return imgs, nil
}
