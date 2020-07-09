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
	// TODO implement
	condition := &authzpb.Rule{}
	return condition, nil
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(kind string, num int) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	listCondition := make([]*authzpb.Condition, 0)

	for i := 0; i < num; i++ {
		condition := &authzpb.Condition{}
		condition.Key = "request.headers[x-token]"
		values := []string{"admin", "guest"}
		condition.NotValues = values
		listCondition = append(listCondition, condition)
	}
	rule.When = listCondition
	return rule, nil
}

type sourceGenerator struct {
}

func (sourceGenerator) generate(kind string, num int) (*authzpb.Rule, error) {
	// TODO implement 
	condition := &authzpb.Rule{}
    return condition, nil
}
