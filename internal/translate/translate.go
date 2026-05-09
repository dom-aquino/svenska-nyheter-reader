package translate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURL = "https://api-free.deepl.com/v2/translate"

var client = &http.Client{Timeout: 5 * time.Second}

type response struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// Word translates a single Swedish word to English using the DeepL free API.
func Word(word, apiKey string) (string, error) {
	data := url.Values{
		"text":        {word},
		"source_lang": {"SV"},
		"target_lang": {"EN"},
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepL status %d", resp.StatusCode)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Translations) == 0 {
		return "", fmt.Errorf("no translation returned")
	}
	return result.Translations[0].Text, nil
}
