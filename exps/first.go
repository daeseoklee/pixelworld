package exps

import (
	"fmt"
	"time"

	"github.com/daeseoklee/pixelworld/colour"
	"github.com/daeseoklee/pixelworld/world"
)

func one(n, i int) []int {
	l := make([]int, n)
	for j := 0; j < n; j++ {
		if j == i%n {
			l[j] = 1
		} else {
			l[j] = 0
		}
	}
	return l
}

func all(n int) []int {
	l := make([]int, n)
	for j := 0; j < n; j++ {
		l[j] = 1
	}
	return l
}

//DoFirst : ~
func DoFirst(iniNut int, kname, mname string, from, duration, limit, savePeriod int, train bool) {
	go First(iniNut, kname, mname, from, duration, limit, savePeriod, train)
	time.Sleep(2000 * time.Millisecond)
	animate()
}

//First : ~
func First(iniNut int, kname, mname string, from, duration, limit, savePeriod int, train bool) {

	var exitCode int
	var newW *world.World
	//var err error
	cn := &world.Connector{}
	world.ConfigureConnector(cn)
	if train {
		w = world.Construct(iniNut, "first", kname, mname, 0, from-1, from, all(4))
		setVars()
		fmt.Println("m,n,k:", m, n, k)
		im := colour.ToImage(w.SnapShot, 1)
		fmt.Println(im.Bounds())
		//time.Sleep(1000 * time.Millisecond)
		//go animate()
		w.TrainMode()
		exitCode = w.Run(cn, limit, savePeriod)
		fmt.Println("exitCode: ", exitCode)
		fmt.Println("last moment: ", w.Moment)
		for i := 1; i < duration; i++ {
			newW = world.Construct(iniNut, "first", kname, mname, i, from+i-1, from+i, all(4))
			time.Sleep(2000 * time.Millisecond)
			w = newW
			w.DoSave()
			w.TrainMode()
			exitCode = w.Run(cn, limit, savePeriod)
			fmt.Println("exitCode: ", exitCode)
			fmt.Println("last moment: ", w.Moment)
		}
	} else {
		w = world.Construct(iniNut, "first", kname, mname, 0, from-1, from, all(4))
		setVars()
		w.DontSave()
		w.TestMode()
		exitCode = w.Run(cn, limit, savePeriod)
		fmt.Println("exitCode: ", exitCode)
		fmt.Println("last moment: ", w.Moment)
	}

}
