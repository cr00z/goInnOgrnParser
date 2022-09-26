package workerpool

import "time"

type ProcessFunction func(args interface{}) (interface{}, error)
type ResultFunction func(args interface{})

type Task struct {
	ID        int
	Args      interface{}
	ProcessFn ProcessFunction
	ResultFn  ResultFunction
}

type Result struct {
	ID       int
	Value    interface{}
	Error    error
	ResultFn ResultFunction
}

func (task Task) Process() Result {
	value, err := task.ProcessFn(task.Args)
	result := Result{
		ID:       task.ID,
		Value:    value,
		Error:    err,
		ResultFn: task.ResultFn,
	}
	return result
}

func (result Result) Process() {
	if result.Value != nil {
		result.ResultFn(result)
	} else {
		time.Sleep(100 * time.Millisecond)
	}
}
