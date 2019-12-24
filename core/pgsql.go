package core

import "time"

type PgsqlPacket struct {
	Id        int
	Query     string
	Error     string
	Timestamp time.Time
}

type PgsqlRepository interface {
	GetAllByApp(appName string, showType string, showedLast int, limit int) []PgsqlPacket
}

type PgsqlService struct {
	repository PgsqlRepository
}

func NewPgsqlService(repository PgsqlRepository) *PgsqlService {
	return &PgsqlService{repository: repository}
}

func (ms *PgsqlService) GetAllByApp(appName string, showType string, showedLast int, limit int) []PgsqlPacket {
	return ms.repository.GetAllByApp(appName, showType, showedLast, limit)
}
