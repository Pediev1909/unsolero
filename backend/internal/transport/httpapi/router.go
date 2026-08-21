package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	admindomain "rigmark/internal/modules/admin/domain"
	aidomain "rigmark/internal/modules/ai/domain"
	analyticsdomain "rigmark/internal/modules/analytics/domain"
	catalog "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	commerce "rigmark/internal/modules/commerce/application"
	commercedomain "rigmark/internal/modules/commerce/domain"
	contentdomain "rigmark/internal/modules/content/domain"
	evidencedomain "rigmark/internal/modules/evidence/domain"
	health "rigmark/internal/modules/health/application"
	identity "rigmark/internal/modules/identity/application"
	"rigmark/internal/modules/identity/domain"
	identityports "rigmark/internal/modules/identity/ports"
	planningports "rigmark/internal/modules/planning/ports"
	recommendationdomain "rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/platform/abuse"
	"rigmark/internal/platform/alerting"
	"rigmark/internal/platform/observability"
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

type AccountSecurityService interface {
	RequestEmailVerification(context.Context, string) (identity.RequestReceipt, error)
	VerifyEmail(context.Context, string) error
	RequestPasswordReset(context.Context, string) (identity.RequestReceipt, error)
	ResetPassword(context.Context, string, string) error
	ChangePassword(context.Context, domain.Principal, string, string) error
	ListSessions(context.Context, domain.Principal) ([]domain.ActiveSession, error)
	RevokeSession(context.Context, domain.Principal, string) error
	RevokeOtherSessions(context.Context, domain.Principal) (int64, error)
	ExportAccount(context.Context, domain.Principal) (domain.AccountExport, error)
	DeleteAccount(context.Context, domain.Principal, string, string) error
	BeginMFAEnrollment(context.Context, domain.Principal, string) (identity.MFAEnrollment, error)
	ConfirmMFAEnrollment(context.Context, domain.Principal, string) (identity.MFAEnabled, error)
	RegenerateRecoveryCodes(context.Context, domain.Principal, string) (identity.MFAEnabled, error)
	CompleteLoginMFA(context.Context, string, string) (identity.AuthenticatedSession, error)
	VerifyStepUp(context.Context, domain.Principal, string) error
	RecentMFA(domain.Principal) bool
	RecordAuthorizationFailure(context.Context, domain.Principal, domain.Permission) error
}

type DevelopmentEmailMessages interface {
	Messages(string) []identityports.DevelopmentMessage
}

type SecurityPolicyConfig struct {
	EnforcePrivilegedMFA   bool
	DevelopmentEmailAccess bool
}

type AuthCookieConfig struct {
	Name                   string
	Secure                 bool
	MaxAge                 int
	AllowedOrigin          string
	AnalyticsSubjectName   string
	AnalyticsSubjectMaxAge int
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
	TrackOfferClick(context.Context, commercedomain.AffiliateClick) (commercedomain.AffiliateRedirectResult, error)
	TrackLegacyLinkClick(context.Context, commercedomain.AffiliateClick) (commercedomain.AffiliateRedirectResult, error)
}

type CommerceOperationsService interface {
	CreateConfiguration(context.Context, domain.UserID, commercedomain.ProviderConfigurationInput) (commercedomain.ProviderConfiguration, error)
	ListConfigurations(context.Context) ([]commercedomain.ProviderConfiguration, error)
	SetLifecycle(context.Context, domain.UserID, commercedomain.ProviderConfigurationID, commercedomain.ProviderLifecycle) (commercedomain.ProviderConfiguration, error)
	TriggerManual(context.Context, domain.UserID, commercedomain.ProviderConfigurationID, string) (commercedomain.ImportRun, error)
	Retry(context.Context, domain.UserID, commercedomain.ImportRunID, string) (commercedomain.ImportRun, error)
	ListImports(context.Context, int, int) ([]commercedomain.ImportRun, int64, error)
	ListFailures(context.Context, commercedomain.ImportRunID, int, int) ([]commercedomain.ImportFailure, int64, error)
}

