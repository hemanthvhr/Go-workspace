package main

import "fmt"
import "time"

func main(){
   n := 0
   i := uint32(3)
   j := 0
   y := 1
   var x uint32
   p := make(map[int]uint32)
   p[1] = 2
   fmt.Println("How many prime no's do you want - ")
   fmt.Scanf("%d",&n)
   fmt.Println();
   start := time.Now()
   if n<3 {
      return
   }
   for len(p)!= n {
      j = 0
      for y=1;y<=(len(p)) && j==0;y++ {
          x = p[y]
          if i%x==0 {
          j=1
          }
      }
      if j==0 {
      p[len(p)+1] = i
      }
      i += 2
   }
   for j=1;j<=len(p);j++ {
   i = p[j]
   fmt.Println(j, i)
   }
   elapsed := time.Since(start)
   fmt.Println("Time Taken to Execute is ",elapsed)
}