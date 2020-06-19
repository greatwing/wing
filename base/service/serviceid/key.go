package serviceid

import (
	"strings"
)

const (
	serviceKeyPrefix = "/service"
)

func GetServiceKey(svcId string) string {
	return serviceKeyPrefix + svcId
}

func IsServiceKey(rawkey string) bool {

	return strings.HasPrefix(rawkey, serviceKeyPrefix)
}

func GetSvcIDByServiceKey(rawkey string) string {

	if IsServiceKey(rawkey) {
		return rawkey[len(serviceKeyPrefix):]
	}

	return ""
}
