package main
 
import (
   "net"
   "fmt"
)
 
func main() {
   addr,err := net.LookupIP("flixvpn.flix-dev.com")
   if err != nil {
      fmt.Println("Unknown host")
   } else {
      fmt.Println("IP address: ", addr)
   }
}