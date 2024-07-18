package handlers

import (
	"fmt"
	"ms-tts-go/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func GetVoiceList(c *gin.Context) {
	locale := c.Query("l")
	voices, err := utils.VoiceList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if locale != "" {
		filteredVoices := make([]interface{}, 0)
		for _, voice := range voices {
			if strings.Contains(voice.(map[string]interface{})["Locale"].(string), locale) {
				filteredVoices = append(filteredVoices, voice)
			}
		}
		voices = filteredVoices
	}

	_, detail := c.GetQuery("d")
	if detail {
		c.JSON(http.StatusOK, gin.H{"voices": voices})
	} else {
		voiceSimpleList := make([]map[string]string, 0)
		for _, voice := range voices {
			localName := voice.(map[string]interface{})["LocalName"].(string)
			shortName := voice.(map[string]interface{})["ShortName"].(string)
			voiceSimpleList = append(voiceSimpleList, map[string]string{
				"LocalName": localName,
				"ShortName": shortName,
			})
		}
		c.JSON(http.StatusOK, gin.H{"voices": voiceSimpleList})
	}

}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "TTS",
	})
}

type SynthesizeVoiceRequest struct {
	Text         string `json:"t"`
	VoiceName    string `json:"v"`
	Rate         string `json:"r"`
	Pitch        string `json:"p"`
	OutputFormat string `json:"o"`
}

func SynthesizeVoice(c *gin.Context) {
	text := c.Query("t")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text is required"})
		return
	}

	voiceName := c.DefaultQuery("v", "zh-CN-XiaoxiaoMultilingualNeural")
	rate := c.DefaultQuery("r", "0")
	pitch := c.DefaultQuery("p", "0")
	outputFormat := c.DefaultQuery("o", "audio-24khz-48kbitrate-mono-mp3")

	log.Infof("Synthesizing voice. Text: %s, Voice: %s, Rate: %s, Pitch: %s, Format: %s", text, voiceName, rate, pitch, outputFormat)

	voice, err := utils.GetVoice(text, voiceName, rate, pitch, outputFormat)
	if err != nil {
		log.Errorf("Failed to synthesize voice: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Infof("Voice synthesized successfully. Size: %s", utils.ByteCountIEC(int64(len(voice))))
	c.Data(http.StatusOK, "audio/mpeg", voice)
}

func SynthesizeVoicePost(c *gin.Context) {
	var request SynthesizeVoiceRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text is required"})
		return
	}

	log.Infof("Synthesizing voice (POST). Text: %s, Voice: %s, Rate: %s, Pitch: %s, Format: %s",
		request.Text, request.VoiceName, request.Rate, request.Pitch, request.OutputFormat)

	voice, err := utils.GetVoice(request.Text, request.VoiceName, request.Rate, request.Pitch, request.OutputFormat)
	if err != nil {
		log.Errorf("Failed to synthesize voice: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Infof("Voice synthesized successfully. Size: %s", utils.ByteCountIEC(int64(len(voice))))
	c.Data(http.StatusOK, "audio/mpeg", voice)
}

// OpenAIModel 结构体用于表示 OpenAI 模型格式
type OpenAIModel struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int    `json:"created"`
    OwnedBy string `json:"owned_by"`
}

// OpenAIModelList 结构体用于表示模型列表响应
type OpenAIModelList struct {
    Object string        `json:"object"`
    Data   []OpenAIModel `json:"data"`
}

// GetModels 处理 /v1/models 请求
func GetModels(c *gin.Context) {
    voices, err := utils.VoiceList()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "message": "Failed to retrieve voice list",
                "type":    "server_error",
                "param":   "",
                "code":    "",
            },
        })
        return
    }

    models := make([]OpenAIModel, 0, len(voices))
    creationTime := int(time.Now().Unix())

    for _, voice := range voices {
        v := voice.(map[string]interface{})
        model := OpenAIModel{
            ID:      v["ShortName"].(string),
            Object:  "model",
            Created: creationTime,
            OwnedBy: "microsoft",
        }
        models = append(models, model)
    }

    response := OpenAIModelList{
        Object: "list",
        Data:   models,
    }

    // 添加 OpenAI 风格的响应头
    c.Header("OpenAI-Organization", "microsoft-organization-id")
    c.Header("OpenAI-Processing-Ms", "50")
    c.Header("OpenAI-Version", "2023-05-15")
    c.Header("X-Request-ID", utils.GenerateRequestID())

    c.JSON(http.StatusOK, response)
}

// CreateSpeechRequest 结构体用于解析 /v1/audio/speech 请求
type CreateSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
}

// CreateSpeech 处理 /v1/audio/speech 请求
func CreateSpeech(c *gin.Context) {
	var request CreateSpeechRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request body",
				"type":    "invalid_request_error",
				"param":   "",
				"code":    "",
			},
		})
		return
	}

	if request.Input == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Input is required",
				"type":    "invalid_request_error",
				"param":   "input",
				"code":    "parameter_missing",
			},
		})
		return
	}

	// 将 speed 转换为 rate
	rate := fmt.Sprintf("%.0f", (request.Speed-1)*100)

	outputFormat := "audio-24khz-48kbitrate-mono-mp3"
	if request.ResponseFormat == "opus" {
		outputFormat = "ogg-48khz-16bit-mono-opus"
	}

	voice, err := utils.GetVoice(request.Input, request.Voice, rate, "0", outputFormat)
	if err != nil {
		log.Errorf("Failed to synthesize voice: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to synthesize speech",
				"type":    "server_error",
				"param":   "",
				"code":    "",
			},
		})
		return
	}

	contentType := "audio/mpeg"
	if request.ResponseFormat == "opus" {
		contentType = "audio/opus"
	}

	// 添加 OpenAI 风格的响应头
	c.Header("OpenAI-Organization", "microsoft-organization-id")
	c.Header("OpenAI-Processing-Ms", "500")
	c.Header("OpenAI-Version", "2023-05-15")
	c.Header("X-Request-ID", utils.GenerateRequestID())

	log.Infof("Voice synthesized successfully. Size: %s", utils.ByteCountIEC(int64(len(voice))))
	c.Data(http.StatusOK, contentType, voice)
}
