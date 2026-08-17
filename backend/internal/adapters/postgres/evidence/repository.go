package evidencepostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/evidence/domain"
	"rigmark/internal/modules/evidence/ports"
	identity "rigmark/internal/modules/identity/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) CreateSource(ctx context.Context, actor identity.UserID, input domain.SourceInput) (domain.Source, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Source{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var source domain.Source
	var sourceURL any
	if input.URL != nil {
		sourceURL = *input.URL
	}
	err = tx.QueryRow(ctx, `INSERT INTO evidence.sources (
		source_type, title, publisher, source_url, is_fictional, created_by_user_id
	) VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id, source_type, title, publisher, source_url, is_fictional,
		review_status, created_by_user_id, created_at, updated_at`,
		input.Type, input.Title, input.Publisher, sourceURL, input.IsFictional, actor).Scan(
		&source.ID, &source.Type, &source.Title, &source.Publisher, &source.URL,
		&source.IsFictional, &source.ReviewStatus, &source.CreatedByUserID,
		&source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return domain.Source{}, mapError("create evidence source", err)
	}
	err = writeAudit(ctx, tx, actor, "evidence.source.create", "evidence_source", source.ID,
		map[string]string{"source_type": string(source.Type), "title": source.Title})
	if err = commit(ctx, tx, err); err != nil {
		return domain.Source{}, mapError("create evidence source", err)
	}
	return source, nil
}

