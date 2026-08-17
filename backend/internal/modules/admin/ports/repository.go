package ports

import (
	"context"
	"errors"

	admin "rigmark/internal/modules/admin/domain"
	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var (
	ErrNotFound = errors.New("admin entity not found")
	ErrConflict = errors.New("admin entity conflicts with an existing record")
)

type Repository interface {
	Dashboard(context.Context) (admin.Dashboard, error)
	References(context.Context) (admin.References, error)
	ListProducts(context.Context, string, int, int) (admin.ProductPage, error)
	GetProduct(context.Context, catalog.ProductID) (catalog.Product, error)
	CreateProduct(context.Context, identity.UserID, admin.ProductInput) (catalog.Product, error)
	UpdateProduct(context.Context, identity.UserID, catalog.ProductID, admin.ProductInput) (catalog.Product, error)
	SetProductStatus(context.Context, identity.UserID, catalog.ProductID, catalog.ProductStatus) error
	AddImage(context.Context, identity.UserID, catalog.ProductID, admin.ImageInput) (catalog.ProductImage, error)
	DeleteImage(context.Context, identity.UserID, catalog.ProductID, string) (string, error)
	UpsertAttribute(context.Context, identity.UserID, catalog.ProductID, admin.AttributeInput) (catalog.Attribute, error)
	DeleteAttribute(context.Context, identity.UserID, catalog.ProductID, string) error
	ListCategories(context.Context) ([]admin.Category, error)
	ListBrands(context.Context) ([]admin.Brand, error)
	ListMerchants(context.Context) ([]admin.Merchant, error)
	ListOffers(context.Context, int, int) (admin.Page[admin.Offer], error)
	CreateOffer(context.Context, identity.UserID, admin.OfferInput) (admin.Offer, error)
	UpdateOffer(context.Context, identity.UserID, string, admin.OfferInput) (admin.Offer, error)
	ListAffiliateLinks(context.Context, int, int) (admin.Page[admin.AffiliateLink], error)
	UpdateAffiliateLink(context.Context, identity.UserID, string, admin.AffiliateLinkInput) (admin.AffiliateLink, error)
	ListRecommendations(context.Context, int, int) (admin.Page[admin.Recommendation], error)
	GetRecommendation(context.Context, string) (admin.RecommendationDetail, error)
	ListUsers(context.Context, int, int) (admin.Page[admin.User], error)
	ListEvents(context.Context, string, int, int) (admin.Page[admin.Event], error)
}
