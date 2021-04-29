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
const ENVIRONMENT_ID_VALUE_ERROR = "Provide a valid environmentId."
const COLLECTION_ID_ERROR = "Invalid action. You can perform this action only after a successful initialization. Check the initialization section for errors."
const COLLECTION_INIT_ERROR = "Invalid action. You can perform this action only after a successful initialization and setting the context. Check the Init and SetContext section for errors."
const CONFIGURATION_FILE_NOT_FOUND_ERROR = "Provide configuration_file value when live_config_update_enabled is false."
const INCORRECT_USAGE_OF_CONTEXT_OPTIONS = "Incorrect usage of context options. At most of one ContextOptions struct should be passed."
const CONFIG_API_ERROR = "Invalid configuration. Verify the collectionId, environmentId, apikey, guid and region."

// configurationHandler
const FETCH_FROM_API_SDK_INIT_ERROR = "fetchFromAPI() - Configuration SDK not initialized with call to init."
const WEBSOCKET_ERROR_READING_MESSAGE = "Error while reading message from the socket."
const WEBSOCKET_RECEIVING_MESSAGE = "Message received from socket."
const CONFIGURATION_UPDATE_LISTENER_METHOD_ERROR = "Configuration update listener should me a method or a function."
const SET_IDENTITY_OBJECT_ID_ERROR = "Provide Id as a first param to GetCurrentValue."
const CONFIGURATION_HANDLER_INIT_ERROR = "Invalid action in ConfigurationHandler. You can perform this action only after a successful initialization. Check the initialization section for errors."
const CREATING_NEW_API_MANAGER_INSTANCE = "Creating new API manager instance."

//
const PARSING_FEATURE_RULES = "Parsing feature rules."
const EVALUATING_SEGMENTS = "Evaluating segments."
const FEATURE_VALUE = "Feature value."
const EVALUATING_FEATURE = "Evaluating feature."
const RETRIEVING_FEATURE = "Retrieving feature current value."
const INVALID_FEATURE_ID = "Invalid feature id - "

//
const PARSING_PROPERTY_RULES = "Parsing property rules."
const PROPERTY_VALUE = "Property value."
const EVALUATING_PROPERTY = "Evaluating property."
const RETRIEVING_PROPERTY = "Retrieving property current value."
const INVALID_PROPERTY_ID = "Invalid property id - "
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
const SETTING_CONTEXT = "Setting context."

const LOADING_DATA = "Loading data."
const CHECK_CONFIGURATION_FILE_PROVIDED = "Checking configuration file is provided by the user or not."
const CONFIGURATION_FILE_PROVIDED = "User provided configuration file."
const LOADING_CONFIGURATIONS = "Loading configurations."
const LIVE_UPDATE_CHECK = "Checking live configuration update is enabled or not."
const FETCH_FROM_CONFIGURATION_FILE = "Fetching from configuration file."
const FETCH_CONFIGURATION_DATA = "Fetching configuration data."
const FETCH_FROM_API = "Fetching from API."
const START_WEB_SOCKET = "Starting web socket connection."
const WEB_SOCKET_CONNECT_ERR = "Error connecting to websocket "
const RETRY_WEB_SCOKET_CONNECT = "Trying web socket connection again."
const UNMARSHAL_JSON_ERR = "Error while unmarshalling JSON "
const MARSHAL_JSON_ERR = "Error while marshalling JSON "

const SET_IN_MEMORY_CACHE = "Setting memory cache."
