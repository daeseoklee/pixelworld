package world

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/util"
	"github.com/phayes/freeport"
)

//Controller : determine behavior, learn etc.
type Controller interface {
	Control(*World)
}

//Random control-------------------------------------------

//Stupid : random controller for test
type Stupid struct {
}

//Control : Stupid control
func (st Stupid) Control(w *World) {
	for mi := range w.Minions {
		moveTy := util.RandInt(8)
		var n int
		switch {
		case moveTy == 0:
			n = 0
		case moveTy <= 2:
			n = mi.SpGet()
		case moveTy <= 4:
			n = mi.SpGet() / 2
		default:
			n = 0
		}
		moveDist := util.RandInt(n + 1)
		jawMoveTy := util.RandInt(3)
		emission := util.RandInt(NumPhero)
		mi.SetBehavior(moveTy, moveDist, jawMoveTy, emission)
	}
}

//learning---------------------------------

func (w *World) writePleasure(mi *minion.Minion) {
	mi.Triple.Pleasure = 0.0
}

func (w *World) writeThird(mi *minion.Minion) {
	w.WriteVision(mi, mi.Triple.After.Vision)
	w.WriteSmell(mi, mi.Triple.After.Smell)
	mi.Triple.After.Hp = float64(mi.HpGet()) / float64(mi.HpMax())
	mi.Triple.After.Sp = float64(mi.SpGet()) / float64(mi.SpMax())
	if mi.Pregnant() {
		mi.Triple.After.Pregnant = 1
	} else {
		mi.Triple.After.Pregnant = 0
	}
}

func (w *World) moveToFirst(mi *minion.Minion) {
	for i := 0; i < len(mi.Triple.Before.Vision); i++ {
		for j := 0; j < len(mi.Triple.Before.Vision[0]); j++ {
			mi.Triple.Before.Vision[i][j] = mi.Triple.After.Vision[i][j]
		}
	}
	for i := 0; i < len(mi.Triple.Before.Smell); i++ {
		mi.Triple.Before.Smell[i] = mi.Triple.After.Smell[i]
	}
	mi.Triple.Before.Sp = mi.Triple.After.Sp
	mi.Triple.Before.Hp = mi.Triple.After.Hp
	mi.Triple.Before.Pregnant = mi.Triple.After.Pregnant
}

/**
type brain struct{

}

//PDQN : a P-DQN controller
type PDQN struct {
	brains
}

//Control : PDQN control
func (pd PDQN) Control() {

}
**/

//Encoding----------------------------------------------------------

type info struct {
	trait    int
	xlen     int
	ylen     int
	id       int
	pleasure float64 //sexual intercourse
	before   *minion.Input
	action   *minion.Behavior
	after    *minion.Input
}

//join : form a signal sequence based on sequence of segments. length <=2^24-1 assumed for each segment
func join(segs [][]byte) []byte {
	m := 1
	for _, seg := range segs {
		if len(seg) > 16000000 {
			panic("too long")
		}
		m += 4 + len(seg)
	}
	bs := make([]byte, m)
	i := 0
	for _, seg := range segs {
		bs[i] = 1
		writeInt(bs, len(seg), i+1, 3)
		for j, b := range seg {
			bs[i+4+j] = b
		}
		i += 4 + len(seg)
	}
	bs[i] = 0
	return bs
}

//writeInt
func writeInt(bs []byte, n, i, d int) int {
	var q, r int
	q = n
	for j := 0; j < d; j++ {
		q, r = q/256, q%256
		bs[i+d-1-j] = byte(r)
	}
	if q != 0 {
		log.Fatal("too big number")
	}
	return i + d
}

func writeFloat64(bs []byte, a float64, i int) int {
	//binary.BigEndian.PutUint64(bs[i:i+8], math.Float64bits(a))
	b := math.Float64bits(a)
	bs[i] = byte(b >> 56)
	bs[i+1] = byte(b >> 48)
	bs[i+2] = byte(b >> 40)
	bs[i+3] = byte(b >> 32)
	bs[i+4] = byte(b >> 24)
	bs[i+5] = byte(b >> 16)
	bs[i+6] = byte(b >> 8)
	bs[i+7] = byte(b)
	return i + 8
}

func writeByte(bs []byte, b byte, i int) int {
	bs[i] = b
	return i + 1
}

