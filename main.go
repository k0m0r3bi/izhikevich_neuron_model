package main

import (
	"fmt"
	"math/rand"
	"math"
	// "sync"
	"time"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
)

var nNeurons int  = 4
var neurons 	  = make([]*NIzkvch, nNeurons)

var nSynapses int = 2
var synapses 	  = make([]*Synapse, nSynapses)

const plankTime 	time.Duration 	= 100 * time.Millisecond
const latencyUpdate time.Duration 	= 20  * time.Millisecond
const simDuration 	time.Duration 	= 30  * time.Second

var RS NIzkvchParams = NIzkvchParams{						// regular spiking
	a:    float64(0.02),
	b:    float64(-0.1),
	// b:  float64(0.1),
	c:    float64(-65 + 15 * math.Pow(rand.Float64(), 2)),
	d:    float64(8 - 6 * math.Pow(rand.Float64(), 2)),
	sOut: float64(-1.5 * rand.Float64()),
}

var FS NIzkvchParams = NIzkvchParams{						// fast spiking
	a:    float64(0.02 + 0.08 * rand.Float64()),
	b:    float64(0.25 + 0.05 * rand.Float64()),
	c:    float64(-65.0),
	d:    float64(2.0),
	sOut: float64(-rand.Float64()),
}

var IB NIzkvchParams = NIzkvchParams{						// intrisically bursting
	a:    float64(0.02 + 0.08 * rand.Float64()),
	b:    float64(0.25 + 0.05 * rand.Float64()),
	c:    float64(-55.0),
	d:    float64(4.0),
	sOut: float64(1.2 * rand.Float64()),
}

var CH NIzkvchParams = NIzkvchParams{						// chattering
	a:    float64(0.02 + 0.08 * rand.Float64()),
	b:    float64(0.25 + 0.05 * rand.Float64()),
	c:    float64(-50.0),
	d:    float64(2.0),
	sOut: float64(0.5 * rand.Float64()),
}

var LTS NIzkvchParams = NIzkvchParams{						// low threshhold spiking
	a:    float64(0.02),
	b:    float64(0.25),
	c:    float64(-65 + 15 * math.Pow(rand.Float64(), 2)),
	d:    float64(8 - 6 * math.Pow(rand.Float64(), 2)),
	sOut: float64(0.5 * rand.Float64()),
}

var RZ NIzkvchParams = NIzkvchParams{						// rezonator
}



type NIzkvchParams struct {
	a  		float64 		// describes the time scale of the recovery variable u. Smaller values result in slower recovery. A typical value is a = 0.02
	b  		float64 		// describes the sensitivity of the recovery variable u to the subthreshold fluctuations of the membrane potential v. typical is b = 0:2
	c  		float64 		// describes the after-spike reset value of the membrane potential v caused by the fast high-threshold K+ conductances. A typical value is c = -65 mV
	d  		float64 		// describes after-spike reset of the recovery variable u caused by slow high-threshold Na+ and K+ conductances. A typical value is d = 2
	sOut 		float64 		// spike out value
}
 

type NIzkvch struct {
	id 		int
	p  		NIzkvchParams
	v  		float64 		// membrane potential
	u  		float64 		// recovery variable
	in 		chan float64
	buffer  	chan float64
	out 		chan float64
	history 	[]float64
}


var restV float64  	= -65.0
var restU float64  	= 0.2 * -65
var restH []float64 	= []float64{-65.0, -65.0}

func NewNIzkvch(id, t int) (n *NIzkvch) {
	// http://www.izhikevich.org/publications/spikes.htm
	switch {
	case t == 0:
		return &NIzkvch{	
			id:		id,
			v:  	 restV,
			p:		 RS,
			u:  	 restU,
			in: 	 make(chan float64, 200),
			buffer:  make(chan float64, 2),
			out: 	 make(chan float64, 200),
			history: restH,
		}
	case t == 1:
		return &NIzkvch{
			id:	 id,
			p:	 FS,
			v:  	 restV,
			u:  	 restU,
			in: 	 make(chan float64, 200),
			buffer:  make(chan float64, 2),
			out: 	 make(chan float64, 200),
			history: restH,
		}
	case t == 2:
		return &NIzkvch{
			id:	 id,
			p:	 LTS,
			v:  	 restV,
			u:  	 restU,
			in: 	 make(chan float64, 200),
			buffer:  make(chan float64, 200),
			out: 	 make(chan float64, 200),
			history: restH,
		}
	case t == 3:
		return &NIzkvch{
			id:	 id,
			p:	 IB,
			v:  	 restV,
			u:  	 restU,
			in: 	 make(chan float64, 200),
			buffer:  make(chan float64, 200),
			out: 	 make(chan float64, 200),
			history: restH,
		}
	}
	return
}

