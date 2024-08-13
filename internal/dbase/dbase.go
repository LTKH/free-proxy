package dbase

import (
    //"os"
    "fmt"
    "time"
    //"sync"
    //"net"
    "io"
    //"time"
    //"errors"
    //"regexp"
    //"strings"
    "crypto/sha1"
    "encoding/hex"
    //"encoding/json"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    //"github.com/ltkh/netmap/internal/config"
)

type DB struct {
    Client     *sql.DB
}

type Proxy struct {
    Id           string                 `json:"id"`
    IP           string                 `json:"ip"`
    Port         string                 `json:"port"`
    Protocol     string                 `json:"protocol"`
    Protocols    []string               `json:"protocols"`
    Latency      float64                `json:"latency"`
    ErrorCount   int64                  `json:"errorCount"`                
}

func GetHash(text string) string {
    h := sha1.New()
    io.WriteString(h, text)
    return hex.EncodeToString(h.Sum(nil))
}

func New(conn string) (*DB, error) {
    client, err := sql.Open("sqlite3", conn)
    if err != nil {
        return &DB{}, err
    }

    return &DB{Client: client}, nil
}

func (db *DB) CreateTables() error {
    _, err := db.Client.Exec(
      `create table if not exists proxyList (
        id             varchar(50) primary key,
        ip             varchar(20) not null,
        port           bigint(20) not null,
        protocol       varchar(10) not null,
        city           varchar(10) null,
        country        varchar(10) null,
        lastChecked    bigint(20) default 0,
        latency        float default 0,
        errorCount     int default 0,
        workingPercent float default 0
      );
      create index if not exists localId ON records (id);
    `)
    if err != nil {
        return err
    }

    return nil
}

func (db *DB) UpdateProxy(percent float64, proxy Proxy) error {

    errCount := ""
    if percent == 0 {
        errCount = ", errorCount = errorCount+1"
    }

    sql := fmt.Sprintf("update proxyList set lastChecked = ?, latency = ?, workingPercent = ? %v where id = ?", errCount)
    //fmt.Printf("test - %v\n", sql)
          
    _, err := db.Client.Exec(
        sql, 
        time.Now().UTC().Unix(),
        proxy.Latency,
        percent,
        proxy.Id,
    )

    if err != nil {
        return err
    }

    return nil
}

func (db *DB) SaveProxies(proxies []Proxy) error {

    sql := "insert or ignore into proxyList (id,ip,port,protocol) values (?,?,?,?)"

    for _, proxy := range proxies {

        proxy.Id = GetHash(fmt.Sprintf("%s://%s:%s@%s:%s", proxy.Protocol, "", "", proxy.IP, proxy.Port))

        if len(proxy.Protocols) == 0 {
            proxy.Protocols = append(proxy.Protocols, proxy.Protocol)
        }

        for _, proto := range proxy.Protocols {
            
            _, err := db.Client.Exec(
                sql, 
                proxy.Id, 
                proxy.IP, 
                proxy.Port, 
                proto,
                //time.Now().UTC().Unix(),
                //rec.LocalAddr.Name, 
                //rec.LocalAddr.IP, 
                //rec.RemoteAddr.Name, 
                //rec.RemoteAddr.IP,
                //relation, 
                //options, 
            )

            if err != nil {
                return err
            }

        }
        
    }

    return nil
}

func (db *DB) LoadProxies() ([]Proxy, error) {

    result := []Proxy{}

    sql := "select id,ip,port,protocol from proxyList where errorCount < 10 order by workingPercent desc, latency asc, errorCount asc"

    rows, err := db.Client.Query(sql, nil)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var pr Proxy
        
        err := rows.Scan(
            &pr.Id, 
            &pr.IP,
            &pr.Port, 
            &pr.Protocol, 
        )

        if err != nil { 
            return nil, err 
        }

        result = append(result, pr) 
    }

    return result, nil
}
