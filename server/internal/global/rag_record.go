/*
 * @Date: 2025-06-24 23:26:09
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-24 23:26:15
 * @FilePath: /thinking-map/server/internal/global/rag_record.go
 */
package global

import "github.com/PGshen/thinking-map/server/internal/repository"

var (
	ragRecordRepo repository.RAGRecord
)

func InitRAGRecordRepository(repo repository.RAGRecord) {
	ragRecordRepo = repo
}

func GetRAGRecordRepository() repository.RAGRecord {
	return ragRecordRepo
}