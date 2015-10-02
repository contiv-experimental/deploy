package nethooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

/*
func httpGet(url string, jdata interface{}) error {

	r, err := http.Get(url)
	if err != nil {
		return err
	}

	switch {
	case r.StatusCode == int(404):
		return errors.New("Page not found!")
	case r.StatusCode == int(403):
		return errors.New("Access denied!")
	case r.StatusCode != int(200):
		log.Errorf("GET Status '%s' status code %d \n", r.Status, r.StatusCode)
		return errors.New("unkown error")
	}

	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(response, jdata); err != nil {
		return err
	}

	return nil
}
*/

func httpPost(url string, jdata interface{}) error {
	buf, err := json.Marshal(jdata)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(buf)

	log.Debugf("Posting url %s: %v \n", url, jdata)
	return nil

	r, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	switch {
	case r.StatusCode == int(404):
		return errors.New("Page not found!")
	case r.StatusCode == int(403):
		return errors.New("Access denied!")
	case r.StatusCode != int(200):
		log.Errorf("POST Status '%s' status code %d \n", r.Status, r.StatusCode)
		return errors.New("unkown error")
	}

	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	log.Infof("POST response: %s ", string(response))

	return nil
}

/*
  func httpDelete(url string) error {

    r, err := http.Get(url)

    if err != nil {
      return err
    }

    switch {
    case r.StatusCode == int(404):
      return errors.New("Page not found!")
    case r.StatusCode == int(403):
      return errors.New("Access denied!")
    case r.StatusCode != int(200):
      fmt.Printf("DELETE Status '%s' status code %d \n", r.Status, r.StatusCode)
      return errors.New("unkown error")
    }

    response, err := ioutil.ReadAll(r.Body)
    if err != nil {
      return err
    }
    fmt.Println(string(response))

    return nil
  }
*/
