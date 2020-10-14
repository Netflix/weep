package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"github.com/netflix/weep/challenge"
	"github.com/netflix/weep/mtls"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/netflix/weep/config"
	"github.com/netflix/weep/consoleme"
	"github.com/netflix/weep/handlers"
	"github.com/netflix/weep/metadata"
	"github.com/netflix/weep/util"
	"github.com/netflix/weep/version"
)

var (
	logFormat  string
	logLevel   string
	port       int
	listenAddr string
	configPath string

	mtlsClient *http.Client
)

func main() {
	versionPtr := flag.Bool("version", false, "Prints version")
	metadataSvcPtr := flag.Bool("meta-data", false, "Starts the Meta-data Service")
	exportPtr := flag.Bool("export", false, "Triggers printing out credentials to stdout")
	listPtr := flag.Bool("list", false, "List Eligible Roles")
	setupPtr := flag.Bool("setup", false, "Print out the commands you should run to get routing setup")
	flag.StringVar(&metadata.Role, "role", "", "Role ARN")
	flag.StringVar(&logFormat, "log_fmt", "tty", "Log Format - json or tty")
	flag.StringVar(&logLevel, "log_level", "info", "Log Level - info, debug, warn")
	flag.StringVar(&listenAddr, "listen_ip", "127.0.0.1", "IP Address to listen on")
	flag.StringVar(&configPath, "config", "", "Config file (yml)")
	flag.BoolVar(&metadata.NoIpRestrict, "no_ip", false, "removes VPN IP restrictions (PAGES SECOPS)")
	flag.StringVar(&metadata.MetadataRegion, "region", "us-east-1", "Region for Metadata service")
	flag.IntVar(&port, "port", 9090, "Listening Port")
	flag.Parse()

	if *versionPtr {
		fmt.Println(version.GetVersion())
		os.Exit(0)
	}

	if *setupPtr {
		// use os-specific routine
		PrintSetup()
		os.Exit(0)
	}

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	// Set the log format.  Default to Text
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}

	// Set the log level and default to INFO
	switch logLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.SetConfigType("yaml")
	viper.SetConfigName(".weep")
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AddConfigPath(home + "/.config/weep/")
	err = viper.ReadInConfig()
	if err == nil {
		log.Debug("Found config")
		err = viper.Unmarshal(&config.Config)
		if err != nil {
			log.Fatalf("unable to decode into struct, %v", err)
		}
	}
	consoleMeUrl := config.Config.ConsoleMeUrl

	authenticationMethod := config.Config.AuthenticationMethod

	var client *consoleme.Client

	if authenticationMethod == "mtls" {
		mtlsClient, err := mtls.NewHTTPClient()
		util.CheckError(err)
		client, err = consoleme.NewClientWithMtls(consoleMeUrl, mtlsClient)
		util.CheckError(err)
	} else if authenticationMethod == "challenge" {
		err = challenge.RefreshChallenge()
		util.CheckError(err)
		httpClient, err := challenge.NewHTTPClient(consoleMeUrl)
		util.CheckError(err)
		client, err = consoleme.NewClientWithJwtAuth(consoleMeUrl, httpClient)
		util.CheckError(err)
	} else {
		log.Fatal("Authentication method unsupported or not provided.")
	}


	if *listPtr {
		roles, err := client.Roles()
		util.CheckError(err)

		fmt.Println("Roles:")
		for i := range roles {
			fmt.Println("  ", roles[i])
		}
		os.Exit(0)
	}

	if !*versionPtr && !*metadataSvcPtr && !*exportPtr {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *metadataSvcPtr && metadata.NoIpRestrict {
		log.Fatal("You cannot have non IP-restricted credentials in the metadata service due to potential Duo lockout")
	}

	if len(metadata.Role) < 1 {
		log.Error("Please provide a Role via the --role command line flag")
		os.Exit(1)
	}

	if *exportPtr {
		creds, err := client.GetRoleCredentials(metadata.Role, metadata.NoIpRestrict)
		util.CheckError(err)
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s && export AWS_SECRET_ACCESS_KEY=%s && export AWS_SESSION_TOKEN=%s\n",
			creds.AccessKeyId, creds.SecretAccessKey, creds.SessionToken)
		os.Exit(0)
	}

	ipaddress := net.ParseIP(listenAddr)

	if ipaddress == nil {
		fmt.Println("Invalid IP: ", listenAddr)
		os.Exit(1)
	}

	listener_addr := fmt.Sprintf("%s:%d", ipaddress, port)

	if *metadataSvcPtr {
		router := mux.NewRouter()
		router.HandleFunc("/{version}/", handlers.MetaDataServiceMiddleware(handlers.BaseVersionHandler))
		router.HandleFunc("/{version}/api/token", handlers.MetaDataServiceMiddleware(handlers.TokenHandler)).Methods("PUT")
		router.HandleFunc("/{version}/meta-data", handlers.MetaDataServiceMiddleware(handlers.BaseHandler))
		router.HandleFunc("/{version}/meta-data/", handlers.MetaDataServiceMiddleware(handlers.BaseHandler))
		router.HandleFunc("/{version}/meta-data/iam/info", handlers.MetaDataServiceMiddleware(handlers.IamInfoHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/", handlers.MetaDataServiceMiddleware(handlers.RoleHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/{role}", handlers.MetaDataServiceMiddleware(handlers.CredentialsHandler))
		router.HandleFunc("/{version}/dynamic/instance-identity/document", handlers.MetaDataServiceMiddleware(handlers.InstanceIdentityDocumentHandler))
		router.HandleFunc("/{path:.*}", handlers.MetaDataServiceMiddleware(handlers.CustomHandler))

		go metadata.StartMetaDataRefresh(client)

		go func() {
			log.Info("Starting weep meta-data service...")
			log.Info("Server started on: ", listener_addr)
			log.Info(http.ListenAndServe(listener_addr, router))
		}()
	}

	// Check for interrupt signal and exit cleanly
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown signal received, exiting weep meta-data service...")
}
