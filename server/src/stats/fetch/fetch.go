package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	noContent       = 204
	badRequest      = 400
	tooManyRequests = 429
)

type Uuid_to_username struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func FetchUsernameFromUUID(uuid string) (Uuid_to_username, error) {
	var mojang_response Uuid_to_username

	res, err := http.Get("https://api.mojang.com/user/profile/" + uuid)
	if err != nil {
		return mojang_response, err
	}

	if res.StatusCode != 200 {
		log.Println("Error: received status code " + fmt.Sprintf("%d", res.StatusCode) + " from Mojang API")
		switch res.StatusCode {
		case noContent:
			return mojang_response, errors.New("no content")
		case badRequest:
			return mojang_response, errors.New("bad request")
		case tooManyRequests:
			return mojang_response, errors.New("too many requests")
		}

	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return mojang_response, err
	}

	err = json.Unmarshal(body, &mojang_response)
	if err != nil {
		return mojang_response, err
	}

	mojang_response.Id = mojangApiToUuid(mojang_response.Id)

	return mojang_response, nil
}

func mojangApiToUuid(undashed string) string {
	var dashPositions = [...]int{8, 13, 18, 23}
	for _, v := range dashPositions {
		undashed = undashed[:v] + "-" + undashed[v:]
	}
	return undashed
}
