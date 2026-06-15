package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"donation-site/services/api/internal/config"
	apimw "donation-site/services/api/internal/http/middleware"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/service"
	"donation-site/services/api/internal/service/payment"
	"donation-site/services/api/internal/util"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminHandler struct {
	cfg      config.Config
	db       *gorm.DB
	admin    *service.AdminService
	export   *service.ExportService
	overview *service.OverviewService
}

func NewAdminHandler(cfg config.Config, db *gorm.DB, admin *service.AdminService, export *service.ExportService, overview *service.OverviewService) *AdminHandler {
	return &AdminHandler{cfg: cfg, db: db, admin: admin, export: export, overview: overview}
}

type tierRequest struct {
	Name        string `json:"name"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Enabled     *bool  `json:"enabled"`
	IsDefault   bool   `json:"is_default"`
}

type tierPatchRequest struct {
	Name        *string `json:"name"`
	AmountCents *int64  `json:"amount_cents"`
	Currency    *string `json:"currency"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
	Enabled     *bool   `json:"enabled"`
	IsDefault   *bool   `json:"is_default"`
}

func (h *AdminHandler) ListTiers(c fiber.Ctx) error {
	var tiers []model.DonationTier
	if err := h.db.WithContext(c).Order("sort_order ASC, created_at ASC").Find(&tiers).Error; err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"items": tiers})
}

func (h *AdminHandler) CreateTier(c fiber.Ctx) error {
	var req tierRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	tier := model.DonationTier{Name: strings.TrimSpace(req.Name), AmountCents: req.AmountCents, Currency: normalizeCurrency(req.Currency), Description: strings.TrimSpace(req.Description), SortOrder: req.SortOrder, Enabled: enabled, IsDefault: req.IsDefault}
	if err := validateTier(tier); err != nil {
		return err
	}
	if err := h.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if tier.IsDefault {
			if err := tx.Model(&model.DonationTier{}).Where("is_default = true").Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&tier).Error
	}); err != nil {
		return err
	}
	_ = h.audit(c, "tier.create", "donation_tier", tier.ID.String(), model.JSONMap{"name": tier.Name})
	return respondOK(c, tier)
}

func (h *AdminHandler) UpdateTier(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("档位 ID 无效", nil)
	}
	var req tierPatchRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	var tier model.DonationTier
	if err := h.db.WithContext(c).First(&tier, "id = ?", id).Error; err != nil {
		return notFoundOrErr(err, "档位不存在")
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
		tier.Name = updates["name"].(string)
	}
	if req.AmountCents != nil {
		updates["amount_cents"] = *req.AmountCents
		tier.AmountCents = *req.AmountCents
	}
	if req.Currency != nil {
		updates["currency"] = normalizeCurrency(*req.Currency)
		tier.Currency = updates["currency"].(string)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
		tier.Description = updates["description"].(string)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
		tier.IsDefault = *req.IsDefault
	}
	if err := validateTier(tier); err != nil {
		return err
	}
	if err := h.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if req.IsDefault != nil && *req.IsDefault {
			if err := tx.Model(&model.DonationTier{}).Where("id <> ?", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.DonationTier{}).Where("id = ?", id).Updates(updates).Error
	}); err != nil {
		return err
	}
	_ = h.audit(c, "tier.update", "donation_tier", id.String(), model.JSONMap{})
	return respondOK(c, fiber.Map{"ok": true})
}

func (h *AdminHandler) DeleteTier(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("档位 ID 无效", nil)
	}
	if err := h.db.WithContext(c).Model(&model.DonationTier{}).Where("id = ?", id).Updates(map[string]any{"enabled": false, "updated_at": time.Now().UTC()}).Error; err != nil {
		return err
	}
	_ = h.audit(c, "tier.delete", "donation_tier", id.String(), model.JSONMap{"soft_delete": true})
	return respondOK(c, fiber.Map{"ok": true})
}

