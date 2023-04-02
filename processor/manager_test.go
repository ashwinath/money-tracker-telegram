package processor

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	database "github.com/ashwinath/money-tracker-telegram/db"
	"github.com/stretchr/testify/assert"
)

func TestHelp(t *testing.T) {
	var tests = []struct {
		name           string
		err            error
		expectedLength int
	}{
		{
			name:           "help with no error",
			err:            nil,
			expectedLength: 278,
		},
		{
			name:           "help with no error",
			err:            errors.New("hello world"),
			expectedLength: 291,
		},
	}
	m, err := NewManager(nil)
	assert.Nil(t, err)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errorString := m.showHelp(tt.err)
			actualLength := len(*errorString)
			assert.Equal(t, tt.expectedLength, actualLength)
		})
	}
}

func TestProcessChunk(t *testing.T) {
	var idPattern = regexp.MustCompile(`^Created Transaction ID: (?P<txID>\d+)$`)

	err := database.RunTest(func(db *database.DB) {
		m, err := NewManager(db)
		assert.Nil(t, err)
		addChunk := &Chunk{
			Instruction:    Add,
			Type:           database.TypeOwn,
			Classification: "hello",
			Amount:         100.2,
		}
		reply := m.processChunk(addChunk, time.Now())
		assert.True(t, strings.HasPrefix(*reply, "Created Transaction ID:"))

		s := strings.Split(*reply, "\n")
		match := idPattern.FindStringSubmatch(s[0])
		result := make(map[string]string)
		for i, name := range idPattern.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		id, err := strconv.Atoi(result["txID"])
		assert.Nil(t, err)

		deleteChunk := &Chunk{
			Instruction: Delete,
			ID:          uint(id),
		}
		deleteReply := m.processChunk(deleteChunk, time.Now())
		assert.True(t, strings.HasPrefix(*deleteReply, "Deleted Transaction ID:"))
	})
	assert.Nil(t, err)
}
