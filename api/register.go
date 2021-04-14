package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

//
func Register(accountType AccountType) []UserAccount {
	const reqName = "Register"

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/sandbox/register",
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create %s http request: %s", reqName, err)
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	//	req.Header.Add("User-Agent", "MSIE/15.0") // добавляем заголовок User-Agent
	req.Header.Add("Authorization", "Bearer "+token)

	payload := struct {
		AccountType AccountType `json:"brokerAccountType"`
	}{accountType}

	data, _ := json.Marshal(payload)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	//log.Println(req)

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
		log.Fatalf("Can't read register response: %s", err)
	}
	//log.Println(reqName, string(respBody))

	var responce UserAccountsResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}

	log.Println(reqName, responce)

	return responce.PayLoad.Accounts
}
