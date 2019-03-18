package move

import "github.com/pufferpanel/pufferd/programs/operations/ops"

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	source := op.OperationArgs["source"].(string)
	target := op.OperationArgs["target"].(string)
	return Move{SourceFile: source, TargetFile: target}
}

func (of OperationFactory) Key() string {
	return "move"
}

var Factory OperationFactory
