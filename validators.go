package paymail

import (
	"net"
	"regexp"
	"strings"
)

var (
	dnsRegEx   = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[._]?$`)
	emailRegEx = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,6}$`)
)

// IsValidHost checks if the string is a valid IP (both v4 and v6) or a valid DNS name
func IsValidHost(host string) bool {
	return IsValidIP(host) || IsValidDNSName(host)
}

// IsValidIP checks if a string is either IP version 4 or 6. Alias for `net.ParseIP`
func IsValidIP(ipAddress string) bool {
	return net.ParseIP(ipAddress) != nil
}

// IsValidDNSName will validate the given string as a DNS name
func IsValidDNSName(dnsName string) bool {
	if dnsName == "" || len(strings.Replace(dnsName, ".", "", -1)) > 255 {
		return false
	}
	return !IsValidIP(dnsName) && dnsRegEx.MatchString(dnsName)
}

// IsValidEmail will validate the given string as an email address
func IsValidEmail(e string) bool {
	return emailRegEx.MatchString(e)
}
