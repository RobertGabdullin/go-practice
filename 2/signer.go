package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}

	chanList := make([]chan interface{}, len(jobs)+1)
	for i := range chanList {
		chanList[i] = make(chan interface{})
	}

	for idx, curJob := range jobs {
		wg.Add(1)
		go func(idx int, curJob job) {
			defer close(chanList[idx+1])
			defer wg.Done()
			curJob(chanList[idx], chanList[idx+1])
		}(idx, curJob)
	}

	close(chanList[0])
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wgGlobal := &sync.WaitGroup{}
	mx := &sync.Mutex{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {

			defer wgGlobal.Done()

			wg := &sync.WaitGroup{}
			wg.Add(2)

			str := fmt.Sprint(data)
			left, right := "", ""

			go func() {
				defer wg.Done()
				left = DataSignerCrc32(str)
			}()

			go func() {
				defer wg.Done()
				mx.Lock()
				md := DataSignerMd5(str)
				mx.Unlock()
				right = DataSignerCrc32(md)
			}()

			wg.Wait()

			res := left + "~" + right
			out <- res
		}(data)
	}
	wgGlobal.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgGlobal := &sync.WaitGroup{}
	for data := range in {
		wgGlobal.Add(1)
		go func(data interface{}) {
			defer wgGlobal.Done()
			str := fmt.Sprint(data)
			builder := strings.Builder{}
			tempList := make([]string, 6)
			wg := &sync.WaitGroup{}

			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					tempList[i] = DataSignerCrc32(fmt.Sprint(i) + str)
				}(i)
			}

			wg.Wait()
			for i := 0; i < 6; i++ {
				builder.WriteString(tempList[i])
			}

			res := builder.String()
			out <- res
		}(data)
	}
	wgGlobal.Wait()
}

func CombineResults(in, out chan interface{}) {
	resList := make([]string, 0)
	for data := range in {
		str := fmt.Sprint(data)
		resList = append(resList, str)
	}

	sort.Slice(resList, func(i, j int) bool {
		return resList[i] < resList[j]
	})

	res := strings.Builder{}
	res.WriteString(resList[0])
	for i := 1; i < len(resList); i++ {
		res.WriteString("_" + resList[i])
	}

	out <- res.String()

}
