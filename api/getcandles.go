package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//
func GetCandles(figi string, from string, to string, interval CandleInterval) ([]Candle, error) {
	const reqName = "getCandles"

	q := url.Values{
		"figi":     []string{figi},
		"from":     []string{from},
		"to":       []string{to},
		"interval": []string{string(interval)},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/market/candles?"+q.Encode(),
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
		log.Println("bad response code ", reqName, resp.Status, resp.Request.URL)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't read response:", reqName, err)
		return nil, err
	}
	var responce CandlesResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println("Can't unmarshal register response: ", reqName, string(respBody), err)
		return nil, err
	}
	//log.Println(reqName, responce)

	candles := responce.Payload.Candles

	log.Println(responce.Payload.Figi, responce.Payload.Interval, "count:", len(candles))
	if len(candles) > 0 {
		for i := range candles {
			log.Println(
				i,
				candles[i].Time,
				"l", candles[i].LowPrice,
				"h", candles[i].HighPrice,
				"o", candles[i].OpenPrice,
				"c", candles[i].ClosePrice,
				"v", candles[i].Volume,
			)
		}
		return responce.Payload.Candles, nil
	}
	return nil, errors.New("count zero")
}
