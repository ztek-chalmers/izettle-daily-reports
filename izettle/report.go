package izettle

import (
	"fmt"
	"izettle-daily-reports/preferences"
	"sort"
	"strconv"
)

type Report struct {
	Date     string
	User     int
	Username string
	Rows     []ReportRow
}

type ReportRow struct {
	Name         string
	Count        int
	Amount       int
	VismaAccount int
}

type VismaRow struct {
	Amount       int
	VismaAccount int
}

func (r Report) Sum() int {
	s := 0
	for _, p := range r.Rows {
		s += p.Amount
	}
	return s
}

func (r Report) RowsByVismaAccount() ([]VismaRow, error) {
	accounts := make(map[int]VismaRow)
	for _, row := range r.Rows {
		if row.VismaAccount == 0 {
			return nil, fmt.Errorf("row contained an item without a visma account: %s", row.Name)
		}
		account := accounts[row.VismaAccount]
		account.VismaAccount = row.VismaAccount
		account.Amount += row.Amount
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

func Reports(purchases Purchases, products []Product) []Report {
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
					VismaAccount: preferences.OtherIncomeAccountNumber,
				})
			}
		}
		reports = append(reports, Report{
			Date:     purchase.Date,
			User:     purchase.User,
			Username: purchase.Username,
			Rows:     rows,
		})
	}

	sort.SliceStable(reports, func(i, j int) bool {
		return reports[i].Date < reports[j].Date
	})
	return reports
}
