package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

//GetOrders - Получение списка активных заявок
func GetOrders(brokerAccountID string) ([]Order, error) {
	const reqName = "getOrders"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/orders?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Println("Can't create http request:", reqName, err)
		return nil, err
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := hClient.Do(req)
	if err != nil {
		log.Println("Can't send request: ", reqName, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("bad response code", reqName, resp.Status, resp.Request.URL)
		return nil, errors.New("bad response code " + strconv.Itoa(resp.StatusCode))
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't read response:", reqName, err)
		return nil, err
	}
	//log.Println(reqName, string(respBody))

	var response OrdersResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Println("Can't unmarshal register response", err, string(respBody))
		return nil, err
	}
	/*
		if len(response.Payload) > 0 {
			log.Println(reqName, "(", brokerAccountID, ")")
			for i := range response.Payload {
				log.Println("   ", response.Payload[i])
			}
		} else {
			log.Println(reqName, "(", brokerAccountID, ")", "No orders")
		}
	*/
	return response.Payload, nil

}
