package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	admin "rigmark/internal/modules/admin/domain"
	"rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
	commerce "rigmark/internal/modules/commerce/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var ErrInvalidInput = errors.New("admin input is invalid")

const productMediaPathPrefix = "/api/media/products/"

var (
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	providerPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)
)

type Service struct {
	repository ports.Repository
	images     ports.ImageStorage
	scanner    ports.ImageScanner
}

func NewService(repository ports.Repository, images ports.ImageStorage) *Service {
	return &Service{repository: repository, images: images}
}

func NewServiceWithMedia(repository ports.Repository, images ports.ImageStorage, scanner ports.ImageScanner) *Service {
	return &Service{repository: repository, images: images, scanner: scanner}
}

func (service *Service) Dashboard(ctx context.Context) (admin.Dashboard, error) {
	return service.repository.Dashboard(ctx)
}

func (service *Service) References(ctx context.Context) (admin.References, error) {
	return service.repository.References(ctx)
}

func (service *Service) ListProducts(ctx context.Context, search string, page, pageSize int) (admin.ProductPage, error) {
	offset, limit, err := pagination(page, pageSize)
	if err != nil || len(strings.TrimSpace(search)) > 100 {
		return admin.ProductPage{}, ErrInvalidInput
	}
	return service.repository.ListProducts(ctx, strings.TrimSpace(search), limit, offset)
}

func (service *Service) GetProduct(ctx context.Context, id catalog.ProductID) (catalog.Product, error) {
	return service.repository.GetProduct(ctx, id)
}

func (service *Service) CreateProduct(ctx context.Context, actor identity.UserID, input admin.ProductInput) (catalog.Product, error) {
	input = normalizeProduct(input)
	if err := validateProductInput(input); err != nil {
		return catalog.Product{}, invalidInput(err)
	}
	return service.repository.CreateProduct(ctx, actor, input)
}

func (service *Service) UpdateProduct(ctx context.Context, actor identity.UserID, id catalog.ProductID, input admin.ProductInput) (catalog.Product, error) {
	input = normalizeProduct(input)
	if id == "" {
		return catalog.Product{}, ErrInvalidInput
	}
	if err := validateProductInput(input); err != nil {
		return catalog.Product{}, invalidInput(err)
	}
	return service.repository.UpdateProduct(ctx, actor, id, input)
}

// invalidInput keeps the sentinel every caller matches on and, when the domain
// named the field that failed, keeps that too so the response can point at it
// rather than at the whole form.
func invalidInput(err error) error {
	var field catalog.FieldError
	if errors.As(err, &field) {
		return fmt.Errorf("%w: %w", ErrInvalidInput, field)
	}
	return ErrInvalidInput
}

func (service *Service) SetProductStatus(ctx context.Context, actor identity.UserID, id catalog.ProductID, status catalog.ProductStatus) error {
	if id == "" || status == catalog.ProductStatusPublished ||
		(status != catalog.ProductStatusDraft && status != catalog.ProductStatusDiscontinued) {
		return ErrInvalidInput
	}
	return service.repository.SetProductStatus(ctx, actor, id, status)
}

func (service *Service) AddImage(ctx context.Context, actor identity.UserID, productID catalog.ProductID, input admin.ImageInput) (catalog.ProductImage, error) {
	input.URL = strings.TrimSpace(input.URL)
	input.AltText = strings.TrimSpace(input.AltText)
	if productID == "" || input.SortOrder < 0 || len(input.AltText) < 1 || len(input.AltText) > 240 || !validImageURL(input.URL) {
		return catalog.ProductImage{}, ErrInvalidInput
	}
	if strings.HasPrefix(input.URL, productMediaPathPrefix) &&
		(service.images == nil || !service.images.BelongsTo(productID, strings.TrimPrefix(input.URL, productMediaPathPrefix))) {
		return catalog.ProductImage{}, ErrInvalidInput
	}
	return service.repository.AddImage(ctx, actor, productID, input)
}

