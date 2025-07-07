package main

import (
	"fmt"
	"sort"
	"sync"
)

var (
	ths = [...]string{"0", "1", "2", "3", "4", "5"}
)

func ExecutePipeline(tasks ...job) {

	chans := make([]chan interface{}, len(tasks)+1)

	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))

	for i := range chans {
		chans[i] = make(chan interface{})
	}

	for i := range tasks {
		go func(itr int) {
			defer wg.Done()
			tasks[itr](chans[itr], chans[itr+1])
			close(chans[itr+1])
		}(i)
	}

	wg.Wait()

}

func SingleHash(in, out chan interface{}) {
	waits := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for elem := range in {
		waits.Add(1)
		go func(curElem interface{}) {
			defer waits.Done()
			str := fmt.Sprint(curElem)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			var temp, temp2 string
			go func() {
				defer wg.Done()
				temp = DataSignerCrc32(str) + "~"
			}()
			mu.Lock()
			temp2 = DataSignerMd5(str)
			mu.Unlock()
			temp2 = DataSignerCrc32(temp2)
			wg.Wait()
			out <- temp + temp2
		}(elem)
	}
	waits.Wait()
}

func MultiHash(in, out chan interface{}) {

	waits := &sync.WaitGroup{}

	for elem := range in {
		waits.Add(1)
		go func(curElem interface{}) {
			defer waits.Done()
			ans := ""
			str := fmt.Sprint(curElem)
			crcSlice := make([]string, 6)

			wg := &sync.WaitGroup{}
			wg.Add(6)

			for i := range ths {
				go func(num int) {
					defer wg.Done()
					crcSlice[num] = DataSignerCrc32(ths[num] + str)
				}(i)
			}

			wg.Wait()

			for _, part := range crcSlice {
				ans += part
			}

			out <- ans
		}(elem)
	}

	waits.Wait()

}

func CombineResults(in, out chan interface{}) {
	strs := make([]string, 0)
	for elem := range in {
		strs = append(strs, fmt.Sprint(elem))
	}

	sort.Slice(strs, func(i, j int) bool {
		return strs[i] < strs[j]
	})

	ans := ""
	for i, elem := range strs {
		ans += elem
		if i != len(strs)-1 {
			ans += "_"
		}
	}

	out <- ans

}
