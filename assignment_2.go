package main

import (
    "github.com/julienschmidt/httprouter"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "fmt"
    "os"
    "log"
    "time"
    "strings"
  )
 

const (
  timeout = time.Duration(time.Second * 10)
)

type (  
    User struct {
        IdReq     bson.ObjectId         `json:"id" bson:"_id"`
        NameReq    string       `json:"name" bson:"name"`
        AddressReq string       `json:"address" bson:"address"`
        CityReq    string       `json:"city" bson:"city"`
        StateReq   string       `json:"state" bson:"state"`
        ZipReq    string       `json:"zip" bson:"zip"` 
    }
)

type (  
    UserResponse struct {
        Id      bson.ObjectId          `json:"id" bson:"_id"`
        Name    string       `json:"name" bson:"name"`
        Address string       `json:"address" bson:"address"`
        City    string       `json:"city" bson:"city"`
        State   string       `json:"state" bson:"state"`
        Zip     string       `json:"zip" bson:"zip"`
        Cc      Coordinate   `json:"coordinate" bson:"coordinate"`
    }
)

type LocationResponse struct {
Results []struct {
Types []string
Formatted_address string
Address_components []struct {
Short_name string
Long_name string
Postcode_localities []string
Types []string
}
Geometry struct {
Location struct {
Lat float64
Lng float64
}
Location_type string
Viewport struct {
Northeast struct {
Lat float64
Lng float64
}
Southwest struct {
Lat float64
Lng float64
}
}
}
}
Status string
}


type Coordinate struct{
        Lat     float64      `json:"lat" bson:"lat"`
        Lng     float64      `json:"lng" bson:"lng"`
}

func post(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    log.Println("inside post")
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }
   var u User
   err = json.Unmarshal(body, &u)
    if (err != nil ) {
        http.Error(rw, "Bad Request, check request payload", http.StatusBadRequest)
        return
    }
  u.IdReq = bson.NewObjectId()
  s := addressString(u)
  var lr = new(LocationResponse)
  lr = lat_lng(s)
  user := &UserResponse{Id:u.IdReq,Name:u.NameReq,Address:u.AddressReq,City:u.CityReq,State:u.StateReq,Zip:u.ZipReq,Cc:Coordinate{Lat : lr.Results[0].Geometry.Location.Lat,Lng : lr.Results[0].Geometry.Location.Lng,},}
  sess, err := mgo.Dial("mongodb://Richa:Indore#1@ds041934.mongolab.com:41934/assignment_2_db")
  if err != nil {
    fmt.Printf("Can't connect to mongo, go error %v\n", err)
    os.Exit(1)
  }
  defer sess.Close()
  sess.SetSafe(&mgo.Safe{})
  collection := sess.DB("assignment_2_db").C("loc")
  err = collection.Insert(user)
  if (err != nil ) {
        http.Error(rw, "error in inserting to database", 400)
        return 
    }
  uj, _ := json.Marshal(user)
  rw.Header().Set("Content-Type", "application/json")
  rw.WriteHeader(201)
  fmt.Fprintf(rw, "%s", uj)
  
}

func get(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    log.Println("inside Get")
    identifier := p.ByName("id")
    //log.Println("id is : ", identifier)
    if !bson.IsObjectIdHex(identifier) {
        rw.WriteHeader(404)
        return
    }
    oid := bson.ObjectIdHex(identifier)
    sess, err := mgo.Dial("mongodb://Richa:Indore#1@ds041934.mongolab.com:41934/assignment_2_db")
    if err != nil {
       fmt.Printf("Can't connect to mongo, go error %v\n", err)
       os.Exit(1)
    }
    defer sess.Close()
    sess.SetSafe(&mgo.Safe{})
    collection := sess.DB("assignment_2_db").C("loc")
    var user UserResponse
    err = collection.Find(bson.M{"_id":oid}).One(&user)
    if err != nil {
    // handle error
       fmt.Fprintf(rw,"Record not found",err)
       return
    }  
    uj, _ := json.Marshal(user)
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", uj)
}

