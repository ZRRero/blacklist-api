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

func (receiver *BlacklistDynamoClient) GetRecordBatch(requests []*blacklist.BlacklistRecordDto) ([]*models.Record, error) {
	if len(requests) > 25 {
		return nil, errors.New("ids list has more than BlacklistDynamoClient max batch (25)")
	}

}

func (receiver *BlacklistDynamoClient) getBatchRequestFromRequest(requests []*blacklist.BlacklistRecordDto) map[string]*dynamodb.KeysAndAttributes {
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
