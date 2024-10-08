// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Processing is an autogenerated mock type for the Processing type
type Processing struct {
	mock.Mock
}

type Processing_Expecter struct {
	mock *mock.Mock
}

func (_m *Processing) EXPECT() *Processing_Expecter {
	return &Processing_Expecter{mock: &_m.Mock}
}

// Processing provides a mock function with given fields: ctx, input
func (_m *Processing) Processing(ctx context.Context, input interface{}) {
	_m.Called(ctx, input)
}

// Processing_Processing_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Processing'
type Processing_Processing_Call struct {
	*mock.Call
}

// Processing is a helper method to define mock.On call
//   - ctx context.Context
//   - input interface{}
func (_e *Processing_Expecter) Processing(ctx interface{}, input interface{}) *Processing_Processing_Call {
	return &Processing_Processing_Call{Call: _e.mock.On("Processing", ctx, input)}
}

func (_c *Processing_Processing_Call) Run(run func(ctx context.Context, input interface{})) *Processing_Processing_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *Processing_Processing_Call) Return() *Processing_Processing_Call {
	_c.Call.Return()
	return _c
}

func (_c *Processing_Processing_Call) RunAndReturn(run func(context.Context, interface{})) *Processing_Processing_Call {
	_c.Call.Return(run)
	return _c
}

// Result provides a mock function with given fields:
func (_m *Processing) Result() chan interface{} {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Result")
	}

	var r0 chan interface{}
	if rf, ok := ret.Get(0).(func() chan interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan interface{})
		}
	}

	return r0
}

// Processing_Result_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Result'
type Processing_Result_Call struct {
	*mock.Call
}

// Result is a helper method to define mock.On call
func (_e *Processing_Expecter) Result() *Processing_Result_Call {
	return &Processing_Result_Call{Call: _e.mock.On("Result")}
}

func (_c *Processing_Result_Call) Run(run func()) *Processing_Result_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Processing_Result_Call) Return(_a0 chan interface{}) *Processing_Result_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Processing_Result_Call) RunAndReturn(run func() chan interface{}) *Processing_Result_Call {
	_c.Call.Return(run)
	return _c
}

// NewProcessing creates a new instance of Processing. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProcessing(t interface {
	mock.TestingT
	Cleanup(func())
}) *Processing {
	mock := &Processing{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
