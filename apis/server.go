package apis

import (
	"blacklist-api/models"
	"blacklist-api/pkg/clients"
	blacklist "blacklist-api/tools/protos"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	notFound = "given record %s with client %s and product %s does not exist"
)

type BlacklistServer struct {
	blacklist.UnimplementedBlacklistServer
	mu        sync.Mutex
	BatchSize int
	Table     string
}

func (receiver *BlacklistServer) GetBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordRequest) (*blacklist.BlacklistRecordDto, error) {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return nil, err
	}
	result, err := client.GetRecordByRequest(request)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New(fmt.Sprintf(notFound, request.RecordId, request.ClientId, request.ProductId))
	}
	return result.ToDto(), nil
}

func (receiver *BlacklistServer) GetBlacklistRecordBatch(stream blacklist.Blacklist_GetBlacklistRecordBatchServer) error {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return err
	}
	requests := make([]*blacklist.BlacklistRecordRequest, 0, receiver.BatchSize)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		requests = append(requests, in)
		if len(requests) == 25 {
			result, err := client.GetRecordBatch(requests)
			if err != nil {
				return err
			}
			for _, record := range result {
				err = stream.Send(record.ToDto())
				if err != nil {
					return err
				}
			}
		}
	}
}

func (receiver *BlacklistServer) SaveBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordDto) (*blacklist.BlacklistRecordDto, error) {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return nil, err
	}
	record, err := client.SaveRecord(models.NewRecord(request.RecordId, request.ClientId, request.ProductId))
	if err != nil {
		return nil, err
	}
	return record.ToDto(), nil
}

func (receiver *BlacklistServer) SaveBlacklistRecordBatch(stream blacklist.Blacklist_SaveBlacklistRecordBatchServer) error {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return err
	}
	requests := make([]*models.Record, 0, receiver.BatchSize)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		record, err := models.FromDto(in)
		if err != nil {
			return err
		}
		requests = append(requests, record)
		if len(requests) == 25 {
			result, err := client.SaveBatchRecords(requests)
			if err != nil {
				return err
			}
			for _, record := range result {
				err = stream.Send(record.ToDto())
				if err != nil {
					return err
				}
			}
		}
	}
}

func (receiver *BlacklistServer) DeleteBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordRequest) (*blacklist.Empty, error) {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return nil, err
	}
	err = client.DeleteRecord(request)
	if err != nil {
		return nil, err
	}
	return &blacklist.Empty{}, nil
}

func (receiver *BlacklistServer) DeleteBatchBlacklistRecord(stream blacklist.Blacklist_DeleteBatchBlacklistRecordServer) error {
	client, err := clients.NewClient(receiver.Table)
	if err != nil {
		return err
	}
	requests := make([]*blacklist.BlacklistRecordRequest, 0, receiver.BatchSize)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		requests = append(requests, in)
		if len(requests) == 25 {
			err = client.DeleteBatchRecords(requests)
			if err != nil {
				return err
			}
		}
	}
}

func (receiver *BlacklistServer) HealthCheck(_ context.Context, request *blacklist.Empty) (*blacklist.Empty, error) {
	return request, nil
}
