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
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
	// externals
	"github.com/gorilla/mux"
	"github.com/superwhiskers/yaml"
	"github.com/tomasen/realip"
	"gitlab.com/superwhiskers/libninty"
	//"gopkg.in/yaml.v3" when yaml.v3 is available, i will use that instead
)

// a set of variables
var (
	maintenanceURL    string
	groupdefsURL      string
	banURL            string
	updateJSON        string
	err               error
	banData           map[string]interface{}
	defaultEndpoints  map[string]interface{}
	maintenanceData   bool
	marshalledXML     []byte
	overrideDiscovery bool
	bcryptCost        int
	endpoints         map[string]interface{}
	groupdefsData     map[string]interface{}
)

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
		servicetoken, err = hash(servicetoken, bcryptCost)
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
		log.Printf("-> unable to decode parampack. shown data is a nullified parampack\n")
	}

	// print out request data
	log.Printf("-> ~ new request ~\n")
	log.Printf("-> service token (hashed): %s\n", servicetoken)
	log.Printf("-> ip address: %s\n", realip.FromRequest(r))
	log.Printf("-> parampack: %+v\n", parampack)

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

			// output an error message if an error occured
			log.Printf("[err]: could not marshal xml...\n")
			log.Printf("       error: %v\n", err)

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
			banned, err := compareHash(r.Header.Get("X-Nintendo-Servicetoken"), hash)

			// check for errors decoding
			if err != nil {

				// show the error
				log.Printf("[err]: %s is not a hexadecimal-encoded hash...\n", hash)

			}

			// check if they're banned
			if banned == true {

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

		// check if we override discovery
		if overrideDiscovery == true {

			// the match variable that we use to check if we ever found a match
			match := false

			// loop over the groupdefs
			for hash, group := range groupdefsData {

				// check if the servicetokens match
				match, err = compareHash(r.Header.Get("X-Nintendo-Servicetoken"), hash)

				// check for errors decoding
				if err != nil {

					// show the error
					log.Printf("[err]: %s is not a hexadecimal-encoded hash...\n", hash)

				}

				// check if they're banned
				if match == true {

					// we do
					endpointset := endpoints[group.(string)].(map[string]interface{})

					// and fabricate the response
					fabricatedXML = &result{
						HasError:   0,
						Version:    1,
						Host:       endpointset["discovery"].(string),
						APIHost:    endpointset["api"].(string),
						PortalHost: endpointset["wiiu"].(string),
						N3DSHost:   endpointset["3ds"].(string),
					}

					// break from the loop
					break

				}

			}

			// if match was never set to true, we need to give them the standard endpoints
			if match == false {

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

			// the match variable that we use to check if we ever found a match
			match := false

			// loop over the groupdefs
			for hash, group := range groupdefsData {

				// check if the servicetokens match
				match, err = compareHash(r.Header.Get("X-Nintendo-Servicetoken"), hash)

				// check for errors decoding
				if err != nil {

					// show the error
					log.Printf("[err]: %s is not a hexadecimal-encoded hash...\n", hash)

				}

				// check if they're banned
				if match == true {

					// we do
					endpointset := endpoints[group.(string)].(map[string]interface{})

					// and fabricate the response
					fabricatedXML = &result{
						HasError:   0,
						Version:    1,
						Host:       r.Host,
						APIHost:    endpointset["api"].(string),
						PortalHost: endpointset["wiiu"].(string),
						N3DSHost:   endpointset["3ds"].(string),
					}

					// break from the loop
					break

				}

			}

			// if match was never set to true, we need to give them the standard endpoints
			if match == false {

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

		// output an error message if an error occured
		log.Printf("[err]: could not marshal xml...\n")
		log.Printf("       error: %v\n", err)

	}

	// send the xml
	w.Header().Set("Content-Type", "application/xml")
	w.Write(marshalledXML)

}

// the main function, obviously
func main() {

	// set the default map type
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})

	// variable that the config is parsed into
	config := make(map[string]interface{})

	// get the file data
	confByte, err := readFileByte("config.yaml")

	// check for errors
	if err != nil {

		// show a message
		fmt.Printf("[err]: error while loading config.yaml.\n")
		fmt.Printf("       you should copy config.example.yaml to config.yaml and edit it.\n")

		// exit
		os.Exit(1)

	}

	// parse it to yaml
	err = yaml.Unmarshal(confByte, config)

	// check for errors
	if err != nil {

		// show a message
		fmt.Printf("[err]: there is an error in your yaml in config.yaml...\n")

		// and show a traceback
		panic(err)

	}

	// predefine some variables
	var (
		pullMaintenanceFromURL bool
		pullBansFromURL        bool
		pullGroupdefsFromURL   bool
	)

	// set some others
	var (
		settings      = config["options"].(map[string]interface{})
		logfile       = settings["logfile"].(string)
		serverPort    = settings["port"].(int)
		cacheSettings = settings["cache"].(map[string]interface{})
	)

	// these have to be left out because we're modifying existing ones
	endpoints = config["endpoints"].(map[string]interface{})
	bcryptCost = settings["hashCost"].(int)
	overrideDiscovery = settings["overrideDiscovery"].(bool)
	endpointForDiscovery := settings["endpoint"].(string)
	defaultEndpoints = endpoints["default"].(map[string]interface{})

	// open the logfile
	file, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// check for errors
	if err != nil {

		// show an error message
		fmt.Printf("[err]: unable to open file %s...\n", logfile)

		// panic
		panic(err)

	}

	// close it when this function returns
	defer file.Close()

	// set the output for the logger
	log.SetOutput(io.MultiWriter(os.Stdout, file))

	// groupdefs is either a url to get a plaintext
	// response from (like this:
	//
	// { "servicetoken-one": "group-name", "servicetoken-two": "group-name" }
	//
	// )
	switch config["groupdefs"].(type) {

	case string:
		pullGroupdefsFromURL = true
		groupdefsURL = config["groupdefs"].(string)
		groupdefsData = map[string]interface{}{}

	case map[string]interface{}:
		pullGroupdefsFromURL = false
		groupdefsData = config["groupdefs"].(map[string]interface{})

	case nil:
		pullGroupdefsFromURL = false
		groupdefsData = map[string]interface{}{}

	default:
		log.Printf("[err]: the groupdefs field in the config must be either a map[string]interface{} or\n")
		log.Printf("       a string of a url to a website where the server can fetch the status...\n")
		log.Printf("current type: %v\n", reflect.TypeOf(config["groupdefs"]))
		os.Exit(1)

	}

	// maintenance is either a url to get a plaintext
	// response from (like this:
	//
	// { "inMaintenance": false }
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
		log.Printf("[err]: the maintenance field in the options must either be a boolean\n")
		log.Printf("       or a string of a url to a website where the server can fetch the status...\n")
		log.Printf("current type: %v\n", reflect.TypeOf(settings["maintenance"]))
		os.Exit(1)

	}

	// banList is either a url to get a plaintext
	// response from (like this:
	//
	// { "one-servicetoken": { "reason": "haha-yes" }, "two-servicetoken": { "reason": "haha-yes" } }
	//
	// ) or a list of banned servicetokens
	switch settings["bans"].(type) {

	case string:
		pullBansFromURL = true
		banURL = settings["bans"].(string)
		banData = map[string]interface{}{}

	case map[string]interface{}:
		pullBansFromURL = false
		banData = settings["bans"].(map[string]interface{})

	case nil:
		pullBansFromURL = false
		banData = map[string]interface{}{}

	default:
		log.Printf("[err]: the bans field in the options must either be a map of strings\n")
		log.Printf("       containing banned servicetokens or a url that points to an endpoint\n")
		log.Printf("       that will return a map of banned servicetokens...\n")
		log.Printf("current type: %v\n", reflect.TypeOf(settings["bans"]))
		os.Exit(1)

	}

	// check if we use a goroutine to update the maintenance status
	if pullMaintenanceFromURL == true {

		// start it
		go func() {

			// do this forever
			for {

				// temporary variables for unpacking the data
				var tmp interface{}

				// update the maintenance status
				updateData, err := get(maintenanceURL)
				if err != nil {

					// same here
					log.Printf("[err]: your maintenance update url might be invalid\n")

				} else {

					// unmarshal json data gotten from the url
					err = json.Unmarshal([]byte(updateData), &tmp)

					// handle errors
					if err != nil {

						// show an error message if needed
						log.Printf("[err]: the data at your maintenance update url is invalid json...\n")

					} else {

						// move this data into the variable
						maintenanceData = tmp.(map[string]interface{})["inMaintenance"].(bool)

						// let the user know that we did it
						log.Printf("-> updated maintenance status\n")

					}

				}

				// timeout
				time.Sleep(time.Duration(cacheSettings["maintenanceTimeout"].(int)) * time.Second)

			}

		}()

	}

	// check if we use a goroutine to update banlists
	if pullBansFromURL == true {

		// start it
		go func() {

			// do this forever
			for {

				// temporary variables for unpacking the data
				var tmp interface{}

				// update the bans
				updateData, err := get(banURL)
				if err != nil {

					// just show a message and go on
					log.Printf("[err]: your banlist update url might be invalid, please check this...\n")

				} else {

					// unmarshal json data gotten from the url
					err = json.Unmarshal([]byte(updateData), &tmp)

					// handle errors
					if err != nil {

						// show an error message if needed
						log.Printf("[err]: the data at your maintenance update url is invalid json...\n")

					} else {

						// move this data into the ban data variable
						banData = tmp.(map[string]interface{})

						// let the user know
						log.Printf("-> updated banlists...\n")

					}

				}

				// timeout
				time.Sleep(time.Duration(cacheSettings["banlistTimeout"].(int)) * time.Second)

			}

		}()

	}

	// check if we use a goroutine to update groupdefs
	if pullGroupdefsFromURL == true {

		// start it
		go func() {

			// do this forever
			for {

				// temporary variables for unpacking the data
				var tmp interface{}

				// update the bans
				updateData, err := get(groupdefsURL)
				if err != nil {

					// just show a message and go on
					log.Printf("[err]: your groupdefs update url might be invalid, please check this...\n")

				} else {

					// unmarshal json data gotten from the url
					err = json.Unmarshal([]byte(updateData), &tmp)

					// handle errors
					if err != nil {

						// show an error message if needed
						log.Printf("[err]: the data at your groupdefs update url is invalid json...\n")

					} else {

						// move this data into the ban data variable
						groupdefsData = tmp.(map[string]interface{})

						// let the user know
						log.Printf("-> updated groupdefs...\n")

					}

				}

				// timeout
				time.Sleep(time.Duration(cacheSettings["groupdefsTimeout"].(int)) * time.Second)

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
	log.Printf("-> starting server...\n")

	// do we use https?
	if settings["https"].(bool) == true {

		// host on https
		log.Fatal(srv.ListenAndServeTLS("tls/cert.pem", "tls/key.pem"))

	} else {

		// host on http
		log.Fatal(srv.ListenAndServe())

	}

}