type paymentMethodRequest struct {
	Code           string        `json:"code"`
	Name           string        `json:"name"`
	Type           string        `json:"type"`
	Provider       string        `json:"provider"`
	IconURL        string        `json:"icon_url"`
	ConfigJSON     model.JSONMap `json:"config_json"`
	Enabled        *bool         `json:"enabled"`
	SortOrder      int           `json:"sort_order"`
	MinAmountCents *int64        `json:"min_amount_cents"`
	MaxAmountCents *int64        `json:"max_amount_cents"`
}

type paymentMethodPatchRequest struct {
	Code           *string        `json:"code"`
	Name           *string        `json:"name"`
	Type           *string        `json:"type"`
	Provider       *string        `json:"provider"`
	IconURL        *string        `json:"icon_url"`
	ConfigJSON     *model.JSONMap `json:"config_json"`
	Enabled        *bool          `json:"enabled"`
	SortOrder      *int           `json:"sort_order"`
	MinAmountCents *int64         `json:"min_amount_cents"`
	MaxAmountCents *int64         `json:"max_amount_cents"`
}

func (h *AdminHandler) ListPaymentMethods(c fiber.Ctx) error {
	var methods []model.PaymentMethod
	if err := h.db.WithContext(c).Order("sort_order ASC, created_at ASC").Find(&methods).Error; err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"items": methods})
}

func (h *AdminHandler) CreatePaymentMethod(c fiber.Ctx) error {
	var req paymentMethodRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = strings.TrimSpace(req.Type)
	}
	method := model.PaymentMethod{Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), Type: strings.TrimSpace(req.Type), Provider: provider, IconURL: strings.TrimSpace(req.IconURL), ConfigJSON: req.ConfigJSON, Enabled: enabled, SortOrder: req.SortOrder, MinAmountCents: req.MinAmountCents, MaxAmountCents: req.MaxAmountCents}
	if method.ConfigJSON == nil {
		method.ConfigJSON = model.JSONMap{}
	}
	if err := validatePaymentMethod(method); err != nil {
		return err
	}
	if err := h.db.WithContext(c).Create(&method).Error; err != nil {
		return err
	}
	_ = h.audit(c, "payment_method.create", "payment_method", method.ID.String(), model.JSONMap{"code": method.Code})
	return respondOK(c, method)
}

