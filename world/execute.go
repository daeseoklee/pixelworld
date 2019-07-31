package world

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/minion"
	"github.com/daeseoklee/pixelworld/pos"
)

//import _ "image/png" //to open png files

var (
	wd, _    = os.Getwd()
	based    = findBase(wd)
	mapd     = filepath.Join(based, "_data/maps")
	kingdomd = filepath.Join(based, "_data/kingdoms")
)

func findBase(d string) string {
	var last string
	for (len(d) >= 5) && (last != "pixelworld") {
		last = filepath.Base(d)
		d = filepath.Dir(d)
	}
	if last == "pixelworld" {
		return filepath.Join(d, last)
	}
	log.Fatal(errors.New("calling execute.go outside /pixelworld"))
	return ""
}

func parseInt(s string) (int, error) {
	t := strings.Trim(s, " \n\r")
	return strconv.Atoi(t)
}

func parseFloat(s string) (float64, error) {
	t := strings.Trim(s, " \n\r")
	return strconv.ParseFloat(t, 64)
}
func readKingdom(kingdomname, mapname string) (names []string, traits []*minion.Trait, nums []int, phs [][]float64) {
	bytes, err := ioutil.ReadFile(filepath.Join(kingdomd, kingdomname+".txt"))
	if err != nil {
		panic(err)
	}
	r := string(bytes)
	lines := strings.Split(r, "\n")
	var numTrait int

	for i := 0; i < len(lines); i++ {
		if len(lines[i]) >= 7 && lines[i][:7] == "items :" {
			numTrait, err = parseInt(lines[i+1])
			if err != nil {
				panic(err)
			}
			fmt.Println("numTrait:", numTrait)
			names = make([]string, numTrait)
			traits = make([]*minion.Trait, numTrait)
			nums = make([]int, numTrait)
			var xlen, ylen, n int
			var taste, a float64
			var bodyColour, geniColour colour.Colour
			var err error
			var R, G, B float64
			var line string
			var segs []string
			for j := 0; j < numTrait; j++ {
				segs := strings.Split(lines[i+3+j], "/")
				names[j] = segs[1][1 : len(segs[1])-1]
				size := strings.Split(segs[2], ",")
				xlen, err = parseInt(size[0])
				if err != nil {
					panic(err)
				}
				ylen, err = parseInt(size[1])
				if err != nil {
					panic(err)
				}
				taste, err = parseFloat(segs[3])
				if err != nil {
					panic(err)
				}
				body := strings.Split(segs[4], ",")
				R, err = parseFloat(body[0])
				if err != nil {
					panic(err)
				}
				G, err = parseFloat(body[1])
				if err != nil {
					panic(err)
				}
				B, err = parseFloat(body[2])
				if err != nil {
					panic(err)
				}
				bodyColour = colour.Colour{M: 0, R: R / 255, G: G / 255, B: B / 255}
				geni := strings.Split(segs[5], ",")
				R, err = parseFloat(geni[0])
				if err != nil {
					panic(err)
				}
				G, err = parseFloat(geni[1])
				if err != nil {
					panic(err)
				}
				B, err = parseFloat(geni[2][:len(geni[2])-1])
				if err != nil {
					panic(err)
				}
				geniColour = colour.Colour{M: 0, R: R / 255, G: G / 255, B: B / 255}
				appear := minion.ConstructAppear(xlen, ylen, bodyColour, geniColour)
				traits[j] = minion.ConstructTrait(appear, taste)
				switch mapname[:3] {
				case "tin":
					n, err = parseInt(segs[6])
					if err != nil {
						panic(err)
					}
				case "sma":
					n, err = parseInt(segs[7])
					if err != nil {
						panic(err)
					}
				case "med":
					n, err = parseInt(segs[8])
					if err != nil {
						panic(err)
					}
				case "big":
					n, err = parseInt(segs[9])
					if err != nil {
						panic(err)
					}
				case "lar":
					n, err = parseInt(segs[10])
					if err != nil {
						panic(err)
					}
				}
				nums[j] = n
			}
			phs = make([][]float64, 30)
			for j := 0; j < 30; j++ {
				phs[j] = make([]float64, 50)
				line = lines[i+3+len(traits)+j]
				segs = strings.Split(line, "/")
				for k := 0; k < 50; k++ {
					a, err = parseFloat(segs[k])
					if err != nil {
						panic(err)
					}
					phs[j][k] = a
				}
			}

		}
	}
	return names, traits, nums, phs
}

