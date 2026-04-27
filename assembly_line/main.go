package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Employee struct {
	id    int
	count int
}

func (e *Employee) work(wg *sync.WaitGroup, jobs <-chan Item, mu *sync.Mutex) {
	defer wg.Done()
	for item := range jobs {
		fmt.Printf("員工: %d 開始處理 %s\n", e.id, item.Name())
		item.Process()
		fmt.Printf("員工 %d 完成處理 %s\n", e.id, item.Name())

		mu.Lock()
		e.count++
		mu.Unlock()
	}
}

type Item1 struct {
	id int
}

type Item2 struct {
	id int
}

type Item3 struct {
	id int
}

type Item interface {
	// Process 這是一個耗時操作
	Process()
	Name() string
}

func (i Item1) Process() { time.Sleep(100 * time.Millisecond) }
func (i Item2) Process() { time.Sleep(200 * time.Millisecond) }
func (i Item3) Process() { time.Sleep(300 * time.Millisecond) }

func (i Item1) Name() string {
	return fmt.Sprintf("Item1-%d", i.id)
}
func (i Item2) Name() string {
	return fmt.Sprintf("Item2-%d", i.id)
}
func (i Item3) Name() string {
	return fmt.Sprintf("Item3-%d", i.id)
}

func main() {
	items := make([]Item, 0, 30)
	for i := 1; i <= 10; i++ {
		items = append(items, Item1{id: i})
		items = append(items, Item2{id: i})
		items = append(items, Item3{id: i})
	}

	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	jobs := make(chan Item, len(items))
	for _, item := range items {
		jobs <- item
	}
	close(jobs)

	employees := make([]*Employee, 5)
	for i := range employees {
		employees[i] = &Employee{id: i + 1}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	start := time.Now()

	for _, emp := range employees {
		wg.Add(1)
		go emp.work(&wg, jobs, &mu)
	}

	wg.Wait()

	elapsed := time.Since(start)

	fmt.Printf("總處理時間: %v\n", elapsed)
	total := 0
	for _, emp := range employees {
		fmt.Printf("員工 %d 處理了 %d 件物品\n", emp.id, emp.count)
		total += emp.count
	}
	fmt.Printf("總處理件數: %d\n", total)
}
