package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/database"
	"jaystar/internal/model/po"
	"log"
	"os"
	"testing"
	"time"
)

func TestGetRecordsWithSettleRecords(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	iConfigEnv := config.ProviderIConfigEnv()
	db := database.ProvidePostgresDB(iConfigEnv)
	repo := ProvideReduceRecordRepository(iConfigEnv)

	records, err := repo.GetRecordsWithSettleRecords(context.TODO(), db.Session(), &po.ReduceRecordCond{
		StudentId:      197763843382837253,
		ClassTimeStart: time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local),
		ClassTimeEnd:   time.Date(2025, 3, 31, 0, 0, 0, 0, time.Local),
	}, &po.Pager{
		Index: 1,
		Size:  20,
		Order: "class_time desc",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("records: %v", records)
}

func TestGetRecordsWithSettleRecordsPager(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	iConfigEnv := config.ProviderIConfigEnv()
	db := database.ProvidePostgresDB(iConfigEnv)
	repo := ProvideReduceRecordRepository(iConfigEnv)

	count, err := repo.GetRecordsWithSettleRecordsPager(context.TODO(), db.Session(), &po.ReduceRecordCond{
		StudentId:      197763843382837253,
		ClassTimeStart: time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local),
		ClassTimeEnd:   time.Date(2025, 3, 31, 0, 0, 0, 0, time.Local),
	}, &po.Pager{
		Index: 1,
		Size:  20,
		Order: "class_time desc",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("count: %v", count)
}
