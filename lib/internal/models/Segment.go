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
	messages "github.com/IBM/appconfiguration-go-sdk/lib/internal/messages"
	utils "github.com/IBM/appconfiguration-go-sdk/lib/internal/utils"
)

// log "github.com/sirupsen/logrus"

type Segment struct {
	Name       string `json:"name"`
	Segment_id string `json:"segment_id"`
	Rules      []Rule `json:"rules"`
}

func (s *Segment) GetName() string {
	return s.Name
}
func (s *Segment) GetSegmentId() string {
	return s.Segment_id
}
func (s *Segment) GetRules() []Rule {
	return s.Rules
}
func (s *Segment) EvaluateRule(identity map[string]interface{}) bool {
	log.Debug(messages.EVAL_SEGMENT_RULE)
	defer utils.GracefullyHandleError()
	for _, rule := range s.GetRules() {
		if !rule.EvaluateRule(identity) {
			return false
		}
	}
	return true
}
