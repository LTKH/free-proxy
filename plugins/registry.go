package plugins

import (
    //"log"
    //"time"
    "github.com/ltkh/free-proxy/internal/dbase"
    //"github.com/ltkh/free-proxy/plugins/geonode"
)

//import "github.com/influxdata/telegraf"
//type Creator func() telegraf.Input

//var Inputs = map[string]Creator{}

//func Add(name string, creator Creator) {
//	Inputs[name] = creator
//}

func New(db *dbase.DB) error {
    /*
    go func(){
        for{
            items, err := geonode.Gather()
            if err != nil {
                log.Printf("[error] %v", err)
            } else {
                err := db.SaveProxies(items)
                if err != nil {
                    log.Printf("[error] %v", err)
                }
            }
            time.Sleep(600 * time.Second)
        }
    }()
    */

    return nil
}