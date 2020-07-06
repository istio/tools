package main

import (
	"fmt"
	"bytes"
	"math/rand"
)

type generator interface {
    generate(kind string, num int) (string, error)
}

type operationGenerator struct {
}

func (operationGenerator) generate(kind string, num int) (string, error) {
	return "", fmt.Errorf("unimplimented")
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(kind string, num int) (string, error) {
	rule := bytes.Buffer{}
	rule.WriteString("  - key: \n")

    for i := 0; i < num; i++ {
		rule.WriteString("    value")
	}

    return rule.String(), nil
}

type sourceGenerator struct {
}

func (sourceGenerator) generate(kind string, num int) (string, error) {
	rule := bytes.Buffer{}

    for i := 0; i < num; i++ {
		rule.WriteString("  - operation \n")
		
		// Currently hardcoded only for the methods call
		rule.WriteString("     methods: ")
		if (rand.Int() % 2 == 0) {
			rule.WriteString("[\"GET\"]\n")
		} else {
			rule.WriteString("[\"POST\"]\n")
		}
	}

    return rule.String(), nil
}

// May want JWTRules as a generator struct for RequestAuthentication
