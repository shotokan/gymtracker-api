package dynamo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shotokan/gymtracker-api/internal/session/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockDynamoClient struct{ mock.Mock }

func (m *mockDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestSessionRepository_Save_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	session := &domain.Session{
		ID:           "session-123",
		UserID:       "user-1",
		PlanID:       "plan-1",
		RoutineID:    "routine-1",
		RoutineName:  "Push Day",
		Date:         "2024-01-15",
		DurationSecs: 3600,
		Exercises: []domain.SessionExercise{
			{
				ExerciseID: "ex-1",
				Name:       "Press Banca",
				Sets: []domain.CompletedSet{
					{Weight: 60.0, Reps: 10, RestSeconds: 90},
				},
			},
		},
		StartedAt:  time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
		FinishedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		if *input.TableName != "gym-tracker" {
			return false
		}
		pk, _ := input.Item["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Item["SK"].(*types.AttributeValueMemberS)
		gsi1pk, _ := input.Item["GSI1PK"].(*types.AttributeValueMemberS)
		gsi1sk, _ := input.Item["GSI1SK"].(*types.AttributeValueMemberS)

		return pk != nil && pk.Value == "USER#user-1" &&
			sk != nil && sk.Value == "SESSION#2024-01-15#session-123" &&
			gsi1pk != nil && gsi1pk.Value == "SESSION#session-123" &&
			gsi1sk != nil && gsi1sk.Value == "USER#user-1"
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Save(context.Background(), session)

	require.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSessionRepository_Save_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	now := time.Now().UTC()
	session := &domain.Session{
		ID:           "session-123",
		UserID:       "user-1",
		RoutineID:    "routine-1",
		RoutineName:  "Workout",
		Date:         "2024-01-15",
		DurationSecs: 1800,
		Exercises:    []domain.SessionExercise{},
		StartedAt:    now,
		FinishedAt:   now,
		CreatedAt:    now,
	}

	client.On("PutItem", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	err := repo.Save(context.Background(), session)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo put session")
	client.AssertExpectations(t)
}

func TestSessionRepository_FindByID_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	record := sessionRec{
		PK:           "USER#user-1",
		SK:           "SESSION#2024-01-15#session-123",
		GSI1PK:       "SESSION#session-123",
		GSI1SK:       "USER#user-1",
		ID:           "session-123",
		UserID:       "user-1",
		PlanID:       "plan-1",
		RoutineID:    "routine-1",
		RoutineName:  "Push Day",
		Date:         "2024-01-15",
		DurationSecs: 3600,
		Exercises: []exerciseRec{
			{
				ExerciseID: "ex-1",
				Name:       "Press",
				Sets: []setRec{
					{Weight: 60.0, Reps: 10, RestSeconds: 90},
				},
			},
		},
		StartedAt:  "2024-01-15T09:00:00Z",
		FinishedAt: "2024-01-15T10:00:00Z",
		CreatedAt:  "2024-01-15T10:30:00Z",
	}

	item, _ := attributevalue.MarshalMap(record)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "gym-tracker" &&
			*input.IndexName == "GSI1" &&
			*input.KeyConditionExpression == "GSI1PK = :pk AND GSI1SK = :sk"
	})).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}}, nil)

	result, err := repo.FindByID(context.Background(), "user-1", "session-123")

	require.NoError(t, err)
	assert.Equal(t, "session-123", result.ID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "Push Day", result.RoutineName)
	assert.Equal(t, "2024-01-15", result.Date)
	assert.Len(t, result.Exercises, 1)
	client.AssertExpectations(t)
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	_, err := repo.FindByID(context.Background(), "user-1", "session-999")

	require.ErrorIs(t, err, domain.ErrSessionNotFound)
	client.AssertExpectations(t)
}

func TestSessionRepository_FindByID_QueryError(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	_, err := repo.FindByID(context.Background(), "user-1", "session-123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo query session")
	client.AssertExpectations(t)
}

func TestSessionRepository_FindAll_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	record1 := sessionRec{
		PK:           "USER#user-1",
		SK:           "SESSION#2024-01-15#session-1",
		GSI1PK:       "USER#user-1",
		GSI1SK:       "SESSION#2024-01-15#session-1",
		ID:           "session-1",
		UserID:       "user-1",
		RoutineID:    "routine-1",
		RoutineName:  "Push Day",
		Date:         "2024-01-15",
		DurationSecs: 3600,
		Exercises:    []exerciseRec{{ExerciseID: "ex-1", Name: "Press", Sets: []setRec{{Weight: 60, Reps: 10, RestSeconds: 0}}}},
		CreatedAt:    "2024-01-15T10:00:00Z",
	}

	record2 := sessionRec{
		PK:           "USER#user-1",
		SK:           "SESSION#2024-01-14#session-2",
		GSI1PK:       "USER#user-1",
		GSI1SK:       "SESSION#2024-01-14#session-2",
		ID:           "session-2",
		UserID:       "user-1",
		RoutineID:    "routine-2",
		RoutineName:  "Pull Day",
		Date:         "2024-01-14",
		DurationSecs: 3000,
		Exercises:    []exerciseRec{{ExerciseID: "ex-2", Name: "Deadlift", Sets: []setRec{{Weight: 100, Reps: 5, RestSeconds: 0}}}},
		CreatedAt:    "2024-01-14T10:00:00Z",
	}

	item1, _ := attributevalue.MarshalMap(record1)
	item2, _ := attributevalue.MarshalMap(record2)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "gym-tracker" &&
			*input.KeyConditionExpression == "PK = :pk AND begins_with(SK, :prefix)" &&
			*input.ScanIndexForward == false && // newest first
			*input.Limit == 20
	})).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item1, item2}}, nil)

	result, err := repo.FindAll(context.Background(), "user-1", 20)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "session-1", result[0].ID)
	assert.Equal(t, "session-2", result[1].ID)
	client.AssertExpectations(t)
}

func TestSessionRepository_FindAll_Empty(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	result, err := repo.FindAll(context.Background(), "user-1", 20)

	require.NoError(t, err)
	assert.Empty(t, result)
	client.AssertExpectations(t)
}

func TestSessionRepository_FindAll_QueryError(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewSessionRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	_, err := repo.FindAll(context.Background(), "user-1", 20)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo query sessions")
	client.AssertExpectations(t)
}

func TestToRecord_KeyConstruction(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	session := &domain.Session{
		ID:           "session-abc",
		UserID:       "user-xyz",
		Date:         "2024-01-20",
		RoutineID:    "routine-1",
		RoutineName:  "Workout",
		DurationSecs: 1800,
		Exercises:    []domain.SessionExercise{},
		StartedAt:    now,
		FinishedAt:   now,
		CreatedAt:    now,
	}

	record := toRecord(session)

	assert.Equal(t, "USER#user-xyz", record.PK)
	assert.Equal(t, "SESSION#2024-01-20#session-abc", record.SK)
	assert.Equal(t, "SESSION#session-abc", record.GSI1PK)
	assert.Equal(t, "USER#user-xyz", record.GSI1SK)
}
