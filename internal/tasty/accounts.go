package tasty

import (
	"context"
	"net/http"
)

type Address struct {
	StreetOne   string `json:"street-one"`
	City        string `json:"city"`
	StateRegion string `json:"state-region"`
	PostalCode  string `json:"postal-code"`
	Country     string `json:"country"`
	IsForeign   bool   `json:"is-foreign"`
	IsDomestic  bool   `json:"is-domestic"`
}

type CustomerSuitability struct {
	ID                                int    `json:"id"`
	MaritalStatus                     string `json:"marital-status"`
	NumberOfDependents                int    `json:"number-of-dependents"`
	EmploymentStatus                  string `json:"employment-status"`
	Occupation                        string `json:"occupation"`
	EmployerName                      string `json:"employer-name"`
	JobTitle                          string `json:"job-title"`
	AnnualNetIncome                   int    `json:"annual-net-income"`
	NetWorth                          int    `json:"net-worth"`
	LiquidNetWorth                    int    `json:"liquid-net-worth"`
	StockTradingExperience            string `json:"stock-trading-experience"`
	CoveredOptionsTradingExperience   string `json:"covered-options-trading-experience"`
	UncoveredOptionsTradingExperience string `json:"uncovered-options-trading-experience"`
	FuturesTradingExperience          string `json:"futures-trading-experience"`
}

type Person struct {
	ExternalID         string `json:"external-id"`
	FirstName          string `json:"first-name"`
	LastName           string `json:"last-name"`
	BirthDate          string `json:"birth-date"`
	CitizenshipCountry string `json:"citizenship-country"`
	USACitizenshipType string `json:"usa-citizenship-type"`
	MaritalStatus      string `json:"marital-status"`
	NumberOfDependents int    `json:"number-of-dependents"`
	EmploymentStatus   string `json:"employment-status"`
	Occupation         string `json:"occupation"`
	EmployerName       string `json:"employer-name"`
	JobTitle           string `json:"job-title"`
}

type CustomerData struct {
	ID                              string              `json:"id"`
	FirstName                       string              `json:"first-name"`
	LastName                        string              `json:"last-name"`
	Address                         Address             `json:"address"`
	MailingAddress                  Address             `json:"mailing-address"`
	CustomerSuitability             CustomerSuitability `json:"customer-suitability"`
	USACitizenshipType              string              `json:"usa-citizenship-type"`
	IsForeign                       bool                `json:"is-foreign"`
	MobilePhoneNumber               string              `json:"mobile-phone-number"`
	Email                           string              `json:"email"`
	TaxNumberType                   string              `json:"tax-number-type"`
	TaxNumber                       string              `json:"tax-number"`
	BirthDate                       string              `json:"birth-date"`
	ExternalID                      string              `json:"external-id"`
	CitizenshipCountry              string              `json:"citizenship-country"`
	SubjectToTaxWithholding         bool                `json:"subject-to-tax-withholding"`
	AgreedToMargining               bool                `json:"agreed-to-margining"`
	AgreedToTerms                   bool                `json:"agreed-to-terms"`
	HasIndustryAffiliation          bool                `json:"has-industry-affiliation"`
	HasPoliticalAffiliation         bool                `json:"has-political-affiliation"`
	HasListedAffiliation            bool                `json:"has-listed-affiliation"`
	IsProfessional                  bool                `json:"is-professional"`
	HasDelayedQuotes                bool                `json:"has-delayed-quotes"`
	HasPendingOrApprovedApplication bool                `json:"has-pending-or-approved-application"`
	IdentifiableType                string              `json:"identifiable-type"`
	Person                          Person              `json:"person"`
}

type CustomerResponse struct {
	Context string       `json:"context"`
	Data    CustomerData `json:"data"`
}

type Account struct {
	AccountNumber         string `json:"account-number"`
	ExternalID            string `json:"external-id"`
	OpenedAt              string `json:"opened-at"`
	Nickname              string `json:"nickname"`
	AccountTypeName       string `json:"account-type-name"`
	DayTraderStatus       bool   `json:"day-trader-status"`
	IsClosed              bool   `json:"is-closed"`
	IsFirmError           bool   `json:"is-firm-error"`
	IsFirmProprietary     bool   `json:"is-firm-proprietary"`
	IsFuturesApproved     bool   `json:"is-futures-approved"`
	IsTestDrive           bool   `json:"is-test-drive"`
	MarginOrCash          string `json:"margin-or-cash"`
	IsForeign             bool   `json:"is-foreign"`
	FundingDate           string `json:"funding-date"`
	InvestmentObjective   string `json:"investment-objective"`
	FuturesAccountPurpose string `json:"futures-account-purpose"`
	SuitableOptionsLevel  string `json:"suitable-options-level"`
	CreatedAt             string `json:"created-at"`
}

type AccountContainer struct {
	Account        Account `json:"account"`
	AuthorityLevel string  `json:"authority-level"`
}

type AccountData struct {
	Items []AccountContainer `json:"items"`
}

type AccountResponse struct {
	Context string  `json:"context"`
	Data    Account `json:"data"`
}

type AccountsResponse struct {
	Context string      `json:"context"`
	Data    AccountData `json:"data"`
}

const (
	CustomerInfoPath = "/customers/me"
	AccountsPath     = "/customers/me/accounts"
)

func (c *TastyAPI) GetCustomer(ctx context.Context) (*CustomerResponse, error) {
	res := &CustomerResponse{}
	path := c.baseurl + CustomerInfoPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccounts(ctx context.Context) (*AccountResponse, error) {
	res := &AccountResponse{}
	path := c.baseurl + AccountsPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}
