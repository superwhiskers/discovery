/*

utils.go -
utilities for the program

*/

package main

import (
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

// function to get data from a URL.
// based on https://www.github.com/thbar/golang-playground/blob/master/download-files.go
func get(url string) (string, error) {

	// attempt to download the contents
	res, err := http.Get(url)

	// error handling
	if err != nil {

		// return an empty string, and the error
		return "", err

	}

	// close request body stream once finished
	defer res.Body.Close()

	// read all data from body
	data, err := ioutil.ReadAll(res.Body)

	// error handling
	if err != nil {

		// return an empty string, and the error
		return "", err

	}

	// convert the bytes to a string
	ret := string(data[:])

	// return the request response
	return ret, nil

}

// this struct contains all of the data a parampack would contain in
// a go-compatible format
type paramPack struct {
	TitleID            string
	AccessKey          string
	PlatformID         int
	RegionID           int
	LanguageID         int
	CountryID          int
	AreaID             int
	NetworkRestriction int
	FriendRestriction  int
	RatingRestriction  int
	RatingOrganization int
	TransferableID     string
	TimezoneName       string
	UTCOffset          int
	RemasterVersion    int
}

var nilParamPack = paramPack{
	TitleID:            "0000000000000000",
	AccessKey:          "",
	PlatformID:         0,
	RegionID:           0,
	LanguageID:         0,
	CountryID:          0,
	AreaID:             0,
	NetworkRestriction: 0,
	FriendRestriction:  0,
	RatingRestriction:  0,
	RatingOrganization: 0,
	TransferableID:     "",
	TimezoneName:       "",
	UTCOffset:          0,
	RemasterVersion:    0,
}

// nintendo servicetoken decoder
func decodeServiceToken(serviceToken string) (string, error) {

	// decode it from base64
	decodedServiceToken, err := base64.StdEncoding.DecodeString(serviceToken)

	// if there is an error
	if err != nil {

		// exit the function and return the error
		return "", err

	}

	// temporary workaround for now
	return hex.EncodeToString(decodedServiceToken), nil

}

// turn a stringified tid to a usable one (wiiu)
func unstringifyWiiUTID(stringifiedTID string) (string, error) {

	// convert the string to an int
	intTID, err := strconv.ParseUint(stringifiedTID, 10, 64)

	// return an error if there is one
	if err != nil {

		// return the error
		return "", err

	}

	// convert it back to a string
	tid := strconv.FormatUint(intTID, 16)

	// pad it to 16 characters
	if len(tid) != 16 {

		// loop until it is 16 characters
		for x := len(tid); x < 16; x++ {

			// pad with zeroes
			tid = "0" + tid

		}

	}

	// return the zero padded tid
	return tid, nil

}

// nintendo parampack decoder
func decodeParampack(parampack string) (paramPack, error) {

	// strip spaces
	paramStripped := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, parampack)

	// decode it from base64
	decodedParampack, err := base64.StdEncoding.DecodeString(paramStripped)

	// if there is an error
	if err != nil {

		// exit the function and return the error
		return nilParamPack, err

	}

	// split it by backslashes
	splitParampack := strings.Split(string(decodedParampack[:]), "\\")

	// variables to be placed into the struct
	titleID := "0000000000000000"
	accessKey := ""
	platformID := 0
	regionID := 0
	languageID := 0
	countryID := 0
	areaID := 0
	networkRestriction := 0
	friendRestriction := 0
	ratingRestriction := 0
	ratingOrganization := 0
	transferableID := ""
	timezoneName := ""
	utcOffset := 0
	remasterVersion := 0

	// iterate over the split parampack
	for ind, ele := range splitParampack {

		// check if it is one of the parts of a parameter pack
		// and assign its value to the corresponding variable
		switch ele {

		case "title_id":

			// titleids are special
			/*
				unstringifiedTID, err := unstringifyTID(splitParampack[ind+1])
				if err != nil {
					unstringifiedTID = "0000000000000000"
				}
			*/
			titleID = splitParampack[ind+1]

		case "access_key":
			accessKey = splitParampack[ind+1]

		case "platform_id":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			platformID = tmp

		case "region_id":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			regionID = tmp

		case "language_id":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			languageID = tmp

		case "country_id":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			countryID = tmp

		case "area_id":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			areaID = tmp

		case "network_restriction":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			networkRestriction = tmp

		case "friend_restriction":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			friendRestriction = tmp

		case "rating_restriction":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			ratingRestriction = tmp

		case "rating_organization":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			ratingOrganization = tmp

		case "transferable_id":
			transferableID = splitParampack[ind+1]

		case "tz_name":
			timezoneName = splitParampack[ind+1]

		case "utc_offset":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			utcOffset = tmp

		case "remaster_version":
			tmp, err := strconv.Atoi(splitParampack[ind+1])
			if err != nil {
				tmp = 0
			}
			remasterVersion = tmp

		}

	}

	// finally, formulate a parampack struct
	returnableParampack := paramPack{
		TitleID:            titleID,
		AccessKey:          accessKey,
		PlatformID:         platformID,
		RegionID:           regionID,
		LanguageID:         languageID,
		CountryID:          countryID,
		AreaID:             areaID,
		NetworkRestriction: networkRestriction,
		FriendRestriction:  friendRestriction,
		RatingRestriction:  ratingRestriction,
		RatingOrganization: ratingOrganization,
		TransferableID:     transferableID,
		TimezoneName:       timezoneName,
		UTCOffset:          utcOffset,
		RemasterVersion:    remasterVersion,
	}

	// and return it
	return returnableParampack, nil

}
