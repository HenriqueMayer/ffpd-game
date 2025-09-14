// jogo.go - Funções para manipular os elementos do jogo e a lógica dos atores concorrentes.
package main

import (
	"bufio"
	"math"
	"os"
	"time"
)

type Coords struct{ x, y int }

type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

type Jogo struct {
	Mapa           [][]Elemento
	PosX, PosY     int
	UltimoVisitado Elemento
	StatusMsg      string
	lock           chan struct{}
	patrulheiros   []*Patrulheiro
	portais        []*Portal
	GameOver       bool
}

// Definição de todos os elementos visuais do jogo (Armadilha agora é estática).
var (
	Personagem     = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo        = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede         = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao      = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio          = Elemento{' ', CorPadrao, CorPadrao, false}
	Armadilha      = Elemento{'*', CorVermelho, CorPadrao, false} // Armadilha estática, sempre armada.
	PlacaDePressao = Elemento{'.', CorCinzaEscuro, CorPadrao, false}
	PortalFechado  = Elemento{'⬱', CorVerde, CorPadrao, true}
	PortalAberto   = Elemento{'O', CorVerde, CorPadrao, false}
)

func jogoNovo() Jogo {
	j := Jogo{
		UltimoVisitado: Vazio,
		lock:           make(chan struct{}, 1),
		patrulheiros:   make([]*Patrulheiro, 0),
		portais:        make([]*Portal, 0),
		GameOver:       false,
	}
	return j
}

func (j *Jogo) Travar()    { j.lock <- struct{}{} }
func (j *Jogo) Destravar() { <-j.lock }

func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil { return err }
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo: e = Parede
			case Inimigo.simbolo: e = Inimigo
			case Vegetacao.simbolo: e = Vegetacao
			case Personagem.simbolo: jogo.PosX, jogo.PosY = x, y
			case Armadilha.simbolo: e = Armadilha // Reconhece a armadilha estática
			case PlacaDePressao.simbolo: e = PlacaDePressao
			case PortalFechado.simbolo: e = PortalFechado
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

// --- LÓGICA DOS PATRULHEIROS ---

type Patrulheiro struct {
	x, y           int
	dx             int
	ultimoVisitado Elemento
	notificacaoCh  chan Coords
	alvo           *Coords
}

func rodarPatrulheiro(jogo *Jogo, p *Patrulheiro) {
	tickerMovimento := time.NewTicker(500 * time.Millisecond)
	defer tickerMovimento.Stop()
	for {
		select {
		case alvoCoords := <-p.notificacaoCh:
			p.alvo = &alvoCoords
		case <-tickerMovimento.C:
			if p.alvo == nil {
				moverPatrulhando(jogo, p)
			} else {
				moverPerseguindo(jogo, p)
			}
		}
	}
}

func moverPatrulhando(jogo *Jogo, p *Patrulheiro) {
	nx := p.x + p.dx
	jogo.Travar()
	if nx == jogo.PosX && p.y == jogo.PosY {
		jogo.StatusMsg = "Voce foi pego por um inimigo! Fim de jogo."
		jogo.GameOver = true
		jogo.Destravar()
		return
	}
	if jogoPodeMoverPara(jogo, nx, p.y) {
		jogo.Mapa[p.y][p.x] = p.ultimoVisitado
		p.ultimoVisitado = jogo.Mapa[p.y][nx]
		jogo.Mapa[p.y][nx] = Inimigo
		p.x = nx
	} else {
		p.dx *= -1
	}
	jogo.Destravar()
}

func moverPerseguindo(jogo *Jogo, p *Patrulheiro) {
	if p.x == p.alvo.x && p.y == p.alvo.y { p.alvo = nil; return }
	dx, dy := 0, 0
	if math.Abs(float64(p.alvo.x-p.x)) > math.Abs(float64(p.alvo.y-p.y)) {
		if p.alvo.x > p.x { dx = 1 } else { dx = -1 }
	} else {
		if p.alvo.y > p.y { dy = 1 } else { dy = -1 }
	}
	nx, ny := p.x+dx, p.y+dy
	jogo.Travar()
	if nx == jogo.PosX && ny == jogo.PosY {
		jogo.StatusMsg = "Voce foi pego por um inimigo! Fim de jogo."
		jogo.GameOver = true
		jogo.Destravar()
		return
	}
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogo.Mapa[p.y][p.x] = p.ultimoVisitado
		p.ultimoVisitado = jogo.Mapa[ny][nx]
		jogo.Mapa[ny][nx] = Inimigo
		p.x = nx
		p.y = ny
	}
	jogo.Destravar()
}

// --- LÓGICA DO PORTAL ---

type Portal struct {
	x, y       int
	ativacaoCh chan bool
}

func rodarPortal(jogo *Jogo, p *Portal) {
	for range p.ativacaoCh {
		jogo.Travar()
		jogo.Mapa[p.y][p.x] = PortalAberto
		jogo.Destravar()
		select {
		case <-time.After(5 * time.Second):
			jogo.Travar()
			if jogo.Mapa[p.y][p.x].simbolo == PortalAberto.simbolo {
				jogo.Mapa[p.y][p.x] = PortalFechado
			}
			jogo.Destravar()
		}
	}
}