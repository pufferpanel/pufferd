package forgedl

import "github.com/pufferpanel/pufferd/programs/operations/ops"

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Key() string {
	return "forgedl"
}
func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	version := op.OperationArgs["version"].(string)
	filename := op.OperationArgs["target"].(string)

	return ForgeDl{Version: version, Filename: filename}
}

var Factory OperationFactory
