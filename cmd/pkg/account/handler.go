package account

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	validator "gopkg.in/go-playground/validator.v9"
)

type fundRequest struct {
	ID         string `json:"id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
	LoadAmount string `json:"load_amount" validate:"required"`
	Time       string `json:"time" validate:"required"`
}

type FundResponse struct {
	ID         string `json:"id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
	Accepted   bool   `json:"accepted" validate:"required"`
}

//FundHandler contains validator to validate fund request
type FundHandler struct {
	validate *validator.Validate
	service  Service
	cache    *cache.Cache
}

//NewHandler will create a new FundHandler for requested fund transaction
func NewHandler(s Service, v *validator.Validate, c *cache.Cache) FundHandler {
	return FundHandler{service: s, validate: v, cache: c}
}

//Run will take json string as request, validate, and process the request
func (h *FundHandler) Run(req string) FundResponse {
	var err error
	input := fundRequest{}
	if err = json.Unmarshal([]byte(req), &input); err != nil {
		log.Print(err)
		return FundResponse{}
	}
	if err = h.validate.Struct(input); err != nil {
		log.Print(err)
		return FundResponse{}
	}
	amount, err := strconv.ParseFloat(strings.TrimPrefix(input.LoadAmount, "$"), 64)
	if err != nil {
		log.Print(err)
		return FundResponse{}
	}
	timestamp, err := time.Parse(time.RFC3339, input.Time)
	if err != nil {
		log.Print(err)
		return FundResponse{}
	}
	fund := Fund{
		ID:         input.ID,
		CustomerID: input.CustomerID,
		LoadAmount: amount,
		Time:       timestamp,
	}
	exists, err := h.service.LoadFund(fund, h.cache)
	//return empty if loadID exists
	if exists {
		log.Print(err)
		return FundResponse{}
	}
	if err != nil {
		return FundResponse{
			ID:         fund.ID,
			CustomerID: fund.CustomerID,
			Accepted:   false,
		}
	}
	return FundResponse{
		ID:         fund.ID,
		CustomerID: fund.CustomerID,
		Accepted:   true,
	}
}
