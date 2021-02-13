package main

import (
	"log"
	"net"
)

type room struct {
	name    string
	members map[net.Addr]*client
}

func (r *room) broadcast(sender *client, msg string) {
	for addr, m := range r.members {
		if addr != sender.conn.RemoteAddr() {
			m.msg(msg)
		} else {
			m.msg("(You) " + msg)
			log.Printf("New message from: %s - %s in %s: %s", sender.nick,
				sender.conn.RemoteAddr().String(), sender.room.name, msg)
		}
	}
}
