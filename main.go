package main


import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"strings"
)

var SENSORS_ACTIVATION_BATCH int = 20

type Sensor struct {
	Id int `json:"id"`
	Role string `json:"role"`
	Mood string `json:"mood"`
	Baseline string `json:"baseline"`
}

type AllSensors struct {
	Sensors []*Sensor
}

var AllSensorsData AllSensors

func handler(w http.ResponseWriter, r *http.Request) {
	
	teamName, apiDomain := parseHost(r)

	fmt.Fprintf(w,"<body>")

	fmt.Fprintf(w, addBanner("Welcome team %s"), teamName)

	fmt.Fprintf(w, addHeader("DevX Mood Analyzer"))

	//process APIs calls and analytics
	if processSensorActivation(apiDomain) != "success" {
		return
	}
	
	if processSensorsMeasurement(apiDomain) != "success" {
		return
	}
	pureHappy,totalHappy,pureSad,totalSad,pureAngry,totalAngry := moodAnalysis()

	//render results section
	fmt.Fprintf(w,addMoodResults(),	pureHappy,pureSad, pureAngry,
									totalHappy,totalSad,totalAngry)
									
	//render happy/sad
	sniffThreshold, err := strconv.ParseFloat(os.Getenv("SNIFF_THRESHOLD"),64)
	if err != nil { fmt.Fprintf(w,"!!Error in converting sniffing threhold to float64")}
	
	if pureHappy > sniffThreshold {
		fmt.Fprintf(w, addDog("happy"))
	} else {
		fmt.Fprintf(w, addDog("sad"))
	}
		
	//render info section
	fmt.Fprintf(w,addDataTitle("Sniffing threshold"))
	fmt.Fprintf(w,addDataContent("Above %.2f%% of pure happiness"),sniffThreshold)
	fmt.Fprintf(w,addDataTitle("/activate API"))
	fmt.Fprintf(w,addDataContent("%d sensors activated"),len(AllSensorsData.Sensors))
	fmt.Fprintf(w,addDataTitle("/measure API"))
	fmt.Fprintf(w,addDataContent(addAPICallsTable()))

	fmt.Fprintf(w,"</body>")

}

func processSensorActivation (apiDomain string) (status string) {

	tlsConfig := &http.Transport{
	 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	 }


	tlsClient := &http.Client{Transport: tlsConfig}

	var activateApiCall = "http://mood-sensors." + apiDomain + "/activate"

	for i := 0; i < SENSORS_ACTIVATION_BATCH ; i++ {
		//response, err := tlsClient.Get("mood-sensors." + apiDomain + "/activate")
		response, err := tlsClient.Get(activateApiCall)
		if err != nil { 
			status = "Error in calling activate API: " + err.Error()
		} 	 	
		defer response.Body.Close()
	}
	status = "success"
	return
}

func processSensorsMeasurement (apiDomain string) (status string) {
	
	tlsConfig := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}


	tlsClient := &http.Client{Transport: tlsConfig}

	var measureApiCall = "http://mood-sensors." + apiDomain + "/measure"
	response, err := tlsClient.Get(measureApiCall)

	if err != nil { 
		status = "Error in calling measure API: " + err.Error()
	} 	 	

	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body) 	

	if err != nil { 	
		status = "Error in reading measure results: " + err.Error()
	}

	json.Unmarshal(responseData, &AllSensorsData.Sensors)

	status = "success"
	return
}

func moodAnalysis () (	float64, float64, //pure happy, total happy
						float64, float64, //pure sad, total sad
						float64, float64) { //pure angry, total angry
	
	var pureHappy,totalHappy,pureSad,totalSad,pureAngry,totalAngry float64 = 0.0,0.0,0.0,0.0,0.0,0.0
	var totalMeasurements float64 = float64(len(AllSensorsData.Sensors))
	
	for _, sensor := range AllSensorsData.Sensors {
		if sensor.Mood == "happy" {
			if sensor.Baseline == "" {
				pureHappy++
			}
			totalHappy++
		} else if sensor.Mood == "sad" {
			if sensor.Baseline == "" {
				pureSad++
			}
			totalSad++
		} else if sensor.Mood == "angry" {
			if sensor.Baseline == "" {
				pureAngry++
			} 
			totalAngry++
		} else { 
			//error
		}
	}
	
	return	(pureHappy/totalMeasurements)*100,(totalHappy/totalMeasurements)*100,
			(pureSad/totalMeasurements)*100,(totalSad/totalMeasurements)*100,
			(pureAngry/totalMeasurements)*100,(totalAngry/totalMeasurements)*100
	
}

