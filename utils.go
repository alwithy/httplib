package httplib

import (
	"encoding/base64"
	"net/url"
)

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func MergeUrlValues(v ...url.Values) (mergedValues url.Values) {
	mergedValues = make(url.Values)
	for _, items := range v {
		for key, values := range items {
			for _, val := range values {
				mergedValues.Add(key, val)
			}
		}
	}
	return
}
