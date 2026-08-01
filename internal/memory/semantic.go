package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var factColumns = []string{"id", "fact", "category", "confidence", "source", "created_at", "updated_at"}

func scanFact(row rowScanner) (Fact, error) {
	var f Fact
	if err := row.Scan(&f.ID, &f.Fact, &f.Category, &f.Confidence, &f.Source, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return Fact{}, err
	}
	return f, nil
}

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
	var f Fact
	scanFn := func(row rowScanner) error {
		var fac Fact
		if err := row.Scan(&fac.ID, &fac.Fact, &fac.Category, &fac.Confidence, &fac.Source, &fac.CreatedAt, &fac.UpdatedAt); err != nil {
			return fmt.Errorf("getting fact: %w", err)
		}
		f = fac
		return nil
	}
	if err := s.scanRow(ctx, buildSelectByID(factColumns, "facts", id), scanFn); err != nil {
		return nil, err
	}
	return &f, nil
}

// GetFacts returns a paginated list of all facts, ordered by most recently
// updated.
func (s *Store) GetFacts(ctx context.Context, limit, offset int) ([]Fact, error) {
	qb := sb.Select(factColumns...).
		From("facts").
		OrderBy("updated_at DESC").
		Limit(uint64(limit)).  //nolint:gosec // limit from user input
		Offset(uint64(offset)) //nolint:gosec // offset from user input

	var facts []Fact
	scanFn := func(row rowScanner) error {
		f, err := scanFact(row)
		if err != nil {
			return fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
		return nil
	}
	if err := s.scanRows(ctx, qb, scanFn, "querying facts"); err != nil {
		return nil, err
	}
	return facts, nil
}

// GetFactsByCategory returns all facts in a given category, ordered by
// confidence descending.
func (s *Store) GetFactsByCategory(ctx context.Context, category string) ([]Fact, error) {
	qb := sb.Select(factColumns...).
		From("facts").
		Where(sq.Eq{"category": category}).
		OrderBy("confidence DESC")

	var facts []Fact
	scanFn := func(row rowScanner) error {
		f, err := scanFact(row)
		if err != nil {
			return fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
		return nil
	}
	if err := s.scanRows(ctx, qb, scanFn, "querying facts by category"); err != nil {
		return nil, err
	}
	return facts, nil
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
	qb := sb.Select("f.id", "f.fact", "f.category", "f.confidence", "f.source", "f.created_at", "f.updated_at").
		From("facts f").
		JoinClause(sq.Expr("INNER JOIN (SELECT id, distance FROM fact_vectors WHERE embedding MATCH ? ORDER BY distance LIMIT ?) fv ON fv.id = f.id", embedding, limit))

	var facts []Fact
	scanFn := func(row rowScanner) error {
		f, err := scanFact(row)
		if err != nil {
			return fmt.Errorf("scanning fact row: %w", err)
		}
		facts = append(facts, f)
		return nil
	}
	if err := s.scanRows(ctx, qb, scanFn, "searching facts"); err != nil {
		return nil, err
	}
	return facts, nil
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
	var e Entity
	scanFn := func(row rowScanner) error {
		var ent Entity
		if err := row.Scan(&ent.ID, &ent.Name, &ent.Type, &ent.Description, &ent.CreatedAt); err != nil {
			return fmt.Errorf("getting entity: %w", err)
		}
		e = ent
		return nil
	}
	if err := s.scanRow(ctx, buildSelectByID([]string{"id", "name", "type", "description", "created_at"}, "entities", id), scanFn); err != nil {
		return nil, err
	}
	return &e, nil
}

// GetEntities returns all entities ordered by name.
func (s *Store) GetEntities(ctx context.Context) ([]Entity, error) {
	qb := sb.Select("id", "name", "type", "description", "created_at").
		From("entities").
		OrderBy("name ASC")

	var entities []Entity
	scanFn := func(row rowScanner) error {
		var e Entity
		if err := row.Scan(&e.ID, &e.Name, &e.Type, &e.Description, &e.CreatedAt); err != nil {
			return fmt.Errorf("scanning entity row: %w", err)
		}
		entities = append(entities, e)
		return nil
	}
	if err := s.scanRows(ctx, qb, scanFn, "querying entities"); err != nil {
		return nil, err
	}
	return entities, nil
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
	qb := sb.Select("id", "source_entity", "target_entity", "relationship", "confidence", "created_at").
		From("relationships")

	var relationships []Relationship
	scanFn := func(row rowScanner) error {
		var r Relationship
		if err := row.Scan(&r.ID, &r.SourceEntity, &r.TargetEntity, &r.Relationship, &r.Confidence, &r.CreatedAt); err != nil {
			return fmt.Errorf("scanning relationship row: %w", err)
		}
		relationships = append(relationships, r)
		return nil
	}
	if err := s.scanRows(ctx, qb, scanFn, "querying relationships"); err != nil {
		return nil, err
	}
	return relationships, nil
}
