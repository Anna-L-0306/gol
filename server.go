//package gol

package main

import (
	"flag"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/gol"
	
	"uk.ac.bris.cs/gameoflife/util"
	
)

const alive = 255
const dead = 0

type DistributedOperations struct{}


func mod(x, m int) int {
	return (x + m) % m
}

func calculateNeighbours(Width, Height, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, Height)][mod(x+j, Width)] == alive {
					
					neighbours++
					
				}
			}
		}
	}
	return neighbours
}

func calculateNextState(a, b int, width int, world [][]byte, res *gol.ResponseCal, turn int) [][]byte {
	newWorld := make([][]byte, b-a)
	for i := range newWorld {
		newWorld[i] = make([]byte,width)
	}
	
	
	for y := 0; y < b - a; y++ {
		for x := 0; x < width; x++ {
			neighbours := calculateNeighbours(width, b-a, x, y+a, world)
			
			
			
			if world[y+a][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
					// res.Events = append(res.Events,gol.CellFlipped{CompletedTurns:turn, Cell:util.Cell{X:x, Y:y}})
					res.X = append(res.X, x)
					res.Y = append(res.Y, y+a)
					res.Turn = append(res.Turn,turn)
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
					// res.Events = append(res.Events,gol.CellFlipped{CompletedTurns:turn, Cell:util.Cell{X:x, Y:y}})
					res.X = append(res.X, x)
					res.Y = append(res.Y, y+a)
					res.Turn = append(res.Turn,turn)
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

// func calculateNextState(a int, b int, width int, r gol.Resource, world [][]byte, events []gol.Event, turn int) [][]byte {

// 	newWorld := make([][]byte, b-a)
// 	for i := range newWorld {
// 		newWorld[i] = make([]byte, width)
// 	}
// 	for y := 0; y < b-a; y++ {
// 		for x := 0; x < width; x++ {
// 			neighbours := calculateNeighbours(r, x, y+a, world)
// 			if world[a+y][x] == alive {
// 				if neighbours == 2 || neighbours == 3 {
// 					newWorld[y][x] = alive

// 				} else {
// 					newWorld[y][x] = dead
// 					events = append(events, gol.CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y + a}})
// 				}
// 			} else {
// 				if neighbours == 3 {
// 					newWorld[y][x] = alive
// 					events = append(events, gol.CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y + a}})
// 				} else {
// 					newWorld[y][x] = dead
// 				}
// 			}
// 		}
// 	}
// 	return newWorld
// }

func makeMatrix(height, width int) [][]byte {
	matrix := make([][]byte, height)
	for i := range matrix {
		matrix[i] = make([]byte, width)
	}
	return matrix
}

func worker(a int, b int, width int, r gol.Resource, world [][]byte, out chan [][]byte,reply *gol.ResponseCal ,turn int) {
	
	imagePart := calculateNextState(a, b, width, world, reply, turn)
	
	out <- imagePart
//	wg.Done()
}

func (d *DistributedOperations) AliveCells(r gol.Resource, cellList *gol.ResponseAlive) (err error){
	alivecells := make([]util.Cell, 0)
	World := r.World
	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			if World[y][x] == alive {
				cell := util.Cell{X: x, Y: y}
				alivecells = append(alivecells, cell)
			}
		}
	}
	
	cellList.Alivecells = alivecells
	return
}

func (d *DistributedOperations) Calculate(r gol.Resource, reply *gol.ResponseCal) (err error){
	
	world := r.World
	reply.X = make([]int, 0)
	reply.Y = make([]int, 0)
	reply.Turn = make([]int, 0)
	workerHeight := r.Height /r.Threads
	for turn := 0; turn < r.Turns; turn++ {

		out := make([]chan [][]byte, r.Threads)
		for i := 0; i < r.Threads; i++ {
			out[i] = make(chan [][]byte)
		}

		//world = calculateNextState(r, world, reply, turn)
		for i := 0; i < r.Threads; i++ {
			if i <= r.Threads-2 {
				go worker(i*workerHeight, (i+1)*workerHeight, r.Width, r, world, out[i], reply, turn)
			} else {
				go worker(i*workerHeight, r.Height, r.Width, r, world, out[i], reply, turn)
			}

		}

		newWorld := makeMatrix(0, 0)

		for i := 0; i < r.Threads; i++ {
			part := <-out[i]

			newWorld = append(newWorld, part...)

		}
		world = newWorld
		reply.X = append(reply.X, -1)
		reply.Y = append(reply.Y, -1)
		reply.Turn = append(reply.Turn, turn)
		
	}
	
	reply.World =  world
	
	return
}

func main() {

	pAddr := flag.String("port", "9090", "Port to listen on")
	flag.Parse()
	// rand.Seed(time.Now().UnixNano())

	rpc.Register(&DistributedOperations{})
	
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()

	rpc.Accept(listener)

}
