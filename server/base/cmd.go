package base

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/cherts/anylink/pkg/utils"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xlzd/gotp"
)

var (
	// Submit id
	CommitId string
	// pass plain text
	passwd string
	// generate otp
	otp bool
	// Generate key
	secret bool
	// Show version information
	rev bool
	// Output debug information
	debug bool

	// Used for flags.
	runSrv bool

	linkViper *viper.Viper
	rootCmd   *cobra.Command
)

// Execute executes the root command.
func execute() {
	initCmd()

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// viper.Debug()
	ref := reflect.ValueOf(linkViper)
	s := ref.Elem()
	ee := s.FieldByName("env")
	if ee.Kind() != reflect.Map {
		panic("Viper env is err")
	}
	rr := ee.MapRange()
	for rr.Next() {
		// fmt.Println(rr.Key(), rr.Value().Index(0))
		envs[rr.Key().String()] = rr.Value().Index(0).String()
	}

	if !runSrv {
		os.Exit(0)
	}
}

func initCmd() {
	linkViper = viper.New()
	rootCmd = &cobra.Command{
		Use:   "anylink",
		Short: "AnyLink VPN Server",
		Long:  `AnyLink is a VPN Server application`,
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Println("cmd：", cmd.Use, args)
			runSrv = true

			if rev {
				printVersion()
				os.Exit(0)
			}
		},
	}

	linkViper.SetEnvPrefix("link")

	// Basic configuration
	for _, v := range configs {
		if v.Typ == cfgStr {
			rootCmd.Flags().StringP(v.Name, v.Short, v.ValStr, v.Usage)
		}
		if v.Typ == cfgInt {
			rootCmd.Flags().IntP(v.Name, v.Short, v.ValInt, v.Usage)
		}
		if v.Typ == cfgBool {
			rootCmd.Flags().BoolP(v.Name, v.Short, v.ValBool, v.Usage)
		}

		_ = linkViper.BindPFlag(v.Name, rootCmd.Flags().Lookup(v.Name))
		_ = linkViper.BindEnv(v.Name)
		// viper.SetDefault(v.Name, v.Value)
	}

	rootCmd.Flags().BoolVarP(&rev, "version", "v", false, "display version info")
	rootCmd.AddCommand(initToolCmd())

	cobra.OnInitialize(func() {
		linkViper.AutomaticEnv()
		conf := linkViper.GetString("conf")

		_, err := os.Stat(conf)
		if errors.Is(err, os.ErrNotExist) {
			// No configuration file, no processing
			panic(err)
		}

		linkViper.SetConfigFile(conf)
		err = linkViper.ReadInConfig()
		if err != nil {
			panic("Config file err:" + err.Error())
		}
	})
}

func initToolCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tool",
		Short: "AnyLink tool",
		Long:  `AnyLink tool is a application`,
	}

	toolCmd.Flags().BoolVarP(&rev, "version", "v", false, "display version info")
	toolCmd.Flags().BoolVarP(&secret, "secret", "s", false, "generate a random jwt secret")
	toolCmd.Flags().StringVarP(&passwd, "passwd", "p", "", "convert the password plaintext")
	toolCmd.Flags().BoolVarP(&otp, "otp", "o", false, "generate a random otp secret")
	toolCmd.Flags().BoolVarP(&debug, "debug", "d", false, "list the config viper.Debug() info")

	toolCmd.Run = func(cmd *cobra.Command, args []string) {
		switch {
		case rev:
			printVersion()
		case secret:
			s, _ := utils.RandSecret(40, 60)
			s = strings.Trim(s, "=")
			fmt.Printf("Secret:%s\n", s)
		case otp:
			s := gotp.RandomSecret(32)
			fmt.Printf("Otp:%s\n\n", s)
			qrstr := fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s", "anylink_admin", "admin@anylink", "anylink_admin", s)
			qr, _ := qrcode.New(qrstr, qrcode.High)
			ss := qr.ToSmallString(false)
			io.WriteString(os.Stderr, ss)
		case passwd != "":
			pass, _ := utils.PasswordHash(passwd)
			fmt.Printf("Passwd:%s\n", pass)
		case debug:
			linkViper.Debug()
		default:
			fmt.Println("Using [anylink tool -h] for help")
		}
	}

	return toolCmd
}

func printVersion() {
	fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
		APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
}
