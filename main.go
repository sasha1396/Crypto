package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const prefix = "000000"

var gogroup sync.WaitGroup // создал переменную для группы горутин

func main() {
	t := time.Now()         // текущее время
	rand.Seed(t.UnixNano()) // задаем семя

	gogroup.Add(runtime.NumCPU()) // группа из 6 горутин

	hash := make(chan [32]byte)
	tm := make(chan time.Duration)
	c := make(chan int)

	//запускаю горутины (количество ядер = количеству горутин)
	for i := 0; i < runtime.NumCPU(); i++ {
		go findhash(i, c, interrupt(), hash, tm)
	}
	//горутина для вывода хеша
	go printhash(hash, c, tm)

	gogroup.Wait() // жду завершения горутин
	fmt.Println("Программа завершена")
}

func findhash(id int, c chan int, quit <-chan os.Signal, hash chan [32]byte, tm chan time.Duration) {
	defer gogroup.Done()

	start := time.Now()
	tflag := true
	buf := make([]byte, binary.MaxVarintLen64) // срез для байтов случайного числа
	number := rand.Int63()                     // случайное число

	for {
		select {
		case <-quit:
			fmt.Println("Поток:", id, "завершил работу")
			return
		default:
			// метка отсчета выполнения поиска
			if tflag {
				start = time.Now()
				tflag = false
			}
			binary.PutVarint(buf, number) // количество байтов
			sum := sha256.Sum256(buf)     // хеш от числа
			hex := fmt.Sprintf("%x", sum[:])
			number++

			if hex[:len(prefix)] == prefix {
				hash <- sum
				t := time.Now()         // метка конца поиска
				elapsed := t.Sub(start) // вычисления разницы между меткой начала и конца
				tm <- elapsed
				c <- id
				tflag = true // возвращаем флаг чтобы снова отсчитывать время
			}

		}
	}
}

// вывод хэша
func printhash(hash chan [32]byte, c chan int, tm chan time.Duration) {
	for {
		fmt.Printf("Хеш: %x\n", <-hash)
		fmt.Println("Время нахождения:", <-tm, "Поток:", <-c)
	}
}

// вызов прерывания
func interrupt() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return c
}
