package apis

import (
	blacklist "blacklist-api/tools/protos"
	"context"
	"sync"
)

type BlacklistServer struct {
	blacklist.UnimplementedBlacklistServer
	mu        sync.Mutex
	BatchSize int
	Table     string
}

func (receiver *BlacklistServer) GetBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordRequest) (*blacklist.BlacklistRecordDto, error) {
	return nil, nil
}

func (receiver *BlacklistServer) GetBlacklistRecordBatch(stream blacklist.Blacklist_GetBlacklistRecordBatchServer) error {
	return nil
}

func (receiver *BlacklistServer) SaveBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordDto) (*blacklist.BlacklistRecordDto, error) {
	return nil, nil
}

func (receiver *BlacklistServer) SaveBlacklistRecordBatch(stream blacklist.Blacklist_SaveBlacklistRecordBatchServer) error {
	return nil
}

func (receiver *BlacklistServer) DeleteBlacklistRecord(_ context.Context, request *blacklist.BlacklistRecordRequest) (*blacklist.Empty, error) {
	return nil, nil
}

func (receiver *BlacklistServer) DeleteBatchBlacklistRecord(stream blacklist.Blacklist_DeleteBatchBlacklistRecordServer) error {
	return nil
}
