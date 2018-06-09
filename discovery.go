/*

discovery/discovery.go

written by superwhiskers, licensed under gnu agpl.
if you want a copy, go to http://www.gnu.org/licenses/

*/

package main

import (
	"strconv"
	"strings"

	"gitlab.com/superwhiskers/libninty"
	// internals

	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	// externals
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// a set of variables
var maintenanceURL string
var banURL string
var updateJSON string
var err error
var banData map[interface{}]interface{}
var defaultEndpoints map[interface{}]interface{}
var maintenanceData bool
var marshalledXML []byte
var overrideDiscovery bool
var groupdefs map[interface{}]interface{}
var endpoints map[interface{}]interface{}

// the handler for the discovery endpoint
func discoveryHandler(w http.ResponseWriter, r *http.Request) {

	// the response
	var fabricatedXML *result

	// trigger to tell if we will actually be able to ban it
	attemptToBan := true

	// get the servicetoken
	servicetoken, err := libninty.DecodeServiceToken(r.Header.Get("X-Nintendo-Servicetoken"))
	if err != nil {

		// display a message
		servicetoken = fmt.Sprintf("unable to decode servicetoken: %v", err)

		// set the attempt to ban flag
		attemptToBan = false

	} else {

		// hash the servicetoken
		servicetoken, err = hash(servicetoken)
		if err != nil {

			// display a message
			servicetoken = fmt.Sprintf("unable to hash servicetoken: %v", err)

			// set the attempt to ban flag
			attemptToBan = false

		}

	}

	// get the unpacked parampack
	parampack, err := libninty.DecodeParampack(r.Header.Get("X-Nintendo-Parampack"))
	if err != nil {
		fmt.Printf("-> unable to decode parampack. shown data is a nullified parampack\n")
	}

	// get x-forwarded-for
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	// print out request data
	fmt.Printf("-> ~ new request ~\n")
	fmt.Printf("-> service token (hashed): %s\n", servicetoken)
	fmt.Printf("-> remoteaddr: %s\n", r.RemoteAddr)
	fmt.Printf("-> x-forwarded-for: %s\n", xForwardedFor)
	fmt.Printf("-> parampack: \n%+v\n", parampack)

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

		// send the xml
		w.Header().Set("Content-Type", "application/xml")
		w.Write(marshalledXML)

		// don't continue
		return

	}

	// otherwise, we check if the person connecting is banned
	if attemptToBan == true {

		// loop over the bans
		for hash, banMap := range banData {

			// check if they're banned
			if compareHash(r.Header.Get("X-Nintendo-Servicetoken"), hash.(string)) == true {

				// they're banned, so we can respond with a ban message
				fabricatedXML = &result{
					HasError:  1,
					Version:   1,
					Code:      400,
					ErrorCode: 7,
					Message:   banMap.(map[interface{}]interface{})["reason"].(string),
				}

				// break from the loop
				break

			}

		}

	}

	// check if we've already created a response to send
	if fabricatedXML == nil {

		// standard mode

		// check if we override discovery
		if overrideDiscovery == true {

			// check if we use a different set of endpoints for this client
			if val, ok := groupdefs[servicetoken]; ok {

				// we do
				endpointset := endpoints[val.(string)].(map[interface{}]interface{})

				// and fabricate the response
				fabricatedXML = &result{
					HasError:   0,
					Version:    1,
					Host:       endpointset["discovery"].(string),
					APIHost:    endpointset["api"].(string),
					PortalHost: endpointset["wiiu"].(string),
					N3DSHost:   endpointset["3ds"].(string),
				}

			} else {

				// otherwise, fabricate the response as normal
				fabricatedXML = &result{
					HasError:   0,
					Version:    1,
					Host:       defaultEndpoints["discovery"].(string),
					APIHost:    defaultEndpoints["api"].(string),
					PortalHost: defaultEndpoints["wiiu"].(string),
					N3DSHost:   defaultEndpoints["3ds"].(string),
				}

			}

		} else {

			// check if we use a different set of endpoints for this client
			if val, ok := groupdefs[servicetoken]; ok {

				// we do
				endpointset := endpoints[val.(string)].(map[interface{}]interface{})

				// and fabricate the response
				fabricatedXML = &result{
					HasError:   0,
					Version:    1,
					Host:       r.Host,
					APIHost:    endpointset["api"].(string),
					PortalHost: endpointset["wiiu"].(string),
					N3DSHost:   endpointset["3ds"].(string),
				}

			} else {

				// otherwise, fabricate the response as normal
				fabricatedXML = &result{
					HasError:   0,
					Version:    1,
					Host:       r.Host,
					APIHost:    defaultEndpoints["api"].(string),
					PortalHost: defaultEndpoints["wiiu"].(string),
					N3DSHost:   defaultEndpoints["3ds"].(string),
				}

			}

		}

	}

	// marshal it
	marshalledXML, err = xml.MarshalIndent(fabricatedXML, "  ", "    ")
	if err != nil {

		fmt.Printf("[err]: could not marshal xml...\n")

	}

	// send the xml
	w.Header().Set("Content-Type", "application/xml")
	w.Write(marshalledXML)

}

