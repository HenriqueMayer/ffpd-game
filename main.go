// main.go - Loop principal do jogo
// O ponto de entrada do programa e o novo loop de jogo concorrente.
package main

import (
	"os"
	"time"
)

func main() {
	// --- INICIALIZAÇÃO ---
	// Inicializa a interface com o termbox.
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento da linha de comando.
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o estado do jogo.
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// --- CONFIGURAÇÃO DA CONCORRÊNCIA ---
	// Cria um canal para receber eventos do teclado.
	eventosTecladoCh := make(chan EventoTeclado)

	// Inicia uma goroutine separada para ler a entrada do teclado de forma contínua.
	go func() {
		for {
			// A chamada `interfaceLerEventoTeclado()` é bloqueante. Ela roda aqui sem travar o loop principal.
			eventosTecladoCh <- interfaceLerEventoTeclado()
		}
	}()

	// Cria um ticker que envia um sinal a cada 200 milissegundos. Este é o "pulso" do jogo.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// --- LOOP PRINCIPAL DO JOGO ---
	// Desenha o estado inicial do jogo antes de entrar no loop.
	interfaceDesenharJogo(&jogo)

loopPrincipal: // Rótulo para permitir a saída do loop de dentro do select. EN: Label to allow breaking the loop from within the select.
	for {
		// O `select` espera por uma mensagem em um dos canais. Ele processa o primeiro que chegar.
		select {

		case evento := <-eventosTecladoCh:
			// Trava o estado do jogo, executa a ação e o redesenho, e depois destrava.
			jogo.Travar()
			continuar := personagemExecutarAcao(evento, &jogo)
			if !continuar {
				jogo.Destravar() // Garante que destravamos antes de sair.
				break loopPrincipal
			}
			interfaceDesenharJogo(&jogo)
			jogo.Destravar()

		case <-ticker.C:
			// No futuro, a lógica dos inimigos virá aqui, também dentro de um par Travar/Destravar.
		}
	}
}