func delete(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
     log.Println("inside Delete")
     identifier := p.ByName("id")
     //log.Println("id is : ", identifier)
    if !bson.IsObjectIdHex(identifier) {
        rw.WriteHeader(404)
        return
    }
    oid := bson.ObjectIdHex(identifier)
    sess, err := mgo.Dial("mongodb://Richa:Indore#1@ds041934.mongolab.com:41934/assignment_2_db")
    if err != nil {
       fmt.Printf("Can't connect to mongo, go error %v\n", err)
       os.Exit(1)
    }
    defer sess.Close()
    sess.SetSafe(&mgo.Safe{})
    collection := sess.DB("assignment_2_db").C("loc")
    err = collection.Remove(bson.M{"_id":oid})
    if err != nil {
    // handle error
       fmt.Printf("Delete Error : ", err)
       //rw.WriteHeader(201)
       fmt.Fprintf(rw,"%s","\"Message\":\"Record not found\"")
       return
    }else{
       fmt.Fprintf(rw,"%s","\"Message\":\"Record deleted successfully\"")
    }

}

func put(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
    log.Println("inside Put/Update")
    identifier := p.ByName("id")
     //log.Println("id is : ", identifier)
    if !bson.IsObjectIdHex(identifier) {
        rw.WriteHeader(404)
        return
    }
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }
   var u User
   err = json.Unmarshal(body, &u)
    oid := bson.ObjectIdHex(identifier)
    sess, err := mgo.Dial("mongodb://Richa:Indore#1@ds041934.mongolab.com:41934/assignment_2_db")
    if err != nil {
       fmt.Printf("Can't connect to mongo, go error %v\n", err)
       os.Exit(1)
    }
    defer sess.Close()
    sess.SetSafe(&mgo.Safe{})
    collection := sess.DB("assignment_2_db").C("loc")
    var user UserResponse  
    err = collection.Find(bson.M{"_id":oid}).One(&user)
    if err != nil {
      fmt.Println("select error ",err)
      return
    } 
    s := addressString(u)
    var lr = new(LocationResponse)
    lr = lat_lng(s) 
    err = collection.Update(bson.M{"_id":oid}, bson.M{"name":u.NameReq,"state": u.StateReq,"city":u.CityReq,"zip":u.ZipReq,"address":u.AddressReq,"coordinate":bson.M{"lat":lr.Results[0].Geometry.Location.Lat,"lng":lr.Results[0].Geometry.Location.Lng}})
    if err != nil {
      fmt.Printf("Can't update document %v\n", err)
      fmt.Fprintf(rw,"%s","\"Message\":\"Update error\"")
      os.Exit(1)
    }
    err = collection.Find(bson.M{"_id":oid}).One(&user)
    if err != nil {
    // handle error
      fmt.Println("Select record error ",err)
    return
    }
    uj, _ := json.Marshal(user)
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", uj)
}

func main() {
    mux := httprouter.New()
    mux.POST("/locations",post)
    mux.GET("/locations/:id",get)
    mux.DELETE("/locations/:id",delete)
    mux.PUT("/locations/:id",put)
    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }
    server.ListenAndServe()
}

func addressString(u User) string{
     var adr string
     var c string
     var s string
     var retString string
     adr=strings.Replace(u.AddressReq," ","+",-1)
     c =strings.Replace(u.CityReq," ","+",-1)
     c = "+"+c
     s = "+"+u.StateReq
     retString = adr + ","+ c + "," + s + "&sensor=false"
     return retString
}

func lat_lng(s string) * LocationResponse{
  url := fmt.Sprintf("http://maps.googleapis.com/maps/api/geocode/json?address=%s",s)
  client := http.Client{Timeout: timeout}
  locresp, err := client.Get(url)
  if err != nil {
         fmt.Println(err)
         //return
  }
  content, err := ioutil.ReadAll(locresp.Body)
  if err != nil {
         fmt.Println(err)
         //return
  }
  var lr = new(LocationResponse)
  err = json.Unmarshal(content, &lr)

  if err != nil { 
        log.Println(err)
  }  
  return lr
}