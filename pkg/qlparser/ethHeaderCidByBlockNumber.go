package qlparser

import (
	"fmt"
	"math/big"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// EthHeaderCidByBlockNumberArg detect graphql query `ethHeaderCidByBlockNumber`
func EthHeaderCidByBlockNumberArg(query []byte) (*big.Int, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: source.NewSource(&source.Source{
			Body: query,
		}),
	})
	if err != nil {
		return nil, err
	}
	var n *big.Int = nil
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
			if field.Name.Value != "ethHeaderCidByBlockNumber" || len(field.Arguments) != 1 {
				continue
			}
			n, ok = new(big.Int).SetString(field.Arguments[0].Value.GetValue().(string), 10)
			if !ok {
				return nil, fmt.Errorf("bad argument")
			}
		}
		if n != nil {
			break
		}
	}
	return n, nil
}
