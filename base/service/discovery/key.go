package discovery

import (
	"strings"
)

const (
	ServiceKeyPrefix  = "/service"
	BalacneKeyPrefix  = "/balance"
	LocationKeyPrefix = "/location"
)

func getKey(rawkey, prefix string) string {
	if strings.HasPrefix(rawkey, prefix) {
		return rawkey
	}
	if !strings.HasPrefix(rawkey, "/") {
		rawkey = "/" + rawkey
	}
	return prefix + rawkey
}

func getKeyWithoutPrefix(key, prefix string) string {
	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}

	return key
}

func GetServiceKey(svcId string) string {
	return getKey(svcId, ServiceKeyPrefix)
}

func IsServiceKey(rawkey string) bool {
	return strings.HasPrefix(rawkey, ServiceKeyPrefix)
}

func GetSvcIDByServiceKey(rawkey string) string {
	return getKeyWithoutPrefix(rawkey, ServiceKeyPrefix)
}

func GetBalanceKey(svcId string) string {
	return getKey(svcId, BalacneKeyPrefix)
}

func IsBalanceKey(rawkey string) bool {
	return strings.HasPrefix(rawkey, BalacneKeyPrefix)
}

func GetSvcIDByBalanceKey(rawkey string) string {
	return getKeyWithoutPrefix(rawkey, BalacneKeyPrefix)
}

func GetLocationKey(svcId string) string {
	return getKey(svcId, LocationKeyPrefix)
}
