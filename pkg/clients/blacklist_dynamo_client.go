package clients

import (
	"blacklist-api/models"
	blacklist "blacklist-api/tools/protos"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type BlacklistDynamoClient struct {
	client dynamodbiface.DynamoDBAPI
	table  string
}

func newSession() (*session.Session, error) {
	sess, err := session.NewSession()
	svc := session.Must(sess, err)
	return svc, err
}

func NewClient(table string) (*BlacklistDynamoClient, error) {
	// Create AWS Session
	sess, err := newSession()
	if err != nil {
		return nil, err
	}
	dynamoClient := &BlacklistDynamoClient{dynamodb.New(sess), table}
	return dynamoClient, nil
}

//Get

func (receiver *BlacklistDynamoClient) GetRecordByRequest(request *blacklist.BlacklistRecordRequest) (*models.Record, error) {
	key := make(map[string]*dynamodb.AttributeValue)
	sortKey := fmt.Sprintf("%s:%s", request.ClientId, request.ProductId)
	key["record_id"] = &dynamodb.AttributeValue{S: &request.RecordId}
	key["sort_id"] = &dynamodb.AttributeValue{S: &sortKey}
	input := &dynamodb.GetItemInput{
		TableName: &receiver.table,
		Key:       key,
	}
	result, err := receiver.client.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	record, err := models.FromDynamoItem(result.Item)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (receiver *BlacklistDynamoClient) GetRecordBatch(requests []*blacklist.BlacklistRecordRequest) ([]*models.Record, error) {
	if len(requests) > 25 {
		return nil, errors.New("ids list has more than BlacklistDynamoClient max batch (25)")
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: receiver.getBatchRequestFromRequest(requests),
	}
	result, err := receiver.client.BatchGetItem(input)
	if err != nil {
		return nil, err
	}
	dynamoRecords := result.Responses[receiver.table]
	records, err := receiver.parseDynamoRecords(dynamoRecords)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (receiver *BlacklistDynamoClient) getBatchRequestFromRequest(requests []*blacklist.BlacklistRecordRequest) map[string]*dynamodb.KeysAndAttributes {
	items := make([]map[string]*dynamodb.AttributeValue, 0, 25)
	for _, request := range requests {
		item := make(map[string]*dynamodb.AttributeValue)
		sortKey := fmt.Sprintf("%s:%s", request.ClientId, request.ProductId)
		key := make(map[string]*dynamodb.AttributeValue)
		key["record_id"] = &dynamodb.AttributeValue{S: &request.RecordId}
		key["sort_id"] = &dynamodb.AttributeValue{S: &sortKey}
		items = append(items, item)
	}
	keyAndAttributes := &dynamodb.KeysAndAttributes{
		Keys: items,
	}
	requestItems := make(map[string]*dynamodb.KeysAndAttributes)
	requestItems[receiver.table] = keyAndAttributes
	return requestItems
}

func (receiver *BlacklistDynamoClient) parseDynamoRecords(dynamoRecords []map[string]*dynamodb.AttributeValue) ([]*models.Record, error) {
	records := make([]*models.Record, 0, 25)
	for _, dynamoRecord := range dynamoRecords {
		record, err := models.FromDynamoItem(dynamoRecord)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

//Save

func (receiver *BlacklistDynamoClient) SaveRecord(record *models.Record) (*models.Record, error) {
	input := &dynamodb.PutItemInput{
		TableName: &receiver.table,
		Item:      record.ToDynamoItem(),
	}
	_, err := receiver.client.PutItem(input)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (receiver *BlacklistDynamoClient) SaveBatchRecords(records []*models.Record) ([]*models.Record, error) {
	if len(records) > 25 {
		return nil, errors.New("ids list has more than BlacklistDynamoClient max batch (25)")
	}
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: receiver.getWriteBatchRequestFromModel(records),
	}
	result, err := receiver.client.BatchWriteItem(input)
	if err != nil {
		return nil, err
	}
	for len(result.UnprocessedItems) != 0 {
		input = &dynamodb.BatchWriteItemInput{
			RequestItems: result.UnprocessedItems,
		}
		result, err = receiver.client.BatchWriteItem(input)
		if err != nil {
			return nil, err
		}
	}
	return records, nil
}

func (receiver *BlacklistDynamoClient) getWriteBatchRequestFromModel(records []*models.Record) map[string][]*dynamodb.WriteRequest {
	items := make(map[string][]*dynamodb.WriteRequest)
	requests := make([]*dynamodb.WriteRequest, 0, len(records))
	for _, record := range records {
		requests = append(requests, &dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{Item: record.ToDynamoItem()}})
	}
	items[receiver.table] = requests
	return items
}

//Delete

func (receiver *BlacklistDynamoClient) DeleteRecord(id *string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &receiver.table,
		Key:       map[string]*dynamodb.AttributeValue{"id": {S: id}},
	}
	_, err := receiver.client.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (receiver *BlacklistDynamoClient) DeleteBatchRecords(ids []*string) error {
	if len(ids) > 25 {
		return errors.New("ids list has more than BlacklistDynamoClient max batch (25)")
	}
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: receiver.getDeleteBatchRequestFromIds(ids),
	}
	result, err := receiver.client.BatchWriteItem(input)
	if err != nil {
		return err
	}
	for len(result.UnprocessedItems) != 0 {
		input = &dynamodb.BatchWriteItemInput{
			RequestItems: result.UnprocessedItems,
		}
		result, err = receiver.client.BatchWriteItem(input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (receiver *BlacklistDynamoClient) getDeleteBatchRequestFromIds(ids []*string) map[string][]*dynamodb.WriteRequest {
	items := make(map[string][]*dynamodb.WriteRequest)
	requests := make([]*dynamodb.WriteRequest, 0, len(ids))
	for _, id := range ids {
		requests = append(requests, &dynamodb.WriteRequest{DeleteRequest: &dynamodb.DeleteRequest{Key: map[string]*dynamodb.AttributeValue{"id": {S: id}}}})
	}
	items[receiver.table] = requests
	return items
}
