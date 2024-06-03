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
				logrus.Info(fmt.Sprintf("[event] Logged keypress [%s]", e.KeyString()))
				conn.Write([]byte(e.KeyString()))
				logrus.Info("Sent key")
			}
		}
	}
}

func sartRecieving(logger keylogger.KeyLogger, host string, p string) {
	listener, err := net.Listen("tcp", host+":"+p)
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
			logrus.Info(fmt.Sprintf("Recieved key [%s]", buff))
			logger.WriteOnce(string(buff))
		}
	}
}

func main() {
	logrus.Info("initializing program")
	if len(os.Args) < 4 {
		println("Usage: tkeyp [host] [port] [sender/reciever]")
		return
	}
	host := os.Args[1]
	port := os.Args[2]
	role := os.Args[3]

	logger := selectKeyboard(false)

	switch role {
	case "sender":
		startSending(logger, host, port)
	case "reciever":
		sartRecieving(logger, host, port)
	default:
		logrus.Fatal("role has to be either [sender] or [reciever]")
	}

}
