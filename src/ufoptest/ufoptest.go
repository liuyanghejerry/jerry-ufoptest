package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"image"
	"bytes"

	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images. Uncomment these
	// two lines to also understand GIF and PNG images:
	_ "image/gif"
	_ "image/png"
	jpeg "image/jpeg"

	_ "ufop"
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

func demoHandler(w http.ResponseWriter, req *http.Request) {
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
	var buf = make([]byte, 512)
	io.ReadFull(resp.Body, buf)
	contentType := http.DetectContentType(buf)
	w.Write([]byte(contentType))
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
	// var buf = make([]byte, 512)
	// io.ReadFull(resp.Body, buf)
	// contentType := http.DetectContentType(buf)
	// w.Write([]byte(contentType))
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Println("invalid image of the url", err)
		return
	}

	croppedImage := image.NewRGBA(img.Bounds())
	img = croppedImage.SubImage(image.Rect(0, 0, 20, 20))

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
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
