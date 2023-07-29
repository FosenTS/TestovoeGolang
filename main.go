package main

import (
	"fmt"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	cT         string //время создания
	fT         string //время выполнения
	taskResult []byte
}

func main() {
	taskCreature := func(a chan Ttype) {
		go func() {
			for {
				ft := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
					ft = "Some error occurred"
				}
				a <- Ttype{cT: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
			}
		}()
	}

	superChan := make(chan Ttype, 10)

	go taskCreature(superChan)

	taskWorker := func(a Ttype, wg *sync.WaitGroup) Ttype {
		defer wg.Done()

		tt, _ := time.Parse(time.RFC3339, a.cT)
		if tt.After(time.Now().Add(-20 * time.Second)) {
			a.taskResult = []byte("task has been succeeded")
		} else {
			a.taskResult = []byte("something went wrong")
		}
		a.fT = time.Now().Format(time.RFC3339Nano)

		time.Sleep(time.Millisecond * 150)

		return a
	}

	var wg sync.WaitGroup

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)

	taskSorter := func(t Ttype) {
		if string(t.taskResult[14:]) == "succeeded" {
			doneTasks <- t
		} else {
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskResult)
		}
	}

	go func() {
		// получение тасков
		for t := range superChan {
			wg.Add(1)
			go func(t Ttype) {
				t = taskWorker(t, &wg)
				taskSorter(t)
			}(t)
		}
		wg.Wait()
		close(doneTasks)
		close(undoneTasks)
	}()

	var result sync.Map
	var err []error

	go func() {
		for r := range doneTasks {
			result.Store(r.id, r)
		}
		for r := range undoneTasks {
			err = append(err, r)
		}
	}()

	time.Sleep(time.Second * 3)

	fmt.Println("Errors:")
	for _, r := range err {
		fmt.Println(r)
	}

	fmt.Println("Done tasks:")
	result.Range(func(key, value interface{}) bool {
		fmt.Println(key)
		return true
	})
}
