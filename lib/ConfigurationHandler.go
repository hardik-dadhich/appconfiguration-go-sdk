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
func (fh *ConfigurationHandler) Init(collectionId string, ac *AppConfiguration) {

	fh.collectionId = collectionId
	fh.urlBuilder = utils.GetInstance()
	fh.urlBuilder.Init(collectionId, ac.GetRegion(), ac.GetGuid(), OverrideServerHost)
	fh.appConfig = ac
	utils.GetMeteringInstance().Init(utils.GetInstance().GetMeteringUrl(), fh.appConfig.GetApiKey(), fh.appConfig.GetGuid(), collectionId)
	fh.configurationFile = ""
	fh.liveConfigUpdateEnabled = true
	fh.isInitialized = true
	fh.retryCount = 3
	fh.retryInterval = 600
}
func (fh *ConfigurationHandler) loadData() {
	if !fh.isInitialized {
		log.Error(messages.CONFIGURATION_HANDLER_INIT_ERROR)
	}
	log.Debug(messages.LOADING_DATA)
	log.Debug(messages.CHECK_CONFIGURATION_FILE_PROVIDED)
	if len(fh.configurationFile) > 0 {
		log.Debug(messages.CONFIGURATION_FILE_PROVIDED)
		fh.getFileData(fh.configurationFile)
	}
	log.Debug(messages.LOADING_CONFIGURATIONS)
	fh.loadConfigurations()
	log.Debug(messages.LIVE_UPDATE_CHECK)
	log.Debug(fh.liveConfigUpdateEnabled)
	if fh.liveConfigUpdateEnabled {
		go fh.FetchConfigurationData()
	}
}
func (fh *ConfigurationHandler) fetchConfigurationFromFile(configurationFilePath string, liveConfigUpdateEnabled bool) {
	fh.configurationFile = configurationFilePath
	fh.liveConfigUpdateEnabled = liveConfigUpdateEnabled
	log.Debug(messages.FETCH_FROM_CONFIGURATION_FILE + configurationFilePath)
	log.Debug(liveConfigUpdateEnabled)
	go fh.loadData()

}
func (fh *ConfigurationHandler) FetchConfigurationData() {
	log.Debug(messages.FETCH_CONFIGURATION_DATA)
	if fh.isInitialized {
		fh.fetchFromApi()
		fh.startWebSocket()
	}
}

func (fh *ConfigurationHandler) fetchFromApi() {
	log.Debug(messages.FETCH_FROM_API)
	if fh.isInitialized {
		fh.retryCount -= 1
		configUrl := fh.urlBuilder.GetConfigUrl()
		apiManager := utils.NewApiManagerInstance(configUrl, "GET", fh.appConfig.GetApiKey(), OverrideServerHost)
		response, statusCode := apiManager.ExecuteApiCall()
		if statusCode >= 200 && statusCode <= 299 {
			fh.writeServerFile(response)
		} else {
			if fh.retryCount > 0 {
				fh.fetchFromApi()
			} else {
				fh.retryCount = 3
				time.AfterFunc(time.Second*time.Duration(fh.retryInterval), func() {
					fh.fetchFromApi()
				})
			}
		}
		log.Debug(response)
	} else {
		log.Debug(messages.FETCH_FROM_API_SDK_INIT_ERROR)
	}
}

func (fh *ConfigurationHandler) startWebSocket() {
	log.Debug(messages.START_WEB_SOCKET)
	apiKey := fh.appConfig.GetApiKey()
	h := http.Header{"Authorization": []string{apiKey}}
	var err error
	if fh.socketConnection != nil {
		fh.socketConnection.Close()
	}
	fh.socketConnection, fh.socketConnectionResponse, err = websocket.DefaultDialer.Dial(fh.urlBuilder.GetWebSocketUrl(), h)
	if err != nil {
		log.Error(messages.WEB_SOCKET_CONNECT_ERR, err, fh.socketConnectionResponse.StatusCode)
		log.Info(messages.RETRY_WEB_SCOKET_CONNECT)
		go fh.startWebSocket()
	}
	// defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := fh.socketConnection.ReadMessage()
			log.Debug(string(message))
			if err != nil {
				log.Error(messages.WEBSOCKET_ERROR_READING_MESSAGE, err.Error())
				return
			}
			if string(message) != "test message" {
				log.Debug(messages.WEBSOCKET_RECEIVING_MESSAGE + string(message))
				fh.fetchFromApi()
			}
		}
	}()

}
func (fh *ConfigurationHandler) loadConfigurations() {
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
	fh.cache = models.GetCacheInstance()
}

func (fh *ConfigurationHandler) getFeatureActions(featureID string) models.Feature {
	fh.loadConfigurations()
	if fh.cache != nil && len(fh.cache.FeatureMap) > 0 {
		if val, ok := fh.cache.FeatureMap[featureID]; ok {
			return val
		} else {
			log.Error(messages.INVALID_FEATURE_ID, featureID)
			return models.Feature{}
		}
	} else {
		return models.Feature{}
	}
}
func (fh *ConfigurationHandler) getFeatures() map[string]models.Feature {
	if fh.cache == nil {
		return map[string]models.Feature{}
	}
	return fh.cache.FeatureMap
}
func (fh *ConfigurationHandler) getFeature(featureID string) models.Feature {
	if fh.cache != nil && len(fh.cache.FeatureMap) > 0 {
		if val, ok := fh.cache.FeatureMap[featureID]; ok {
			return val
		} else {
			return fh.getFeatureActions(featureID)
		}
	} else {
		return fh.getFeatureActions(featureID)
	}
}

func (fh *ConfigurationHandler) getPropertyActions(propertyID string) models.Property {
	fh.loadConfigurations()
	if fh.cache != nil && len(fh.cache.PropertyMap) > 0 {
		if val, ok := fh.cache.PropertyMap[propertyID]; ok {
			return val
		} else {
			log.Error(messages.INVALID_PROPERTY_ID, propertyID)
			return models.Property{}
		}
	} else {
		return models.Property{}
	}
}
func (fh *ConfigurationHandler) getProperties() map[string]models.Property {
	if fh.cache == nil {
		return map[string]models.Property{}
	}
	return fh.cache.PropertyMap
}
func (fh *ConfigurationHandler) getProperty(propertyID string) models.Property {
	if fh.cache != nil && len(fh.cache.PropertyMap) > 0 {
		if val, ok := fh.cache.PropertyMap[propertyID]; ok {
			return val
		} else {
			return fh.getPropertyActions(propertyID)
		}
	} else {
		return fh.getPropertyActions(propertyID)
	}
}

func (fh *ConfigurationHandler) registerConfigurationUpdateListener(fhl configurationUpdateListenerFunc) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(messages.CONFIGURATION_UPDATE_LISTENER_METHOD_ERROR)
		}
	}()
	if fh.isInitialized {
		fh.configurationUpdateListener = fhl
	} else {
		log.Error(messages.COLLECTION_ID_ERROR)
	}
}

func (fh *ConfigurationHandler) writeServerFile(content string) {
	if fh.liveConfigUpdateEnabled {
		fh.writeToFile(content)
	}
}
func (fh *ConfigurationHandler) writeToFile(content string) {
	utils.StoreFiles(content)
	fh.loadConfigurations()
	if fh.configurationUpdateListener != nil {
		fh.configurationUpdateListener()
	}
}

func (fh *ConfigurationHandler) getFileData(filePath string) {
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
	fh.writeToFile(string(out))
}