func parseHost (r *http.Request) (teamName string, apiDomain string)  {
    host := r.Host
    hostParts := strings.Split(host, ".")
	teamName = strings.ToUpper(hostParts[1])
	apiDomain = hostParts[1] + "." + hostParts[2] + "." + hostParts[3]

	return
 
}

func addMoodResults () (htmlOutput string) {

	htmlOutput += "<p align='center'>"
	htmlOutput += "<table align='center' border='0'>"
	
	//pure mood row
	htmlOutput += "<tr>"
	htmlOutput += "<td style='font-size:30px;color:DarkGreen'>%.2f%% Happy</td>"
	htmlOutput += "<td>&nbsp;&nbsp;&nbsp;</td>"
	htmlOutput += "<td style='font-size:30px;color:DarkRed'>%.2f%% Sad</td>"
	htmlOutput += "<td>&nbsp;&nbsp;&nbsp;</td>"
	htmlOutput += "<td style='font-size:30px;color:DarkOrange'>%.2f%% Angry</td>"
	htmlOutput += "</tr>"

	//brownfield row	
	htmlOutput += "<tr style='font-size:15px;color:gray'>"
	htmlOutput += "<td>(%.2f%% including brownfield)</td>"
	htmlOutput += "<td>&nbsp;&nbsp;&nbsp;</td>"
	htmlOutput += "<td>(%.2f%% including brownfield)</td>"
	htmlOutput += "<td>&nbsp;&nbsp;&nbsp;</td>"
	htmlOutput += "<td>(%.2f%% including brownfield)</td>"
	htmlOutput += "</tr>"
	
	htmlOutput += "</table></p>"
	return
}

func addBanner (myBanner string) (htmlOutput string) {

    htmlOutput += "<p style='font-size:20px;color:green' align='left'>"
	htmlOutput += myBanner
	htmlOutput += "</p>"
	return
}

func addHeader (myHeader string) (htmlOutput string) {

    htmlOutput += "<p style='font-size:40px;color:navy' align='center'>"
	htmlOutput += myHeader
	htmlOutput += "</p>"
	return
}

func addDog (imgPrefix string) (htmlOutput string) {

	htmlOutput += "<p style='font-size:20px;color:purple' align='center'>"
	htmlOutput += "<img src='https://raw.githubusercontent.com/dektlong/devx-mood/main/" + imgPrefix + "-dog.jpg' alt=''>"
	htmlOutput += "</p>"
	return
}

func addAPICallsTable () (htmlOutput string) {

	htmlOutput += "<p align='left'>"
	htmlOutput += "<table style='font-size:15px;color:gray' border='1'>"
	
	htmlOutput += "<tr>"
	htmlOutput += "<th>Sensor</th>" + "<th>Role</th>" + "<th>Greenfield Mood</th>"+ "<th>Brownfield Mood</th>"
	htmlOutput += "</tr>"

	for _, sensor := range AllSensorsData.Sensors {
  		htmlOutput += "<tr>"
		htmlOutput += "<td>" + strconv.Itoa(sensor.Id) + "</td>"
		htmlOutput += "<td>" + sensor.Role + "&nbsp;</td>"
		htmlOutput += "<td>" + sensor.Mood + "&nbsp;</td>"
		htmlOutput += "<td>" + sensor.Baseline + "</td>"
		htmlOutput += "</tr>"
	}

	htmlOutput += "</table></p>"
	return
}

func addDataTitle (title string) (htmlOutput string) {

	htmlOutput += "<p style='font-size:15px;color:purple' align='left'>"
	htmlOutput += title
	htmlOutput += "</p>"
	return
}

func addDataContent (content string) (htmlOutput string) {

	htmlOutput += "<p style='font-size:15px;color:gray' align='left'>"
	htmlOutput += content
	htmlOutput += "</p><BR>"
	return
}

func main() {
	
	http.HandleFunc("/", handler)

	var addr = flag.String("addr", ":8080", "addr to bind to")
	http.ListenAndServe(*addr, nil)
}