func (service *Service) UploadImage(ctx context.Context, actor identity.UserID, productID catalog.ProductID, upload admin.ImageUpload) (catalog.ProductImage, error) {
	if service.images == nil || productID == "" || len(upload.Data) == 0 || len(upload.Data) > 5*1024*1024 || len(strings.TrimSpace(upload.AltText)) < 1 || upload.SortOrder < 0 {
		return catalog.ProductImage{}, ErrInvalidInput
	}
	extension := ""
	switch upload.MIMEType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		return catalog.ProductImage{}, ErrInvalidInput
	}
	if service.scanner == nil {
		return catalog.ProductImage{}, ports.ErrMediaScanUnavailable
	}
	if err := service.scanner.Scan(ctx, upload.Data, upload.MIMEType); err != nil {
		return catalog.ProductImage{}, err
	}
	path, created, err := service.images.Save(ctx, productID, upload.Data, extension)
	if err != nil {
		return catalog.ProductImage{}, err
	}
	image, err := service.repository.AddImage(ctx, actor, productID, admin.ImageInput{URL: productMediaPathPrefix + path, AltText: strings.TrimSpace(upload.AltText), SortOrder: upload.SortOrder, IsPrimary: upload.IsPrimary})
	if err != nil {
		if created {
			if deleteErr := service.images.Delete(ctx, path); deleteErr != nil {
				_ = service.repository.ScheduleMediaDeletion(ctx, productID, path)
			}
		}
		return catalog.ProductImage{}, err
	}
	return image, nil
}

func (service *Service) OpenImage(ctx context.Context, name string) ([]byte, string, error) {
	if service.images == nil {
		return nil, "", ports.ErrNotFound
	}
	return service.images.Open(ctx, name)
}

