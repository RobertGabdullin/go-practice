package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type UserDTO struct {
	Row []struct {
		Id        int    `xml:"id"`
		FirstName string `xml:"first_name"`
		LastName  string `xml:"last_name"`
		Age       int    `xml:"age"`
		Gender    string `xml:"gender"`
		About     string `xml:"about"`
	} `xml:"row"`
}

func validate(r *http.Request) *SearchRequest {

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		return nil
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		return nil
	}

	order, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		return nil
	}

	req := SearchRequest{
		Limit:      limit,
		Offset:     offset,
		OrderBy:    order,
		OrderField: r.FormValue("order_field"),
		Query:      r.FormValue("query"),
	}

	if req.OrderField == "" {
		req.OrderField = "Name"
	}

	orderBy := req.OrderBy < 2 && req.OrderBy > -2
	orderField := req.OrderField == "Name" || req.OrderField == "Id" || req.OrderField == "Age"
	if orderBy && orderField {
		return &req
	}
	return nil

}

func less(i, j int, field string, users []User) bool {
	switch field {
	case "Name":
		return users[i].Name < users[j].Name
	case "Id":
		return users[i].Id < users[j].Id
	case "Age":
		return users[i].Age < users[j].Age
	}
	return false
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("AccessToken")
	if token != "Test" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := validate(r)
	if req == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	records, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var users UserDTO
	err = xml.Unmarshal(records, &users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var ans []User
	for _, elem := range users.Row {
		name := elem.FirstName + " " + elem.LastName
		ans = append(ans, User{
			Name:   name,
			Id:     elem.Id,
			Age:    elem.Age,
			Gender: elem.Gender,
			About:  elem.About,
		})
	}

	pos := 0
	for i := 0; i < len(ans); i++ {
		if strings.Contains(ans[i].Name, req.Query) || strings.Contains(ans[i].About, req.Query) {
			ans[pos] = ans[i]
			pos++
		}
	}

	ans = ans[:pos]

	if req.OrderBy != 0 {
		sort.Slice(ans, func(i, j int) bool {
			if req.OrderBy == -1 {
				i, j = j, i
			}
			return less(i, j, req.OrderField, ans)
		})
	}

	if req.Offset >= len(ans) || req.Limit == 0 {
		w.Write([]byte("{}"))
		return
	}

	var end int
	if req.Offset+req.Limit > len(ans) {
		end = len(ans)
	} else {
		end = req.Offset + req.Limit
	}

	if req.Offset >= len(ans) || req.Limit == 0 {
		ans = []User{}
	} else {
		ans = ans[req.Offset:end]
	}

	out, err := json.Marshal(ans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(out)

}

type TestCase struct {
	request SearchRequest
	e       bool
}

func TestFindUsersOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		AccessToken: "Test",
		URL:         ts.URL,
	}

	tc := []TestCase{
		{
			request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "Nulla",
				OrderField: "Name",
				OrderBy:    1,
			},
			e: false,
		},
		{
			request: SearchRequest{
				Limit:      30,
				Offset:     0,
				Query:      "Nulla",
				OrderField: "Name",
				OrderBy:    1,
			},
			e: false,
		},
	}

	for _, req := range tc {
		_, err := client.FindUsers(req.request)
		if err != nil && !req.e {
			t.Errorf("Expected no error, but found error")
		}
		if err == nil && req.e {
			t.Errorf("Expected error, but no error found")
		}
	}

}

func TestFindUserError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		AccessToken: "Test",
		URL:         ts.URL,
	}

	tc := []TestCase{
		{
			request: SearchRequest{
				Limit:      -25,
				Offset:     0,
				Query:      "Nulla",
				OrderField: "Name",
				OrderBy:    1,
			},
			e: true,
		},
		{
			request: SearchRequest{
				Limit:      30,
				Offset:     -25,
				Query:      "Nulla",
				OrderField: "Name",
				OrderBy:    1,
			},
			e: true,
		},
	}

	for _, elem := range tc {
		_, err := client.FindUsers(elem.request)
		if err == nil {
			t.Errorf("Expected error, no error found")
		}
	}

}
