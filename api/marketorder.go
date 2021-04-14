package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//SetMarketOrder - Создание рыночной заявки
func SetMarketOrder(brokerAccountID string, figi string, lots int, operation OperationType) (*TPlacedOrder, error) {
	const reqName = "SetMarketOrder"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
		"figi":            []string{figi},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/orders/market-order?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Println("Can't create http request", reqName, err)
		return nil, err
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

	payload := struct {
		Lots      int           `json:"lots"`
		Operation OperationType `json:"operation"`
	}{Lots: lots, Operation: operation}

	data, _ := json.Marshal(payload)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	resp, err := hClient.Do(req)
	if err != nil {
		log.Println("Can't send request", reqName, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("bad response code", reqName, resp.Status, resp.Request.URL)
		str, _ := ioutil.ReadAll(req.Body)
		log.Println(reqName, string(str))
		return nil, errors.New(resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't read response", reqName, err)
		return nil, err
	}
	log.Println(reqName, string(respBody))

	var response OrderResponse

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Println("Can't unmarshal register response", string(respBody), err)
		return nil, err
	}
	log.Println(reqName)
	log.Println("   ", response.Payload)

	return &response.Payload, nil
}
