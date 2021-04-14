package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

//SetBalance - Выставление баланса по валютным позициям
func SetBalance(accountID string, balance float64, currency string) bool {
	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/sandbox/currencies/balance",
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create register http request: %s", err)
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	//	req.Header.Add("User-Agent", "MSIE/15.0") // добавляем заголовок User-Agent
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("brokerAccountId", accountID)

	payload := struct {
		Currency  string  `json:"currency"`
		Balance   float64 `json:"balance"`
		AccountID string  `json:"brokerAccountId,omitempty"`
	}{currency, balance, accountID}

	data, _ := json.Marshal(payload)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	resp, err := hClient.Do(req)
	if err != nil {
		log.Fatalf("Can't send setBalance request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("setBalance, bad response code '%s' from '%s'", resp.Status, resp.Request.URL)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Can't read setBalance response: %s", err)
	}

	var responce EmptyResponce

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}
	log.Println("setBalance", payload, responce)

	return true
}
