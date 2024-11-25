// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	utils "github.com/valinurovdenis/urlshortener/internal/app/utils"
)

// URLStorage is an autogenerated mock type for the URLStorage type
type URLStorage struct {
	mock.Mock
}

// Clear provides a mock function with given fields:
func (_m *URLStorage) Clear() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Clear")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLongURLWithContext provides a mock function with given fields: _a0, shortURL
func (_m *URLStorage) GetLongURLWithContext(_a0 context.Context, shortURL string) (string, error) {
	ret := _m.Called(_a0, shortURL)

	if len(ret) == 0 {
		panic("no return value specified for GetLongURLWithContext")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(_a0, shortURL)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(_a0, shortURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, shortURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetShortURLWithContext provides a mock function with given fields: _a0, longURL
func (_m *URLStorage) GetShortURLWithContext(_a0 context.Context, longURL string) (string, error) {
	ret := _m.Called(_a0, longURL)

	if len(ret) == 0 {
		panic("no return value specified for GetShortURLWithContext")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(_a0, longURL)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(_a0, longURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, longURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserURLs provides a mock function with given fields: _a0, userID
func (_m *URLStorage) GetUserURLs(_a0 context.Context, userID string) ([]utils.URLPair, error) {
	ret := _m.Called(_a0, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetUserURLs")
	}

	var r0 []utils.URLPair
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]utils.URLPair, error)); ok {
		return rf(_a0, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []utils.URLPair); ok {
		r0 = rf(_a0, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]utils.URLPair)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields:
func (_m *URLStorage) Ping() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreManyWithContext provides a mock function with given fields: _a0, long2ShortUrls, userID
func (_m *URLStorage) StoreManyWithContext(_a0 context.Context, long2ShortUrls []utils.URLPair, userID string) ([]error, error) {
	ret := _m.Called(_a0, long2ShortUrls, userID)

	if len(ret) == 0 {
		panic("no return value specified for StoreManyWithContext")
	}

	var r0 []error
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []utils.URLPair, string) ([]error, error)); ok {
		return rf(_a0, long2ShortUrls, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []utils.URLPair, string) []error); ok {
		r0 = rf(_a0, long2ShortUrls, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]error)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []utils.URLPair, string) error); ok {
		r1 = rf(_a0, long2ShortUrls, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StoreWithContext provides a mock function with given fields: _a0, longURL, shortURL, userID
func (_m *URLStorage) StoreWithContext(_a0 context.Context, longURL string, shortURL string, userID string) error {
	ret := _m.Called(_a0, longURL, shortURL, userID)

	if len(ret) == 0 {
		panic("no return value specified for StoreWithContext")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(_a0, longURL, shortURL, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewURLStorage creates a new instance of URLStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewURLStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *URLStorage {
	mock := &URLStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
