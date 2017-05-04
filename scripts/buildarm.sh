#!/bin/bash
GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc go build -o out/mongoose-bot
