// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	// Verifica se o movimento é permitido
	if jogoPodeMoverPara(jogo, nx, ny) {
		// Restaura o elemento sobre o qual o personagem estava
		jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado
		// Guarda o elemento para o qual o personagem vai se mover
		jogo.UltimoVisitado = jogo.Mapa[ny][nx]
		// Atualiza a posição do personagem
		jogo.PosX, jogo.PosY = nx, ny

		// --- NOVA LÓGICA DE INTERAÇÃO ---
		// Verifica se o personagem pisou em uma armadilha armada.
		if jogo.UltimoVisitado.simbolo == ArmadilhaArmada.simbolo {
			jogo.StatusMsg = "Cuidado! Voce ativou uma armadilha!"
		}
	}
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
func personagemInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		return false
	case "interagir":
		personagemInteragir(jogo)
	case "mover":
		// Limpa a mensagem de status antes de mover
		jogo.StatusMsg = ""
		personagemMover(ev.Tecla, jogo)
	}
	return true
}