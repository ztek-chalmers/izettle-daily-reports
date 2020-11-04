package izettle

import (
	"fmt"
	"izettle-daily-reports/util"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type Report struct {
	Date        util.Date
	UserID      int
	Username    string
	Rows        []ReportRow
	Attachments [][]byte
}

type ReportRow struct {
	Name         string
	Count        int
	Amount       util.Money
	VismaAccount int
}

type VismaRow struct {
	Amount       util.Money
	VismaAccount int
}

func (r Report) Sum() util.Money {
	d := decimal.Zero
	for _, p := range r.Rows {
		d = d.Add(p.Amount.Decimal)
	}
	return util.Money{d}
}

func (r Report) RowsByVismaAccount() ([]VismaRow, error) {
	accounts := make(map[int]VismaRow)
	for _, row := range r.Rows {
		if row.VismaAccount == 0 {
			return nil, fmt.Errorf("row contained an item without a visma account: %s", row.Name)
		}
		account := accounts[row.VismaAccount]
		account.VismaAccount = row.VismaAccount
		account.Amount = util.Money{account.Amount.Add(row.Amount.Decimal)}
		accounts[account.VismaAccount] = account
	}
	accountList := make([]VismaRow, 0)
	for _, a := range accounts {
		accountList = append(accountList, a)
	}
	return accountList, nil
}

func findProductVariant(uuid string, products []Product) (Product, Variant, bool) {
	for _, p := range products {
		for _, v := range p.Variants {
			if v.UUID == uuid {
				return p, v, true
			}
		}
	}
	return Product{}, Variant{}, false
}

func Reports(purchases Purchases, products []Product, defaultAccountNumber int) []Report {
	reports := []Report{}
	purchaseUnits := purchases.Group()
	for _, purchase := range purchaseUnits {
		rows := []ReportRow{}
		purchaseVariants := purchase.Summary()
		for variantUUID, pv := range purchaseVariants {
			product, variant, found := findProductVariant(variantUUID, products)
			if found {
				var name string
				if variant.Name == "" {
					name = product.Name
				} else {
					name = product.Name + ", " + variant.Name
				}
				barcode, _ := strconv.Atoi(variant.Barcode)
				s := pv.Summary()
				rows = append(rows, ReportRow{
					Name:         name,
					Count:        s.Count,
					Amount:       s.Amount,
					VismaAccount: barcode,
				})
			} else {
				name := "Custom product"
				s := pv.Summary()
				rows = append(rows, ReportRow{
					Name:         name,
					Count:        s.Count,
					Amount:       s.Amount,
					VismaAccount: defaultAccountNumber,
				})
			}
		}
		userName := strings.TrimSpace(strings.Split(purchase.Username, ".")[0])
		reports = append(reports, Report{
			Date:     purchase.Date,
			UserID:   purchase.User,
			Username: userName,
			Rows:     rows,
		})
	}

	sort.SliceStable(reports, func(i, j int) bool {
		return reports[i].Date.Before(reports[j].Date)
	})
	return reports
}
