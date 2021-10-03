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
	"errors"
	"os"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/models"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/utils/log"
)

// AppConfiguration : Struct having init and configInstance.
type AppConfiguration struct {
	isInitialized                bool
	isInitializedConfig          bool
	configurationHandlerInstance *ConfigurationHandler
}

// ContextOptions : Struct having PersistentCacheDirectory path, BootstrapFile (ConfigurationFile) path and LiveConfigUpdateEnabled flag.
type ContextOptions struct {
	PersistentCacheDirectory string
	BootstrapFile            string
	ConfigurationFile        string
	LiveConfigUpdateEnabled  bool
}

var appConfigurationInstance *AppConfiguration

// OverrideServerHost : Override server host
var OverrideServerHost = ""

// var log = logrus.New()

// REGION_US_SOUTH : Dallas Region
const REGION_US_SOUTH = "us-south"

// REGION_EU_GB : London Region
const REGION_EU_GB = "eu-gb"

// REGION_AU_SYD : Sydney Region
const REGION_AU_SYD = "au-syd"

func init() {
	log.SetLogLevel("info")
}

// GetInstance : Get App Configuration Instance
func GetInstance() *AppConfiguration {
	log.Debug(messages.RetrieveingAppConfig)
	if appConfigurationInstance == nil {
		appConfigurationInstance = new(AppConfiguration)
	}
	return appConfigurationInstance
}

// Init : Init App Configuration Instance
func (ac *AppConfiguration) Init(region string, guid string, apikey string) {
	if len(region) == 0 || len(guid) == 0 || len(apikey) == 0 {
		if len(region) == 0 {
			log.Error(messages.RegionError)
		}
		if len(guid) == 0 {
			log.Error(messages.GUIDError)
		}
		if len(apikey) == 0 {
			log.Error(messages.ApikeyError)
		}
		return
	}
	ac.configurationHandlerInstance = GetConfigurationHandlerInstance()
	ac.configurationHandlerInstance.Init(region, guid, apikey)
	ac.isInitialized = true
}

// SetContext : Set Context
func (ac *AppConfiguration) SetContext(collectionID string, environmentID string, options ...ContextOptions) {
	log.Debug(messages.SettingContext)
	if !ac.isInitialized {
		log.Error(messages.CollectionIDError)
		return
	}
	if len(collectionID) == 0 {
		log.Error(messages.CollectionIDValueError)
		return
	}
	if len(environmentID) == 0 {
		log.Error(messages.EnvironmentIDValueError)
		return
	}
	switch len(options) {
	case 0:
		ac.configurationHandlerInstance.SetContext(collectionID, environmentID, ContextOptions{
			LiveConfigUpdateEnabled: true,
		})
	case 1:
		var temp = options[0]
		if len(temp.ConfigurationFile) > 0 && len(temp.BootstrapFile) == 0 {
			temp.BootstrapFile = temp.ConfigurationFile
			log.Info(messages.ContextOptionsParameterDeprecation)
		}
		if !temp.LiveConfigUpdateEnabled && len(temp.BootstrapFile) == 0 {
			log.Error(messages.BootstrapFileNotFoundError)
			return
		}
		ac.configurationHandlerInstance.SetContext(collectionID, environmentID, temp)
	default:
		log.Error(messages.IncorrectUsageOfContextOptions)
		return
	}
	ac.isInitializedConfig = true
	// If the cache is not having data make a blocking call and load the data in in-memory cache , else use the existing cache data and asynchronously update it.
	// This scenario can happen if the user uses setcontext second time in the code , in that case cache would not be empty.
	if ac.configurationHandlerInstance.cache == nil {
		ac.configurationHandlerInstance.loadData()
	} else {
		go ac.configurationHandlerInstance.loadData()
	}
}

// FetchConfigurations : Fetch Configurations
func (ac *AppConfiguration) FetchConfigurations() {
	if ac.isInitialized && ac.isInitializedConfig {
		go ac.configurationHandlerInstance.loadData()
	} else {
		log.Error(messages.CollectionInitError)
	}
}

// RegisterConfigurationUpdateListener : Register Configuration Update Listener
func (ac *AppConfiguration) RegisterConfigurationUpdateListener(fhl configurationUpdateListenerFunc) {
	if ac.isInitialized && ac.isInitializedConfig {
		ac.configurationHandlerInstance.registerConfigurationUpdateListener(fhl)
	} else {
		log.Error(messages.CollectionInitError)
	}
}

// GetFeature : Get Feature
func (ac *AppConfiguration) GetFeature(featureID string) (models.Feature, error) {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getFeature(featureID)
	}
	log.Error(messages.CollectionInitError)
	return models.Feature{}, errors.New(messages.ErrorInvalidFeatureAction)
}

// GetFeatures : Get Features
func (ac *AppConfiguration) GetFeatures() (map[string]models.Feature, error) {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getFeatures()
	}
	log.Error(messages.CollectionInitError)
	return nil, errors.New(messages.InitError)
}

// GetProperty : Get Property
func (ac *AppConfiguration) GetProperty(propertyID string) (models.Property, error) {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getProperty(propertyID)
	}
	log.Error(messages.CollectionInitError)
	return models.Property{}, errors.New(messages.ErrorInvalidPropertyAction)
}

// GetProperties : Get Properties
func (ac *AppConfiguration) GetProperties() (map[string]models.Property, error) {
	if ac.isInitializedConfig == true && ac.configurationHandlerInstance != nil {
		return ac.configurationHandlerInstance.getProperties()
	}
	log.Error(messages.CollectionInitError)
	return nil, errors.New(messages.InitError)
}

// EnableDebug : Enable Debug
func (ac *AppConfiguration) EnableDebug(enabled bool) {
	if enabled {
		os.Setenv("ENABLE_DEBUG", "true")
		log.SetLogLevel("debug")
	} else {
		os.Setenv("ENABLE_DEBUG", "false")
		log.SetLogLevel("info")
	}
}
