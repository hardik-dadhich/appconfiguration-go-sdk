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

package messages

const REGION_ERROR = "Provide a valid region."
const GUID_ERRROR = "Provide a valid guid."
const APIKEY_ERROR = "Provide a valid apiKey."
const COLLECTION_ID_VALUE_ERROR = "Provide a valid collectionId."
const COLLECTION_ID_ERROR = "Invalid action. You can perform this action only after a successful initialization. Check the initialization section for errors."
const COLLECTION_SUB_ERROR = "Invalid action. You can perform this action only after a successful initialization and set a collectionId value. Check the initialization and SetCollectionId section for errors."
const FEATURE_FILE_NOT_FOUND_ERROR = "Provide feature_file value when live_feature_update_enabled is false."

// feaureHandler
const FETCH_FROM_API_SDK_INIT_ERROR = "fetchFromAPI() - Feature SDK not initialized with call to init."
const WEBSOCKET_ERROR_READING_MESSAGE = "Error while reading message from the socket."
const WEBSOCKET_RECEIVING_MESSAGE = "Message received from socket."
const FEATURES_UPDATE_LISTENER_METHOD_ERROR = "Features update listener should me a method or a function."
const SET_IDENTITY_OBJECT_ID_ERROR = "Provide Id as a first param to GetCurrentValue."
const FEATURE_HANDLER_INIT_ERROR = "Invalid action in FeatureHandler. You can perform this action only after a successful initialization. Check the initialization section for errors."
const CREATING_NEW_API_MANAGER_INSTANCE = "Creating new API manager instance."

//
const PARSING_FEATURE_RULES = "Parsing feature rules."
const EVALUATING_SEGMENTS = "Evaluating segments."
const FEATURE_VALUE = "Feature value."
const EVALUATING_FEATURE = "Evaluating feature."
const RETRIEVING_FEATURE = "Retrieving feature current value."

//
const EVAL_SEGMENT_RULE = "Evaluating segment rule."

//
const EXEC_API_CALL = "Executing API call."
const API_CALL_ERROR = "Error executing API call "

//
const ENCODE_JSON_ERR = "Error while encoding json "
const WRITE_FILE_ERR = "Error while writing to a file "
const READ_FILE_ERR = "Error while reading file "

//
const STORE_FILE = "Storing file."
const READ_FILE = "Reading file."

//
const RETRIEVE_METERING_INSTANCE = "Retrieving metering instance."
const START_SENDING_METERING_DATA = "Start sending metering data in the background."
const ADD_METERING = "Add metering."
const RECORD_EVAL = "Record evaluation."
const TEN_MIN_EXPIRY = "10 mins expired, sending metering data if any."
const SEND_METERING_SERVER = "Sending to metering server."
const SEND_METERING_SERVER_ERR = "Error while sending metering data to server "

//
const RETRIEVEING_APP_CONFIG = "Retrieving App Configuration instance."
const CREATING_NEW_APP_CONFIG = "Creating new App Configuration instance."
const APP_CONFIG_ALREADY_INSTANTIATED = "App Configuration instance is already instantiated."
const SETTING_COLLECTION_ID = "Setting collectionId."

const LOADING_DATA = "Loading data."
const CHECK_FEATURE_FILE_PROVIDED = "Checking feature file is provided by the user or not."
const FEATURE_FILE_PROVIDED = "User provided feature file."
const LOADING_FEATURES = "Loading features."
const LIVE_UPDATE_CHECK = "Checking live feature update is enabled or not."
const FETCH_FROM_FEATURE_FILE = "Fetching from feature file."
const FETCH_FEATURES_DATA = "Fetching features data."
const FETCH_FROM_API = "Fetching from API."
const START_WEB_SOCKET = "Starting web socket connection."
const WEB_SOCKET_CONNECT_ERR = "Error connecting to websocket "
const RETRY_WEB_SCOKET_CONNECT = "Trying web socket connection again."
const UNMARSHAL_JSON_ERR = "Error while unmarshalling JSON "
const MARSHAL_JSON_ERR = "Error while marshalling JSON "

const SET_IN_MEMORY_CACHE = "Setting memory cache."
