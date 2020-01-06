package izettle

type Product struct {
	UUID              string    `json:"uuid"`
	Categories        []string  `json:"categories"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	ImageLookupKeys   []string  `json:"imageLookupKeys"`
	Variants          []Variant `json:"variants"`
	ExternalReference string    `json:"externalReference"`
	Etag              string    `json:"etag"`
	Updated           string    `json:"updated"`
	UpdatedBy         string    `json:"updatedBy"`
	Created           string    `json:"created"`
	UnitName          string    `json:"unitName"`
	VatPercentage     string    `json:"vatPercentage"`
	TaxCode           string    `json:"taxCode"`
	Category          Category  `json:"category"`
}
type Price struct {
	Amount     int    `json:"amount"`
	CurrencyID string `json:"currencyId"`
}
type CostPrice struct {
	Amount     int    `json:"amount"`
	CurrencyID string `json:"currencyId"`
}
type Variant struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Sku         string    `json:"sku"`
	Barcode     string    `json:"barcode"`
	Price       Price     `json:"price"`
	CostPrice   CostPrice `json:"costPrice"`
}
type Category struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
type Source struct {
	Name     string `json:"name"`
	External bool   `json:"external"`
}

func (c *Client) Products() ([]Product, error) {
	resource := "/organizations/self/library"
	resp := struct {
		Products []Product
	}{}
	err := c.GetRequest(productURL, resource, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Products, nil
}
