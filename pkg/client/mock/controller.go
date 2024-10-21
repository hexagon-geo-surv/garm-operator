// SPDX-License-Identifier: MIT
// Code generated by MockGen. DO NOT EDIT.
// Source: ../controller.go
//
// Generated by this command:
//
//	mockgen -package mock -destination=controller.go -source=../controller.go GarmServerConfig
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	controller "github.com/cloudbase/garm/client/controller"
	controller_info "github.com/cloudbase/garm/client/controller_info"
	gomock "go.uber.org/mock/gomock"
)

// MockControllerClient is a mock of ControllerClient interface.
type MockControllerClient struct {
	ctrl     *gomock.Controller
	recorder *MockControllerClientMockRecorder
}

// MockControllerClientMockRecorder is the mock recorder for MockControllerClient.
type MockControllerClientMockRecorder struct {
	mock *MockControllerClient
}

// NewMockControllerClient creates a new mock instance.
func NewMockControllerClient(ctrl *gomock.Controller) *MockControllerClient {
	mock := &MockControllerClient{ctrl: ctrl}
	mock.recorder = &MockControllerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockControllerClient) EXPECT() *MockControllerClientMockRecorder {
	return m.recorder
}

// GetControllerInfo mocks base method.
func (m *MockControllerClient) GetControllerInfo() (*controller_info.ControllerInfoOK, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetControllerInfo")
	ret0, _ := ret[0].(*controller_info.ControllerInfoOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetControllerInfo indicates an expected call of GetControllerInfo.
func (mr *MockControllerClientMockRecorder) GetControllerInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetControllerInfo", reflect.TypeOf((*MockControllerClient)(nil).GetControllerInfo))
}

// UpdateController mocks base method.
func (m *MockControllerClient) UpdateController(params *controller.UpdateControllerParams) (*controller.UpdateControllerOK, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateController", params)
	ret0, _ := ret[0].(*controller.UpdateControllerOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateController indicates an expected call of UpdateController.
func (mr *MockControllerClientMockRecorder) UpdateController(params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateController", reflect.TypeOf((*MockControllerClient)(nil).UpdateController), params)
}