package panel

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"x-panel/global"
	"x-panel/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIconEditsRespectOwnerAndGroup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:icon_access?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.ItemIconGroup{}, &models.ItemIcon{}); err != nil {
		t.Fatal(err)
	}
	old := global.Db
	global.Db = db
	t.Cleanup(func() { global.Db = old })
	own := models.ItemIconGroup{Title: "own", UserId: 1}
	other := models.ItemIconGroup{Title: "other", UserId: 2}
	if err := db.Create(&own).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	icon := models.ItemIcon{Title: "original", UserId: 2, ItemIconGroupId: int(other.ID)}
	if err := db.Create(&icon).Error; err != nil {
		t.Fatal(err)
	}
	iconJSON, err := json.Marshal(models.ItemIcon{}.Icon)
	if err != nil {
		t.Fatal(err)
	}
	ownIcon := models.ItemIcon{Title: "same", UserId: 1, ItemIconGroupId: int(own.ID), IconJson: string(iconJSON)}
	if err := db.Create(&ownIcon).Error; err != nil {
		t.Fatal(err)
	}

	call := func(handler gin.HandlerFunc, body any) int {
		t.Helper()
		payload, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userInfo", models.User{BaseModel: models.BaseModel{ID: 1}})
		handler(c)
		var result struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Code
	}
	if code := call((&ItemIcon{}).Edit, models.ItemIcon{BaseModel: models.BaseModel{ID: icon.ID}, Title: "stolen", ItemIconGroupId: int(own.ID)}); code == 0 {
		t.Fatal("foreign icon edit succeeded")
	}
	if code := call((&ItemIconGroup{}).Edit, models.ItemIconGroup{BaseModel: models.BaseModel{ID: other.ID}, Title: "stolen"}); code == 0 {
		t.Fatal("foreign group edit succeeded")
	}
	if code := call((&ItemIcon{}).Edit, models.ItemIcon{Title: "bad group", ItemIconGroupId: int(other.ID)}); code == 0 {
		t.Fatal("foreign group accepted by Edit")
	}
	if code := call((&ItemIcon{}).AddMultiple, []models.ItemIcon{{Title: "bad group", ItemIconGroupId: int(other.ID)}}); code == 0 {
		t.Fatal("foreign group accepted by AddMultiple")
	}
	if code := call((&ItemIcon{}).Edit, models.ItemIcon{BaseModel: models.BaseModel{ID: ownIcon.ID}, Title: ownIcon.Title, ItemIconGroupId: ownIcon.ItemIconGroupId}); code != 0 {
		t.Fatalf("unchanged own icon edit failed: code %d", code)
	}
	if code := call((&ItemIconGroup{}).Edit, models.ItemIconGroup{BaseModel: models.BaseModel{ID: own.ID}, Title: own.Title}); code != 0 {
		t.Fatalf("unchanged own group edit failed: code %d", code)
	}
	var got models.ItemIcon
	if err := db.First(&got, icon.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Title != "original" || got.UserId != 2 || got.ItemIconGroupId != int(other.ID) {
		t.Fatalf("icon changed: %+v", got)
	}
	var gotGroup models.ItemIconGroup
	if err := db.First(&gotGroup, other.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotGroup.Title != "other" || gotGroup.UserId != 2 {
		t.Fatalf("group changed: %+v", gotGroup)
	}
}
