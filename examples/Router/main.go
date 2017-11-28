package main

import (
	"log"
	"time"

	"github.com/Chyroc/gor"
	"fmt"
)

func Logger(req *gor.Req, res *gor.Res, next gor.Next) {
	req.AddContext("time", time.Now())

	next()

	fmt.Printf("========================")

	//startTime := req.GetContext("time").(time.Time)
	//log.Printf("startTime %+v , endTime %+v\n", startTime.UTC(), time.Now().UTC())
}

func main() {
	app := gor.NewGor()
	router := gor.NewRouter()

	router.Get("/sub/:uu", func(req *gor.Req, res *gor.Res) {
		time.Sleep(time.Microsecond * 100)
		res.JSON(map[string]interface{}{
			"query":  req.Query,
			"params": req.Params,
		})
	})
	app.Use(Logger)

	app.Use("/m", router)

	log.Fatal(app.Listen(":3000"))
}
