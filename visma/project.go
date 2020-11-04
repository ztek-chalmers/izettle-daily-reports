package visma

import (
	"fmt"
	"izettle-daily-reports/util"
	"time"
)

type Project struct {
	ID           string    `json:"Id"`
	Number       string    `json:"Number"`
	Name         string    `json:"Name"`
	StartDate    util.Date `json:"StartDate"`
	EndDate      util.Date `json:"EndDate"`
	CustomerID   string    `json:"CustomerId"`
	CustomerName string    `json:"CustomerName"`
	Notes        string    `json:"Notes"`
	Status       int       `json:"Status"`
	ModifiedUtc  time.Time `json:"ModifiedUtc"`
}

func (c *Client) Projects(id ...string) ([]Project, error) {
	resource := "projects"
	if len(id) > 1 {
		return nil, fmt.Errorf("projects can only take one optional id")
	} else if len(id) == 1 {
		resource = resource + "/" + id[0]
	}
	page := 1
	pageSize := 1000
	projects := []Project{}
	for {
		resp := struct {
			Meta Meta
			Data []Project
		}{}
		err := c.GetRequestPage(resource, page, pageSize, &resp)
		if err != nil {
			return nil, err
		}
		for _, p := range resp.Data {
			projects = append(projects, p)
		}
		if resp.Meta.CurrentPage >= resp.Meta.TotalNumberOfPages {
			break
		}
		page = resp.Meta.CurrentPage + 1
		pageSize = resp.Meta.PageSize
	}
	return projects, nil
}
