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

//SetLimitOrder - Создание лимитной заявки
func SetLimitOrder(brokerAccountID string, figi string, lots int, price float64, operation OperationType) (*TPlacedOrder, error) {
	const reqName = "SetLimitOrder"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
		"figi":            []string{figi},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/orders/limit-order?"+q.Encode(),
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
		Price     float64       `json:"price"`
		Operation OperationType `json:"operation"`
	}{Lots: lots, Operation: operation, Price: price}

	data, _ := json.Marshal(payload)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	resp, err := hClient.Do(req)
	if err != nil {
		log.Println("Can't send request:", reqName, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("bad response code", reqName, resp.Status, resp.Request.URL)
		return nil, errors.New(resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Can't read %s response: %s", reqName, err)
		return nil, err
	}
	log.Println(reqName, string(respBody))

	var responce OrderResponse

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println("Can't unmarshal register response", string(respBody), err)
		return nil, err
	}

	log.Println(reqName, responce.Payload)

	if responce.Payload.Status == "Rejected" {
		return nil, errors.New(responce.Payload.Message)
	}

	return &responce.Payload, nil
}