type ConversionOperationsService interface {
	IngestWebhook(context.Context, commercedomain.ProviderConfigurationID, commercedomain.WebhookRequest) (commerce.WebhookResult, error)
	SetProviderEnabled(context.Context, domain.UserID, commercedomain.ProviderConfigurationID, bool) (commercedomain.ProviderConfiguration, error)
	TriggerManualImport(context.Context, domain.UserID, commercedomain.ProviderConfigurationID, string) (commercedomain.ConversionImportRun, error)
	RetryImport(context.Context, domain.UserID, commercedomain.ConversionImportRunID, string) (commercedomain.ConversionImportRun, error)
	ListConversions(context.Context, commercedomain.ConversionFilter) ([]commercedomain.Conversion, int64, error)
	ListImports(context.Context, int, int) ([]commercedomain.ConversionImportRun, int64, error)
	ListReconciliations(context.Context, int, int) ([]commercedomain.ReconciliationRun, int64, error)
	Reconcile(context.Context, domain.UserID, commercedomain.ConversionImportRunID, string) (commercedomain.ReconciliationRun, error)
	Metrics(context.Context, time.Time, time.Time) (commercedomain.MonetizationReport, error)
}

type AnalyticsService interface {
	RecordClientEvent(context.Context, analyticsdomain.Event) (analyticsdomain.IngestionResult, error)
	SetConsent(context.Context, analyticsdomain.ConsentDecision) (analyticsdomain.Consent, error)
	GetConsent(context.Context, analyticsdomain.Subject) (analyticsdomain.Consent, error)
	ClaimIdentity(context.Context, []byte, string) error
}

type AnalyticsReportingService interface {
	Report(context.Context, analyticsdomain.ReportQuery) (analyticsdomain.Report, error)
}

type ContentService interface {
	List(context.Context, string, string, int) ([]contentdomain.Summary, error)
	Get(context.Context, string) (contentdomain.Entry, error)
	Author(context.Context, string) (contentdomain.Author, []contentdomain.Summary, error)
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
	List(context.Context, domain.UserID, int, int) (planningports.WishlistPage, error)
	Save(context.Context, domain.UserID, catalogdomain.ProductID) error
	Delete(context.Context, domain.UserID, catalogdomain.ProductID) error
}

type OperationalMetricsSource interface {
	Collect(context.Context) (map[string]float64, error)
}

type PublicServices struct {
	Catalog              CatalogService
	Commerce             CommerceService
	CommerceOperations   CommerceOperationsService
	ConversionOperations ConversionOperationsService
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
	RateLimiter          abuse.Limiter
	RateLimitKey         []byte
	Metrics              observability.Recorder
	MetricsToken         string
	OperationalMetrics   OperationalMetricsSource
	Alerts               alerting.Notifier
	HandlerTimeout       time.Duration
	// SPAShellURL is where the built index.html can be fetched so per-route
	// metadata can be injected into it. Empty disables injection and the edge
	// serves the static shell unchanged.
	SPAShellURL      string
	Security         AccountSecurityService
	DevelopmentEmail DevelopmentEmailMessages
	SecurityPolicy   SecurityPolicyConfig
}

