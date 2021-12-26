package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"distributed_bank/bank"
)



func main() {

	allNodeInfo := NodesInfo{
		ServerIDlist: []int{1, 2, 3, 4, 5, 6, 7},
		ServerAddrmap: map[int]string{
			1: "127.0.0.1:12110",
			2: "127.0.0.1:12111",
			3: "127.0.0.1:12112",
			4: "127.0.0.1:12113",
			5: "127.0.0.1:12114",
			6: "127.0.0.1:12115",
			7: "127.0.0.1:12116",
		},
	}

	//map for storing account number and balance
	aInfo := map[int]bank.Account{
		42: bank.Account{Number: 42,
			Balance: 1000},
		52: bank.Account{Number: 52,
			Balance: 10000},
		62: bank.Account{Number: 62,
			Balance: 2000},
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Type in your ID(1-7): ")
	scanner.Scan()
	ID, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal("Fail to transfor mstring to int: ", err)
	}

	fmt.Println("Enter Current Server IDs (seperate with comma):")
	scanner.Scan()
	txt := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading text:", err)
	}
	s := strings.Split(txt, ",")
	currentSerList := []int{}
	for _, v := range s {
		vInt, _ := strconv.Atoi(v)
		currentSerList = append(currentSerList, vInt)
	}

	deadsig := make(chan os.Signal, 1)

	distNet := NewDistNetwork(ID, currentSerList, allNodeInfo, aInfo)
	if err != nil {
		log.Fatal("Server err: ", err)
	}

	go distNet.StartServerWS()
	go distNet.startClientWS()
	go distNet.handleChan()
	<-deadsig
	fmt.Println("Goodbye")
}
