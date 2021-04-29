// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	AppConfiguration "github.com/IBM/appconfiguration-go-sdk/lib"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to Sample App HomePage!")
}

func main() {
	appConfiguration := AppConfiguration.GetInstance()
	appConfiguration.Init(AppConfiguration.REGION_US_SOUTH, "<guid>", "<apikey>")
	appConfiguration.SetContext("<collectionId>", "<environmentId>")
	identityId := "user123"
	identityAttributes := make(map[string]interface{})
	identityAttributes["city"] = "Bangalore"
	identityAttributes["radius"] = 60

	appConfiguration.RegisterConfigurationUpdateListener(func() {

		fmt.Println("\n\nFEATURE FLAG OPERATIONS\n")
		feature := appConfiguration.GetFeature("<featureId>")
		fmt.Println("Feature Name:", feature.GetFeatureName())
		fmt.Println("Feature Id:", feature.GetFeatureId())
		fmt.Println("Feature Data type:", feature.GetFeatureDataType())
		fmt.Println("Is Feature enabled?", feature.IsEnabled())
		fmt.Println("Feature evaluated value is:", feature.GetCurrentValue(identityId, identityAttributes))

		fmt.Println("\n\nPROPERTY OPERATIONS\n")
		property := appConfiguration.GetProperty("<propertyId>")
		fmt.Println("Property Name:", property.GetPropertyName())
		fmt.Println("Property Id:", property.GetPropertyId())
		fmt.Println("Property Data type:", property.GetPropertyDataType())
		fmt.Println("Property evaluated value is:", property.GetCurrentValue(identityId, identityAttributes))

	})

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
