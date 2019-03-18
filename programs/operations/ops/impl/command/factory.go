package command

import (
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
)

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	cmds := common.ToStringArray(op.OperationArgs["commands"])
	return Command{Commands: cmds, Env: op.EnvironmentVariables}
}

func (of OperationFactory) Key() string {
	return "command"
}

var Factory OperationFactory