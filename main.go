package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/MarinX/keylogger"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

func selectKeyboard(interactive bool) keylogger.KeyLogger {
	// Find all keyboards and select the first one
	keyboards := keylogger.FindAllKeyboardDevices()
	active_keyboard := ""
	logrus.Info("keyboards found: ", len(keyboards))
	for i := 0; i < len(keyboards); i++ {
		path := "/sys/class/input/%s/device/name"
		device := strings.Split(keyboards[i], "/")[3]
		input := fmt.Sprintf(path, device)
		buff, err := os.ReadFile(input)
		if err != nil {
			logrus.Errorln("Found error on read file")
			logrus.Fatal(err)
		}
		title := fmt.Sprintf("Keyboard %d\n", i)
		out := fmt.Sprintf("fs dev: %s\n fs input: %s\n Device name: %s", keyboards[i], input, string(buff))
		style := lipgloss.NewStyle().
			SetString(title).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTop(true).BorderBottom(true).BorderLeft(true).BorderRight(true).Padding(0, 1, 0, 1)
		fmt.Println(style.Render(out))
	}
	if !interactive {
		active_keyboard = keylogger.FindKeyboardDevice()
		logrus.Infoln("Auto selected: ", active_keyboard)
	} else {
		// TODO: Implement interactive keyboard selection
		logrus.Fatal("Interactive keyboard selection not implemented yet")
	}

	logger, err := keylogger.New(active_keyboard)
	if err != nil {
		logrus.Error("Error creating the logger")
		logrus.Fatal(err)
	}
	return *logger
}

func startSending(logger keylogger.KeyLogger, host string, p string) {
	tcpServer, err := net.ResolveTCPAddr("tcp", host+":"+p)
	if err != nil {
		logrus.Error("Unable to resolve TPC address")
		logrus.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		logrus.Errorln("Found error dialing TCP address")
		logrus.Fatal(err)
	}

	events := logger.Read()
	logrus.Info("Now sending keboard events")
	for e := range events {
		switch e.Type {
		case keylogger.EvKey:
			if e.KeyPress() {
				// Careful with the logs, they could contain sensitive information
				// logrus.Info(fmt.Sprintf("[event] Logged keypress [%s]", e.KeyString()))
				conn.Write([]byte(e.KeyString()))
				logrus.Info("Sent key")
			}
		}
	}
}

func sartRecieving(logger keylogger.KeyLogger) {
	listener, err := net.Listen("tcp", "localhost:6094")
	if err != nil {
		logrus.Fatal(err)
	}
	defer listener.Close()
	for {
		logrus.Info("Waiting on message")
		mssg, err := listener.Accept()
		defer mssg.Close()
		logrus.Info("recieved message")
		if err != nil {
			logrus.Error("Unable to accept request")
			logrus.Error(err)
		}
		buff := make([]byte, 1)
		for {
			n, err := mssg.Read(buff[:])
			if err != nil || n == 0 {
				break
			}
			// Careful with the logs, they could contain sensitive information
			// logrus.Info(fmt.Sprintf("Recieved key [%s]", buff))
			logrus.Info(fmt.Sprintf("Recieved key"))
			logger.WriteOnce(string(buff))
		}
	}
}

func main() {
	logrus.Info("initializing program")
	logger := selectKeyboard(false)

	if os.Args[1] != "sender" && os.Args[1] != "reciever" {
		logrus.Fatal("Role has to be either [sender] or [reciever]")
	}

	role := os.Args[1]
	if role == "sender" {
		if len(os.Args) != 4 {
			logrus.Fatal("The sender role needs a host and port to send the data to")
		}
		ip := net.ParseIP(os.Args[2])
		if ip == nil {
			logrus.Fatal("Invalid IP address")
		}
		host := os.Args[2]
		port := os.Args[3]
		startSending(logger, host, port)
	}

	if role == "reciever" {
		if len(os.Args) > 2 {
			logrus.Fatal("The reciever role does not need any arguments")
		}
		sartRecieving(logger)
	}

	logrus.Fatal("Usage: sudo ./keylogger [role] [host] [port]")
}
