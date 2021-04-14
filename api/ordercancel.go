package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//OrderCancel - Отмена заявки
func OrderCancel(brokerAccountID string, orderID string) EmptyResponce {
	const reqName = "OrderCancel"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
		"orderId":         []string{orderID},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/orders/cancel?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Fatalf("Can't create %s http request: %s", reqName, err)
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

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
	log.Println(reqName, string(respBody))

	var responce EmptyResponce

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}

	log.Println(reqName)
	log.Println("   ", responce.Payload)

	return responce
}
