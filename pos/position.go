package pos

import (
	"errors"
	"fmt"
)

//Abs : struct for absolute position
type Abs struct {
	X int
	Y int
}

//Member : check whether it is in the given slice
func (pos Abs) Member(l []Abs) bool {
	for _, a := range l {
		if pos == a {
			return true
		}
	}
	return false
}

//ToAbs : given reference point, head direction and relative position, returns the absolute position
func ToAbs(ref Abs, head Direc, relPos Rel) Abs {
	switch head {
	case Direc{"n"}:
		return Abs{X: ref.X + relPos.Z, Y: ref.Y + relPos.W}
	case Direc{"w"}:
		return Abs{X: ref.X - relPos.W, Y: ref.Y + relPos.Z}
	case Direc{"s"}:
		return Abs{X: ref.X - relPos.Z, Y: ref.Y - relPos.W}
	case Direc{"e"}:
		return Abs{X: ref.X + relPos.W, Y: ref.Y - relPos.Z}
	default:
		panic(errors.New("invalid absolute direction"))
	}
}

//Rel : struct for relative position
type Rel struct {
	Z int
	W int
}

//ToMoveTy : given head direction and direction to proceed, return moveTy
func ToMoveTy(head, d Direc) int {
	m := (NumFromDirec(d) - NumFromDirec(head)) % 4
	if m < 0 {
		m += 4
	}
	switch m {
	case 0:
		return 1
	case 1:
		return 3
	case 2:
		return 2
	case 3:
		return 4
	default:
		panic(errors.New("impossible"))
	}
}

//Member : check whether it is in the given slice
func (pos Rel) Member(l []Rel) bool {
	for _, a := range l {
		if pos == a {
			return true
		}
	}
	return false
}

//ToRel : inverse of ToAbs on the last argument
func ToRel(ref Abs, head Direc, loc Abs) Rel {
	switch head {
	case Direc{"n"}:
		return Rel{Z: loc.X - ref.X, W: loc.Y - ref.Y}
	case Direc{"w"}:
		return Rel{W: ref.X - loc.X, Z: loc.Y - ref.Y}
	case Direc{"s"}:
		return Rel{Z: ref.X - loc.X, W: ref.Y - loc.Y}
	case Direc{"e"}:
		return Rel{W: loc.X - ref.X, Z: ref.Y - loc.Y}
	default:
		fmt.Println("wrong head:", head)
		panic(errors.New("invalid absolute direction"))
	}
}

//Direc : struct for absolute direction
type Direc struct {
	D string //"n","w","s","e"
}

//DirecFromNum : absolute direction from number
func DirecFromNum(n int) Direc {
	m := n % 4
	if m < 0 {
		m += 4
	}
	switch m {
	case 0:
		return Direc{"n"}
	case 1:
		return Direc{"w"}
	case 2:
		return Direc{"s"}
	case 3:
		return Direc{"e"}
	default:
		panic(errors.New("impossible"))
	}
}

//NumFromDirec : get absolute direction number
func NumFromDirec(d Direc) int {
	switch d {
	case Direc{"n"}:
		return 0
	case Direc{"w"}:
		return 1
	case Direc{"s"}:
		return 2
	case Direc{"e"}:
		return 3
	default:
		panic(errors.New("invalid direction"))
	}
}

//Rotate : absolute direction rotated
func Rotate(d Direc, angle int) Direc {
	return DirecFromNum(NumFromDirec(d) + angle)
}

//ToDirec : head direction and relative direction, return the absolute direction
func ToDirec(head Direc, relDirec RelDirec) Direc {
	a := NumFromDirec(head)
	var b int
	switch NumFromRelDirec(relDirec) {
	case 1:
		b = 0
	case 2:
		b = 2
	case 3:
		b = 1
	case 4:
		b = 3
	}
	return DirecFromNum(a + b)
}

//RelDirec : struct for relative direction
type RelDirec struct {
	D string //"f","b","l","r"
}

//RelDirecFromNum : relative direction from number
func RelDirecFromNum(n int) RelDirec {
	switch n {
	case 1:
		return RelDirec{"f"}
	case 2:
		return RelDirec{"b"}
	case 3:
		return RelDirec{"l"}
	case 4:
		return RelDirec{"r"}
	default:
		panic(errors.New("invalid movety for parallel move"))
	}
}

//NumFromRelDirec : get movety from relative direction
func NumFromRelDirec(d RelDirec) int {
	switch d {
	case RelDirec{"f"}:
		return 1
	case RelDirec{"b"}:
		return 2
	case RelDirec{"l"}:
		return 3
	case RelDirec{"r"}:
		return 4
	default:
		panic(errors.New("invalid relative direction"))
	}
}

//Next : position after one-step toward the direction
func Next(p Abs, d Direc) Abs {
	switch d {
	case Direc{"n"}:
		return Abs{X: p.X, Y: p.Y + 1}
	case Direc{"w"}:
		return Abs{X: p.X - 1, Y: p.Y}
	case Direc{"s"}:
		return Abs{X: p.X, Y: p.Y - 1}
	case Direc{"e"}:
		return Abs{X: p.X + 1, Y: p.Y}
	default:
		panic(errors.New("invalid absolute direction"))
	}
}

//Previous : position one-step backward in the direction
func Previous(p Abs, d Direc) Abs {
	return Next(p, DirecFromNum(NumFromDirec(d)+2))
}
