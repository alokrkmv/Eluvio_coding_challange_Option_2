package fetch_data

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

const BASE_URL = "https://challenges.qluv.io/items/"

// A basic get request handler
func get_request_handler(url string) (res interface{}, err error) {

	client := &http.Client{}
	final_url := BASE_URL + url
	req, _ := http.NewRequest("GET", final_url, nil)
	req.Header.Set("Authorization", "Y1JGMmR2RFpRc211MzdXR2dLNk1UY0w3WGpI")

	resp, err := client.Do(req)

	// Log and return the error in case the request fails
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// Close the request body to avoid any kind of leaks
	defer resp.Body.Close()
	// Check for the API response type .
	// If the API response is something else than 200 then exit
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		err = errors.New("API response status is not 200")
		log.Fatal(err)
		return nil, err
	}
	// Parse the response body and handle any error that occurs while parsing
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.New("somthing went wrong in reading the response body")
		log.Fatal(err)
		return nil, err
	}
	return responseData, err
}

func GetData(urls []string) (final_res map[string][]string, failed_ids []string) {

	final_res = make(map[string][]string)

	
	for _, url := range urls {
		res, err := get_request_handler(url)
		if err != nil {
			log.Fatal(err)
			failed_ids = append(failed_ids, url)
		}
		// Convert the response from []uint8 to string
		final_res[url] = append(final_res[url], string(res.([]uint8)))
	}
	return final_res, failed_ids

}
