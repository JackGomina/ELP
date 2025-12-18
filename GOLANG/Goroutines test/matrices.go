package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func multiplyMatrices(a, b [][]int64) [][]int64 {
	L := len(a)
	var N = int64(L)
	C := make([][]int64, N)
	for i := range C {
		C[i] = make([]int64, N)
	}

	maxWorkers := runtime.NumCPU() * 2

	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	start := time.Now()

	for i := int64(0); i < N; i++ {
		wg.Add(1)
		sem <- struct{}{} // Réserve une place dans le sémaphore

		go func(i int64) {
			defer wg.Done()
			defer func() { <-sem }() // Libère la place dans le sémaphore une fois le travail terminé

			fmt.Printf("Goroutine lancée. Nombre total de Goroutines : %d\n", runtime.NumGoroutine())

			for j := int64(0); j < N; j++ {
				var sum int64
				for k := int64(0); k < N; k++ {
					sum += a[i][k] * b[k][j]
				}
				C[i][j] = sum
			}
		}(i) // Lance la goroutine au rang i, en copiant la valeur de i et non en prenant la variable pour éviter les problèmes
	}

	// Mesures avec runtime
	fmt.Printf("Nombre de CPUs utilisés : %d\n", runtime.NumCPU())
	fmt.Printf("Nombre de processeurs logiques : %d\n", runtime.GOMAXPROCS(0))

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Temps d'exécution de la multiplication des matrices: %v\n", duration)
	return C
}

func main() {

	A := [][]int64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	} // Attention à l'int overflow -> float64 si besoin

	B := [][]int64{
		{9, 8, 7},
		{6, 5, 4},
		{3, 2, 1},
	}

	C := multiplyMatrices(A, B)

	fmt.Println("Resultat de la multiplication des matrices A et B :")
	for _, row := range C {
		fmt.Println(row)
	}
}
