package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

var posx,posy int

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}