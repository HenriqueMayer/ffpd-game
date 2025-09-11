// main.go - Loop principal do jogo
package main

import (
	"math"
	"os"
	"time"
)

// iniciarPatrulheiros foi atualizado para criar o canal de notificação e adicionar à lista do jogo.
func iniciarPatrulheiros(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Inimigo.simbolo {
				patrulheiro := &Patrulheiro{
					x:              x,
					y:              y,
					dx:             1,
					ultimoVisitado: Vazio,
					notificacaoCh:  make(chan Coords, 1), // Cria o canal (buffer de 1 para não bloquear).
					alvo:           nil,
				}
				jogo.patrulheiros = append(jogo.patrulheiros, patrulheiro) // Adiciona à lista do jogo.
				go rodarPatrulheiro(jogo, patrulheiro)
			}
		}
	}
}

// ... (iniciarArmadilhas e outras funções continuam aqui) ...
func iniciarArmadilhas(jogo *Jogo) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == ArmadilhaArmada.simbolo {
				armadilha := &Armadilha{
					x: x, y: y, armada: true,
				}
				go rodarArmadilha(jogo, armadilha)
			}
		}
	}
}


func main() {
	// ... (inicialização do jogo como antes) ...
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

	iniciarPatrulheiros(&jogo)
	iniciarArmadilhas(&jogo)

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

			// --- LÓGICA DO RADAR (Nova) ---
			// Raio de visão do inimigo.
			const raioDeVisao = 8.0
			for _, p := range jogo.patrulheiros {
				// Calcula a distância euclidiana.
				distancia := math.Sqrt(math.Pow(float64(p.x-jogo.PosX), 2) + math.Pow(float64(p.y-jogo.PosY), 2))

				if distancia < raioDeVisao {
					// JOGADOR ESTÁ PERTO: Envia as coordenadas do jogador para o canal do patrulheiro.
					// Usamos um select para não bloquear caso o canal esteja cheio.
					select {
					case p.notificacaoCh <- Coords{jogo.PosX, jogo.PosY}:
					default: // Se o canal estiver cheio, não faz nada e segue o jogo.
					}
				}
			}

			interfaceDesenharJogo(&jogo)
			jogo.Destravar()
		}
	}
}