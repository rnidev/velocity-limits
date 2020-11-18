package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/rnidev/velocity-limits/cmd/pkg/account"
	validator "gopkg.in/go-playground/validator.v9"
)

func main() {
	inputs, err := readInputFile("../../input.txt")
	if err != nil {
		panic(err)
	}
	output, err := os.OpenFile("../../output.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	for _, input := range inputs {
		var s account.Service
		s = account.CustomerAccount{}
		v := validator.New()
		handler := account.NewHandler(s, v, c)
		response := handler.Run(input)
		//ignore empty response
		if (response == account.FundResponse{}) {
			continue
		}
		jsonByte, err := json.Marshal(&response)
		if err != nil {
			log.Print(err)
			continue
		}
		_, err = fmt.Fprintln(output, string(jsonByte))
		if err != nil {
			log.Print(err)
			continue
		}
	}
	err = output.Close()
	if err != nil {
		panic(err)
	}
	c.Flush()
}

func readInputFile(filePath string) ([]string, error) {
	var inputs []string
	file, err := os.Open(filePath)
	if err != nil {
		return inputs, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		var (
			buffer   bytes.Buffer
			line     []byte
			isPrefix bool
		)
		for {
			line, isPrefix, err = reader.ReadLine()
			buffer.Write(line)
			if !isPrefix {
				break
			}
			if err != nil {
				if err != io.EOF {
					return inputs, err
				}
				break
			}
		}
		lineString := buffer.String()
		if err == io.EOF {
			break
		}
		if len(lineString) > 0 {
			inputs = append(inputs, lineString)
		}
	}
	if err != io.EOF {
		return inputs, err
	}
	return inputs, nil
}
