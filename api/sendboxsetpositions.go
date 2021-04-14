package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//SetSandboxPositions - Выставление баланса по инструментным позициям
func SetSandboxPositions(brokerAccountID string, figi string, balance float64) EmptyResponce {
	const reqName = "SetSandboxPositions"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/sandbox/positions/balance?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create %s http request: %s", reqName, err)
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

	payload := struct {
		Figi    string  `json:"figi"`
		Balance float64 `json:"balance"`
	}{Figi: figi, Balance: balance}

	data, _ := json.Marshal(payload)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	resp, err := hClient.Do(req)
	if err != nil {
		log.Fatalf("Can't send %s request: %s", reqName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s, bad response code '%s' from '%s'", reqName, resp.Status, resp.Request.URL)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Can't read %s response: %s", reqName, err)
	}
	//log.Println(reqName, string(respBody))

	var responce EmptyResponce

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}
	//log.Println(reqName, responce)
	log.Println(reqName)
	log.Println("   ", responce.Payload)

	return responce
}
