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

import "regexp"

type UrlBuilder struct {
	baseUrl       string
	wsUrl         string
	path          string
	service       string
	httpBase      string
	webSocketBase string
	reWriteDomain string
	events        string
	region        string
	guid          string
}

var urlBuilderInstance *UrlBuilder

func GetInstance() *UrlBuilder {
	if urlBuilderInstance == nil {
		urlBuilderInstance = &UrlBuilder{
			baseUrl:       ".apprapp.cloud.ibm.com",
			wsUrl:         "/wsfeature",
			path:          "/feature/v1/instances/",
			service:       "/apprapp",
			httpBase:      "https://",
			webSocketBase: "wss://",
			events:        "/events/v1/instances/",
			region:        "",
			guid:          "",
		}
	}
	return urlBuilderInstance
}

func (ub *UrlBuilder) Init(collectionId string, environmentId string, region string, guid string, overrideServerHost string) {
	ub.region = region
	ub.guid = guid
	if len(overrideServerHost) > 0 {
		ub.httpBase = overrideServerHost
		var compile, _ = regexp.Compile(`http([a-z]*)://`)
		ub.webSocketBase += compile.ReplaceAllString(overrideServerHost, "")
		ub.reWriteDomain = overrideServerHost
	} else {
		ub.httpBase += region
		ub.httpBase += ub.baseUrl
		ub.webSocketBase += region
		ub.webSocketBase += ub.baseUrl
		ub.reWriteDomain = ""
	}
	ub.httpBase += ub.service + ub.path + guid + "/collections/" + collectionId + "/config?environment_id=" + environmentId
	ub.webSocketBase += ub.service + ub.wsUrl + "?instance_id=" + guid + "&collection_id=" + collectionId + "&environment_id=" + environmentId
}
func (ub *UrlBuilder) GetConfigUrl() string {
	return ub.httpBase
}

func (ub *UrlBuilder) GetWebSocketUrl() string {
	return ub.webSocketBase
}
func (ub *UrlBuilder) GetMeteringUrl() string {
	base := "https://" + ub.region + ub.baseUrl + ub.service
	if len(ub.reWriteDomain) > 0 {
		base = ub.reWriteDomain + ub.service
	}
	return base + ub.events
}
