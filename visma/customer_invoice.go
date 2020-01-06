package visma

import (
	"izettle-daily-reports/util"
	"time"
)

func (c *Client) NewCustomerInvoice(voucher CustomerInvoice) (*CustomerInvoice, error) {
	resource := "customerinvoices"
	resp := &CustomerInvoice{}
	err := c.PostRequest(resource, voucher, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type CustomerInvoice struct {
	ID                                       string    `json:"Id"`
	EuThirdParty                             bool      `json:"EuThirdParty"`
	IsCreditInvoice                          bool      `json:"IsCreditInvoice"`
	CurrencyCode                             string    `json:"CurrencyCode"`
	CurrencyRate                             int       `json:"CurrencyRate"`
	CreatedByUserID                          string    `json:"CreatedByUserId"`
	TotalAmount                              int       `json:"TotalAmount"`
	TotalVatAmount                           int       `json:"TotalVatAmount"`
	TotalRoundings                           int       `json:"TotalRoundings"`
	TotalAmountInvoiceCurrency               int       `json:"TotalAmountInvoiceCurrency"`
	TotalVatAmountInvoiceCurrency            int       `json:"TotalVatAmountInvoiceCurrency"`
	SetOffAmountInvoiceCurrency              int       `json:"SetOffAmountInvoiceCurrency"`
	CustomerID                               string    `json:"CustomerId"`
	Rows                                     []Rows    `json:"Rows"`
	InvoiceDate                              util.Date `json:"InvoiceDate"`
	DueDate                                  util.Date `json:"DueDate"`
	DeliveryDate                             util.Date `json:"DeliveryDate"`
	RotReducedInvoicingType                  int       `json:"RotReducedInvoicingType"`
	RotReducedInvoicingAmount                int       `json:"RotReducedInvoicingAmount"`
	RotReducedInvoicingPercent               int       `json:"RotReducedInvoicingPercent"`
	RotReducedInvoicingPropertyName          string    `json:"RotReducedInvoicingPropertyName"`
	RotReducedInvoicingOrgNumber             string    `json:"RotReducedInvoicingOrgNumber"`
	Persons                                  []Persons `json:"Persons"`
	RotReducedInvoicingAutomaticDistribution bool      `json:"RotReducedInvoicingAutomaticDistribution"`
	ElectronicReference                      string    `json:"ElectronicReference"`
	ElectronicAddress                        string    `json:"ElectronicAddress"`
	EdiServiceDelivererID                    string    `json:"EdiServiceDelivererId"`
	OurReference                             string    `json:"OurReference"`
	YourReference                            string    `json:"YourReference"`
	BuyersOrderReference                     string    `json:"BuyersOrderReference"`
	InvoiceCustomerName                      string    `json:"InvoiceCustomerName"`
	InvoiceAddress1                          string    `json:"InvoiceAddress1"`
	InvoiceAddress2                          string    `json:"InvoiceAddress2"`
	InvoicePostalCode                        string    `json:"InvoicePostalCode"`
	InvoiceCity                              string    `json:"InvoiceCity"`
	InvoiceCountryCode                       string    `json:"InvoiceCountryCode"`
	DeliveryCustomerName                     string    `json:"DeliveryCustomerName"`
	DeliveryAddress1                         string    `json:"DeliveryAddress1"`
	DeliveryAddress2                         string    `json:"DeliveryAddress2"`
	DeliveryPostalCode                       string    `json:"DeliveryPostalCode"`
	DeliveryCity                             string    `json:"DeliveryCity"`
	DeliveryCountryCode                      string    `json:"DeliveryCountryCode"`
	DeliveryMethodName                       string    `json:"DeliveryMethodName"`
	DeliveryTermName                         string    `json:"DeliveryTermName"`
	DeliveryMethodCode                       string    `json:"DeliveryMethodCode"`
	DeliveryTermCode                         string    `json:"DeliveryTermCode"`
	CustomerIsPrivatePerson                  bool      `json:"CustomerIsPrivatePerson"`
	TermsOfPaymentID                         string    `json:"TermsOfPaymentId"`
	CustomerEmail                            string    `json:"CustomerEmail"`
	InvoiceNumber                            int       `json:"InvoiceNumber"`
	CustomerNumber                           string    `json:"CustomerNumber"`
	PaymentReferenceNumber                   string    `json:"PaymentReferenceNumber"`
	RotPropertyType                          int       `json:"RotPropertyType"`
	SalesDocumentAttachments                 []string  `json:"SalesDocumentAttachments"`
	HasAutoInvoiceError                      bool      `json:"HasAutoInvoiceError"`
	IsNotDelivered                           bool      `json:"IsNotDelivered"`
	ReverseChargeOnConstructionServices      bool      `json:"ReverseChargeOnConstructionServices"`
	WorkHouseOtherCosts                      int       `json:"WorkHouseOtherCosts"`
	RemainingAmount                          int       `json:"RemainingAmount"`
	RemainingAmountInvoiceCurrency           int       `json:"RemainingAmountInvoiceCurrency"`
	ReferringInvoiceID                       string    `json:"ReferringInvoiceId"`
	CreatedFromOrderID                       string    `json:"CreatedFromOrderId"`
	CreatedFromDraftID                       string    `json:"CreatedFromDraftId"`
	VoucherNumber                            string    `json:"VoucherNumber"`
	VoucherID                                string    `json:"VoucherId"`
	CreatedUtc                               time.Time `json:"CreatedUtc"`
	ModifiedUtc                              time.Time `json:"ModifiedUtc"`
	ReversedConstructionVatInvoicing         bool      `json:"ReversedConstructionVatInvoicing"`
	IncludesVat                              bool      `json:"IncludesVat"`
	SendType                                 int       `json:"SendType"`
	PaymentReminderIssued                    bool      `json:"PaymentReminderIssued"`
}
type Rows struct {
	ID                      string `json:"Id"`
	ArticleNumber           string `json:"ArticleNumber"`
	ArticleID               string `json:"ArticleId"`
	AmountNoVat             int    `json:"AmountNoVat"`
	PercentVat              int    `json:"PercentVat"`
	LineNumber              int    `json:"LineNumber"`
	IsTextRow               bool   `json:"IsTextRow"`
	Text                    string `json:"Text"`
	UnitPrice               int    `json:"UnitPrice"`
	UnitAbbreviation        string `json:"UnitAbbreviation"`
	UnitAbbreviationEnglish string `json:"UnitAbbreviationEnglish"`
	DiscountPercentage      int    `json:"DiscountPercentage"`
	Quantity                int    `json:"Quantity"`
	IsWorkCost              bool   `json:"IsWorkCost"`
	IsVatFree               bool   `json:"IsVatFree"`
	CostCenterItemID1       string `json:"CostCenterItemId1"`
	CostCenterItemID2       string `json:"CostCenterItemId2"`
	CostCenterItemID3       string `json:"CostCenterItemId3"`
	UnitID                  string `json:"UnitId"`
	ProjectID               string `json:"ProjectId"`
	WorkCostType            int    `json:"WorkCostType"`
	WorkHours               int    `json:"WorkHours"`
	MaterialCosts           int    `json:"MaterialCosts"`
}
type Persons struct {
	Ssn    string `json:"Ssn"`
	Amount int    `json:"Amount"`
}
