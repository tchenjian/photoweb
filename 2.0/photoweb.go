package main

import (
	"path"
	"io"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"html/template"
	"path/filepath"
	"time"
	"strings"
)

var	root=""
const(
	Assert_dir="/assert"
	Upload_dir="/uploads"
	View_dir="/views"
)
var MyTemplates=make(map[string] *template.Template)
//给templates加载所有views文件夹下的模版文件
func Loadtmpl(){	
	fileInfoArr,err:=ioutil.ReadDir(root+View_dir)
	check(err)
	
	var temlateName,temlatePath string
	for _,fileInfo:=range fileInfoArr{
		temlateName=fileInfo.Name()
		log.Println(temlateName)
		//检查后缀名
		if ext:=path.Ext(temlateName);ext!=".html"{
			continue
		}
		temlatePath=root+View_dir+"/"+temlateName
		log.Println("Loadtmpl()",temlatePath)
		t:=template.Must(template.ParseFiles(temlatePath))
		MyTemplates[temlateName]=t
		log.Println(len(MyTemplates),MyTemplates[temlateName])
	}
}
func check(err error){
	if err!=nil{
		panic(err)
	}
}
func renderHtml(w http.ResponseWriter,tmpl string,locals map[string] interface{}){	
	err:=MyTemplates[tmpl+".html"].Execute(w,locals)
	check(err)
}

func checkUploadDir(){
	Uploadinfo, err := os.Stat(root+Upload_dir)
	if err != nil {
	    os.Mkdir(root+Upload_dir,os.ModePerm)
		log.Println("os.Mkdir("+root+Upload_dir)
	    return
	}
	if Uploadinfo.IsDir() {
	    // it's a file
	} else {
	    os.Mkdir(root+Upload_dir,os.ModePerm)
	}

	_, err= os.Stat(root+View_dir)
	if err != nil {
		log.Fatal("views file not found! require template file")
	    // no such file or dir
	    return
	}

}

func uploadHandler(w http.ResponseWriter,r *http.Request){
	
	switch r.Method{
		case "GET":
			renderHtml(w,"upload",nil)
		case "POST":
			r.ParseMultipartForm(32<<20)
			files:=r.MultipartForm.File["image"]
			len:=len(files)
			var tmpPath string
			for i:=0;i<len;i++{
				file,err:=files[i].Open()
				defer file.Close()
				check(err)
				t:=time.Now()
				tmpPath =t.Format("200601021504")
				dir :=filepath.Join(photoph,tmpPath)
				os.Mkdir(dir,os.ModePerm)

				cur,err:= os.Create(filepath.Join(dir,files[i].Filename))
				defer cur.Close()
				check(err)

				_,err=io.Copy(cur,file)
				check(err)
				log.Println(files[i].Filename)
			}
			http.Redirect(w,r,"/list?f="+tmpPath,http.StatusFound)
	}
}


func viewHandler(w http.ResponseWriter,r *http.Request){
	imageid:=r.FormValue("id")

	keys, ok := r.URL.Query()["f"]
	var key string
	if ok && len(keys) >0 {
		key = keys[0]
	}

	imagepath:=filepath.Join(photoph,key,imageid)
	if _,err:=os.Stat(imagepath);err!=nil{
		http.NotFound(w,r)
	}
	w.Header().Set("Content-Type","image")
	http.ServeFile(w,r,imagepath)
}

func listHandler(w http.ResponseWriter,r *http.Request){
	keys, ok := r.URL.Query()["f"]
	var key string
	if ok && len(keys) >0 {
		key = keys[0]
		if key!=""{
			//root+"/"+"uploads/"+key)
			if _, err := os.Stat(filepath.Join(photoph,key)); err != nil {
				key =""
			}
		}
	}

	fileInfoArr,err:=ioutil.ReadDir(filepath.Join(photoph,key))

	check(err)
	
	locals:=make(map[string]interface{})
	images:=[]string{}
	dirs:=[]string{}
	for _,fileInfo:=range fileInfoArr{
		if fileInfo.IsDir(){
			if _,err:=os.Stat(filepath.Join(photoph,fileInfo.Name()));err ==nil{
				dirmid,err:=ioutil.ReadDir(filepath.Join(photoph,fileInfo.Name()))
				check(err)
				has:=false
				for _,fi:=range dirmid{
					if strings.HasSuffix(fi.Name(),".png")||strings.HasSuffix(fi.Name(),".jpg"){
						has = true
						break
					}
				}
				if has{
					dirs =append(dirs,fileInfo.Name())
				}
			}
		}else {
			if strings.HasSuffix(fileInfo.Name(),".png")||strings.HasSuffix(fileInfo.Name(),".jpg"){
				images=append(images,fileInfo.Name())
			}
		}
	}

	locals["images"]=images
	locals["dirs"]=dirs
	
	renderHtml(w,"list",locals)
}
func staticDirHandler(mux *http.ServeMux,prefix string,staticDir string,flags int){
	mux.HandleFunc(prefix,
		func(w http.ResponseWriter,r *http.Request){
			log.Println(r.URL.Path)
			file:=root+r.URL.Path
			log.Println(file)
			http.ServeFile(w,r,file)
		})
}

var photoph string
func main(){
    var mux=http.NewServeMux()
	
	root,_=os.Getwd()
	photoph=filepath.Join(root,Upload_dir)
	if len(os.Args)>1{
		photoph=os.Args[1]
	}

	checkUploadDir()
	Loadtmpl()
	staticDirHandler(mux,"/assets/",root+"/assets",0)
	mux.HandleFunc("/upload",uploadHandler)
	mux.HandleFunc("/list",listHandler)
	mux.HandleFunc("/",listHandler)
	mux.HandleFunc("/views",viewHandler)
	err:=http.ListenAndServe(":8090",mux)
	log.Println("http.ListenAndServe(:8090)")
	if err!=nil{
		log.Fatal("ListenAndServe:",err.Error())
	}
}
