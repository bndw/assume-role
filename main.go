package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <role> [<command> <args...>]\n", os.Args[0])
	flag.PrintDefaults()
}

type Assumable struct {
	Credentials *credentials.Value
	Region      *string
}

func init() {
	flag.Usage = usage
}

func defaultFormat() string {
	switch runtime.GOOS {
	case "windows":
		if os.Getenv("SHELL") == "" {
			return "powershell"
		}
		fallthrough
	default:
		return "bash"
	}
}

func main() {
	var (
		duration = flag.Duration("duration", time.Hour, "The duration that the credentials will be valid for.")
		format   = flag.String("format", defaultFormat(), "Format can be 'bash' or 'powershell'.")
	)
	flag.Parse()
	argv := flag.Args()
	if len(argv) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	stscreds.DefaultDuration = *duration

	role := argv[0]
	args := argv[1:]

	if os.Getenv("ASSUMED_ROLE") != "" {
		cleanEnv()
	}
	assume, err := assumeProfile(role)
	must(err)

	if len(args) == 0 {
		switch *format {
		case "powershell":
			printPowerShellCredentials(role, assume)
		case "bash":
			printCredentials(role, assume)
		default:
			flag.Usage()
			os.Exit(1)
		}
		return
	}

	must(exportVariables(args, assume))
}

func cleanEnv() {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
	os.Unsetenv("AWS_SECURITY_TOKEN")
}

func exportVariables(argv []string, assume *Assumable) error {
	argv0, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}

	os.Setenv("AWS_ACCESS_KEY_ID", assume.Credentials.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", assume.Credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", assume.Credentials.SessionToken)
	os.Setenv("AWS_SECURITY_TOKEN", assume.Credentials.SessionToken)
	os.Setenv("AWS_REGION", *assume.Region)

	env := os.Environ()
	return syscall.Exec(argv0, argv, env)
}

// printCredentials prints the credentials in a way that can easily be sourced
// with bash.
func printCredentials(role string, assume *Assumable) {
	fmt.Printf("export AWS_ACCESS_KEY_ID=\"%s\"\n", assume.Credentials.AccessKeyID)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=\"%s\"\n", assume.Credentials.SecretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=\"%s\"\n", assume.Credentials.SessionToken)
	fmt.Printf("export AWS_SECURITY_TOKEN=\"%s\"\n", assume.Credentials.SessionToken)
	fmt.Printf("export AWS_REGION=\"%s\"\n", *assume.Region)
	fmt.Printf("export ASSUMED_ROLE=\"%s\"\n", role)
	fmt.Printf("# Run this to configure your shell:\n")
	fmt.Printf("# eval $(%s)\n", strings.Join(os.Args, " "))
}

// printPowerShellCredentials prints the credentials in a way that can easily be sourced
// with Windows powershell using Invoke-Expression.
func printPowerShellCredentials(role string, assume *Assumable) {
	fmt.Printf("$env:AWS_ACCESS_KEY_ID=\"%s\"\n", assume.Credentials.AccessKeyID)
	fmt.Printf("$env:AWS_SECRET_ACCESS_KEY=\"%s\"\n", assume.Credentials.SecretAccessKey)
	fmt.Printf("$env:AWS_SESSION_TOKEN=\"%s\"\n", assume.Credentials.SessionToken)
	fmt.Printf("$env:AWS_SECURITY_TOKEN=\"%s\"\n", assume.Credentials.SessionToken)
	fmt.Printf("$env:AWS_REGION=\"%s\"\n", *assume.Region)
	fmt.Printf("$env:ASSUMED_ROLE=\"%s\"\n", role)
	fmt.Printf("# Run this to configure your shell:\n")
	fmt.Printf("# %s | Invoke-Expression \n", strings.Join(os.Args, " "))
}

// assumeProfile assumes the named profile which must exist in ~/.aws/config
// (https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html) and returns the temporary STS
// credentials.
func assumeProfile(profile string) (*Assumable, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:                 profile,
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: readTokenCode,
	}))

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	return &Assumable{
		Credentials: &creds,
		Region:      sess.Config.Region,
	}, nil
}

// readTokenCode reads the MFA token from Stdin.
func readTokenCode() (string, error) {
	r := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "MFA code: ")
	text, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func must(err error) {
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Errors are already on Stderr.
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