func (h *AdminHandler) UpdatePaymentMethod(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("支付方式 ID 无效", nil)
	}
	var req paymentMethodPatchRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	var method model.PaymentMethod
	if err := h.db.WithContext(c).First(&method, "id = ?", id).Error; err != nil {
		return notFoundOrErr(err, "支付方式不存在")
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Code != nil {
		method.Code = strings.TrimSpace(*req.Code)
		updates["code"] = method.Code
	}
	if req.Name != nil {
		method.Name = strings.TrimSpace(*req.Name)
		updates["name"] = method.Name
	}
	if req.Type != nil {
		method.Type = strings.TrimSpace(*req.Type)
		updates["type"] = method.Type
	}
	if req.Provider != nil {
		method.Provider = strings.TrimSpace(*req.Provider)
		updates["provider"] = method.Provider
	}
	if req.IconURL != nil {
		method.IconURL = strings.TrimSpace(*req.IconURL)
		updates["icon_url"] = method.IconURL
	}
	if req.ConfigJSON != nil {
		method.ConfigJSON = *req.ConfigJSON
		if method.ConfigJSON == nil {
			method.ConfigJSON = model.JSONMap{}
		}
		updates["config_json"] = method.ConfigJSON
	}
	if req.Enabled != nil {
		method.Enabled = *req.Enabled
		updates["enabled"] = method.Enabled
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.MinAmountCents != nil {
		method.MinAmountCents = req.MinAmountCents
		updates["min_amount_cents"] = *req.MinAmountCents
	}
	if req.MaxAmountCents != nil {
		method.MaxAmountCents = req.MaxAmountCents
		updates["max_amount_cents"] = *req.MaxAmountCents
	}
	if err := validatePaymentMethod(method); err != nil {
		return err
	}
	if err := h.db.WithContext(c).Model(&model.PaymentMethod{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	_ = h.audit(c, "payment_method.update", "payment_method", id.String(), model.JSONMap{})
	return respondOK(c, fiber.Map{"ok": true})
}

func (h *AdminHandler) DeletePaymentMethod(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("支付方式 ID 无效", nil)
	}
	if err := h.db.WithContext(c).Model(&model.PaymentMethod{}).Where("id = ?", id).Updates(map[string]any{"enabled": false, "updated_at": time.Now().UTC()}).Error; err != nil {
		return err
	}
	_ = h.audit(c, "payment_method.delete", "payment_method", id.String(), model.JSONMap{"soft_delete": true})
	return respondOK(c, fiber.Map{"ok": true})
}

func (h *AdminHandler) UploadPaymentMethodQR(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("支付方式 ID 无效", nil)
	}
	file, err := c.FormFile("file")
	if err != nil {
		return service.Validation("请上传二维码图片", nil)
	}
	ext, err := validateUploadedImage(file)
	if err != nil {
		return err
	}
	var method model.PaymentMethod
	if err := h.db.WithContext(c).First(&method, "id = ?", id).Error; err != nil {
		return notFoundOrErr(err, "支付方式不存在")
	}
	dir := filepath.Join(h.cfg.UploadDir, "payment-methods")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	name := uuid.NewString() + ext
	dest := filepath.Join(dir, name)
	if err := c.SaveFile(file, dest); err != nil {
		return err
	}
	url := "/uploads/payment-methods/" + name
	cfg := method.ConfigJSON
	if cfg == nil {
		cfg = model.JSONMap{}
	}
	cfg["qr_image_url"] = url
	if err := h.db.WithContext(c).Model(&model.PaymentMethod{}).Where("id = ?", id).Updates(map[string]any{"config_json": cfg, "updated_at": time.Now().UTC()}).Error; err != nil {
		return err
	}
	_ = h.audit(c, "payment_method.upload_qr", "payment_method", id.String(), model.JSONMap{"url": url})
	return respondOK(c, fiber.Map{"url": url})
}

func (h *AdminHandler) ListDonations(c fiber.Ctx) error {
	query, err := h.filteredDonations(c)
	if err != nil {
		return err
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return err
	}
	pagination := util.ParsePagination(c.Query("page"), c.Query("page_size"), 20, 100)
	var donations []model.Donation
	if err := query.Preload("Tier").Preload("PaymentMethod").Order("created_at DESC").Limit(pagination.PageSize).Offset(pagination.Offset).Find(&donations).Error; err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"items": donations, "total": total, "page": pagination.Page, "page_size": pagination.PageSize})
}

func (h *AdminHandler) GetDonation(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("捐赠记录 ID 无效", nil)
	}
	var donation model.Donation
	if err := h.db.WithContext(c).Preload("Tier").Preload("PaymentMethod").First(&donation, "id = ?", id).Error; err != nil {
		return notFoundOrErr(err, "捐赠记录不存在")
	}
	return respondOK(c, donation)
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *AdminHandler) UpdateDonationStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service.Validation("捐赠记录 ID 无效", nil)
	}
	var req statusRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	now := time.Now().UTC()
	var donation model.Donation
	if err := h.db.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&donation, "id = ?", id).Error; err != nil {
			return notFoundOrErr(err, "捐赠记录不存在")
		}
		if err := service.ApplyAdminStatusTransition(&donation, strings.TrimSpace(req.Status), now); err != nil {
			return err
		}
		updates := map[string]any{"status": donation.Status, "updated_at": now}
		if donation.PaidAt != nil {
			updates["paid_at"] = donation.PaidAt
		}
		return tx.Model(&model.Donation{}).Where("id = ?", id).Updates(updates).Error
	}); err != nil {
		return err
	}
	_ = h.audit(c, "donation.status", "donation", id.String(), model.JSONMap{"status": donation.Status})
	return respondOK(c, fiber.Map{"id": id, "status": donation.Status, "paid_at": donation.PaidAt})
}

