package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"in-share-server/app/models"
	"in-share-server/app/router"
	"in-share-server/app/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)
import "in-share-server/app/db"

//Access-Control-Max-Age
type Files struct {
	Router         router.Router
	DatabaseClient db.DatabaseClient
}

type UploadFileFields struct {
	Name             string
	Size             int64
	FileId           string `json:"fileId"`
	BinaryString     string `json:"binaryString"`
	ChunkSize        int64  `json:"chunkSize"`
	Position         int
	TotalChunks      int    `json:"totalChunks"`
	UsersSharedWith  string `json:"usersSharedWith"`
	GroupsSharedWith string `json:"groupsSharedWith"`
}

const collectionName = "files"

const gridFsFilesCollectionName = "fs.files"
const gridFsChunksCollectionName = "fs.chunks"

const maxChunkSize = 255000

func (files *Files) Init() {
	files.Router.Put("/files/new", files.uploadNewFile)
	files.Router.Put("/files/legacy/new", files.uploadFile)
	files.Router.Put("/files/resume", files.resumeFileUpload)
	files.Router.Get("/files/{fileId}/{fileName}", files.downloadFile)
	files.Router.Head("/files/{fileId}/{fileName}", files.downloadFile)
}

func (files *Files) uploadNewFile(w http.ResponseWriter, req *http.Request) {
	user := req.Context().Value("user").(*models.User)

	decoder := json.NewDecoder(req.Body)
	var body UploadFileFields
	decoder.Decode(&body)

	usersSharedWith := strings.Split(body.UsersSharedWith, ",")
	groupsSharedWith := strings.Split(body.GroupsSharedWith, ",")

	fileId := files.DatabaseClient.Insert(gridFsFilesCollectionName, bson.M{"filename": body.Name, "length": body.Size, "chunkSize": maxChunkSize, "uploadDate": time.Now()})

	fileObjectId, _ := primitive.ObjectIDFromHex(utils.ObjectIdToString(fileId))

	fmt.Println("bLength", len(body.BinaryString))
	byteString, err := base64.StdEncoding.DecodeString(body.BinaryString)

	if err != nil {
		fmt.Println("error ", err)
	}
	ioutil.WriteFile("./test.txt", byteString, 0644)

	binary := primitive.Binary{Data: byteString, Subtype: 0}

	fmt.Println("fLength ", len(byteString))
	files.DatabaseClient.Insert(gridFsChunksCollectionName, bson.M{"files_id": fileObjectId, "n": 0, "data": binary, "size": len(body.BinaryString)})

	fileUid := uuid.NewV4()

	files.DatabaseClient.Insert(collectionName, bson.M{"fileId": fileUid.String(), "fileDId": fileObjectId, "fileSize": body.Size, "fileName": body.Name, "userId": user.Id, "ownerId": user.Id, "users": usersSharedWith, "groups": groupsSharedWith})

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, fmt.Sprintf("{\"fileId\": \"%s\", \"fileUid\": \"%s\"}", utils.ObjectIdToString(fileId), fileUid.String()))

}

func (files *Files) resumeFileUpload(w http.ResponseWriter, req *http.Request) {
	//user := req.Context().Value("user").(*models.User)

	decoder := json.NewDecoder(req.Body)
	var body UploadFileFields
	decoder.Decode(&body)

	fileObjectId, _ := primitive.ObjectIDFromHex(body.FileId)

	binary := primitive.Binary{Data: []byte(body.BinaryString), Subtype: 0}
	files.DatabaseClient.Insert(gridFsChunksCollectionName, bson.M{"files_id": fileObjectId, "n": body.Position, "data": binary, "size": len(body.BinaryString)})

	fmt.Fprintf(w, "done")

}
func (files *Files) uploadFile(w http.ResponseWriter, req *http.Request) {
	user := req.Context().Value("user").(*models.User)

	usersSharedWith := strings.Split(req.FormValue("usersSharedWith"), ",")
	groupsSharedWith := strings.Split(req.FormValue("groupsSharedWith"), ",")

	//TODO optional validate usersSharedWith, groupsSharedWith

	file, header, err := req.FormFile("file")

	if req.Header.Get("x-contains-all") == "1" {
		// request contains all files. max 255 kilobytes

	}

	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)

	fileId, _ := files.DatabaseClient.InsertFile(data, header.Filename)

	fmt.Println("fid", fileId)
	fileUid := uuid.NewV4()

	files.DatabaseClient.Insert(collectionName, bson.M{"fileId": fileUid.String(), "fileDId": fileId, "fileSize": len(data), "fileName": header.Filename, "userId": user.Id, "ownerId": user.Id, "users": usersSharedWith, "groups": groupsSharedWith})

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fileUid.String())

}

func (files *Files) downloadFile(w http.ResponseWriter, req *http.Request) {

	user := req.Context().Value("user").(*models.User)

	fmt.Println(req.Method)

	vars := mux.Vars(req)
	fileId := vars["fileId"]

	fileName := vars["fileName"]

	file := &models.File{}

	fmt.Println("here")

	err := files.DatabaseClient.FindOne(collectionName, bson.M{"fileId": fileId}, file)

	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !userHasAccess(user, file) {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")

	if req.Method == "HEAD" {
		w.Header().Set("Content-Length", strconv.Itoa(file.FileSize))
		w.Write(nil)
		return
	}

	r := req.Header.Get("range")

	var fileBytes []byte
	if r != "" {
		fmt.Println("range ", r)
		byteRange := strings.Split(strings.Split(r, "=")[1], "-")
		start, _ := strconv.Atoi(byteRange[0])
		end := 0
		if byteRange[1] != "" {
			end, _ = strconv.Atoi(byteRange[1])
		} else {
			end = file.FileSize - 1
		}

		chunkSize := (end - start) + 1

		fmt.Println("chunk size", chunkSize)
		w.Header().Set("Content-Length", strconv.Itoa(chunkSize))
		contentRage := fmt.Sprintf("bytes %d-%d/%d ", start, end, file.FileSize)
		fmt.Println(contentRage)
		w.Header().Set("Content-Range", contentRage)
		w.WriteHeader(http.StatusPartialContent)

		skip := 0

		if start != 0 {
			skip = start + 1
			fmt.Println("skip ", skip)
		}
		fileBytes, _ = files.DatabaseClient.DownloadFile(file.FileDid, chunkSize, int64(skip))
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(file.FileSize))
		fileBytes, _ = files.DatabaseClient.DownloadFile(file.FileDid, file.FileSize, 0)
	}

	fmt.Println("to return ", len(fileBytes))
	//fmt.Println("to return ", fileBytes)

	w.Write(fileBytes)

}

func userHasAccess(user *models.User, file *models.File) bool {

	if file.OwnerId == user.Id {
		return true
	}

	if (utils.Include(file.Groups, "general") && !user.Public) || utils.Include(file.Groups, "public") {
		return true
	}

	for _, group := range file.Groups {
		if utils.Include(user.Groups, group) {
			return true
		}
	}

	for _, fileUserId := range file.Users {
		if user.Id == fileUserId {
			return true
		}
	}
	return false
}
