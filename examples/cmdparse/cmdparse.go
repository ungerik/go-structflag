package main

import (
	"encoding/json"
	"fmt"

	"github.com/ungerik/go-structflag"
)

// Config is the global configuration
type Config struct {
	String1 string
	String2 string `flag:"string2" json:"string2"`
	String3 string `default:"XXX"`

	Int int `default:"-1"`

	Bool1 bool `default:"false"`
	Bool2 bool `default:"true"`
	Bool3 bool
}

var config Config

func main() {
	structflag.StructVar(&config)
	_, err := structflag.Parse()
	if err != nil {
		panic(err)
	}
	j, err := json.MarshalIndent(&config, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}
