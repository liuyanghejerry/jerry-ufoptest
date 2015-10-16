package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"bytes"
	"strconv"
	"io"
	"regexp"

	"image"
	"image/draw"
	"image/gif"

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

func parseCmd(cmd string) (width int, height int) {
	capWidth := regexp.MustCompile(`\/w\/([0-9]+)`)
	capHeight := regexp.MustCompile(`\/h\/([0-9]+)`)

	w, h := 100, 100

	results := capWidth.FindStringSubmatch(cmd)

	if len(results) < 2 {
		w = 100
	}

	w, err := strconv.Atoi(results[1])

	if err != nil {
		w = 100
	}

	results = capHeight.FindStringSubmatch(cmd)

	if len(results) < 2 {
		h = 100
	}

	h, err = strconv.Atoi(results[1])

	if err != nil {
		h = 100
	}

	return w, h
}

func thumbImage(r io.Reader, width int, height int) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)

	imageData, err := ioutil.ReadAll(r)

	if err != nil {
		log.Println("cannot read response", err)
		return nil, err
	}

	imageBuffer := bytes.NewBuffer(imageData)
	img, formatString, err := image.Decode(imageBuffer)
	if err != nil {
		log.Println("cannot decode image", err)
		return
	}

	switch formatString {
	case "jpg":
		fallthrough
	case "jpeg":
		croppedImg := imaging.Thumbnail(img, width, height, imaging.Lanczos)
		imaging.Encode(buf, croppedImg, imaging.JPEG)
		return
	case "png":
		croppedImg := imaging.Thumbnail(img, width, height, imaging.Lanczos)
		imaging.Encode(buf, croppedImg, imaging.PNG)
	case "bmp":
		croppedImg := imaging.Thumbnail(img, width, height, imaging.Lanczos)
		imaging.Encode(buf, croppedImg, imaging.BMP)
		return
	case "gif":
		imageBuffer = bytes.NewBuffer(imageData)
		g, err := gif.DecodeAll(imageBuffer)
		if err != nil {
			log.Println("cannot decode gif", err)
			return nil, err
		}
		for i := range g.Image {
			thumb := imaging.Thumbnail(g.Image[i], width, height, imaging.Lanczos)
			g.Image[i] = image.NewPaletted(image.Rect(0, 0, width, height), g.Image[i].Palette)
			draw.Draw(g.Image[i], image.Rect(0, 0, width, height), thumb, image.Pt(0, 0), draw.Over)
		}
		g.Config.Width, g.Config.Height = width, height
		err = gif.EncodeAll(buf, g)
		if err != nil {
			log.Println("cannot encode gif", err)
			return nil, err
		}
	}
	return
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

	log.Println("processing cmd:", args.Cmd)
	width, height := parseCmd(args.Cmd)
	log.Println("width, height: ", width, height)

	log.Println("processing url:", args.Src.Url)
	resp, err := http.Get(args.Src.Url)
	if err != nil {
		w.WriteHeader(400)
		log.Println("fetch resource failed:", err)
		return
	}
	defer resp.Body.Close()

	buf, err := thumbImage(resp.Body, width, height)
	if err != nil {
		w.WriteHeader(500)
		log.Println("cannot encode or decode image: ", err)
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

