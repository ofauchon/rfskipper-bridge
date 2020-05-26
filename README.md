rfskipper-decoder 

# Introduction

RFSkipper-decoder main functions :  

- Connect the RFSkipper Gateway through the serial / usb serial interface
- Listen of incoming RF pulses  
- Publish the pulses in a 'signal_raw' topic
- Try to decode the pulses with embedded decoders and eventually publish decoded messages 
on the 'signal_decoded' MQTT topic


# Build and run 

  * Run local MQTT Broker (Optionnal)

If you don't already have a MQTT Broker, you can install one local
instance of Mosquitto MQTT broker. 

$ pacman -S mosquitto 
$ systemctl start mosquitto
$ systemctl enable mosquitto

  * Run RFSkipper-decoder 

Have a look to the options first: 

```
go run rfskipper-decoder.go --help
Usage of /tmp/go-build041482982/b001/exe/rfskipper-decoder:
  -daemon
    	Run in background
  -log string
    	Path to log file
  -mqtt_url string
    	MQTT Broker Address (default "mqtt://localhost:1883")
  -serial-baudrate int
    	Serial port baudrate (default 57600)
  -serial-dev string
    	Serial port device (default "/dev/ttyUSB0")
  -topic-signal-decoded string
    	MQTT Topic for publishing decoded signals (default "signal_decoded")
  -topic-signal-raw string
    	MQTT Topic for publishing raw signals (default "signal_raw")

 Run it : 

```
$ go run rfskipper-decoder.go --serial-dev /dev/ttyACMX --serial-baudrate 57600
```

  * Connect with a mosquitto subscriber

```
$ mosquitto_sub  -L mqtt://localhost:1883/signal_raw 
{id:20,count:131,pulses:[10,40,30,30,20,230,20,40,30,30,20,700,30,40,10,40,30,700,20,40,30,30,30,230,20,470,20,40,20,40,10,240,20,40,30,720,10,40,20,40,20,40,30,30,30,570,20,30,30,40,20,40,20,230,500,40,20,290,30,40,20,720,30,40,20,40,20,630,10,490,0,800,30,30,30,40,30,490,10,40,30,30,30,230,20,40,20,40,10,480,10,430,0,50,500,30,30,220,30,30,30,40,20,40,10,630,30,40,20,40,20,40,30,30,30,40,30,500,20,40,30,30,30,290,20,490,10,40,20,230,30]}
```







Then, it sends initialisation sequence: 

``` 
net chan 11
net pan 61453
net addr 00:00:00:00:00:00:00:01^M
```

before starting packet dump mode : 

```
net pktdump
```

# Run

First, you need to install required libraries: 

```
go get github.com/tarm/serial
go get github.com/influxdata/influxdb/client/v2
```

Then edit and run start.sh wrapper: 

```
./start.sh
```

you'll find the logs in logs/


