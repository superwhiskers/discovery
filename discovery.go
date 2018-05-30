/*

discovery/discovery.go

written by superwhiskers, licensed under gnu agpl.
if you want a copy, go to http://www.gnu.org/licenses/

*/

package main

import (
	// internals
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"time"
	// externals
	"github.com/labstack/echo"
	"gopkg.in/yaml.v2"
)

var maintenanceURL string
var banURL string
var updateJSON string
var err error
var banData []interface{}
var maintenanceData bool
var fabricatedXML *result
var marshalledXML []byte

func main() {

	config := make(map[interface{}]interface{})

	// get the file data
	confByte, err := readFileByte("config.yaml")

	// check for errors
	if err != nil {

		// show a message
		fmt.Printf("\n[err]: error while loading config.yaml.\n")
		fmt.Printf("       you should copy config.example.yaml to config.yaml and edit it.\n")

		// exit
		os.Exit(1)

	}

	// parse it to yaml
	err = yaml.Unmarshal(confByte, config)

	// check for errors
	if err != nil {

		// show a message
		fmt.Printf("\n[err]: there is an error in your yaml in config.yaml...\n")

		// and show a traceback
		panic(err)

	}

	// predefine some variables
	var pullMaintenanceFromURL bool
	var pullBansFromURL bool

	// get some config sections from the config
	settings := config["options"].(map[interface{}]interface{})
	endpoints := config["endpoints"].(map[interface{}]interface{})

	// endpoints
	host := endpoints["discovery"].(string)
	apiHost := endpoints["api"].(string)
	portalHost := endpoints["wiiu"].(string)
	nintendo3dsHost := endpoints["3ds"].(string)

	// settings

	// the endpoint to place the discovery data on
	endpointForDiscovery := settings["endpoint"].(string)

	// port for the server
	serverPort := settings["port"].(int)

	// cache settings
	cacheSettings := settings["cache"].(map[interface{}]interface{})

	// do we allow using timeouts to update cache
	updateCacheByTimeout := cacheSettings["useTimeout"].(bool)

	// timeouts for the automatic cache update and endpoint, respectively
	timeoutForAutomatic := cacheSettings["autoTimeout"].(int)

	// maintenance is either a url to get a json
	// response from (like this:
	// { inMaintenance: false }
	// ) or a boolean
	switch settings["maintenance"].(type) {

	case string:
		pullMaintenanceFromURL = true
		maintenanceURL = settings["maintenance"].(string)

	case bool:
		pullMaintenanceFromURL = false
		maintenanceData = settings["maintenance"].(bool)

	default:
		fmt.Printf("\n[err]: the maintenance field in the options must either be a boolean\n")
		fmt.Printf("       or a string of a url to a website where the server can fetch the status...\n")
		os.Exit(1)

	}

	// banList is either a url to get a json
	// response from (like this:
	// { bans: [
	// 	{ "token": "one-servicetoken", "reason": "haha-yes" },
	// 	{ "token": "two-servicetoken", "reason": "haha&yes" },
	// 	{ "token": "three-servicetoken", "reason": "haha*yes" }
	// ] }
	// ) or a list of banned servicetokens
	switch settings["bans"].(type) {

	case string:
		pullBansFromURL = true
		banURL = settings["bans"].(string)

	case []interface{}:
		pullBansFromURL = false
		banData = settings["bans"].([]interface{})

	default:
		fmt.Printf("[err]: the bans field in the options must either be a list of strings\n")
		fmt.Printf("       containing banned servicetokens or a url that points to an endpoint\n")
		fmt.Printf("       that will return a list of banned servicetokens...")
		os.Exit(1)

	}

	// check if we need to start a goroutine to update the status of the server
	if (pullMaintenanceFromURL == true || pullBansFromURL == true) && (updateCacheByTimeout == true) {

		// start it
		go func() {

			// temporary variables for unpacking the data
			var tmp interface{}

			// check if we update the bans via url
			if pullBansFromURL == true {

				// update the bans
				updateJSON, err = get(banURL)
				if err != nil {

					// just show a message and go on
					fmt.Printf("[err]: your banlist update url might be invalid, please check this...")

				}

				// unpack the data
				if err := json.Unmarshal([]byte(updateJSON), &tmp); err != nil {

					// print an error message and go on
					fmt.Printf("[err]: error while unpacking the json data from the banlist into a go-supported type...")

				}

				// move this data into the ban data variable
				banData = tmp.(map[interface{}]interface{})["bans"].([]interface{})

			}

			// check if we update the maintenance via url
			if pullMaintenanceFromURL == true {

				// update the maintenance status
				updateJSON, err = get(maintenanceURL)
				if err != nil {

					// same here
					fmt.Printf("[err]: your maintenance update url might be invalid")

				}

				// unpack the data
				if err := json.Unmarshal([]byte(updateJSON), &tmp); err != nil {

					// print an error message and go on
					fmt.Printf("[err]: error while unpacking the json data from the maintenance endpoint into a go-supported type...")

				}

				// move this data into the variable
				maintenanceData = tmp.(map[interface{}]interface{})["inMaintenance"].(bool)

			}

			// timeout
			time.Sleep(time.Duration(timeoutForAutomatic) * time.Second)

		}()

	}

	// load echo
	e := echo.New()

	// set a get handler for the xml that nintendo consoles that support miiverse use
	e.GET(endpointForDiscovery, func(c echo.Context) error {

		// we have a request
		fmt.Printf("-> request to discovery...\n")

		// first, check if we are in maintenance mode
		if maintenanceData == true {

			// then we are

			// fabricate the response
			fabricatedXML = &result{
				HasError:  1,
				Version:   1,
				Code:      400,
				ErrorCode: 3,
				Message:   "SERVICE_MAINTENANCE",
			}

			// marshal it
			marshalledXML, err = xml.MarshalIndent(fabricatedXML, "  ", "    ")
			if err != nil {

				fmt.Printf("[err]: could not marshal xml...\n")

			}

			// send the blob
			return c.XMLBlob(http.StatusOK, marshalledXML)

		} else {

			/*
				// otherwise, we check if the person connecting is banned
				for _, b := range banData {
					if b == a {
						return true
					}
				}
			*/

			// standard mode

			// fabricate the response
			fabricatedXML = &result{
				HasError:   0,
				Version:    1,
				Host:       host,
				APIHost:    apiHost,
				PortalHost: portalHost,
				N3DSHost:   nintendo3dsHost,
			}

			// marshal it
			marshalledXML, err = xml.MarshalIndent(fabricatedXML, "  ", "    ")
			if err != nil {

				fmt.Printf("[err]: could not marshal xml...\n")

			}

			// send the blob
			return c.XMLBlob(http.StatusOK, marshalledXML)

		}

	})

	// hide the startup banner
	e.HideBanner = true

	// run the server
	e.Logger.Fatal(e.StartTLS(fmt.Sprintf(":%d", serverPort), "tls/cert.pem", "tls/key.pem"))

}
