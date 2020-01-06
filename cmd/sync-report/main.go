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
	izettleEmail := mustGetEnv("IZETTLE_EMAIL")
	izettlePassword := mustGetEnv("IZETTLE_PASSWORD")
	izettleApplicationID := mustGetEnv("IZETTLE_CLIENT_ID")
	izettleApplicationSecret := mustGetEnv("IZETTLE_CLIENT_SECRET")
	vismaApplicationID := mustGetEnv("VISMA_CLIENT_ID")
	vismaApplicationSecret := mustGetEnv("VISMA_CLIENT_SECRET")
	fmt.Println("DONE")

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
	var uncategorizedIzettlePrj visma.Project
	for _, p := range projects {
		if p.Number == preferences.VismaUncategorizedProjectNumber {
			uncategorizedIzettlePrj = p
			break
		}
	}
	if uncategorizedIzettlePrj.ID == "" {
		handleError(fmt.Errorf("unable to find poject with number: %s", preferences.VismaUncategorizedProjectNumber))
	}

	fmt.Print("Fetching visma vouchers... ")
	vouchers, err := vi.Vouchers(currentYear.ID)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Fetching izettle products... ")
	products, err := iz.Products()
	handleError(err)
	fmt.Println("DONE")
	fmt.Print("Fetching izettle purchases... ")
	purchases, err := iz.Purchases(currentYear.StartDate, currentYear.EndDate)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Generating vouchers... ")
	reports := izettle.Reports(*purchases, products)
	unmatchedVouchers, err := getUnmatchedVouchers(reports, vouchers, cc[0].Items)
	handleError(err)
	unmatchedReports, err := getUnmatchedReports(reports, vouchers, cc[0].Items)
	handleError(err)
	pendingVouchers, ignoredReports, err := generatePendingVouchers(unmatchedReports, cc[0].Items, uncategorizedIzettlePrj)
	handleError(err)
	fmt.Println("DONE")

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
	fmt.Println()

	fmt.Printf("Upploading vouchers...\n")
	for _, v := range pendingVouchers {
		sum, err := getVoucherSum(v)
		handleError(err)
		fmt.Printf(" + %s\t%s\t%s...", v.VoucherDate.String(), v.VoucherText, sum.String())
		_, err = vi.NewVoucher(v)
		if err != nil {
			fmt.Printf(" FAILED! \n %s\n", err)
		} else {
			fmt.Println(" DONE!")
		}
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Failed to get environment variable '%s'", name)
	}
	return value
}
