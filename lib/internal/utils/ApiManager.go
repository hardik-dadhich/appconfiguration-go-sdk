/**
 * (C) Copyright IBM Corp. 2021.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	messages "github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"

	"github.com/sirupsen/logrus"
)

const (
	X_REWRITE_DOMAIN = "X-REWRITE-DOMAIN"
	AUTHORIZATION    = "Authorization"
	APPLICATION_JSON = "application/json"
	CONTENT_TYPE     = "Content-Type"
)

type ApiManager struct {
	url    string
	method string
	req    *http.Request
	body   map[string]string
}

var log = logrus.New()

func init() {
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}
func NewApiManagerInstance(url string, method string, apikey string, overrideServerHost string) *ApiManager {

	var ap *ApiManager
	ap = new(ApiManager)
	ap.url = url
	ap.method = method
	ap.req, _ = http.NewRequest(method, url, nil)
	ap.req.Header.Add(AUTHORIZATION, apikey)
	ap.req.Header.Add(CONTENT_TYPE, APPLICATION_JSON)

	if len(overrideServerHost) > 0 {
		ap.req.Header.Add(X_REWRITE_DOMAIN, overrideServerHost)
	}
	return ap
}

func (ap *ApiManager) setRequestBody(body map[string]string) {
	ap.body = body
}

func (ap *ApiManager) ExecuteApiCall() (string, int) {
	log.Debug(messages.EXEC_API_CALL)
	client := &http.Client{}
	if len(ap.body) > 0 {
		jsonEncoded, _ := json.Marshal(ap.body)
		jsonString := string(jsonEncoded)
		ap.req.Body = ioutil.NopCloser(strings.NewReader(jsonString))
	}
	resp, err := client.Do(ap.req)
	if err != nil {
		log.Error(messages.API_CALL_ERROR, err)
		return "", resp.StatusCode
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Println(string([]byte(body)))
	}
	return string([]byte(body)), resp.StatusCode
}
