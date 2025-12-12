// Package search provides full-text search functionality for beans using Bleve.
package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/hmans/beans/internal/bean"
)

// Index wraps a Bleve in-memory index for searching beans.
type Index struct {
	index bleve.Index
}

// beanDocument is the structure stored in the Bleve index.
type beanDocument struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Body  string `json:"body"`

	// Link targets for reverse lookup (finding incoming links)
	MilestoneTarget string   `json:"milestone_target,omitempty"`
	EpicTarget      string   `json:"epic_target,omitempty"`
	FeatureTarget   string   `json:"feature_target,omitempty"`
	BlockTargets    []string `json:"block_targets,omitempty"`
	RelatedTargets  []string `json:"related_targets,omitempty"`
	DuplicateTargets []string `json:"duplicate_targets,omitempty"`
}

// NewIndex creates a new in-memory Bleve index.
func NewIndex() (*Index, error) {
	indexMapping := buildIndexMapping()
	idx, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, err
	}

	return &Index{index: idx}, nil
}

// buildIndexMapping creates the Bleve index mapping for bean documents.
func buildIndexMapping() mapping.IndexMapping {
	// Create a text field mapping with the standard analyzer
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "standard"

	// Create a keyword field mapping for ID (stored but not analyzed)
	keywordFieldMapping := bleve.NewKeywordFieldMapping()

	// Create the document mapping
	beanMapping := bleve.NewDocumentMapping()
	beanMapping.AddFieldMappingsAt("id", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("slug", textFieldMapping)
	beanMapping.AddFieldMappingsAt("title", textFieldMapping)
	beanMapping.AddFieldMappingsAt("body", textFieldMapping)

	// Link target fields for reverse lookup (keyword fields for exact matching)
	beanMapping.AddFieldMappingsAt("milestone_target", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("epic_target", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("feature_target", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("block_targets", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("related_targets", keywordFieldMapping)
	beanMapping.AddFieldMappingsAt("duplicate_targets", keywordFieldMapping)

	// Create the index mapping with BM25 scoring for better relevance ranking
	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = beanMapping
	indexMapping.DefaultAnalyzer = "standard"
	indexMapping.IndexDynamic = false
	indexMapping.StoreDynamic = false

	// Use BM25 scoring algorithm (available in Bleve v2.5.0+)
	// BM25 provides better relevance ranking than TF-IDF, especially for:
	// - Handling term frequency saturation (repeated terms don't over-boost)
	// - Normalizing for document length (short docs aren't unfairly penalized)
	indexMapping.ScoringModel = "bm25"

	return indexMapping
}

// Close closes the index.
func (idx *Index) Close() error {
	return idx.index.Close()
}

// IndexBean adds or updates a bean in the search index.
func (idx *Index) IndexBean(b *bean.Bean) error {
	doc := beanDocument{
		ID:               b.ID,
		Slug:             b.Slug,
		Title:            b.Title,
		Body:             b.Body,
		MilestoneTarget:  b.Milestone,
		EpicTarget:       b.Epic,
		FeatureTarget:    b.Feature,
		BlockTargets:     b.Blocks,
		RelatedTargets:   b.Related,
		DuplicateTargets: b.Duplicates,
	}
	return idx.index.Index(b.ID, doc)
}

// DeleteBean removes a bean from the search index.
func (idx *Index) DeleteBean(id string) error {
	return idx.index.Delete(id)
}

// DefaultSearchLimit is the default maximum number of search results.
const DefaultSearchLimit = 1000

// Search executes a search query and returns matching bean IDs.
// The limit parameter controls the maximum number of results (0 uses DefaultSearchLimit).
func (idx *Index) Search(queryStr string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	// Use query string syntax which supports:
	// - Simple terms: "authentication"
	// - Boolean operators: "user AND password"
	// - Wildcards: "auth*"
	// - Phrases: "\"user login\""
	// - Field-specific: "title:login"
	query := bleve.NewQueryStringQuery(queryStr)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"id"} // Only return ID field

	result, err := idx.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		ids = append(ids, hit.ID)
	}

	return ids, nil
}

// IndexBeans indexes multiple beans in a batch for efficiency.
func (idx *Index) IndexBeans(beans []*bean.Bean) error {
	batch := idx.index.NewBatch()
	for _, b := range beans {
		doc := beanDocument{
			ID:               b.ID,
			Slug:             b.Slug,
			Title:            b.Title,
			Body:             b.Body,
			MilestoneTarget:  b.Milestone,
			EpicTarget:       b.Epic,
			FeatureTarget:    b.Feature,
			BlockTargets:     b.Blocks,
			RelatedTargets:   b.Related,
			DuplicateTargets: b.Duplicates,
		}
		if err := batch.Index(b.ID, doc); err != nil {
			return err
		}
	}
	return idx.index.Batch(batch)
}

// IncomingLinkResult represents a bean that links to a target with a specific link type.
type IncomingLinkResult struct {
	FromID   string
	LinkType string
}

// FindIncomingLinks finds all beans that link to the given target ID.
// Returns a slice of results with the source bean ID and link type.
func (idx *Index) FindIncomingLinks(targetID string) ([]IncomingLinkResult, error) {
	var results []IncomingLinkResult

	// Search each link type field
	linkFields := []struct {
		field    string
		linkType string
	}{
		{"milestone_target", "milestone"},
		{"epic_target", "epic"},
		{"feature_target", "feature"},
		{"block_targets", "blocks"},
		{"related_targets", "related"},
		{"duplicate_targets", "duplicates"},
	}

	for _, lf := range linkFields {
		query := bleve.NewTermQuery(targetID)
		query.SetField(lf.field)

		searchRequest := bleve.NewSearchRequest(query)
		searchRequest.Size = DefaultSearchLimit

		result, err := idx.index.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		for _, hit := range result.Hits {
			results = append(results, IncomingLinkResult{
				FromID:   hit.ID,
				LinkType: lf.linkType,
			})
		}
	}

	return results, nil
}
