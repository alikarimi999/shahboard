package stockfish

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Stockfish struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

func NewStockfish(path string) (*Stockfish, error) {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(stdoutPipe)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	sf := &Stockfish{
		cmd:    cmd,
		stdin:  stdin,
		stdout: reader,
	}

	// Initialize UCI
	sf.sendCommand("uci")
	sf.waitFor("uciok")
	return sf, nil
}

func (sf *Stockfish) sendCommand(cmd string) {
	_, _ = sf.stdin.Write([]byte(cmd + "\n"))
}

func (sf *Stockfish) waitFor(expected string) {
	for {
		line, _ := sf.stdout.ReadString('\n')
		if strings.Contains(line, expected) {
			return
		}
	}
}

func (sf *Stockfish) BestMove(fen string, depth int) (string, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	sf.sendCommand("ucinewgame")
	sf.sendCommand("isready")
	sf.waitFor("readyok")

	sf.sendCommand("position fen " + fen)
	sf.sendCommand(fmt.Sprintf("go depth %d", depth))

	for {
		line, err := sf.stdout.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
}

func (sf *Stockfish) BestMoveWithMultiPV(fen string, depth, multipv int) (string, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	sf.sendCommand("ucinewgame")
	sf.sendCommand("isready")
	sf.waitFor("readyok")

	sf.sendCommand(fmt.Sprintf("setoption name MultiPV value %d", multipv))
	sf.sendCommand("position fen " + fen)
	sf.sendCommand(fmt.Sprintf("go depth %d", depth))

	var moves []string

	for {
		line, err := sf.stdout.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "info depth") && strings.Contains(line, "multipv") {
			parts := strings.Split(line, " ")
			if len(parts) >= 22 && parts[2] == strconv.Itoa(depth) {
				move := parts[21]
				moves = append(moves, move)
			}
		}

		if strings.HasPrefix(line, "bestmove") {
			break
		}
	}

	randIndex := rand.Intn(len(moves))
	return moves[randIndex], nil
}
