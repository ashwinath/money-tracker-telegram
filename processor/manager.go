package processor

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/ashwinath/money-tracker-telegram/db"
	database "github.com/ashwinath/money-tracker-telegram/db"
)

const HelpMessageTemplate = `
{{ if .Error }}
{{ .Error }}
{{ end }}

_Adding a transaction_

User: {{ .UserAddCommandHelp }}
Service returns: {{ .UserAddResultHelp }}

_Deleting a transaction_

User: {{ .UserDelCommandHelp }}
Service returns: {{ .UserDelResultHelp }}`

const UserAddCommandHelp = "`ADD <TYPE> <CLASSIFICATION> <PRICE (no $ sign)> <Optional date in yyyy-mm-dd format>`"
const UserAddResultHelp = "`Created Transaction ID: <ID>, Transaction: <transaction>`"
const UserDelCommandHelp = "`DEL <ID>`"
const UserDelResultHelp = "`Deleted Transaction ID: <ID>, Transaction: <transaction>`"

const oneMonth = 1
const databaseQueryErrorFormat = "Could not query database, error: %s"

type ProcessorManager struct {
	db       *database.DB
	helpTmpl *template.Template
}

func NewManager(db *database.DB) (*ProcessorManager, error) {
	tmpl, err := template.New("help").Parse(HelpMessageTemplate)
	if err != nil {
		return nil, err
	}

	return &ProcessorManager{
		db:       db,
		helpTmpl: tmpl,
	}, nil
}

type UserHelp struct {
	Error              string
	UserAddCommandHelp string
	UserAddResultHelp  string
	UserDelCommandHelp string
	UserDelResultHelp  string
}

func (m *ProcessorManager) showHelp(err error) *string {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	args := &UserHelp{
		Error:              errString,
		UserAddCommandHelp: UserAddCommandHelp,
		UserAddResultHelp:  UserAddResultHelp,
		UserDelCommandHelp: UserDelCommandHelp,
		UserDelResultHelp:  UserDelResultHelp,
	}

	buf := &bytes.Buffer{}
	err = m.helpTmpl.Execute(buf, args)
	if err != nil {
		helpString := fmt.Sprintf(
			"error templating your message, please contact the author. error: %s",
			err,
		)
		return &helpString
	}
	helpString := buf.String()
	return &helpString
}

func (m *ProcessorManager) ProcessMessage(message string, messageTime time.Time) *string {
	chunk, err := Parse(message)
	if err != nil {
		return m.showHelp(err)
	}

	return m.processChunk(chunk, messageTime)
}

func (m *ProcessorManager) processChunk(chunk *Chunk, messageTime time.Time) *string {
	switch chunk.Instruction {
	case Add:
		return m.processChunkAdd(chunk, messageTime)
	case Delete:
		return m.processChunkDelete(chunk, messageTime)
	case Generate:
		return m.processChunkGenerate(chunk)
	}

	return m.showHelp(nil)

}

func (m *ProcessorManager) processChunkAdd(chunk *Chunk, messageTime time.Time) *string {
	t := &db.Transaction{
		Date:           messageTime,
		Type:           chunk.Type,
		Classification: chunk.Classification,
		Amount:         chunk.Amount,
	}

	if chunk.Date != nil {
		t.Date = *chunk.Date
	}

	t, err := m.db.InsertTransaction(t)
	if err != nil {
		return m.showHelp(err)
	}

	returnString := fmt.Sprintf(
		"Created Transaction ID: %d\n%s",
		t.ID,
		t,
	)
	return &returnString
}

func (m *ProcessorManager) processChunkDelete(chunk *Chunk, messageTime time.Time) *string {
	deletedTx, err := m.db.DeleteTransaction(chunk.ID)
	if err != nil {
		return m.showHelp(err)
	}

	returnString := fmt.Sprintf(
		"Deleted Transaction ID: %d\n%s",
		chunk.ID,
		deletedTx,
	)
	return &returnString
}

type AsyncAggregateResult struct {
	Result *float64
	Error  error
}

type AsyncTransactionResults struct {
	Result []db.Transaction
	Error  error
}

