package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func MiningHandler(difficulty int, timeout time.Duration) http.HandlerFunc {
	// redundant conversion: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		block := Block{}
		err = json.Unmarshal(body, &block)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		timer := time.NewTimer(timeout)
		// В случае, если main goroutine вылетает по таймауту, нужно:
		// 1. Сделать dataChan буфферизованным (в таком случае dataChan не удалится GC и в буфферизованный канал с размером 1
		// можно записать одно значение в неблокирующем режиме. Тогда "go generateHash" запишет свое значение в канал, не будет
		// заблокированна и завершится)
		dataChan := make(chan BlockMetadata, 1)
		// 2. Сделать chan done, доступный только на запись, для передачи информации. И закрывать его в случае падения по таймауту.
		done := make(chan struct{})
		go generateHash(block, difficulty, dataChan, done)
		select {
		case val := <-dataChan:
			block.Metadata.IterationCount, block.Metadata.Hash = val.IterationCount, val.Hash

			res, err := json.Marshal(block)
			if err != nil {
				timer.Stop()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// если работа горутины generateHash завершилась, стопаем таймер, чтобы его смог подчистить GC
			timer.Stop()
			w.WriteHeader(http.StatusOK)
			w.Write(res)
			return

		case <-timer.C:
			// если таймер подошел к концу, закрываем канал done, чтобы сообщить горутине generateHash о прекращении работы
			close(done)
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
	}
}

func generateHash(block Block, difficulty int, dataChan chan<- BlockMetadata, done <-chan struct{}) {
	prefix := strings.Repeat("0", difficulty)
	for i := int64(0); ; i++ {
		block.Metadata.IterationCount = i
		hash := block.Hash()
		if strings.HasPrefix(hash, prefix) {
			block.Metadata.Hash = hash
			dataChan <- block.Metadata
			break
		}
		select {
		case <-done:
			// если канал закрыт, значит мы вывалились по таймауту -> прерываем работу горутины
			return
		default:
			// если канал не закрыт, возвращаемся в новый цикл итерции
		}
	}
}
