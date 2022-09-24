package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()
	for {
		//Loop infinito para receber mensagem e escrever todos os
		//conteúdos (processo que enviou, relógio recebido e texto)
		//na tela
		buf := make([]byte, 1024)
		n, _, err := Connection.ReadFromUDP(buf)
		msg := string(buf[0:n])
		msg_parser := strings.Split(msg, ",")
		fmt.Println("Received message....\n from ID:", msg_parser[0], "\nwith Logical Clock:", msg_parser[1], "\nBODY: ", msg_parser[2])
		CheckError(err)

	}
	time.Sleep(time.Second * 1)
}
