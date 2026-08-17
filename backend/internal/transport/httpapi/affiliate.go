package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"

	commerce "rigmark/internal/modules/commerce/application"
	commercedomain "rigmark/internal/modules/commerce/domain"
	commerceports "rigmark/internal/modules/commerce/ports"
)

func (h *Handler) affiliateClickRedirect(response http.ResponseWriter, request *http.Request) {
	offerID := request.PathValue("offerID")
	if !validUUID(offerID) {
		h.writeAffiliateNotFound(response)
		return
	}
	click, ok := h.affiliateAttribution(response, request)
	if !ok {
		return
	}
	click.OfferID = commercedomain.OfferID(offerID)
	destination, err := h.commerce.TrackOfferClick(request.Context(), click)
	h.finishAffiliateRedirect(response, request, destination, err)
}

// outboundRedirect preserves previously issued tracked URLs while new offer
// responses expose only the provider-agnostic offer endpoint.
func (h *Handler) outboundRedirect(response http.ResponseWriter, request *http.Request) {
	linkID := request.PathValue("affiliateLinkID")
	if !validUUID(linkID) {
		h.writeAffiliateNotFound(response)
		return
	}
	click, ok := h.affiliateAttribution(response, request)
	if !ok {
		return
	}
	click.LinkID = commercedomain.AffiliateLinkID(linkID)
	destination, err := h.commerce.TrackLegacyLinkClick(request.Context(), click)
	h.finishAffiliateRedirect(response, request, destination, err)
}

func (h *Handler) affiliateAttribution(response http.ResponseWriter, request *http.Request) (commercedomain.AffiliateClick, bool) {
	values := request.URL.Query()
	source := strings.TrimSpace(values.Get("source"))
	if source == "" {
		source = "product_detail"
	}
	sessionID := strings.TrimSpace(values.Get("session_id"))
	if sessionID == "" {
		var err error
		sessionID, err = randomUUID()
		if err != nil {
			h.logger.Error("create anonymous affiliate session", "error", err)
			writeAPIError(response, http.StatusServiceUnavailable, "redirect_unavailable", "We could not open this merchant right now.", nil, h.logger)
			return commercedomain.AffiliateClick{}, false
		}
	} else if !validUUID(sessionID) {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_attribution", "The merchant attribution is invalid.", nil, h.logger)
		return commercedomain.AffiliateClick{}, false
	}
	click := commercedomain.AffiliateClick{Source: source, SessionID: &sessionID}
	if principal, authenticated := principalFromContext(request.Context()); authenticated {
		userID := string(principal.UserID)
		click.UserID = &userID
	} else {
		click.AnonymousID = &sessionID
	}
	if campaign := strings.TrimSpace(values.Get("campaign")); campaign != "" {
		click.Campaign = &campaign
	}
	if source := strings.TrimSpace(values.Get("traffic_source")); source != "" {
		click.TrafficSource = &source
	}
	if medium := strings.TrimSpace(values.Get("traffic_medium")); medium != "" {
		click.TrafficMedium = &medium
	}
	if recommendationID := strings.TrimSpace(values.Get("recommendation_id")); recommendationID != "" {
		if !validUUID(recommendationID) {
			writeAPIError(response, http.StatusUnprocessableEntity, "invalid_attribution", "The merchant attribution is invalid.", nil, h.logger)
			return commercedomain.AffiliateClick{}, false
		}
		click.RecommendationID = &recommendationID
	}
	if referrer := normalizedReferrer(request.Referer()); referrer != "" {
		click.Referrer = &referrer
	}
	if referrerHost := normalizedReferrerHost(request.Referer()); referrerHost != "" {
		click.ReferrerHost = &referrerHost
	}
	if requestID := strings.TrimSpace(request.Header.Get("X-Request-ID")); requestID != "" && len(requestID) <= 128 {
		click.RequestID = &requestID
	}
	return click, true
}

func normalizedReferrerHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

func (h *Handler) finishAffiliateRedirect(response http.ResponseWriter, request *http.Request, destination string, err error) {
	if errors.Is(err, commerceports.ErrAffiliateDestinationNotFound) || errors.Is(err, commerce.ErrInvalidAttribution) {
		h.writeAffiliateNotFound(response)
		return
	}
	if err != nil {
		h.logger.Error("track affiliate click", "error", err)
		writeAPIError(response, http.StatusServiceUnavailable, "redirect_unavailable", "We could not open this merchant right now.", nil, h.logger)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(response, request, destination, http.StatusFound)
}

func (h *Handler) writeAffiliateNotFound(response http.ResponseWriter) {
	writeAPIError(response, http.StatusNotFound, "offer_not_found", "This merchant offer is no longer available.", nil, h.logger)
}

func normalizedReferrer(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return ""
	}
	value := parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
	if len(value) > 500 {
		return parsed.Scheme + "://" + parsed.Host
	}
	return value
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := hex.EncodeToString(value)
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}
