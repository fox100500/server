package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//SearchByFigi - Получение инструмента по FIGI
func SearchByFigi(figi string) SearchMarketInstrument {
	const reqName = "SearchByFigi"

	q := url.Values{
		"figi": []string{figi},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/market/search/by-figi?"+q.Encode(),
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
	//log.Println(reqName, string(respBody))

	var responce SearchMarketInstrumentResponse

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}
	log.Println(reqName, responce)

	return responce.Payload
}
