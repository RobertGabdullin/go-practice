package main

import (
	"bufio"
	"fmt"
	"github.com/mailru/easyjson"
	"hw3/models"
	"io"
	"os"
	"regexp"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("@")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	reader := bufio.NewReader(file)

	users := make([]models.User, 0)
	for {
		line, eof := reader.ReadString('\n')

		if len(line) == 0 {
			break
		}

		user := &models.User{}
		err = easyjson.Unmarshal([]byte(line), user)
		if err != nil {
			panic(err)
		}
		users = append(users, *user)

		if eof != nil {
			break
		}

	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers

		for _, browserRaw := range browsers {
			browser := browserRaw

			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			browser := browserRaw

			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
