/*

utils.go -
utilities for the program

*/

package main

import (
	"io/ioutil"
	"net/http"
)

// function to get data from a URL.
// based on https://www.github.com/thbar/golang-playground/blob/master/download-files.go
func get(url string) (string, string, error) {

	// attempt to download the contents
	res, err := http.Get(url)

	// error handling
	if err != nil {

		// return an empty string, and the error
		return "", "", err

	}

	// close request body stream once finished
	defer res.Body.Close()

	// read all data from body
	data, err := ioutil.ReadAll(res.Body)

	// error handling
	if err != nil {

		// return an empty string, and the error
		return "", "", err

	}

	// convert the bytes to a string
	ret := string(data[:])

	// return the request response
	return ret, res.Header.Get("Content-Type"), nil

}
