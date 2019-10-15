package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CertConsul syncs file based certificate (.crt & .key) into Consul Key/Value store.
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: certconsul <directory> <consul_url_1> <consul_url_2> ...\n\n" +
			"    directory   Directory where certificates are stored.\n" +
			"    consul_url  List of consul URL where certificates will be synced. (eg: http://127.0.0.1:8500/v1/kv/certs/)\n\n" +
			"This app is part of github.com/mataharimall/mesos-consul.")
		os.Exit(0)
	}

	certs := findCertificates(os.Args[1])
	for _, v := range os.Args[2:] {
		log.Println("Checking consul kv:", v)
		checkConsul(v, certs)
	}
	log.Println("Done")
}

// findCertificates returns the map of certificate filename and its concatenated content
// if you have mm.com.crt and mm.com.key, it would result {"mm.com": "content"}
// content is as showed by `cat mm.com.crt mm.com.key`
func findCertificates(directory string) map[string]string {
	res := make(map[string]string)
	files := make(map[string]string)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".issuer.crt") {
			return nil
		}
		if !strings.HasSuffix(path, ".crt") && !strings.HasSuffix(path, ".key") {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println("Error reading file:", err)
			return nil
		}
		files[info.Name()] = string(b)
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	for k, c := range files {
		if !strings.HasSuffix(k, ".crt") {
			continue
		}
		name := k[:len(k)-4]
		if files[name+".key"] == "" {
			continue
		}
		res[name] = c + files[name+".key"]
	}

	return res
}

// checkConsul check Consul key value store for given certificates
func checkConsul(consulURL string, cert map[string]string) {
	if !strings.HasSuffix(consulURL, "/") {
		consulURL += "/"
	}
	for k, v := range cert {
		url := consulURL + k + ".pem"
		s, err := consulGet(url + "?raw")
		if err != nil {
			log.Printf("failed to get %v: %v\n", url, err)
			continue
		}

		if v != s {
			log.Println("updating ", url)
			err := consulUpdate(url, v)
			if err != nil {
				fmt.Println("update failed:", err)
			}
		}
	}
}

func consulGet(url string) (string, error) {
	c := http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	if res == nil || res.Body == nil {
		return "", errors.New("empty response")
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	return string(b), err
}

func consulUpdate(url string, cert string) error {
	c := http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(cert))
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	return err
}
