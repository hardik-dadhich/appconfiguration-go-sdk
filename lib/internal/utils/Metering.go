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
	"fmt"
	"github.com/IBM/appconfiguration-go-sdk/lib/internal/constants"

	"sync"
	"time"

	messages "github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"

	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron"
)

type Usages struct {
	Feature_id      string      `json:"feature_id,omitempty"`
	Property_id     string      `json:"property_id,omitempty"`
	Identity_id     string      `json:"identity_id"`
	Segment_id      interface{} `json:"segment_id"`
	Evaluation_time string      `json:"evaluation_time"`
	Count           int64       `json:"count"`
}
type CollectionUsages struct {
	Collection_id  string   `json:"collection_id"`
	Environment_id string   `json:"environment_id"`
	Usages         []Usages `json:"usages"`
}

type featureMetric struct {
	count          int64
	evaluationTime string
}
type Metering struct {
	url                  string
	apiKey               string
	collectionId         string
	environmentId        string
	guid                 string
	mu                   sync.Mutex
	meteringFeatureData  map[string]map[string]map[string]map[string]map[string]map[string]featureMetric //guid->environmentId->collectionId->featureId->identityId->segmentId
	meteringPropertyData map[string]map[string]map[string]map[string]map[string]map[string]featureMetric //guid->environmentId->collectionId->propertyId->identityId->segmentId
}

const SEND_INTERVAL = "10m"

var meteringInstance *Metering

func GetMeteringInstance() *Metering {
	log.Debug(messages.RETRIEVE_METERING_INSTANCE)
	if meteringInstance == nil {
		meteringInstance = &Metering{}
		guidFeatureMap := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
		guidPropertyMap := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
		meteringInstance.meteringFeatureData = guidFeatureMap
		meteringInstance.meteringPropertyData = guidPropertyMap
		// start sending metering data in the background
		log.Debug(messages.START_SENDING_METERING_DATA)
		c := cron.New()
		c.AddFunc("@every "+SEND_INTERVAL, meteringInstance.sendMetering)
		c.Start()

	}
	return meteringInstance
}

func (mt *Metering) Init(url string, apiKey string, guid string, environmentId string, collectionId string) {
	mt.url = url
	mt.apiKey = apiKey
	mt.guid = guid
	mt.environmentId = environmentId
	mt.collectionId = collectionId
}

