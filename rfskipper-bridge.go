package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tarm/serial"
)

var cnfSerialDev string
var cnfSerialBaudRate int
var cnfMqttURL *url.URL
var cnfLogFile string
var cnfIsDaemon bool
var logfile *os.File
var cnfTopicSignalRaw, cnfTopicSignalDecoded string

// doLog write message to log file
func doLog(format string, a ...interface{}) {
	t := time.Now()

	if cnfLogFile != "" && logfile == os.Stdout {
		fmt.Printf("INFO: Creating log file '%s'\n", cnfLogFile)
		tf, err := os.OpenFile(cnfLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("ERROR: Can't open log file '%s' for writing, logging to Stdout\n", cnfLogFile)
			fmt.Printf("ERROR: %s\n", err)
		} else {
			fmt.Printf("INFO: log file '%s' is ready for writing\n", cnfLogFile)
			logfile = tf
		}
	}

	// logfile default is os.Stdout
	fmt.Fprintf(logfile, "%s ", string(t.Format("20060102 150405")))
	fmt.Fprintf(logfile, format, a...)
}

func serialworker(sig chan string) {
	c := &serial.Config{Name: cnfSerialDev, Baud: cnfSerialBaudRate, ReadTimeout: time.Second * 1}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	n, err := s.Write([]byte("10;RFDEBUG=ON;\n"))
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 1)
	buf := make([]byte, 2048)
	/*
		20;188;DEBUG;Pulses=301;Pulses(uSec)=210,210,210,210,...,200,220,200,210;';
	*/
	// <PAYLOAD:46575645523A303130343B434150413A303030343B4241544C45563A323636313B4157414B455F5345433A303B4D41494E5F4C4F4F503A303B4552524F523A303031323B4800; SMAC:00:00:00:00:00:00:00:25;LQI:120;RSSI:61>
	r, _ := regexp.Compile(`(\d+);(\d+);DEBUG;Pulses=(\d+);Pulses\(uSec\)=([^;]+);`)
	doLog("Start reading UART, forever\n")
	for {
		n, _ = s.Read(buf)
		if n > 0 {
			buffer = append(buffer, buf[:n]...)
			//doLog("Serial: Read %d bytes [%s]\n", n, buffer)

			for {
				m := r.FindIndex(buffer)
				if m == nil {
					break
				}
				rs := r.FindSubmatch(buffer)
				pulses := "{id:" + string(rs[1]) + ",count:" + string(rs[3]) + ",pulses:[" + string(rs[4]) + "]}"
				doLog("Found Pulse [%s]\n", pulses)
				buffer = buffer[m[1]:]

				// Hope that help.
				n, err = s.Write([]byte("10;PING;\n"))
				if err != nil {
					log.Fatal(err)
				}

				// Send signal
				sig <- pulses

			}

		}
	} // Endless loop for
}

/* MQTT *********************************************************/

func createClientOptions(clientID string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientID)
	return opts
}

func connect(clientID string, url *url.URL) mqtt.Client {
	opts := createClientOptions(clientID, url)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})
}

// push_mqtt push messages to queues
func pushMqtt(client mqtt.Client, ps *list.List) {
	doLog("Buffer contains %d elements\n", ps.Len())
	for ps.Len() > 0 {
		// Get First In, and remove it
		e := ps.Front()

		if token := client.Publish("signal_raw", 0, false, e.Value); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
		doLog("Publish '%s' \n", e.Value)

		ps.Remove(e)
	}

}

/* Misc functions **********************************************************/

func parseArgs() {
	//mqtt://<user>:<pass>@<server>.cloudmqtt.com:<port>
	var tMqtt string
	flag.StringVar(&tMqtt, "mqtt_url", "mqtt://localhost:1883", "MQTT Broker Address")
	//	flag.StringVar(&influx_db, "mqtt_topic_raw", "signal_raw", "InfluxDB database")
	//	flag.StringVar(&influx_user, "mqtt_topic_decoded", "", "InfluxDB user")
	//	flag.StringVar(&influx_pass, "influx_pass", "", "InfluxDB password")

	// cnfSignalRawTopic, cnfSignalDecodedTopic
	flag.StringVar(&cnfTopicSignalRaw, "topic-signal-raw", "signal_raw", "MQTT Topic for publishing raw signals")
	flag.StringVar(&cnfTopicSignalDecoded, "topic-signal-decoded", "signal_decoded", "MQTT Topic for publishing decoded signals")

	flag.StringVar(&cnfSerialDev, "serial-dev", "/dev/ttyUSB0", "Serial port device")
	flag.IntVar(&cnfSerialBaudRate, "serial-baudrate", 57600, "Serial port baudrate")
	flag.StringVar(&cnfLogFile, "log", "", "Path to log file")
	flag.BoolVar(&cnfIsDaemon, "daemon", false, "Run in background")
	flag.Parse()

	var err error
	cnfMqttURL, err = url.Parse(tMqtt)
	if err != nil {
		log.Fatal(err)
	}

	doLog("Conf : MQTT broker URL : %s\n", cnfMqttURL)
	doLog("Conf : MQTT signalRaw topic name : %s\n", cnfTopicSignalRaw)
	doLog("Conf : MQTT signalDecoded topic name : %s\n", cnfTopicSignalDecoded)
	doLog("Conf : Serial port: %s (baudrate:%d)\n", cnfSerialDev, cnfSerialBaudRate)

}

func main() {

	logfile = os.Stdout
	parseArgs()
	doLog("Starting RFSkipper decoder \n")

	// Inter routines communicatin
	signals := make(chan string)

	// Serial Port processing routine
	go serialworker(signals)

	ps := list.New()

	// Connect MQTT
	client := connect("rfs-decoder", cnfMqttURL)

	/*
		go func(sig chan string) {
			for {
				test := "coucou"
				sig <- test
				fmt.Println("coucou")
				time.Sleep(time.Second)
			}
		}(signals)
	*/

	// Loop until someting happens
	for {
		p := <-signals
		ps.PushBack(p)
		pushMqtt(client, ps)

	}

}
