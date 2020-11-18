package account

import (
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	DailyNumberOfLoadsLimit = 3
	DailyFundLimit          = 5000.00
	WeeklyFundLimit         = 20000.00
)

type Service interface {
	LoadFund(Fund, *cache.Cache) (bool, error)
	checkIfLoadExists(string) error
	checkDailyLimit(*Fund) error
	checkWeeklyLimit(*Fund) error
}
type CustomerAccount struct {
	ID           string
	LoadIDs      []string
	Transactions map[string][]Fund
}
type Fund struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	LoadAmount float64   `json:"load_amount"`
	Time       time.Time `json:"time"`
}

//LoadFund will validate dupe transaction, and check account velocity limits before load fund into account
func (a CustomerAccount) LoadFund(fund Fund, c *cache.Cache) (bool, error) {
	var err error
	//Try to find customer account in cache, if not found create a new account
	if x, found := c.Get(fund.CustomerID); found {
		a = x.(CustomerAccount)
	} else {
		a = CustomerAccount{
			ID: fund.CustomerID,
		}
	}
	//Check against customer account to see if loadID alreay exits. If yes, set skip to true
	if err = a.checkIfLoadExists(fund.ID); err != nil {
		return true, err
	}
	//Log LoadID even if the load doesn't pass validation
	a.LoadIDs = append(a.LoadIDs, fund.ID)
	c.Set(a.ID, a, cache.DefaultExpiration)
	if err = a.checkDailyLimit(&fund); err != nil {
		return false, err
	}
	if err = a.checkWeeklyLimit(&fund); err != nil {
		return false, err
	}
	//use date as key to group loads together as transaction history in account
	date := fund.Time.Format("01/02/2019")
	if len(a.Transactions) == 0 {
		transactions := make(map[string][]Fund)
		transactions[date] = []Fund{fund}
		a.Transactions = transactions
	} else {
		a.Transactions[date] = append(a.Transactions[date], fund)
	}
	c.Set(a.ID, a, cache.DefaultExpiration)
	return false, nil
}
func (a CustomerAccount) checkIfLoadExists(loadID string) error {
	if find(a.LoadIDs, loadID) {
		return fmt.Errorf("loadID: %s exists", loadID)
	}
	return nil
}
func (a CustomerAccount) checkDailyLimit(fund *Fund) error {
	//There are two limits for daily fund: number of loads per day, and maximum amount per day.
	date := fund.Time.Format("01/02/2019")
	if len(a.Transactions[date]) >= DailyNumberOfLoadsLimit {
		return fmt.Errorf("accountID: %s exceed daily number of loads limit when process loadID: %s", a.ID, fund.ID)
	}
	var totalAmount float64
	if len(a.Transactions[date]) > 0 {
		for _, load := range a.Transactions[date] {
			totalAmount += load.LoadAmount
		}
	}
	if (totalAmount + fund.LoadAmount) > DailyFundLimit {
		return fmt.Errorf("accountID: %s exceed daily fund limit when process loadID: %s", a.ID, fund.ID)
	}
	return nil
}
func (a CustomerAccount) checkWeeklyLimit(fund *Fund) error {
	//Weeks start on Monday, and time.Monday is 1
	//Find all loads between Monday and current fund request date, then check if excced amount limit per week
	var totalAmount float64
	for i := 0; i < int(fund.Time.Weekday()); i++ {
		date := fund.Time.AddDate(0, 0, -i).Format("01/02/2019")
		if len(a.Transactions[date]) > 0 {
			for _, load := range a.Transactions[date] {
				totalAmount += load.LoadAmount
			}
		}
	}
	if (totalAmount + fund.LoadAmount) > WeeklyFundLimit {
		return fmt.Errorf("accountID: %s exceed weekly fund limit when process loadID: %s", a.ID, fund.ID)
	}
	return nil
}
func find(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}
