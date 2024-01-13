package npm

import (
	"fmt"
	"net/http"
)

func FetchNpmRegistry(dependency string) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", dependency))
	if err != nil {
		fmt.Println(err)
	}

	return resp, err
}