func (n *NIzkvch) Activate() {
	// go n.Listen()
	go n.React()
}


func (n *NIzkvch) Listen() {
	buf := float64(0)
	clk := time.NewTicker(20*time.Millisecond)
	for {
		select {
		case <-clk.C:
			n.buffer <- buf
			buf = 0
		case spike := <- n.in:
			buf += spike
		}
	}
}



func (n *NIzkvch) React() {
	for {
		I := <- n.in
		n.v = math.Min(30.0, n.v + 20*(0.04*math.Pow(n.v, 2) + 4.1*n.v + 108 - n.u + I))
		n.u = n.p.a*(n.p.b*n.v - n.u)
		switch {
		case n.v == 30:
			n.out <- n.p.sOut
			n.history = append(n.history, n.p.sOut)
			n.v = n.p.c
			n.u = n.u + n.p.d
		default:
			n.history = append(n.history, n.v)
		}
	}
}

type Synapse struct {
	in 		chan float64
	w		float64
	out 	chan float64
}

func NewSynapse(in chan float64, w float64, out chan float64) (n *Synapse) {
	return &Synapse{
		in:  in,
		w:	 w,
		out: out,
	}
}

func (s *Synapse) Transmit() {
	for {
		spike := <- s.in
		s.out <- s.w * spike
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())


	plank := time.NewTicker(plankTime)
	done := make(chan bool)


	neurons[0] = NewNIzkvch(0, 0)
	neurons[0].Activate()
	neurons[1] = NewNIzkvch(1, 0)
	neurons[1].Activate()
	neurons[2] = NewNIzkvch(2, 2)
	neurons[2].Activate()
	neurons[3] = NewNIzkvch(3, 3)
	neurons[3].Activate()


	// synapses[0] = NewSynapse(neurons[0].out, -0.2, neurons[2].in)	
	// go synapses[0].Transmit()
	// synapses[1] = NewSynapse(neurons[1].out, 3.0, neurons[3].in)	
	// go synapses[1].Transmit()


    
    go func() {
        for {
            select {
            case <- done:
                return
            case <- plank.C:
            	for i := 0; i < len(neurons); i++ {
	            	switch {
	            	case rand.Float64() < 0.5:
	            		switch {
	            		case rand.Float64() < 0.6:
	            			neurons[i].in <- -0.4 *rand.Float64()
	            		default:
	            			neurons[i].in <- 0.8 *rand.Float64()
	            		}
	            	default:
	            		switch {
	            		case rand.Float64() < 0.3:
	            			neurons[i].in <- -1.8 * rand.Float64()
	            		default:
	            			neurons[i].in <- 1.3 * rand.Float64()
	            		}
	            	}
              	}
            }
        }
    }()

    Plotting(neurons, plank)
    time.Sleep(50 * time.Second)
    done <- true
}




func Plotting(neurons []*NIzkvch, tic *time.Ticker) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Title = "Press q to quit"
	p.Text = "Neuromorphic network"
	p.SetRect(0, 0, 50, 5)
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	widg := make([]*widgets.Plot, len(neurons))
	height := 25
	width := 200
	x0 := 0
	for i, _:= range neurons {

		widg[i] = widgets.NewPlot()
		widg[i].Title = fmt.Sprintf("Neuron_%v", i)
		widg[i].Data = make([][]float64, 1)
		switch {
		case len(neurons[i].history) < 50:
			widg[i].Data[0] = neurons[i].history[:len(neurons[i].history)-1]
		default:
			widg[i].Data[0] = neurons[i].history[len(neurons[i].history)-51:len(neurons[i].history)-1]
		}

		widg[i].SetRect(x0, x0+height*i, x0+width, x0+height*(i+1))
		widg[i].AxesColor = ui.ColorWhite
		widg[i].LineColors[0] = ui.ColorRed
	}

	draw := func() {
		for i, w := range widg {
			w.Data[0] = neurons[i].history
			ui.Render(w)
		}
	}

	draw()
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-tic.C:
			draw()
		}
	}
	return
}
