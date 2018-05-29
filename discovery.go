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
)

func main() {
	/*
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
	*/

	// construct the xml
	returnXML := &result{
		HasError:   0,
		Version:    1,
		Host:       "discovery.foxverse.xyz",
		APIHost:    "api-all.foxverse.xyz",
		PortalHost: "example.com",
		N3DSHost:   "3ds-all.foxverse.xyz",
	}

	// indent the xml
	output, err := xml.MarshalIndent(returnXML, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// output the xml
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
	fmt.Println()

}
