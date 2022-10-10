package config

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func IsRunningAsComponent() bool {
	return os.Getenv("LW_COMPONENT_NAME") != ""
}

func RootCommandName() string {
	if IsRunningAsComponent() {
		return os.Getenv("LW_COMPONENT_NAME")
	}
	return "soluble"
}

func CommandInvocation() string {
	if IsRunningAsComponent() {
		return fmt.Sprintf("lacework %s", RootCommandName())
	}
	return "soluble"
}

func ExpandCommandInvocation(s string) string {
	t := template.New("expand")
	if _, err := t.Parse(s); err != nil {
		panic(err)
	}
	dat := map[string]interface{}{
		"RootCommand":       RootCommandName(),
		"CommandInvocation": CommandInvocation(),
	}
	w := &strings.Builder{}
	err := t.Execute(w, dat)
	if err != nil {
		panic(err)
	}
	return w.String()
}
