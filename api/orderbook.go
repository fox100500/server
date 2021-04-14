package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

//
func GetOrderBookOld(Figi string, Depth int) (*OrderBookOld, error) {
	const reqName = "getOrderBook"

	q := url.Values{
		"depth": []string{strconv.Itoa(Depth)},
		"figi":  []string{Figi},
	}

	req, err := http.NewRequest(
		"GET",
		apiURL+"/market/orderbook?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Println("Can't create http request", reqName, err)
		return nil, err
	}

	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := hClient.Do(req)
	if err != nil {
		log.Println("Can't send request", reqName, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("bad response code", reqName, resp.Status, resp.Request.URL)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't read response", reqName, err)
		return nil, err
	}

	var responce OrderBookResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println("Can't unmarshal register response", reqName, string(respBody), err)
		return nil, err
	}

	/*
		log.Println(reqName, responce.Payload.FIGI, responce.Payload.TradeStatus)
		if len(responce.Payload.Asks) > 0 {
			for i := range responce.Payload.Asks {
				log.Println("asks ", responce.Payload.Asks[i].Price, "(", responce.Payload.Asks[i].Quantity, ")", "bids ", responce.Payload.Bids[i].Price, "(", responce.Payload.Bids[i].Quantity, ")")
			}
		}
	*/

	return &responce.Payload, nil
}
