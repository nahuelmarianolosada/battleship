package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"strconv"
)

const (
	//TCP CONFIG
	connHost = "localhost"
	connPort = "8080"
	connType = "tcp"

	//Board dimensions
	maxY = 10
	maxX = 10

	//Attack status
	hit = "Hit"
	miss = "Miss"
	alreadyHit = "Already Hit"
	invalid = "Invalid"

	//ship position
	ship = "s"
	empty = ""
)

type Player struct {
	name, addr string
	conn net.Conn
}

var (
	board = [maxX][maxY] string {
		{ship, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty, empty, ship},
	}

	validCommands = map[string]func(net.Conn, []string) bool {
		"login":login,
		"logout":logout,
		"attack":attack,
	}

	player1, player2 *Player
)

func checkVictory() bool {
	nonShips := true
	for _, v := range board {
		for _, cel := range v {
			if cel == ship {
				nonShips = false
			}
		}
	}
	return nonShips
}

func login(conn net.Conn, command []string) bool {
	success := false
	if n := command[1]; n != "" {
		if player1 == nil {
			player1 = &Player {n, conn.RemoteAddr().String(), conn}
			log.Println("Ready player 1: ", player1)
			success = true
		} else if player2 == nil {
			player2 = &Player{n, conn.RemoteAddr().String(), conn}
			log.Println("Ready player 2: ", player2)
			success = true
		} else {
			log.Println("Ya están conectados los dos jugadores. Sin espacio para %v", n)
		}
	}

	return success
}

func logout(conn net.Conn, command []string) bool {
	success := false
	if n := command[1]; n != "" {
		if player1.name == n {
			player1.conn.Close()
			player1 = nil
			success = true
		} else if player2.name == n {
			player2.conn.Close()
			player2 = nil
			success = true
		} else {
			log.Println(fmt.Sprintf("No se encontró el jugador %s", n))
		}
	}
	if success {
		conn.Close()
	}
	return success
}

func attack(conn net.Conn, command []string) bool {
	result := ""
	playerWhoAttack := ""
	if conn.RemoteAddr().String() == player1.addr {
		playerWhoAttack = player1.name
	} else if conn.RemoteAddr().String() == player2.addr {
		playerWhoAttack = player2.name
	}

	if xStr := command[1]; xStr != "" {
		if yStr := command[2]; yStr != "" {
			x, _ := strconv.Atoi(xStr)
			y, _ := strconv.Atoi(yStr)


			if x > maxX || y > maxY || x < 0 || y < 0 {
				result = invalid
			} else {
				switch board[x][y] {
					case ship:
						board[x][y] = alreadyHit
						result = hit
					case empty: result = miss
					case hit: result = alreadyHit
					case alreadyHit: result = alreadyHit
				}
			}
			msg := fmt.Sprintf("%s:%s %s %s \n", playerWhoAttack, result, xStr, yStr)
			log.Println(msg)
			if player1.conn != nil {
				player1.conn.Write([]byte(msg))
			}

			if player2.conn != nil {
				player2.conn.Write([]byte(msg))
			}

		}
	}

	if checkVictory() {
		conn.Write([]byte("Game won by " + playerWhoAttack))
		logout(conn, []string{"logout", player1.name})
		logout(conn, []string{"logout", player2.name})
	}
	return false
}

func main() {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	data := make(chan string)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client connected.")

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		go handleConnection(c, data)
	}
}

func handleConnection(conn net.Conn, data chan string) {
	buffer, err := bufio.NewReader(conn).ReadBytes('\n')

	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}

	clientMsg := string(buffer[:len(buffer)-1])
	clientMsgReq := strings.Split(clientMsg, " ")
	if isAValidRequest(clientMsgReq) {
		f := validCommands[clientMsgReq[0]]
		f(conn, clientMsgReq)
	}
	handleConnection(conn, data)
}

func isAValidRequest(req []string) bool {
	if len(req) > 3 { return false}
	if !isAValidCommand(req[0]) { return false }
	return true
}

func isAValidCommand(c string) bool {
	if _, ok := validCommands[strings.ToLower(c)]; ok {
		return true
	}
	return false
}

