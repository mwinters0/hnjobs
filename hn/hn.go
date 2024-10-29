package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var baseURL = "https://hacker-news.firebaseio.com/v0" // to facilitate testing

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+url, nil)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// TODO: Everything in the HN API is an "item". We should fetch items and validate type before converting to Comment / etc.

// https://hacker-news.firebaseio.com/v0/user/whoishiring/submitted.json
func FetchSubmissions(ctx context.Context, user string) ([]int, error) {
	resp, err := fetch(ctx, "/user/"+user+"/submitted.json")
	if err != nil {
		return nil, fmt.Errorf("can't fetch: %v", err)
	}

	var i []int
	err = json.Unmarshal(resp, &i)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal: %v", err)
	}
	return i, nil
}

// https://hacker-news.firebaseio.com/v0/item/41709301.json
type Story struct {
	Id            int
	Kids          []int
	Time          int64
	GoTime        time.Time `json:"-"`
	Title         string
	FetchedTime   int64     `json:"-"`
	FetchedGoTime time.Time `json:"-"`
}

func FetchStory(ctx context.Context, id int) (*Story, error) {
	var s Story
	resp, err := fetch(ctx, "/item/"+strconv.Itoa(id)+".json")
	if err != nil {
		return &s, fmt.Errorf("can't fetch: %v", err)
	}
	err = json.Unmarshal(resp, &s)
	if err != nil {
		return &s, fmt.Errorf("can't unmarshal: %v", err)
	}
	s.FetchedGoTime = time.Now()
	s.FetchedTime = s.FetchedGoTime.Unix()
	s.GoTime = time.Unix(s.Time, 0)
	return &s, nil
}

// https://hacker-news.firebaseio.com/v0/item/41733646.json
type Comment struct {
	Id            int
	Parent        int
	Text          string
	Time          int64
	GoTime        time.Time `json:"-"`
	FetchedTime   int64     `json:"-"`
	FetchedGoTime time.Time `json:"-"`
}

func FetchComment(ctx context.Context, id int) (*Comment, error) {
	var c Comment
	resp, err := fetch(ctx, "/item/"+strconv.Itoa(id)+".json")
	if err != nil {
		return nil, fmt.Errorf("can't fetch: %v", err)
	}
	err = json.Unmarshal(resp, &c)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal: %v", err)
	}
	c.GoTime = time.Unix(c.Time, 0)
	now := time.Now()
	c.FetchedTime = now.Unix()
	c.FetchedGoTime = now
	return &c, nil
}