type Handler struct {
	health               HealthService
	auth                 AuthenticationService
	catalog              CatalogService
	commerce             CommerceService
	commerceOperations   CommerceOperationsService
	conversionOperations ConversionOperationsService
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
	security             AccountSecurityService
	developmentEmail     DevelopmentEmailMessages
	securityPolicy       SecurityPolicyConfig
	cookie               AuthCookieConfig
	logger               *slog.Logger
	metrics              observability.Recorder
	metricsToken         string
	operationalSource    OperationalMetricsSource
	shell                *spaShellProvider
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
		handler.commerceOperations = publicServices[0].CommerceOperations
		handler.conversionOperations = publicServices[0].ConversionOperations
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
		handler.security = publicServices[0].Security
		handler.developmentEmail = publicServices[0].DevelopmentEmail
		handler.securityPolicy = publicServices[0].SecurityPolicy
		handler.metrics = publicServices[0].Metrics
		handler.metricsToken = publicServices[0].MetricsToken
		handler.operationalSource = publicServices[0].OperationalMetrics
		if publicServices[0].SPAShellURL != "" {
			handler.shell = newSPAShellProvider(publicServices[0].SPAShellURL)
		}
	}
	mux := http.NewServeMux()

	// Public foundation contract.
	mux.HandleFunc("GET /api/health", handler.healthCheck)

	// Infrastructure probes remain separate so orchestration can distinguish
	// process liveness from dependency readiness.
	mux.HandleFunc("GET /api/v1/health/live", handler.live)
	mux.HandleFunc("GET /api/v1/health/ready", handler.healthCheck)
	// The production web edge uses this bounded resolver before serving the SPA
	// shell. Direct API requests without X-Original-URI fail closed.
	mux.HandleFunc("GET /api/v1/public-route", handler.publicRouteStatus)
	if handler.metrics != nil && handler.metricsToken != "" {
		mux.HandleFunc("GET /api/v1/metrics", handler.operationalMetrics)
		mux.HandleFunc("GET /api/v1/metrics/openmetrics", handler.openMetrics)
	}

	mux.HandleFunc("POST /api/auth/register", handler.register)
	mux.HandleFunc("POST /api/auth/login", handler.login)
	mux.HandleFunc("POST /api/auth/logout", handler.logout)
	mux.Handle("GET /api/auth/me", handler.requireAuthentication(http.HandlerFunc(handler.me)))
	if handler.security != nil {
		mux.HandleFunc("POST /api/auth/email-verification/request", handler.requestEmailVerification)
		mux.HandleFunc("POST /api/auth/email-verification/complete", handler.completeEmailVerification)
		mux.HandleFunc("POST /api/auth/password-reset/request", handler.requestPasswordReset)
		mux.HandleFunc("POST /api/auth/password-reset/complete", handler.completePasswordReset)
		mux.HandleFunc("POST /api/auth/mfa/complete", handler.completeMFALogin)
		mux.Handle("POST /api/account/security/password", handler.requireAuthentication(http.HandlerFunc(handler.changePassword)))
		mux.Handle("GET /api/account/security/sessions", handler.requireAuthentication(http.HandlerFunc(handler.listAccountSessions)))
		mux.Handle("DELETE /api/account/security/sessions/{sessionID}", handler.requireAuthentication(http.HandlerFunc(handler.revokeAccountSession)))
		mux.Handle("DELETE /api/account/security/sessions", handler.requireAuthentication(http.HandlerFunc(handler.revokeOtherAccountSessions)))
		mux.Handle("GET /api/account/export", handler.requireAuthentication(http.HandlerFunc(handler.exportAccount)))
		mux.Handle("DELETE /api/account", handler.requireAuthentication(http.HandlerFunc(handler.deleteAccount)))
		mux.Handle("POST /api/account/security/mfa/enroll", handler.requireAuthentication(http.HandlerFunc(handler.beginMFAEnrollment)))
		mux.Handle("POST /api/account/security/mfa/verify", handler.requireAuthentication(http.HandlerFunc(handler.confirmMFAEnrollment)))
		mux.Handle("POST /api/account/security/mfa/recovery-codes", handler.requireAuthentication(http.HandlerFunc(handler.regenerateRecoveryCodes)))
		mux.Handle("POST /api/account/security/mfa/step-up", handler.requireAuthentication(http.HandlerFunc(handler.stepUpMFA)))
	}
	if handler.developmentEmail != nil && handler.securityPolicy.DevelopmentEmailAccess {
		mux.HandleFunc("GET /api/dev/email-deliveries", handler.listDevelopmentEmails)
	}

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
		// Registered before the slug route so an author path is not read as a
		// content slug.
		mux.HandleFunc("GET /api/content/authors/{slug}", handler.getAuthor)
		mux.HandleFunc("GET /api/content/{slug}", handler.getContent)
		mux.HandleFunc("GET /sitemap.xml", handler.sitemap)
		mux.HandleFunc("GET /robots.txt", handler.robots)
		mux.HandleFunc("GET /llms.txt", handler.llmsTxt)
	}
	if handler.catalog != nil && handler.commerce != nil {
		mux.HandleFunc("GET /api/catalog/products/{slug}/offers", handler.listOffers)
		mux.Handle("GET /api/affiliate/click/{offerID}", handler.attachOptionalAuthenticationFailOpen(http.HandlerFunc(handler.affiliateClickRedirect)))
		mux.Handle("GET /api/out/{affiliateLinkID}", handler.attachOptionalAuthenticationFailOpen(http.HandlerFunc(handler.outboundRedirect)))
	}
	if handler.analytics != nil {
		mux.Handle("POST /api/analytics/events", handler.attachAuthentication(http.HandlerFunc(handler.recordAnalyticsEvent)))
		mux.Handle("GET /api/analytics/consent", handler.attachAuthentication(http.HandlerFunc(handler.getAnalyticsConsent)))
		mux.Handle("PUT /api/analytics/consent", handler.attachAuthentication(http.HandlerFunc(handler.setAnalyticsConsent)))
		mux.Handle("POST /api/analytics/identity/claim", handler.requireAuthentication(http.HandlerFunc(handler.claimAnalyticsIdentity)))
	}
	if handler.conversionOperations != nil {
		mux.HandleFunc("POST /api/webhooks/commerce/{providerConfigurationID}", handler.commerceConversionWebhook)
	}
	allowed := func(permission domain.Permission, next http.HandlerFunc) http.Handler {
		return handler.requirePermission(permission, next)
	}
	if handler.admin != nil {
		mux.HandleFunc("GET /api/media/products/{file}", handler.productImage)
		mux.Handle("GET /api/admin/dashboard", allowed(domain.PermissionAdminRead, handler.adminDashboard))
		if handler.analyticsReporting != nil {
			mux.Handle("GET /api/admin/analytics", allowed(domain.PermissionAnalyticsRead, handler.adminAnalyticsReport))
		}
		mux.Handle("GET /api/admin/references", allowed(domain.PermissionCatalogRead, handler.adminReferences))
		mux.Handle("GET /api/admin/products", allowed(domain.PermissionCatalogRead, handler.adminListProducts))
		mux.Handle("POST /api/admin/products", allowed(domain.PermissionCatalogCreate, handler.adminCreateProduct))
		mux.Handle("GET /api/admin/products/{productID}", allowed(domain.PermissionCatalogRead, handler.adminGetProduct))
		mux.Handle("PATCH /api/admin/products/{productID}", allowed(domain.PermissionCatalogUpdate, handler.adminUpdateProduct))
		mux.Handle("PUT /api/admin/products/{productID}/status", allowed(domain.PermissionCatalogDelete, handler.adminProductStatus))
		mux.Handle("POST /api/admin/products/{productID}/images", allowed(domain.PermissionCatalogUpdate, handler.adminAddImage))
		mux.Handle("DELETE /api/admin/products/{productID}/images/{imageID}", allowed(domain.PermissionCatalogUpdate, handler.adminDeleteImage))
		mux.Handle("PUT /api/admin/products/{productID}/attributes/{key}", allowed(domain.PermissionCatalogUpdate, handler.adminUpsertAttribute))
		mux.Handle("DELETE /api/admin/products/{productID}/attributes/{key}", allowed(domain.PermissionCatalogUpdate, handler.adminDeleteAttribute))
		mux.Handle("GET /api/admin/categories", allowed(domain.PermissionCatalogRead, handler.adminListCategories))
		mux.Handle("GET /api/admin/brands", allowed(domain.PermissionCatalogRead, handler.adminListBrands))
		mux.Handle("GET /api/admin/merchants", allowed(domain.PermissionCommerceRead, handler.adminListMerchants))
		mux.Handle("GET /api/admin/offers", allowed(domain.PermissionCommerceRead, handler.adminListOffers))
		mux.Handle("POST /api/admin/offers", allowed(domain.PermissionCommerceCreate, handler.adminCreateOffer))
		mux.Handle("PATCH /api/admin/offers/{offerID}", allowed(domain.PermissionCommerceUpdate, handler.adminUpdateOffer))
		mux.Handle("GET /api/admin/affiliate-links", allowed(domain.PermissionCommerceRead, handler.adminListAffiliateLinks))
		mux.Handle("PATCH /api/admin/affiliate-links/{linkID}", allowed(domain.PermissionCommerceUpdate, handler.adminUpdateAffiliateLink))
		mux.Handle("GET /api/admin/recommendations", allowed(domain.PermissionAnalyticsRead, handler.adminListRecommendations))
		mux.Handle("GET /api/admin/recommendations/{recommendationID}", allowed(domain.PermissionAnalyticsRead, handler.adminGetRecommendation))
		mux.Handle("GET /api/admin/users", allowed(domain.PermissionUsersRead, handler.adminListUsers))
		mux.Handle("GET /api/admin/events", allowed(domain.PermissionAnalyticsRawRead, handler.adminListEvents))
		if handler.evidence != nil {
			evidenceEditor := func(next http.HandlerFunc) http.Handler {
				return allowed(domain.PermissionEvidenceCreate, next)
			}
			evidenceReviewer := func(next http.HandlerFunc) http.Handler {
				return allowed(domain.PermissionEvidenceApprove, next)
			}
			mux.Handle("GET /api/admin/evidence/products", allowed(domain.PermissionEvidenceRead, handler.adminListProductGovernance))
			mux.Handle("GET /api/admin/evidence/products/{productID}", allowed(domain.PermissionEvidenceRead, handler.adminGetProductGovernance))
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
				return allowed(domain.PermissionPolicyCreate, next)
			}
			policyReviewer := func(next http.HandlerFunc) http.Handler {
				return allowed(domain.PermissionPolicyApprove, next)
			}
			mux.Handle("GET /api/admin/recommendation-policies", allowed(domain.PermissionPolicyRead, handler.adminListRecommendationPolicies))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/submit", policyEditor(handler.adminSubmitRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/approve", policyReviewer(handler.adminApproveRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/reject", policyReviewer(handler.adminRejectRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/activate", policyReviewer(handler.adminActivateRecommendationPolicy))
			mux.Handle("POST /api/admin/recommendation-policies/{version}/deactivate", policyReviewer(handler.adminDeactivateRecommendationPolicy))
		}
	}
	if handler.commerceOperations != nil {
		mux.Handle("GET /api/admin/commerce/providers", allowed(domain.PermissionCommerceRead, handler.adminListCommerceProviders))
		mux.Handle("POST /api/admin/commerce/providers", allowed(domain.PermissionCommerceCreate, handler.adminCreateCommerceProvider))
		mux.Handle("PUT /api/admin/commerce/providers/{providerID}/lifecycle", allowed(domain.PermissionCommerceActivate, handler.adminSetCommerceProviderLifecycle))
		mux.Handle("GET /api/admin/commerce/imports", allowed(domain.PermissionCommerceRead, handler.adminListCommerceImports))
		mux.Handle("POST /api/admin/commerce/imports", allowed(domain.PermissionCommerceUpdate, handler.adminTriggerCommerceImport))
		mux.Handle("POST /api/admin/commerce/imports/{importID}/retry", allowed(domain.PermissionCommerceUpdate, handler.adminRetryCommerceImport))
		mux.Handle("GET /api/admin/commerce/imports/{importID}/failures", allowed(domain.PermissionCommerceRead, handler.adminListCommerceImportFailures))
	}
	if handler.conversionOperations != nil {
		mux.Handle("PUT /api/admin/commerce/providers/{providerID}/conversions", allowed(domain.PermissionCommerceActivate, handler.adminSetConversionProvider))
		mux.Handle("GET /api/admin/commerce/conversions", allowed(domain.PermissionCommerceRead, handler.adminListConversions))
		mux.Handle("GET /api/admin/commerce/conversion-imports", allowed(domain.PermissionCommerceRead, handler.adminListConversionImports))
		mux.Handle("POST /api/admin/commerce/conversion-imports", allowed(domain.PermissionCommerceUpdate, handler.adminTriggerConversionImport))
		mux.Handle("POST /api/admin/commerce/conversion-imports/{importID}/retry", allowed(domain.PermissionCommerceUpdate, handler.adminRetryConversionImport))
		mux.Handle("GET /api/admin/commerce/reconciliations", allowed(domain.PermissionCommerceRead, handler.adminListConversionReconciliations))
		mux.Handle("POST /api/admin/commerce/conversion-imports/{importID}/reconcile", allowed(domain.PermissionCommerceUpdate, handler.adminReconcileConversionImport))
		mux.Handle("GET /api/admin/commerce/metrics", allowed(domain.PermissionCommerceRead, handler.adminMonetizationMetrics))
	}
	if handler.wishlist != nil {
		mux.Handle("GET /api/account/wishlist", handler.requireAuthentication(http.HandlerFunc(handler.listWishlist)))
		mux.Handle("PUT /api/account/wishlist/{productID}", handler.requireAuthentication(http.HandlerFunc(handler.saveWishlist)))
		mux.Handle("DELETE /api/account/wishlist/{productID}", handler.requireAuthentication(http.HandlerFunc(handler.deleteWishlist)))
	}
	if handler.recommendations != nil {
		mux.Handle("POST /api/recommendations/generate", handler.attachAuthentication(http.HandlerFunc(handler.generateRecommendation)))
		mux.Handle("POST /api/recommendations/preview", handler.attachAuthentication(http.HandlerFunc(handler.previewRecommendation)))
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
					rateLimitRequestsWithBackend(
						requestDeadline(securityRequestContext(apiCacheDefaults(captureRoutePattern(mux))), publicHandlerTimeout(publicServices)),
						publicRateLimits(publicServices), publicRateLimiter(publicServices), publicRateLimitKey(publicServices),
						publicMetrics(publicServices), publicAlerts(publicServices), logger,
					),
					logger,
					cookie.AllowedOrigin,
				),
			),
			logger,
		),
		logger, publicMetrics(publicServices),
	)
}

