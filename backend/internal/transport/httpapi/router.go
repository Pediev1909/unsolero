package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	admindomain "rigmark/internal/modules/admin/domain"
	aidomain "rigmark/internal/modules/ai/domain"
	analyticsdomain "rigmark/internal/modules/analytics/domain"
	catalog "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	commercedomain "rigmark/internal/modules/commerce/domain"
	contentdomain "rigmark/internal/modules/content/domain"
	evidencedomain "rigmark/internal/modules/evidence/domain"
	health "rigmark/internal/modules/health/application"
	identity "rigmark/internal/modules/identity/application"
	"rigmark/internal/modules/identity/domain"
	recommendationdomain "rigmark/internal/modules/recommendation/domain"
)

type HealthService interface {
	Live() health.Report
	Check(context.Context) (health.Report, error)
}

type AuthenticationService interface {
	Register(context.Context, string, string) (identity.AuthenticatedSession, error)
	Login(context.Context, string, string) (identity.AuthenticatedSession, error)
	Logout(context.Context, string) error
	Authenticate(context.Context, string) (domain.Principal, error)
}

type AuthCookieConfig struct {
	Name          string
	Secure        bool
	MaxAge        int
	AllowedOrigin string
}

type CatalogService interface {
	Search(context.Context, catalog.Query) (catalog.Page, error)
	GetProduct(context.Context, string) (catalog.ProductDetail, error)
	ListCategories(context.Context) ([]catalogdomain.Category, error)
	GetCategory(context.Context, string) (catalogdomain.Category, error)
	ListBrands(context.Context) ([]catalogdomain.Brand, error)
	GetBrand(context.Context, string) (catalogdomain.Brand, error)
}

type CommerceService interface {
	ListOffers(context.Context, catalogdomain.ProductID, string) ([]commercedomain.Offer, error)
	TrackOfferClick(context.Context, commercedomain.AffiliateClick) (string, error)
	TrackLegacyLinkClick(context.Context, commercedomain.AffiliateClick) (string, error)
}

type AnalyticsService interface {
	RecordClientEvent(context.Context, analyticsdomain.Event) error
}

type AnalyticsReportingService interface {
	Report(context.Context) (analyticsdomain.Report, error)
}

type ContentService interface {
	List(context.Context, string, string, int) ([]contentdomain.Summary, error)
	Get(context.Context, string) (contentdomain.Entry, error)
	Sitemap(context.Context) ([]contentdomain.SitemapEntry, error)
	AbsoluteURL(string) string
}

// AIService is retained in the server composition root for future bounded
// routes or workers. No public AI endpoint is exposed in this phase.
type AIService interface {
	UnderstandUserInput(context.Context, aidomain.UnderstandUserInputRequest) (aidomain.UserInputUnderstanding, error)
	ExtractRequirements(context.Context, aidomain.ExtractRequirementsRequest) (aidomain.RequirementsDraft, error)
	AskClarifyingQuestion(context.Context, aidomain.ClarifyingQuestionRequest) (aidomain.ClarifyingQuestion, error)
	ExplainRecommendation(context.Context, aidomain.ExplainRecommendationRequest) (aidomain.ExplanationPlan, error)
	RefineRecommendation(context.Context, aidomain.RefineRecommendationRequest) (aidomain.Refinement, error)
	CompareProducts(context.Context, aidomain.CompareProductsRequest) (aidomain.ComparisonPlan, error)
}

