package mesos

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Client struct {
	c       *http.Client
	baseUrl string
	OnEvent func(ev EventStatusUpdate)
}

func NewClient(c *http.Client, baseUrl string) *Client {
	return &Client{c, baseUrl, func(EventStatusUpdate) {}}
}

func (c *Client) List() ([]App, error) {
	return c.list("")
}

func (c *Client) App(id string) (App, error) {
	x, err := c.list(id)
	if err != nil {
		return App{}, err
	}
	if len(x) != 1 {
		return App{}, errors.New("id not found: " + id)
	}
	return x[0], nil
}

func (c *Client) Subscribe() error {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/events", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "text/event-stream")

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	reader := bufio.NewReader(res.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
		}

		if len(line) < 7 {
			//log.Println(string(line))
			continue
		}

		if string(line[0:5]) != "data:" {
			//log.Println(string(line))
			continue
		}

		var e EventStatusUpdate
		if err := json.Unmarshal(line[6:], &e); err != nil {
			log.Printf("error %v, data: %v\n", err, string(line[6:]))
		}
		switch e.Type {
		case "status_update_event":
			c.OnEvent(e)
		}
	}
	return nil
}

func (c *Client) list(id string) ([]App, error) {
	if id != "" {
		id = "&id=" + id
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/apps?embed=apps.tasks"+id, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	var a Apps
	err = json.NewDecoder(res.Body).Decode(&a)
	return a.Apps, err
}
