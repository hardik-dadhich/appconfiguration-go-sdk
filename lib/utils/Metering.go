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

	"sync"
	"time"

	messages "github.com/IBM/appconfiguration-go-sdk/lib/messages"

	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron"
)

type FeatureUsage struct {
	Feature_id      string `json:"feature_id"`
	Evaluation_time string `json:"evaluation_time"`
	Count           int64  `json:"count"`
}
type CollectionUsages struct {
	Collection_id string         `json:"collection_id"`
	Usages        []FeatureUsage `json:"usages"`
}

type featureMetric struct {
	count          int64
	evaluationTime string
}
type Metering struct {
	url          string
	apiKey       string
	collectionId string
	guid         string
	mu           sync.Mutex
	meteringData map[string]map[string]map[string]featureMetric //guid->collectionid->featureid
}

const SEND_INTERVAL = "10m"

var meteringInstance *Metering

func GetMeteringInstance() *Metering {
	log.Debug(messages.RETRIEVE_METERING_INSTANCE)
	if meteringInstance == nil {
		meteringInstance = &Metering{}
		// start sending metering data in the background
		log.Debug(messages.START_SENDING_METERING_DATA)
		c := cron.New()
		c.AddFunc("@every "+SEND_INTERVAL, meteringInstance.sendMetering)
		c.Start()

	}
	return meteringInstance
}

func (mt *Metering) Init(url string, apiKey string, guid string, collectionId string) {
	mt.url = url
	mt.apiKey = apiKey
	mt.guid = guid
	mt.collectionId = collectionId
}

func (mt *Metering) addMetering(guid string, collectionId string, featureId string) {
	log.Debug(messages.ADD_METERING)
	defer GracefullyHandleError()
	mt.mu.Lock()
	t := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	var fm featureMetric
	fm.evaluationTime = formattedTime
	fm.count = 1

	if _, ok := mt.meteringData[guid]; ok {
		guidVal := mt.meteringData[guid]
		if _, ok := guidVal[collectionId]; ok {
			collectionIdVal := guidVal[collectionId]
			if _, ok := collectionIdVal[featureId]; ok {
				featureIdVal := collectionIdVal[featureId]
				featureIdVal.evaluationTime = formattedTime
				featureIdVal.count = featureIdVal.count + 1
				collectionIdVal[featureId] = featureIdVal
			} else {
				mt.meteringData[guid][collectionId][featureId] = fm
			}
		} else {
			collectionMap := make(map[string]map[string]featureMetric)
			featureMap := make(map[string]featureMetric)
			featureMap[featureId] = fm
			collectionMap[collectionId] = featureMap
			mt.meteringData[guid] = collectionMap
		}
	} else {
		guidMap := make(map[string]map[string]map[string]featureMetric)
		collectionMap := make(map[string]map[string]featureMetric)
		featureMap := make(map[string]featureMetric)
		featureMap[featureId] = fm
		collectionMap[collectionId] = featureMap
		guidMap[guid] = collectionMap
		mt.meteringData = guidMap
	}
	mt.mu.Unlock()
}
func (mt *Metering) RecordEvaluation(featureId string) {
	log.Debug(messages.RECORD_EVAL)
	mt.addMetering(mt.guid, mt.collectionId, featureId)
}
func (mt *Metering) sendMetering() {
	log.Debug(messages.TEN_MIN_EXPIRY)
	defer GracefullyHandleError()
	log.Debug(mt.meteringData)
	mt.mu.Lock()
	if len(mt.meteringData) <= 0 {
		mt.mu.Unlock()
		return
	}
	sendMeteringData := mt.meteringData
	meteringDataMap := make(map[string]map[string]map[string]featureMetric)
	mt.meteringData = meteringDataMap

	mt.mu.Unlock()
	guidMap := make(map[string][]CollectionUsages)
	for guid, collectionMap := range sendMeteringData {
		var collectionUsageArray []CollectionUsages
		for collectionId, featureMap := range collectionMap {
			var usagesArray []FeatureUsage
			for featureId, val := range featureMap {
				var featureUsage FeatureUsage
				featureUsage.Feature_id = featureId
				featureUsage.Evaluation_time = val.evaluationTime
				featureUsage.Count = val.count
				usagesArray = append(usagesArray, featureUsage)
			}
			var collectionUsageElem CollectionUsages
			collectionUsageElem.Collection_id = collectionId
			collectionUsageElem.Usages = usagesArray
			collectionUsageArray = append(collectionUsageArray, collectionUsageElem)
		}
		guidMap[guid] = collectionUsageArray
	}

	for guid, val := range guidMap {
		for _, collectionUsage := range val {
			mt.sendToServer(guid, collectionUsage)
		}
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
