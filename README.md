# iZettle Daily Reports

This is a small tool to mirror izettle sales to visma. The report generator can be run with `./run.sh`.
izettle-daily-reports only imports reports which are 2 or more days old. This is to make sure that
reports which are half-done, if an import would happen in the between two sales on the dame day.
If it detects a partial import, it will fail.

## Installation

The report generator requires a go version `>1.13` so a installation script is included for installing
go on a Raspberry PI since the default package repository only includes an outdated version. 
To install go run `./raspberypi-install.sh`

## Configuration

The application can be configured by setting the following environment variables. Or create a file called
`.env` in the same folder as this README file with a copy of the follow snippet. Before you can run the
tool, you have to get the passwords, id's and secrets from their respective place.

This information should NEVER be saved in a public place!!!

```bash
# Your iZettle email adress
IZETTLE_EMAIL=
# Your iZettle password
IZETTLE_PASSWORD=
# The client which you can get by creating a new project at https://developer.izettle.com/login
IZETTLE_CLIENT_ID=
# The client secret which you can get by creating a new project at https://developer.izettle.com/login 
IZETTLE_CLIENT_SECRET=

# The client ID of the visma integration. Can be found:
# * In an e-mail from eAccountingApi@visma.com
# * Or by regestering an account at https://selfservice.developer.vismaonline.com/
VISMA_CLIENT_ID=
# The client secret of the visma integration. Can be found:
# * In an e-mail from eAccountingApi@visma.com
# * Or by regestering an account at https://selfservice.developer.vismaonline.com/
VISMA_CLIENT_SECRET=
```

Except for setting the authentication information, a few extra things must be specified.
Create a new file called `config.json` in the same folder as `.env` and copy the following
snippet. All the values can be changed based on the situation.

* `izettleLedgerAccountNumber` specifies the debit account used in the voucher
* `otherIncomeAccountNumber` specifies the default account to use if the voucher can not
                             be classified.
* `vismaUncategorizedProjectNumber` specifies the project number to use when creating a voucher.
                                    This is preferably one with a name like `Uncategorized iZettle Import`.
                                    
* `izettleVismaMap` specifies which izettle user should belong to which visma "cost center" (Kommitté).
                    FILL_THIS_IN is left blank since it is a name of an old treasurer. This name can be found
                    in izettle when looking at sales reports.

```json
{
  "fromDate": "2020-01-01",
  "izettleLedgerAccountNumber": 1690,
  "otherIncomeAccountNumber": 3110,
  "vismaUncategorizedProjectNumber": "1",
  "izettleVismaMap": [
    ["FILL_THIS_IN", "Ztyret"],
    ["DaltonZ .", "DaltonZ"],
    ["ZEXET .", "ZEXET"],
    ["ZnollK .", "ZØK"],
    ["ZIK .", "ZIK"],
    ["Argz .", "ArgZ"],
    ["Zenith .", "Zenith"],
    ["SNZ .", "SNZ"]
  ]
}
```