// the main function, obviously
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
	endpoints = config["endpoints"].(map[interface{}]interface{})

	// default endpoints
	defaultEndpoints = endpoints["default"].(map[interface{}]interface{})

	// group definitions
	groupdefs = config["groupdefs"].(map[interface{}]interface{})

	// settings

	// do we override the automatic discovery endpoint calculation
	overrideDiscovery = settings["overrideDiscovery"].(bool)

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

	// maintenance is either a url to get a plaintext
	// response from (like this:
	//
	// inMaintenance: false
	//
	// ) or a boolean
	switch settings["maintenance"].(type) {

	case string:
		pullMaintenanceFromURL = true
		maintenanceURL = settings["maintenance"].(string)
		maintenanceData = false

	case bool:
		pullMaintenanceFromURL = false
		maintenanceData = settings["maintenance"].(bool)

	default:
		fmt.Printf("\n[err]: the maintenance field in the options must either be a boolean\n")
		fmt.Printf("       or a string of a url to a website where the server can fetch the status...\n")
		os.Exit(1)

	}

	// banList is either a url to get a plaintext
	// response from (like this:
	//
	// one-servicetoken: haha-yes
	// two-servicetoken: haha&yes
	// three-servicetoken: haha*yes
	//
	// ) or a list of banned servicetokens
	switch settings["bans"].(type) {

	case string:
		pullBansFromURL = true
		banURL = settings["bans"].(string)
		banData = map[interface{}]interface{}{}

	case map[interface{}]interface{}:
		pullBansFromURL = false
		banData = settings["bans"].(map[interface{}]interface{})

	default:
		fmt.Printf("[err]: the bans field in the options must either be a map of strings\n")
		fmt.Printf("       containing banned servicetokens or a url that points to an endpoint\n")
		fmt.Printf("       that will return a map of banned servicetokens...\n")
		os.Exit(1)

	}

	// check if we need to start a goroutine to update the status of the server
	if (pullMaintenanceFromURL == true || pullBansFromURL == true) && (updateCacheByTimeout == true) {

		// start it
		go func() {

			// do this forever
			for {

				// temporary variables for unpacking the data
				tmp := make(map[interface{}]interface{})

				// check if we update the bans via url
				if pullBansFromURL == true {

					// update the bans
					updateData, contenttype, err := get(banURL)
					if err != nil {

						// just show a message and go on
						fmt.Printf("[err]: your banlist update url might be invalid, please check this...\n")

					} else if contenttype != "text/plain" {

						// show a message and go on
						fmt.Printf("[err]: content type of banlist url is not text/plain...\n")

					} else {

						// create a variable with the request data split on newlines
						sepUpdatedData := strings.Split(updateData, "\n")

						// create another variable that will hold the split data of each index of the sepUpdatedData variable
						almostUpdatedData := [][]string{}

						// loop over and fill up the almostUpdatedData variable with that data
						for _, val := range sepUpdatedData {

							// split the index into two substrings
							almostUpdatedData = append(almostUpdatedData, strings.SplitN(val, ": ", 2))

						}

						// loop over almostUpdatedData filling up the tmp variable with maps
						for _, val := range almostUpdatedData {

							// add the map
							tmp[val[0]] = map[interface{}]interface{}{"reason": val[1]}

						}

						// move this data into the ban data variable
						banData = tmp

						// let the user know
						fmt.Printf("-> updated banlists...\n")

					}

				}

				// check if we update the maintenance via url
				if pullMaintenanceFromURL == true {

					// update the maintenance status
					updateData, contenttype, err := get(maintenanceURL)
					if err != nil {

						// same here
						fmt.Printf("[err]: your maintenance update url might be invalid\n")

					} else if contenttype != "text/plain" {

						// show a message and go on
						fmt.Printf("[err]: content type of banlist url is not text/plain...\n")

					} else {

						// create a variable with the request data split on newlines
						sepUpdatedData := strings.Split(updateData, "\n")

						// create another variable that will hold the split data of each index of the sepUpdatedData variable
						almostUpdatedData := [][]string{}

						// loop over and fill up the almostUpdatedData variable with that data
						for _, val := range sepUpdatedData {

							// split the index into two substrings
							almostUpdatedData = append(almostUpdatedData, strings.SplitN(val, ": ", 2))

						}

						// loop over almostUpdatedData filling up the tmp variable with maps
						for _, val := range almostUpdatedData {

							// parse boolean
							tmp2, err := strconv.ParseBool(val[1])

							// handle errors
							if err != nil {

								// show a message
								fmt.Printf("[err]: incorrectly formatted boolean: %s\n", val[1])

								// reset the temporary variable
								tmp2 = false

							}

							// add the map
							tmp[val[0]] = tmp2

						}

						// move this data into the variable
						maintenanceData = tmp["inMaintenance"].(bool)

						// let the user know that we did it
						fmt.Printf("-> updated maintenance status\n")

					}

				}

				// timeout
				time.Sleep(time.Duration(timeoutForAutomatic) * time.Second)

			}

		}()

	}

	// create a new router
	r := mux.NewRouter()

	// register the handler for the discovery endpoint
	r.HandleFunc(endpointForDiscovery, discoveryHandler)

	// server configuration
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", serverPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// start the server
	fmt.Printf("-> starting server...\n")

	// do we use https?
	if settings["https"].(bool) == true {

		// host on https
		log.Fatal(srv.ListenAndServeTLS("tls/cert.pem", "tls/key.pem"))

	} else {

		// host on http
		log.Fatal(srv.ListenAndServe())

	}

}
