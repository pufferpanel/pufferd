package spongeforgedl

import "github.com/pufferpanel/pufferd/programs/operations/ops"

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Key() string {
	return "spongeforgedl"
}
func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	releaseType, ok := op.OperationArgs["releaseType"].(string)
	if !ok {
		releaseType = "recommended"
	}

	return SpongeForgeDl{ReleaseType: releaseType}
}

var Factory OperationFactory
