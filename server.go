package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms    map[string]*room
	commands chan command
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args[1])
		case CMD_JOIN:
			s.join(cmd.client, cmd.args[1])
		case CMD_CREATE_ROOM:
			s.createRoom(cmd.client, cmd.args[1])
		case CMD_ROOMS:
			s.listRooms(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
		}
	}
}

func (s *server) newClient(conn net.Conn) {
	log.Printf("New client from: %s", conn.RemoteAddr().String())

	c := &client{
		conn:     conn,
		nick:     "Anon",
		commands: s.commands,
	}

	c.ReadInput()
}

func (s *server) nick(c *client, nick string) {
	c.nick = nick
	c.msg(fmt.Sprintf("Your nickname now is: %s", c.nick))
}

func (s *server) join(c *client, roomName string) {
	r, ok := s.rooms[roomName]
	if !ok {
		c.msg(fmt.Sprint("Wrong room name! Check /rooms or /createroom"))
		return
	}

	c.room.broadcast(c, fmt.Sprintf("%s has created the room: %s", c.nick, roomName))

	r.members[c.conn.RemoteAddr()] = c

	c.room = r

	c.room.broadcast(c, fmt.Sprintf("%s has joined the room", c.nick))
	c.msg(fmt.Sprintf("Welcome to %s", r.name))
}

func (s *server) createRoom(c *client, roomName string) {
	r := &room{
		name:    roomName,
		members: make(map[net.Addr]*client),
	}
	s.rooms[roomName] = r

	c.msg(fmt.Sprintf("You created new room:  %s", roomName))
}

func (s *server) listRooms(c *client) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	if len(rooms) != 0 {
		c.msg(fmt.Sprintf("Available rooms are: %s", strings.Join(rooms, ", ")))
	} else {
		c.msg(fmt.Sprint("There is no rooms. You can create room by /join 'room name'"))
	}

}

func (s *server) msg(c *client, args []string) {
	if c.room == nil {
		c.err(errors.New("you have to join room first"))
		return
	}

	c.room.broadcast(c, c.nick+": "+strings.Join(args[1:len(args)], " "))
}

func (s *server) quit(c *client) {
	log.Printf("Client has disconnected: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)

	c.msg("See you later")
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		delete(c.room.members, c.conn.RemoteAddr())
		c.room.broadcast(c, fmt.Sprintf("%s has left the room", c.nick))
	}
}
