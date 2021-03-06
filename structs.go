/*

discovery/structs.go

contains structs used to do things

written by superwhiskers, licensed under gnu agpl.
if you want a copy, go to http://www.gnu.org/licenses/

*/

package main

// result is a result that can or cannot be errored
type result struct {
	HasError   int    `xml:"has_error"`
	Version    int    `xml:"version"`
	Host       string `xml:"endpoint>host,omitempty"`
	APIHost    string `xml:"endpoint>api_host,omitempty"`
	PortalHost string `xml:"endpoint>portal_host,omitempty"`
	N3DSHost   string `xml:"endpoint>n3ds_host,omitempty"`
	Code       int    `xml:"code,omitempty"`
	ErrorCode  int    `xml:"error_code,omitempty"`
	Message    string `xml:"message,omitempty"`
}
