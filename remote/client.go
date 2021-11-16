package remote

import (
	"CloudWars/core"
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
)

// TcpClient is an API to access a server
// and to remotely control a player cloud.
type TcpClient struct {
	conn *net.TCPConn
	tp   *textproto.Reader
	mux  *sync.Mutex
}

// NewTcpClient init a TcpClient
func NewTcpClient(host, port string) *TcpClient {

	// address
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("NewTcpClient: %v\n", err)
	}

	// connection
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("NewTcpClient: %v\n", err)
	}

	// config client
	t := &TcpClient{
		conn: conn,
		tp:   textproto.NewReader(bufio.NewReader(conn)),
		mux:  new(sync.Mutex),
	}

	// return client
	return t
}

//--------------------------------------------------------------------------------------------------------------------//

// Close disconnects from the server.
// The controlled cloud remains unchanged (use Kill() before this call).
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Close() string {
	t.mux.Lock()
	defer t.mux.Unlock()

	ret := comWriteRead(t, "quit")
	_ = t.conn.Close()
	return ret
}

// List returns the world status as a json string.
// Use core.World FromJson() to parse the string.
func (t *TcpClient) List() string {
	t.mux.Lock()
	defer t.mux.Unlock()

	return comWriteRead(t, "list")
}

// Name set the player name
// Use this before calling Play()
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Name(name string) string {
	t.mux.Lock()
	defer t.mux.Unlock()

	com := fmt.Sprintf("name%s", name)
	return comWriteRead(t, com)
}

// Color set the player color.
// 'blue', 'gray', 'orange', 'purple' or 'red'
// Use this before calling Play()
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Color(color string) string {
	t.mux.Lock()
	defer t.mux.Unlock()

	com := fmt.Sprintf("type%s", color)
	return comWriteRead(t, com)
}

// Play creates a new player cloud.
// The attributes of Name() and Color() are used.
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Play() string {
	t.mux.Lock()
	defer t.mux.Unlock()

	return comWriteRead(t, "play")
}

// Move sends a move command for your player cloud to the server.
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Move(v *core.Velocity) string {
	t.mux.Lock()
	defer t.mux.Unlock()

	if v == nil {
		return "err: nil"
	}
	com := fmt.Sprintf("move%f;%f", v.X, v.Y)
	return comWriteRead(t, com)
}

// Kill blasts the controlled cloud and removes it from the game.
// Returns the server response (OK or ERR) as a string.
func (t *TcpClient) Kill() string {
	t.mux.Lock()
	defer t.mux.Unlock()

	return comWriteRead(t, "kill")
}

//----- Helper -------------------------------------------------------------------------------------------------------//

func comWriteRead(t *TcpClient, com string) string {
	// remove protocol break
	com = strings.ReplaceAll(com, "\n", "")
	com = strings.ReplaceAll(com, "\r", "")
	// send command
	if comWrite(t.conn, com) {
		return "err"
	}
	// read response
	resp, err := t.tp.ReadLine()
	if err != nil {
		fmt.Printf("comWriteRead: %v\n", err)
	}
	return resp
}