type AdminService interface {
	Dashboard(context.Context) (admindomain.Dashboard, error)
	References(context.Context) (admindomain.References, error)
	ListProducts(context.Context, string, int, int) (admindomain.ProductPage, error)
	GetProduct(context.Context, catalogdomain.ProductID) (catalogdomain.Product, error)
	CreateProduct(context.Context, domain.UserID, admindomain.ProductInput) (catalogdomain.Product, error)
	UpdateProduct(context.Context, domain.UserID, catalogdomain.ProductID, admindomain.ProductInput) (catalogdomain.Product, error)
	SetProductStatus(context.Context, domain.UserID, catalogdomain.ProductID, catalogdomain.ProductStatus) error
	AddImage(context.Context, domain.UserID, catalogdomain.ProductID, admindomain.ImageInput) (catalogdomain.ProductImage, error)
	UploadImage(context.Context, domain.UserID, catalogdomain.ProductID, admindomain.ImageUpload) (catalogdomain.ProductImage, error)
	OpenImage(context.Context, string) ([]byte, string, error)
	DeleteImage(context.Context, domain.UserID, catalogdomain.ProductID, string) error
	UpsertAttribute(context.Context, domain.UserID, catalogdomain.ProductID, admindomain.AttributeInput) (catalogdomain.Attribute, error)
	DeleteAttribute(context.Context, domain.UserID, catalogdomain.ProductID, string) error
	ListCategories(context.Context) ([]admindomain.Category, error)
	ListBrands(context.Context) ([]admindomain.Brand, error)
	ListMerchants(context.Context) ([]admindomain.Merchant, error)
	ListOffers(context.Context, int, int) (admindomain.Page[admindomain.Offer], error)
	CreateOffer(context.Context, domain.UserID, admindomain.OfferInput) (admindomain.Offer, error)
	UpdateOffer(context.Context, domain.UserID, string, admindomain.OfferInput) (admindomain.Offer, error)
	ListAffiliateLinks(context.Context, int, int) (admindomain.Page[admindomain.AffiliateLink], error)
	UpdateAffiliateLink(context.Context, domain.UserID, string, admindomain.AffiliateLinkInput) (admindomain.AffiliateLink, error)
	ListRecommendations(context.Context, int, int) (admindomain.Page[admindomain.Recommendation], error)
	GetRecommendation(context.Context, string) (admindomain.RecommendationDetail, error)
	ListUsers(context.Context, int, int) (admindomain.Page[admindomain.User], error)
	ListEvents(context.Context, string, int, int) (admindomain.Page[admindomain.Event], error)
}

type EvidenceService interface {
	CreateSource(context.Context, domain.UserID, evidencedomain.SourceInput) (evidencedomain.Source, error)
	ReviewSource(context.Context, domain.UserID, string, evidencedomain.ReviewStatus, string) (evidencedomain.Source, error)
	CreateObservation(context.Context, domain.UserID, evidencedomain.ObservationInput) (evidencedomain.Observation, error)
	CreateRevision(context.Context, domain.UserID, evidencedomain.RevisionInput) (evidencedomain.Revision, error)
	Submit(context.Context, domain.UserID, string) (evidencedomain.Revision, error)
	Approve(context.Context, domain.UserID, string, string) (evidencedomain.Revision, error)
	Reject(context.Context, domain.UserID, string, string) (evidencedomain.Revision, error)
	Publish(context.Context, domain.UserID, string) (evidencedomain.Revision, error)
	GetProduct(context.Context, catalogdomain.ProductID) (evidencedomain.ProductGovernance, error)
	ListProducts(context.Context, int, int) ([]evidencedomain.ProductGovernance, int64, error)
}

type RecommendationPolicyService interface {
	List(context.Context) ([]recommendationdomain.PolicySummary, error)
	Transition(context.Context, domain.UserID, string, recommendationdomain.PolicyWorkflowStatus, string) error
}

type WishlistService interface {
	List(context.Context, domain.UserID) ([]catalogdomain.ProductID, error)
	Save(context.Context, domain.UserID, catalogdomain.ProductID) error
	Delete(context.Context, domain.UserID, catalogdomain.ProductID) error
}

type PublicServices struct {
	Catalog              CatalogService
	Commerce             CommerceService
	Wishlist             WishlistService
	Recommendations      recommendationService
	Comparison           ComparisonService
	Analytics            AnalyticsService
	AnalyticsReporting   AnalyticsReportingService
	Admin                AdminService
	AI                   AIService
	Content              ContentService
	Evidence             EvidenceService
	RecommendationPolicy RecommendationPolicyService
	RateLimits           RateLimitConfig
}

