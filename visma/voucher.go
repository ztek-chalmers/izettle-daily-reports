package visma

import (
	"fmt"
	"izettle-daily-reports/util"
	"time"
)

const ManualVoucher = 2
const BankAccountTransferDeposit = 5
const BankAccountTransferWithDrawal = 6
const PurchaseReceipt = 7
const VatReport = 8
const SieImport = 9
const BankTransactionDeposit = 10
const BankTransactionWithdrawal = 11
const SupplierInvoiceDebit = 12
const SupplierInvoiceCredit = 13
const CustomerInvoiceDebit = 14
const CustomerInvoiceCredit = 15
const ClaimOnCardAcquirer = 16
const TaxReturn = 17
const AllocationPeriod = 18
const AllocationPeriodCorrection = 19
const InventoryEvent = 20
const EmployerReport = 21
const Payslip = 22
const CustomerQuickInvoiceDebit = 23
const CustomerQuickInvoiceCredit = 24
const SupplierQuickInvoiceDebit = 25
const SupplierQuickInvoiceCredit = 26
const IZettleVoucher = 27

func (c *Client) Vouchers(id ...string) ([]Voucher, error) {
	resource := "vouchers"
	resp := &struct {
		Meta Meta
		Data []Voucher
	}{}
	if len(id) > 2 {
		return nil, fmt.Errorf("vouchers can only take one optional fiscal year and voucher id")
	} else if len(id) == 1 {
		resource = resource + "/" + id[0]
	} else if len(id) == 2 {
		resource = resource + "/" + id[0] + "/" + id[1]
	}
	err := c.GetRequest(resource, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) NewVoucher(voucher Voucher) (*Voucher, error) {
	resource := "vouchers"
	resp := &Voucher{}
	err := c.PostRequest(resource, voucher, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type VoucherRow struct {
	AccountNumber      int        `url:"AccountNumber,omitempty"`
	AccountDescription string     `url:"AccountDescription,omitempty"`
	DebitAmount        util.Money `url:"DebitAmount,omitempty"`
	CreditAmount       util.Money `url:"CreditAmount,omitempty"`
	TransactionText    string     `url:"TransactionText,omitempty"`
	CostCenterItemID1  string     `url:"CostCenterItemId1,omitempty"`
	CostCenterItemID2  string     `url:"CostCenterItemId2,omitempty"`
	CostCenterItemID3  string     `url:"CostCenterItemId3,omitempty"`
	VatCodeID          string     `url:"VatCodeId,omitempty"`
	VatCodeAndPercent  string     `url:"VatCodeAndPercent,omitempty"`
	Quantity           int        `url:"Quantity,omitempty"`
	Weight             int        `url:"Weight,omitempty"`
	DeliveryDate       time.Time  `url:"DeliveryDate,omitempty"`
	HarvestYear        int        `url:"HarvestYear,omitempty"`
	ProjectID          string     `url:"ProjectId,omitempty"`
}

type VoucherAttachment struct {
	DocumentID    string   `url:"DocumentId,omitempty"`
	DocumentType  int      `url:"DocumentType,omitempty"`
	AttachmentIds []string `url:"AttachmentIds,omitempty"`
}

type Voucher struct {
	ID                    string             `url:"Id,omitempty"`
	VoucherDate           util.Date          `url:"VoucherDate,omitempty"`
	VoucherText           string             `url:"VoucherText,omitempty"`
	Rows                  []VoucherRow       `url:"Rows,omitempty"`
	NumberAndNumberSeries string             `url:"NumberAndNumberSeries,omitempty"`
	NumberSeries          string             `url:"NumberSeries,omitempty"`
	Attachments           *VoucherAttachment `url:"Attachments,omitempty"`
	ModifiedUtc           time.Time          `url:"ModifiedUtc,,omitempty"`
	VoucherType           int                `url:"VoucherType,omitempty"`
	SourceID              string             `url:"SourceId,omitempty"`
	CreatedUtc            time.Time          `url:"CreatedUtc,omitempty"`
}
