package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var mutex sync.Mutex

// Размер буфера
const lenghtOfBuf int = 10

func main() {
	//test commit
	//test commit 2
	//1)Отвечает за постоянное считывание данных с консоли, ведет запись в кольцевой буффер
	//2)Отвечает за отправку сигнала о завершении программы при введении "end"
	wait := func(done chan bool, permission []chan bool, arr *[lenghtOfBuf]int, arrIndex *int, arrLastIndex *int) {
		var p int
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Введите набор чисел:")
		for {
			<-permission[0]
			slice, _ := reader.ReadString('\n')
			*arrLastIndex = *arrIndex
			for _, k := range strings.Fields(slice) {

				if k == "end" {
					done <- true
					return
				}

				p, _ = strconv.Atoi(k)

				mutex.Lock()

				arr[*arrIndex] = p
				*arrIndex++
				if *arrIndex > lenghtOfBuf-1 {
					*arrIndex = 0
				}

				mutex.Unlock()
			}
			fmt.Println("Данные получены,текущее состояние буфера:", *arr)
			permission[1] <- true
		}
	}

	//Отвечает передачу информации из кольцевого буффера в канал
	init := func(permission []chan bool, arr *[lenghtOfBuf]int, arrIndex *int, arrLastIndex *int) <-chan int {
		output := make(chan int, lenghtOfBuf)
		go func() {
			for {
				<-permission[1]
				for i := *arrLastIndex; i != *arrIndex; i = (i + 1) % len(arr) {
					output <- arr[i]
				}
				permission[2] <- true
			}
		}()
		return output
	}

	//Фильтр положительных значений
	positive := func(input <-chan int) <-chan int {
		filtredPos := make(chan int)
		go func() {
			defer close(filtredPos)
			for {
				select {
				case i, isChannelOpen := <-input:
					if !isChannelOpen {
						return
					}
					if i > 0 {
						select {
						case filtredPos <- i:
						}
					}
				}
			}

		}()
		return filtredPos
	}

	//Фильтр значений кратных 3-ем
	multiplicity := func(input <-chan int) <-chan int {
		filtredThree := make(chan int)
		go func() {
			defer close(filtredThree)
			for {
				select {
				case i, isChannelOpen := <-input:
					if !isChannelOpen {
						return
					}
					if i%3 == 0 {
						select {
						case filtredThree <- i:
						}
					}
				}
			}

		}()
		return filtredThree
	}

	var arr [lenghtOfBuf]int
	var arrIndex int
	var arrLastIndex int
	permission := make([]chan bool, 3)
	done := make(chan bool)

	for i := range permission {
		permission[i] = make(chan bool, 0)
	}

	go wait(done, permission, &arr, &arrIndex, &arrLastIndex)
	permission[0] <- true
	intStream := init(permission, &arr, &arrIndex, &arrLastIndex)

	for {
		select {
		case <-permission[2]:
			permission[0] <- true
			pipline := multiplicity(positive(intStream))
			go func() {
				for v := range pipline {
					fmt.Println(v)
				}
			}()
		case <-done:
			fmt.Println("Программа завершена")
			return
		}
	}
}
