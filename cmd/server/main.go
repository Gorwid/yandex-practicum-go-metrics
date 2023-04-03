/*
 * Cервер для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.
 * Агент вам предстоит реализовать в следующем инкременте — в качестве источника метрик вы будете использовать пакет runtime.
 *
 * Сервер должен быть доступен по адресу http://localhost:8080, а также:
 * Принимать и хранить произвольные метрики двух типов:
 * Тип gauge, float64 — новое значение должно замещать предыдущее.
 * Тип counter, int64 — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
 * Принимать метрики по протоколу HTTP методом POST.
 * Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain.
 * При успешном приёме возвращать http.StatusOK.
 *
 * Для хранения метрик объявите тип MemStorage. Рекомендуем использовать тип struct с полем-коллекцией внутри (slice или map).
 *
 * В будущем это позволит добавлять к объекту хранилища новые поля, например логер или мьютекс, чтобы можно было использовать их в методах.
 * Опишите интерфейс для взаимодействия с этим хранилищем.
 *
 * Пример запроса к серверу:
 * POST /update/counter/someMetric/527 HTTP/1.1
 * Host: localhost:8080
 * Content-Length: 0
 * Content-Type: text/plain
 *
 * Пример ответа от сервера:
 * HTTP/1.1 200 OK
 * Date: Tue, 21 Feb 2023 02:51:35 GMT
 * Content-Length: 11
 * Content-Type: text/plain; charset=utf-8
 */

package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type storager interface {
	storageUpdater(val []string)
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (m MemStorage) storageUpdater(val []string) error {
	var (
		err      error
		numGauge float64
		numCount int64
	)
	if val[2] != "" && val[3] != "" && val[4] != "" {
		switch {
		case val[2] == "gauge":
			numGauge, err = strconv.ParseFloat(val[4], 64)
			if err != nil {
				return err
			}
			m.gauge[val[3]] = numGauge
		case val[2] == "counter":
			numCount, err = strconv.ParseInt(val[4], 10, 64)
			if err != nil {
				return err
			}
			if _, ok := m.counter[val[3]]; ok {
				m.counter[val[3]] = m.counter[val[4]] + numCount
			} else {
				m.counter[val[3]] = numCount
			}
		default:
			return errors.New("invalid update request")
		}
	} else {
		return errors.New("invalid update request")
	}
	return nil

}

// const useThePostMethodMsg = `Use the POST method:
// http://localhost:8080/update/<metric_type>/<metric_name>/<metric_value>`

func mainpage(store MemStorage) func(answer http.ResponseWriter, req *http.Request) {

	return func(answer http.ResponseWriter, req *http.Request) {
		splitURL := strings.Split(req.URL.Path, "/")
		if req.Method == http.MethodPost && splitURL[1] == "update" {
			err := store.storageUpdater(splitURL)
			if err != nil {
				answer.WriteHeader(http.StatusBadRequest)
			}
			answer.Header().Add("Content-Type", "text/plain")
			answer.WriteHeader(http.StatusOK)
		} else {
			answer.WriteHeader(http.StatusNotFound)
		}
	}
}

func main() {
	var vault MemStorage
	vault.gauge = make(map[string]float64)
	vault.counter = make(map[string]int64)

	metrics := http.NewServeMux()
	metrics.HandleFunc("/", mainpage(vault))

	err := http.ListenAndServe(`:8080`, metrics)
	if err != nil {
		panic(err)
	}
}
