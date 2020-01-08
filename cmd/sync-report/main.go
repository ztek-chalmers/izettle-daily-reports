package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"izettle-daily-reports/generate"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/util"
	"izettle-daily-reports/visma"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Preferences struct {
	FromDate                        util.Date
	IzettleLedgerAccountNumber      int
	OtherIncomeAccountNumber        int
	VismaUncategorizedProjectNumber string
	IzettleVismaMap                 []generate.IzettleVismaMapping
}

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

	fmt.Print("Reading config.json... ")
	prefData, err := ioutil.ReadFile("config.json")
	handleError(err)
	pref := Preferences{}
	err = json.Unmarshal(prefData, &pref)
	handleError(err)
	fromDate := currentYear.StartDate
	if pref.FromDate.After(fromDate) {
		fromDate = pref.FromDate
	} else {
		fmt.Println("\n * From date was before the start of this year and is therefor ignored.")
	}
	fmt.Println("DONE")

	var uncategorizedIzettlePrj visma.Project
	for _, p := range projects {
		if p.Number == pref.VismaUncategorizedProjectNumber {
			uncategorizedIzettlePrj = p
			break
		}
	}
	if uncategorizedIzettlePrj.ID == "" {
		handleError(fmt.Errorf("unable to find poject with number: %s", pref.VismaUncategorizedProjectNumber))
	}

	fmt.Printf("Fetching visma vouchers between %s and %s... ", fromDate.String(), currentYear.EndDate.String())
	vouchers, err := vi.Vouchers(pref.FromDate, currentYear.ID)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Fetching izettle products... ")
	products, err := iz.Products()
	handleError(err)
	fmt.Println("DONE")
	fmt.Printf("Fetching izettle purchases between %s and %s... ", fromDate.String(), currentYear.EndDate.String())
	purchases, err := iz.Purchases(fromDate, currentYear.EndDate)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("Generating vouchers... ")
	matcher := generate.NewMatcher(pref.IzettleLedgerAccountNumber, pref.IzettleVismaMap)
	generator := generate.NewGenerator(matcher)
	reports := izettle.Reports(*purchases, products, pref.OtherIncomeAccountNumber)
	unmatchedVouchers, err := matcher.GetUnmatchedVouchers(reports, vouchers, cc[0].Items)
	handleError(err)
	unmatchedReports, err := matcher.GetUnmatchedReports(reports, vouchers, cc[0].Items)
	handleError(err)
	pendingVouchers, ignoredReports, err := generator.GeneratePendingVouchers(unmatchedReports, cc[0].Items, uncategorizedIzettlePrj)
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
			fmt.Printf(" - %s\t%s\n", r.Date.String(), r.Username)
		}
		fmt.Println()
	}

	if len(pendingVouchers) == 0 {
		fmt.Printf("All %d reports are already imported into visma. Just chilaxing for now.", len(reports))
		os.Exit(0)
	}

	fmt.Printf("Preparing to upload %d vouchers\n", len(pendingVouchers))
	for _, v := range pendingVouchers {
		sum, err := matcher.GetVoucherSum(v)
		handleError(err)
		fmt.Printf(" * %s\t%s\t%s\n", v.VoucherDate.String(), v.VoucherText, sum.String())
		for _, r := range v.Rows {
			costCenter, err := matcher.GetVoucherCostCenter(v, cc[0].Items)
			handleError(err)
			fmt.Printf("     %d\t%s\t%s\t%s\n", r.AccountNumber, costCenter.ShortName, r.DebitAmount.String(), r.CreditAmount.String())
		}
		handleError(err)
	}
	fmt.Println()

	fmt.Printf("Upploading vouchers...\n")
	for _, v := range pendingVouchers {
		sum, err := matcher.GetVoucherSum(v)
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
