package main

import (
	"encoding/json"
	// "io"
	"io/ioutil"
	"log"
	"net/http"

	// "image"
	"bytes"

	_ "image/gif"
	_ "image/png"
	_ "image/jpeg"

	"github.com/disintegration/imaging"
)

type ReqArgs struct {
	Cmd string `json:"cmd"`
	Src struct {
		Url      string `json:"url"`
		Mimetype string `json:"mimetype"`
		Fsize    int32  `json:"fsize"`
		Bucket   string `json:"bucket"`
		Key      string `json:"key"`
	} `json: "src"`
}

func imageHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Println("read request body failed:", err)
		return
	}
	var args ReqArgs
	err = json.Unmarshal(body, &args)
	if err != nil {
		w.WriteHeader(400)
		log.Println("invalid request body:", err)
		return
	}
	resp, err := http.Get(args.Src.Url)
	if err != nil {
		w.WriteHeader(400)
		log.Println("fetch resource failed:", err)
		return
	}
	defer resp.Body.Close()

	img, err := imaging.Decode(resp.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Println("invalid image of the url", err)
		return
	}

	croppedImg := imaging.Fit(img, 100, 100, imaging.Lanczos)

	if err != nil {
		w.WriteHeader(500)
		log.Println("cannot crop image", err)
		return
	}

	buf := new(bytes.Buffer)
	// err = jpeg.Encode(buf, croppedImg, nil)
	imaging.Encode(buf, croppedImg, imaging.JPEG)
	if err != nil {
		w.WriteHeader(500)
		log.Println("cannot encode jpeg", err)
		return
	}
	result := buf.Bytes()
	w.Write(result)
}

func main() {
	http.HandleFunc("/uop", imageHandler)
	err := http.ListenAndServe(":9100", nil)
	if err != nil {
		log.Fatal("Demo server failed to start:", err)
	}
	log.Println("jerry-ufoptest is up now.")
}

