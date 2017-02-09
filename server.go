package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/db"
	"github.com/intervention-engine/ie/huddles"
	"github.com/intervention-engine/ie/models"
	"github.com/intervention-engine/ie/utilities"
	"github.com/robfig/cron"
)

// Setup the huddle flag to accumulate huddle config file paths
type huddleSlice []string

func (h *huddleSlice) String() string {
	return fmt.Sprint(*h)
}
func (h *huddleSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

var huddleFlag huddleSlice

func init() {
	flag.Var(&huddleFlag, "huddle", "path to a huddle configuration file (repeat flag for multiple paths)")
}

func main() {
	codeLookupFlag := flag.Bool("loadCodes", false, "flag to enable download of icd-9 and icd-10 code lookup")
	icd9URL := flag.String("icd9URL", "https://www.cms.gov/Medicare/Coding/ICD9ProviderDiagnosticCodes/Downloads/ICD-9-CM-v32-master-descriptions.zip", "url for icd-9 code definition zip")
	icd10URL := flag.String("icd10URL", "https://www.cms.gov/Medicare/Coding/ICD10/Downloads/2016-Code-Descriptions-in-Tabular-Order.zip", "url for icd-10 code definition zip")
	subscriptionFlag := flag.Bool("subscriptions", false, "enables limited support for resource subscriptions (default: false)")
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	logFileFlag := flag.String("logdir", "", "Path to a directory for ie and gin logs to be written to.")
	flag.Parse()

	lfpath := getConfigValue(logFileFlag, "IE_LOG_DIR", "")
	if lfpath != "" {
		err := os.Mkdir(lfpath, 0755)
		if err != nil && !os.IsExist(err) {
			log.Println("Error creating log directory:" + err.Error())
		}

		lf, err := os.OpenFile(lfpath+"/ie.log", os.O_RDWR|os.O_APPEND, 0755)
		if os.IsNotExist(err) {
			lf, err = os.Create(lfpath + "/ie.log")
		}
		if err != nil {
			log.Println("Unable to create ie log file:" + err.Error())
		} else {
			log.SetOutput(lf)
		}
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	riskServiceURL := os.Getenv("RISK_SERVICE_URL")
	if riskServiceURL == "" {
		riskServiceURL = "http://localhost:9000"
	}

	huddleConfigs := huddleFlag
	if len(huddleConfigs) == 0 && os.Getenv("HUDDLE_CONFIGS") != "" {
		huddleConfigs = huddleSlice(strings.Split(os.Getenv("HUDDLE_CONFIGS"), ","))
	}

	if *codeLookupFlag {
		utilities.LoadICDFromCMS(mongoURL, *icd9URL, "ICD-9")
		utilities.LoadICDFromCMS(mongoURL, *icd10URL, "ICD-10")
	}

	var ip net.IP
	var selfURL string
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		log.Println("Unable to lookup IP based on hostname, defaulting to localhost.")
		selfURL = "http://localhost:3001"
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip = ipv4
			selfURL = "http://" + ip.String() + ":3001"
		}
	}

	s := server.NewServer(mongoURL)

	api := s.Engine.Group("/api")

	sess, dbase := db.SetupDBConnection("fhir")
	defer sess.Close()

	controllers.RegisterController(dbase, "patients", api, models.Patient{})
	controllers.RegisterController(dbase, "care_teams", api, models.CareTeam{})

	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}

	// Since the huddle controller needs info from the command line, set it up here.  When IE is refactored
	// to take out globals (and other stuff), this should be rethought.
	c := cron.New()
	huddleController := new(huddles.HuddleSchedulerController)
	for _, huddlePath := range huddleConfigs {
		data, err := ioutil.ReadFile(huddlePath)
		if err != nil {
			panic(err)
		}
		config := new(huddles.HuddleConfig)
		if err := json.Unmarshal(data, config); err != nil {
			panic(err)
		}
		huddleController.AddConfig(config)
		scheduleIt := func() {
			_, err := huddles.NewHuddleScheduler(config).ScheduleHuddles()
			if err != nil {
				log.Printf("ERROR: Could not schedule huddles for huddle with name %s: %v", config.Name, err)
			}
		}

		// Wait 1 minute before doing initial runs or setting up cron jobs.  This allows the server to get
		// started (since it needs to initiate the db connection, etc).
		time.AfterFunc(1*time.Minute, func() {
			// Do an initial run
			log.Println("Initial scheduling for huddle with name ", config.Name)
			scheduleIt()

			// Set up the cron job for future runs
			if config.SchedulerCronSpec != "" {
				err := c.AddFunc(config.SchedulerCronSpec, scheduleIt)
				if err != nil {
					panic(err)
				}
				log.Printf("Huddle with name %s scheduled with cron spec: %s\n", config.Name, config.SchedulerCronSpec)
			} else {
				log.Printf("Warning: Huddle with name %s is not configured with a scheduler cron job.\n", config.Name)
			}
		})
	}
	s.Engine.GET("/ScheduleHuddles", huddleController.ScheduleHandler)

	if lfpath != "" {
		glf, err := os.OpenFile(lfpath+"/gin.log", os.O_RDWR|os.O_APPEND, 0755)
		if os.IsNotExist(err) {
			glf, err = os.Create(lfpath + "/gin.log")
		}
		if err != nil {
			log.Println("Unable to create ie log file:" + err.Error())
		} else {
			s.Engine.Use(gin.LoggerWithWriter(glf))
		}
	}

	if len(huddleConfigs) > 0 {
		c.Start()
		defer c.Stop()
	}

	closer := controllers.RegisterRoutes(s, selfURL, riskServiceURL, *subscriptionFlag)
	defer closer()

	s.Run(server.DefaultConfig)
}

func getConfigValue(parsedFlag *string, envVar string, defaultVal string) string {
	val := *parsedFlag
	if val == "" {
		val = os.Getenv(envVar)
		if val == "" {
			val = defaultVal
		}
	}
	return val
}
