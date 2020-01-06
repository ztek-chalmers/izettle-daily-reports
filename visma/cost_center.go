package visma

import "time"

type CostCenterItem struct {
	CostCenterID string    `json:"CostCenterId"`
	ID           string    `json:"Id"`
	Name         string    `json:"Name"`
	ShortName    string    `json:"ShortName"`
	IsActive     bool      `json:"IsActive"`
	CreatedUtc   time.Time `json:"CreatedUtc"`
}
type CostCenter struct {
	Name     string           `json:"Name"`
	Number   int              `json:"Number"`
	IsActive bool             `json:"IsActive"`
	Items    []CostCenterItem `json:"Items"`
	ID       string           `json:"Id"`
}

func (c *Client) CostCenters() ([]CostCenter, error) {
	resource := "costcenters"
	resp := &struct {
		Meta Meta
		Data []CostCenter
	}{}
	err := c.GetRequest(resource, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
