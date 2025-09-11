// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"os"
	"time" // Adicionado para usar o time.Ticker na lógica do patrulheiro
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
	lock           chan struct{}  // Canal para controlar o acesso concorrente ao estado do jogo (exclusão mútua).
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	j := Jogo{
		UltimoVisitado: Vazio,
		lock:           make(chan struct{}, 1),
	}
	return j
}

// Travar adquire o lock de exclusão mútua do jogo.
func (j *Jogo) Travar() {
	j.lock <- struct{}{}
}

// Destravar libera o lock de exclusão mútua do jogo.
func (j *Jogo) Destravar() {
	<-j.lock
}

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

// --- LÓGICA DOS PATRULHEIROS ---

// Patrulheiro define o estado de um inimigo que se move horizontalmente.
type Patrulheiro struct {
	x, y           int      // Posição atual no mapa.
	dx             int      // Direção do movimento horizontal (-1 para esquerda, 1 para direita).
	ultimoVisitado Elemento // O que estava na célula antes do patrulheiro ocupá-la.
}

// rodarPatrulheiro é a função principal para a goroutine de um único patrulheiro.
func rodarPatrulheiro(jogo *Jogo, p *Patrulheiro) {
	// Cada patrulheiro tem seu próprio ticker para decidir quando se mover.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		nx := p.x + p.dx // Calcula a próxima posição desejada.

		jogo.Travar() // Trava o jogo para poder ler e modificar o mapa com segurança.

		if jogoPodeMoverPara(jogo, nx, p.y) {
			// Movimentação válida.
			jogo.Mapa[p.y][p.x] = p.ultimoVisitado     // Restaura a posição antiga.
			p.ultimoVisitado = jogo.Mapa[p.y][nx]      // Guarda o que está na nova posição.
			jogo.Mapa[p.y][nx] = Inimigo               // Move o inimigo para a nova posição.
			p.x = nx                                   // Atualiza a coordenada interna do patrulheiro.
		} else {
			// Movimentação inválida (bateu numa parede), inverte a direção.
			p.dx *= -1
		}
		jogo.Destravar() // Destrava o jogo para outras goroutines.
	}
}