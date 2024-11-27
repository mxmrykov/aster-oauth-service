package utils

import (
	"strings"
)

type DeviceInfo struct {
	DeviceName    string
	OSName        string
	OSVersion     string
	Client        string
	ClientVersion string
}

func GetDeviceInfo(userAgent string) DeviceInfo {
	deviceInfo := DeviceInfo{}

	start := strings.Index(userAgent, "(")
	end := strings.Index(userAgent, ")")
	if start != -1 && end != -1 && end > start {
		deviceInfo.DeviceName = strings.TrimSpace(userAgent[start+1 : end])
	}

	osParts := []string{"Windows NT", "Mac OS X", "Linux", "Android", "iPhone OS", "iPad OS", "iOS"}
	for _, os := range osParts {
		if strings.Contains(userAgent, os) {
			deviceInfo.OSName = os
			osVersionStart := strings.Index(userAgent, os) + len(os)
			if osVersionStart < len(userAgent) {
				osVersionEnd := strings.IndexAny(userAgent[osVersionStart:], " ;") // Find space or semicolon after OS
				if osVersionEnd != -1 {
					deviceInfo.OSVersion = strings.TrimSpace(userAgent[osVersionStart : osVersionStart+osVersionEnd])
				}
			}
			break
		}
	}

	clientParts := []string{"Firefox", "Chrome", "Safari", "Opera", "MSIE", "Trident"}
	for _, client := range clientParts {
		if strings.Contains(userAgent, client) {
			deviceInfo.Client = client
			clientVersionStart := strings.Index(userAgent, client) + len(client)
			if clientVersionStart < len(userAgent) {
				versionSeparator := userAgent[clientVersionStart]
				if versionSeparator == '/' || versionSeparator == ' ' {
					clientVersionStart++
				}
				clientVersionEnd := strings.IndexAny(userAgent[clientVersionStart:], " ;") // Find space or semicolon after version
				if clientVersionEnd != -1 {
					deviceInfo.ClientVersion = strings.TrimSpace(userAgent[clientVersionStart : clientVersionStart+clientVersionEnd])
				}
			}
			break
		}
	}

	return deviceInfo
}
