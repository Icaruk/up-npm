package npm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchNpmRegistry(dependency string, token string) (map[string]interface{}, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://registry.npmjs.org/%s", dependency), nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package %s, status code: %d", dependency, resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result, nil

}
