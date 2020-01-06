package main

import (
	"fmt"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/preferences"
	"izettle-daily-reports/visma"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Printf("izettle-report-generator run at %s\n\n", time.Now())

	fmt.Print("Loading .env file... ")
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("FAILED: %s\n", err)
	} else {
		fmt.Println("DONE")
	}

	fmt.Print("Reading environment variables... ")
	//googleApplicationID := MustGetEnv("GDRIVE_CLIENT_ID")
	//googleApplicationSecret := MustGetEnv("GDRIVE_CLIENT_SECRET")
	//driveDirectoryID := MustGetEnv("GDRIVE_FOLDER_ID")
	izettleEmail := MustGetEnv("IZETTLE_EMAIL")
	izettlePassword := MustGetEnv("IZETTLE_PASSWORD")
	izettleApplicationID := MustGetEnv("IZETTLE_CLIENT_ID")
	izettleApplicationSecret := MustGetEnv("IZETTLE_CLIENT_SECRET")
	vismaApplicationID := MustGetEnv("VISMA_CLIENT_ID")
	vismaApplicationSecret := MustGetEnv("VISMA_CLIENT_SECRET")
	fmt.Println("DONE")

	//fmt.Print("Logging in to your google account... ")
	//token, err := google.Login(googleApplicationID, googleApplicationSecret, []string{drive.DriveScope, gmail.GmailSendScope})
	//handleError(err)
	//driveService, err := drive.NewService(context.Background(), option.WithTokenSource(token))
	//handleError(err)
	//folders, err := storage.Children(driveService, driveDirectoryID)
	//handleError(err)
	//fmt.Println("DONE")

	fmt.Print("Logging in to your izettle account... ")
	iz, err := izettle.Login(izettleEmail, izettlePassword, izettleApplicationID, izettleApplicationSecret)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Logging in to your visma account... ")
	vi, err := visma.Login(vismaApplicationID, vismaApplicationSecret, visma.Sandbox)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Fetching visma metadata... ")
	cc, err := vi.CostCenters()
	handleError(err)
	projects, err := vi.Projects()
	handleError(err)
	currentYear, err := vi.CurrentFiscalYear()
	handleError(err)
	fmt.Println("DONE")
	uncategorizedIzettlePrj := projects[0]

	fmt.Print("Fetching visma vouchers... ")
	vouchers, err := vi.Vouchers(currentYear.ID)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Fetching izettle products... ")
	products, err := iz.Products()
	handleError(err)
	fmt.Println("DONE")
	fmt.Print("Fetching izettle purchases... ")
	purchases, err := iz.Purchases()
	handleError(err)
	fmt.Println("DONE")
	reports := izettle.Reports(*purchases, products)

	unmatchedVouchers, err := getUnmatchedVouchers(reports, vouchers, cc[0].Items)
	handleError(err)
	unmatchedReports, err := getUnmatchedReports(reports, vouchers, cc[0].Items)
	handleError(err)

	ignoredReports := []izettle.Report{}
	pendingVouchers := []visma.Voucher{}
	for _, report := range unmatchedReports {
		date, err := visma.DateFromString(report.Date)
		handleError(err)
		costCenter, err := getReportCostCenter(report, cc[0].Items)
		if err != nil {
			// We did not manage to lookup the cost center for the report,
			// this is probably due to a name being wrong in izettle
			// or a new cost center (utskott/kommitte) has been
			// added to izettle but not to visma
			ignoredReports = append(ignoredReports, report)
			continue
		}
		handleError(err)
		vismaAccountRows, err := report.RowsByVismaAccount()
		handleError(err)

		var rows []visma.VoucherRow
		rows = append(rows, visma.VoucherRow{
			AccountNumber:     preferences.IzettleLedgerAccountNumber,
			DebitAmount:       decimal.NewFromInt(int64(report.Sum())).Div(decimal.NewFromInt(100)),
			CostCenterItemID1: costCenter.ID,
			ProjectID:         uncategorizedIzettlePrj.ID,
		})
		for _, s := range vismaAccountRows {
			rows = append(rows, visma.VoucherRow{
				AccountNumber:     s.VismaAccount,
				CreditAmount:      decimal.NewFromInt(int64(s.Amount)).Div(decimal.NewFromInt(100)),
				CostCenterItemID1: costCenter.ID,
				ProjectID:         uncategorizedIzettlePrj.ID,
			})
		}
		pendingVouchers = append(pendingVouchers, visma.Voucher{
			VoucherDate: date,
			VoucherText: "Uncategorized iZettle Import",
			Rows:        rows,
		})
	}

	if len(unmatchedVouchers) > 0 {
		fmt.Printf("Found the following vouchers not belonging to any report!\n")
		for _, v := range unmatchedVouchers {
			fmt.Printf(" - %s\t%s\n", v.VoucherDate.String(), v.VoucherText)
		}
		fmt.Println()
	}

	if len(ignoredReports) > 0 {
		fmt.Printf("Failed to generate vouchers for the following repports\n")
		for _, r := range ignoredReports {
			fmt.Printf(" - %s\t%s\n", r.Date, r.Username)
		}
		fmt.Println()
	}

	if len(pendingVouchers) == 0 {
		fmt.Printf("Everything is up to date!")
		os.Exit(0)
	}

	fmt.Printf("Preparing to uppload %d vouchers\n", len(pendingVouchers))
	for _, v := range pendingVouchers {
		sum, err := getVoucherSum(v)
		handleError(err)
		fmt.Printf(" * %s\t%s\t%s\n", v.VoucherDate.String(), v.VoucherText, sum.String())
		for _, r := range v.Rows {
			costCenter, err := getVoucherCostCenter(v, cc[0].Items)
			handleError(err)
			fmt.Printf("     %d\t%s\t%s\t%s\n", r.AccountNumber, costCenter.ShortName, r.DebitAmount.String(), r.CreditAmount.String())
		}
		handleError(err)
	}
	if true {
		os.Exit(0)
	}
	fmt.Println()
	fmt.Printf("Upploading vouchers...")
	for _, v := range pendingVouchers {
		sum, err := getVoucherSum(v)
		handleError(err)
		fmt.Printf(" + %s\t%s\t%d...", v.VoucherDate.String(), v.VoucherText, sum.String())
		_, err = vi.NewVoucher(v)
		if err != nil {
			fmt.Printf(" FAILED! %s", err)
		} else {
			fmt.Println(" DONE!")
		}
	}
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
			if voucher.VoucherDate.String() != report.Date {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			sum, err := getVoucherSum(voucher)
			handleError(err)
			if *sum != decimal.NewFromInt(int64(report.Sum())) {
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
			if voucher.VoucherDate.String() != report.Date {
				// If the dates do not match,
				// so we know it can't be the same sale
				continue
			}
			sum, err := getVoucherSum(voucher)
			handleError(err)
			if *sum != decimal.NewFromInt(int64(report.Sum())) {
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

func getVoucherSum(voucher visma.Voucher) (*decimal.Decimal, error) {
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

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func MustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Failed to get environment variable '%s'", name)
	}
	return value
}
