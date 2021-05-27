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
	"net/http"
	"time"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/models"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/utils"

	"github.com/gorilla/websocket"
)

type configurationUpdateListenerFunc func()
type ConfigurationHandler struct {
	isInitialized               bool
	collectionId                string
	environmentId               string
	apikey                      string
	guid                        string
	region                      string
	urlBuilder                  *utils.UrlBuilder
	appConfig                   *AppConfiguration
	cache                       *models.Cache
	configurationUpdateListener configurationUpdateListenerFunc
	configurationFile           string
	liveConfigUpdateEnabled     bool
	retryCount                  int
	retryInterval               int64
	socketConnection            *websocket.Conn
	socketConnectionResponse    *http.Response
}

var configurationHandlerInstance *ConfigurationHandler

func GetConfigurationHandlerInstance() *ConfigurationHandler {
	if configurationHandlerInstance == nil {
		configurationHandlerInstance = new(ConfigurationHandler)
	}
	return configurationHandlerInstance
}
func (ch *ConfigurationHandler) Init(region, guid, apikey string) {
	ch.region = region
	ch.guid = guid
	ch.apikey = apikey
}
func (ch *ConfigurationHandler) SetContext(collectionId, environmentId, configurationFile string, liveConfigUpdateEnabled bool) {
	ch.collectionId = collectionId
	ch.environmentId = environmentId
	ch.urlBuilder = utils.GetInstance()
	ch.urlBuilder.Init(ch.collectionId, ch.environmentId, ch.region, ch.guid, OverrideServerHost)
	utils.GetMeteringInstance().Init(utils.GetInstance().GetMeteringUrl(), ch.apikey, ch.guid, environmentId, collectionId)
	ch.configurationFile = configurationFile
	ch.liveConfigUpdateEnabled = liveConfigUpdateEnabled
	ch.isInitialized = true
	ch.retryCount = 3
	ch.retryInterval = 600
}
func (ch *ConfigurationHandler) loadData() {
	if !ch.isInitialized {
		log.Error(messages.CONFIGURATION_HANDLER_INIT_ERROR)
	}
	log.Debug(messages.LOADING_DATA)
	log.Debug(messages.CHECK_CONFIGURATION_FILE_PROVIDED)
	if len(ch.configurationFile) > 0 {
		log.Debug(messages.CONFIGURATION_FILE_PROVIDED)
		ch.getFileData(ch.configurationFile)
	}
	log.Debug(messages.LOADING_CONFIGURATIONS)
	ch.loadConfigurations()
	log.Debug(messages.LIVE_UPDATE_CHECK)
	log.Debug(ch.liveConfigUpdateEnabled)
	if ch.liveConfigUpdateEnabled {
		ch.FetchConfigurationData()
	}
}
func (ch *ConfigurationHandler) FetchConfigurationData() {
	log.Debug(messages.FETCH_CONFIGURATION_DATA)
	if ch.isInitialized {
		ch.fetchFromApi()
		go ch.startWebSocket()
	}
}

func (ch *ConfigurationHandler) fetchFromApi() {
	log.Debug(messages.FETCH_FROM_API)
	if ch.isInitialized {
		ch.retryCount -= 1
		configUrl := ch.urlBuilder.GetConfigUrl()
		apiManager := utils.NewApiManagerInstance(configUrl, "GET", ch.apikey, OverrideServerHost)
		response, statusCode := apiManager.ExecuteApiCall()
		if statusCode >= 200 && statusCode <= 299 {
			ch.writeServerFile(response)
		} else {
			if ch.retryCount > 0 {
				log.Error(messages.CONFIG_API_ERROR)
				ch.fetchFromApi()
			} else {
				ch.retryCount = 3
				time.AfterFunc(time.Second*time.Duration(ch.retryInterval), func() {
					ch.fetchFromApi()
				})
			}
		}
	} else {
		log.Debug(messages.FETCH_FROM_API_SDK_INIT_ERROR)
	}
}

