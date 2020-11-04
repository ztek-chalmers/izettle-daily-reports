package generate

import (
	"fmt"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/util"
	"izettle-daily-reports/visma"
)

type User struct {
	Izettle IZettleUser
	Visma   VismaUser
}

type VismaUser struct {
	Name string
}

type IZettleUser struct {
	Name string
	UUID string
}

type Matcher struct {
	ledgerAccountNumber int
	bankAccountNumbers  []int
	users               []User
}

func NewMatcher(ledgerAccountNumber int, bankAccountNumbers []int, users []User) Matcher {
	return Matcher{
		ledgerAccountNumber: ledgerAccountNumber,
		bankAccountNumbers:  bankAccountNumbers,
		users:               users,
	}
}

func (m *Matcher) GetReportCostCenter(report izettle.Report, costCenterItems []visma.CostCenterItem) (*visma.CostCenterItem, error) {
	for _, cc := range costCenterItems {
		if m.IsSameUser(report, cc) {
			return &cc, nil
		}
	}
	return nil, fmt.Errorf("failed to lookup cost center for report: %s %s", report.Date, report.Username)
}

func (m *Matcher) IsIZettleRelated(voucher visma.Voucher) bool {
	hasLedger := false
	hasBank := false
	for _, row := range voucher.Rows {
		if row.AccountNumber == m.ledgerAccountNumber {
			hasLedger = true
		}
		for _, n := range m.bankAccountNumbers {
			if row.AccountNumber == n {
				hasBank = true
			}
		}
	}
	return hasLedger && !hasBank
}

func (m *Matcher) GetVoucherCostCenter(voucher visma.Voucher, costCenterItems []visma.CostCenterItem) (*visma.CostCenterItem, error) {
	for _, row := range voucher.Rows {
		if row.AccountNumber != m.ledgerAccountNumber {
			continue
		}
		for _, costCenterItem := range costCenterItems {
			if costCenterItem.ID == row.CostCenterItemID1 {
				return &costCenterItem, nil
			}
		}
		if row.CostCenterItemID1 == "" {
			return nil, fmt.Errorf("voucher does not have a cost center set: %s - %s", voucher.VoucherText, voucher.VoucherDate.String())
		}
		return nil, fmt.Errorf("failed to lookup cost center for voucher: %s", voucher.ID)
	}
	return nil, fmt.Errorf("voucher is not imported: %s", voucher.ID)
}

func (m *Matcher) GetVoucherSum(voucher visma.Voucher) (*util.Money, error) {
	for _, row := range voucher.Rows {
		if row.AccountNumber != m.ledgerAccountNumber {
			continue
		}
		return &row.DebitAmount, nil
	}
	return nil, fmt.Errorf("failed to get sum from voucher")
}

func (m *Matcher) isImportedVoucher(voucher visma.Voucher) bool {
	if voucher.VoucherType != visma.SieImport {
		// It's not an imported voucher,
		// so we know it can't be the same sale
		return false
	}
	return m.IsIZettleRelated(voucher)
}

func (m *Matcher) GetUnmatchedReports(reports []izettle.Report, vouchers []visma.Voucher, costCenterItems []visma.CostCenterItem) ([]izettle.Report, error) {
	var unmatchedReports []izettle.Report
	for _, report := range reports {
		exists := false
		for _, voucher := range vouchers {
			if !m.IsIZettleRelated(voucher) {
				// The voucher is not related to iZettle,
				// so it's not a sale
				continue
			}
			costCenter, err := m.GetVoucherCostCenter(voucher, costCenterItems)
			if err != nil {
				return nil, err
			}
			sum, err := m.GetVoucherSum(voucher)
			if err != nil {
				return nil, err
			}
			if !voucher.VoucherDate.Equal(report.Date) {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			if !m.IsSameUser(report, *costCenter) {
				// The report and voucher did not reference the same user
				// so we know it can't be the same sale
				continue
			}
			if !sum.Equal(report.Sum().Decimal) {
				// The price amounts do not match,
				// so we know it can't be the same sale
				return nil, fmt.Errorf("found voucher with the correct date and user but not the same sum: %s %s %s", report.Date.String(), report.Username, voucher.NumberAndNumberSeries)
			}
			if !m.isImportedVoucher(voucher) {
				fmt.Printf("The voucher %s is manually created, but is will be used in place of a new voucher", voucher.ID)
			}
			exists = true
			break
		}
		if !exists {
			unmatchedReports = append(unmatchedReports, report)
		}
	}
	return unmatchedReports, nil
}

func (m *Matcher) GetUnmatchedVouchers(reports []izettle.Report, vouchers []visma.Voucher, costCenterItems []visma.CostCenterItem) ([]visma.Voucher, error) {
	var unmatchedVouchers []visma.Voucher
	for _, voucher := range vouchers {
		if !m.isImportedVoucher(voucher) {
			// It's not an imported voucher,
			// so we know it can't be the same sale
			continue
		}
		costCenter, err := m.GetVoucherCostCenter(voucher, costCenterItems)
		if err != nil {
			return nil, err
		}
		sum, err := m.GetVoucherSum(voucher)
		if err != nil {
			return nil, err
		}
		exists := false
		for _, report := range reports {
			if !voucher.VoucherDate.Equal(report.Date) {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			if !m.IsSameUser(report, *costCenter) {
				// The report and voucher did not reference the same user
				// so we know it can't be the same sale
				continue
			}
			if !sum.Equal(report.Sum().Decimal) {
				// The price amounts do not match,
				// so we know it can't be the same sale
				return nil, fmt.Errorf("found voucher with the correct date and user but not the same sum: %s %s %s", report.Date.String(), report.Username, voucher.NumberAndNumberSeries)
			}
			exists = true
			break
		}
		if !exists {
			unmatchedVouchers = append(unmatchedVouchers, voucher)
		}
	}
	return unmatchedVouchers, nil
}

func (m *Matcher) IsSameUser(report izettle.Report, costCenter visma.CostCenterItem) bool {
	for _, r := range m.users {
		if r.Izettle.Name == report.Username && r.Visma.Name == costCenter.ShortName {
			return true
		}
	}
	return false
}