type Handler struct {
	health               HealthService
	auth                 AuthenticationService
	catalog              CatalogService
	commerce             CommerceService
	wishlist             WishlistService
	recommendations      recommendationService
	comparison           ComparisonService
	analytics            AnalyticsService
	analyticsReporting   AnalyticsReportingService
	admin                AdminService
	ai                   AIService
	content              ContentService
	evidence             EvidenceService
	recommendationPolicy RecommendationPolicyService
	cookie               AuthCookieConfig
	logger               *slog.Logger
}

func NewRouter(
	healthService HealthService,
	authService AuthenticationService,
	cookie AuthCookieConfig,
	logger *slog.Logger,
	publicServices ...PublicServices,
) http.Handler {
	handler := &Handler{health: healthService, auth: authService, cookie: cookie, logger: logger}
	if len(publicServices) > 0 {
		handler.catalog = publicServices[0].Catalog
		handler.commerce = publicServices[0].Commerce
		handler.wishlist = publicServices[0].Wishlist
		handler.recommendations = publicServices[0].Recommendations
		handler.comparison = publicServices[0].Comparison
		handler.analytics = publicServices[0].Analytics
		handler.analyticsReporting = publicServices[0].AnalyticsReporting
		handler.admin = publicServices[0].Admin
		handler.ai = publicServices[0].AI
		handler.content = publicServices[0].Content
		handler.evidence = publicServices[0].Evidence
		handler.recommendationPolicy = publicServices[0].RecommendationPolicy
	}
	mux := http.NewServeMux()

	// Public foundation contract.
	mux.HandleFunc("GET /api/health", handler.healthCheck)

	// Infrastructure probes remain separate so orchestration can distinguish
	// process liveness from dependency readiness.
	mux.HandleFunc("GET /api/v1/health/live", handler.live)
	mux.HandleFunc("GET /api/v1/health/ready", handler.healthCheck)

	mux.HandleFunc("POST /api/auth/register", handler.register)
	mux.HandleFunc("POST /api/auth/login", handler.login)
	mux.HandleFunc("POST /api/auth/logout", handler.logout)
	mux.Handle("GET /api/auth/me", handler.requireAuthentication(http.HandlerFunc(handler.me)))

	if handler.catalog != nil {
		mux.HandleFunc("GET /api/catalog/products", handler.listProducts)
		mux.HandleFunc("GET /api/catalog/products/{slug}", handler.getProduct)
		mux.HandleFunc("GET /api/catalog/categories", handler.listCategories)
		mux.HandleFunc("GET /api/catalog/categories/{slug}", handler.getCategory)
		mux.HandleFunc("GET /api/catalog/brands", handler.listBrands)
		mux.HandleFunc("GET /api/catalog/brands/{slug}", handler.getBrand)
	}
	if handler.content != nil {
		mux.HandleFunc("GET /api/content", handler.listContent)
		mux.HandleFunc("GET /api/content/{slug}", handler.getContent)
		mux.HandleFunc("GET /sitemap.xml", handler.sitemap)
		mux.HandleFunc("GET /robots.txt", handler.robots)
	}
	if handler.catalog != nil && handler.commerce != nil {
		mux.HandleFunc("GET /api/catalog/products/{slug}/offers", handler.listOffers)
		mux.Handle("GET /api/affiliate/click/{offerID}", handler.attachAuthentication(http.HandlerFunc(handler.affiliateClickRedirect)))
		mux.Handle("GET /api/out/{affiliateLinkID}", handler.attachAuthentication(http.HandlerFunc(handler.outboundRedirect)))
	}
	if handler.analytics != nil {
		mux.Handle("POST /api/analytics/events", handler.attachAuthentication(http.HandlerFunc(handler.recordAnalyticsEvent)))
	}
	if handler.admin != nil {
		mux.HandleFunc("GET /api/media/products/{file}", handler.productImage)
		adminOnly := func(next http.HandlerFunc) http.Handler {
			return handler.requireRole(domain.RoleAdmin, next)
		}
		mux.Handle("GET /api/admin/dashboard", adminOnly(handler.adminDashboard))
		if handler.analyticsReporting != nil {
			mux.Handle("GET /api/admin/analytics", adminOnly(handler.adminAnalyticsReport))
		}
		mux.Handle("GET /api/admin/references", adminOnly(handler.adminReferences))
		mux.Handle("GET /api/admin/products", adminOnly(handler.adminListProducts))
		mux.Handle("POST /api/admin/products", adminOnly(handler.adminCreateProduct))
		mux.Handle("GET /api/admin/products/{productID}", adminOnly(handler.adminGetProduct))
		mux.Handle("PATCH /api/admin/products/{productID}", adminOnly(handler.adminUpdateProduct))
		mux.Handle("PUT /api/admin/products/{productID}/status", adminOnly(handler.adminProductStatus))
		mux.Handle("POST /api/admin/products/{productID}/images", adminOnly(handler.adminAddImage))
		mux.Handle("DELETE /api/admin/products/{productID}/images/{imageID}", adminOnly(handler.adminDeleteImage))
		mux.Handle("PUT /api/admin/products/{productID}/attributes/{key}", adminOnly(handler.adminUpsertAttribute))
		mux.Handle("DELETE /api/admin/products/{productID}/attributes/{key}", adminOnly(handler.adminDeleteAttribute))
		mux.Handle("GET /api/admin/categories", adminOnly(handler.adminListCategories))
		mux.Handle("GET /api/admin/brands", adminOnly(handler.adminListBrands))
		mux.Handle("GET /api/admin/merchants", adminOnly(handler.adminListMerchants))
		mux.Handle("GET /api/admin/offers", adminOnly(handler.adminListOffers))
		mux.Handle("POST /api/admin/offers", adminOnly(handler.adminCreateOffer))
		mux.Handle("PATCH /api/admin/offers/{offerID}", adminOnly(handler.adminUpdateOffer))
		mux.Handle("GET /api/admin/affiliate-links", adminOnly(handler.adminListAffiliateLinks))
		mux.Handle("PATCH /api/admin/affiliate-links/{linkID}", adminOnly(handler.adminUpdateAffiliateLink))
		mux.Handle("GET /api/admin/recommendations", adminOnly(handler.adminListRecommendations))
		mux.Handle("GET /api/admin/recommendations/{recommendationID}", adminOnly(handler.adminGetRecommendation))
		mux.Handle("GET /api/admin/users", adminOnly(handler.adminListUsers))
		mux.Handle("GET /api/admin/events", adminOnly(handler.adminListEvents))
		if handler.evidence != nil {
			evidenceEditor := func(next http.HandlerFunc) http.Handler {
				return handler.requireRole(domain.RoleEvidenceEditor, next)
			}
			evidenceReviewer := func(next http.HandlerFunc) http.Handler {
				return handler.requireRole(domain.RoleEvidenceReviewer, next)
			}
			mux.Handle("GET /api/admin/evidence/products", adminOnly(handler.adminListProductGovernance))
			mux.Handle("GET /api/admin/evidence/products/{productID}", adminOnly(handler.adminGetProductGovernance))
			mux.Handle("POST /api/admin/evidence/sources", evidenceEditor(handler.adminCreateEvidenceSource))
			mux.Handle("PUT /api/admin/evidence/sources/{sourceID}/review", evidenceReviewer(handler.adminReviewEvidenceSource))
			mux.Handle("POST /api/admin/evidence/observations", evidenceEditor(handler.adminCreateEvidenceObservation))
			mux.Handle("POST /api/admin/evidence/products/{productID}/revisions", evidenceEditor(handler.adminCreateEvidenceRevision))
			mux.Handle("POST /api/admin/evidence/revisions/{revisionID}/submit", evidenceEditor(handler.adminSubmitEvidenceRevision))
			mux.Handle("POST /api/admin/evidence/revisions/{revisionID}/approve", evidenceReviewer(handler.adminApproveEvidenceRevision))
			mux.Handle("POST /api/admin/evidence/revisions/{revisionID}/reject", evidenceReviewer(handler.adminRejectEvidenceRevision))
			mux.Handle("POST /api/admin/evidence/revisions/{revisionID}/publish", evidenceReviewer(handler.adminPublishEvidenceRevision))
		}
		if handler.recommendationPolicy != nil {
			policyEditor := func(next http.HandlerFunc) http.Handler {
				return handler.requireRole(domain.RolePolicyEditor, next)
			}
			policyReviewer := func(next http.HandlerFunc) http.Handler {
				return handler.requireRole(domain.RolePolicyReviewer, next)
			}
			mux.Handle("GET /api/admin/recommendation-policies", adminOnly(handler.adminListRecommendationPolicies))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/submit", policyEditor(handler.adminSubmitRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/approve", policyReviewer(handler.adminApproveRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/reject", policyReviewer(handler.adminRejectRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/activate", policyReviewer(handler.adminActivateRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/deactivate", policyReviewer(handler.adminDeactivateRecommendationPolicy))
		}
	}
	if handler.wishlist != nil {
		mux.Handle("GET /api/account/wishlist", handler.requireAuthentication(http.HandlerFunc(handler.listWishlist)))
		mux.Handle("PUT /api/account/wishlist/{productID}", handler.requireAuthentication(http.HandlerFunc(handler.saveWishlist)))
		mux.Handle("DELETE /api/account/wishlist/{productID}", handler.requireAuthentication(http.HandlerFunc(handler.deleteWishlist)))
	}
	if handler.recommendations != nil {
		mux.Handle("POST /api/recommendations/generate", handler.attachAuthentication(http.HandlerFunc(handler.generateRecommendation)))
		mux.Handle("GET /api/recommendations/draft", handler.requireAuthentication(http.HandlerFunc(handler.getRecommendationDraft)))
		mux.Handle("PUT /api/recommendations/draft", handler.requireAuthentication(http.HandlerFunc(handler.saveRecommendationDraft)))
		mux.Handle("DELETE /api/recommendations/draft", handler.requireAuthentication(http.HandlerFunc(handler.deleteRecommendationDraft)))
		mux.Handle("GET /api/account/setups", handler.requireAuthentication(http.HandlerFunc(handler.listSetups)))
		mux.Handle("GET /api/account/setups/{setupID}", handler.requireAuthentication(http.HandlerFunc(handler.getSetup)))
		mux.Handle("PATCH /api/account/setups/{setupID}", handler.requireAuthentication(http.HandlerFunc(handler.renameSetup)))
		mux.Handle("DELETE /api/account/setups/{setupID}", handler.requireAuthentication(http.HandlerFunc(handler.deleteSetup)))
	}
	if handler.comparison != nil {
		mux.Handle("GET /api/account/comparison", handler.requireAuthentication(http.HandlerFunc(handler.listComparison)))
		mux.Handle("PUT /api/account/comparison", handler.requireAuthentication(http.HandlerFunc(handler.replaceComparison)))
	}

	return requestObservability(
		recoverPanics(
			securityHeaders(
				sameOriginProtection(
					rateLimitRequests(apiCacheDefaults(mux), publicRateLimits(publicServices), logger),
					logger,
					cookie.AllowedOrigin,
				),
			),
			logger,
		),
		logger,
	)
}

func publicRateLimits(services []PublicServices) RateLimitConfig {
	if len(services) == 0 {
		return DefaultRateLimitConfig()
	}
	return services[0].RateLimits.withDefaults()
}

func (h *Handler) live(response http.ResponseWriter, _ *http.Request) {
	writeReport(response, http.StatusOK, h.health.Live(), h.logger)
}

func (h *Handler) healthCheck(response http.ResponseWriter, request *http.Request) {
	report, err := h.health.Check(request.Context())
	if err != nil {
		h.logger.Warn("health check failed", "error", err)
		writeReport(response, http.StatusServiceUnavailable, report, h.logger)
		return
	}
	writeReport(response, http.StatusOK, report, h.logger)
}

func writeReport(response http.ResponseWriter, status int, report health.Report, logger *slog.Logger) {
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, status, report, logger)
}
