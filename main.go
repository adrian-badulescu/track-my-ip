package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	gomail "gopkg.in/mail.v2"
)

type ip struct {
	Ip string
}

func getIp(c chan ip) {
	defer close(c)
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		fmt.Println("Err in getting IP: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var IP ip

	json.Unmarshal(body, &IP)

	c <- IP
}

func sendMail(channel chan cmdLn) {
	t := time.Now()
	c := <-channel

	m := gomail.NewMessage()

	// Set E-Mail sender

	m.SetHeader("From", *c.from)

	// Set E-Mail receivers

	m.SetHeader("To", *c.to)

	// Set E-Mail subject
	m.SetHeader("Subject", "Your last known IP: "+c.ip)
	
	fmt.Println(c.ip)
	

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "System Local Time: "+t.String())

	// Settings for SMTP server
	d := gomail.NewDialer(*c.server, *c.port, *c.user, *c.password)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	fmt.Println("ISV ", *c.isv)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: *c.isv}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type cmdLn struct {
	server   *string
	port     *int
	user     *string
	password *string
	from     *string
	to       *string
	isv      *bool
	ip       string
}

func commandLine(cn chan cmdLn, IP chan ip) {
	// closing the channel on the emitter end
	defer close(cn)

	// reading CLI the params 
	server := flag.String("server", "", "smtp.example.com")
	port := flag.Int("port", 0, "587")
	user := flag.String("user", "", "Jakie_Chan")
	password := flag.String("password", "", "Ksjn4##$jLK!")
	from := flag.String("from", "", "Let the receiver know from whom is the email")
	to := flag.String("to", "", "Drop the receiver email address here")
	isv := flag.Bool("isv", true, "Only needed when SSL/TLS certificate is not valid on server")
	flag.Parse()

	// getting the ip from the channel IP and introducing it into the cmdLn struct
	iP := <-IP

	var c cmdLn = cmdLn{server, port, user, password, from, to, isv, iP.Ip}

	cn <- c
}
func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	ipChannel := make(chan ip)
	commandChannel := make(chan cmdLn)

	go func() {
		getIp(ipChannel)
		wg.Done()
	}()

	go func() {
		commandLine(commandChannel, ipChannel)
		wg.Done()
	}()

	go func() {
		sendMail(commandChannel)
		wg.Done()
	}()

	wg.Wait()
}

// -server=smtp.mailgun.org -port=587 -user=postmaster@sandbox8427fbc34c6148a2b09edf38e069e889.mailgun.org -password=9b44b099a236c7d6f8630ec1fc6a09f0-3e51f8d2-ee4d46d5 -from=me@techdocs.co.uk -to=adrianb2104@gmail.com
