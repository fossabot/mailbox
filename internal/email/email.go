package email

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/harryzcy/mailbox/internal/util/format"
)

var (
	// tableName represents DynamoDB table name
	tableName = os.Getenv("DYNAMODB_TABLE")
	// gsiIndexName represents DynamoDB's GSI name
	gsiIndexName = os.Getenv("DYNAMODB_TIME_INDEX")
)

// The constants representing email types
const (
	// EmailTypeInbox represents an inbox email
	EmailTypeInbox = "inbox"
	// EmailTypeInbox represents a sent email
	EmailTypeSent = "sent"
	// EmailTypeInbox represents a draft email
	EmailTypeDraft = "draft"
)

// TimeIndex represents the index attributes of an email
type TimeIndex struct {
	MessageID string `json:"messageID"`
	Type      string `json:"type"`

	// TimeReceived is used by inbox emails
	TimeReceived string `json:"timeReceived,omitempty"`

	// TimeUpdated is used by draft emails
	TimeUpdated string `json:"timeUpdated,omitempty"`

	// TimeSent is used by sent emails
	TimeSent string `json:"timeSent,omitempty"`
}

// GSIIndex represents Global Secondary Index of an email
type GSIIndex struct {
	MessageID     string `dynamodbav:"MessageID"`
	TypeYearMonth string `dynamodbav:"TypeYearMonth"`
	DateTime      string `dynamodbav:"DateTime"`
}

// ToTimeIndex returns TimeIndex
func (gsi GSIIndex) ToTimeIndex() (*TimeIndex, error) {
	index := &TimeIndex{
		MessageID: gsi.MessageID,
	}

	var emailTime string
	var err error
	index.Type, emailTime, err = parseGSI(gsi.TypeYearMonth, gsi.DateTime)

	switch index.Type {
	case EmailTypeInbox:
		index.TimeReceived = emailTime
	case EmailTypeSent:
		index.TimeSent = emailTime
	case EmailTypeDraft:
		index.TimeUpdated = emailTime
	}
	return index, err
}

func unmarshalGSI(item map[string]types.AttributeValue) (emailType, emailTime string, err error) {
	var typeYearMonth string
	var dt string // date-time
	err = attributevalue.Unmarshal(item["TypeYearMonth"], &typeYearMonth)
	if err != nil {
		fmt.Printf("unmarshal TypeYearMonth failed: %v", err)
		return
	}
	err = attributevalue.Unmarshal(item["DateTime"], &dt)
	if err != nil {
		fmt.Printf("unmarshal DateTime failed: %v", err)
		return
	}
	return parseGSI(typeYearMonth, dt)
}

func parseGSI(typeYearMonth, dt string) (emailType, emailTime string, err error) {
	var ym string // YYYY-MM
	emailType, ym, err = format.ExtractTypeYearMonth(typeYearMonth)
	if err != nil {
		fmt.Printf("extract TypeYearMonth failed: %v\n", err)
		return
	}
	emailTime = format.RejoinDate(ym, dt)
	return
}
