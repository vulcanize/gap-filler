package qlparser

import (
	"errors"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// List of errors
var (
	ErrNotFound = errors.New("Not found")
	ErrBadType  = errors.New("Bad type")
)

// QueryParams get graphql query names and params
func QueryParams(request []byte, names []string) (map[string][]*ast.Argument, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: source.NewSource(&source.Source{
			Body: request,
		}),
	})
	if err != nil {
		return nil, err
	}

	index := make(map[string]bool)
	for i := range names {
		index[names[i]] = true
	}

	data := make(map[string][]*ast.Argument)
	for i := range doc.Definitions {
		if doc.Definitions[i].GetKind() != "OperationDefinition" {
			continue
		}
		op := doc.Definitions[i].(*ast.OperationDefinition)
		if op.Operation != "query" && op.Kind != "field" {
			continue
		}
		for j := range op.SelectionSet.Selections {
			field, ok := op.SelectionSet.Selections[j].(*ast.Field)
			if ok && index[field.Name.Value] {
				data[field.Name.Value] = field.Arguments
			}
		}
	}

	return data, nil
}

// GetParams get graphql params from request for given query
func GetParams(request []byte, queryName string) ([]*ast.Argument, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: source.NewSource(&source.Source{
			Body: request,
		}),
	})
	if err != nil {
		return nil, err
	}
	for i := range doc.Definitions {
		if doc.Definitions[i].GetKind() != "OperationDefinition" {
			continue
		}
		op := doc.Definitions[i].(*ast.OperationDefinition)
		if op.Operation != "query" && op.Kind != "field" {
			continue
		}
		for j := range op.SelectionSet.Selections {
			field, ok := op.SelectionSet.Selections[j].(*ast.Field)
			if !ok || field.Name.Value != queryName {
				continue
			}
			return field.Arguments, nil
		}
	}
	return nil, ErrNotFound
}

// GetParam get graphql param from request for given query by argument name
func GetParam(request []byte, queryName string, argName string) (*ast.Argument, error) {
	args, err := GetParams(request, queryName)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, ErrNotFound
	}
	for i := range args {
		if args[i].Name.Value == argName {
			tmp := *args[i]
			return &tmp, nil
		}
	}
	return nil, ErrNotFound
}