func (h *AdminHandler) ExportDonations(c fiber.Ctx) error {
	if format := c.Query("format", "csv"); format != "csv" {
		return service.Validation("仅支持 csv 导出", nil)
	}
	start, end, err := parseDateBounds(c.Query("start"), c.Query("end"))
	if err != nil {
		return err
	}
	csvBytes, err := h.export.CSV(c, service.DonationExportFilter{Start: start, End: end, Status: c.Query("status"), TimeField: c.Query("time_field")})
	if err != nil {
		return err
	}
	filename := service.ExportFilename(c.Query("start"), c.Query("end"))
	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", filename))
	return c.Send(csvBytes)
}

func (h *AdminHandler) Overview(c fiber.Ctx) error {
	dr, err := service.ParseDateRange(c.Query("start"), c.Query("end"), 30)
	if err != nil {
		return err
	}
	overview, err := h.overview.Overview(c, dr)
	if err != nil {
		return err
	}
	return respondOK(c, overview)
}

func (h *AdminHandler) GetSettings(c fiber.Ctx) error {
	site, err := h.admin.Settings().Get(c, "site")
	if err != nil {
		return err
	}
	admin, err := h.admin.Settings().Get(c, "admin")
	if err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"site": site, "admin": admin})
}

func (h *AdminHandler) UpdateSiteSettings(c fiber.Ctx) error {
	var req model.JSONMap
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	site, err := h.admin.Settings().Get(c, "site")
	if err != nil {
		return err
	}
	for key, value := range req {
		switch key {
		case "name", "hero_title", "hero_subtitle", "currency", "timezone":
			text, ok := value.(string)
			if !ok {
				return service.Validation(key+" 必须是字符串", nil)
			}
			if key == "currency" {
				text = normalizeCurrency(text)
				if !util.ValidCurrency(text) {
					return service.Validation("币种无效", nil)
				}
			}
			site[key] = strings.TrimSpace(text)
		case "goal_cents":
			amount, ok := numberToInt64(value)
			if !ok || amount < 0 {
				return service.Validation("goal_cents 必须是非负整数", nil)
			}
			site[key] = amount
		case "show_goal":
			show, ok := value.(bool)
			if !ok {
				return service.Validation("show_goal 必须是布尔值", nil)
			}
			site[key] = show
		default:
			return service.Validation("不支持的设置项", map[string]string{"key": key})
		}
	}
	if err := h.admin.Settings().Upsert(c, "site", site); err != nil {
		return err
	}
	_ = h.audit(c, "settings.site", "app_settings", "site", model.JSONMap{})
	return respondOK(c, site)
}

type adminPathRequest struct {
	BasePath string `json:"base_path"`
}

func (h *AdminHandler) UpdateAdminPath(c fiber.Ctx) error {
	var req adminPathRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	basePath := strings.TrimSpace(req.BasePath)
	if err := service.ValidateAdminBasePath(basePath); err != nil {
		return err
	}
	value := service.AdminSettingValue(basePath)
	if err := h.admin.Settings().Upsert(c, "admin", value); err != nil {
		return err
	}
	_ = h.audit(c, "settings.admin_path", "app_settings", "admin", model.JSONMap{"base_path": basePath})
	return respondOK(c, fiber.Map{"base_path": basePath, "message": "后台入口已更新，请使用新地址访问。"})
}

func (h *AdminHandler) filteredDonations(c fiber.Ctx) (*gorm.DB, error) {
	query := h.db.WithContext(c).Model(&model.Donation{})
	timeField := c.Query("time_field", "created_at")
	if timeField != "created_at" && timeField != "paid_at" {
		return nil, service.Validation("time_field 无效", nil)
	}
	start, end, err := parseDateBounds(c.Query("start"), c.Query("end"))
	if err != nil {
		return nil, err
	}
	if start != nil {
		query = query.Where(timeField+" >= ?", *start)
	}
	if end != nil {
		query = query.Where(timeField+" < ?", *end)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("order_no ILIKE ? OR nickname ILIKE ? OR email ILIKE ?", like, like, like)
	}
	return query, nil
}

