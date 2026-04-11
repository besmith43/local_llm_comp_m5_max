#!/usr/bin/env bash


./reset.sh

templ generate

go run ./cmd/main.go
