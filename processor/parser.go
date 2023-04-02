package processor

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ashwinath/money-tracker-telegram/db"
)

// Only used for parser
const TypeNone db.TransactionType = "NO_OP"

type Instruction int

const (
	Add Instruction = iota
	Delete
	Help
	NoOp
)

const (
	errorFormatWrappedToUser      = "Error: %s, Message: %s"
	errorFormatInstructionToken   = "could not parse instruction token: %s"
	errorFormatTypeToken          = "unable to parse type: %s"
	errorFormatAmountToken        = "unable to parse type: %s"
	errorEmptyClassificationToken = "empty classification token"
	errorEmptyTypeToken           = "empty type token"
	errorEmptyAmountToken         = "empty amount token"
	errorEmptyInstructionToken    = "empty instruction token"
	errorEmptyIDToken             = "empty ID token"
)

type Chunk struct {
	Instruction    Instruction
	Type           db.TransactionType
	Classification string
	Amount         float64
	ID             uint
	Date           *time.Time
}

type parser struct {
	message  string
	tokens   []string
	position int
	length   int
}

// Parse parses the command sent by the user
func Parse(message string) (*Chunk, error) {
	tokens := strings.Fields(message)
	parserObj := parser{
		message:  message,
		tokens:   tokens,
		position: 0,
		length:   len(tokens),
	}

	return parserObj.Parse()
}

func (p *parser) Parse() (*Chunk, error) {
	instruction, err := p.instruction()
	if err != nil {
		return nil, p.wrapError(err)
	}

	chunk := &Chunk{
		Instruction: instruction,
	}

	switch instruction {
	case Add:
		transactionType, err := p.transactionType()
		if err != nil {
			return nil, p.wrapError(err)
		}
		chunk.Type = transactionType

		classification, err := p.classification()
		if err != nil {
			return nil, p.wrapError(err)
		}
		chunk.Classification = *classification

		amount, err := p.amount()
		if err != nil {
			return nil, p.wrapError(err)
		}
		chunk.Amount = *amount

		date, err := p.date()
		if err != nil {
			return nil, p.wrapError(err)
		}
		if date != nil {
			// Date is optional
			chunk.Date = date
		}

	case Delete:
		id, err := p.id()
		if err != nil {
			return nil, p.wrapError(err)
		}
		chunk.ID = *id
	case Help:
		// Do nothing
	}

	return chunk, nil
}

func (p *parser) wrapError(err error) error {
	return fmt.Errorf(errorFormatWrappedToUser, err, p.message)
}

func (p *parser) incrementPos() {
	p.position++
}

func (p *parser) peekCurrent() *string {
	if (p.position) >= p.length {
		return nil
	}

	return &p.tokens[p.position]
}

func (p *parser) next() *string {
	if (p.position) >= p.length {
		return nil
	}

	token := p.tokens[p.position]
	p.incrementPos()

	return &token
}

func (p *parser) instruction() (Instruction, error) {
	token := p.next()
	if token == nil {
		return NoOp, errors.New(errorEmptyInstructionToken)
	}

	switch tokenUpper := strings.ToUpper(*token); tokenUpper {
	case "ADD":
		return Add, nil
	case "DEL":
		return Delete, nil
	case "HELP":
		return Help, nil
	default:
		return NoOp, fmt.Errorf(errorFormatInstructionToken, *token)
	}
}

func (p *parser) transactionType() (db.TransactionType, error) {
	token := p.next()
	if token == nil {
		return TypeNone, errors.New(errorEmptyTypeToken)
	}

	switch tokenUpper := strings.ToUpper(*token); tokenUpper {
	case "SHARED":
		if strings.ToUpper(*p.peekCurrent()) == "REIM" {
			p.next()
			return db.TypeSharedReimburse, nil
		}
		return db.TypeShared, nil
	case "SPECIAL":
		if strings.ToUpper(*p.peekCurrent()) == "SHARED" {
			p.next()
			if strings.ToUpper(*p.peekCurrent()) == "REIM" {
				p.next()
				return db.TypeSpecialSharedReimburse, nil
			}
			return db.TypeSpecialShared, nil
		} else if strings.ToUpper(*p.peekCurrent()) == "OWN" {
			p.next()
			return db.TypeSpecialOwn, nil
		}
		return TypeNone, fmt.Errorf(errorFormatTypeToken, fmt.Sprintf("%s %s", *token, *p.next()))
	case "REIM":
		return db.TypeReimburse, nil
	case "OWN":
		return db.TypeOwn, nil
	default:
		return TypeNone, fmt.Errorf(errorFormatTypeToken, *token)
	}
}

func (p *parser) classification() (*string, error) {
	classificationSlice := []string{}

	if p.peekCurrent() == nil {
		return nil, errors.New(errorEmptyClassificationToken)
	}

	for {
		// Not consuming the token as of yet in case we don't want to use it.
		if p.peekCurrent() == nil {
			// This is empty amount, e.g. Add own bicycle.
			// Let parser.amount() handle the error.
			break
		}

		if _, err := strconv.ParseFloat(*p.peekCurrent(), 64); err == nil {
			// Found a number
			break
		}

		classificationSlice = append(classificationSlice, *p.next())
	}

	// next token could be a number
	if len(classificationSlice) == 0 {
		return nil, errors.New(errorEmptyClassificationToken)
	}

	classification := strings.Join(classificationSlice, " ")

	return &classification, nil
}

func (p *parser) amount() (*float64, error) {
	token := p.next()
	if token == nil {
		// EOF error
		return nil, errors.New(errorEmptyAmountToken)
	}

	amount, err := strconv.ParseFloat(*token, 64)
	if err != nil {
		// This cannot happen, or I can't imagine it.
		// parser.classification() should consume all non floats
		return nil, errors.New(fmt.Sprintf(errorFormatAmountToken, *token))
	}

	return &amount, nil
}

func (p *parser) id() (*uint, error) {
	token := p.next()
	if token == nil {
		// EOF error
		return nil, errors.New(errorEmptyIDToken)
	}

	idInt, err := strconv.Atoi(*token)
	if err != nil {
		return nil, err
	}

	id := uint(idInt)

	return &id, nil
}

func (p *parser) date() (*time.Time, error) {
	token := p.next()
	if token == nil {
		// Date is optional
		return nil, nil
	}

	loc, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		return nil, err
	}

	parsed, err := time.ParseInLocation(time.DateOnly, *token, loc)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
