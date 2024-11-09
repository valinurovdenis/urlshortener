// Code generated by mockery v2.46.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

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

// GetLongURL provides a mock function with given fields: shortURL
func (_m *URLStorage) GetLongURL(shortURL string) (string, error) {
	ret := _m.Called(shortURL)

	if len(ret) == 0 {
		panic("no return value specified for GetLongURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(shortURL)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(shortURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(shortURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetShortURL provides a mock function with given fields: longURL
func (_m *URLStorage) GetShortURL(longURL string) (string, error) {
	ret := _m.Called(longURL)

	if len(ret) == 0 {
		panic("no return value specified for GetShortURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(longURL)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(longURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(longURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: longURL, shortURL
func (_m *URLStorage) Store(longURL string, shortURL string) error {
	ret := _m.Called(longURL, shortURL)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(longURL, shortURL)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreMany provides a mock function with given fields: long2ShortUrls
func (_m *URLStorage) StoreMany(long2ShortUrls map[string]string) error {
	ret := _m.Called(long2ShortUrls)

	if len(ret) == 0 {
		panic("no return value specified for StoreMany")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]string) error); ok {
		r0 = rf(long2ShortUrls)
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
