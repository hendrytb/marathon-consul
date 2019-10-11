package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Client struct {
	c       *http.Client
	baseUrl string
}

func NewClient(c *http.Client, baseUrl string) *Client {
	return &Client{c, baseUrl}
}

func (c *Client) Register(s Service) error {
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(s); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/v1/agent/service/register", b)
	if err != nil {
		return err
	}

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("error response: " + res.Status)
	}
	return nil
}

func (c *Client) DeRegister(id string) error {
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/v1/agent/service/deregister/"+id, nil)
	if err != nil {
		return err
	}

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("error response: " + res.Status)
	}
	return nil
}

func (c *Client) List() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v1/agent/services?filter=mesos+in+Tags", nil)
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

	var a map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&a); err != nil {
		return nil, err
	}
	s := make([]string, 0)
	for k := range a {
		s = append(s, k)
	}
	return s, nil
}
