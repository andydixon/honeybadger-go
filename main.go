package main

import (
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pborman/getopt/v2"
	"honeybadger/randomdata"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func main() {
	targetUrl := getopt.StringLong("url", 'u', "", "URL to attack")
	getopt.Parse()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err := url.ParseRequestURI(*targetUrl)
	if err != nil {
		panic(err)
	} else {
		res, err := http.Get(*targetUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		iter := 0
		for {
			iter++

			// Find the review items
			doc.Find("form").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the band and title
				formMethod, _ := s.Attr("method")
				formTarget, _ := s.Attr("action")
				//fmt.Printf("Form method %s - target is %s\n", formMethod, formTarget)
				target, _ := url.Parse(*targetUrl)
				finalTarget, _ := target.Parse(formTarget)
				if iter == 1 {
					fmt.Println("Form Target (" + formMethod + "): " + finalTarget.String())
				}
				params := url.Values{}
				s.Find("input").Each(func(i int, element *goquery.Selection) {
					inputType, _ := element.Attr("type")
					inputValue, _ := element.Attr("value")
					inputName, _ := element.Attr("name")

					//title := s.Find("i").Text()
					if iter == 1 {
						fmt.Printf("Element %s is a %s - %s\n", inputName, inputType, inputValue)
					}
					if inputType == "password" {
						// Generate a random string as a password
						params.Add(inputName, randomdata.RandStringRunes(rand.Intn(4)+8))
					} else if inputType == "hidden" {
						// Generate a random string as a password
						params.Add(inputName, inputValue)
					} else if inputType == "checkbox" {
						// Generate a random string as a password
						params.Add(inputName, "")
					} else if inputType == "submit" || inputName == "" {
						//Looking for any submit button on the form target
						params.Add(inputType, inputValue)
					} else if inputValue == "" {
						// Meh, lets just put email addresses in everywhere for now
						params.Add(inputName, randomdata.Email())
					} else {
						//This would be for any hidden fields with data in already/
						params.Add(inputName, inputValue)
					}
				}) // End of each input field

				body := strings.NewReader(params.Encode())
				req, err := http.NewRequest(formMethod, finalTarget.String(), body)
				if err != nil {
					// handle err
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("User-Agent", randomdata.UserAgentString())
				fmt.Print("Hit " + strconv.Itoa(iter) + ": ")
				resp, _ := http.DefaultClient.Do(req)
				fmt.Print(strconv.Itoa(resp.StatusCode) + "(" + resp.Status + ") " + string(rune(27)) + "[K")
				defer resp.Body.Close()
			}) // End of each form
		}
	}

}
