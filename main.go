package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const prefix = "000000"

var gogroup sync.WaitGroup // создал переменную для группы горутин

func main() {
	t := time.Now()         // текущее время
	rand.Seed(t.UnixNano()) // задаем семя

	gogroup.Add(runtime.NumCPU()) // группа из 6 горутин

	//запускаю горутины (количество ядер = количеству горутин)
	for i := 0; i < runtime.NumCPU(); i++ {
		go findhash(i)
	}

	gogroup.Wait() // жду завершения горутин
}

func findhash(id int) {
	defer gogroup.Done()

	start := time.Now()
	tflag := true

	buf := make([]byte, binary.MaxVarintLen64) // срез для байтов случайного числа
	number := rand.Int63()                     // случайное число

	for {
		// метка отсчета выполнения поиска
		if tflag {
			start = time.Now()
			tflag = false
		}
		n := binary.PutVarint(buf, number) // количество байтов
		sum := sha256.Sum256(buf)          // хеш от числа
		hex := fmt.Sprintf("%x", sum[:])
		number++

		if hex[:len(prefix)] == prefix {
			fmt.Printf("хеш: %x\n", sum)
			fmt.Printf("%x\n", buf[:n])
			t := time.Now()         // метка конца поиска
			elapsed := t.Sub(start) // вычисления разницы между меткой начала и конца
			fmt.Println("Поток:", id, "Время нахождения:", elapsed)
			tflag = true // возвращаем флаг чтобы снова отсчитывать время
		}
	}
}
