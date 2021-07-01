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

package models

import (
	constants "github.com/IBM/appconfiguration-go-sdk/lib/internal/constants"
	messages "github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	utils "github.com/IBM/appconfiguration-go-sdk/lib/internal/utils"

	"sort"

	"github.com/IBM/appconfiguration-go-sdk/lib/internal/utils/log"
)

// Property : Property struct
type Property struct {
	Name         string        `json:"name"`
	PropertyID   string        `json:"property_id"`
	DataType     string        `json:"type"`
	Value        interface{}   `json:"value"`
	SegmentRules []SegmentRule `json:"segment_rules"`
}

// GetPropertyName : Get Property Name
func (f *Property) GetPropertyName() string {
	return f.Name
}

// GetPropertyID : Get Property Id
func (f *Property) GetPropertyID() string {
	return f.PropertyID
}

// GetPropertyDataType : Get Property Data Type
func (f *Property) GetPropertyDataType() string {
	return f.DataType
}

// GetValue : Get Value
func (f *Property) GetValue() interface{} {
	return f.Value
}

// GetSegmentRules : Get Segment Rules
func (f *Property) GetSegmentRules() []SegmentRule {
	return f.SegmentRules
}

// GetCurrentValue : Get Current Value
func (f *Property) GetCurrentValue(entityID string, entityAttributes map[string]interface{}) interface{} {
	log.Debug(messages.RetrievingProperty)
	if len(entityID) <= 0 {
		log.Error(messages.SetEntityObjectIDError)
		return nil
	}

	if f.isPropertyValid() {
		val := f.propertyEvaluation(entityID, entityAttributes)
		return getTypeCastedValue(val, f.GetPropertyDataType())
	}
	return nil
}

func (f *Property) isPropertyValid() bool {
	return !(f.Name == "" || f.PropertyID == "" || f.DataType == "" || f.Value == nil)
}

func (f *Property) propertyEvaluation(entityID string, entityAttributes map[string]interface{}) interface{} {

	var evaluatedSegmentID string = constants.DefaultSegmentID
	defer func() {
		utils.GetMeteringInstance().RecordEvaluation("", f.GetPropertyID(), entityID, evaluatedSegmentID)
	}()

	log.Debug(messages.EvaluatingProperty)
	defer utils.GracefullyHandleError()

	if len(entityAttributes) < 0 {
		log.Debug(f.GetValue())
		return f.GetValue()
	}

	if len(f.GetSegmentRules()) > 0 {

		var rulesMap map[int]SegmentRule
		rulesMap = f.parseRules(f.GetSegmentRules())

		// sort the map elements as per ascending order of keys

		var keys []int
		for k := range rulesMap {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		// after sorting , pick up each map element as per keys order
		for _, k := range keys {
			segmentRule := rulesMap[k]
			for _, rule := range segmentRule.GetRules() {
				for _, segmentKey := range rule.Segments {
					if f.evaluateSegment(string(segmentKey), entityAttributes) {
						evaluatedSegmentID = segmentKey
						if segmentRule.GetValue() == "$default" {
							log.Debug(messages.PropertyValue)
							log.Debug(f.GetValue())
							return f.GetValue()
						}
						log.Debug(messages.PropertyValue)
						log.Debug(segmentRule.GetValue())
						return segmentRule.GetValue()
					}
				}
			}
		}
	} else {
		return f.GetValue()
	}
	return f.GetValue()
}
func (f *Property) parseRules(segmentRules []SegmentRule) map[int]SegmentRule {
	log.Debug(messages.ParsingPropertyRules)
	defer utils.GracefullyHandleError()
	var rulesMap map[int]SegmentRule
	rulesMap = make(map[int]SegmentRule)
	for _, rule := range segmentRules {
		rulesMap[rule.GetOrder()] = rule
	}
	log.Debug(rulesMap)
	return rulesMap
}
func (f *Property) evaluateSegment(segmentKey string, entityAttributes map[string]interface{}) bool {
	log.Debug(messages.EvaluatingSegments)
	segment, ok := GetCacheInstance().SegmentMap[segmentKey]
	if ok {
		return segment.EvaluateRule(entityAttributes)
	}
	return false
}
