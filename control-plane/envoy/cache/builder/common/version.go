package common

import (
	"strconv"
	"time"
)

func GenerateResourceVersion() string {
	return VersionToString(time.Now().UnixNano())
}

func VersionToString(version int64) string {
	return strconv.FormatInt(version, 10)
}

func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func makeTimestampMilli() int64 {
	return unixMilli(time.Now())
}
