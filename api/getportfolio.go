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

//GetPortfolio - Получение портфеля клиента
func GetPortfolio(brokerAccountID string) ([]PositionBalance, error) {
	const reqName = "getPortfolio"

	q := url.Values{
		"brokerAccountId": []string{brokerAccountID},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/portfolio?"+q.Encode(),
		nil,
	)
	if err != nil {
		log.Println(reqName, "Can't create http request:", err)
		return nil, err
	}
	// добавляем заголовки
	req.Header.Add("Accept", "text/html") // добавляем заголовок Accept
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := hClient.Do(req)
	if err != nil {
		log.Println(reqName, "Can't send %s request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(reqName, " bad response code", resp.Status, "from ", resp.Request.URL)
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return nil, errors.New("too many requests")
		default:
			return nil, errors.New("bad responce " + strconv.Itoa(resp.StatusCode))
		}
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(reqName, "Can't read response:", err)
		return nil, err
	}
	//log.Println(reqName, string(respBody))

	var responce PortfolioResponse
	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Println(reqName, "Can't unmarshal register response", string(respBody), "err:", err)
		return nil, err
	}

	//	log.Println(reqName, responce)

	if len(responce.Payload.Positions) > 0 {
		/*
			log.Println(reqName, "(", brokerAccountID, ")")
			for i := range responce.Payload.Positions {
				log.Println("   ",
					responce.Payload.Positions[i].Ticker,
					responce.Payload.Positions[i].AveragePositionPrice.Currency,
					responce.Payload.Positions[i].AveragePositionPrice.Value,
					responce.Payload.Positions[i].AveragePositionPriceNoNkd.Currency,
					responce.Payload.Positions[i].AveragePositionPriceNoNkd.Value,
					responce.Payload.Positions[i].Balance,
					responce.Payload.Positions[i].Blocked,
					responce.Payload.Positions[i].ExpectedYield.Currency,
					responce.Payload.Positions[i].ExpectedYield.Value,
					responce.Payload.Positions[i].FIGI,
					responce.Payload.Positions[i].ISIN,
					responce.Payload.Positions[i].InstrumentType,
					responce.Payload.Positions[i].Lots,
					responce.Payload.Positions[i].Name,
				)
			}
		*/
	} else {
		log.Println(reqName, "(", brokerAccountID, ")", "No Positions")
	}

	return responce.Payload.Positions, nil
}
