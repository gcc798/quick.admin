package utils

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBindJSONWithTypeCasting_StringInt64_SingleField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type requestBody struct {
		EnvId int64 `json:"envId" typecast:"stringInt64"`
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var body requestBody
		if err := BindJSONWithTypeCasting(c, &body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, body)
	})

	reqBody := map[string]interface{}{
		"envId": "1234567890123456789",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1234567890123456789), resp["envId"])
}

func TestBindJSONWithTypeCasting_StringInt64_ArrayField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type requestBody struct {
		IDs []int64 `json:"ids" typecast:"stringInt64"`
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var body requestBody
		if err := BindJSONWithTypeCasting(c, &body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, body)
	})

	reqBody := map[string]interface{}{
		"ids": []interface{}{"1", "2", "3"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	ids, ok := resp["ids"].([]interface{})
	if assert.True(t, ok) && assert.Len(t, ids, 3) {
		assert.Equal(t, float64(1), ids[0])
		assert.Equal(t, float64(2), ids[1])
		assert.Equal(t, float64(3), ids[2])
	}
}

func TestBindJSONWithTypeCasting_ToTime_SingleField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type requestBody struct {
		CreatedAt time.Time `json:"createdAt" typecast:"toTime"`
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var body requestBody
		if err := BindJSONWithTypeCasting(c, &body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, body)
	})

	sec := int64(1700000000)
	reqBody := map[string]interface{}{
		"createdAt": sec,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp struct {
		CreatedAt time.Time `json:"createdAt"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, time.Unix(sec, 0).UTC(), resp.CreatedAt.UTC())
}
