# IBM Cloud App Configuration Go server SDK

IBM Cloud App Configuration SDK is used to perform feature flag and property evaluation based on the configuration on IBM Cloud App Configuration service.

## Table of Contents

  - [Overview](#overview)
  - [Installation](#installation)
  - [Import the SDK](#import-the-sdk)
  - [Initialize SDK](#initialize-sdk)
  - [License](#license)

## Overview

IBM Cloud App Configuration is a centralized feature management and configuration service on [IBM Cloud](https://www.cloud.ibm.com) for use with web and mobile applications, microservices, and distributed environments.

Instrument your applications with App Configuration Go SDK, and use the App Configuration dashboard, API or CLI to define feature flags or properties, organized into collections and targeted to segments. Change feature flag states in the cloud to activate or deactivate features in your application or environment, when required. You can also manage the properties for distributed applications centrally.

## Installation

Install using the command.

```bash
go get -u github.com/IBM/appconfiguration-go-sdk
```

## Import the SDK

To import the module 

```go
import AppConfiguration "github.com/IBM/appconfiguration-go-sdk/lib"
```

## Initialize SDK
Initialize the sdk to connect with your App Configuration service instance.
```go
appConfiguration = AppConfiguration.GetInstance()
appConfiguration.Init("region", "guid", "apikey")

// Set the collection Id
appConfiguration.SetCollectionId("collectionId")
```

- region : Region name where the App Configuration service instance is created. Use
    - `AppConfiguration.REGION_US_SOUTH` for Dallas
    - `AppConfiguration.REGION_EU_GB` for London
    - `AppConfiguration.REGION_AU_SYD` for Sydney
- guid : Instance Id of the App Configuration service. Get it from the service credentials section of the dashboard.
- apikey : ApiKey of the App Configuration service. Get it from the service credentials section of the dashboard.

* collectionId: Id of the collection created in App Configuration service instance.
> Here, by default live update from the server is enabled. To turn off this mode see the [below section](#work-offline-with-local-configuration-file)



### Work offline with local configuration file
You can also work offline with local configuration file and perform feature and property related operations.

After setting the collection Id, follow the below steps
```go
appConfiguration.FetchConfigurationFromFile( "configurationFilePath", "liveConfigUpdateEnabled")
```
- configurationFilePath: Path to the JSON file, which contains configuration details.
- liveConfigUpdateEnabled: Set this value to `false` if the new configuration values shouldn't be fetched from the server. Make sure to provide a proper JSON file in the path. By default, this value is enabled.


## Get single feature

```go
feature := appConfiguration.GetFeature("featureId")

if(feature.IsEnabled()) {
        // enable feature
} else {
        // disable the feature
}
fmt.Println(feature);
fmt.Println("Feature Name %s", feature.GetFeatureName());
fmt.Println("Feature Id  %s", feature.GetFeatureId());
fmt.Println("Feature Type %s", feature.GetFeatureDataType());
fmt.Println("Feature is enabled %t ", feature.IsEnabled());
```
## Get all features

```go
features := appConfiguration.GetFeatures()

feature := features["featureId"];

fmt.Println("Feature Name %s", feature.GetFeatureName());
fmt.Println("Feature Id  %s", feature.GetFeatureId());
fmt.Println("Feature Type %s", feature.GetFeatureDataType());
fmt.Println("Feature is enabled %t ", feature.IsEnabled());
```
## Evaluate a feature

You can use the ` feature.GetCurrentValue(identityId, identityAttributes)` method to evaluate the value of the feature flag.
You should pass an unique identityId as the parameter to perform the feature flag evaluation. If the feature flag is configured with segments in the App Configuration service, you can set the attributes values as a map.

```go
identityId := "identityId"
identityAttributes := make(map[string]interface{})
identityAttributes["email"] = "ibm.com"
identityAttributes["city"] = "Bengaluru"

featureVal := feature.GetCurrentValue(identityId, identityAttributes)
```

## Get single property

```go
property := appConfiguration.GetProperty("propertyId")

fmt.Println(property);
fmt.Println("Property Name %s", property.GetPropertyName());
fmt.Println("Property Id  %s", property.GetPropertyId());
fmt.Println("Property Type %s", property.GetPropertyDataType());
```
## Get all properties

```go
properties := appConfiguration.GetProperties()

property := properties["propertyId"];

fmt.Println("Property Name %s", property.GetPropertyName());
fmt.Println("Property Id  %s", property.GetPropertyId());
fmt.Println("Property Type %s", property.GetPropertyDataType());
```
## Evaluate a property

You can use the ` property.GetCurrentValue(identityId, identityAttributes)` method to evaluate the value of the property.
You should pass an unique identityId as the parameter to perform the property evaluation. If the property is configured with segments in the App Configuration service, you can set the attributes values as a map.

```go
identityId := "identityId"
identityAttributes := make(map[string]interface{})
identityAttributes["email"] = "ibm.com"
identityAttributes["city"] = "Bengaluru"

propertyVal := property.GetCurrentValue(identityId, identityAttributes)
```



## Set listener for feature or property data changes

To listen to the data changes add the following code in your application

```go
appConfiguration.RegisterConfigurationUpdateListener(func() {
    fmt.Println("Get your feature/property value now ")
})
```


## Fetch latest data

```go
appConfiguration.FetchConfigurations()
```

## Enable debugger (Optional)

```go
appConfiguration.EnableDebug(true)
```

## License

This project is released under the Apache 2.0 license. The license's full text can be found in [LICENSE](/LICENSE)
