// jogo.go - Funções para manipular os elementos do jogo...
package main

import (
	"bufio"
	"math"
	"os"
	"time"
)

// Coords é uma struct simples para representar coordenadas X, Y.
type Coords struct {
	x, y int
}

type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

// Jogo contém o estado atual do jogo. Adicionamos uma lista de patrulheiros.
type Jogo struct {
	Mapa           [][]Elemento
	PosX, PosY     int
	UltimoVisitado Elemento
	StatusMsg      string
	lock           chan struct{}
	patrulheiros   []*Patrulheiro // Lista de todos os patrulheiros ativos no jogo.
}

var (
	Personagem         = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo            = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede             = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao          = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio              = Elemento{' ', CorPadrao, CorPadrao, false}
	ArmadilhaArmada    = Elemento{'*', CorVermelho, CorPadrao, false}
	ArmadilhaDesarmada = Elemento{'o', CorCinzaEscuro, CorPadrao, false}
)

func jogoNovo() Jogo {
	j := Jogo{
		UltimoVisitado: Vazio,
		lock:           make(chan struct{}, 1),
		patrulheiros:   make([]*Patrulheiro, 0), // Inicializa a lista de patrulheiros.
	}
	return j
}
func (j *Jogo) Travar()    { j.lock <- struct{}{} }
func (j *Jogo) Destravar() { <-j.lock }

// ... (código existente como jogoCarregarMapa, jogoPodeMoverPara, etc. continua aqui) ...
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()
	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y
			case ArmadilhaArmada.simbolo:
				e = ArmadilhaArmada
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	return scanner.Err()
}
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) || x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}
	return !jogo.Mapa[y][x].tangivel
}
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy
	elemento := jogo.Mapa[y][x]
	jogo.Mapa[y][x] = jogo.UltimoVisitado
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]
	jogo.Mapa[ny][nx] = elemento
}


// --- LÓGICA DOS PATRULHEIROS (Atualizada) ---

// Patrulheiro foi atualizado para incluir um canal de notificação e um estado de perseguição.
type Patrulheiro struct {
	x, y           int
	dx             int
	ultimoVisitado Elemento
	notificacaoCh  chan Coords // Canal exclusivo para este patrulheiro receber alertas.
	alvo           *Coords     // Ponteiro para as coordenadas do alvo. nil se não estiver perseguindo.
}

// rodarPatrulheiro foi reescrito para usar 'select' e ter dois comportamentos.
func rodarPatrulheiro(jogo *Jogo, p *Patrulheiro) {
	tickerMovimento := time.NewTicker(500 * time.Millisecond)
	defer tickerMovimento.Stop()

	for {
		select {
		case alvoCoords := <-p.notificacaoCh:
			// MENSAGEM RECEBIDA: O jogador está perto!
			// Atualiza o alvo do patrulheiro com as novas coordenadas.
			p.alvo = &alvoCoords

		case <-tickerMovimento.C:
			// HORA DE MOVER: Executa um passo de movimento.
			if p.alvo == nil {
				// MODO PATRULHA: Nenhum alvo definido, continua o movimento horizontal.
				moverPatrulhando(jogo, p)
			} else {
				// MODO PERSEGUIÇÃO: Move-se em direção ao alvo.
				moverPerseguindo(jogo, p)
			}
		}
	}
}

// moverPatrulhando contém a lógica original de movimento horizontal.
func moverPatrulhando(jogo *Jogo, p *Patrulheiro) {
	nx := p.x + p.dx // Calcula a próxima posição de patrulha.

	jogo.Travar()
	if jogoPodeMoverPara(jogo, nx, p.y) {
		// Move o patrulheiro
		jogo.Mapa[p.y][p.x] = p.ultimoVisitado
		p.ultimoVisitado = jogo.Mapa[p.y][nx]
		jogo.Mapa[p.y][nx] = Inimigo
		p.x = nx
	} else {
		p.dx *= -1 // Bateu, inverte a direção
	}
	jogo.Destravar()
}

// moverPerseguindo contém a nova lógica para se mover em direção a um alvo.
func moverPerseguindo(jogo *Jogo, p *Patrulheiro) {
	if p.x == p.alvo.x && p.y == p.alvo.y {
		// Já alcançou o alvo, volta a patrulhar.
		p.alvo = nil
		return
	}

	// Lógica de movimento simples: tenta reduzir a maior distância primeiro (horizontal ou vertical).
	dx, dy := 0, 0
	if math.Abs(float64(p.alvo.x-p.x)) > math.Abs(float64(p.alvo.y-p.y)) {
		// Anda na horizontal
		if p.alvo.x > p.x {
			dx = 1
		} else {
			dx = -1
		}
	} else {
		// Anda na vertical
		if p.alvo.y > p.y {
			dy = 1
		} else {
			dy = -1
		}
	}

	nx, ny := p.x+dx, p.y+dy

	jogo.Travar()
	if jogoPodeMoverPara(jogo, nx, ny) {
		// Move o patrulheiro
		jogo.Mapa[p.y][p.x] = p.ultimoVisitado
		p.ultimoVisitado = jogo.Mapa[ny][nx]
		jogo.Mapa[ny][nx] = Inimigo
		p.x = nx
		p.y = ny
	}
	jogo.Destravar()
}

// --- LÓGICA DAS ARMADILHAS (Existente) ---
type Armadilha struct {
	x, y   int
	armada bool
}

func rodarArmadilha(jogo *Jogo, a *Armadilha) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		jogo.Travar()
		a.armada = !a.armada
		if a.armada {
			jogo.Mapa[a.y][a.x] = ArmadilhaArmada
		} else {
			jogo.Mapa[a.y][a.x] = ArmadilhaDesarmada
		}
		jogo.Destravar()
	}
}