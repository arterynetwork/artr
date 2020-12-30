package types

import "fmt"

// Query endpoints supported by the schedule querier
const (
	QueryTasks  = "tasks"
	QueryParams = "params"
)

type QueryTasksParams struct {
	BlockHeight int64 `json:"block_height" yaml:"block_height"`
}

func (params QueryTasksParams) String() string {
	return fmt.Sprintf("BlockHeight: %d\n", params.BlockHeight)
}

func NewQueryTasksParams(blockHeight int64) QueryTasksParams {
	return QueryTasksParams{
		BlockHeight: blockHeight,
	}
}