func publicRateLimiter(services []PublicServices) abuse.Limiter {
	if len(services) == 0 || services[0].RateLimiter == nil {
		return abuse.NewLocalLimiter()
	}
	return services[0].RateLimiter
}

func publicRateLimitKey(services []PublicServices) []byte {
	if len(services) == 0 || len(services[0].RateLimitKey) == 0 {
		return []byte("local-router-rate-limit-key")
	}
	return services[0].RateLimitKey
}

func publicMetrics(services []PublicServices) observability.Recorder {
	if len(services) == 0 || services[0].Metrics == nil {
		return observability.DisabledRecorder{}
	}
	return services[0].Metrics
}

func publicAlerts(services []PublicServices) alerting.Notifier {
	if len(services) == 0 || services[0].Alerts == nil {
		return alerting.Disabled{}
	}
	return services[0].Alerts
}

func publicHandlerTimeout(services []PublicServices) time.Duration {
	if len(services) == 0 || services[0].HandlerTimeout <= 0 {
		return 20 * time.Second
	}
	return services[0].HandlerTimeout
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
		if h.metrics != nil {
			recordReadinessFailure(h.metrics, report)
		}
		h.logger.Warn("health check failed", "error", err)
		writeReport(response, http.StatusServiceUnavailable, report, h.logger)
		return
	}
	writeReport(response, http.StatusOK, report, h.logger)
}

func recordReadinessFailure(metrics observability.Recorder, report health.Report) {
	metrics.Increment(observability.MetricReadinessFailure)
	if report.Checks["database"] == "unavailable" {
		metrics.Increment(observability.MetricDatabaseAcquireFailure)
	}
	if report.Checks["schema"] == "unavailable" {
		metrics.Increment(observability.MetricMigrationFailure)
	}
	if report.Checks["rate_limit"] == "unavailable" {
		metrics.Increment(observability.MetricRedisUnavailable)
	}
	if report.Checks["media_storage"] == "unavailable" {
		metrics.Increment(observability.MetricStorageFailure)
	}
}

func writeReport(response http.ResponseWriter, status int, report health.Report, logger *slog.Logger) {
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, status, report, logger)
}
