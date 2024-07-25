package main

import (
	facility "github.com/rumus007/lldp-xir-generator"
	"gitlab.com/mergetb/xir/v0.3/go/build"
)
func main() {
	build.Run(facility.CreateFacility())
}