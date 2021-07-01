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

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/constants"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/models"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/utils"
	"github.com/IBM/go-sdk-core/v5/core"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/utils/log"
	"github.com/gorilla/websocket"
)

type configurationUpdateListenerFunc func()

// ConfigurationHandler : Configuration Handler
type ConfigurationHandler struct {
	isInitialized               bool
	collectionID                string
	environmentID               string
	apikey                      string
	guid                        string
	region                      string
	urlBuilder                  *utils.URLBuilder
	appConfig                   *AppConfiguration
	cache                       *models.Cache
	configurationUpdateListener configurationUpdateListenerFunc
	configurationFile           string
	liveConfigUpdateEnabled     bool
	retryCount                  int
	retryInterval               int64
	socketConnection            *websocket.Conn
	socketConnectionResponse    *http.Response
	mu                          sync.Mutex
}

var configurationHandlerInstance *ConfigurationHandler

// GetConfigurationHandlerInstance : Get Configuration Handler Instance
func GetConfigurationHandlerInstance() *ConfigurationHandler {
	if configurationHandlerInstance == nil {
		configurationHandlerInstance = new(ConfigurationHandler)
	}
	return configurationHandlerInstance
}

// Init : Init App Configuration Instance
func (ch *ConfigurationHandler) Init(region, guid, apikey string) {
	ch.region = region
	ch.guid = guid
	ch.apikey = apikey
}

// SetContext : Set Context
func (ch *ConfigurationHandler) SetContext(collectionID, environmentID, configurationFile string, liveConfigUpdateEnabled bool) {
	ch.collectionID = collectionID
	ch.environmentID = environmentID
	ch.urlBuilder = utils.GetInstance()
	ch.urlBuilder.Init(ch.collectionID, ch.environmentID, ch.region, ch.guid, ch.apikey, OverrideServerHost)
	utils.GetMeteringInstance().Init(ch.guid, environmentID, collectionID)
	ch.configurationFile = configurationFile
	ch.liveConfigUpdateEnabled = liveConfigUpdateEnabled
	ch.isInitialized = true
	ch.retryCount = 3
	ch.retryInterval = 600
}
func (ch *ConfigurationHandler) loadData() {
	if !ch.isInitialized {
		log.Error(messages.ConfigurationHandlerInitError)
	}
	log.Debug(messages.LoadingData)
	log.Debug(messages.CheckConfigurationFileProvided)
	if len(ch.configurationFile) > 0 {
		log.Debug(messages.ConfigurationFileProvided)
		ch.getFileData(ch.configurationFile)
	}
	log.Debug(messages.LiveUpdateCheck)
	log.Debug(fmt.Sprint(ch.liveConfigUpdateEnabled))
	if ch.liveConfigUpdateEnabled {
		ch.FetchConfigurationData()
	}
}

// FetchConfigurationData : Fetch Configuration Data
func (ch *ConfigurationHandler) FetchConfigurationData() {
	log.Debug(messages.FetchConfigurationData)
	if ch.isInitialized {
		ch.fetchFromAPI()
		go ch.startWebSocket()
	}
}
func (ch *ConfigurationHandler) saveInCache(data string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	configResponse := models.ConfigResponse{}
	err := json.Unmarshal([]byte(data), &configResponse)
	if err != nil {
		log.Error(messages.UnmarshalJSONErr, err)
		return
	}
	log.Debug(configResponse)
	featureMap := make(map[string]models.Feature)
	for _, feature := range configResponse.Features {
		featureMap[feature.GetFeatureID()] = feature
	}

	propertyMap := make(map[string]models.Property)
	for _, property := range configResponse.Properties {
		propertyMap[property.GetPropertyID()] = property
	}

	segmentMap := make(map[string]models.Segment)
	for _, segment := range configResponse.Segments {
		segmentMap[segment.GetSegmentID()] = segment
	}
	log.Debug(messages.SetInMemoryCache)
	models.SetCache(featureMap, propertyMap, segmentMap)
	ch.cache = models.GetCacheInstance()
}
func (ch *ConfigurationHandler) fetchFromAPI() {
	if ch.isInitialized {
		ch.retryCount--
		builder := core.NewRequestBuilder(core.GET)
		builder.AddQuery("environment_id", ch.environmentID)
		pathParamsMap := map[string]string{
			"guid":          ch.guid,
			"collection_id": ch.collectionID,
		}
		_, err := builder.ResolveRequestURL(ch.urlBuilder.GetBaseServiceURL(), `/apprapp/feature/v1/instances/{guid}/collections/{collection_id}/config`, pathParamsMap)
		if err != nil {
			return
		}
		builder.AddHeader("Accept", "application/json")
		builder.AddHeader("User-Agent", constants.UserAgent)
		response := utils.GetAPIManagerInstance().Request(builder)
		if response != nil && response.StatusCode >= 200 && response.StatusCode <= 299 {
			if ch.liveConfigUpdateEnabled {
				jsonData, _ := json.Marshal(response.Result)
				ch.saveInCache(string(jsonData))
				if ch.configurationUpdateListener != nil {
					ch.configurationUpdateListener()
				}
			}
		} else {
			if ch.retryCount > 0 {
				log.Error(messages.ConfigAPIError)
				ch.fetchFromAPI()
			} else {
				ch.retryCount = 3
				time.AfterFunc(time.Second*time.Duration(ch.retryInterval), func() {
					ch.fetchFromAPI()
				})
			}
		}
	} else {
		log.Debug(messages.FetchFromAPISdkInitError)
	}
}