func validateTier(tier model.DonationTier) error {
	if tier.Name == "" || len([]rune(tier.Name)) > 80 {
		return service.Validation("档位名称不能为空且不能超过 80 字", nil)
	}
	if tier.AmountCents <= 0 {
		return service.Validation("档位金额必须大于 0", nil)
	}
	if !util.ValidCurrency(tier.Currency) {
		return service.Validation("币种无效", nil)
	}
	if len([]rune(tier.Description)) > 300 {
		return service.Validation("档位描述不能超过 300 字", nil)
	}
	return nil
}

func validatePaymentMethod(method model.PaymentMethod) error {
	if method.Code == "" || len(method.Code) > 50 {
		return service.Validation("支付方式 code 不能为空且不能超过 50 字", nil)
	}
	if method.Name == "" || len([]rune(method.Name)) > 80 {
		return service.Validation("支付方式名称不能为空且不能超过 80 字", nil)
	}
	if !payment.SupportedPaymentTypes()[method.Type] {
		return service.Validation("支付类型无效", nil)
	}
	if method.MinAmountCents != nil && *method.MinAmountCents <= 0 {
		return service.Validation("最小金额必须大于 0", nil)
	}
	if method.MaxAmountCents != nil && *method.MaxAmountCents <= 0 {
		return service.Validation("最大金额必须大于 0", nil)
	}
	if method.MinAmountCents != nil && method.MaxAmountCents != nil && *method.MinAmountCents > *method.MaxAmountCents {
		return service.Validation("最小金额不能大于最大金额", nil)
	}
	return nil
}

func validateUploadedImage(file *multipart.FileHeader) (string, error) {
	if file.Size <= 0 || file.Size > 2*1024*1024 {
		return "", service.Validation("二维码图片大小必须在 2MB 以内", nil)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mime := strings.ToLower(file.Header.Get("Content-Type"))
	allowedMIME := map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}
	if !allowedMIME[mime] {
		return "", service.Validation("二维码图片 MIME 类型无效", nil)
	}
	fh, err := file.Open()
	if err != nil {
		return "", err
	}
	defer fh.Close()
	header := make([]byte, 12)
	n, err := io.ReadFull(fh, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	header = header[:n]
	switch {
	case ext == ".png" && mime == "image/png" && hasPrefix(header, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}):
		return ".png", nil
	case (ext == ".jpg" || ext == ".jpeg") && mime == "image/jpeg" && hasPrefix(header, []byte{0xff, 0xd8, 0xff}):
		return ".jpg", nil
	case ext == ".webp" && mime == "image/webp" && len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP":
		return ".webp", nil
	default:
		return "", service.Validation("二维码图片签名无效", nil)
	}
}

func hasPrefix(value, prefix []byte) bool {
	if len(value) < len(prefix) {
		return false
	}
	for i := range prefix {
		if value[i] != prefix[i] {
			return false
		}
	}
	return true
}

func parseDateBounds(startValue, endValue string) (*time.Time, *time.Time, error) {
	var start *time.Time
	var end *time.Time
	if startValue != "" {
		parsed, err := time.ParseInLocation("2006-01-02", startValue, time.UTC)
		if err != nil {
			return nil, nil, service.Validation("开始日期无效", nil)
		}
		start = &parsed
	}
	if endValue != "" {
		parsed, err := time.ParseInLocation("2006-01-02", endValue, time.UTC)
		if err != nil {
			return nil, nil, service.Validation("结束日期无效", nil)
		}
		parsed = parsed.AddDate(0, 0, 1)
		end = &parsed
	}
	if start != nil && end != nil && !start.Before(*end) {
		return nil, nil, service.Validation("开始日期必须早于结束日期", nil)
	}
	return start, end, nil
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "CNY"
	}
	return currency
}

func notFoundOrErr(err error, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service.NotFound(message)
	}
	return err
}

func (h *AdminHandler) audit(c fiber.Ctx, action, targetType, targetID string, metadata model.JSONMap) error {
	admin, ok := c.Locals(apimw.LocalAdmin).(*model.Admin)
	if !ok {
		return nil
	}
	return h.admin.Audit(c, admin.ID, action, targetType, targetID, metadata, c.IP())
}

func numberToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		if v < 0 || v != float64(int64(v)) {
			return 0, false
		}
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}
