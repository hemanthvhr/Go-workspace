package main 

import "fmt"

/*

Imagine a gene that is 108 letters with A, T, G, C in random sequence. 
Assume that every year, there is a random change — one of the letters somewhere on this
gene mutates and is replaced by one of the other three. After each year, you compare 
the current copy of the gene with the original and tally how many letters have changed. 
After a certain time “the evolutionary clock will have slowed to a crawl” — that is, 
the number of changed letters will have stopped rising. The evolutionary rate from here 
on is zero. 
How many letters of the original gene will have changed at that point? 
How many years will it take to get to this point? 
Is the curve exponential?

Source  = https://www.quantamagazine.org/20170316-the-evolutionary-clock-puzzle/

*/
var D int = 108 // No of genes given
var p float64 = 1.0/(108.0) //Probability of selecting any one letter
var q float64 = 1/3  

func min(x, y int) int {
    if(x <= y) {
    	return x
    }
    return y
}
func prob(store map[int]map[int]float64,X , i int) float64 { //Recursive Definition of the probability after X Years
	if(i > min(X,D)) { return 0.0 }
	if(X < 0 || i < 0) { return 0.0 }
	temp := store[X-1]
	return (temp[i-1]*(float64)(D-i+1)*p + temp[i]*((float64)(2*i)*p*q) + temp[i+1]*(float64)(i+1)*p*q)
}
func main() {

	var X int
	var i int
	store := make(map[int]map[int]float64)
	store[1] = make(map[int]float64)
	(store[1])[0] = 0.0
	(store[1])[1] = 1.0 
	for X = 2;X < 2*D;X++ {
		store[X] = make(map[int]float64)
		temp := store[X]
		for i = 0;i<=min(X,D);i++ {
			temp[i] = prob(store,X,i)
		} 
	}
 	for X = 1;X<2*D;X++ {
 		temp := store[X]
 		for i = 0;i<min(X,D);i++ {
 			fmt.Print(temp[i]," ")
 		}
 		fmt.Println()
 	}
}
