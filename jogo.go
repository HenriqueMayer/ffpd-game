// jogo.go - Funções para manipular os elementos do jogo...
package main

import (
	"bufio"
	"os"
	"time"
)

// ... (structs Jogo e Elemento continuam aqui) ...
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
}


// Adicionamos os novos elementos Armadilha
var (
	Personagem         = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo            = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede             = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao          = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio              = Elemento{' ', CorPadrao, CorPadrao, false}
	ArmadilhaArmada    = Elemento{'*', CorVermelho, CorPadrao, false}   // Armadilha ligada, não bloqueia
	ArmadilhaDesarmada = Elemento{'o', CorCinzaEscuro, CorPadrao, false} // Armadilha desligada, não bloqueia
)

// ... (funções jogoNovo, Travar, Destravar, etc. continuam aqui, sem alteração) ...
func jogoNovo() Jogo {
	j := Jogo{
		UltimoVisitado: Vazio,
		lock:           make(chan struct{}, 1),
	}
	return j
}
func (j *Jogo) Travar()    { j.lock <- struct{}{} }
func (j *Jogo) Destravar() { <-j.lock }
// ... (código existente) ...
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
			case ArmadilhaArmada.simbolo: // Reconhece o símbolo da armadilha no mapa
				e = ArmadilhaArmada
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	return scanner.Err()
}
// ... (código existente) ...
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


// --- LÓGICA DOS PATRULHEIROS (Existente) ---
type Patrulheiro struct {
	x, y           int
	dx             int
	ultimoVisitado Elemento
}
func rodarPatrulheiro(jogo *Jogo, p *Patrulheiro) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		nx := p.x + p.dx
		jogo.Travar()
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
}

// --- LÓGICA DAS ARMADILHAS (Nova) ---

// Armadilha define o estado de uma armadilha que pisca.
type Armadilha struct {
	x, y   int
	armada bool
}

// rodarArmadilha é a função principal para a goroutine de uma armadilha.
func rodarArmadilha(jogo *Jogo, a *Armadilha) {
	// Ticker mais lento para a armadilha piscar a cada 1.5 segundos.
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		jogo.Travar() // Trava o jogo para modificar o mapa.

		// Inverte o estado da armadilha.
		a.armada = !a.armada

		// Atualiza o símbolo no mapa de acordo com o novo estado.
		if a.armada {
			jogo.Mapa[a.y][a.x] = ArmadilhaArmada
		} else {
			jogo.Mapa[a.y][a.x] = ArmadilhaDesarmada
		}

		jogo.Destravar() // Libera o jogo.
	}
}