package download

import (
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
)

type OperationFactory struct {
	ops.OperationFactory
}

func (of OperationFactory) Create(op ops.CreateOperation) ops.Operation {
	files := common.ToStringArray(op.OperationArgs["files"])
	return &Download{Files: files}
}

func (of OperationFactory) Key() string {
	return "download"
}

var Factory OperationFactory
