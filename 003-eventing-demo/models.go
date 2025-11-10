package kndemo

type Payload struct {
	Content    string      `json:"content"`
	ParsedData *ParsedData `json:"parsed"`
	Customer   *Customer   `json:"customer"`
	Apology    string      `json:"apology"`
}

type ParsedData struct {
	CompanyID    *string `json:"company_id"`
	CompanyName  *string `json:"company_name"`
	Country      *string `json:"country"`
	CustomerName *string `json:"customer_name"`
	EmailAddress *string `json:"email_address"`
	Escalate     bool    `json:"escalate"`
	Phone        *string `json:"phone"`
	ProductName  *string `json:"product_name"`
	Sentiment    *string `json:"sentiment"`
}

type Customer struct {
	CustomerID   string `json:"customer_id"`
	CompanyName  string `json:"company_name"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
}
