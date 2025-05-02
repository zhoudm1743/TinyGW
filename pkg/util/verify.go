package util

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io/ioutil"
	"mime/multipart"
	"tinyGW/pkg/plugin/response"
)

var VerifyUtil = verifyUtil{}

// verifyUtil 参数验证工具类
type verifyUtil struct{}

func (vu verifyUtil) VerifyJSON(c *gin.Context, obj any) (e error) {
	if err := c.ShouldBindBodyWith(obj, binding.JSON); err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

func (vu verifyUtil) VerifyJSONArray(c *gin.Context, obj any) (e error) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	err = json.Unmarshal(body, &obj)
	if err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

func (vu verifyUtil) VerifyBody(c *gin.Context, obj any) (e error) {
	if err := c.ShouldBind(obj); err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

func (vu verifyUtil) VerifyHeader(c *gin.Context, obj any) (e error) {
	if err := c.ShouldBindHeader(obj); err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

func (vu verifyUtil) VerifyQuery(c *gin.Context, obj any) (e error) {
	if err := c.ShouldBindQuery(obj); err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

func (vu verifyUtil) VerifyFile(c *gin.Context, name string) (file *multipart.FileHeader, e error) {
	file, err := c.FormFile(name)
	if err != nil {
		e = response.ParamsValidError.MakeData(err.Error())
		return
	}
	return
}

// VerifyData 验证请求数据
func (vu verifyUtil) Verify(c *gin.Context, obj ...any) (e error) {
	for _, o := range obj {
		switch c.Request.Method {
		case "POST":
			return VerifyUtil.VerifyJSON(c, o)
		case "GET":
			return VerifyUtil.VerifyQuery(c, o)
		case "PUT":
			return VerifyUtil.VerifyJSON(c, o)
		case "DELETE":
			return VerifyUtil.VerifyQuery(c, o)
		default:
			return response.ParamsValidError.MakeData("请求方式错误")
		}
	}
	return
}
