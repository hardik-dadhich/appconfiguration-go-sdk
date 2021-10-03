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
	"net/http"
	"regexp"

	"github.com/IBM/go-sdk-core/v5/core"
)

// URLBuilder : URLBuilder struct
type URLBuilder struct {
	baseURL       string
	wsURL         string
	path          string
	service       string
	httpBase      string
	webSocketBase string
	events        string
	region        string
	guid          string
	iamURL        string
	authenticator core.Authenticator
}

var urlBuilderInstance *URLBuilder

// GetInstance : Get Instance
func GetInstance() *URLBuilder {
	if urlBuilderInstance == nil {
		urlBuilderInstance = &URLBuilder{
			baseURL:       ".apprapp.cloud.ibm.com",
			wsURL:         "/wsfeature",
			path:          "/feature/v1/instances/",
			service:       "/apprapp",
			httpBase:      "https://",
			webSocketBase: "wss://",
			events:        "/events/v1/instances/",
			region:        "",
			guid:          "",
			iamURL:        "https://iam.cloud.ibm.com",
		}
	}
	return urlBuilderInstance
}

// Init : Init
func (ub *URLBuilder) Init(collectionID string, environmentID string, region string, guid string, apikey string, overrideServerHost string) {
	ub.region = region
	ub.guid = guid
	if len(overrideServerHost) > 0 {
		ub.httpBase = overrideServerHost
		ub.iamURL = "https://iam.test.cloud.ibm.com"
		var compile, _ = regexp.Compile(`http([a-z]*)://`)
		ub.webSocketBase += compile.ReplaceAllString(overrideServerHost, "")
	} else {
		ub.httpBase += region
		ub.httpBase += ub.baseURL
		ub.webSocketBase += region
		ub.webSocketBase += ub.baseURL
	}
	ub.webSocketBase += ub.service + ub.wsURL + "?instance_id=" + guid + "&collection_id=" + collectionID + "&environment_id=" + environmentID
	// Create the authenticator.
	ub.authenticator = &core.IamAuthenticator{
		ApiKey: apikey,
		URL:    ub.iamURL,
	}
}

// GetBaseServiceURL returns base service url
func (ub *URLBuilder) GetBaseServiceURL() string {
	return ub.httpBase
}

// GetAuthenticator returns iam authenticator
func (ub *URLBuilder) GetAuthenticator() core.Authenticator {
	return ub.authenticator
}

// GetWebSocketURL returns web socket url
func (ub *URLBuilder) GetWebSocketURL() string {
	return ub.webSocketBase
}

// GetToken returns the string "Bearer <token>"
func (ub *URLBuilder) GetToken() string {
	req, _ := http.NewRequest("GET", "https://localhost", nil)
	var err error
	err = ub.authenticator.Authenticate(req)
	if err != nil {
		return ""
	}
	return req.Header.Get("Authorization")
}

// SetWebSocketURL : sets web socket url
func (ub *URLBuilder) SetWebSocketURL(webSocketURL string) {
	ub.webSocketBase = webSocketURL
}

// SetAuthenticator : assigns an authenticator to the url builder instance authenticator member variable.
func (ub *URLBuilder) SetAuthenticator(authenticator core.Authenticator) {
	ub.authenticator = authenticator
}
