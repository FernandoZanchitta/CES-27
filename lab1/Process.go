package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// Variáveis globais interessantes para o processo
var err string
var myPort string          //porta do meu servidor
var nServers int           //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
// dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
// mensagens dos outros processos)

var id int                    // numero de identificador de processo
var my_logical_clock int      // inicia a contagem do relogio logico para 0
var estou_na_cs bool          // verifica caso HELD
var estou_esperando bool      // verifica caso HOLD em caso de falso duplo => released
var received_all_replies bool // verificacao dos replies para acesso a CS
var shared_resource *net.UDPConn
var queued_request []int
var lc_requisicao int
var replied_received []int

func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}
func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}
func doServerJob() { //Loop infinito mesmo
	for {
		//Ler (uma vez somente) da conexão UDP a mensagem
		//Escrever na tela a msg recebida (indicando o
		//endereço de quem enviou)
		//FALTA ALGO AQUI
		buf := make([]byte, 1024)
		n, addr, err := ServConn.ReadFromUDP(buf)
		msg = string(buf[0:n])
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
func doClientJob(otherProcess int, i int) {
	//Enviar uma mensagem (com valor i) para o servidor do processo //otherServer.

	// Ler (uma vez somente) da conexão UDP a mensagem
	// Escrever na tela a msg recebida (indicando o
	// endereço de quem enviou)
	// FALTA ALGO AQUI
	msg := strconv.Itoa(i)
	buf := []byte(msg)
	_, err := CliConn[otherProcess].Write(buf)
	// FALTA ALGO AQUI
	if err != nil {
		fmt.Println(msg, err)
	}
}

// ESSA FUNÇÃO ESTÁ PRONTA
func initConnections() {
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[id+1]
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira a primeira porta (que é a minha). As demais portas são dos outros processos*/
	CliConn = make([]*net.UDPConn, nServers)
	/*Outros códigos para deixar ok a conexão do meu servidor (onde re-
	cebo msgs). O processo já deve ficar habilitado a receber msgs.*/

	/*Outros códigos para deixar ok a minha conexão com cada servidor dos outros processos. Colocar tais conexões no vetor CliConn.*/
	for servidores := 0; servidores < nServers; servidores++ {
		ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+os.Args[2+servidores])
		CheckError(err)
		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		CheckError(err)
		Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
		CliConn[servidores] = Conn
		CheckError(err)
	}
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)
	shared_resource, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	// iniciando o clock logico do processo
	my_logical_clock = 0
}
func readInput(ch chan string) {
	// Rotina que "escuta" o stdin
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func request_CS() {

}
func use_CS() {

}
func reply_any_queued_request() {

}
func exit_CS() {
	estou_esperando = false
	estou_na_cs = false
	received_all_replies = false
	reply_any_queued_request()
	replied_received = nil

}
func Ricart_Agrawala(logical_clock_req int, text_mensagem string) {
	estou_esperando = true
	request_CS(logical_clock_req)
	fmt.Println("Estou esperando receber os replies\n ")
	for !received_all_replies {
	}
	fmt.Println("Entrei na CS!")
	use_CS(logical_clock_req, text_mensagem)
	fmt.Println("Sai da CS!")
	exit_CS()
	fmt.Println("Liberei a CS!")
}

// ESSA FUNÇÃO ESTÁ PRONTA
func main() {
	initConnections()
	estou_na_cs = false
	estou_esperando = false
	//O fechamento de conexões deve ficar aqui, assim só fecha //conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	/*Todos Process fará a mesma coisa: ficar ouvindo mensagens e man- dar infinitos i’s para os outros processos*/

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)
	go doServerJob()
	for {
		//Verificar (de forma nao bloqueante ) se tem algo no
		//stdin (input do terminal)
		select {
		case x, valid := <-ch:
			if valid {
				compare, _ = strconv.Atoi(x)
				if compare != id && x == "x" {
					if estou_na_cs || estou_esperando {
						fmt.Println("x ignorado\n")
					} else {
						fmt.Println("Solicitando acesso a CS com ID = %d e Logical Clock = %d\n", id, my_logical_clock)
						text_mensagem = "TEXTO TESTE MENSAGEM"
						lc_requisicao = my_logical_clock
						go Ricart_Agrawala(lc_requisicao, text_mensagem)
					}
				}
			} else {
				fmt.Println("Canal fechado!")
			}
		default:
			// Fazer nada!
			// Mas não fica bloqueado esperando o teclado
			time.Sleep(time.Second * 1)
		}
		//Esperar um pouco
		time.Sleep(time.Second * 1)

	}
}
