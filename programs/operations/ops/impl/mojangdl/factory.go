package mojangdl

import "github.com/pufferpanel/pufferd/programs/operations/ops"

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	version := op.OperationArgs["version"].(string)
	target := op.OperationArgs["target"].(string)

	return MojangDl{Version: version, Target: target}
}

func (of OperationFactory) Key() string {
	return "mojangdl"
}

var Factory OperationFactory
