package models

type BalanceResponse struct {
	Result string `json:"result"`
	Answer struct {
		Prepay     string `json:"prepay"`
		Credit     string `json:"credit"`
		CreditInfo struct {
			Active  string `json:"active"`
			Debt    string `json:"debt"`
			PayTill string `json:"pay_till"`
		} `json:"credit_info"`
	} `json:"answer"`
}
