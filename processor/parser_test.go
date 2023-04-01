package processor

import (
	"errors"
	"testing"

	"github.com/ashwinath/money-tracker-telegram/db"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	var tests = []struct {
		name          string
		testString    string
		expected      Chunk
		expectedError error
	}{
		{
			name:       "Add reim",
			testString: "Add reim food 24.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeReimburse,
				Classification: "food",
				Amount:         24.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add shared reim",
			testString: "Add shared reim christmas's dinner 60.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeSharedReimburse,
				Classification: "christmas's dinner",
				Amount:         60.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add special shared reim",
			testString: "Add special shared reim furniture 2400.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeSpecialSharedReimburse,
				Classification: "furniture",
				Amount:         2400.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add shared",
			testString: "Add shared lunch 10.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeShared,
				Classification: "lunch",
				Amount:         10.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add special shared",
			testString: "Add special shared washing machine 810.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeSpecialShared,
				Classification: "washing machine",
				Amount:         810.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add special own",
			testString: "Add special own holiday to japan 810.5",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeSpecialOwn,
				Classification: "holiday to japan",
				Amount:         810.5,
			},
			expectedError: nil,
		},
		{
			name:       "Add own",
			testString: "Add own computer and monitors 2010",
			expected: Chunk{
				Instruction:    Add,
				Type:           db.TypeOwn,
				Classification: "computer and monitors",
				Amount:         2010.0,
			},
			expectedError: nil,
		},
		{
			name:       "Delete transaction",
			testString: "Del 100",
			expected: Chunk{
				Instruction: Delete,
				ID:          100,
			},
			expectedError: nil,
		},
		{
			name:          "wrong instruction",
			testString:    "update own computer and monitors 2010",
			expected:      Chunk{},
			expectedError: errors.New("Error: could not parse instruction token: update, Message: update own computer and monitors 2010"),
		},
		{
			name:          "wrong Add when should have been Del",
			testString:    "Add 100",
			expected:      Chunk{},
			expectedError: errors.New("Error: unable to parse type: 100, Message: Add 100"),
		},
		{
			name:          "wrong special but shared or own does not come after",
			testString:    "Add special something toy 100.4",
			expected:      Chunk{},
			expectedError: errors.New("Error: unable to parse type: special something, Message: Add special something toy 100.4"),
		},
		{
			name:          "wrong type",
			testString:    "Add property sale 1000000.4",
			expected:      Chunk{},
			expectedError: errors.New("Error: unable to parse type: property, Message: Add property sale 1000000.4"),
		},
		{
			name:          "classification is a number",
			testString:    "Add reim 100.4",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty classification token, Message: Add reim 100.4"),
		},
		{
			name:          "empty classification",
			testString:    "Add reim",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty classification token, Message: Add reim"),
		},
		{
			name:          "empty amount",
			testString:    "Add own bicycle",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty amount token, Message: Add own bicycle"),
		},
		{
			name:          "empty string",
			testString:    "",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty instruction token, Message: "),
		},
		{
			name:          "empty transaction type",
			testString:    "Add",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty type token, Message: Add"),
		},
		{
			name:          "empty id for delete",
			testString:    "Del",
			expected:      Chunk{},
			expectedError: errors.New("Error: empty ID token, Message: Del"),
		},
		{
			name:          "delete cannot parse id",
			testString:    "Del you",
			expected:      Chunk{},
			expectedError: errors.New("Error: strconv.Atoi: parsing \"you\": invalid syntax, Message: Del you"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Parse(tt.testString)
			if tt.expectedError == nil {
				assert.Equal(t, tt.expected, *actual)
			} else {
				assert.Equal(t, tt.expectedError, err)
			}
		})
	}
}