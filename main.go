package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/stianeikeland/go-rpio/v4"
)

//go:embed root.html
var rootHTML []byte

var (
	exportedPins map[int]rpio.Pin
)

type Config struct {
	Pins map[string]int
	Port int
}

type GpioRequest struct {
	Type  string
	Pin   int
	Value bool
}

type GpioResponse struct {
	Pin   int
	State bool
}

func (c *Config) serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(rootHTML); err != nil {
			log.Panicf("%v\n", err)
		}
	})
	http.HandleFunc("/gpio", func(w http.ResponseWriter, r *http.Request) {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		g := GpioRequest{}
		if err := json.Unmarshal(bb, &g); err != nil {
			log.Printf("ERROR::%v\n", err)
			return
		}
		switch g.Type {
		case "list":
			jp, err := json.Marshal(c.Pins)
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			if _, wErr := w.Write(jp); wErr != nil {
				log.Printf("%v\n", wErr)
				return
			}
		case "write":
			log.Printf("Request: set gpio %d to %v", g.Pin, g.Value)
			switch g.Value {
			case true:
				exportedPins[g.Pin].High()
			case false:
				exportedPins[g.Pin].Low()
			}
			if _, err := w.Write([]byte(fmt.Sprintf("response: set gpio %d to %v", g.Pin, g.Value))); err != nil {
				log.Printf("%v\n", err)
				return
			}
		case "read":
			res := GpioResponse{
				Pin: g.Pin,
				State: func() bool {
					switch exportedPins[g.Pin].Read() {
					case 1:
						return true
					case 0:
						return false
					}
					return false
				}(),
			}
			mjd, err := json.Marshal(res)
			if err != nil {
				log.Printf("%v\n", err)
			}
			if _, wErr := w.Write(mjd); wErr != nil {
				log.Printf("%v\n", wErr)
			}
		}
	})
	http.HandleFunc("/temp", func(w http.ResponseWriter, r *http.Request) {
		// TODO : return temp data from ds18b20's or dht22
	})
	log.Printf("Starting server on port %d", c.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil); err != nil {
		log.Printf("ERROR::http.ListenAndServe()::%v", err)
	}
}

func init() {
	if err := rpio.Open(); err != nil {
		panic(err)
	}
}

func (c *Config) addPin(pin int, label string) {
	log.Printf("adding pin %d as %s\n", pin, label)
	c.Pins[label] = pin
}

func (c *Config) saveConfig(configFilename string) {
	mj, err := c.marshalJson()
	if err != nil {
		log.Panicf("%v\n", err)
		return
	}
	if wErr := os.WriteFile(configFilename, mj, 0666); err != nil {
		log.Panicf("%v\n", wErr)
	}
}

func (c *Config) loadConfig(configFilename string) {
	fr, err := os.ReadFile(configFilename)
	if err != nil {
		log.Panicf("%v\n", err)
		return
	}
	if jErr := json.Unmarshal(fr, c); jErr != nil {
		log.Panicf("%v\n", jErr)
		return
	}
	fmt.Printf("loading config file: %s", configFilename)
	if pErr := c.printConfig(); pErr != nil {
		return
	}
}

func (c *Config) initGpio() {
	for s, i := range c.Pins {
		fmt.Printf("Exporting: %s %d\n", s, i)
		pin := rpio.Pin(i)
		pin.Output()
		pin.Low()
		exportedPins[i] = pin
	}
}

func (c *Config) printConfig() error {
	mj, err := c.marshalJson()
	if err != nil {
		return err
	}
	fmt.Println(string(mj))
	return nil
}

func (c *Config) marshalJson() ([]byte, error) {
	mj, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return nil, err
	}
	return mj, nil
}

func main() {
	c := Config{}
	c.Pins = make(map[string]int)
	configFilename := flag.String("config", "config.json", "filename of config")
	add := flag.Bool("add", false, "add pin")
	pin := flag.Int("pin", 0, "pin to add")
	label := flag.String("label", "default pin label", "pin label")
	c.Port = *flag.Int("port", 8729, "server port")
	flag.Parse()
	c.loadConfig(*configFilename)
	if *add {
		c.addPin(*pin, *label)
		if err := c.printConfig(); err != nil {
			log.Printf("%v\n", err)
		}
		c.saveConfig(*configFilename)
		return
	}
	exportedPins = make(map[int]rpio.Pin)
	c.initGpio()
	c.serve()
}
