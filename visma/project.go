package visma

import (
	"fmt"
	"time"
)

type Project struct {
	ID           string    `json:"Id"`
	Number       string    `json:"Number"`
	Name         string    `json:"Name"`
	StartDate    Date      `json:"StartDate"`
	EndDate      Date      `json:"EndDate"`
	CustomerID   string    `json:"CustomerId"`
	CustomerName string    `json:"CustomerName"`
	Notes        string    `json:"Notes"`
	Status       int       `json:"Status"`
	ModifiedUtc  time.Time `json:"ModifiedUtc"`
}

func (c *Client) Projects(id ...string) ([]Project, error) {
	resource := "projects"
	resp := &struct {
		Meta Meta
		Data []Project
	}{}
	if len(id) > 1 {
		return nil, fmt.Errorf("projects can only take one optional id")
	} else if len(id) == 1 {
		resource = resource + "/" + id[0]
	}
	err := c.GetRequest(resource, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
