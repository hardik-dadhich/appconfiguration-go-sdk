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
	entityID := "user123"
	entityAttributes := make(map[string]interface{})
	entityAttributes["city"] = "Bangalore"
	entityAttributes["radius"] = 60

	fmt.Println("\n\nFEATURE FLAG OPERATIONS\n")
	feature, err := appConfiguration.GetFeature("<featureId>")
	if err == nil {
		fmt.Println("Feature Name:", feature.GetFeatureName())
		fmt.Println("Feature Id:", feature.GetFeatureID())
		fmt.Println("Feature Data type:", feature.GetFeatureDataType())
		fmt.Println("Is Feature enabled?", feature.IsEnabled())
		fmt.Println("Feature evaluated value is:", feature.GetCurrentValue(entityID, entityAttributes))
	}

	fmt.Println("\n\nPROPERTY OPERATIONS\n")
	property, err := appConfiguration.GetProperty("<propertyId>")
	if err == nil {
		fmt.Println("Property Name:", property.GetPropertyName())
		fmt.Println("Property Id:", property.GetPropertyID())
		fmt.Println("Property Data type:", property.GetPropertyDataType())
		fmt.Println("Property evaluated value is:", property.GetCurrentValue(entityID, entityAttributes))
	}
	//whenever the configurations get changed/updated on the app configuration service instance the function inside this listener is triggered.
	//So, to keep track of live changes to configurations use this listener.
	appConfiguration.RegisterConfigurationUpdateListener(func() {
		fmt.Println("configurations updated")
	})

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
