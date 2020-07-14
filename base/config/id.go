package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 全局唯一的svcid 格式:  /svcName/svcGroup/svcIndex

// 构造服务ID
func MakeSvcID(svcName string, svcIndex int, svcGroup string) string {
	return fmt.Sprintf("/%s/%s/%d", svcName, svcGroup, svcIndex)
}

// 构造指定服务的ID
func MakeLocalSvcID(svcName string) string {
	return MakeSvcID(svcName, GetSvcIndex(), GetSvcGroup())
}

// 获得本进程的服务id
func GetLocalSvcID() string {
	return MakeLocalSvcID(GetSvcName())
}

// 解析服务id
func ParseSvcID(svcid string) (svcName string, svcIndex int, svcGroup string, err error) {

	result := strings.FieldsFunc(svcid, func(c rune) bool {
		if c == '/' {
			return true
		} else {
			return false
		}
	})

	if len(result) != 3 {
		err = errors.New("service id format is wrong")
		return
	}

	svcName = result[0]
	svcGroup = result[1]
	svcIndex, err = strconv.Atoi(result[2])
	return
}
