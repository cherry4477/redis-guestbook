/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis"
	"github.com/garyburd/redigo/redis"
)

var (
	masterPool *simpleredis.ConnectionPool
	slavePool  *simpleredis.ConnectionPool
)

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	list := simpleredis.NewList(slavePool, key)
	members := HandleError(list.GetAll()).([]string)
	membersJSON := HandleError(json.MarshalIndent(members, "", "  ")).([]byte)
	rw.Write(membersJSON)
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(masterPool, key)
	HandleError(nil, list.Add(value))
	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info := HandleError(masterPool.Get(0).Do("INFO")).([]byte)
	rw.Write(info)
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	envJSON := HandleError(json.MarshalIndent(environment, "", "  ")).([]byte)
	rw.Write(envJSON)
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		panic(err)
	}
	return result
}

func main() {
	sentinel := os.Getenv(os.Getenv("EnvName_SentinelAddr"))
	cluster := os.Getenv(os.Getenv("EnvName_ClusterName"))
	password := os.Getenv(os.Getenv("EnvName_Password"))
	password = strings.TrimSpace(password)
	
	//masterPool = simpleredis.NewConnectionPoolHost("redis-master:6379")
	master := strings.Join(getRedisMasterAddr(sentinel, cluster), ":")
	if password != "" {
		master = password + "@" + master
	}
	masterPool = simpleredis.NewConnectionPoolHost(master)
	defer masterPool.Close()

	//slavePool = simpleredis.NewConnectionPoolHost("redis-slave:6379")
	slave := strings.Join(getRedisSlaveAddr(sentinel, cluster), ":")
	if password != "" {
		slave = password + "@" + slave
	}
	slavePool = simpleredis.NewConnectionPoolHost(slave)
	defer slavePool.Close()

	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
}

//=================================



func getRedisMasterAddr(sentinelAddr, clusterName string) []string {
	if len(sentinelAddr) == 0 {
		//log.Printf("Redis sentinelAddr is nil.")
		//return "", ""
		return []string{"", ""}
	}

	conn, err := redis.DialTimeout("tcp", sentinelAddr, time.Second*10, time.Second*10, time.Second*10)
	if err != nil {
		//log.Printf("redis dial timeout(\"tcp\", \"%s\", %d) error(%v)", sentinelAddr, time.Second, err)
		//return "", ""
		return []string{"", ""}
	}
	defer conn.Close()

	redisMasterPair, err := redis.Strings(conn.Do("SENTINEL", "get-master-addr-by-name", clusterName))
	if err != nil {
		//log.Printf("conn.Do(\"SENTINEL\", \"get-master-addr-by-name\", \"%s\") error(%v)", clusterName, err)
		//return "", ""
		return []string{"", ""}
	}

	if len(redisMasterPair) != 2 {
		//return "", ""
		return []string{"", ""}
	}
	//return redisMasterPair[0], redisMasterPair[1]
	return redisMasterPair[:2]
}

func getRedisSlaveAddr(sentinelAddr, clusterName string) []string {
	if len(sentinelAddr) == 0 {
		//log.Printf("Redis sentinelAddr is nil.")
		//return "", ""
		return []string{"", ""}
	}

	conn, err := redis.DialTimeout("tcp", sentinelAddr, time.Second*10, time.Second*10, time.Second*10)
	if err != nil {
		//log.Printf("redis dial timeout(\"tcp\", \"%s\", %d) error(%v)", sentinelAddr, time.Second, err)
		//return "", ""
		return []string{"", ""}
	}
	defer conn.Close()

	redisSlavePair, err := redis.Strings(conn.Do("SENTINEL", "slaves", clusterName))
	if err != nil {
		//log.Printf("conn.Do(\"SENTINEL\", \"get-master-addr-by-name\", \"%s\") error(%v)", clusterName, err)
		//return "", ""
		return []string{"", ""}
	}

	if len(redisSlavePair) < 2 {
		//return "", ""
		return []string{"", ""}
	}
	//return redisSlavePair[0], redisSlavePair[1]
	return redisSlavePair[:2]
}