func (service *Service) DeleteImage(ctx context.Context, actor identity.UserID, productID catalog.ProductID, imageID string) error {
	if productID == "" || strings.TrimSpace(imageID) == "" {
		return ErrInvalidInput
	}
	imageURL, err := service.repository.DeleteImage(ctx, actor, productID, imageID)
	if err != nil {
		return err
	}
	if imageURL == "" {
		return nil
	}
	if service.images != nil && strings.HasPrefix(imageURL, productMediaPathPrefix) {
		name := strings.TrimPrefix(imageURL, productMediaPathPrefix)
		if !service.images.BelongsTo(productID, name) {
			return ErrInvalidInput
		}
		if err := service.images.Delete(ctx, name); err != nil {
			_ = service.repository.FailMediaDeletion(ctx, name, "storage.delete_failed", time.Now().UTC())
			return err
		}
		if err := service.repository.CompleteMediaDeletion(ctx, name, time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) UpsertAttribute(ctx context.Context, actor identity.UserID, productID catalog.ProductID, input admin.AttributeInput) (catalog.Attribute, error) {
	input.Key = strings.TrimSpace(input.Key)
	if input.TextValue != nil {
		value := strings.TrimSpace(*input.TextValue)
		input.TextValue = &value
	}
	attribute := catalog.Attribute{Key: input.Key, Type: input.Type, NumericValue: input.NumericValue, TextValue: input.TextValue, BooleanValue: input.BooleanValue, Unit: input.Unit, IsFilterable: input.IsFilterable}
	if productID == "" || attribute.Validate() != nil {
		return catalog.Attribute{}, ErrInvalidInput
	}
	return service.repository.UpsertAttribute(ctx, actor, productID, input)
}

func (service *Service) DeleteAttribute(ctx context.Context, actor identity.UserID, productID catalog.ProductID, key string) error {
	if productID == "" || !regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(key) {
		return ErrInvalidInput
	}
	return service.repository.DeleteAttribute(ctx, actor, productID, key)
}

func (service *Service) ListCategories(ctx context.Context) ([]admin.Category, error) {
	return service.repository.ListCategories(ctx)
}

func (service *Service) ListBrands(ctx context.Context) ([]admin.Brand, error) {
	return service.repository.ListBrands(ctx)
}

func (service *Service) ListMerchants(ctx context.Context) ([]admin.Merchant, error) {
	return service.repository.ListMerchants(ctx)
}

func (service *Service) ListOffers(ctx context.Context, page, pageSize int) (admin.Page[admin.Offer], error) {
	offset, limit, err := pagination(page, pageSize)
	if err != nil {
		return admin.Page[admin.Offer]{}, err
	}
	return service.repository.ListOffers(ctx, limit, offset)
}

func (service *Service) CreateOffer(ctx context.Context, actor identity.UserID, input admin.OfferInput) (admin.Offer, error) {
	input = normalizeOffer(input)
	if validateOfferInput(input) != nil {
		return admin.Offer{}, ErrInvalidInput
	}
	return service.repository.CreateOffer(ctx, actor, input)
}

func (service *Service) UpdateOffer(ctx context.Context, actor identity.UserID, id string, input admin.OfferInput) (admin.Offer, error) {
	input = normalizeOffer(input)
	if strings.TrimSpace(id) == "" || validateOfferInput(input) != nil {
		return admin.Offer{}, ErrInvalidInput
	}
	return service.repository.UpdateOffer(ctx, actor, id, input)
}

func (service *Service) ListAffiliateLinks(ctx context.Context, page, pageSize int) (admin.Page[admin.AffiliateLink], error) {
	offset, limit, err := pagination(page, pageSize)
	if err != nil {
		return admin.Page[admin.AffiliateLink]{}, err
	}
	return service.repository.ListAffiliateLinks(ctx, limit, offset)
}

func (service *Service) UpdateAffiliateLink(ctx context.Context, actor identity.UserID, id string, input admin.AffiliateLinkInput) (admin.AffiliateLink, error) {
	input = normalizeAffiliate(input)
	if strings.TrimSpace(id) == "" || validateAffiliate(input) != nil {
		return admin.AffiliateLink{}, ErrInvalidInput
	}
	return service.repository.UpdateAffiliateLink(ctx, actor, id, input)
}

func (service *Service) ListRecommendations(ctx context.Context, page, pageSize int) (admin.Page[admin.Recommendation], error) {
	offset, limit, err := pagination(page, pageSize)
	if err != nil {
		return admin.Page[admin.Recommendation]{}, err
	}
	return service.repository.ListRecommendations(ctx, limit, offset)
}

func (service *Service) GetRecommendation(ctx context.Context, id string) (admin.RecommendationDetail, error) {
	if strings.TrimSpace(id) == "" {
		return admin.RecommendationDetail{}, ErrInvalidInput
	}
	return service.repository.GetRecommendation(ctx, id)
}

func (service *Service) ListUsers(ctx context.Context, page, pageSize int) (admin.Page[admin.User], error) {
	offset, limit, err := pagination(page, pageSize)
	if err != nil {
		return admin.Page[admin.User]{}, err
	}
	return service.repository.ListUsers(ctx, limit, offset)
}

func (service *Service) ListEvents(ctx context.Context, name string, page, pageSize int) (admin.Page[admin.Event], error) {
	offset, limit, err := pagination(page, pageSize)
	name = strings.TrimSpace(name)
	if err != nil || len(name) > 100 {
		return admin.Page[admin.Event]{}, ErrInvalidInput
	}
	return service.repository.ListEvents(ctx, name, limit, offset)
}

func pagination(page, pageSize int) (int, int, error) {
	if page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		return 0, 0, ErrInvalidInput
	}
	return (page - 1) * pageSize, pageSize, nil
}

func normalizeProduct(input admin.ProductInput) admin.ProductInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Description = strings.TrimSpace(input.Description)
	input.Price.Currency = strings.ToUpper(strings.TrimSpace(input.Price.Currency))
	input.Billing = input.Billing.Normalized()
	input.Material = strings.TrimSpace(input.Material)
	return input
}

