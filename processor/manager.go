package processor

import (
	"bytes"
	"fmt"
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

const UserAddCommandHelp = "`ADD <TYPE> <CLASSIFICATION> <PRICE (no $ sign)>`"
const UserAddResultHelp = "`Created Transaction ID: <ID>, Transaction: <transaction>`"
const UserDelCommandHelp = "`DEL <ID>`"
const UserDelResultHelp = "`Deleted Transaction ID: <ID>, Transaction: <transaction>`"

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
		"Deleted Transaction ID: %d, Transaction: %+v",
		chunk.ID,
		deletedTx,
	)
	return &returnString
}
