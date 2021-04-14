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

type Feature struct {
	Name           string        `json:"name"`
	Feature_id     string        `json:"feature_id"`
	DataType       string        `json:"type"`
	Enabled_value  interface{}   `json:"enabled_value"`
	Disabled_value interface{}   `json:"disabled_value"`
	Segment_rules  []SegmentRule `json:"segment_rules"`
	Enabled        bool          `json:"isEnabled"`
}

func (f *Feature) GetFeatureName() string {
	return f.Name
}
func (f *Feature) GetDisabledValue() interface{} {
	return f.Disabled_value
}
func (f *Feature) GetEnabledValue() interface{} {
	return f.Enabled_value
}
func (f *Feature) GetFeatureId() string {
	return f.Feature_id
}
func (f *Feature) GetFeatureDataType() string {
	return f.DataType
}
func (f *Feature) IsEnabled() bool {
	return f.Enabled
}
func (f *Feature) GetSegmentRules() []SegmentRule {
	return f.Segment_rules
}
func (f *Feature) GetCurrentValue(id string, identity map[string]interface{}) interface{} {
	log.Debug(messages.RETRIEVING_FEATURE)
	if len(id) <= 0 {
		log.Error(messages.SET_IDENTITY_OBJECT_ID_ERROR)
		return nil
	}

	val := f.featureEvaluation(id, identity)
	return getTypeCastedValue(val, f.GetFeatureDataType())
}
func (f *Feature) featureEvaluation(id string, identity map[string]interface{}) interface{} {

	var evaluatedSegmentId string = constants.DEFAULT_SEGMENT_ID
	defer func() { utils.GetMeteringInstance().RecordEvaluation(f.GetFeatureId(), "", id, evaluatedSegmentId) }()

	if f.IsEnabled() {
		log.Debug(messages.EVALUATING_FEATURE)
		defer utils.GracefullyHandleError()

		if len(identity) < 0 {
			log.Debug(f.GetEnabledValue())
			return f.GetEnabledValue()
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
								log.Debug(messages.FEATURE_VALUE)
								log.Debug(f.GetEnabledValue())
								return f.GetEnabledValue()
							} else {
								log.Debug(messages.FEATURE_VALUE)
								log.Debug(segmentRule.GetValue())
								return segmentRule.GetValue()
							}
						}
					}
				}
			}
		} else {
			return f.GetEnabledValue()
		}
		return f.GetEnabledValue()
	} else {
		return f.GetDisabledValue()
	}
}
func (f *Feature) parseRules(segmentRules []SegmentRule) map[int]SegmentRule {
	log.Debug(messages.PARSING_FEATURE_RULES)
	defer utils.GracefullyHandleError()
	var rulesMap map[int]SegmentRule
	rulesMap = make(map[int]SegmentRule)
	for _, rule := range segmentRules {
		rulesMap[rule.GetOrder()] = rule
	}
	log.Debug(rulesMap)
	return rulesMap
}
func (f *Feature) evaluateSegment(segmentKey string, identity map[string]interface{}) bool {
	log.Debug(messages.EVALUATING_SEGMENTS)
	segment, ok := GetCacheInstance().SegmentMap[segmentKey]
	if ok {
		return segment.EvaluateRule(identity)
	}
	return false
}
