package serviceid

import (
	"strings"
)

const (
	ServiceKeyPrefix = "/service"
)

func GetServiceKey(svcId string) string {
	return ServiceKeyPrefix + svcId
}

func IsServiceKey(rawkey string) bool {

	return strings.HasPrefix(rawkey, ServiceKeyPrefix)
}

func GetSvcIDByServiceKey(rawkey string) string {

	if IsServiceKey(rawkey) {
		return rawkey[len(ServiceKeyPrefix):]
	}

	return ""
}
