// Copyright (c) 2023. staking Inc. All rights reserved.
// Author icicd
// Create Time 2023/11/9

package nobot

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func httpPost(apiUrl string, formData map[string]string) (body []byte, err error) {

	var r http.Request

	r.ParseForm()
	for k, v := range formData {
		r.Form.Add(k, v)
	}

	payload := strings.NewReader(strings.TrimSpace(r.Form.Encode()))

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiUrl, payload)

	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad HTTP Response: %v, url:%v", resp.Status, apiUrl)
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}