func (m *ProcessorManager) processChunkGenerate(chunk *Chunk) *string {
	endDate := chunk.StartDate.AddDate(0, oneMonth, 0)

	// Generate TypeOwn (Others field)
	othersResultChannel := make(chan AsyncAggregateResult)
	go m.queryTypeOwnSum(*chunk.StartDate, endDate, othersResultChannel)

	// Generate Reimbursement (reimbursement field, will be negative)
	reimResultChannel := make(chan AsyncAggregateResult)
	go m.queryReimburseSum(*chunk.StartDate, endDate, reimResultChannel)

	// Generate shared expenses (list of transactions)
	sharedResultChannel := make(chan AsyncTransactionResults)
	go m.querySharedTransactions(*chunk.StartDate, endDate, sharedResultChannel)

	othersResult := <-othersResultChannel
	if othersResult.Error != nil {
		err := fmt.Sprintf(databaseQueryErrorFormat, othersResult.Error)
		return &err
	}

	reimResult := <-reimResultChannel
	if reimResult.Error != nil {
		err := fmt.Sprintf(databaseQueryErrorFormat, reimResult.Error)
		return &err
	}

	sharedResult := <-sharedResultChannel
	if sharedResult.Error != nil {
		err := fmt.Sprintf(databaseQueryErrorFormat, sharedResult.Error)
		return &err
	}

	// build the string
	var resStrings []string

	// others + reimbursements
	resStrings = append(resStrings, "```")
	resStrings = append(resStrings, "---expenses.csv---")
	resStrings = append(
		resStrings,
		fmt.Sprintf("%s,Others,%f", chunk.StartDate.Format(time.DateOnly), *othersResult.Result),
	)
	resStrings = append(
		resStrings,
		fmt.Sprintf("%s,Reimbursement,%f", chunk.StartDate.Format(time.DateOnly), *reimResult.Result*-1),
	)

	// other shared spending
	resStrings = append(resStrings, "shared_expenses.csv")
	for _, tx := range sharedResult.Result {
		var typeBuilder []string
		if strings.Contains(string(tx.Type), "SPECIAL") {
			typeBuilder = append(typeBuilder, "Special:")
		}
		typeBuilder = append(typeBuilder, tx.Classification)
		type_ := strings.Join(typeBuilder, "")

		date := tx.Date
		amount := tx.Amount
		resStrings = append(
			resStrings,
			fmt.Sprintf("%s,%s,%f", date.Format(time.DateOnly), type_, amount),
		)
	}

	resStrings = append(resStrings, "```")
	res := strings.Join(resStrings, "\n")

	return &res
}

func (m *ProcessorManager) queryTypeOwnSum(startDate, endDate time.Time, result chan<- AsyncAggregateResult) {
	defer close(result)
	othersOption := &db.FindTransactionOptions{
		StartDate: startDate,
		EndDate:   endDate,
		Types:     []db.TransactionType{db.TypeOwn},
	}

	othersTotal, err := m.db.AggregateTransactions(othersOption)
	result <- AsyncAggregateResult{
		Result: othersTotal,
		Error:  err,
	}
}

func (m *ProcessorManager) queryReimburseSum(startDate, endDate time.Time, result chan<- AsyncAggregateResult) {
	defer close(result)
	reimOption := &db.FindTransactionOptions{
		StartDate: startDate,
		EndDate:   endDate,
		Types: []db.TransactionType{
			db.TypeReimburse,
			db.TypeSharedReimburse,
			db.TypeSpecialSharedReimburse,
		},
	}

	reimTotal, err := m.db.AggregateTransactions(reimOption)
	result <- AsyncAggregateResult{
		Result: reimTotal,
		Error:  err,
	}
}

func (m *ProcessorManager) querySharedTransactions(startDate, endDate time.Time, result chan<- AsyncTransactionResults) {
	defer close(result)
	sharedOption := &db.FindTransactionOptions{
		StartDate: startDate,
		EndDate:   endDate,
		Types: []db.TransactionType{
			db.TypeSharedReimburse,
			db.TypeSpecialSharedReimburse,
			db.TypeSpecialShared,
			db.TypeShared,
		},
	}
	sharedTransactions, err := m.db.QueryTransactionByOptions(sharedOption)
	result <- AsyncTransactionResults{
		Result: sharedTransactions,
		Error:  err,
	}
}
