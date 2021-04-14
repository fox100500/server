package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

//GetAvalableCurrencies - Получение списка валютных пар
func GetAvalableCurrencies() ([]MarketInstrument, error) {
	const reqName = "getAvalableCurrencies"

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/market/currencies",
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
		log.Println("Can't send request:", reqName, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("bad response code from ", reqName, resp.Status, resp.Request.URL)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't read response", reqName, err)
		return nil, err
	}

	var responce MarketInstrumentListResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println("Can't unmarshal register response", reqName, string(respBody), err)
		return nil, err
	}

	log.Println(reqName, "Total currencies", responce.Payload.Total)

	for i := range responce.Payload.Instruments {
		log.Println("   ",
			responce.Payload.Instruments[i].Name,
			responce.Payload.Instruments[i].FIGI,
			responce.Payload.Instruments[i].Ticker,
		)
	}

	return responce.Payload.Instruments, nil

}
