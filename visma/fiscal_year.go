package visma

import (
	"fmt"
	"izettle-daily-reports/util"
	"time"
)

type FiscalYear struct {
	ID                    string    `json:"Id"`
	StartDate             util.Date `json:"StartDate"`
	EndDate               util.Date `json:"EndDate"`
	IsLockedForAccounting bool      `json:"IsLockedForAccounting"`
	BookkeepingMethod     int       `json:"BookkeepingMethod"`
}

func (c *Client) FiscalYears() ([]FiscalYear, error) {
	resource := "fiscalyears"
	resp := &struct {
		Meta Meta
		Data []FiscalYear
	}{}
	err := c.GetRequest(resource, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CurrentFiscalYear() (*FiscalYear, error) {
	years, err := c.FiscalYears()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, year := range years {
		if now.After(year.StartDate.Time()) && now.Before(year.EndDate.Time()) {
			return &year, nil
		}
	}
	return nil, fmt.Errorf("failed to get current year")
}
