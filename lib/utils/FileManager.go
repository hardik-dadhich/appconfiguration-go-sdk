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
	constants "github.com/IBM/appconfiguration-go-sdk/lib/constants"
	messages "github.com/IBM/appconfiguration-go-sdk/lib/messages"

	"encoding/json"
	"io/ioutil"
)

const featureFile = constants.FEATURE_FILE

func StoreFiles(content string) {
	log.Debug(messages.STORE_FILE)

	file, err := json.MarshalIndent(json.RawMessage(content), "", "\t")
	if err != nil {
		log.Error(messages.ENCODE_JSON_ERR, err)
		return
	}
	err = ioutil.WriteFile(featureFile, (file), 0644)
	if err != nil {
		log.Error(messages.WRITE_FILE_ERR, err)
		return
	}

}

func ReadFiles(filePath string) []byte {
	log.Debug(messages.READ_FILE)
	fileToRead := featureFile
	if len(filePath) > 0 {
		fileToRead = filePath
	}
	file, err := ioutil.ReadFile(fileToRead)
	if err != nil {
		log.Error(messages.READ_FILE_ERR, err)
		return nil
	}
	return file
}
