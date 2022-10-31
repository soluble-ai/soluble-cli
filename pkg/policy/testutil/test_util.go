package testutil

import (
	"bytes"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func readYamlFile(path string) map[interface{}]interface{} {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[interface{}]interface{})

	if err := yaml.Unmarshal(f, &data); err != nil {
		log.Fatal(err)
	}
	return data
}

func CompareYamlFiles(file1, file2 string) int {
	yaml1 := readYamlFile(file1)
	yaml2 := readYamlFile(file2)

	data1, _ := yaml.Marshal(yaml1)
	data2, _ := yaml.Marshal(yaml2)

	diff := bytes.Compare(data1, data2)
	return diff
}
