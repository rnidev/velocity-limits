package account

import (
	"errors"
	"reflect"
	"testing"
	"time"

	cache "github.com/patrickmn/go-cache"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type FundTestSuite struct {
	suite.Suite
	request      string
	err          error
	expectedErr  error
	resp         interface{}
	expectedResp interface{}
}

type MockCustomerAccount struct {
	mock.Mock
}

func (mock *MockCustomerAccount) LoadFund(fund Fund, c *cache.Cache) (bool, error) {
	args := mock.Called()
	return args.Bool(0), args.Error(1)
}
func (mock *MockCustomerAccount) checkIfLoadExists(request string) error {
	args := mock.Called(request)
	return args.Error(0)
}
func (mock *MockCustomerAccount) checkDailyLimit(fund *Fund) error {
	args := mock.Called(fund)
	return args.Error(0)
}
func (mock *MockCustomerAccount) checkWeeklyLimit(fund *Fund) error {
	args := mock.Called(fund)
	return args.Error(0)
}

func TestFundLoads(t *testing.T) {
	suite.Run(t, new(FundTestSuite))
}

func (s *FundTestSuite) Reset() {
	s.request = ""
	s.err = nil
	s.expectedErr = nil
	s.resp = nil
	s.expectedResp = nil
}

func (s *FundTestSuite) TestInvalidJsonString() {
	s.Reset()
	s.request = `{:"324","load_amount":"$4810.91","time":"2000-02-05T17:05:16Z"}`
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}

func (s *FundTestSuite) TestMissingParamInRequest() {
	s.Reset()
	s.request = `{"id":"16710","customer_id":"783","load_amount":"$750.87","time":""}`
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}
func (s *FundTestSuite) TestInvalidDataForLoadAmount() {
	s.Reset()
	s.request = `{"id":"29360","customer_id":"18","load_amount":"$$5745.70","time":"2000-02-04T12:27:00Z"}`
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}

func (s *FundTestSuite) TestInvalidDataTime() {
	s.Reset()
	s.request = `{"id":"29360","customer_id":"18","load_amount":"$5745.70","time":"20000-02-04T12:27:00Z"}`
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}

func (s *FundTestSuite) TestSuccessfulLoad() {
	s.Reset()
	s.request = `{"id":"29360","customer_id":"18","load_amount":"$4745.70","time":"2000-02-04T12:27:00Z"}`
	mock := new(MockCustomerAccount)
	mock.On("LoadFund").Return(false, nil)
	c := cache.New(5*time.Minute, 10*time.Minute)
	v := validator.New()
	handler := FundHandler{
		service:  mock,
		validate: v,
		cache:    c,
	}
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{
		ID:         "29360",
		CustomerID: "18",
		Accepted:   true,
	}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}
func (s *FundTestSuite) TestFailToLoadExceedLimit() {
	s.Reset()
	s.request = `{"id":"29360","customer_id":"18","load_amount":"$5745.70","time":"2000-02-04T12:27:00Z"}`
	mock := new(MockCustomerAccount)
	mock.On("LoadFund").Return(false, errors.New("some error"))
	c := cache.New(5*time.Minute, 10*time.Minute)
	v := validator.New()
	handler := FundHandler{
		service:  mock,
		validate: v,
		cache:    c,
	}
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{
		ID:         "29360",
		CustomerID: "18",
		Accepted:   false,
	}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}

func (s *FundTestSuite) TestFailToLoadLoadIDExists() {
	s.Reset()
	s.request = `{"id":"29360","customer_id":"18","load_amount":"$5745.70","time":"2000-02-04T12:27:00Z"}`
	mock := new(MockCustomerAccount)
	mock.On("LoadFund").Return(true, errors.New("LoadID exists"))
	c := cache.New(5*time.Minute, 10*time.Minute)
	v := validator.New()
	handler := FundHandler{
		service:  mock,
		validate: v,
		cache:    c,
	}
	s.resp = handler.Run(s.request)
	s.expectedResp = FundResponse{}
	if s.expectedErr != nil && s.err == nil {
		s.T().Error("error was expected, but no error return")
	}
	if s.err != nil && !reflect.DeepEqual(s.err, s.expectedErr) {
		s.T().Errorf("error expected was %s, but error returned was %s.", s.expectedErr, s.err)
	}
	if !reflect.DeepEqual(s.resp, s.expectedResp) {
		s.T().Errorf("response: %v, expected response: %v", s.resp, s.expectedResp)
	}
}
