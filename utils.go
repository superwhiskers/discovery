/*

utils.go -
utilities for the program

*/

package main

import (
	// internals
	"encoding/hex"
	"io/ioutil"
	"net/http"
	// externals
	"golang.org/x/crypto/bcrypt"
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

// object hashing
func hash(object string, cost int) (string, error) {

	// use bcrypt
	bytes, err := bcrypt.GenerateFromPassword([]byte(object), cost)

	// return that data as hexadecimal
	return hex.EncodeToString(bytes), err

}

// compare a hash and an object
func compareHash(object, hash string) (bool, error) {

	// decode it from hexadecimal
	byteHash, err := hex.DecodeString(hash)

	// check for errors
	if err != nil {

		// return it
		return false, err

	}

	// compare them
	err = bcrypt.CompareHashAndPassword(byteHash, []byte(object))

	// return if they're the same
	return err == nil, nil

}
