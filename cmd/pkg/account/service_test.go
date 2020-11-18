package account

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	validator "gopkg.in/go-playground/validator.v9"

	cache "github.com/patrickmn/go-cache"

	"github.com/stretchr/testify/suite"
)

type CustomerAccountTestSuite struct {
	suite.Suite
	request      Fund
	err          error
	expectedErr  error
	resp         interface{}
	expectedResp interface{}
}

func TestCustomerAccount(t *testing.T) {
	suite.Run(t, new(CustomerAccountTestSuite))
}

func (s *CustomerAccountTestSuite) Reset() {
	s.request = Fund{}
	s.err = nil
	s.expectedErr = nil
	s.resp = nil
	s.expectedResp = nil
}

func (s *CustomerAccountTestSuite) TestLoadIDExists() {
	s.Reset()
	s.request = Fund{
		ID:         "29360",
		CustomerID: "18",
		LoadAmount: 4000.00,
		Time:       time.Now(),
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	transactions := make(map[string][]Fund)
	transactions[time.Now().Format("01/02/2019")] = []Fund{
		Fund{
			ID:         "29360",
			CustomerID: "18",
			LoadAmount: 4000.00,
			Time:       time.Now(),
		},
	}
	data := CustomerAccount{
		ID:           "18",
		LoadIDs:      []string{"29360"},
		Transactions: transactions,
	}
	c.Set("18", data, cache.DefaultExpiration)
	s.resp, s.err = handler.service.LoadFund(s.request, c)
	s.expectedResp = true
	s.expectedErr = fmt.Errorf("loadID: %s exists", s.request.ID)
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
func (s *CustomerAccountTestSuite) TestExceedDailyAmountLimit() {
	s.Reset()
	s.request = Fund{
		ID:         "29360",
		CustomerID: "18",
		LoadAmount: 5000.01,
		Time:       time.Now(),
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	s.resp, s.err = handler.service.LoadFund(s.request, c)
	s.expectedResp = false
	s.expectedErr = fmt.Errorf(
		"accountID: %s exceed daily fund limit when process loadID: %s",
		s.request.CustomerID, s.request.ID,
	)
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
func (s *CustomerAccountTestSuite) TestExceedDailyLoadsLimit() {
	s.Reset()
	s.request = Fund{
		ID:         "29370",
		CustomerID: "18",
		LoadAmount: 1.00,
		Time:       time.Now(),
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	transactions := make(map[string][]Fund)
	transactions[time.Now().Format("01/02/2019")] = []Fund{
		Fund{
			ID:         "29360",
			CustomerID: "18",
			LoadAmount: 1.00,
			Time:       time.Now(),
		},
		Fund{
			ID:         "29361",
			CustomerID: "18",
			LoadAmount: 1.00,
			Time:       time.Now(),
		},
		Fund{
			ID:         "29362",
			CustomerID: "18",
			LoadAmount: 1.00,
			Time:       time.Now(),
		},
	}
	data := CustomerAccount{
		ID:           "18",
		LoadIDs:      []string{"29360", "29361", "29362"},
		Transactions: transactions,
	}
	c.Set("18", data, cache.DefaultExpiration)
	s.resp, s.err = handler.service.LoadFund(s.request, c)
	s.expectedResp = false
	s.expectedErr = fmt.Errorf(
		"accountID: %s exceed daily number of loads limit when process loadID: %s",
		s.request.CustomerID, s.request.ID,
	)
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

func (s *CustomerAccountTestSuite) TestExceedDailyAmountLimitWithMultiple() {
	s.Reset()
	s.request = Fund{
		ID:         "29370",
		CustomerID: "18",
		LoadAmount: 2000.01,
		Time:       time.Now(),
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	transactions := make(map[string][]Fund)
	transactions[time.Now().Format("01/02/2019")] = []Fund{
		Fund{
			ID:         "29360",
			CustomerID: "18",
			LoadAmount: 1000.00,
			Time:       time.Now(),
		},
		Fund{
			ID:         "29361",
			CustomerID: "18",
			LoadAmount: 2000.00,
			Time:       time.Now(),
		},
	}
	data := CustomerAccount{
		ID:           "18",
		LoadIDs:      []string{"29360", "29361", "29362"},
		Transactions: transactions,
	}
	c.Set("18", data, cache.DefaultExpiration)
	s.resp, s.err = handler.service.LoadFund(s.request, c)
	s.expectedResp = false
	s.expectedErr = fmt.Errorf(
		"accountID: %s exceed daily fund limit when process loadID: %s",
		s.request.CustomerID, s.request.ID,
	)
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

func (s *CustomerAccountTestSuite) TestExceedWeeklyAmountLimit() {
	s.Reset()
	date := time.Date(2020, 11, 18, 0, 0, 0, 0, time.UTC)
	s.request = Fund{
		ID:         "29370",
		CustomerID: "18",
		LoadAmount: 0.01,
		Time:       date,
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	var service Service
	service = CustomerAccount{}
	v := validator.New()
	handler := NewHandler(service, v, c)
	transactions := make(map[string][]Fund)
	transactions[date.AddDate(0, 0, -1).Format("01/02/2019")] = []Fund{
		Fund{
			ID:         "29360",
			CustomerID: "18",
			LoadAmount: 10000.00,
			Time:       time.Now(),
		},
	}
	transactions[date.AddDate(0, 0, -2).Format("01/02/2019")] = []Fund{
		Fund{
			ID:         "29361",
			CustomerID: "18",
			LoadAmount: 10000.00,
			Time:       time.Now(),
		},
	}
	data := CustomerAccount{
		ID:           "18",
		LoadIDs:      []string{"29360", "29361"},
		Transactions: transactions,
	}
	c.Set("18", data, cache.DefaultExpiration)
	s.resp, s.err = handler.service.LoadFund(s.request, c)
	s.expectedResp = false
	s.expectedErr = fmt.Errorf(
		"accountID: %s exceed weekly fund limit when process loadID: %s",
		s.request.CustomerID, s.request.ID,
	)
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
