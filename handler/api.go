package handler

import (
	"net/http"
	"strings"

	"zufen/model"
	"zufen/service"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Type        string `json:"type" binding:"required,oneof=team person"`
	Score       int    `json:"score" binding:"min=0,max=2026"`
	MatchMode   string `json:"match_mode" binding:"required,oneof=exact fuzzy"`
	QRCodeImage string `json:"qrcode_image" binding:"required"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func isValidImageType(base64Image string) bool {
	validPrefixes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/jpg;base64,",
		"data:image/gif;base64,",
	}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(base64Image, prefix) {
			return true
		}
	}
	return false
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证图片大小 (base64 约为原始大小的 1.37 倍，2MB 原图约 2.7MB base64)
	maxBase64Size := 3 * 1024 * 1024
	if len(req.QRCodeImage) > maxBase64Size {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "图片大小不能超过 2MB",
		})
		return
	}

	// 验证图片类型
	if !isValidImageType(req.QRCodeImage) {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "只支持 PNG、JPG、GIF 格式的图片",
		})
		return
	}

	clientIP := c.ClientIP()

	maxPerDay := service.GetConfigInt("max_per_day_ip", 3)
	canSubmit, err := service.CheckIPLimit(clientIP, maxPerDay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	if !canSubmit {
		c.JSON(http.StatusTooManyRequests, Response{
			Code:    429,
			Message: "今日提交次数已达上限，请明天再试",
		})
		return
	}

	qrContent, err := service.DecodeQRCode(req.QRCodeImage)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "二维码解析失败: " + err.Error(),
		})
		return
	}

	if !service.ValidateQRCodeContent(qrContent) {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "二维码内容无效，请使用正确的邀请二维码",
		})
		return
	}

	isUnique, err := service.CheckQRCodeDuplicate(qrContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "系统错误",
		})
		return
	}
	if !isUnique {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "该二维码已被提交过，请勿重复提交",
		})
		return
	}

	p := &model.Participant{
		Type:          model.ParticipantType(req.Type),
		Score:         req.Score,
		MatchMode:     model.MatchMode(req.MatchMode),
		QRCodeContent: qrContent,
		ClientIP:      clientIP,
	}

	if err := service.CreateParticipant(p); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建失败: " + err.Error(),
		})
		return
	}

	qrPath, err := service.SaveQRCodeImage(req.QRCodeImage, p.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "保存二维码失败: " + err.Error(),
		})
		return
	}

	if err := service.UpdateQRCodePath(p.UUID, qrPath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新二维码路径失败: " + err.Error(),
		})
		return
	}

	result, err := service.TryMatch(p.UUID)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "已提交，等待匹配",
			Data: map[string]interface{}{
				"uuid":   p.UUID,
				"status": "pending",
			},
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "匹配成功",
			Data: map[string]interface{}{
				"uuid":       p.UUID,
				"status":     "matched",
				"match_info": result.MatchInfo,
			},
		})
	} else {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "已提交，等待匹配",
			Data: map[string]interface{}{
				"uuid":   p.UUID,
				"status": "pending",
			},
		})
	}
}

func GetStatus(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "UUID不能为空",
		})
		return
	}

	result, err := service.TryMatch(uuid)
	if err != nil {
		p, matchInfo, err := service.GetStatus(uuid)
		if err != nil {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: err.Error(),
			})
			return
		}

		if p.Status == model.StatusMatched && matchInfo != nil {
			c.JSON(http.StatusOK, Response{
				Code:    0,
				Message: "匹配成功",
				Data: map[string]interface{}{
					"status":     "matched",
					"match_info": matchInfo,
				},
			})
		} else {
			c.JSON(http.StatusOK, Response{
				Code:    0,
				Message: "等待匹配中",
				Data: map[string]interface{}{
					"status": "pending",
				},
			})
		}
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "匹配成功",
			Data: map[string]interface{}{
				"status":     "matched",
				"match_info": result.MatchInfo,
			},
		})
	} else {
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "等待匹配中",
			Data: map[string]interface{}{
				"status": "pending",
			},
		})
	}
}

func GetConfig(c *gin.Context) {
	targetScore, fuzzyMin, fuzzyMax := service.GetMatchConfig()

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"target_score": targetScore,
			"fuzzy_min":    fuzzyMin,
			"fuzzy_max":    fuzzyMax,
		},
	})
}
