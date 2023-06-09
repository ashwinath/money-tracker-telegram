package webhandler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	database "github.com/ashwinath/money-tracker-telegram/db"
	"github.com/ashwinath/money-tracker-telegram/utils"
)

const oneMonth = 1
const databaseQueryErrorFormat = "Could not query database, error: %s"

type DataDumpHandler struct {
	db *database.DB
}

func NewDataDumpHandler(db *database.DB) *DataDumpHandler {
	return &DataDumpHandler{
		db: db,
	}
}

type expensesResponse struct {
	Expenses []expenseStruct `json:"expenses"`
}

type expenseStruct struct {
	Date   time.Time `json:"date"`
	Type   string    `json:"type"`
	Amount float64   `json:"amount"`
}

// all numerical
// /expenses?month=<?>&year=<?>
func (h *DataDumpHandler) expenses(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := getDatesToProcess(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	// Generate TypeOwn (Others field)
	othersResultChannel := make(chan database.AsyncAggregateResult)
	go h.db.QueryTypeOwnSum(*startDate, *endDate, othersResultChannel)

	// Generate Reimbursement (reimbursement field, will be negative)
	reimResultChannel := make(chan database.AsyncAggregateResult)
	go h.db.QueryReimburseSum(*startDate, *endDate, reimResultChannel)

	// Generate shared expenses (list of transactions)
	miscResultChannel := make(chan database.AsyncTransactionResults)
	go h.db.QueryMiscTransactions(*startDate, *endDate, miscResultChannel)

	othersResult := <-othersResultChannel
	if othersResult.Error != nil {
		err := fmt.Errorf(databaseQueryErrorFormat, othersResult.Error)
		badRequest(w, err)
		return
	}

	reimResult := <-reimResultChannel
	if reimResult.Error != nil {
		err := fmt.Errorf(databaseQueryErrorFormat, reimResult.Error)
		badRequest(w, err)
		return
	}

	miscResult := <-miscResultChannel
	if miscResult.Error != nil {
		err := fmt.Errorf(databaseQueryErrorFormat, miscResult.Error)
		badRequest(w, err)
		return
	}

	var expenses []expenseStruct
	expenses = append(expenses, expenseStruct{
		Date:   utils.SetDateToEndOfMonth(*startDate),
		Type:   "Others",
		Amount: *othersResult.Result,
	})
	expenses = append(expenses, expenseStruct{
		Date:   utils.SetDateToEndOfMonth(*startDate),
		Type:   "Reimbursement",
		Amount: *reimResult.Result * -1,
	})
	for _, result := range miscResult.Result {
		expenses = append(expenses, expenseStruct{
			Date:   utils.SetDateToEndOfMonth(*startDate),
			Type:   strings.Title(strings.ToLower(string(result.Type))),
			Amount: result.Amount,
		})
	}

	ok(w, expensesResponse{expenses})
}

// all numerical
// /shared-expenses?month=<?>&year=<?>
func (h *DataDumpHandler) sharedExpenses(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := getDatesToProcess(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	sharedResultChannel := make(chan database.AsyncTransactionResults)
	go h.db.QuerySharedTransactions(*startDate, *endDate, sharedResultChannel)

	sharedReimCCResultChannel := make(chan database.AsyncTransactionResults)
	go h.db.QuerySharedReimCCTransactions(*startDate, *endDate, sharedReimCCResultChannel)

	sharedResult := <-sharedResultChannel
	if sharedResult.Error != nil {
		err := fmt.Errorf(databaseQueryErrorFormat, sharedResult.Error)
		badRequest(w, err)
		return
	}

	sharedReimCCResult := <-sharedReimCCResultChannel
	if sharedReimCCResult.Error != nil {
		err := fmt.Errorf(databaseQueryErrorFormat, sharedReimCCResult.Error)
		badRequest(w, err)
		return
	}

	nonSpecialSpend := 0.0
	var otherSpendingDate time.Time

	var sharedExpenses []expenseStruct
	for _, tx := range sharedResult.Result {
		if strings.Contains(string(tx.Type), "SPECIAL") {
			type_ := fmt.Sprintf("Special:%s", string(tx.Classification))
			sharedExpenses = append(sharedExpenses, expenseStruct{
				Date:   utils.SetDateToEndOfMonth(tx.Date),
				Type:   type_,
				Amount: tx.Amount,
			})
			continue
		}

		// Combine all non special spends
		nonSpecialSpend += tx.Amount
		otherSpendingDate = utils.SetDateToEndOfMonth(tx.Date)
	}

	// subtract shared reim cc result
	sharedCCReimAmount := 0.0
	for _, tx := range sharedReimCCResult.Result {
		sharedCCReimAmount += tx.Amount
	}

	sharedExpenses = append(sharedExpenses, expenseStruct{
		Date:   otherSpendingDate,
		Type:   "others",
		Amount: nonSpecialSpend - sharedCCReimAmount,
	})

	ok(w, expensesResponse{sharedExpenses})
}

func (h *DataDumpHandler) Serve(port int) {
	log.Printf("Starting http server on port: %d", port)
	http.Handle("/health", http.HandlerFunc(health))
	http.Handle("/expenses", http.HandlerFunc(h.expenses))
	http.Handle("/shared-expenses", http.HandlerFunc(h.sharedExpenses))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getDatesToProcess(r *http.Request) (*time.Time, *time.Time, error) {
	month := getArg(r, "month")
	if month == "" {
		return nil, nil, fmt.Errorf("missing parameter: month")
	}

	year := getArg(r, "year")
	if year == "" {
		return nil, nil, fmt.Errorf("missing parameter: year")
	}

	loc, _ := time.LoadLocation("Asia/Singapore")

	startDate, err := time.ParseInLocation(time.DateOnly, fmt.Sprintf("%s-%s-01", year, month), loc)
	if err != nil {
		return nil, nil, err
	}

	endDate := startDate.AddDate(0, oneMonth, 0)

	return &startDate, &endDate, nil
}