func validateProductInput(input admin.ProductInput) error {
	product := catalog.Product{ID: "validation", CategoryID: input.CategoryID, BrandID: input.BrandID, Name: input.Name, Slug: input.Slug, Description: input.Description, Price: input.Price, Billing: input.Billing, Dimensions: input.Dimensions, WeightGrams: input.WeightGrams, MaxCapacityGrams: input.MaxCapacityGrams, Material: input.Material, WarrantyMonths: input.WarrantyMonths, Scores: input.Scores, Status: catalog.ProductStatusDraft}
	if len(input.Name) > 180 || !slugPattern.MatchString(input.Slug) {
		return ErrInvalidInput
	}
	return product.Validate()
}

func normalizeOffer(input admin.OfferInput) admin.OfferInput {
	input.MerchantSKU = strings.TrimSpace(input.MerchantSKU)
	input.ProductURL = strings.TrimSpace(input.ProductURL)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Availability = strings.TrimSpace(input.Availability)
	input.Condition = strings.TrimSpace(input.Condition)
	if input.Affiliate != nil {
		normalized := normalizeAffiliate(*input.Affiliate)
		input.Affiliate = &normalized
	}
	return input
}

func validateOfferInput(input admin.OfferInput) error {
	if input.MerchantID == "" || input.ProductID == "" || len(input.MerchantSKU) < 1 || len(input.MerchantSKU) > 120 || input.PriceMinor < 0 || input.ShippingMinor < 0 || len(input.Currency) != 3 || !validMerchantAdminURL(input.ProductURL) {
		return ErrInvalidInput
	}
	if input.Availability != "in_stock" && input.Availability != "backorder" && input.Availability != "out_of_stock" && input.Availability != "discontinued" {
		return ErrInvalidInput
	}
	if input.Condition != "new" && input.Condition != "refurbished" && input.Condition != "used" {
		return ErrInvalidInput
	}
	if input.Affiliate != nil {
		return validateAffiliate(*input.Affiliate)
	}
	return nil
}

func normalizeAffiliate(input admin.AffiliateLinkInput) admin.AffiliateLinkInput {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.DestinationURL = strings.TrimSpace(input.DestinationURL)
	input.DisclosureLabel = strings.TrimSpace(input.DisclosureLabel)
	input.CommissionType = strings.TrimSpace(input.CommissionType)
	if input.ExternalReference != nil {
		value := strings.TrimSpace(*input.ExternalReference)
		input.ExternalReference = &value
	}
	if input.ProgramID != nil {
		value := strings.TrimSpace(*input.ProgramID)
		input.ProgramID = &value
	}
	if input.CommissionCurrency != nil {
		value := strings.ToUpper(strings.TrimSpace(*input.CommissionCurrency))
		input.CommissionCurrency = &value
	}
	return input
}

func validateAffiliate(input admin.AffiliateLinkInput) error {
	if !providerPattern.MatchString(input.Provider) || !validMerchantAdminURL(input.DestinationURL) || len(input.DisclosureLabel) < 1 || input.Priority < -1000 || input.Priority > 1000 {
		return ErrInvalidInput
	}
	switch input.CommissionType {
	case "unknown":
		if input.CommissionRateBPS != nil || input.CommissionAmount != nil || input.CommissionCurrency != nil {
			return ErrInvalidInput
		}
	case "percentage":
		if input.CommissionRateBPS == nil || *input.CommissionRateBPS < 0 || *input.CommissionRateBPS > 10000 || input.CommissionAmount != nil || input.CommissionCurrency != nil {
			return ErrInvalidInput
		}
	case "fixed":
		if input.CommissionAmount == nil || *input.CommissionAmount < 0 || input.CommissionCurrency == nil || len(*input.CommissionCurrency) != 3 || input.CommissionRateBPS != nil {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}
	return nil
}

func validHTTPSURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func validMerchantAdminURL(value string) bool {
	if commerce.SafeMerchantURL(value) {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil && parsed.Fragment == "" &&
		strings.HasSuffix(strings.ToLower(parsed.Hostname()), ".invalid")
}

func validImageURL(value string) bool {
	return validHTTPSURL(value) || strings.HasPrefix(value, "/images/") || strings.HasPrefix(value, "/api/media/products/")
}
