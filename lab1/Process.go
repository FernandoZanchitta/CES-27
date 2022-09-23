package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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
func pushReplyQueue(pj_id int) {
	queued_request = append(queued_request, pj_id)
}
func sendReply(pj_id int, lc_pj int) {
	// preciso mandar mensagem para o outro processo
	// conteudo da mensagem: id my_logical_clock reply
	msg := strconv.Itoa(id) + "," + strconv.Itoa(my_logical_clock) + "," + "reply"
	sendMsg(pj_id, msg)

}
func amIPriority(pj_id int, pj_lc int) bool {
	// criterios: para um clock menor > Mesmo clock: id menor > resto
	if my_logical_clock < pj_lc {
		return true
	} else if my_logical_clock == pj_lc {
		if id < pj_id { // todo: verificar se tem problema de igualdade, ou se pode deixar
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
func receiveReply(pj_id int) {
	//todo: revisar essa funcao
	//devemos verificar se a mensagem de reply ja se encontra na lista
	is_new := true
	for _, content := range replied_received {
		if pj_id == content {
			is_new = false
		}
	}
	if is_new {
		replied_received = append(replied_received, pj_id)
	}
	if len(replied_received) >= nServers { // todo: verificar se não é nServer - 1
		received_all_replies = true
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
		msg := string(buf[0:n])
		msg_parse := strings.Split(msg, ",")
		str_pj_id := msg_parse[0]
		str_pj_lc := msg_parse[1]
		str_pj_content := msg_parse[1]
		pj_id, err := strconv.Atoi(str_pj_id)
		lc_pj, err := strconv.Atoi(str_pj_lc)
		//TODO: Preencher essa função
		if str_pj_id != string(id) {
			// caso mensagem venha de outro processo
			if str_pj_content == "reply" {
				// caso mensagem seja de reply
				receiveReply(pj_id)

			} else if str_pj_content == "request" {
				// recebido o request
				// Caso esteja Held || Wanted com menor prioridade:
				if estou_na_cs || (estou_esperando && amIPriority(pj_id, lc_pj)) {
					//devo colocar oprocesso na fila de prioridade
					pushReplyQueue(pj_id)
				} else {
					// Caso contrario: enviar reply
					sendReply(pj_id, lc_pj)
				}
			} else {
				fmt.Println("Mensagem não identificada: %s \n", str_pj_content)
			}

		} else {
			// caso mensagem tenha id igual ao meu id
		}

		fmt.Println("Received ", msg, " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
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
	PrintError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	PrintError(err)
	shared_resource, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	PrintError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	PrintError(err)
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

func sendMsg(other_process int, msg string) {
	//Enviar uma mensagem (com valor i) para o servidor do processo //otherServer.x
	buf := []byte(msg)
	_, err := CliConn[other_process-1].Write(buf)
	// FALTA ALGO AQUI
	PrintError(err)
}

func requestCS(logical_clock_req int) {
	// devo mandar mensagem de request a todos os outros
	//processos, com o conteudo da mensagem: id,lc,request

	// formatando a mensagem que vai ser enviada
	text_msg := "request"
	id_msg := strconv.Itoa(id)
	lc_msg := strconv.Itoa(logical_clock_req)
	msg := id_msg + "," + lc_msg + "," + text_msg
	for other_process := 1; other_process <= nServers; other_process++ {
		if other_process != id {
			sendMsg(other_process, msg)
		}
	}
	// enviar mensagem para todos os processos existentes

}
func useCS(logical_clock_req int, text_mensagem string) {
	// setando flag de HELD
	estou_na_cs = true

	// formatando a mensagem
	lc_msg := strconv.Itoa(logical_clock_req)
	id_msg := strconv.Itoa(id)
	msg := id_msg + "," + lc_msg + "," + text_mensagem
	buf := []byte(msg)
	// enviar mensagem para o shared_resource
	_, err := shared_resource.Write(buf)
	PrintError(err)
	//esperar
	time.Sleep(time.Second * 2)
}
func replyAnyQueuedRequest() {
	text_mensagem := "reply"
	id_msg := strconv.Itoa(id)
	lc_msg := strconv.Itoa(my_logical_clock)
	msg := id_msg + "," + lc_msg + "," + text_mensagem
	buf := []byte(msg)
	for _, pjid := range replied_received {
		// dentro do replied_received contem todos os process_id de cada um dos que pediram acesso
		index := pjid - 1
		_, err := CliConn[index].Write(buf)
		PrintError(err)
	}
}
func exitCS() {
	estou_esperando = false
	estou_na_cs = false
	received_all_replies = false
	replyAnyQueuedRequest()
	replied_received = nil

}
func Ricart_Agrawala(logical_clock_req int, text_mensagem string) {
	estou_esperando = true
	requestCS(logical_clock_req)
	fmt.Println("Estou esperando receber os replies\n ")
	for !received_all_replies {
	}
	fmt.Println("Entrei na CS!")
	useCS(logical_clock_req, text_mensagem)
	fmt.Println("Sai da CS!")
	exitCS()
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
				compare, _ := strconv.Atoi(x)
				if compare != id && x == "x" {
					if estou_na_cs || estou_esperando {
						fmt.Println("x ignorado\n")
					} else {
						fmt.Println("Solicitando acesso a CS com ID = %d e Logical Clock = %d\n", id, my_logical_clock)
						text_mensagem := "TEXTO TESTE MENSAGEM"
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
