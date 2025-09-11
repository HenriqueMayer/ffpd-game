// main.go - Loop principal do jogo
package main

import (
	"os"
	"time"
)

// iniciarPatrulheiros (Existente)
func iniciarPatrulheiros(jogo *Jogo) {
	// ... (código sem alteração)
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Inimigo.simbolo {
				patrulheiro := &Patrulheiro{
					x: x, y: y, dx: 1, ultimoVisitado: Vazio,
				}
				go rodarPatrulheiro(jogo, patrulheiro)
			}
		}
	}
}

// iniciarArmadilhas (Nova)
// Varre o mapa, encontra todas as armadilhas e inicia uma goroutine para cada uma.
func iniciarArmadilhas(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == ArmadilhaArmada.simbolo {
				// Cria uma nova armadilha para este local.
				armadilha := &Armadilha{
					x:      x,
					y:      y,
					armada: true, // Começa armada.
				}
				// Inicia a goroutine que vai controlar esta armadilha.
				go rodarArmadilha(jogo, armadilha)
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

	// Inicia as goroutines para os elementos concorrentes.
	iniciarPatrulheiros(&jogo)
	iniciarArmadilhas(&jogo) // << NOVA CHAMADA AQUI

	// ... (resto do main sem alterações)
	eventosTecladoCh := make(chan EventoTeclado)
	go func() {
		for {
			eventosTecladoCh <- interfaceLerEventoTeclado()
		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
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
			interfaceDesenharJogo(&jogo)
			jogo.Destravar()
		case <-ticker.C:
			jogo.Travar()
			interfaceDesenharJogo(&jogo)
			jogo.Destravar()
		}
	}
}