func (repository *Repository) ReviewSource(ctx context.Context, actor identity.UserID, id string, status domain.ReviewStatus, note string) (domain.Source, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Source{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var creator sql.NullString
	if err = tx.QueryRow(ctx, `SELECT created_by_user_id FROM evidence.sources WHERE id=$1 FOR UPDATE`, id).Scan(&creator); errors.Is(err, pgx.ErrNoRows) {
		return domain.Source{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Source{}, fmt.Errorf("lock evidence source: %w", err)
	}
	if creator.Valid && creator.String == string(actor) {
		return domain.Source{}, ports.ErrSeparationOfDuties
	}
	var source domain.Source
	err = tx.QueryRow(ctx, `UPDATE evidence.sources SET review_status=$2,
		reviewer_user_id=$3, reviewed_at=now(), review_note=$4, updated_at=now()
		WHERE id=$1 RETURNING id, source_type, title, publisher, source_url,
		is_fictional, review_status, reviewer_user_id, reviewed_at, review_note,
		created_by_user_id, created_at, updated_at`, id, status, actor, note).Scan(
		&source.ID, &source.Type, &source.Title, &source.Publisher, &source.URL,
		&source.IsFictional, &source.ReviewStatus, &source.ReviewerUserID,
		&source.ReviewedAt, &source.ReviewNote, &source.CreatedByUserID,
		&source.CreatedAt, &source.UpdatedAt)
	if err == nil {
		err = writeAudit(ctx, tx, actor, "evidence.source.review", "evidence_source", id,
			map[string]string{"status": string(status), "note": note})
	}
	if err = commit(ctx, tx, err); err != nil {
		return domain.Source{}, mapError("review evidence source", err)
	}
	return source, nil
}

func (repository *Repository) CreateObservation(ctx context.Context, actor identity.UserID, input domain.ObservationInput) (domain.Observation, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Observation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var observation domain.Observation
	err = tx.QueryRow(ctx, `INSERT INTO evidence.observations (
		source_id, product_id, observed_at, expires_at, confidence, notes, created_by_user_id
	) SELECT $1,$2,$3,$4,$5,$6,$7
	WHERE EXISTS (SELECT 1 FROM evidence.sources WHERE id=$1)
	  AND EXISTS (SELECT 1 FROM catalog.products WHERE id=$2)
	RETURNING id, source_id, product_id, observed_at, expires_at, confidence, notes, created_at`,
		input.SourceID, input.ProductID, input.ObservedAt, input.ExpiresAt,
		input.Confidence, input.Notes, actor).Scan(&observation.ID, &observation.SourceID,
		&observation.ProductID, &observation.ObservedAt, &observation.ExpiresAt,
		&observation.Confidence, &observation.Notes, &observation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Observation{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Observation{}, mapError("create evidence observation", err)
	}
	err = writeAudit(ctx, tx, actor, "evidence.observation.create", "evidence_observation", observation.ID,
		map[string]string{"product_id": string(observation.ProductID), "source_id": observation.SourceID})
	if err = commit(ctx, tx, err); err != nil {
		return domain.Observation{}, mapError("create evidence observation", err)
	}
	return observation, nil
}

func (repository *Repository) CreateRevision(ctx context.Context, actor identity.UserID, input domain.RevisionInput) (domain.Revision, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Revision{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT true FROM catalog.products WHERE id=$1 FOR UPDATE`, input.Product.ID).Scan(&exists); errors.Is(err, pgx.ErrNoRows) {
		return domain.Revision{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Revision{}, fmt.Errorf("lock product for revision: %w", err)
	}
	observationIDs := revisionObservationIDs(input)
	var matching int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM evidence.observations
		WHERE product_id=$1 AND id=ANY($2::uuid[])`, input.Product.ID, observationIDs).Scan(&matching); err != nil {
		return domain.Revision{}, fmt.Errorf("validate revision observations: %w", err)
	}
	if matching != len(observationIDs) {
		return domain.Revision{}, ports.ErrIncompleteProvenance
	}
	var factVersion, scoreVersion int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(max(version),0)+1 FROM evidence.product_fact_revisions WHERE product_id=$1`, input.Product.ID).Scan(&factVersion); err != nil {
		return domain.Revision{}, err
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE(max(version),0)+1 FROM evidence.score_revisions WHERE product_id=$1`, input.Product.ID).Scan(&scoreVersion); err != nil {
		return domain.Revision{}, err
	}
	var factID, scoreID string
	p := input.Product
	err = tx.QueryRow(ctx, `INSERT INTO evidence.product_fact_revisions (
		product_id, version, category_id, brand_id, name, slug, description,
		price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
		max_capacity_grams, material, warranty_months, created_by_user_id
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	RETURNING id`, p.ID, factVersion, p.CategoryID, p.BrandID, p.Name, p.Slug,
		p.Description, p.Price.AmountMinor, p.Price.Currency, p.Dimensions.LengthMM,
		p.Dimensions.WidthMM, p.Dimensions.HeightMM, p.WeightGrams,
		p.MaxCapacityGrams, p.Material, p.WarrantyMonths, actor).Scan(&factID)
	if err != nil {
		return domain.Revision{}, mapError("insert fact revision", err)
	}
	s := input.Scores
	err = tx.QueryRow(ctx, `INSERT INTO evidence.score_revisions (
		product_id, fact_revision_id, version, quality_score, value_score,
		durability_score, beginner_score, advanced_score, apartment_score,
		noise_score, portability_score, created_by_user_id
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		p.ID, factID, scoreVersion, s.Quality, s.Value, s.Durability, s.Beginner,
		s.Advanced, s.Apartment, s.Noise, s.Portability, actor).Scan(&scoreID)
	if err != nil {
		return domain.Revision{}, mapError("insert score revision", err)
	}
	for _, link := range input.FactLinks {
		if _, err = tx.Exec(ctx, `INSERT INTO evidence.fact_provenance
			(fact_revision_id, fact_key, observation_id, public_classification)
			VALUES ($1,$2,$3,$4)`, factID, link.FactKey, link.ObservationID,
			link.Classification); err != nil {
			return domain.Revision{}, mapError("insert fact provenance", err)
		}
	}
	for _, rationale := range input.Rationales {
		if _, err = tx.Exec(ctx, `INSERT INTO evidence.score_rationales
			(score_revision_id, score_key, rationale, observation_id)
			VALUES ($1,$2,$3,$4)`, scoreID, rationale.ScoreKey,
			rationale.Rationale, rationale.ObservationID); err != nil {
			return domain.Revision{}, mapError("insert score rationale", err)
		}
	}
	if err = writeAudit(ctx, tx, actor, "evidence.revision.create", "product_fact_revision", factID,
		map[string]string{"product_id": string(p.ID), "fact_version": fmt.Sprint(factVersion), "score_version": fmt.Sprint(scoreVersion)}); err != nil {
		return domain.Revision{}, err
	}
	if err = commit(ctx, tx, nil); err != nil {
		return domain.Revision{}, err
	}
	actorCopy := actor
	return domain.Revision{FactRevisionID: factID, ScoreRevisionID: scoreID,
		ProductID: p.ID, FactVersion: factVersion, ScoreVersion: scoreVersion,
		Status: domain.WorkflowDraft, CreatedByUserID: &actorCopy, CreatedAt: time.Now()}, nil
}

func revisionObservationIDs(input domain.RevisionInput) []string {
	set := make(map[string]bool)
	for _, link := range input.FactLinks {
		set[link.ObservationID] = true
	}
	for _, rationale := range input.Rationales {
		set[rationale.ObservationID] = true
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (repository *Repository) TransitionRevision(ctx context.Context, actor identity.UserID, factID string, target domain.WorkflowStatus, note string) (domain.Revision, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Revision{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var revision domain.Revision
	var createdBy, submittedBy sql.NullString
	err = tx.QueryRow(ctx, `SELECT facts.product_id, facts.version, scores.id, scores.version,
		facts.workflow_status, facts.created_by_user_id, facts.submitted_by_user_id, facts.created_at
		FROM evidence.product_fact_revisions facts
		JOIN evidence.score_revisions scores ON scores.fact_revision_id=facts.id
		WHERE facts.id=$1 FOR UPDATE OF facts, scores`, factID).Scan(&revision.ProductID,
		&revision.FactVersion, &revision.ScoreRevisionID, &revision.ScoreVersion,
		&revision.Status, &createdBy, &submittedBy, &revision.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Revision{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Revision{}, fmt.Errorf("lock evidence revision: %w", err)
	}
	revision.FactRevisionID = factID
	allowed := false
	switch target {
	case domain.WorkflowInReview:
		allowed = revision.Status == domain.WorkflowDraft || revision.Status == domain.WorkflowRejected
		if allowed {
			_, err = tx.Exec(ctx, `UPDATE evidence.product_fact_revisions SET workflow_status='in_review',
				submitted_by_user_id=$2, submitted_at=now(), reviewed_by_user_id=NULL,
				reviewed_at=NULL, review_note='' WHERE id=$1`, factID, actor)
			if err == nil {
				_, err = tx.Exec(ctx, `UPDATE evidence.score_revisions SET workflow_status='in_review',
					submitted_by_user_id=$2, submitted_at=now(), reviewed_by_user_id=NULL,
					reviewed_at=NULL, review_note='' WHERE id=$1`, revision.ScoreRevisionID, actor)
			}
		}
	case domain.WorkflowApproved, domain.WorkflowRejected:
		allowed = revision.Status == domain.WorkflowInReview
		if allowed && ((createdBy.Valid && createdBy.String == string(actor)) ||
			(submittedBy.Valid && submittedBy.String == string(actor))) {
			return domain.Revision{}, ports.ErrSeparationOfDuties
		}
		if allowed {
			_, err = tx.Exec(ctx, `UPDATE evidence.product_fact_revisions SET workflow_status=$2,
				reviewed_by_user_id=$3, reviewed_at=now(), review_note=$4 WHERE id=$1`, factID, target, actor, note)
			if err == nil {
				_, err = tx.Exec(ctx, `UPDATE evidence.score_revisions SET workflow_status=$2,
					reviewed_by_user_id=$3, reviewed_at=now(), review_note=$4 WHERE id=$1`, revision.ScoreRevisionID, target, actor, note)
			}
		}
	}
	if !allowed {
		return domain.Revision{}, ports.ErrConflict
	}
	if err == nil {
		err = writeAudit(ctx, tx, actor, "evidence.revision."+string(target), "product_fact_revision", factID,
			map[string]string{"status": string(target), "note": note})
	}
	if err = commit(ctx, tx, err); err != nil {
		return domain.Revision{}, mapError("transition evidence revision", err)
	}
	revision.Status, revision.ReviewNote = target, note
	return revision, nil
}

func (repository *Repository) PublishRevision(ctx context.Context, actor identity.UserID, factID string) (domain.Revision, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Revision{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var revision domain.Revision
	var maxCapacity sql.NullInt64
	var createdBy, submittedBy, reviewedBy sql.NullString
	var scoreStatus domain.WorkflowStatus
	err = tx.QueryRow(ctx, `SELECT facts.product_id, facts.version, scores.id, scores.version,
		facts.workflow_status, scores.workflow_status, facts.created_by_user_id, facts.submitted_by_user_id,
		facts.reviewed_by_user_id, facts.max_capacity_grams, facts.created_at
		FROM evidence.product_fact_revisions facts
		JOIN evidence.score_revisions scores ON scores.fact_revision_id=facts.id
		WHERE facts.id=$1 FOR UPDATE OF facts, scores`, factID).Scan(&revision.ProductID,
		&revision.FactVersion, &revision.ScoreRevisionID, &revision.ScoreVersion,
		&revision.Status, &scoreStatus, &createdBy, &submittedBy, &reviewedBy, &maxCapacity, &revision.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Revision{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Revision{}, fmt.Errorf("lock approved revision: %w", err)
	}
	if revision.Status != domain.WorkflowApproved || scoreStatus != domain.WorkflowApproved {
		return domain.Revision{}, ports.ErrConflict
	}
	if !reviewedBy.Valid || reviewedBy.String == string(actor) {
		return domain.Revision{}, ports.ErrSeparationOfDuties
	}
	revision.FactRevisionID = factID
	requiredFacts := 10
	if maxCapacity.Valid {
		requiredFacts++
	}
	var factCount, scoreCount, invalidEvidence int
	var validUntil sql.NullTime
	err = tx.QueryRow(ctx, `SELECT
		(SELECT count(DISTINCT fact_key) FROM evidence.fact_provenance WHERE fact_revision_id=$1),
		(SELECT count(DISTINCT score_key) FROM evidence.score_rationales WHERE score_revision_id=$2),
		(SELECT count(*) FROM (
			SELECT observations.id FROM evidence.fact_provenance links
			JOIN evidence.observations observations ON observations.id=links.observation_id
			JOIN evidence.sources sources ON sources.id=observations.source_id
			WHERE links.fact_revision_id=$1 AND (observations.product_id<>$3
				OR sources.review_status<>'verified'
				OR links.public_classification <> CASE sources.source_type
					WHEN 'manufacturer_documentation' THEN 'manufacturer_claim'
					WHEN 'independent_testing' THEN 'verified_fact'
					WHEN 'verified_merchant_data' THEN 'merchant_observation'
					ELSE 'editorial_assessment'
				END
				OR (observations.expires_at IS NOT NULL AND observations.expires_at<=now()))
			UNION ALL
			SELECT observations.id FROM evidence.score_rationales rationales
			JOIN evidence.observations observations ON observations.id=rationales.observation_id
			JOIN evidence.sources sources ON sources.id=observations.source_id
			WHERE rationales.score_revision_id=$2 AND (observations.product_id<>$3 OR sources.review_status<>'verified' OR (observations.expires_at IS NOT NULL AND observations.expires_at<=now()))
		) invalid),
		(SELECT min(expires_at) FROM evidence.observations WHERE id IN (
			SELECT observation_id FROM evidence.fact_provenance WHERE fact_revision_id=$1
			UNION SELECT observation_id FROM evidence.score_rationales WHERE score_revision_id=$2
		))`, factID, revision.ScoreRevisionID, revision.ProductID).Scan(&factCount, &scoreCount, &invalidEvidence, &validUntil)
	if err != nil {
		return domain.Revision{}, fmt.Errorf("validate publication provenance: %w", err)
	}
	if factCount != requiredFacts || scoreCount != 8 || invalidEvidence != 0 {
		return domain.Revision{}, ports.ErrIncompleteProvenance
	}
	_, err = tx.Exec(ctx, `UPDATE evidence.product_fact_revisions SET workflow_status='superseded'
		WHERE product_id=$1 AND workflow_status='published' AND id<>$2`, revision.ProductID, factID)
	if err == nil {
		_, err = tx.Exec(ctx, `UPDATE evidence.score_revisions SET workflow_status='superseded'
			WHERE product_id=$1 AND workflow_status='published' AND id<>$2`, revision.ProductID, revision.ScoreRevisionID)
	}
	if err == nil {
		_, err = tx.Exec(ctx, `UPDATE evidence.product_fact_revisions SET workflow_status='published',
			published_by_user_id=$2, published_at=now(), valid_until=$3 WHERE id=$1`, factID, actor, validUntil)
	}
	if err == nil {
		_, err = tx.Exec(ctx, `UPDATE evidence.score_revisions SET workflow_status='published',
			published_by_user_id=$2, published_at=now() WHERE id=$1`, revision.ScoreRevisionID, actor)
	}
	if err == nil {
		_, err = tx.Exec(ctx, `UPDATE catalog.products products SET
			category_id=facts.category_id, brand_id=facts.brand_id, name=facts.name,
			slug=facts.slug, description=facts.description, price_minor=facts.price_minor,
			currency=facts.currency, length_mm=facts.length_mm, width_mm=facts.width_mm,
			height_mm=facts.height_mm, weight_grams=facts.weight_grams,
			max_capacity_grams=facts.max_capacity_grams, material=facts.material,
			warranty_months=facts.warranty_months, quality_score=scores.quality_score,
			value_score=scores.value_score, durability_score=scores.durability_score,
			beginner_score=scores.beginner_score, advanced_score=scores.advanced_score,
			apartment_score=scores.apartment_score, noise_score=scores.noise_score,
			portability_score=scores.portability_score, status='published',
			published_fact_revision_id=facts.id, published_score_revision_id=scores.id,
			updated_at=now()
			FROM evidence.product_fact_revisions facts
			JOIN evidence.score_revisions scores ON scores.fact_revision_id=facts.id
			WHERE products.id=facts.product_id AND facts.id=$1`, factID)
	}
	if err == nil {
		err = writeAudit(ctx, tx, actor, "evidence.revision.publish", "product_fact_revision", factID,
			map[string]string{"product_id": string(revision.ProductID), "score_revision_id": revision.ScoreRevisionID})
	}
	if err = commit(ctx, tx, err); err != nil {
		return domain.Revision{}, mapError("publish evidence revision", err)
	}
	revision.Status = domain.WorkflowPublished
	if validUntil.Valid {
		revision.ValidUntil = &validUntil.Time
	}
	return revision, nil
}

func (repository *Repository) ListProductGovernance(ctx context.Context, limit, offset int) ([]domain.ProductGovernance, int64, error) {
	rows, err := repository.pool.Query(ctx, `SELECT count(*) OVER(), products.id, products.name,
		products.status, products.published_fact_revision_id, products.published_score_revision_id
		FROM catalog.products products ORDER BY products.updated_at DESC, products.name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list product governance: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ProductGovernance, 0)
	var total int64
	for rows.Next() {
		var item domain.ProductGovernance
		if err = rows.Scan(&total, &item.ProductID, &item.ProductName, &item.Status,
			&item.PublishedFactRevisionID, &item.PublishedScoreRevisionID); err != nil {
			return nil, 0, fmt.Errorf("scan product governance: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (repository *Repository) GetProductGovernance(ctx context.Context, productID catalog.ProductID) (domain.ProductGovernance, error) {
	var result domain.ProductGovernance
	err := repository.pool.QueryRow(ctx, `SELECT id, name, status,
		published_fact_revision_id, published_score_revision_id
		FROM catalog.products WHERE id=$1`, productID).Scan(&result.ProductID,
		&result.ProductName, &result.Status, &result.PublishedFactRevisionID,
		&result.PublishedScoreRevisionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProductGovernance{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.ProductGovernance{}, fmt.Errorf("get governed product: %w", err)
	}
	if result.Revisions, err = repository.listRevisions(ctx, productID); err != nil {
		return domain.ProductGovernance{}, err
	}
	if result.Provenance, err = repository.listProvenance(ctx, productID); err != nil {
		return domain.ProductGovernance{}, err
	}
	if result.Audit, err = repository.listAudit(ctx, productID); err != nil {
		return domain.ProductGovernance{}, err
	}
	return result, nil
}

func (repository *Repository) listRevisions(ctx context.Context, productID catalog.ProductID) ([]domain.Revision, error) {
	rows, err := repository.pool.Query(ctx, `SELECT facts.id, scores.id, facts.product_id,
		facts.version, scores.version, facts.workflow_status, facts.created_by_user_id,
		facts.submitted_by_user_id, facts.reviewed_by_user_id, facts.published_by_user_id,
		facts.created_at, facts.submitted_at, facts.reviewed_at, facts.published_at,
		facts.valid_until, facts.review_note
		FROM evidence.product_fact_revisions facts
		JOIN evidence.score_revisions scores ON scores.fact_revision_id=facts.id
		WHERE facts.product_id=$1 ORDER BY facts.version DESC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list evidence revisions: %w", err)
	}
	defer rows.Close()
	result := make([]domain.Revision, 0)
	for rows.Next() {
		var item domain.Revision
		if err = rows.Scan(&item.FactRevisionID, &item.ScoreRevisionID, &item.ProductID,
			&item.FactVersion, &item.ScoreVersion, &item.Status, &item.CreatedByUserID,
			&item.SubmittedByUserID, &item.ReviewedByUserID, &item.PublishedByUserID,
			&item.CreatedAt, &item.SubmittedAt, &item.ReviewedAt, &item.PublishedAt,
			&item.ValidUntil, &item.ReviewNote); err != nil {
			return nil, fmt.Errorf("scan evidence revision: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repository *Repository) listProvenance(ctx context.Context, productID catalog.ProductID) ([]domain.Provenance, error) {
	rows, err := repository.pool.Query(ctx, `SELECT links.fact_key, '' AS score_key,
		links.public_classification, '' AS rationale, observations.id,
		observations.source_id, observations.product_id, observations.observed_at,
		observations.expires_at, observations.confidence, observations.notes,
		observations.created_at, sources.id, sources.source_type, sources.title,
		sources.publisher, sources.source_url, sources.is_fictional,
		sources.review_status, sources.reviewer_user_id, sources.reviewed_at,
		sources.review_note, sources.created_by_user_id, sources.created_at, sources.updated_at
		FROM evidence.fact_provenance links
		JOIN evidence.product_fact_revisions facts ON facts.id=links.fact_revision_id
		JOIN evidence.observations observations ON observations.id=links.observation_id
		JOIN evidence.sources sources ON sources.id=observations.source_id
		WHERE facts.product_id=$1
		UNION ALL
		SELECT '' AS fact_key, rationales.score_key, '' AS classification,
		rationales.rationale, observations.id, observations.source_id,
		observations.product_id, observations.observed_at, observations.expires_at,
		observations.confidence, observations.notes, observations.created_at,
		sources.id, sources.source_type, sources.title, sources.publisher,
		sources.source_url, sources.is_fictional, sources.review_status,
		sources.reviewer_user_id, sources.reviewed_at, sources.review_note,
		sources.created_by_user_id, sources.created_at, sources.updated_at
		FROM evidence.score_rationales rationales
		JOIN evidence.score_revisions scores ON scores.id=rationales.score_revision_id
		JOIN evidence.observations observations ON observations.id=rationales.observation_id
		JOIN evidence.sources sources ON sources.id=observations.source_id
		WHERE scores.product_id=$1
		ORDER BY fact_key, score_key, title`, productID)
	if err != nil {
		return nil, fmt.Errorf("list provenance: %w", err)
	}
	defer rows.Close()
	result := make([]domain.Provenance, 0)
	for rows.Next() {
		var item domain.Provenance
		if err = rows.Scan(&item.FactKey, &item.ScoreKey, &item.Classification,
			&item.Rationale, &item.Observation.ID, &item.Observation.SourceID,
			&item.Observation.ProductID, &item.Observation.ObservedAt,
			&item.Observation.ExpiresAt, &item.Observation.Confidence,
			&item.Observation.Notes, &item.Observation.CreatedAt,
			&item.Source.ID, &item.Source.Type, &item.Source.Title,
			&item.Source.Publisher, &item.Source.URL, &item.Source.IsFictional,
			&item.Source.ReviewStatus, &item.Source.ReviewerUserID,
			&item.Source.ReviewedAt, &item.Source.ReviewNote,
			&item.Source.CreatedByUserID, &item.Source.CreatedAt,
			&item.Source.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provenance: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repository *Repository) listAudit(ctx context.Context, productID catalog.ProductID) ([]domain.AuditEvent, error) {
	rows, err := repository.pool.Query(ctx, `SELECT audit.action, users.email, audit.changes, audit.occurred_at
		FROM admin.audit_log audit
		LEFT JOIN identity.users users ON users.id=audit.actor_user_id
		WHERE (audit.entity_type='product_fact_revision' AND audit.entity_id IN (
			SELECT id::text FROM evidence.product_fact_revisions WHERE product_id=$1
		)) OR (audit.entity_type='evidence_observation' AND audit.entity_id IN (
			SELECT id::text FROM evidence.observations WHERE product_id=$1
		)) OR (audit.entity_type='evidence_source' AND audit.entity_id IN (
			SELECT source_id::text FROM evidence.observations WHERE product_id=$1
		)) OR (audit.entity_type='product' AND audit.entity_id=$1::text)
		ORDER BY audit.occurred_at DESC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list evidence audit: %w", err)
	}
	defer rows.Close()
	result := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var item domain.AuditEvent
		var raw []byte
		if err = rows.Scan(&item.Action, &item.ActorEmail, &raw, &item.OccurredAt); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &item.Changes); err != nil {
			return nil, fmt.Errorf("decode evidence audit: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func writeAudit(ctx context.Context, tx pgx.Tx, actor identity.UserID, action, entityType, entityID string, changes map[string]string) error {
	payload, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO admin.audit_log
		(actor_user_id, action, entity_type, entity_id, changes)
		VALUES ($1,$2,$3,$4,$5)`, actor, action, entityType, entityID, payload)
	return err
}

func commit(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func mapError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && (postgresError.Code == "23505" || postgresError.Code == "23514" || postgresError.Code == "23503") {
		return ports.ErrConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}

var _ ports.Repository = (*Repository)(nil)