func (mt *Metering) addMetering(guid string, environmentId string, collectionId string, identityId string, segmentId string, featureId string, propertyId string) {
	log.Debug(messages.ADD_METERING)
	defer GracefullyHandleError()
	mt.mu.Lock()
	t := time.Now().UTC()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	var fm featureMetric
	fm.evaluationTime = formattedTime
	fm.count = 1

	meteringData := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
	var modifyKey string
	if featureId != "" {
		meteringData = meteringInstance.meteringFeatureData
		modifyKey = featureId
	} else {
		meteringData = meteringInstance.meteringPropertyData
		modifyKey = propertyId
	}

	if _, ok := meteringData[guid]; ok {
		guidVal := meteringData[guid]
		if _, ok := guidVal[environmentId]; ok {
			envIdVal := guidVal[environmentId]
			if _, ok := envIdVal[collectionId]; ok {
				collectionIdVal := envIdVal[collectionId]
				if _, ok := collectionIdVal[modifyKey]; ok {
					modifyKeyVal := collectionIdVal[modifyKey]
					if _, ok := modifyKeyVal[identityId]; ok {
						identityIdVal := modifyKeyVal[identityId]
						if _, ok := identityIdVal[segmentId]; ok {
							segmentIdVal := identityIdVal[segmentId]
							segmentIdVal.evaluationTime = formattedTime
							segmentIdVal.count = segmentIdVal.count + 1
							identityIdVal[segmentId] = segmentIdVal
						} else {
							identityIdVal[segmentId] = fm
						}
					} else {
						segmentMap := make(map[string]featureMetric)
						segmentMap[segmentId] = fm
						modifyKeyVal[identityId] = segmentMap
					}
				} else {
					segmentMap := make(map[string]featureMetric)
					identityMap := make(map[string]map[string]featureMetric)
					segmentMap[segmentId] = fm
					identityMap[identityId] = segmentMap
					collectionIdVal[modifyKey] = identityMap
				}
			} else {
				segmentMap := make(map[string]featureMetric)
				identityMap := make(map[string]map[string]featureMetric)
				modifyKeyMap := make(map[string]map[string]map[string]featureMetric)
				segmentMap[segmentId] = fm
				identityMap[identityId] = segmentMap
				modifyKeyMap[modifyKey] = identityMap
				envIdVal[collectionId] = modifyKeyMap
			}
		} else {
			segmentMap := make(map[string]featureMetric)
			identityMap := make(map[string]map[string]featureMetric)
			modifyKeyMap := make(map[string]map[string]map[string]featureMetric)
			collectionMap := make(map[string]map[string]map[string]map[string]featureMetric)
			segmentMap[segmentId] = fm
			identityMap[identityId] = segmentMap
			modifyKeyMap[modifyKey] = identityMap
			collectionMap[collectionId] = modifyKeyMap
			guidVal[guid] = collectionMap
		}
	} else {
		segmentMap := make(map[string]featureMetric)
		identityMap := make(map[string]map[string]featureMetric)
		modifyKeyMap := make(map[string]map[string]map[string]featureMetric)
		collectionMap := make(map[string]map[string]map[string]map[string]featureMetric)
		environmentMap := make(map[string]map[string]map[string]map[string]map[string]featureMetric)
		segmentMap[segmentId] = fm
		identityMap[identityId] = segmentMap
		modifyKeyMap[modifyKey] = identityMap
		collectionMap[collectionId] = modifyKeyMap
		environmentMap[environmentId] = collectionMap
		meteringData[guid] = environmentMap
	}
	mt.mu.Unlock()
}
func (mt *Metering) RecordEvaluation(featureId string, propertyId string, identityId string, segmentId string) {
	log.Debug(messages.RECORD_EVAL)
	mt.addMetering(mt.guid, mt.environmentId, mt.collectionId, identityId, segmentId, featureId, propertyId)
}
func (mt *Metering) buildRequestBody(sendMeteringData map[string]map[string]map[string]map[string]map[string]map[string]featureMetric, guidMap map[string][]CollectionUsages, key string) {

	for guid, environmentMap := range sendMeteringData {
		var collectionUsageArray []CollectionUsages
		if _, ok := guidMap[guid]; !ok {
			guidMap[guid] = collectionUsageArray
		}
		for environmentId, collectionMap := range environmentMap {
			for collectionId, featureMap := range collectionMap {
				var usagesArray []Usages
				for featureId, identityMap := range featureMap {
					for identityId, segmentMap := range identityMap {
						for segmentId, val := range segmentMap {
							var usages Usages
							if key == "feature_id" {
								usages.Feature_id = featureId
							} else {
								usages.Property_id = featureId
							}
							if segmentId == constants.DEFAULT_SEGMENT_ID {
								usages.Segment_id = nil
							} else {
								usages.Segment_id = segmentId
							}
							usages.Identity_id = identityId
							usages.Evaluation_time = val.evaluationTime
							usages.Count = val.count
							usagesArray = append(usagesArray, usages)
						}
					}
				}
				var collectionUsageElem CollectionUsages
				collectionUsageElem.Collection_id = collectionId
				collectionUsageElem.Environment_id = environmentId
				collectionUsageElem.Usages = usagesArray
				collectionUsageArray = append(collectionUsageArray, collectionUsageElem)
			}
		}
		guidMap[guid] = append(guidMap[guid], collectionUsageArray...)
	}
}
func (mt *Metering) sendMetering() {
	log.Debug(messages.TEN_MIN_EXPIRY)
	defer GracefullyHandleError()
	log.Debug(mt.meteringFeatureData)
	log.Debug(mt.meteringPropertyData)
	mt.mu.Lock()
	if len(mt.meteringFeatureData) <= 0 && len(mt.meteringPropertyData) <= 0 {
		mt.mu.Unlock()
		return
	}
	sendFeatureData := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
	sendFeatureData = mt.meteringFeatureData
	meteringFeatureDataMap := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
	mt.meteringFeatureData = meteringFeatureDataMap

	sendPropertyData := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
	sendPropertyData = mt.meteringPropertyData
	meteringPropertyDataMap := make(map[string]map[string]map[string]map[string]map[string]map[string]featureMetric)
	mt.meteringPropertyData = meteringPropertyDataMap

	mt.mu.Unlock()

	guidMap := make(map[string][]CollectionUsages)

	if len(sendFeatureData) > 0 {
		mt.buildRequestBody(sendFeatureData, guidMap, "feature_id")
	}

	if len(sendPropertyData) > 0 {
		mt.buildRequestBody(sendPropertyData, guidMap, "property_id")
	}

	for guid, val := range guidMap {
		for _, collectionUsage := range val {
			var count int = len(collectionUsage.Usages)
			if count > constants.DEFAULT_USAGE_LIMIT {
				mt.sendSplitMetering(guid, collectionUsage, count)
			} else {
				mt.sendToServer(guid, collectionUsage)
			}
		}
	}

}
func (mt *Metering) sendSplitMetering(guid string, collectionUsages CollectionUsages, count int) {
	var lim int = 0
	subUsages := collectionUsages.Usages
	for lim <= count {
		var endIndex int
		if lim+constants.DEFAULT_USAGE_LIMIT >= count {
			endIndex = count
		} else {
			endIndex = lim + constants.DEFAULT_USAGE_LIMIT
		}
		var collectionUsageElem CollectionUsages
		collectionUsageElem.Collection_id = collectionUsages.Collection_id
		collectionUsageElem.Environment_id = collectionUsages.Environment_id
		for i := lim; i < endIndex; i++ {
			collectionUsageElem.Usages = append(collectionUsageElem.Usages, subUsages[i])
		}
		mt.sendToServer(guid, collectionUsageElem)
		lim = lim + constants.DEFAULT_USAGE_LIMIT
	}
}
func (mt *Metering) sendToServer(guid string, collectionUsages CollectionUsages) {
	log.Debug(messages.SEND_METERING_SERVER)
	log.Debug(collectionUsages)
	url := mt.url + guid + "/usage"
	encodedData, _ := json.Marshal(collectionUsages)

	log.Debug(string(encodedData))
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", mt.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(string(encodedData)).Post(url)
	log.Debug("Status code " + fmt.Sprint(resp.StatusCode()))
	if err != nil {
		log.Error(messages.SEND_METERING_SERVER_ERR, err)
		return
	}

}
