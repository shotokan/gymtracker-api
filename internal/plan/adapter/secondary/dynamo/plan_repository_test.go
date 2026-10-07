package dynamo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockDynamoClient struct{ mock.Mock }

func (m *mockDynamoClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *mockDynamoClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestPlanRepository_Save_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan",
		Active:    false,
		Routines:  []domain.Routine{},
		CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		if *input.TableName != "gym-tracker" {
			return false
		}
		pk, _ := input.Item["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Item["SK"].(*types.AttributeValueMemberS)

		return pk != nil && pk.Value == "USER#user-1" &&
			sk != nil && sk.Value == "PLAN#plan-123"
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Save(context.Background(), plan)

	require.NoError(t, err)
	client.AssertExpectations(t)
}

func TestPlanRepository_Save_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	plan := &domain.Plan{
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Plan",
		Routines:  []domain.Routine{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	client.On("PutItem", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	err := repo.Save(context.Background(), plan)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo put plan")
}

func TestPlanRepository_FindByID_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	record := planRec{
		PK:        "USER#user-1",
		SK:        "PLAN#plan-123",
		ID:        "plan-123",
		UserID:    "user-1",
		Name:      "Mi Plan",
		Active:    true,
		Routines:  []routineRec{},
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-15T10:00:00Z",
	}

	item, _ := attributevalue.MarshalMap(record)

	client.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		pk, _ := input.Key["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Key["SK"].(*types.AttributeValueMemberS)
		return *input.TableName == "gym-tracker" &&
			pk.Value == "USER#user-1" &&
			sk.Value == "PLAN#plan-123"
	})).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	result, err := repo.FindByID(context.Background(), "user-1", "plan-123")

	require.NoError(t, err)
	assert.Equal(t, "plan-123", result.ID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "Mi Plan", result.Name)
	assert.True(t, result.Active)
	client.AssertExpectations(t)
}

func TestPlanRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

	_, err := repo.FindByID(context.Background(), "user-1", "plan-999")

	require.ErrorIs(t, err, domain.ErrPlanNotFound)
}

func TestPlanRepository_FindAll_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	record1 := planRec{
		PK: "USER#user-1", SK: "PLAN#plan-1", ID: "plan-1", UserID: "user-1",
		Name: "Plan A", Active: true, Routines: []routineRec{},
		CreatedAt: "2024-01-15T10:00:00Z", UpdatedAt: "2024-01-15T10:00:00Z",
	}
	record2 := planRec{
		PK: "USER#user-1", SK: "PLAN#plan-2", ID: "plan-2", UserID: "user-1",
		Name: "Plan B", Active: false, Routines: []routineRec{},
		CreatedAt: "2024-01-14T10:00:00Z", UpdatedAt: "2024-01-14T10:00:00Z",
	}

	item1, _ := attributevalue.MarshalMap(record1)
	item2, _ := attributevalue.MarshalMap(record2)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "gym-tracker" &&
			*input.KeyConditionExpression == "PK = :pk AND begins_with(SK, :prefix)"
	})).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item1, item2}}, nil)

	result, err := repo.FindAll(context.Background(), "user-1", "")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "plan-1", result[0].ID)
	assert.Equal(t, "plan-2", result[1].ID)
}

func TestPlanRepository_FindActive_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	record := planRec{
		PK: "USER#user-1", SK: "PLAN#plan-active", ID: "plan-active", UserID: "user-1",
		Name: "Active Plan", Active: true, Routines: []routineRec{},
		CreatedAt: "2024-01-15T10:00:00Z", UpdatedAt: "2024-01-15T10:00:00Z",
	}

	item, _ := attributevalue.MarshalMap(record)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.FilterExpression == "active = :true" && *input.Limit == 1
	})).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}}, nil)

	result, err := repo.FindActive(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "plan-active", result.ID)
	assert.True(t, result.Active)
}

func TestPlanRepository_FindActive_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	_, err := repo.FindActive(context.Background(), "user-1")

	require.ErrorIs(t, err, domain.ErrNoActivePlan)
}

func TestPlanRepository_Update_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	plan := &domain.Plan{
		ID: "plan-123", UserID: "user-1", Name: "Updated Plan",
		Active: true, Routines: []domain.Routine{},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}

	client.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Update(context.Background(), plan)

	require.NoError(t, err)
}

func TestPlanRepository_Delete_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewPlanRepository(client, "gym-tracker")

	client.On("DeleteItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
		pk, _ := input.Key["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Key["SK"].(*types.AttributeValueMemberS)
		return *input.TableName == "gym-tracker" &&
			pk.Value == "USER#user-1" &&
			sk.Value == "PLAN#plan-123"
	})).Return(&dynamodb.DeleteItemOutput{}, nil)

	err := repo.Delete(context.Background(), "user-1", "plan-123")

	require.NoError(t, err)
}