//bytesFromInfo  : convert Info to byte slice
func bytesFromInfo(inf info) []byte {
	bs := make([]byte, 2+4+1+1+8+2*(4*len(inf.before.Vision)*len(inf.before.Vision[0])+8*2*DimPhero+8+8+1)+(1+8+1+1))
	i := 0
	i = writeInt(bs, inf.trait, i, 2)
	i = writeInt(bs, inf.id, i, 4)
	i = writeInt(bs, inf.xlen, i, 1)
	i = writeInt(bs, inf.ylen, i, 1)
	i = writeFloat64(bs, reward(inf), i)
	//before
	for j := 0; j < 3*inf.xlen; j++ {
		for k := 0; k < 3*inf.ylen; k++ {
			i = writeByte(bs, byte(math.Floor(255*inf.before.Vision[j][k].M)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.before.Vision[j][k].R)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.before.Vision[j][k].G)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.before.Vision[j][k].B)), i)
			/**
			bs[i+4*(j*3*inf.ylen+k)] = byte(math.Floor(255 * inf.before.Vision[j][k].M))
			bs[i+4*(j*3*inf.ylen+k)+1] = byte(math.Floor(255 * inf.before.Vision[j][k].R))
			bs[i+4*(j*3*inf.ylen+k)+2] = byte(math.Floor(255 * inf.before.Vision[j][k].G))
			bs[i+4*(j*3*inf.ylen+k)+3] = byte(math.Floor(255 * inf.before.Vision[j][k].B))
			**/
		}
	}
	//i += 36 * inf.xlen * inf.ylen
	for j := 0; j < 2*DimPhero; j++ {
		i = writeFloat64(bs, inf.before.Smell[j], i)
	}
	i = writeFloat64(bs, inf.before.Hp, i)
	i = writeFloat64(bs, inf.before.Sp, i)
	i = writeByte(bs, byte(inf.before.Pregnant), i)
	//action
	i = writeInt(bs, inf.action.MoveTy, i, 1)
	//i = writeInt(bs, inf.action.MoveDist, i, 2)
	i = writeFloat64(bs, float64(inf.action.MoveDist)/float64(minion.SpMaxFromSize(inf.xlen, inf.ylen)), i)
	i = writeInt(bs, inf.action.JawMoveTy, i, 1)
	i = writeInt(bs, inf.action.Emission, i, 1)
	//after
	for j := 0; j < 3*inf.xlen; j++ {
		for k := 0; k < 3*inf.ylen; k++ {
			i = writeByte(bs, byte(math.Floor(255*inf.after.Vision[j][k].M)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.after.Vision[j][k].R)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.after.Vision[j][k].G)), i)
			i = writeByte(bs, byte(math.Floor(255*inf.after.Vision[j][k].B)), i)
			/**
			bs[i+4*(j*3*inf.xlen+k)] = byte(math.Floor(255 * inf.after.Vision[j][k].M))
			bs[i+4*(j*3*inf.xlen+k)+1] = byte(math.Floor(255 * inf.after.Vision[j][k].R))
			bs[i+4*(j*3*inf.xlen+k)+2] = byte(math.Floor(255 * inf.after.Vision[j][k].G))
			bs[i+4*(j*3*inf.xlen+k)+3] = byte(math.Floor(255 * inf.after.Vision[j][k].B))
			**/
		}
	}
	//i += 36 * inf.xlen * inf.ylen
	for j := 0; j < 2*DimPhero; j++ {
		i = writeFloat64(bs, inf.after.Smell[j], i)
	}
	i = writeFloat64(bs, inf.after.Hp, i)
	i = writeFloat64(bs, inf.after.Sp, i)
	i = writeByte(bs, byte(inf.after.Pregnant), i)
	return bs
}

//Encode : encode all necessary state information
func (w *World) Encode() []byte {
	segs := make([][]byte, 10+len(w.Minions))
	i := 0
	segs[i] = []byte(w.Title)
	i++
	segs[i] = []byte(w.Kingdomname)
	i++
	segs[i] = []byte(w.Mapname)
	i++
	segs[i] = make([]byte, 4)
	_ = writeInt(segs[i], w.Epoch, 0, 4)
	i++
	segs[i] = make([]byte, 4)
	_ = writeInt(segs[i], w.Weightfrom, 0, 4)
	i++
	segs[i] = make([]byte, 4)
	_ = writeInt(segs[i], w.Saveat, 0, 4)
	i++
	segs[i] = make([]byte, len(w.Dolearn))
	for j := 0; j < len(w.Dolearn); j++ {
		segs[i][j] = byte(w.Dolearn[j])
	}
	i++
	segs[i] = make([]byte, 5)
	_ = writeInt(segs[i], w.Moment, 0, 5)
	i++
	segs[i] = make([]byte, 1)
	if w.Save {
		_ = writeByte(segs[i], 1, 0)
	} else {
		_ = writeByte(segs[i], 0, 0)
	}
	i++
	segs[i] = make([]byte, 1)
	if w.Train {
		_ = writeByte(segs[i], 1, 0)
	} else {
		_ = writeByte(segs[i], 0, 0)
	}
	i++
	for mi := range w.Minions {
		for t := 0; t < len(w.Traits); t++ {
			if w.Traits[t] == mi.GetTrait() {
				segs[i] = bytesFromInfo(info{t, mi.Xlen(), mi.Ylen(), mi.GetID(), mi.Triple.Pleasure, mi.Triple.Before, mi.Triple.Action, mi.Triple.After})
			}
		}
		i++
	}
	return join(segs)
}

