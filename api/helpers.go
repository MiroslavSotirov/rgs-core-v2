package api

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"net/http"
	"strings"
)

// GetURLScheme returns https if TLS is present else returns http
func GetURLScheme(r *http.Request) string {
	//return r.URL.Scheme+"//"
	//todo: find another way to handle this because TLS is missing sometimes from request object when it shouldn't be
	if config.GlobalConfig.Local == true {
		return "http://"
	}
	return "https://"
}

func processAuthorization(request *http.Request) (string, rgserror.IRGSError) {

	tokenInfo := strings.Split(request.Header.Get("Authorization"), " ")
	switch tokenInfo[0] {
	default:
		return "", rgserror.ErrInvalidCredentials
	case "MAVERICK-Host-Token":
		logger.Debugf("Auth Token: %v; Auth Header: %v", tokenInfo[1], request.Header.Get("Authorization"))
	case "DUMMY-MAVERICK-Host-Token":
		logger.Debugf("Auth Token: %v; Auth Header: %v", tokenInfo[1], request.Header.Get("Authorization"))
	}
	if strings.Contains(tokenInfo[1], "token=\"") {
		return strings.Split(tokenInfo[1], "\"")[1], nil
	}

	return tokenInfo[1], nil
}
