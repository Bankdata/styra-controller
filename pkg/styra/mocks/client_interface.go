// Code generated by mockery v2.23.2. DO NOT EDIT.

package mocks

import (
	context "context"

	styra "github.com/bankdata/styra-controller/pkg/styra"
	mock "github.com/stretchr/testify/mock"
)

// ClientInterface is an autogenerated mock type for the ClientInterface type
type ClientInterface struct {
	mock.Mock
}

// CreateInvitation provides a mock function with given fields: ctx, email, name
func (_m *ClientInterface) CreateInvitation(ctx context.Context, email bool, name string) (*styra.CreateInvitationResponse, error) {
	ret := _m.Called(ctx, email, name)

	var r0 *styra.CreateInvitationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, bool, string) (*styra.CreateInvitationResponse, error)); ok {
		return rf(ctx, email, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, bool, string) *styra.CreateInvitationResponse); ok {
		r0 = rf(ctx, email, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.CreateInvitationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, bool, string) error); ok {
		r1 = rf(ctx, email, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateRoleBinding provides a mock function with given fields: ctx, request
func (_m *ClientInterface) CreateRoleBinding(ctx context.Context, request *styra.CreateRoleBindingRequest) (*styra.CreateRoleBindingResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *styra.CreateRoleBindingResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *styra.CreateRoleBindingRequest) (*styra.CreateRoleBindingResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *styra.CreateRoleBindingRequest) *styra.CreateRoleBindingResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.CreateRoleBindingResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *styra.CreateRoleBindingRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSystem provides a mock function with given fields: ctx, request
func (_m *ClientInterface) CreateSystem(ctx context.Context, request *styra.CreateSystemRequest) (*styra.CreateSystemResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *styra.CreateSystemResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *styra.CreateSystemRequest) (*styra.CreateSystemResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *styra.CreateSystemRequest) *styra.CreateSystemResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.CreateSystemResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *styra.CreateSystemRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateUpdateSecret provides a mock function with given fields: ctx, secretID, request
func (_m *ClientInterface) CreateUpdateSecret(ctx context.Context, secretID string, request *styra.CreateUpdateSecretsRequest) (*styra.CreateUpdateSecretResponse, error) {
	ret := _m.Called(ctx, secretID, request)

	var r0 *styra.CreateUpdateSecretResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.CreateUpdateSecretsRequest) (*styra.CreateUpdateSecretResponse, error)); ok {
		return rf(ctx, secretID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.CreateUpdateSecretsRequest) *styra.CreateUpdateSecretResponse); ok {
		r0 = rf(ctx, secretID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.CreateUpdateSecretResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *styra.CreateUpdateSecretsRequest) error); ok {
		r1 = rf(ctx, secretID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteDatasource provides a mock function with given fields: ctx, id
func (_m *ClientInterface) DeleteDatasource(ctx context.Context, id string) (*styra.DeleteDatasourceResponse, error) {
	ret := _m.Called(ctx, id)

	var r0 *styra.DeleteDatasourceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.DeleteDatasourceResponse, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.DeleteDatasourceResponse); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.DeleteDatasourceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteRoleBindingV2 provides a mock function with given fields: ctx, id
func (_m *ClientInterface) DeleteRoleBindingV2(ctx context.Context, id string) (*styra.DeleteRoleBindingV2Response, error) {
	ret := _m.Called(ctx, id)

	var r0 *styra.DeleteRoleBindingV2Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.DeleteRoleBindingV2Response, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.DeleteRoleBindingV2Response); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.DeleteRoleBindingV2Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSystem provides a mock function with given fields: ctx, id
func (_m *ClientInterface) DeleteSystem(ctx context.Context, id string) (*styra.DeleteSystemResponse, error) {
	ret := _m.Called(ctx, id)

	var r0 *styra.DeleteSystemResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.DeleteSystemResponse, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.DeleteSystemResponse); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.DeleteSystemResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatasource provides a mock function with given fields: ctx, id
func (_m *ClientInterface) GetDatasource(ctx context.Context, id string) (*styra.GetDatasourceResponse, error) {
	ret := _m.Called(ctx, id)

	var r0 *styra.GetDatasourceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.GetDatasourceResponse, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.GetDatasourceResponse); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.GetDatasourceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOPAConfig provides a mock function with given fields: ctx, systemID
func (_m *ClientInterface) GetOPAConfig(ctx context.Context, systemID string) (styra.OPAConfig, error) {
	ret := _m.Called(ctx, systemID)

	var r0 styra.OPAConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (styra.OPAConfig, error)); ok {
		return rf(ctx, systemID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) styra.OPAConfig); ok {
		r0 = rf(ctx, systemID)
	} else {
		r0 = ret.Get(0).(styra.OPAConfig)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, systemID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSystem provides a mock function with given fields: ctx, id
func (_m *ClientInterface) GetSystem(ctx context.Context, id string) (*styra.GetSystemResponse, error) {
	ret := _m.Called(ctx, id)

	var r0 *styra.GetSystemResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.GetSystemResponse, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.GetSystemResponse); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.GetSystemResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUser provides a mock function with given fields: ctx, name
func (_m *ClientInterface) GetUser(ctx context.Context, name string) (*styra.GetUserResponse, error) {
	ret := _m.Called(ctx, name)

	var r0 *styra.GetUserResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*styra.GetUserResponse, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *styra.GetUserResponse); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.GetUserResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListRoleBindingsV2 provides a mock function with given fields: ctx, params
func (_m *ClientInterface) ListRoleBindingsV2(ctx context.Context, params *styra.ListRoleBindingsV2Params) (*styra.ListRoleBindingsV2Response, error) {
	ret := _m.Called(ctx, params)

	var r0 *styra.ListRoleBindingsV2Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *styra.ListRoleBindingsV2Params) (*styra.ListRoleBindingsV2Response, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *styra.ListRoleBindingsV2Params) *styra.ListRoleBindingsV2Response); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.ListRoleBindingsV2Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *styra.ListRoleBindingsV2Params) error); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateRoleBindingSubjects provides a mock function with given fields: ctx, id, request
func (_m *ClientInterface) UpdateRoleBindingSubjects(ctx context.Context, id string, request *styra.UpdateRoleBindingSubjectsRequest) (*styra.UpdateRoleBindingSubjectsResponse, error) {
	ret := _m.Called(ctx, id, request)

	var r0 *styra.UpdateRoleBindingSubjectsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpdateRoleBindingSubjectsRequest) (*styra.UpdateRoleBindingSubjectsResponse, error)); ok {
		return rf(ctx, id, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpdateRoleBindingSubjectsRequest) *styra.UpdateRoleBindingSubjectsResponse); ok {
		r0 = rf(ctx, id, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.UpdateRoleBindingSubjectsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *styra.UpdateRoleBindingSubjectsRequest) error); ok {
		r1 = rf(ctx, id, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSystem provides a mock function with given fields: ctx, id, request
func (_m *ClientInterface) UpdateSystem(ctx context.Context, id string, request *styra.UpdateSystemRequest) (*styra.UpdateSystemResponse, error) {
	ret := _m.Called(ctx, id, request)

	var r0 *styra.UpdateSystemResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpdateSystemRequest) (*styra.UpdateSystemResponse, error)); ok {
		return rf(ctx, id, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpdateSystemRequest) *styra.UpdateSystemResponse); ok {
		r0 = rf(ctx, id, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.UpdateSystemResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *styra.UpdateSystemRequest) error); ok {
		r1 = rf(ctx, id, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertDatasource provides a mock function with given fields: ctx, id, request
func (_m *ClientInterface) UpsertDatasource(ctx context.Context, id string, request *styra.UpsertDatasourceRequest) (*styra.UpsertDatasourceResponse, error) {
	ret := _m.Called(ctx, id, request)

	var r0 *styra.UpsertDatasourceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpsertDatasourceRequest) (*styra.UpsertDatasourceResponse, error)); ok {
		return rf(ctx, id, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *styra.UpsertDatasourceRequest) *styra.UpsertDatasourceResponse); ok {
		r0 = rf(ctx, id, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.UpsertDatasourceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *styra.UpsertDatasourceRequest) error); ok {
		r1 = rf(ctx, id, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// VerifyGitConfiguration provides a mock function with given fields: ctx, request
func (_m *ClientInterface) VerifyGitConfiguration(ctx context.Context, request *styra.VerfiyGitConfigRequest) (*styra.VerfiyGitConfigResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *styra.VerfiyGitConfigResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *styra.VerfiyGitConfigRequest) (*styra.VerfiyGitConfigResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *styra.VerfiyGitConfigRequest) *styra.VerfiyGitConfigResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*styra.VerfiyGitConfigResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *styra.VerfiyGitConfigRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewClientInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewClientInterface creates a new instance of ClientInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClientInterface(t mockConstructorTestingTNewClientInterface) *ClientInterface {
	mock := &ClientInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