//connector---------------------------------------

//Connector : controller that communicates with the existing python program
type Connector struct {
	conn net.Conn
	cmd  *exec.Cmd
}

func (cn *Connector) runPythonPDQN(port int) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Command("python", filepath.Join(based, "/brain/pdqn/main.py"), strconv.Itoa(port))
	cn.cmd = cmd
	cn.cmd.Stdout = stdout
	cn.cmd.Stderr = stderr
	err := cn.cmd.Run()
	fmt.Println("python program exitted")
	if err != nil {
		fmt.Println("stdout :", string(stdout.Bytes()))
		fmt.Println("stderr :", string(stderr.Bytes()))
		panic(err)
	}
}

//ConfigureConnector : run the python part, set the communication channel
func ConfigureConnector(cn *Connector) {
	//var err error
	port, err := freeport.GetFreePort()
	fmt.Println("port:", port)
	if err != nil {
		panic(err)
	}
	go cn.runPythonPDQN(port)
	fmt.Println("python program running")
	cn.conn, err = net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	fmt.Println("right after connection:", cn.conn)
	if err != nil {
		panic(err)
	}
	fmt.Println("connected")
}

func readFloat64(bytes []byte, i int) (float64, int) {
	bits := binary.BigEndian.Uint64(bytes[i : i+8])
	return math.Float64frombits(bits), i + 8
}

func readInt(bytes []byte, i, n int) (int, int) {
	sum := 0
	for j := 0; j < n; j++ {
		sum = 256*sum + int(bytes[i+j])
	}
	return sum, i + n
}
func readByte(bytes []byte, i int) (int, int) {
	return int(bytes[i]), i + 1
}

func readToken(bytes []byte, i int) (bool, int, int) {
	b, i := readByte(bytes, i)
	switch b {
	case 1:
		n, i := readInt(bytes, i, 3)
		return true, n, i
	case 0:
		return false, 0, i + 1
	default:
		panic(errors.New("unexpected encoding"))
	}
}

func decode(bytes []byte) map[int]minion.Behavior {
	i := 0
	var id, xlen, ylen, moveTy, jawMoveTy, emission int
	var moveDist float64
	var b bool
	d := make(map[int]minion.Behavior)
	for b, _, i = readToken(bytes, i); b; b, _, i = readToken(bytes, i) {
		id, i = readInt(bytes, i, 4)
		xlen, i = readByte(bytes, i)
		ylen, i = readByte(bytes, i)
		moveTy, i = readByte(bytes, i)
		moveDist, i = readFloat64(bytes, i)
		jawMoveTy, i = readByte(bytes, i)
		emission, i = readByte(bytes, i)
		d[id] = minion.Behavior{MoveTy: moveTy, MoveDist: int(moveDist * float64(minion.SpMaxFromSize(xlen, ylen))), JawMoveTy: jawMoveTy, Emission: emission}
	}
	return d
}

//Communicate : send to python and receive message
func (cn Connector) Communicate(s string) string {
	fmt.Fprintf(cn.conn, s)
	b := make([]byte, 4)
	_, err := cn.conn.Read(b)
	if err != nil {
		fmt.Println("pythonError:", cn.cmd.Stderr)
		fmt.Println("pythonOut:", cn.cmd.Stdout)
		panic(err)
	}
	message := string(b)
	return message
}

//Control : Connector is a Controller
func (cn Connector) Control(w *World) {
	var err error
	var action minion.Behavior
	//send current inputs
	encoded := w.Encode()
	fmt.Println("encoded length:", len(encoded))
	if err = ioutil.WriteFile(filepath.Join(based, "brain/tmp/dear_python.txt"), encoded, 0644); err != nil {
		panic(err)
	}

	//receive
	switch message := cn.Communicate("sent"); message {
	case "sent":
		received, err := ioutil.ReadFile(filepath.Join(based, "brain/tmp/dear_go.txt"))
		if err != nil {
			panic(err)
		}
		actions := decode(received)
		for mi := range w.Minions {
			action = actions[mi.GetID()]
			mi.SetBehavior(action.MoveTy, action.MoveDist, action.JawMoveTy, action.Emission)
		}
	default:
		panic(errors.New("unexpected message"))
	}
}
