package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
)

var db = dbSetup()

type (
	Config struct {
		ID   int
		Info ConfigInfo
	}

	ConfigInfo struct {
		URL       string
		DBName    string
		OdooPass  string // Odoo Master Password
		BackupDir string
		Version   float64
	}
)

func dbSetup() *bolt.DB {
	dbDir := os.Getenv("HOME") + "/.odoobup"
	dbPath := dbDir + "/config.db"

	//check if .odoobup directory not found
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.Mkdir(dbDir, 0777)
		if err != nil {
			log.Fatalln(err)
		}
	}

	//create and open a database
	db, err := bolt.Open(dbPath, 0777, nil)
	if err != nil {
		log.Fatalln(err)
	}

	//create config bucket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("config"))
		if err != nil {
			log.Fatalln(err)
		}
		return nil
	})

	return db
}

func (c *Config) Encode() (data []byte, err error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(c.Info)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *Config) Decode(data []byte) error {
	r := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&c.Info)
}

func NewConfig(ci *ConfigInfo) (*Config, error) {
	var c Config

	err := db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("config"))

		seq, _ := bkt.NextSequence()
		c.ID = int(seq)
		c.Info = *ci

		encoding, err := c.Encode()
		if err != nil {
			return err
		}

		if err := bkt.Put(itob(c.ID), encoding); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func ConfigByID(id int) (*Config, error) {
	var config Config

	err := db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte("config")).Get(itob(id))
		if v == nil {
			idStr := fmt.Sprint(id)
			return errors.New("The ID " + idStr + " was not Found")
		}

		if err := config.Decode(v); err != nil {
			return err
		}
		config.ID = id

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func AllConfig() ([]Config, error) {
	var config Config
	var allConfig []Config

	err := db.View(func(tx *bolt.Tx) error {
		if err := tx.Bucket([]byte("config")).ForEach(func(k, v []byte) error {
			config.ID, _ = strconv.Atoi(string(k))
			if err := config.Decode(v); err != nil {
				return err
			}

			allConfig = append(allConfig, config)

			return nil

		}); err != nil {
			return err
		}

		return nil

	})
	if err != nil {
		return nil, err
	}

	return allConfig, nil
}

func DeleteConfig(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("config")).Delete(itob(id))
	})
}

func itob(v int) []byte {
	id := fmt.Sprintf("%08d", v)
	return []byte(id)
}
