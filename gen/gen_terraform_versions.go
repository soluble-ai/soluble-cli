package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const terraform_ = "terraform_"

func gen() error {
	resp, err := http.Get("https://releases.hashicorp.com/terraform/")
	if err != nil {
		return err
	}
	versions := &bytes.Buffer{}
	scanner := bufio.NewScanner(resp.Body)
	versionRegexp := regexp.MustCompile(`<a href="[^"]+">(.+)</a>`)
	for scanner.Scan() {
		line := scanner.Text()
		m := versionRegexp.FindStringSubmatch(line)
		if m != nil {
			v := m[1]
			if strings.HasPrefix(v, terraform_) && !strings.ContainsRune(v, '-') {
				fmt.Fprintf(versions, "%s\n", v[len(terraform_):])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile("terraform_versions.txt", versions.Bytes(), 0600)
}

func main() {
	if err := gen(); err != nil {
		log.Fatal("generate failed:", err)
	}
}