//Construct : Load from pixelworld/_data/maps and construct World
func Construct(iniNut int, title, kingdomname, mapname string, epoch, weightfrom, saveat int, dolearn []int) *World {
	f, err := os.Open(filepath.Join(mapd, mapname+".png"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	ima, err := png.Decode(f) //ima : image.Image(*image.NRGBA)
	if err != nil {
		panic(err)
	}
	im := ima.(*image.NRGBA)
	Xlen := im.Bounds().Max.X - im.Bounds().Min.X
	Ylen := im.Bounds().Max.Y - im.Bounds().Min.Y
	count := 0
	var r, g, b uint8
	for i := 0; i < Xlen; i++ {
		for j := 0; j < Ylen; j++ {
			r = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4]
			g = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4+1]
			b = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4+2]
			if !(r == 0 && g == 0 && b == 0) {
				count++
			}
		}
	}
	shape := make([]pos.Rel, count)
	index := 0
	render := make(map[pos.Rel]colour.Colour)
	for i := 0; i < Xlen; i++ {
		for j := 0; j < Ylen; j++ {
			r = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4]
			g = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4+1]
			b = im.Pix[(j-im.Rect.Min.Y)*im.Stride+(i-im.Rect.Min.X)*4+2]
			if !(r == 0 && g == 0 && b == 0) {
				shape[index] = pos.Rel{Z: i, W: j}
				index++
				render[pos.Rel{Z: i, W: j}] = colour.Colour{M: 0, R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}
			}
		}
	}
	wall := &wall{shape: shape, render: render, visible: false}

	min := make(map[pos.Abs]*Mineral)
	nut := make(map[pos.Abs]int)
	snapShot := make([][]colour.Colour, Xlen)
	for i := 0; i < Xlen; i++ {
		snapShot[i] = make([]colour.Colour, Ylen)
	}
	outer := &outer{Xlen: Xlen, Ylen: Ylen}
	w := &World{
		Xlen:        Xlen,
		Ylen:        Ylen,
		Moment:      0,
		Outer:       outer,
		Objects:     make(map[Obj]bool),
		Minions:     make(map[*minion.Minion]bool),
		Traits:      nil,
		OccupiedBy:  make(map[pos.Abs]Obj),
		Loc:         make(map[Obj]pos.Abs),
		Head:        make(map[Obj]pos.Direc),
		Min:         min,
		Nut:         nut,
		SnapShot:    snapShot,
		ID:          0,
		Title:       title,
		Kingdomname: kingdomname,
		Mapname:     mapname,
		Epoch:       epoch,
		Weightfrom:  weightfrom,
		Saveat:      saveat,
		Dolearn:     dolearn,
		Save:        false,
		Train:       true,
	}
	for _, p := range w.Poss() {
		w.Min[p] = &Mineral{p: 0, a: 0}
		w.Nut[p] = iniNut
	}
	w.Register(outer, pos.Abs{X: 0, Y: 0}, pos.DirecFromNum(0), true)
	w.Register(wall, pos.Abs{X: 0, Y: 0}, pos.DirecFromNum(0), true)
	names, traits, nums, phs := readKingdom(kingdomname, mapname)
	SetPheros(phs)
	fmt.Println("names:", names)
	w.Traits = traits
	for i := 0; i < len(traits); i++ {
		for j := 0; j < nums[i]; j++ {
			mi := minion.Construct(traits[i])
			if w.LocateAndRegister(mi) {
				if (w.Head[mi] == pos.Direc{}) {
					panic(errors.New("here"))
				}
			}
		}
	}
	w.Render()
	return w
}

//DoSave : w.Save=true
func (w *World) DoSave() {
	w.Save = true
}

//DontSave : w.Save=false
func (w *World) DontSave() {
	w.Save = false
}

//TrainMode : w.Train=true
func (w *World) TrainMode() {
	w.Train = true
}

//TestMode : w.Train=false
func (w *World) TestMode() {
	w.Train = false
}

//Run : run until limit, exitCode - 0:limit achieved, 1:all died
func (w *World) Run(cont Controller, limit, savePeriod int) (exitCode int) { //exitCode - 0:limit achieved, 1:all died
	var err error
	for i := 0; i < limit; i++ {
		if i%savePeriod == 0 {
			w.DoSave()
		} else {
			w.DontSave()
		}
		err = w.Proceed(cont)
		if err != nil {
			panic(err)
		}
		if len(w.Minions) == 0 {
			return 1
		}
	}
	return 0
}

//Proceed : one step forward
func (w *World) Proceed(cont Controller) error {
	if w == nil {
		panic(errors.New("proceeded with invalid world setting"))
	}
	fmt.Println("moment :", w.Moment)
	w.pregnancyRatio()
	w.Render()
	for mi := range w.Minions {
		w.writeThird(mi)
	}
	cont.Control(w) //learning happend based on states so far, and new behavior has set
	for mi := range w.Minions {
		w.writePleasure(mi) //set "pleasure" to default. While execution, it will be adjusted
	}
	for mi := range w.Minions {
		w.Jaw(mi, mi.GetBehavior().JawMoveTy)
	}
	for mi := range w.Minions {
		b := mi.GetBehavior()
		if w.Minions[mi] {
			fmt.Println(b.MoveTy, b.MoveDist, b.JawMoveTy, b.Emission)
			//fmt.Println(mi.Triple.After.Smell[0:3])
			w.Move(mi, b.MoveTy, b.MoveDist, true, mi.Mass(), mi.Mass())
		}
	}
	for mi := range w.Minions {
		w.Excrete(mi, mi.Area())
	}
	if w.Moment%Period == 0 {
		w.Conversion()
	}
	w.Moment++
	for mi := range w.Minions {
		w.IncreaseAge(mi)
	}
	ok, loc, obj := w.CheckCompatible()
	if !ok {
		fmt.Println(loc, obj)
		panic(errors.New("something's wrong"))
	}
	for mi := range w.Minions {
		w.moveToFirst(mi)
	}
	return nil
}

//debugging-----------------------------------------------------------------

//CheckCompatible : w.Loc,w.Head,w.OccupiedBy are compatible
func (w *World) CheckCompatible() (bool, pos.Abs, Obj) {
	for obj := range w.Objects {
		for _, loc := range w.Occupying(obj) {
			if w.OccupiedBy[loc] != obj {
				return false, loc, obj
			}
		}
	}
	return true, pos.Abs{}, nil
}

//------------------------------------------------------------------------------
func (w *World) pregnancyRatio() {
	num, total := 0, 0
	for minion := range w.Minions {
		total++
		if minion.Pregnant() {
			num++
		}
	}
	//fmt.Printf("pregnant : %v out of %v out of %v\n", num, total, w.ID)
}
