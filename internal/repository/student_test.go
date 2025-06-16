package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/database"
	"jaystar/internal/model/po"
	"log"
	"os"
	"testing"
)

func TestGetStudentRefIds(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	iConfigEnv := config.ProviderIConfigEnv()
	db := database.ProvidePostgresDB(iConfigEnv)
	repo := ProvideStudentRepository()

	ids, err := repo.GetStudentRefIds(context.TODO(), db.Session(), &po.StudentCond{StudentName: "施鈞捷", ParentPhone: "0975670131"})
	if err != nil {
		panic(err)
	}

	log.Printf("ids: %v", ids)
}
