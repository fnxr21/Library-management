package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Category struct {
	ID          string `json:"ID"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetCategories() ([]Category, error) {
	url := "http://localhost:6002/api/categorys"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("[ERROR] : Failed to fetch categories", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("[ERROR] : Non-OK HTTP status:", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch categories, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[ERROR] : Failed to read response body", err)
		return nil, err
	}

	var result struct {
		Result []Category `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("[ERROR] : Failed to unmarshal response", err)
		return nil, err
	}

	return result.Result, nil
}
