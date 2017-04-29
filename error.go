package main

import "log"

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func LogIf(err error, logger log.Logger) {
	if err != nil {
		logger.Println(err)
	}
}