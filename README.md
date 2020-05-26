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

```
$ go run rfskipper-decoder.go 
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


