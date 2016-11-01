package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	version = "1.0.0.beta2"

	//subcommands
	add, del, skip *flag.FlagSet

	//flags
	id string
	//add
	OdooURL, dbName, pass, backupDir string
	odooVerion                       float64
	//del
	delID int
)

func init() {
	//subcommands
	flag.NewFlagSet("version", flag.ExitOnError)
	flag.NewFlagSet("help", flag.ExitOnError)
	flag.NewFlagSet("show", flag.ExitOnError)
	add = flag.NewFlagSet("add", flag.ExitOnError)
	del = flag.NewFlagSet("del", flag.ExitOnError)

	//flags
	flag.StringVar(&id, "n", "", "set id number to backup by id, example 'odoobup -n=1' or 'odoobup -n=1,2,3'.")

	//add
	add.StringVar(&OdooURL, "url", "", "set odoo url.")
	add.StringVar(&dbName, "db_name", "", "set odoo database name.")
	add.StringVar(&pass, "password", "", "set odoo master password.")
	add.StringVar(&backupDir, "backup_dir", "", "set backup path directory.")
	add.Float64Var(&odooVerion, "version", 0.0, "set odoo version.")
	//del
	del.IntVar(&delID, "n", 0, "set id number to delete it.")
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 && len(allConfig()) == 0 {
		fmt.Fprintln(os.Stderr, "no configuration setting found. See `odoobup help`\n"+
			"or try `odoobup add` to add new configuration setting")
		os.Exit(1)
	} else {
		if id == "" && len(flag.Arg(0)) == 0 {
			backup()
			os.Exit(0)
		} else {
			if id != "" {
				idSliceStr := strings.Split(id, ",")
				idSlice, err := sliceAtoi(idSliceStr)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}

				backup(idSlice...)
				os.Exit(0)
			}
		}
	}

	if flag.Arg(0) == "version" { // version subcommand
		fmt.Printf("odoobup version %s \n", version)
		os.Exit(0)
	} else if flag.Arg(0) == "help" { // help subcommand
		helpCommand()
	} else if flag.Arg(0) == "add" { // add subcommand
		addCommand()
	} else if flag.Arg(0) == "show" { // show subcommand
		showCommand()
	} else if flag.Arg(0) == "del" { // show subcommand
		delCommand()
	} else {
		fmt.Fprintf(os.Stderr, "unknown subcommand `%s`, See `odoobup help`\n", flag.Arg(0))
		os.Exit(1)
	}
}

func helpCommand() {
	fmt.Printf("The commands are:\n\n")
	fmt.Printf("\t%s\t%s\n", "add", "add new configuration setting")
	fmt.Printf("\t%s\t%s\n", "show", "show all configurations")
	fmt.Printf("\t%s\t%s\n", "del", "delete configuration setting by id number")
	fmt.Printf("\t%s\t%s\n", "version", "show program version number")
	fmt.Println()
	fmt.Println("Use \"odoobup [subcommand] -h\" for more information about a command.")
	fmt.Println()
}

func addCommand() {
	add.Parse(flag.Args()[1:])

	if OdooURL == "" || dbName == "" || pass == "" || backupDir == "" || odooVerion == 0.0 {
		fmt.Fprintln(os.Stderr, "all flags required.")
		fmt.Fprintln(os.Stderr, "example: `odoobup add -url='http://localhost:8069' -db_name='odoo'"+
			" -password='odoo mastre password' -backup_dir='/home/odoo/odoo_backup' -version=8.0`")
		os.Exit(1)
	}

	if odooVerion < 8.0 {
		fmt.Fprintln(os.Stderr, "odoobup not support version", odooVerion)
		os.Exit(1)
	}

	ci := &ConfigInfo{
		URL:       OdooURL,
		DBName:    dbName,
		OdooPass:  pass,
		BackupDir: backupDir,
		Version:   odooVerion,
	}

	_, err := NewConfig(ci)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func showCommand() {
	for _, value := range allConfig() {
		fmt.Printf("{%d: {url: %s, database name: %s, backup path directory: %s, odoo version: %.1f}}\n",
			value.ID, value.Info.URL, value.Info.DBName, value.Info.BackupDir, value.Info.Version)
	}
}

func delCommand() {
	del.Parse(flag.Args()[1:])

	if delID == 0 {
		fmt.Fprintln(os.Stderr, "Please add id number to delete it using -n flag.\nexample: `odoobup delete -n=1`")
		os.Exit(1)
	}

	err := DeleteConfig(delID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func backup(i ...int) {
	urlPath := "/web/database/backup"

	if len(i) == 0 {
		for _, value := range allConfig() {
			if value.Info.Version == 8.0 {
				backupGET(urlPath, value.Info)
			} else {
				backupPOST(urlPath, value.Info)
			}
		}
	} else {
		for _, v := range i {
			c, err := ConfigByID(v)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			if c.Info.Version == 8.0 {
				backupGET(urlPath, c.Info)
			} else {
				backupPOST(urlPath, c.Info)
			}
		}
	}
}

func backupGET(urlPath string, ci ConfigInfo) {
	token := strconv.Itoa(int(time.Now().UTC().UnixNano()) / int(time.Millisecond))

	address := fmt.Sprintf("%s%s?token=%s&backup_db=%s&backup_format=%s&backup_pwd=%s",
		ci.URL, urlPath, token, ci.DBName, "zip", ci.OdooPass)

	resp, err := http.Get(address)

	response(resp, ci, err)
}

func backupPOST(urlPath string, ci ConfigInfo) {
	address := fmt.Sprintf("%s%s", ci.URL, urlPath)

	resp, err := http.PostForm(address,
		url.Values{"master_pwd": {ci.OdooPass}, "name": {ci.DBName}, "backup_format": {"zip"}})

	response(resp, ci, err)
}

// Processing response
func response(resp *http.Response, ci ConfigInfo, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s, database:%s... connection refused\n", ci.URL, ci.DBName)
		return
	}

	defer resp.Body.Close()

	fileName := resp.Header.Get("Content-Disposition")
	fileName = strings.Replace(fileName, "attachment; filename*=UTF-8''", "", -1)

	out, err := os.Create(ci.BackupDir + "/" + fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s, database:%s... Database does not exist or Access denied.\n", ci.URL, ci.DBName)
		return
	}
	defer out.Close()

	prefix := fmt.Sprintf("Get: %s, database:%s... ", ci.URL, ci.DBName)
	bar := pb.New64(resp.ContentLength).SetUnits(pb.U_BYTES).Prefix(prefix)
	bar.ShowBar = false
	bar.ShowPercent = false
	bar.ShowFinalTime = false
	bar.ShowTimeLeft = false
	bar.Start()

	reader := bar.NewProxyReader(resp.Body)

	_, err = io.Copy(out, reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	bar.Postfix("Done")
	bar.Finish()
}

func allConfig() []Config {
	allConfig, err := AllConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return allConfig
}

//convert string slice to integer slice
func sliceAtoi(sa []string) ([]int, error) {
	si := make([]int, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.Atoi(a)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}
