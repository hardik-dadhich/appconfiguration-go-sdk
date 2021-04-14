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
)

type Property struct {
	Name          string        `json:"name"`
	Property_id   string        `json:"property_id"`
	DataType      string        `json:"type"`
	Value         interface{}   `json:"value"`
	Segment_rules []SegmentRule `json:"segment_rules"`
}

func (f *Property) GetPropertyName() string {
	return f.Name
}
func (f *Property) GetPropertyId() string {
	return f.Property_id
}
func (f *Property) GetPropertyDataType() string {
	return f.DataType
}
func (f *Property) GetValue() interface{} {
	return f.Value
}
func (f *Property) GetSegmentRules() []SegmentRule {
	return f.Segment_rules
}
func (f *Property) GetCurrentValue(id string, identity map[string]interface{}) interface{} {
	log.Debug(messages.RETRIEVING_PROPERTY)
	if len(id) <= 0 {
		log.Error(messages.SET_IDENTITY_OBJECT_ID_ERROR)
		return nil
	}

	val := f.propertyEvaluation(id, identity)
	return getTypeCastedValue(val, f.GetPropertyDataType())
}
func (f *Property) propertyEvaluation(id string, identity map[string]interface{}) interface{} {

	var evaluatedSegmentId string = constants.DEFAULT_SEGMENT_ID
	defer func() { utils.GetMeteringInstance().RecordEvaluation("", f.GetPropertyId(), id, evaluatedSegmentId) }()

	log.Debug(messages.EVALUATING_PROPERTY)
	defer utils.GracefullyHandleError()

	if len(identity) < 0 {
		log.Debug(f.GetValue())
		return f.GetValue()
	}

	if len(f.GetSegmentRules()) > 0 {

		var rulesMap map[int]SegmentRule
		rulesMap = f.parseRules(f.GetSegmentRules())

		// sort the map elements as per ascending order of keys

		var keys []int
		for k, _ := range rulesMap {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		// after sorting , pick up each map element as per keys order
		for _, k := range keys {
			segmentRule := rulesMap[k]
			for _, rule := range segmentRule.GetRules() {
				for _, segmentKey := range rule.Segments {
					if f.evaluateSegment(string(segmentKey), identity) {
						evaluatedSegmentId = segmentKey
						if segmentRule.GetValue() == "$default" {
							log.Debug(messages.PROPERTY_VALUE)
							log.Debug(f.GetValue())
							return f.GetValue()
						} else {
							log.Debug(messages.PROPERTY_VALUE)
							log.Debug(segmentRule.GetValue())
							return segmentRule.GetValue()
						}
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
	log.Debug(messages.PARSING_PROPERTY_RULES)
	defer utils.GracefullyHandleError()
	var rulesMap map[int]SegmentRule
	rulesMap = make(map[int]SegmentRule)
	for _, rule := range segmentRules {
		rulesMap[rule.GetOrder()] = rule
	}
	log.Debug(rulesMap)
	return rulesMap
}
func (f *Property) evaluateSegment(segmentKey string, identity map[string]interface{}) bool {
	log.Debug(messages.EVALUATING_SEGMENTS)
	segment, ok := GetCacheInstance().SegmentMap[segmentKey]
	if ok {
		return segment.EvaluateRule(identity)
	}
	return false
}
