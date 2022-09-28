// Code generated by mocktail; DO NOT EDIT.

package commands

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
)

// storeMock mock of Store.
type storeMock struct{ mock.Mock }

// newStoreMock creates a new storeMock.
func newStoreMock(tb testing.TB) *storeMock {
	tb.Helper()

	m := &storeMock{}
	m.Mock.Test(tb)

	tb.Cleanup(func() { m.AssertExpectations(tb) })

	return m
}

func (_m *storeMock) ListPendingCommands(_ context.Context) ([]platform.Command, error) {
	_ret := _m.Called()

	_ra0, _ := _ret.Get(0).([]platform.Command)
	_rb1 := _ret.Error(1)

	return _ra0, _rb1
}

func (_m *storeMock) OnListPendingCommands() *storeListPendingCommandsCall {
	return &storeListPendingCommandsCall{Call: _m.Mock.On("ListPendingCommands"), Parent: _m}
}

func (_m *storeMock) OnListPendingCommandsRaw() *storeListPendingCommandsCall {
	return &storeListPendingCommandsCall{Call: _m.Mock.On("ListPendingCommands"), Parent: _m}
}

type storeListPendingCommandsCall struct {
	*mock.Call
	Parent *storeMock
}

func (_c *storeListPendingCommandsCall) Panic(msg string) *storeListPendingCommandsCall {
	_c.Call = _c.Call.Panic(msg)
	return _c
}

func (_c *storeListPendingCommandsCall) Once() *storeListPendingCommandsCall {
	_c.Call = _c.Call.Once()
	return _c
}

func (_c *storeListPendingCommandsCall) Twice() *storeListPendingCommandsCall {
	_c.Call = _c.Call.Twice()
	return _c
}

func (_c *storeListPendingCommandsCall) Times(i int) *storeListPendingCommandsCall {
	_c.Call = _c.Call.Times(i)
	return _c
}

func (_c *storeListPendingCommandsCall) WaitUntil(w <-chan time.Time) *storeListPendingCommandsCall {
	_c.Call = _c.Call.WaitUntil(w)
	return _c
}

func (_c *storeListPendingCommandsCall) After(d time.Duration) *storeListPendingCommandsCall {
	_c.Call = _c.Call.After(d)
	return _c
}

func (_c *storeListPendingCommandsCall) Run(fn func(args mock.Arguments)) *storeListPendingCommandsCall {
	_c.Call = _c.Call.Run(fn)
	return _c
}

func (_c *storeListPendingCommandsCall) Maybe() *storeListPendingCommandsCall {
	_c.Call = _c.Call.Maybe()
	return _c
}

func (_c *storeListPendingCommandsCall) TypedReturns(a []platform.Command, b error) *storeListPendingCommandsCall {
	_c.Call = _c.Return(a, b)
	return _c
}

func (_c *storeListPendingCommandsCall) ReturnsFn(fn func() ([]platform.Command, error)) *storeListPendingCommandsCall {
	_c.Call = _c.Return(fn)
	return _c
}

func (_c *storeListPendingCommandsCall) TypedRun(fn func()) *storeListPendingCommandsCall {
	_c.Call = _c.Call.Run(func(args mock.Arguments) {
		fn()
	})
	return _c
}

func (_c *storeListPendingCommandsCall) OnListPendingCommands() *storeListPendingCommandsCall {
	return _c.Parent.OnListPendingCommands()
}

func (_c *storeListPendingCommandsCall) OnSendCommandReports(reports []platform.CommandReport) *storeSendCommandReportsCall {
	return _c.Parent.OnSendCommandReports(reports)
}

func (_c *storeListPendingCommandsCall) OnListPendingCommandsRaw() *storeListPendingCommandsCall {
	return _c.Parent.OnListPendingCommandsRaw()
}

func (_c *storeListPendingCommandsCall) OnSendCommandReportsRaw(reports interface{}) *storeSendCommandReportsCall {
	return _c.Parent.OnSendCommandReportsRaw(reports)
}

func (_m *storeMock) SendCommandReports(_ context.Context, reports []platform.CommandReport) error {
	_ret := _m.Called(reports)

	if _rf, ok := _ret.Get(0).(func([]platform.CommandReport) error); ok {
		return _rf(reports)
	}

	_ra0 := _ret.Error(0)

	return _ra0
}

func (_m *storeMock) OnSendCommandReports(reports []platform.CommandReport) *storeSendCommandReportsCall {
	return &storeSendCommandReportsCall{Call: _m.Mock.On("SendCommandReports", reports), Parent: _m}
}

func (_m *storeMock) OnSendCommandReportsRaw(reports interface{}) *storeSendCommandReportsCall {
	return &storeSendCommandReportsCall{Call: _m.Mock.On("SendCommandReports", reports), Parent: _m}
}

type storeSendCommandReportsCall struct {
	*mock.Call
	Parent *storeMock
}

func (_c *storeSendCommandReportsCall) Panic(msg string) *storeSendCommandReportsCall {
	_c.Call = _c.Call.Panic(msg)
	return _c
}

func (_c *storeSendCommandReportsCall) Once() *storeSendCommandReportsCall {
	_c.Call = _c.Call.Once()
	return _c
}

func (_c *storeSendCommandReportsCall) Twice() *storeSendCommandReportsCall {
	_c.Call = _c.Call.Twice()
	return _c
}

func (_c *storeSendCommandReportsCall) Times(i int) *storeSendCommandReportsCall {
	_c.Call = _c.Call.Times(i)
	return _c
}

func (_c *storeSendCommandReportsCall) WaitUntil(w <-chan time.Time) *storeSendCommandReportsCall {
	_c.Call = _c.Call.WaitUntil(w)
	return _c
}

