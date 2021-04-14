package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//
func GetOperations(brokerAccountID string, Figi string, From string, To string) OperationResponse {
	const reqName = "getOperations"

	q := url.Values{
		"from":            []string{From},
		"to":              []string{To},
		"figi":            []string{Figi},
		"brokerAccountId": []string{brokerAccountID},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/operations?"+q.Encode(),
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

	var responce OperationResponse

	err = json.Unmarshal(respBody, &responce)
	if err != nil {
		log.Fatalf("Can't unmarshal register response: '%s' \nwith error: %s", string(respBody), err)
	}
	//log.Println(reqName, responce)

	if len(responce.Payload.Operations) > 0 {
		log.Println(reqName)
		for i := range responce.Payload.Operations {
			log.Println("   ", responce.Payload.Operations[i])
		}
	} else {
		log.Println(reqName, "No Operation")
	}

	return responce
}
