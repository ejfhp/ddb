package main

import "embed"

const (
	imgLogo = "resource/logo.png"
)

//go:embed resource/*
var resources embed.FS
