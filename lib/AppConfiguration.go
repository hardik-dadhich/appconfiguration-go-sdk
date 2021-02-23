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
	region                 string
	guid                   string
	apikey                 string
	isInitialized          bool
	isInitializedFeature   bool
	featureHandlerInstance *FeatureHandler
}

var appConfigurationInstance *AppConfiguration
var OverrideServerHost = ""
var log = logrus.New()
var REGION_US_SOUTH = "us-south"
var REGION_EU_GB = "eu-gb"

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

	ac.featureHandlerInstance = GetFeatureHandlerInstance()
	ac.featureHandlerInstance.Init(collectionId, ac)
	ac.isInitializedFeature = true
	go ac.featureHandlerInstance.loadData()
}

func (ac *AppConfiguration) FetchFeatureData() {
	if ac.isInitialized && ac.isInitializedFeature {
		go ac.featureHandlerInstance.loadData()
	} else {
		log.Error(messages.COLLECTION_SUB_ERROR)
	}
}

func (ac *AppConfiguration) FetchFromFeatureFile(featureFilePath string, enableLiveUpdate bool) {
	if !ac.isInitialized || !ac.isInitializedFeature {
		log.Error(messages.COLLECTION_ID_ERROR)
		return
	}
	if !enableLiveUpdate && len(featureFilePath) == 0 {
		log.Error(messages.FEATURE_FILE_NOT_FOUND_ERROR)
		return
	}
	ac.featureHandlerInstance.fetchFromFeatureFile(featureFilePath, enableLiveUpdate)
}

func (ac *AppConfiguration) RegisterFeaturesUpdateListener(fhl featureUpdateListenerFunc) {
	if ac.isInitialized && ac.isInitializedFeature {
		ac.featureHandlerInstance.registerFeaturesUpdateListener(fhl)
	} else {
		log.Error(messages.COLLECTION_SUB_ERROR)
	}
}

func (ac *AppConfiguration) GetFeature(featureId string) models.Feature {
	if ac.isInitializedFeature == true && ac.featureHandlerInstance != nil {
		return ac.featureHandlerInstance.getFeature(featureId)
	} else {
		log.Error(messages.COLLECTION_SUB_ERROR)
		return models.Feature{}
	}
}
func (ac *AppConfiguration) GetFeatures() map[string]models.Feature {
	if ac.isInitializedFeature == true && ac.featureHandlerInstance != nil {
		return ac.featureHandlerInstance.getFeatures()
	} else {
		log.Error(messages.COLLECTION_SUB_ERROR)
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
