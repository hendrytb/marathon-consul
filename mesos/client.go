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
	OnEvent func(Task)
	apps    map[string]App
	tasks   map[string]Task
}

func NewClient(c *http.Client, baseUrl string) (*Client, error) {
	cl := &Client{c, baseUrl, func(Task) {}, make(map[string]App), make(map[string]Task)}

	apps, err := cl.getApps()
	if err != nil {
		return cl, err
	}

	for _, app := range apps.Apps {
		hc := app.HealthCheck.Get()
		if hc == nil {
			continue
		}

		cl.addApp(app.App)

		// register tasks
		for _, task := range app.Tasks {
			task.App = app.App
			cl.tasks[task.ID] = task
		}
	}

	return cl, nil
}

func (cl *Client) Tasks() map[string]Task {
	return cl.tasks
}

func (cl *Client) App(id string) (App, error) {
	if a, ok := cl.apps[id]; ok {
		return a, nil
	}

	return App{}, errors.New("App ID not found: " + id)
}

func (cl *Client) Subscribe() error {
	req, err := http.NewRequest(http.MethodGet, cl.baseUrl+"/v2/events", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "text/event-stream")

	res, err := cl.c.Do(req)
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
			return err
		}

		if len(line) < 7 {
			//log.Println(string(line))
			continue
		}

		if string(line[0:5]) != "data:" {
			//log.Println(string(line))
			continue
		}

		var e Event
		if err := json.Unmarshal(line[6:], &e); err != nil {
			log.Printf("error %v, data: %v\n", err, string(line[6:]))
			continue
		}

		switch e.Type {
		case "api_post_event":
			cl.addApp(e.App)

		case "status_update_event":
			h := cl.apps[e.AppID].HealthCheck.Get()
			if h != nil {
				t := Task{
					ID:    e.TaskID,
					Host:  e.Host,
					Ports: e.Ports,
					State: e.TaskState,
					App:   cl.apps[e.AppID],
				}

				switch t.State {
				case StateStaging, StateStarting:
					// Do nothing
				case StateRunning:
					cl.tasks[t.ID] = t
				case StateFinished, StateFailed, StateKilling, StateKilled, StateLost:
					delete(cl.tasks, t.ID)
				}

				cl.OnEvent(t)
			}
		}
	}
}

func (cl *Client) getApps() (Apps, error) {
	var a Apps
	req, err := http.NewRequest(http.MethodGet, cl.baseUrl+"/v2/apps?embed=apps.tasks", nil)
	if err != nil {
		return a, err
	}

	res, err := cl.c.Do(req)
	if err != nil {
		return a, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	err = json.NewDecoder(res.Body).Decode(&a)
	return a, err
}

func (cl *Client) addApp(app App) {
	if app.HealthCheck.Get() != nil {
		cl.apps[app.ID] = app
	}
}
