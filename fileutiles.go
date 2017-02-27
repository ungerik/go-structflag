package structflag

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// LoadFile loads a struct from a JSON or XML file.
// The file type is determined by the file extension.
func LoadFile(filename string, structPtr interface{}) error {
	// Load and unmarshal struct from file
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return LoadJSON(filename, structPtr)
	case ".xml":
		return LoadXML(filename, structPtr)
	}
	return errors.New("File extension not supported: " + ext)
}

// LoadXML loads a struct from a XML file
func LoadXML(filename string, structPtr interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, structPtr)
}

// SaveXML saves a struct as a XML file
func SaveXML(filename string, structPtr interface{}, indent ...string) error {
	data, err := xml.MarshalIndent(structPtr, "", strings.Join(indent, ""))
	if err != nil {
		return err
	}
	data = append([]byte(xml.Header), data...)
	return ioutil.WriteFile(filename, data, 0660)
}

// LoadJSON loads a struct from a JSON file
func LoadJSON(filename string, structPtr interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, structPtr)
}

// SaveJSON saves a struct as a JSON file
func SaveJSON(filename string, structPtr interface{}, indent ...string) error {
	data, err := json.MarshalIndent(structPtr, "", strings.Join(indent, ""))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0660)
}
