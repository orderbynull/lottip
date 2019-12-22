package app

import "time"

type PgsqlPacket struct {
	Query     string
	Error     string
	Timestamp time.Time
}

type PgsqlRepository interface {
	GetAllByApp(appName string) []PgsqlPacket
}

type PgsqlService struct {
	repository PgsqlRepository
}

func NewPgsqlService(repository PgsqlRepository) *PgsqlService {
	return &PgsqlService{repository: repository}
}

func (ms *PgsqlService) GetAllByApp(appName string) []PgsqlPacket {
	return ms.repository.GetAllByApp(appName)
}