func (ch *ConfigurationHandler) startWebSocket() {
	defer utils.GracefullyHandleError()
	log.Debug(messages.StartWebSocket)
	h := http.Header{"Authorization": []string{ch.urlBuilder.GetToken()}}
	var err error
	if ch.socketConnection != nil {
		ch.socketConnection.Close()
	}
	ch.socketConnection, ch.socketConnectionResponse, err = websocket.DefaultDialer.Dial(ch.urlBuilder.GetWebSocketURL(), h)
	if err != nil {
		if ch.socketConnectionResponse != nil {
			log.Error(messages.WebSocketConnectErr, err, ch.socketConnectionResponse.StatusCode)
		}
		go ch.startWebSocket()
		return
	}
	// defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if ch.socketConnection != nil {
				_, message, err := ch.socketConnection.ReadMessage()
				log.Debug(string(message))
				if err != nil {
					log.Error(messages.WebsocketErrorReadingMessage, err.Error())
					go ch.startWebSocket()
					return
				}
				if string(message) != "test message" {
					log.Debug(messages.WebsocketReceivingMessage + string(message))
					ch.fetchFromAPI()
				}
			} else {
				go ch.startWebSocket()
				return
			}
		}
	}()

}

func (ch *ConfigurationHandler) getFeatureActions(featureID string) (models.Feature, error) {
	if ch.cache != nil && len(ch.cache.FeatureMap) > 0 {
		if val, ok := ch.cache.FeatureMap[featureID]; ok {
			return val, nil
		}
		log.Error(messages.InvalidFeatureID, featureID)
		return models.Feature{}, errors.New(messages.ErrorInvalidFeatureID + featureID)
	}
	return models.Feature{}, errors.New(messages.ErrorInvalidFeatureID + featureID)
}
func (ch *ConfigurationHandler) getFeatures() map[string]models.Feature {
	if ch.cache == nil {
		return map[string]models.Feature{}
	}
	return ch.cache.FeatureMap
}
func (ch *ConfigurationHandler) getFeature(featureID string) (models.Feature, error) {
	if ch.cache != nil && len(ch.cache.FeatureMap) > 0 {
		if val, ok := ch.cache.FeatureMap[featureID]; ok {
			return val, nil
		}
		return ch.getFeatureActions(featureID)
	}
	return ch.getFeatureActions(featureID)

}

func (ch *ConfigurationHandler) getPropertyActions(propertyID string) (models.Property, error) {
	if ch.cache != nil && len(ch.cache.PropertyMap) > 0 {
		if val, ok := ch.cache.PropertyMap[propertyID]; ok {
			return val, nil
		}
		log.Error(messages.InvalidPropertyID, propertyID)
		return models.Property{}, errors.New(messages.ErrorInvalidPropertyID + propertyID)
	}
	return models.Property{}, errors.New(messages.ErrorInvalidPropertyID + propertyID)
}
func (ch *ConfigurationHandler) getProperties() map[string]models.Property {
	if ch.cache == nil {
		return map[string]models.Property{}
	}
	return ch.cache.PropertyMap
}
func (ch *ConfigurationHandler) getProperty(propertyID string) (models.Property, error) {
	if ch.cache != nil && len(ch.cache.PropertyMap) > 0 {
		if val, ok := ch.cache.PropertyMap[propertyID]; ok {
			return val, nil
		}
		return ch.getPropertyActions(propertyID)
	}
	return ch.getPropertyActions(propertyID)
}

func (ch *ConfigurationHandler) registerConfigurationUpdateListener(chl configurationUpdateListenerFunc) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(messages.ConfigurationUpdateListenerMethodError)
		}
	}()
	if ch.isInitialized {
		ch.configurationUpdateListener = chl
	} else {
		log.Error(messages.CollectionIDError)
	}
}

func (ch *ConfigurationHandler) getFileData(filePath string) {
	data := utils.ReadFiles(filePath)
	configResp := models.ConfigResponse{}
	err := json.Unmarshal(data, &configResp)
	if err != nil {
		log.Error(messages.UnmarshalJSONErr, err)
		return
	}
	log.Debug(fmt.Sprint(configResp))
	out, err := json.Marshal(configResp)
	if err != nil {
		log.Error(messages.MarshalJSONErr, err)
		return
	}
	ch.saveInCache(string(out))
	if ch.configurationUpdateListener != nil {
		ch.configurationUpdateListener()
	}
}
