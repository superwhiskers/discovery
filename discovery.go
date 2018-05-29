/*

discovery/discovery.go

written by superwhiskers, licensed under gnu agpl.
if you want a copy, go to http://www.gnu.org/licenses/

*/

package main

import (
	// internals
	"encoding/xml"
	"fmt"
	"os"
	// externals
	//"github.com/labstack/echo"
	"gopkg.in/yaml.v2"
)

var maintenance interface{}
var bans interface{}

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
	var pullBansFromURL        bool

	// get some config sections from the config
	settings  := config["options"].(map[interface{}]interface{})
	endpoints := config["endpoints"].(map[interface{}]interface{})

	// endpoints
	host            := endpoints["discovery"].(string)
	apiHost         := endpoints["api"].(string)
	portalHost      := endpoints["wiiu"].(string)
	nintendo3dsHost := endpoints["3ds"].(string)
	
	// settings
	
	// maintenance is either a url to get a json
	// response from (like this:
	// { inMaintenance: false }
	// ) or a boolean
	switch settings["maintenance"].(type) {

	case string:
		pullMaintenanceFromURL = true	
		maintenance = settings["maintenance"].(string)

	case bool:
		pullMaintenanceFromURL = false
		maintenance = settings["maintenance"].(bool)

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
		bans = settings["bans"].(string)

	case []interface{}:
		pullBansFromURL = false
		bans = settings["bans"].([]interface{})

	default:
		fmt.Printf("[err]: the bans field in the options must either be a list of strings\n")
		fmt.Printf("       containing banned servicetokens or a url that points to an endpoint\n")
		fmt.Printf("       that will return a list of banned servicetokens...")
		os.Exit(1)

	}

	// construct the standard xml
	standardReturnXML := &result{
		HasError:   0,
		Version:    1,
		Host:       host,
		APIHost:    apiHost,
		PortalHost: portalHost,
		N3DSHost:   nintendo3dsHost,
	}

	// indent the xml
	output, err := xml.MarshalIndent(standardReturnXML, "  ",
"    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// output the xml
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
	fmt.Println()

	// output some stuff so go doesn't whine
	fmt.Printf("maintenance from url: %b\n", pullMaintenanceFromURL)
	fmt.Printf("bans from url: %b\n", pullBansFromURL)

	// load echo
	//e := echo.New()

	// set a get handler
	//e.GET("

}
