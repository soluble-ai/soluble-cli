package model

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/print"
)

type ColumnComputerType string

var columnComputers = map[string]print.ColumnComputer{}

func (t ColumnComputerType) validate() error {
	if _, ok := columnComputers[string(t)]; !ok {
		return fmt.Errorf("invalid column computer %s", t)
	}
	return nil
}

func (t ColumnComputerType) GetComlumnComputer() print.ColumnComputer {
	return columnComputers[string(t)]
}

func RegisterColumnComputer(name string, c print.ColumnComputer) {
	columnComputers[name] = c
}
