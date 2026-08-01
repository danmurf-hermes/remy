package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// SaveFact inserts a new fact into the database.
func (s *Store) SaveFact(ctx context.Context, fact *Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("facts").
		Columns("id", "fact", "category", "confidence", "source", "created_at", "updated_at").
		Values(fact.ID, fact.Fact, fact.Category, fact.Confidence, fact.Source, fact.CreatedAt, fact.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving fact: %w", err)
	}
	return nil
}

// GetFact retrieves a single fact by its ID.
func (s *Store) GetFact(ctx context.Context, id string) (*Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "fact", "category", "confidence", "source", "created_at", "updated_at").
		From("facts").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var f Fact
	if err := row.Scan(&f.ID, &f.Fact, &f.Category, &f.Confidence, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, fmt.Errorf("getting fact: %w", err)
	}
	return &f, nil
}

// GetFacts returns a paginated list of all facts, ordered by most recently
// updated.
func (s *Store) GetFacts(ctx context.Context, limit, offset int) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "fact", "category", "confidence", "source", "created_at", "updated_at").
		From("facts").
		OrderBy("updated_at DESC").
		Limit(uint64(limit)).   //nolint:gosec // limit from user input
		Offset(uint64(offset)). //nolint:gosec // offset from user input
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying facts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.ID, &f.Fact, &f.Category, &f.Confidence, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
	}
	return facts, rows.Err()
}

// GetFactsByCategory returns all facts in a given category, ordered by
// confidence descending.
func (s *Store) GetFactsByCategory(ctx context.Context, category string) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "fact", "category", "confidence", "source", "created_at", "updated_at").
		From("facts").
		Where(sq.Eq{"category": category}).
		OrderBy("confidence DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying facts by category: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.ID, &f.Fact, &f.Category, &f.Confidence, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
	}
	return facts, rows.Err()
}

// UpdateFact updates the fact text, category, confidence, and source of an
// existing fact identified by its ID.
func (s *Store) UpdateFact(ctx context.Context, fact *Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Update("facts").
		Set("fact", fact.Fact).
		Set("category", fact.Category).
		Set("confidence", fact.Confidence).
		Set("source", fact.Source).
		Set("updated_at", fact.UpdatedAt).
		Where(sq.Eq{"id": fact.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building update: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating fact: %w", err)
	}
	return nil
}

// DeleteFact removes a fact from the database by its ID.
func (s *Store) DeleteFact(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Delete("facts").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building delete: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting fact: %w", err)
	}
	return nil
}

// SearchFacts performs a vector similarity search over facts using the
// given embedding, returning the top-N most similar facts.
func (s *Store) SearchFacts(ctx context.Context, embedding []byte, limit int) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("f.id", "f.fact", "f.category", "f.confidence", "f.source", "f.created_at", "f.updated_at").
		From("facts f").
		JoinClause(sq.Expr("INNER JOIN (SELECT id, distance FROM fact_vectors WHERE embedding MATCH ? ORDER BY distance LIMIT ?) fv ON fv.id = f.id", embedding, limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building search query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("searching facts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.ID, &f.Fact, &f.Category, &f.Confidence, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
	}
	return facts, rows.Err()
}

// SaveEntity inserts a new entity into the database.
func (s *Store) SaveEntity(ctx context.Context, entity Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("entities").
		Columns("id", "name", "type", "description", "created_at").
		Values(entity.ID, entity.Name, entity.Type, entity.Description, entity.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving entity: %w", err)
	}
	return nil
}

// GetEntity retrieves a single entity by its ID.
func (s *Store) GetEntity(ctx context.Context, id string) (*Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "name", "type", "description", "created_at").
		From("entities").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var e Entity
	if err := row.Scan(&e.ID, &e.Name, &e.Type, &e.Description, &e.CreatedAt); err != nil {
		return nil, fmt.Errorf("getting entity: %w", err)
	}
	return &e, nil
}

// GetEntities returns all entities ordered by name.
func (s *Store) GetEntities(ctx context.Context) ([]Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "name", "type", "description", "created_at").
		From("entities").
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying entities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.Description, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning entity row: %w", err)
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// SaveRelationship inserts a new relationship between two entities into
// the database.
func (s *Store) SaveRelationship(ctx context.Context, rel *Relationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("relationships").
		Columns("id", "source_entity", "target_entity", "relationship", "confidence", "created_at").
		Values(rel.ID, rel.SourceEntity, rel.TargetEntity, rel.Relationship, rel.Confidence, rel.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving relationship: %w", err)
	}
	return nil
}

// GetRelationships returns all relationships in the database.
func (s *Store) GetRelationships(ctx context.Context) ([]Relationship, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "source_entity", "target_entity", "relationship", "confidence", "created_at").
		From("relationships").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying relationships: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var relationships []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.ID, &r.SourceEntity, &r.TargetEntity, &r.Relationship, &r.Confidence, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning relationship row: %w", err)
		}
		relationships = append(relationships, r)
	}
	return relationships, rows.Err()
}
