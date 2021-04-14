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
	"os"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/models"

	"github.com/sirupsen/logrus"
)

type AppConfiguration struct {
	region                       string
	guid                         string
	apikey                       string
	isInitialized                bool
	isInitializedConfig          bool
	configurationHandlerInstance *ConfigurationHandler
}

var appConfigurationInstance *AppConfiguration
var OverrideServerHost = ""
var log = logrus.New()
var REGION_US_SOUTH = "us-south"
var REGION_EU_GB = "eu-gb"
var REGION_AU_SYD = "au-syd"

func init() {
	log.SetLevel(logrus.InfoLevel)
}
func GetInstance() *AppConfiguration {
	log.Debug(messages.RETRIEVEING_APP_CONFIG)
	if appConfigurationInstance == nil {
		appConfigurationInstance = new(AppConfiguration)
	}
	return appConfigurationInstance
}

func (ac *AppConfiguration) Init(region string, guid string, apikey string) {
	if len(region) == 0 || len(guid) == 0 || len(apikey) == 0 {
		if len(region) == 0 {
			log.Error(messages.REGION_ERROR)
		}
		if len(guid) == 0 {
			log.Error(messages.APIKEY_ERROR)
		}
		if len(apikey) == 0 {
			log.Error(messages.GUID_ERRROR)
		}
		return
	}
	ac.region = region
	ac.guid = guid
	ac.apikey = apikey
	ac.isInitialized = true
}

func (ac *AppConfiguration) SetCollectionId(collectionId string) {
	log.Debug(messages.SETTING_COLLECTION_ID)
	if !ac.isInitialized {
		log.Error(messages.COLLECTION_ID_ERROR)
		return
	}
	if len(collectionId) == 0 {
		log.Error(messages.COLLECTION_ID_VALUE_ERROR)
		return
	}

	ac.configurationHandlerInstance = GetConfigurationHandlerInstance()
	ac.configurationHandlerInstance.Init(collectionId, ac)
	ac.isInitializedConfig = true
	go ac.configurationHandlerInstance.loadData()
}

func (ac *AppConfiguration) FetchConfigurations() {
	if ac.isInitialized && ac.isInitializedConfig {
		go ac.configurationHandlerInstance.loadData()
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
	}
}

func (ac *AppConfiguration) FetchConfigurationFromFile(configurationFile string, liveConfigUpdateEnabled bool) {
	if !ac.isInitialized || !ac.isInitializedConfig {
		log.Error(messages.COLLECTION_ID_ERROR)
		return
	}
	if !liveConfigUpdateEnabled && len(configurationFile) == 0 {
		log.Error(messages.CONFIGURATION_FILE_NOT_FOUND_ERROR)
		return
	}
	ac.configurationHandlerInstance.fetchConfigurationFromFile(configurationFile, liveConfigUpdateEnabled)
}

func (ac *AppConfiguration) RegisterConfigurationUpdateListener(fhl configurationUpdateListenerFunc) {
	if ac.isInitialized && ac.isInitializedConfig {
		ac.configurationHandlerInstance.registerConfigurationUpdateListener(fhl)
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
	}
}

func (ac *AppConfiguration) GetFeature(featureId string) models.Feature {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getFeature(featureId)
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
		return models.Feature{}
	}
}
func (ac *AppConfiguration) GetFeatures() map[string]models.Feature {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getFeatures()
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
		return nil
	}
}
func (ac *AppConfiguration) GetProperty(propertyId string) models.Property {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getProperty(propertyId)
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
		return models.Property{}
	}
}
func (ac *AppConfiguration) GetProperties() map[string]models.Property {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getProperties()
	} else {
		log.Error(messages.COLLECTION_INIT_ERROR)
		return nil
	}
}

func (ac *AppConfiguration) GetRegion() string {
	return ac.region
}
func (ac *AppConfiguration) GetGuid() string {
	return ac.guid
}
func (ac *AppConfiguration) GetApiKey() string {
	return ac.apikey
}
func (ac *AppConfiguration) EnableDebug(enabled bool) {
	if enabled {
		log.SetLevel(logrus.DebugLevel)
		os.Setenv("ENABLE_DEBUG", "true")
	} else {
		log.SetLevel(logrus.InfoLevel)
		os.Setenv("ENABLE_DEBUG", "false")
	}
}
