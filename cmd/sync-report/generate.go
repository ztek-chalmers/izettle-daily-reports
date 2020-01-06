package main

import (
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/preferences"
	"izettle-daily-reports/util"
	"izettle-daily-reports/visma"
)

func generatePendingVouchers(unmatchedReports []izettle.Report, costCenterItems []visma.CostCenterItem, vismaProject visma.Project) ([]visma.Voucher, []izettle.Report, error) {
	ignoredReports := []izettle.Report{}
	pendingVouchers := []visma.Voucher{}
	for _, report := range unmatchedReports {
		costCenter, err := getReportCostCenter(report, costCenterItems)
		if err != nil {
			// We did not manage to lookup the cost center for the report,
			// this is probably due to a name being wrong in izettle
			// or a new cost center (utskott/kommitte) has been
			// added to izettle but not to visma
			ignoredReports = append(ignoredReports, report)
			continue
		}
		vismaAccountRows, err := report.RowsByVismaAccount()
		if err != nil {
			return nil, nil, err
		}
		var rows []visma.VoucherRow
		rows = append(rows, visma.VoucherRow{
			AccountNumber:     preferences.IzettleLedgerAccountNumber,
			DebitAmount:       util.DecimalFromFixedPointInt(report.Sum()),
			CostCenterItemID1: costCenter.ID,
			ProjectID:         vismaProject.ID,
		})
		for _, s := range vismaAccountRows {
			rows = append(rows, visma.VoucherRow{
				AccountNumber:     s.VismaAccount,
				CreditAmount:      util.DecimalFromFixedPointInt(s.Amount),
				CostCenterItemID1: costCenter.ID,
				ProjectID:         vismaProject.ID,
			})
		}
		pendingVouchers = append(pendingVouchers, visma.Voucher{
			VoucherDate: report.Date,
			VoucherText: "Uncategorized iZettle Import",
			Rows:        rows,
		})
	}
	return pendingVouchers, ignoredReports, nil
}
