// personagem.go - Funções para movimentação e ações do personagem
package main

func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w': dy = -1
	case 'a': dx = -1
	case 's': dy = 1
	case 'd': dx = 1
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	// Verificação de colisão com inimigo ANTES de mover.
	if ny >= 0 && ny < len(jogo.Mapa) && nx >= 0 && nx < len(jogo.Mapa[ny]) {
		if jogo.Mapa[ny][nx].simbolo == Inimigo.simbolo {
			jogo.StatusMsg = "Voce colidiu com um inimigo! Fim de jogo."
			jogo.GameOver = true
			return
		}
	}

	if jogoPodeMoverPara(jogo, nx, ny) {
		jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado
		jogo.UltimoVisitado = jogo.Mapa[ny][nx]
		jogo.PosX, jogo.PosY = nx, ny

		// Verifica se pisou em uma placa de pressão.
		if jogo.UltimoVisitado.simbolo == PlacaDePressao.simbolo {
			jogo.StatusMsg = "Voce ativou um mecanismo!"
			for _, portal := range jogo.portais {
				select {
				case portal.ativacaoCh <- true:
				default:
				}
			}
		}

		// Pisar na armadilha agora causa Game Over.
		if jogo.UltimoVisitado.simbolo == Armadilha.simbolo {
			jogo.StatusMsg = "Armadilha ativada! Fim de jogo."
			jogo.GameOver = true
			return
		}
	}
}

func personagemInteragir(jogo *Jogo) {
	adjacentes := []Coords{
		{jogo.PosX, jogo.PosY - 1}, {jogo.PosX, jogo.PosY + 1},
		{jogo.PosX - 1, jogo.PosY}, {jogo.PosX + 1, jogo.PosY},
	}
	portalEncontrado := false
	for _, coord := range adjacentes {
		if coord.y >= 0 && coord.y < len(jogo.Mapa) && coord.x >= 0 && coord.x < len(jogo.Mapa[coord.y]) {
			if jogo.Mapa[coord.y][coord.x].simbolo == PortalFechado.simbolo {
				for _, portal := range jogo.portais {
					if portal.x == coord.x && portal.y == coord.y {
						select {
						case portal.ativacaoCh <- true:
							jogo.StatusMsg = "Voce abriu um portal!"
							portalEncontrado = true
						default:
						}
						break
					}
				}
			}
		}
		if portalEncontrado { break }
	}
	if !portalEncontrado {
		jogo.StatusMsg = "Nao ha nada para interagir aqui."
	}
}

func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	jogo.StatusMsg = ""
	switch ev.Tipo {
	case "sair": return false
	case "interagir": personagemInteragir(jogo)
	case "mover": personagemMover(ev.Tecla, jogo)
	}
	return true
}