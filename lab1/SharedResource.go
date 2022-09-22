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
		//TODO: Preencher essa função
	}
}