func (ch *ConfigurationHandler) startWebSocket() {
	defer utils.GracefullyHandleError()
	log.Debug(messages.START_WEB_SOCKET)
	apiKey := ch.apikey
	h := http.Header{"Authorization": []string{apiKey}}
	var err error
	if ch.socketConnection != nil {
		ch.socketConnection.Close()
	}
	ch.socketConnection, ch.socketConnectionResponse, err = websocket.DefaultDialer.Dial(ch.urlBuilder.GetWebSocketUrl(), h)
	if err != nil {
		if ch.socketConnectionResponse != nil {
			log.Error(messages.WEB_SOCKET_CONNECT_ERR, err, ch.socketConnectionResponse.StatusCode)
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
					log.Error(messages.WEBSOCKET_ERROR_READING_MESSAGE, err.Error())
					go ch.startWebSocket()
					return
				}
				if string(message) != "test message" {
					log.Debug(messages.WEBSOCKET_RECEIVING_MESSAGE + string(message))
					ch.fetchFromApi()
				}
			} else {
				go ch.startWebSocket()
				return
			}
		}
	}()

}
func (ch *ConfigurationHandler) loadConfigurations() {
	log.Debug(messages.LOADING_CONFIGURATIONS)
	defer utils.GracefullyHandleError()
	data := utils.ReadFiles("")
	configResponse := models.ConfigResponse{}
	err := json.Unmarshal(data, &configResponse)
	if err != nil {
		log.Error(messages.UNMARSHAL_JSON_ERR, err)
		return
	}
	log.Debug(configResponse)
	featureMap := make(map[string]models.Feature)
	for _, feature := range configResponse.Features {
		featureMap[feature.GetFeatureId()] = feature
	}

	propertyMap := make(map[string]models.Property)
	for _, property := range configResponse.Properties {
		propertyMap[property.GetPropertyId()] = property
	}

	segmentMap := make(map[string]models.Segment)
	for _, segment := range configResponse.Segments {
		segmentMap[segment.GetSegmentId()] = segment
	}

	// initialise cache
	log.Debug(messages.SET_IN_MEMORY_CACHE)
	models.SetCache(featureMap, propertyMap, segmentMap)
	ch.cache = models.GetCacheInstance()
}

func (ch *ConfigurationHandler) getFeatureActions(featureID string) (models.Feature, error) {
	ch.loadConfigurations()
	if ch.cache != nil && len(ch.cache.FeatureMap) > 0 {
		if val, ok := ch.cache.FeatureMap[featureID]; ok {
			return val, nil
		} else {
			log.Error(messages.INVALID_FEATURE_ID, featureID)
			return models.Feature{}, errors.New(messages.ERROR_INVALID_FEATURE_ID + featureID)
		}
	} else {
		return models.Feature{}, errors.New(messages.ERROR_INVALID_FEATURE_ID + featureID)
	}
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
		} else {
			return ch.getFeatureActions(featureID)
		}
	} else {
		return ch.getFeatureActions(featureID)
	}
}

func (ch *ConfigurationHandler) getPropertyActions(propertyID string) (models.Property, error) {
	ch.loadConfigurations()
	if ch.cache != nil && len(ch.cache.PropertyMap) > 0 {
		if val, ok := ch.cache.PropertyMap[propertyID]; ok {
			return val, nil
		} else {
			log.Error(messages.INVALID_PROPERTY_ID, propertyID)
			return models.Property{}, errors.New(messages.ERROR_INVALID_PROPERTY_ID + propertyID)
		}
	} else {
		return models.Property{}, errors.New(messages.ERROR_INVALID_PROPERTY_ID + propertyID)
	}
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
		} else {
			return ch.getPropertyActions(propertyID)
		}
	} else {
		return ch.getPropertyActions(propertyID)
	}
}

func (ch *ConfigurationHandler) registerConfigurationUpdateListener(chl configurationUpdateListenerFunc) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(messages.CONFIGURATION_UPDATE_LISTENER_METHOD_ERROR)
		}
	}()
	if ch.isInitialized {
		ch.configurationUpdateListener = chl
	} else {
		log.Error(messages.COLLECTION_ID_ERROR)
	}
}

func (ch *ConfigurationHandler) writeServerFile(content string) {
	if ch.liveConfigUpdateEnabled {
		ch.writeToFile(content)
	}
}
func (ch *ConfigurationHandler) writeToFile(content string) {
	utils.StoreFiles(content)
	ch.loadConfigurations()
	if ch.configurationUpdateListener != nil {
		ch.configurationUpdateListener()
	}
}

func (ch *ConfigurationHandler) getFileData(filePath string) {
	data := utils.ReadFiles(filePath)
	configResp := models.ConfigResponse{}
	err := json.Unmarshal(data, &configResp)
	if err != nil {
		log.Error(messages.UNMARSHAL_JSON_ERR, err)
		return
	}
	log.Debug(configResp)
	out, err := json.Marshal(configResp)
	if err != nil {
		log.Error(messages.MARSHAL_JSON_ERR, err)
		return
	}
	ch.writeToFile(string(out))
}
