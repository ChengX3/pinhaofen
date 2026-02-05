package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zufen/database"
	"zufen/model"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MatchResult struct {
	Success   bool
	MatchInfo *model.MatchInfo
}

func DecodeQRCode(base64Image string) (string, error) {
	base64Data := base64Image
	if idx := strings.Index(base64Image, ","); idx != -1 {
		base64Data = base64Image[idx+1:]
	}

	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", errors.New("无效的base64图片数据")
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return "", errors.New("无法解码图片")
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", errors.New("无法处理图片")
	}

	reader := qrcode.NewQRCodeReader()
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return "", errors.New("无法识别二维码")
	}

	content := result.GetText()
	log.Printf("[DEBUG] 二维码识别内容: %s", content)

	return content, nil
}

func ValidateQRCodeContent(content string) bool {
	prefix := GetValidURLPrefix()
	if prefix == "" {
		return true
	}
	return strings.HasPrefix(content, prefix)
}

func SaveQRCodeImage(base64Image string, uuid string) (string, error) {
	uploadDir := GetConfigValue("upload_dir")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	base64Data := base64Image
	ext := ".png"
	if idx := strings.Index(base64Image, ","); idx != -1 {
		header := base64Image[:idx]
		base64Data = base64Image[idx+1:]
		if strings.Contains(header, "jpeg") || strings.Contains(header, "jpg") {
			ext = ".jpg"
		} else if strings.Contains(header, "gif") {
			ext = ".gif"
		}
	}

	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}

	filename := uuid + ext
	filePath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func CheckIPLimit(clientIP string, maxPerDay int) (bool, error) {
	db := database.Get()

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int64
	if err := db.Model(&model.Participant{}).
		Where("client_ip = ?", clientIP).
		Where("created_at >= ?", todayStart).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count < int64(maxPerDay), nil
}

func CheckQRCodeDuplicate(qrContent string) (bool, error) {
	db := database.Get()

	var count int64
	if err := db.Model(&model.Participant{}).
		Where("qrcode_content = ?", qrContent).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count == 0, nil
}

func CreateParticipant(p *model.Participant) error {
	p.UUID = model.GenerateUUID()
	p.Status = model.StatusPending

	return database.Get().Create(p).Error
}

func UpdateQRCodePath(uuid string, qrPath string) error {
	return database.Get().Model(&model.Participant{}).
		Where("uuid = ?", uuid).
		Update("qrcode_path", qrPath).Error
}

func TryMatch(uuid string) (*MatchResult, error) {
	db := database.Get()
	targetScore, fuzzyMin, fuzzyMax := GetMatchConfig()

	var result *MatchResult

	err := db.Transaction(func(tx *gorm.DB) error {
		var self model.Participant
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("uuid = ?", uuid).
			First(&self).Error; err != nil {
			return err
		}

		if self.Status == model.StatusMatched {
			var matched model.Participant
			if err := tx.Where("uuid = ?", *self.MatchedUUID).First(&matched).Error; err != nil {
				return err
			}
			result = &MatchResult{
				Success: true,
				MatchInfo: &model.MatchInfo{
					Score:      matched.Score,
					QRCodePath: matched.QRCodePath,
				},
			}
			return nil
		}

		var oppositeType model.ParticipantType
		if self.Type == model.TypeTeam {
			oppositeType = model.TypePerson
		} else {
			oppositeType = model.TypeTeam
		}

		needScore := targetScore - self.Score

		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("type = ?", oppositeType).
			Where("status = ?", model.StatusPending).
			Where("uuid != ?", uuid).
			Order("created_at ASC")

		if self.MatchMode == model.ModeExact {
			query = query.Where("score = ?", needScore)
		} else {
			query = query.Where("match_mode = ?", model.ModeFuzzy).
				Where("score + ? >= ?", self.Score, fuzzyMin).
				Where("score + ? <= ?", self.Score, fuzzyMax)
		}

		var candidate model.Participant
		if err := query.First(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				result = &MatchResult{Success: false}
				return nil
			}
			return err
		}

		if err := tx.Model(&model.Participant{}).
			Where("uuid = ?", self.UUID).
			Updates(map[string]interface{}{
				"status":       model.StatusMatched,
				"matched_uuid": candidate.UUID,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Participant{}).
			Where("uuid = ?", candidate.UUID).
			Updates(map[string]interface{}{
				"status":       model.StatusMatched,
				"matched_uuid": self.UUID,
			}).Error; err != nil {
			return err
		}

		result = &MatchResult{
			Success: true,
			MatchInfo: &model.MatchInfo{
				Score:      candidate.Score,
				QRCodePath: candidate.QRCodePath,
			},
		}

		return nil
	})

	return result, err
}

func GetStatus(uuid string) (*model.Participant, *model.MatchInfo, error) {
	db := database.Get()

	var p model.Participant
	if err := db.Where("uuid = ?", uuid).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("UUID不存在")
		}
		return nil, nil, err
	}

	if p.Status == model.StatusMatched && p.MatchedUUID != nil {
		var matched model.Participant
		if err := db.Where("uuid = ?", *p.MatchedUUID).First(&matched).Error; err != nil {
			return &p, nil, nil
		}
		return &p, &model.MatchInfo{
			Score:      matched.Score,
			QRCodePath: matched.QRCodePath,
		}, nil
	}

	return &p, nil, nil
}
