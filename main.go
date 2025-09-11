// main.go - Loop principal do jogo
package main

import (
	"os"
	"time"
)

// iniciarPatrulheiros varre o mapa, encontra todos os inimigos e inicia uma goroutine para cada um.
func iniciarPatrulheiros(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Inimigo.simbolo {
				// Cria um novo patrulheiro para este inimigo.
				patrulheiro := &Patrulheiro{
					x:              x,
					y:              y,
					dx:             1, // Começa movendo para a direita.
					ultimoVisitado: Vazio, // Assume que o inimigo começa sobre uma célula vazia.
				}
				// Inicia a goroutine que vai controlar este patrulheiro.
				go rodarPatrulheiro(jogo, patrulheiro)
			}
		}
	}
}

func main() {
	// --- INICIALIZAÇÃO ---
	interfaceIniciar()
	defer interfaceFinalizar()

	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Inicia as goroutines para os patrulheiros após o mapa ser carregado.
	iniciarPatrulheiros(&jogo)

	// --- CONFIGURAÇÃO DA CONCORRÊNCIA ---
	eventosTecladoCh := make(chan EventoTeclado)
	go func() {
		for {
			eventosTecladoCh <- interfaceLerEventoTeclado()
		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond) // Ticker principal para redesenho.
	defer ticker.Stop()

	// --- LOOP PRINCIPAL DO Jogo ---
	interfaceDesenharJogo(&jogo)

loopPrincipal:
	for {
		select {

		case evento := <-eventosTecladoCh:
			jogo.Travar()
			continuar := personagemExecutarAcao(evento, &jogo)
			if !continuar {
				jogo.Destravar()
				break loopPrincipal
			}
			// Redesenha imediatamente após a ação do jogador para resposta rápida.
			interfaceDesenharJogo(&jogo)
			jogo.Destravar()

		case <-ticker.C:
			// O ticker principal pulsa, então redesenhamos a tela para ver
			// as mudanças feitas pelos patrulheiros em suas goroutines.
			jogo.Travar()
			interfaceDesenharJogo(&jogo)
			jogo.Destravar()
		}
	}
}