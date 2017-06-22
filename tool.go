package main

import (
	"fmt"
	//"net/mail"
	"time"
	"os"
	"bufio"
)

type FileRW struct {
    name string
    file *os.File
    writer *bufio.Writer
    reader *bufio.Reader
}

func testMail(t time.Time) {
	//t, _ := mail.ParseDate(s)
	a, b, c := t.Date()
	fmt.Printf("--- %d/%d/%d \n", a, b, c)
}

func openFile(fileName string, flag int)(out FileRW) {
    f, err := os.OpenFile(fileName, flag, 0664)
    if err != nil {
        panic(err)
    }
    out.name = fileName
    out.file = f
    out.writer = bufio.NewWriter(f)
    out.reader = bufio.NewReader(f)
    return out
}