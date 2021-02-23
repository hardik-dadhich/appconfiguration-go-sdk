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
	"reflect"
	"strconv"
	"strings"

	utils "github.com/IBM/appconfiguration-go-sdk/lib/utils"
)

type Rule struct {
	values         []interface{} `json:"values"`
	operator       string
	attribute_Name string
}

func (r *Rule) GetValues() []interface{} {
	return r.values
}
func (r *Rule) GetOperator() string {
	return r.operator
}
func (r *Rule) GetAttributeName() string {
	return r.attribute_Name
}
func (r *Rule) operatorCheck(key interface{}, value interface{}) bool {

	var result bool = false

	if key == nil || value == nil {
		return result
	}

	switch r.GetOperator() {
	case "endsWith":
		result = strings.HasSuffix(key.(string), value.(string))
		break
	case "startsWith":
		result = strings.HasPrefix(key.(string), value.(string))
		break
	case "contains":
		result = strings.Contains(key.(string), value.(string))
		break
	case "is":
		if isNumber(key) && !(isNumber(value)) {
			// compare number
			if reflect.TypeOf(value).String() == "string" {
				key = key.(float64)
				floatVal, _ := strconv.ParseFloat((value.(string)), 64)
				result = (key == floatVal)
			}
		} else if !isNumber(key) && isNumber(value) {
			// compare string
			if reflect.TypeOf(key).String() == "string" {
				floatKey, _ := strconv.ParseFloat(key.(string), 64)
				result = (floatKey == value.(float64))
			}
		} else {
			result = (key == value)
		}
		break
	case "greaterThan":
		if isNumber(key) && isNumber(value) {
			result = key.(float64) > value.(float64)
		} else if isString(key) || isString(value) {

			if isString(key) {
				key, _ = strconv.ParseFloat(key.(string), 64)
			}
			if isString(value) {
				value, _ = strconv.ParseFloat(value.(string), 64)
			}
			result = key.(float64) > value.(float64)
		}
		break
	case "lesserThan":
		if isNumber(key) && isNumber(value) {
			result = key.(float64) < value.(float64)
		} else if isString(key) || isString(value) {

			if isString(key) {
				key, _ = strconv.ParseFloat(key.(string), 64)
			}
			if isString(value) {
				value, _ = strconv.ParseFloat(value.(string), 64)
			}
			result = key.(float64) < value.(float64)
		}
		break
	case "greaterThanEquals":
		if isNumber(key) && isNumber(value) {
			result = key.(float64) >= value.(float64)
		} else if isString(key) || isString(value) {

			if isString(key) {
				key, _ = strconv.ParseFloat(key.(string), 64)
			}
			if isString(value) {
				value, _ = strconv.ParseFloat(value.(string), 64)
			}
			result = key.(float64) >= value.(float64)
		}
		break
	case "lesserThanEquals":

		if isNumber(key) && isNumber(value) {
			result = key.(float64) >= value.(float64)
		} else if isString(key) || isString(value) {

			if isString(key) {
				key, _ = strconv.ParseFloat(key.(string), 64)
			}
			if isString(value) {
				value, _ = strconv.ParseFloat(value.(string), 64)
			}
			result = key.(float64) >= value.(float64)
		}
		break
	default:
		result = false
	}
	return result
}
func isNumber(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}
func isString(val interface{}) bool {
	return reflect.TypeOf(val).String() == "string"
}
func (r *Rule) EvaluateRule(identity map[string]interface{}) bool {
	defer utils.GracefullyHandleError()
	var result = false
	key, ok := identity[r.GetAttributeName()]
	if !ok {
		return false
	}
	for _, val := range r.GetValues() {
		if r.operatorCheck(key, val) {
			result = true
		}
	}
	return result
}
