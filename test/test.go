// Package test provides tools for testing the library
package test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/gobl"
	ubl "github.com/invopop/gobl.ubl"
	"github.com/invopop/gobl/bill"
	// "github.com/lestrrat-go/libxml2"
	// "github.com/lestrrat-go/libxml2/xsd"
)

// NewDocumentFrom creates a cii Document from a GOBL file in the `test/data` folder
func NewDocumentFrom(name string) (*ubl.Document, error) {
	env, err := LoadTestEnvelope(name)
	if err != nil {
		return nil, err
	}

	return ubl.NewDocument(env)
}

// LoadTestXMLDoc returns a CII XMLDoc from a file in the test data folder
func LoadTestXMLDoc(name string) (*structs.Invoice, error) {
	src, err := os.Open(filepath.Join(GetConversionTypePath("*.xml"), name))
	if err != nil {
		return nil, err
	}
	defer src.Close()

	inData, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	doc := new(structs.Invoice)
	if err := xml.Unmarshal(inData, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// LoadTestInvoice returns a GOBL Invoice from a file in the `test/data` folder
func LoadTestInvoice(name string) (*bill.Invoice, error) {
	env, err := LoadTestEnvelope(name)
	if err != nil {
		return nil, err
	}

	return env.Extract().(*bill.Invoice), nil
}

// LoadTestEnvelope returns a GOBL Envelope from a file in the `test/data` folder
func LoadTestEnvelope(name string) (*gobl.Envelope, error) {
	src, _ := os.Open(filepath.Join(GetConversionTypePath("*.json"), name))
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, err
	}
	env := new(gobl.Envelope)
	if err := json.Unmarshal(buf.Bytes(), env); err != nil {
		return nil, err
	}

	return env, nil
}

// GenerateXInvoiceFrom returns a XInvoice Document from a GOBL Invoice
func GenerateXInvoiceFrom(inv *bill.Invoice) (*ubl.Document, error) {
	env, err := gobl.Envelop(inv)
	if err != nil {
		return nil, err
	}

	return ubl.NewDocument(env)
}

// LoadOutputFile returns byte data from a file in the `test/data/out` folder
func LoadOutputFile(name string) ([]byte, error) {
	var pattern string
	if strings.HasSuffix(name, ".json") {
		pattern = "*.xml"
	} else {
		pattern = "*.json"
	}
	src, _ := os.Open(filepath.Join(GetOutPath(pattern), name))

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SaveOutputFile writes byte data to a file in the `test/data/out` folder
func SaveOutputFile(name string, data []byte) error {
	var pattern string
	if strings.HasSuffix(name, ".json") {
		pattern = "*.xml"
	} else {
		pattern = "*.json"
	}
	return os.WriteFile(filepath.Join(GetOutPath(pattern), name), data, 0644)
}

// LoadSchema returns a XSD Schema from a file in the `test/data/schema` folder
// func LoadSchema(name string) (*xsd.Schema, error) {
// 	return xsd.ParseFromFile(filepath.Join(GetSchemaPath(), name))
// }

// // ValidateXML validates a XML document against a XSD Schema
// func ValidateXML(schema *xsd.Schema, data []byte) error {
// 	xmlDoc, err := libxml2.Parse(data)
// 	if err != nil {
// 		return err
// 	}

// 	err = schema.Validate(xmlDoc)
// 	if err != nil {
// 		return err.(xsd.SchemaValidationError).Errors()[0]
// 	}

// 	return nil
// }

// GetDataGlob returns a list of files in the `test/data` folder that match the pattern
func GetDataGlob(pattern string) ([]string, error) {
	return filepath.Glob(filepath.Join(GetConversionTypePath(pattern), pattern))
}

// GetSchemaPath returns the path to the `test/data/schema` folder
func GetSchemaPath(pattern string) string {
	return filepath.Join(GetConversionTypePath(pattern), "schema")
}

// GetOutPath returns the path to the `test/data/out` folder
func GetOutPath(pattern string) string {
	return filepath.Join(GetConversionTypePath(pattern), "out")
}

// GetDataPath returns the path to the `test/data` folder
func GetDataPath() string {
	return filepath.Join(GetTestPath(), "data")
}

func GetConversionTypePath(pattern string) string {
	if pattern == "*.xml" {
		return filepath.Join(GetDataPath(), "ctog")
	}
	return filepath.Join(GetDataPath(), "gtoc")
}

// GetTestPath returns the path to the `test` folder
func GetTestPath() string {
	return filepath.Join(getRootFolder(), "test")
}

// TODO: adapt to new folder structure
func getRootFolder() string {
	cwd, _ := os.Getwd()

	for !isRootFolder(cwd) {
		cwd = removeLastEntry(cwd)
	}
	return cwd
}

func isRootFolder(dir string) bool {
	files, _ := os.ReadDir(dir)

	for _, file := range files {
		if file.Name() == "go.mod" {
			return true
		}
	}

	return false
}

func removeLastEntry(dir string) string {
	lastEntry := "/" + filepath.Base(dir)
	i := strings.LastIndex(dir, lastEntry)
	return dir[:i]
}
