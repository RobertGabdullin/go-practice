package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду с параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	lines := strings.Split(string(fileContents), "\n")

	users := make([]map[string]interface{}, 0)
	for _, line := range lines {
		user := make(map[string]interface{})
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}

			if ok = strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user["email"].(string), " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