func (_c *storeSendCommandReportsCall) After(d time.Duration) *storeSendCommandReportsCall {
	_c.Call = _c.Call.After(d)
	return _c
}

func (_c *storeSendCommandReportsCall) Run(fn func(args mock.Arguments)) *storeSendCommandReportsCall {
	_c.Call = _c.Call.Run(fn)
	return _c
}

func (_c *storeSendCommandReportsCall) Maybe() *storeSendCommandReportsCall {
	_c.Call = _c.Call.Maybe()
	return _c
}

func (_c *storeSendCommandReportsCall) TypedReturns(a error) *storeSendCommandReportsCall {
	_c.Call = _c.Return(a)
	return _c
}

func (_c *storeSendCommandReportsCall) ReturnsFn(fn func([]platform.CommandReport) error) *storeSendCommandReportsCall {
	_c.Call = _c.Return(fn)
	return _c
}

func (_c *storeSendCommandReportsCall) TypedRun(fn func([]platform.CommandReport)) *storeSendCommandReportsCall {
	_c.Call = _c.Call.Run(func(args mock.Arguments) {
		_reports, _ := args.Get(0).([]platform.CommandReport)
		fn(_reports)
	})
	return _c
}

func (_c *storeSendCommandReportsCall) OnListPendingCommands() *storeListPendingCommandsCall {
	return _c.Parent.OnListPendingCommands()
}

func (_c *storeSendCommandReportsCall) OnSendCommandReports(reports []platform.CommandReport) *storeSendCommandReportsCall {
	return _c.Parent.OnSendCommandReports(reports)
}

func (_c *storeSendCommandReportsCall) OnListPendingCommandsRaw() *storeListPendingCommandsCall {
	return _c.Parent.OnListPendingCommandsRaw()
}

func (_c *storeSendCommandReportsCall) OnSendCommandReportsRaw(reports interface{}) *storeSendCommandReportsCall {
	return _c.Parent.OnSendCommandReportsRaw(reports)
}

// handlerMock mock of Handler.
type handlerMock struct{ mock.Mock }

// newHandlerMock creates a new handlerMock.
func newHandlerMock(tb testing.TB) *handlerMock {
	tb.Helper()

	m := &handlerMock{}
	m.Mock.Test(tb)

	tb.Cleanup(func() { m.AssertExpectations(tb) })

	return m
}

func (_m *handlerMock) Handle(_ context.Context, id string, requestedAt time.Time, data json.RawMessage) *platform.CommandReport {
	_ret := _m.Called(id, requestedAt, data)

	if _rf, ok := _ret.Get(0).(func(string, time.Time, json.RawMessage) *platform.CommandReport); ok {
		return _rf(id, requestedAt, data)
	}

	_ra0, _ := _ret.Get(0).(*platform.CommandReport)

	return _ra0
}

func (_m *handlerMock) OnHandle(id string, requestedAt time.Time, data json.RawMessage) *handlerHandleCall {
	return &handlerHandleCall{Call: _m.Mock.On("Handle", id, requestedAt, data), Parent: _m}
}

func (_m *handlerMock) OnHandleRaw(id interface{}, requestedAt interface{}, data interface{}) *handlerHandleCall {
	return &handlerHandleCall{Call: _m.Mock.On("Handle", id, requestedAt, data), Parent: _m}
}

type handlerHandleCall struct {
	*mock.Call
	Parent *handlerMock
}

func (_c *handlerHandleCall) Panic(msg string) *handlerHandleCall {
	_c.Call = _c.Call.Panic(msg)
	return _c
}

func (_c *handlerHandleCall) Once() *handlerHandleCall {
	_c.Call = _c.Call.Once()
	return _c
}

func (_c *handlerHandleCall) Twice() *handlerHandleCall {
	_c.Call = _c.Call.Twice()
	return _c
}

func (_c *handlerHandleCall) Times(i int) *handlerHandleCall {
	_c.Call = _c.Call.Times(i)
	return _c
}

func (_c *handlerHandleCall) WaitUntil(w <-chan time.Time) *handlerHandleCall {
	_c.Call = _c.Call.WaitUntil(w)
	return _c
}

func (_c *handlerHandleCall) After(d time.Duration) *handlerHandleCall {
	_c.Call = _c.Call.After(d)
	return _c
}

func (_c *handlerHandleCall) Run(fn func(args mock.Arguments)) *handlerHandleCall {
	_c.Call = _c.Call.Run(fn)
	return _c
}

func (_c *handlerHandleCall) Maybe() *handlerHandleCall {
	_c.Call = _c.Call.Maybe()
	return _c
}

func (_c *handlerHandleCall) TypedReturns(a *platform.CommandReport) *handlerHandleCall {
	_c.Call = _c.Return(a)
	return _c
}

func (_c *handlerHandleCall) ReturnsFn(fn func(string, time.Time, json.RawMessage) *platform.CommandReport) *handlerHandleCall {
	_c.Call = _c.Return(fn)
	return _c
}

func (_c *handlerHandleCall) TypedRun(fn func(string, time.Time, json.RawMessage)) *handlerHandleCall {
	_c.Call = _c.Call.Run(func(args mock.Arguments) {
		_id := args.String(0)
		_requestedAt, _ := args.Get(1).(time.Time)
		_data, _ := args.Get(2).(json.RawMessage)
		fn(_id, _requestedAt, _data)
	})
	return _c
}

func (_c *handlerHandleCall) OnHandle(id string, requestedAt time.Time, data json.RawMessage) *handlerHandleCall {
	return _c.Parent.OnHandle(id, requestedAt, data)
}

func (_c *handlerHandleCall) OnHandleRaw(id interface{}, requestedAt interface{}, data interface{}) *handlerHandleCall {
	return _c.Parent.OnHandleRaw(id, requestedAt, data)
}
