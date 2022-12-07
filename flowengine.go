package microgate

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/workanator/go-floc/v3"
	"net/http"
)

var restclient *resty.Client

func GetRestClient() *resty.Client {
	if restclient == nil {
		restclient = resty.New()
	}
	return restclient
}

type RestRequestData struct {
	QueryParams map[string]string
	Header      map[string]string
	Token       string
	BaseUrl     string
	Path        string
	Method      string
}

func Print(restCallKey int, status, header, resp bool) func(ctx floc.Context, ctrl floc.Control) error {
	return func(ctx floc.Context, ctrl floc.Control) error {
		response := ctx.Value(restCallKey).(*resty.Response)
		if status {
			fmt.Println(response.Status())
		}
		if header {
			for key, value := range response.Header() {
				fmt.Println(key+":", value)
			}
		}
		if resp {
			fmt.Println(string(response.Body()))
		}
		return nil
	}
}

type MappingTable struct {
	Mappings []Mapping
}

type Mapping struct {
	FromElement  MappingElement
	ToElement    MappingElement
	OptConverter func(from MappingElement) MappingElement
}

type MappingElement struct {
	PathToObject string
	Key          string
	Type         string
}

func (m MappingTable) DoMapping(restCallKey int) func(ctx floc.Context, ctrl floc.Control) error {
	return func(ctx floc.Context, ctrl floc.Control) error {
		response := ctx.Value(restCallKey).(*resty.Response)
		fromObject := MarshalToDataObject(response.Body())
		toObject := NewDataObject()
		for _, mapping := range m.Mappings {
			fromElement := ""
			subobj := fromObject.GetObjectByPath(mapping.FromElement.PathToObject)
			if mapping.FromElement.Type == "string" {
				fromElement = subobj.GetString(mapping.FromElement.Key)
				subTo := toObject.PutObjectByPath(mapping.ToElement.PathToObject)
				subTo.Put(mapping.ToElement.Key, fromElement)
			}
		}
		fmt.Println(&toObject)
		return nil
	}
}

type LimeObject map[string]interface{}

func RestCall(restCallKey int, r *RestRequestData) func(ctx floc.Context, ctrl floc.Control) error {
	return func(ctx floc.Context, ctrl floc.Control) error {

		client := GetRestClient()
		request := client.R()
		if len(r.QueryParams) > 0 {
			request.SetQueryParams(r.QueryParams)
		}
		if len(r.Header) > 0 {
			request.SetHeaders(r.Header)
		}
		if r.Token != "" {
			request.SetAuthToken(r.Token)
		}
		var response *resty.Response
		var err error
		if r.Method == "GET" {
			response, err = request.Get(r.BaseUrl + r.Path)
		} else if r.Method == "POST" {
			response, err = request.Post(r.BaseUrl + r.Path)
		} else if r.Method == "PUT" {
			response, err = request.Put(r.BaseUrl + r.Path)
		} else if r.Method == "DELETE" {
			response, err = request.Delete(r.BaseUrl + r.Path)
		} else if r.Method == "HEAD" {
			response, err = request.Head(r.BaseUrl + r.Path)
		} else if r.Method == "OPTIONS" {
			response, err = request.Options(r.BaseUrl + r.Path)
		} else if r.Method == "PATCH" {
			response, err = request.Patch(r.BaseUrl + r.Path)
		}
		if err != nil {
			return err
		}
		ctx.AddValue(restCallKey, response)
		return nil
	}
}

func ProcessHandler(jobs floc.Job) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ct := floc.NewContext()
		ctrl := floc.NewControl(ct)
		_, _, err := floc.RunWith(ct, ctrl, jobs)
		if err != nil {
			panic(err)
		}
		result := ct.Value("finalresult")
		ctx.JSON(http.StatusOK, result)
	}
}
