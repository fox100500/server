package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//GetCurrencies - Получение валютных активов клиента
func GetCurrencies(brokerAccountID string) ([]CurrencyPosition, error) {
	const reqName = "getCurrencies"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/portfolio/currencies?"+q.Encode(),
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

	var responce PortfolioCurrenciesResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println("Can't unmarshal register response", reqName, string(respBody), err)
		return nil, err
	}

	/*
		log.Println(reqName, "(", brokerAccountID, ")")
		for i := range responce.PayLoad.Currencies {
			log.Println("   ", responce.PayLoad.Currencies[i].Balance, responce.PayLoad.Currencies[i].TCurrence)
		}
	*/

	return responce.PayLoad.Currencies, nil
}
