package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"izettle-daily-reports/generate"
	"izettle-daily-reports/izettle"
	"izettle-daily-reports/util"
	"izettle-daily-reports/visma"
	"log"
	"time"
)

type Preferences struct {
	DryRun   bool
	FromDate util.Date
	Users    []generate.User
	Visma    VismaPreferences
	IZettle  IZettlePreferences
}

type IZettlePreferences struct {
	Email        string
	Password     string
	ClientID     string
	ClientSecret string
}

type VismaPreferences struct {
	LedgerAccountNumber        int
	BankAccountNumbers         []int
	OtherIncomeAccountNumber   int
	UncategorizedProjectNumber string
	ClientID                   string
	ClientSecret               string
}

func main() {
	fmt.Printf("izettle-report-generator run at %s\n\n", time.Now())

	fmt.Print("Reading config.json... ")
	prefData, err := ioutil.ReadFile("config.json")
	handleError(err)
	pref := Preferences{}
	err = json.Unmarshal(prefData, &pref)
	handleError(err)
	fmt.Println("DONE")
	fmt.Println()

	fmt.Println("Logging in:")
	fmt.Print("  izettle account using official API... ")
	iz, err := izettle.Login(pref.IZettle.Email, pref.IZettle.Password, pref.IZettle.ClientID, pref.IZettle.ClientSecret)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("  izettle account using browser cookie... ")
	token, err := ioutil.ReadFile("tokens/_izsessionat.token")
	izBrowser := izettle.BrowserLoginCookie(string(token))
	if !izBrowser.IsLoggedIn() {
		cookie := ""
		izBrowser, cookie, err = izettle.BrowserLoginEmail(pref.IZettle.Email, pref.IZettle.Password)
		err = ioutil.WriteFile("tokens/_izsessionat.token", []byte(cookie), 0644)
		handleError(err)
	}
	fmt.Println("DONE")

	fmt.Print("  visma account... ")
	vi, err := visma.Login(pref.Visma.ClientID, pref.Visma.ClientSecret, visma.Sandbox)
	handleError(err)
	fmt.Println("DONE")
	fmt.Println()

	fmt.Println("Fetching:")
	fmt.Print("  visma metadata... ")
	cc, err := vi.CostCenters()
	handleError(err)
	projects, err := vi.Projects()
	handleError(err)
	currentYear, err := vi.CurrentFiscalYear()
	handleError(err)
	fmt.Println("DONE")

	// We only import reports created more than 2 days ago, this is to make sure that we do not
	// import a half finished report.
	toDate := util.DateFromTime(time.Now().AddDate(0, 0, -2))
	fromDate := currentYear.StartDate
	if !pref.FromDate.Before(fromDate) {
		fromDate = pref.FromDate
	} else {
		fmt.Println("\n * From date was before the start of this year and is therefor ignored.")
	}
	if fromDate.After(toDate) {
		fmt.Println("\nThe from date is after the to date. This will not result in any imports.\nABORTING!")
		return
	}

	// Find the uncategorized project id
	var uncategorizedIzettlePrj visma.Project
	for _, p := range projects {
		if p.Number == pref.Visma.UncategorizedProjectNumber {
			uncategorizedIzettlePrj = p
			break
		}
	}
	if uncategorizedIzettlePrj.ID == "" {
		handleError(fmt.Errorf("unable to find poject with number: %s", pref.Visma.UncategorizedProjectNumber))
	}

	fmt.Printf("  visma vouchers between %s and %s... ", fromDate.String(), toDate.String())
	vouchers, err := vi.Vouchers(fromDate, toDate, currentYear.ID)
	handleError(err)
	fmt.Println("DONE")

	fmt.Print("  izettle products... ")
	products, err := iz.Products()
	handleError(err)
	fmt.Println("DONE")
	fmt.Printf("  izettle purchases between %s and %s... ", fromDate.String(), toDate.String())
	purchases, err := iz.Purchases(fromDate, toDate)
	handleError(err)
	fmt.Println("DONE")
	fmt.Println()

	fmt.Print("Matching iZettle reports with Visma vouchers... ")
	matcher := generate.NewMatcher(pref.Visma.LedgerAccountNumber, pref.Visma.BankAccountNumbers, pref.Users)
	generator := generate.NewGenerator(matcher)
	reports := izettle.Reports(*purchases, products, pref.Visma.OtherIncomeAccountNumber)
	unmatchedVouchers, err := matcher.GetUnmatchedVouchers(reports, vouchers, cc[0].Items)
	handleError(err)
	unmatchedReports, err := matcher.GetUnmatchedReports(reports, vouchers, cc[0].Items)
	handleError(err)
	fmt.Println("DONE")
	fmt.Println()

	if len(unmatchedReports) == 0 {
		fmt.Printf("All %d reports are already imported into visma. Just chilaxing for now.\n", len(reports))
		return
	}

	fmt.Println("Generating:")
	if !pref.DryRun {
		fmt.Println("  PDFs...")
		for i, r := range unmatchedReports {
			fmt.Printf(" * %d of %d (%s %s)\n", i+1, len(unmatchedReports), r.Username, r.Date.String())
			pdf, err := izBrowser.DayReportToPDF(r)
			handleError(err)
			data, err := ioutil.ReadAll(pdf)
			handleError(err)
			unmatchedReports[i].Attachments = append(unmatchedReports[i].Attachments, data)
		}
	} else {
		fmt.Println("  no PDFs: dry run")
	}

	fmt.Print("  vouchers... ")
	pendingVouchers, ignoredReports, err := generator.GeneratePendingVouchers(unmatchedReports, cc[0].Items, uncategorizedIzettlePrj)
	handleError(err)
	fmt.Println("DONE")
	fmt.Println()

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

	fmt.Printf("Preparing to upload %d vouchers\n", len(pendingVouchers))
	for _, v := range pendingVouchers {
		sum, err := matcher.GetVoucherSum(v.Voucher)
		handleError(err)
		fmt.Printf("  * %s\t%s\t%s\n", v.Voucher.VoucherDate.String(), v.Voucher.VoucherText, sum.String())
		for _, r := range v.Voucher.Rows {
			costCenter, err := matcher.GetVoucherCostCenter(v.Voucher, cc[0].Items)
			handleError(err)
			fmt.Printf("          %d\t%s\t%s\t%s\n", r.AccountNumber, costCenter.ShortName, r.DebitAmount.String(), r.CreditAmount.String())
		}
		handleError(err)
	}
	fmt.Println()

	if pref.DryRun {
		fmt.Println("This was a dry run so no new vouchers where uploaded.")
		return
	}

	fmt.Printf("Upploading vouchers...\n")
	for _, v := range pendingVouchers {
		sum, err := matcher.GetVoucherSum(v.Voucher)
		handleError(err)
		fmt.Printf(" + %s\t%s\t%s...", v.Voucher.VoucherDate.String(), v.Voucher.VoucherText, sum.String())

		attachmentData := base64.StdEncoding.EncodeToString(v.Attachments[0])
		handleError(err)
		attachmentName := fmt.Sprintf("Autogenerated_%s_%s.pdf", v.Voucher.Rows[0].CostCenterItemID1, v.Voucher.VoucherDate.String())
		attachment, err := vi.NewAttachment(attachmentName, "application/pdf", attachmentData)
		if err != nil {
			fmt.Printf(" FAILED! \n %s\n", err)
			continue
		}
		v.Voucher.Attachments = &visma.VoucherAttachment{
			DocumentType:  2, // Receipt
			AttachmentIds: []string{attachment.ID},
		}
		_, err = vi.NewVoucher(v.Voucher)
		handleError(err)
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
