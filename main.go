// main.go - Ponto de entrada e loop principal concorrente do jogo.
package main

import (
	"math"
	"os"
	"time"
)

func iniciarPatrulheiros(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Inimigo.simbolo {
				patrulheiro := &Patrulheiro{
					x: x, y: y, dx: 1, ultimoVisitado: Vazio,
					notificacaoCh: make(chan Coords, 1), alvo: nil,
				}
				jogo.patrulheiros = append(jogo.patrulheiros, patrulheiro)
				go rodarPatrulheiro(jogo, patrulheiro)
			}
		}
	}
}

func iniciarPortais(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == PortalFechado.simbolo {
				portal := &Portal{
					x: x, y: y, ativacaoCh: make(chan bool, 1),
				}
				jogo.portais = append(jogo.portais, portal)
				go rodarPortal(jogo, portal)
			}
		}
	}
}

func main() {
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

	// Inicia as goroutines para os elementos concorrentes (sem armadilhas).
	iniciarPatrulheiros(&jogo)
	iniciarPortais(&jogo)

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
			if !continuar || jogo.GameOver {
				interfaceDesenharJogo(&jogo)
				time.Sleep(2 * time.Second)
				jogo.Destravar()
				break loopPrincipal
			}

			interfaceDesenharJogo(&jogo)
			jogo.Destravar()

		case <-ticker.C:
			jogo.Travar()
			if jogo.GameOver {
				interfaceDesenharJogo(&jogo)
				time.Sleep(2 * time.Second)
				jogo.Destravar()
				break loopPrincipal
			}

			// LÓGICA DO RADAR
			const raioDeVisao = 8.0
			for _, p := range jogo.patrulheiros {
				distancia := math.Sqrt(math.Pow(float64(p.x-jogo.PosX), 2) + math.Pow(float64(p.y-jogo.PosY), 2))
				if distancia < raioDeVisao {
					select {
					case p.notificacaoCh <- Coords{jogo.PosX, jogo.PosY}:
					default:
					}
				}
			}

			interfaceDesenharJogo(&jogo)
			jogo.Destravar()
		}
	}
}