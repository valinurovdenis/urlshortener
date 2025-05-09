// Code generated by mockery v2.46.3. DO NOT EDIT.

// Package mocks contains mocks for testing.
package mocks

import mock "github.com/stretchr/testify/mock"

// ShortCutGenerator is an autogenerated mock type for the ShortCutGenerator type
type ShortCutGenerator struct {
	mock.Mock
}

// Generate provides a mock function with given fields:
func (_m *ShortCutGenerator) Generate() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Generate")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewShortCutGenerator creates a new instance of ShortCutGenerator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShortCutGenerator(t interface {
	mock.TestingT
	Cleanup(func())
}) *ShortCutGenerator {
	mock := &ShortCutGenerator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
