package email

import (
	"context"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUntrash(t *testing.T) {
	tests := []struct {
		client      func(t *testing.T) UpdateItemAPI
		messageID   string
		expectedErr error
	}{
		{
			client: func(t *testing.T) UpdateItemAPI {
				return mockUpdateItemAPI(func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					assert.Len(t, params.Key, 1)
					assert.IsType(t, params.Key["MessageID"], &types.AttributeValueMemberS{})
					assert.Equal(t,
						params.Key["MessageID"].(*types.AttributeValueMemberS).Value,
						"exampleMessageID",
					)
					assert.Equal(t, "REMOVE TrashedTime", *params.UpdateExpression)
					assert.Equal(t, "attribute_exists(TrashedTime) AND NOT begins_with(TypeYearMonth, :v_type)",
						*params.ConditionExpression)

					return &dynamodb.UpdateItemOutput{}, nil
				})
			},
			messageID: "exampleMessageID",
		},
		{
			client: func(t *testing.T) UpdateItemAPI {
				return mockUpdateItemAPI(func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					return &dynamodb.UpdateItemOutput{}, errors.Wrap(&types.ConditionalCheckFailedException{}, "")
				})
			},
			messageID:   "",
			expectedErr: ErrNotTrashed,
		},
		{
			client: func(t *testing.T) UpdateItemAPI {
				return mockUpdateItemAPI(func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
					t.Helper()
					return &dynamodb.UpdateItemOutput{}, ErrNotFound
				})
			},
			messageID:   "",
			expectedErr: ErrNotFound,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := context.TODO()
			err := Untrash(ctx, test.client(t), test.messageID)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
