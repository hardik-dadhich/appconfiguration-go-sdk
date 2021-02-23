// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	AppConfiguration "github.com/IBM/appconfiguration-go-sdk/lib"
)

var appConfiguration *AppConfiguration.AppConfiguration

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Sample App HomePage!")
}

func setUp() {

	appConfiguration = AppConfiguration.GetInstance()
	appConfiguration.EnableDebug(true)
	appConfiguration.Init("region", "guid", "api_key")
	appConfiguration.SetCollectionId("collectionId")
	appConfiguration.RegisterFeaturesUpdateListener(func() {
		fmt.Println("Get all your features now.")
		features := appConfiguration.GetFeatures()
		feature := features["featureId"]

		fmt.Println("Feature Name ", feature.GetFeatureName())
		fmt.Println("Feature Id  ", feature.GetFeatureId())
		fmt.Println("Feature Type ", feature.GetFeatureDataType())
		fmt.Println("Feature is enabled ", feature.IsEnabled())

		fmt.Println("Get your feature specific value now.")
		feature = appConfiguration.GetFeature("featureId")
		if feature.IsEnabled() {
			fmt.Println("Enable the feature.")
		} else {
			fmt.Println("Disable the feature.")
		}
		fmt.Println(feature)
		fmt.Println("Feature Name ", feature.GetFeatureName())
		fmt.Println("Feature Id ", feature.GetFeatureId())
		fmt.Println("Feature Type ", feature.GetFeatureDataType())
		fmt.Println("Feature is enabled ", feature.IsEnabled())
	})

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	setUp()
}
