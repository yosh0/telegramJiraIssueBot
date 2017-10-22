package main

import (
	"fmt"
	"strconv"
	"runtime"
)

func fName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func TimeString(st float64) string {
	var diff int64 = int64(st)
	hour := (diff%86400)/3600
	min := (diff%3600)/60
	sec := (diff%3600)%60
	hs := strconv.FormatInt(hour, 10)
	ms := strconv.FormatInt(min, 10)
	ss := strconv.FormatInt(sec, 10)
	return fmt.Sprintf("%02s:%02s:%02s", hs, ms, ss)
}
