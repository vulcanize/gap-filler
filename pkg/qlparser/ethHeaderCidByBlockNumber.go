package qlparser

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// EthHeaderCidByBlockNumberArg detect graphql query `ethHeaderCidByBlockNumber`
func EthHeaderCidByBlockNumberArg(input io.Reader) (*big.Int, *int, error) {
	body, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, nil, err
	}
	doc, err := parser.Parse(parser.ParseParams{
		Source: source.NewSource(&source.Source{
			Body: body,
		}),
	})
	if err != nil {
		return nil, nil, err
	}
	var (
		n *big.Int = nil
		s *int     = nil
	)
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
			if !ok {
				continue
			}
			if field.Name.Value != "ethHeaderCidByBlockNumber" && len(field.Arguments) != 1 {
				continue
			}
			n, ok = new(big.Int).SetString(field.Arguments[0].Value.GetValue().(string), 10)
			if !ok {
				return nil, nil, fmt.Errorf("bad argument")
			}
			s = &j
		}
		if n != nil {
			break
		}
	}
	return n, s, nil
}
