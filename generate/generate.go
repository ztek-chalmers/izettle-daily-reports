package generate

import (
	"fmt"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/visma"
)

type Generator struct {
	matcher Matcher
}

func NewGenerator(matcher Matcher) Generator {
	return Generator{
		matcher: matcher,
	}
}

type PendingVoucher struct {
	Voucher     visma.Voucher
	Attachments [][]byte
}

func (g *Generator) GeneratePendingVouchers(unmatchedReports []izettle.Report, costCenterItems []visma.CostCenterItem, vismaProject visma.Project) ([]PendingVoucher, []izettle.Report, error) {
	ignoredReports := []izettle.Report{}
	pendingVouchers := []PendingVoucher{}
	for _, report := range unmatchedReports {
		costCenter, err := g.matcher.GetReportCostCenter(report, costCenterItems)
		if err != nil {
			fmt.Printf("We did not manage to lookup the cost center for the user %s\n"+
				"this is probably due to a name being wrong in izettle"+
				"or a new cost center (utskott/kommitte) has been"+
				"added to izettle but not to visma)\n", report.Username)
			ignoredReports = append(ignoredReports, report)
			continue
		}
		vismaAccountRows, err := report.RowsByVismaAccount()
		if err != nil {
			return nil, nil, err
		}
		var rows []visma.VoucherRow
		rows = append(rows, visma.VoucherRow{
			AccountNumber:     g.matcher.ledgerAccountNumber,
			DebitAmount:       report.Sum(),
			CostCenterItemID1: costCenter.ID,
			ProjectID:         vismaProject.ID,
		})
		for _, s := range vismaAccountRows {
			rows = append(rows, visma.VoucherRow{
				AccountNumber:     s.VismaAccount,
				CreditAmount:      s.Amount,
				CostCenterItemID1: costCenter.ID,
				ProjectID:         vismaProject.ID,
			})
		}
		voucher := visma.Voucher{
			VoucherDate: report.Date,
			VoucherText: "Uncategorized iZettle Import",
			Rows:        rows,
		}
		pendingVouchers = append(pendingVouchers, PendingVoucher{
			Voucher:     voucher,
			Attachments: report.Attachments,
		})
	}
	return pendingVouchers, ignoredReports, nil
}
