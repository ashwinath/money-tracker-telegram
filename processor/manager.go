package processor

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	database "github.com/ashwinath/money-tracker-telegram/db"
)

const HelpMessageTemplate = `
{{ if .Error }}
{{ .Error }}
{{ end }}

Classification Types:
- REIM
- SHARED REIM
- SPECIAL SHARED REIM
- SHARED
- SPECIAL SHARED
- OWN
- SPECIAL OWN

_Adding a transaction_

User: {{ .UserAddCommandHelp }}
Service returns: {{ .UserAddResultHelp }}

_Adding a transaction for tax/insurance/tithe/credit card_

User: {{ .UserAddOthersCommandHelp }}
Service returns: {{ .UserAddResultHelp }}

_Deleting a transaction_

User: {{ .UserDelCommandHelp }}
Service returns: {{ .UserDelResultHelp }}

_Generating a monthly report_

User: {{ .UserGenCommandHelp }}
Service returns: {{ .UserGenResultHelp }}
`

const UserAddCommandHelp = "`ADD <TYPE> <CLASSIFICATION> <PRICE (no $ sign)> <Optional date in yyyy-mm-dd format>`"
const UserAddResultHelp = "`Created Transaction ID: <ID>, Transaction: <transaction>`"

const UserAddOthersCommandHelp = "`ADD <TYPE> <PRICE (no $ sign)> <Optional date in yyyy-mm-dd format>`"

const UserDelCommandHelp = "`DEL <ID>`"
const UserDelResultHelp = "`Deleted Transaction ID: <ID>, Transaction: <transaction>`"

const UserGenCommandHelp = "`GEN <Month> <Year>`"
const UserGenResultHelp = "```\n" +
	"---expenses.csv---\n" +
	"<Date>,Others,<Amount>\n" +
	"<Date>,Reimbursement,<Amount>\n" +
	"---shared_expenses.csv---\n" +
	"<Date>,<(Special):Classification>,<Amount>\n```"

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
	Error                    string
	UserAddCommandHelp       string
	UserAddResultHelp        string
	UserDelCommandHelp       string
	UserDelResultHelp        string
	UserGenCommandHelp       string
	UserGenResultHelp        string
	UserAddOthersCommandHelp string
}

func (m *ProcessorManager) showHelp(err error) *string {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	args := &UserHelp{
		Error:                    errString,
		UserAddCommandHelp:       UserAddCommandHelp,
		UserAddResultHelp:        UserAddResultHelp,
		UserDelCommandHelp:       UserDelCommandHelp,
		UserDelResultHelp:        UserDelResultHelp,
		UserGenCommandHelp:       UserGenCommandHelp,
		UserGenResultHelp:        UserGenResultHelp,
		UserAddOthersCommandHelp: UserAddOthersCommandHelp,
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
	t := &database.Transaction{
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
		"```\nCreated Transaction ID: %d\n%s```",
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
		"```\nDeleted Transaction ID: %d\n%s```",
		chunk.ID,
		deletedTx,
	)
	return &returnString
}

func (m *ProcessorManager) processChunkGenerate(chunk *Chunk) *string {
	endDate := chunk.StartDate.AddDate(0, oneMonth, 0)

	// Generate TypeOwn (Others field)
	othersResultChannel := make(chan database.AsyncAggregateResult)
	go m.db.QueryTypeOwnSum(*chunk.StartDate, endDate, othersResultChannel)

	// Generate Reimbursement (reimbursement field, will be negative)
	reimResultChannel := make(chan database.AsyncAggregateResult)
	go m.db.QueryReimburseSum(*chunk.StartDate, endDate, reimResultChannel)

	// Generate shared expenses (list of transactions)
	sharedResultChannel := make(chan database.AsyncTransactionResults)
	go m.db.QuerySharedTransactions(*chunk.StartDate, endDate, sharedResultChannel)

	// Generate all misc expenses
	miscResultChannel := make(chan database.AsyncTransactionResults)
	go m.db.QueryMiscTransactions(*chunk.StartDate, endDate, miscResultChannel)

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

	miscResult := <-miscResultChannel
	if miscResult.Error != nil {
		err := fmt.Sprintf(databaseQueryErrorFormat, miscResult.Error)
		return &err
	}

	// build the string
	var resStrings []string

	// others + reimbursements
	endDateOfMonth := endDate.AddDate(0, 0, -1).Format(time.DateOnly)
	resStrings = append(resStrings, "```")
	resStrings = append(resStrings, "---expenses.csv---")
	resStrings = append(
		resStrings,
		fmt.Sprintf("%s,Others,%.2f", endDateOfMonth, *othersResult.Result),
	)
	resStrings = append(
		resStrings,
		fmt.Sprintf("%s,Reimbursement,%.2f", endDateOfMonth, *reimResult.Result*-1),
	)

	// misc results
	for _, result := range miscResult.Result {
		res := fmt.Sprintf(
			"%s,%s,%.2f",
			endDateOfMonth,
			strings.Title(strings.ToLower(string(result.Type))),
			result.Amount,
		)
		resStrings = append(resStrings, res)
	}

	// other shared spending
	resStrings = append(resStrings, "---shared_expenses.csv---")
	for _, tx := range sharedResult.Result {
		type_ := ""
		if strings.Contains(string(tx.Type), "SPECIAL") {
			type_ += "Special:"
		}
		type_ += string(tx.Classification)

		date := tx.Date
		amount := tx.Amount
		resStrings = append(
			resStrings,
			fmt.Sprintf("%s,%s,%.2f", date.Format(time.DateOnly), type_, amount),
		)
	}

	resStrings = append(resStrings, "```")
	res := strings.Join(resStrings, "\n")

	return &res
}
