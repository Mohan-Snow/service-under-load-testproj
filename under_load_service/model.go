package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

type Block struct {
	From     string        `json:"from"`
	To       string        `json:"to"`
	Value    int64         `json:"value"`
	Metadata BlockMetadata `json:"metadata"`
}

type BlockMetadata struct {
	IterationCount int64  `json:"IterationCount"`
	Hash           string `json:"Hash"`
}

func (t Block) Hash() string {
	buf := bytes.Buffer{}
	buf.WriteString(t.From)
	buf.WriteString("->")
	buf.WriteString(t.To)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(int(t.Value)))
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(int(t.Metadata.IterationCount)))

	v := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(v[:])

	// для оптимизиции можно заменить функцию Sprintf из пакета fmt на Буффер байтов или StringsBuilder
	// (Использовать такую оптимизацию, если есть большая нагрузка на сервис)
	// v := sha256.Sum256([]byte(fmt.Sprintf("%s->%s:%d:%d", t.From, t.To, t.Value, t.Metadata.IterationCount)))
	// return fmt.Sprintf("%x", v)
}
