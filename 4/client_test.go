package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func less(a, b Person, orderField string) bool {
	switch orderField {
	case "Age":
		return a.Age < b.Age
	case "Name":
		return a.FirstName+" "+a.LastName < b.FirstName+" "+b.LastName
	case "Id":
		return a.ID < b.ID
	}
	return false
}

func sortResult(orderBy int, orderField string, list []Person) {
	if orderBy == 0 {
		return
	}
	sort.Slice(list, func(i, j int) bool {
		if orderBy == 1 {
			i, j = j, i
		}
		return less(list[i], list[j], orderField)
	})
}

func convertToUser(list []Person) []User {
	ans := make([]User, len(list))
	for i, person := range list {
		ans[i] = User{
			Id:     person.ID,
			Name:   person.FirstName + " " + person.LastName,
			Age:    person.Age,
			About:  person.About,
			Gender: person.Gender,
		}
	}
	return ans
}

func FreakServer(w http.ResponseWriter, req *http.Request) {
	trash := Person{
		ID:        0,
		GUID:      "",
		IsActive:  false,
		Balance:   "",
		Picture:   "",
		Age:       0,
		EyeColor:  "",
		FirstName: "",
		LastName:  "",
	}
	ans, err := json.Marshal(&trash)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ans)
}

func MissServer(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "Not found", http.StatusBadRequest)
}

func BadRequestServer(w http.ResponseWriter, req *http.Request) {
	res := SearchErrorResponse{
		Error: "unknown error",
	}
	ans, _ := json.Marshal(res)
	http.Error(w, string(ans), http.StatusBadRequest)
	return
}

func SearchServer(w http.ResponseWriter, req *http.Request) {
	query := req.FormValue("query")
	orderField := req.FormValue("order_field")
	orderBy := req.FormValue("order_by")
	limit := req.FormValue("limit")
	offset := req.FormValue("offset")

	token := req.Header.Get("AccessToken")
	if token != "token" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users := &UserList{}
	err = xml.NewDecoder(file).Decode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	resultList := make([]Person, 0)
	for _, user := range users.Users {
		name := user.FirstName + " " + user.LastName
		about := user.About

		if name == query || strings.Contains(about, query) {
			resultList = append(resultList, user)
		}
	}

	if orderField == "" {
		orderField = "Name"
	}

	if orderField != "Age" && orderField != "Name" && orderField != "Id" {
		res := SearchErrorResponse{
			Error: "ErrorBadOrderField",
		}
		ans, _ := json.Marshal(res)
		http.Error(w, string(ans), http.StatusBadRequest)
		return
	}

	order, err := strconv.Atoi(orderBy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid order_by: %s", orderBy), http.StatusBadRequest)
		return
	}

	sortResult(order, orderField, resultList)

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid limit: %s", limit), http.StatusBadRequest)
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid offset: %s", offset), http.StatusBadRequest)
		return
	}

	finalList := make([]Person, 0)
	for i := offsetInt; i < offsetInt+limitInt; i++ {
		if i >= len(resultList) {
			break
		}
		finalList = append(finalList, resultList[i])
	}

	ansList := convertToUser(finalList)

	ans, err := json.Marshal(ansList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(ans)

}

func InternalErrorServer(w http.ResponseWriter, req *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func TestClientCorrect(t *testing.T) {
	t.Parallel()

	tableTests := []struct {
		request  SearchRequest
		expected SearchResponse
	}{
		{
			request: SearchRequest{
				Limit:      30,
				Offset:     0,
				Query:      "Boyd Wolf",
				OrderField: "",
				OrderBy:    OrderByDesc,
			},
			expected: SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
		},
		{
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "enim",
				OrderField: "Id",
				OrderBy:    OrderByAsc,
			},
			expected: SearchResponse{
				Users: []User{
					{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
	}

	for i, tt := range tableTests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(SearchServer))
			defer ts.Close()

			client := SearchClient{
				AccessToken: "token",
				URL:         ts.URL,
			}

			res, err := client.FindUsers(tt.request)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Users, res.Users)
			assert.Equal(t, tt.expected.NextPage, res.NextPage)
		})
	}

}

func TimeoutServer(w http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
}

func TestClientError(t *testing.T) {
	t.Parallel()

	tableTests := []struct {
		handler http.HandlerFunc
		request SearchRequest
		token   string
	}{
		{
			token:   "token",
			handler: InternalErrorServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "Boyd Wolf",
				OrderField: "",
				OrderBy:    OrderByAsc,
			},
		},
		{
			token:   "token",
			handler: SearchServer,
			request: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "token",
			handler: SearchServer,
			request: SearchRequest{
				Limit:      0,
				Offset:     -1,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "bad token",
			handler: SearchServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "token",
			handler: SearchServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: ErrorBadOrderField,
				OrderBy:    0,
			},
		},
		{
			token:   "token",
			handler: TimeoutServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "token",
			handler: FreakServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "token",
			handler: MissServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
		{
			token:   "token",
			handler: BadRequestServer,
			request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
		},
	}

	for i, tt := range tableTests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := SearchClient{
				AccessToken: tt.token,
				URL:         ts.URL,
			}
			res, err := client.FindUsers(tt.request)

			assert.Nil(t, res)
			assert.Error(t, err)
		})
	}

}

func TestErrorClient(t *testing.T) {
	t.Parallel()

	client := SearchClient{
		AccessToken: "token",
		URL:         "http://127.0.0.1:1234",
	}

	requst := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsIs,
	}

	res, err := client.FindUsers(requst)

	assert.Nil(t, res)
	assert.Error(t, err)
}
