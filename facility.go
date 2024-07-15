package facility

import (
	"fmt"
)

var config map[string]string

func init() {
  config = map[string]string{"name": "Go Program from facility folder"}
}

func CreateFacility() {
  fmt.Println(config["name"])
}