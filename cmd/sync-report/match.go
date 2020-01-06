package main

import (
	"fmt"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/preferences"
	"izettle-daily-reports/util"
	"izettle-daily-reports/visma"

	"github.com/shopspring/decimal"
)

func getReportCostCenter(report izettle.Report, costCenterItems []visma.CostCenterItem) (*visma.CostCenterItem, error) {
	for _, cc := range costCenterItems {
		if isSameUser(report, cc) {
			return &cc, nil
		}
	}
	return nil, fmt.Errorf("failed to lookup cost center for report: %s %s", report.Date, report.Username)
}

func getVoucherCostCenter(voucher visma.Voucher, costCenterItems []visma.CostCenterItem) (*visma.CostCenterItem, error) {
	for _, row := range voucher.Rows {
		if row.AccountNumber != preferences.IzettleLedgerAccountNumber {
			continue
		}
		for _, costCenterItem := range costCenterItems {
			if costCenterItem.ID == row.CostCenterItemID1 {
				return &costCenterItem, nil
			}
		}
		if row.CostCenterItemID1 == "" {
			return nil, fmt.Errorf("voucher does not have a cost center set: %s", voucher.ID)
		}
		return nil, fmt.Errorf("failed to lookup cost center for voucher: %s", voucher.ID)
	}
	return nil, fmt.Errorf("voucher is not imported: %s", voucher.ID)
}

func getVoucherSum(voucher visma.Voucher) (*util.Decimal, error) {
	for _, row := range voucher.Rows {
		if row.AccountNumber != preferences.IzettleLedgerAccountNumber {
			continue
		}
		return &row.DebitAmount, nil
	}
	return nil, fmt.Errorf("failed to get sum from voucher")
}

func isImportedVoucher(voucher visma.Voucher) bool {
	if voucher.VoucherType != visma.SieImport {
		// It's not an imported voucher,
		// so we know it can't be the same sale
		return false
	}
	for _, row := range voucher.Rows {
		if row.AccountNumber == preferences.IzettleLedgerAccountNumber {
			return true
		}
	}
	return false
}

func getUnmatchedReports(reports []izettle.Report, vouchers []visma.Voucher, costCenterItems []visma.CostCenterItem) ([]izettle.Report, error) {
	var unmatchedReports []izettle.Report
	for _, report := range reports {
		exists := false
		for _, voucher := range vouchers {
			if !isImportedVoucher(voucher) {
				// It's not an imported voucher,
				// so we know it can't be the same sale
				continue
			}
			if voucher.VoucherDate.Equal(report.Date) {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			sum, err := getVoucherSum(voucher)
			handleError(err)
			if !sum.Equal(decimal.NewFromInt(int64(report.Sum()))) {
				// The price amounts do not match,
				// so we know it can't be the same sale
				continue
			}
			costCenter, err := getVoucherCostCenter(voucher, costCenterItems)
			if err != nil {
				return nil, err
			}
			if !isSameUser(report, *costCenter) {
				// The report and voucher did not reference the same user
				// so we know it can't be the same sale
				continue
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

func getUnmatchedVouchers(reports []izettle.Report, vouchers []visma.Voucher, costCenterItems []visma.CostCenterItem) ([]visma.Voucher, error) {
	var unmatchedVouchers []visma.Voucher
	for _, voucher := range vouchers {
		if !isImportedVoucher(voucher) {
			// It's not an imported voucher,
			// so we know it can't be the same sale
			continue
		}
		exists := false
		for _, report := range reports {
			if voucher.VoucherDate.Equal(report.Date) {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			sum, err := getVoucherSum(voucher)
			handleError(err)
			if !sum.Equal(decimal.NewFromInt(int64(report.Sum()))) {
				// The price amounts do not match,
				// so we know it can't be the same sale
				continue
			}
			costCenter, err := getVoucherCostCenter(voucher, costCenterItems)
			if err != nil {
				return nil, err
			}
			if !isSameUser(report, *costCenter) {
				// The report and voucher did not reference the same user
				// so we know it can't be the same sale
				continue
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

func isSameUser(report izettle.Report, costCenter visma.CostCenterItem) bool {
	for _, r := range preferences.IzettleVismaMap {
		if r.Izettle == report.Username && r.Visma == costCenter.ShortName {
			return true
		}
	}
	return false
}
