package main

import (
	authzpb "istio.io/api/security/v1beta1"
)

type generator interface {
    generate(kind string, num int) (*authzpb.Rule, error)
}

type operationGenerator struct {
}

func (operationGenerator) generate(kind string, num int) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listOperation []*authzpb.Rule_To

	for i := 0; i < num; i++ {
		operation := &authzpb.Rule_To{}
		operation.Operation = &authzpb.Operation{} 
		operation.Operation.Methods = []string{"GET", "HEAD"}
		operation.Operation.Paths = []string{"/admin"}
		listOperation = append(listOperation, operation)
	}
	rule.To = listOperation
	return rule, nil
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(kind string, num int) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listCondition []*authzpb.Condition

	for i := 0; i < num; i++ {
		condition := &authzpb.Condition{}
		condition.Key = "request.headers[x-token]"
		condition.NotValues = []string{"admin", "guest"}
		listCondition = append(listCondition, condition)
	}
	rule.When = listCondition
	return rule, nil
}

type sourceGenerator struct {
}

func (sourceGenerator) generate(kind string, num int) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listSource []*authzpb.Rule_From

	for i := 0; i < num; i++ {
		source := &authzpb.Rule_From{}
		source.Source = &authzpb.Source{}
		source.Source.Namespaces = []string{"twopods-istio"}
		listSource = append(listSource, source)
	}
	rule.From = listSource
    return rule, nil
}
