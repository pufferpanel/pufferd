package mkdir

import "github.com/pufferpanel/pufferd/programs/operations/ops"

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	target := op.OperationArgs["target"].(string)
	return &Mkdir{TargetFile: target}
}

func (of OperationFactory) Key() string {
	return "mkdir"
}

var Factory OperationFactory
