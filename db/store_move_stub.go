package db

import (
	"database/sql"
	"fmt"
)

// =============================================================================
// MOVE OPERATIONS — issue #397
// =============================================================================
//
// MoveSystem and MoveApp implement atomic reparenting of the
// Ecosystem -> Domain -> System -> App hierarchy. Both rewrite the moved
// row's denormalized parent FKs together with any cascading child FKs in a
// single transaction so the hierarchy invariants stay consistent:
//
//	system.EcosystemID == newDomain.EcosystemID
//	app.DomainID       == app.System.DomainID (when SystemID is set)
//
// Schema reminders (see migration 026_nullable_hierarchy_fks):
//   - apps.domain_id  is nullable (ON DELETE SET NULL)
//   - apps.system_id  is nullable (ON DELETE SET NULL)
//   - systems.domain_id and systems.ecosystem_id are nullable
//   - apps has NO ecosystem_id column (no app-level ecosystem denormalization)
//
// All writes go through ds.driver.Begin() / Commit() / Rollback() — see
// DeleteSystem in store_system.go for the canonical pattern.
// =============================================================================

// MoveSystem reparents a system to a new domain in a single atomic transaction.
//
// Behavior:
//   - Validates the system exists.
//   - If newDomainID.Valid, validates the target domain exists and uses its
//     EcosystemID to rewrite systems.ecosystem_id (denormalization).
//   - If newDomainID.Valid is false, both systems.domain_id and
//     systems.ecosystem_id are set to NULL (detach to ecosystem level).
//   - Cascades to all child apps: rewrites apps.domain_id to match the new
//     parent domain (or NULL when detached) for every app whose system_id
//     equals systemID.
//
// On any failure the transaction is rolled back and no rows are modified.
func (ds *SQLDataStore) MoveSystem(systemID int, newDomainID sql.NullInt64) error {
	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Validate the system exists up front so we can return a clean error
	// before any writes happen.
	var existsCheck int
	if err := tx.QueryRow(`SELECT id FROM systems WHERE id = ?`, systemID).Scan(&existsCheck); err != nil {
		if err == sql.ErrNoRows {
			return NewErrNotFound("system", systemID)
		}
		return fmt.Errorf("failed to look up system: %w", err)
	}

	// Resolve the new EcosystemID from the target Domain. When detaching
	// (newDomainID.Valid == false) we leave both FKs NULL.
	var newEcosystemID sql.NullInt64
	if newDomainID.Valid {
		if err := tx.QueryRow(
			`SELECT ecosystem_id FROM domains WHERE id = ?`,
			newDomainID.Int64,
		).Scan(&newEcosystemID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("domain id %d not found", newDomainID.Int64)
			}
			return fmt.Errorf("failed to look up target domain: %w", err)
		}
	}

	// Rewrite the system's denormalized parent FKs.
	if _, err := tx.Execute(
		`UPDATE systems SET domain_id = ?, ecosystem_id = ?, updated_at = `+ds.queryBuilder.Now()+` WHERE id = ?`,
		newDomainID, newEcosystemID, systemID,
	); err != nil {
		return fmt.Errorf("failed to update system parent FKs: %w", err)
	}

	// Cascade: rewrite apps.domain_id for every child app of this system so
	// the denormalization invariant (app.DomainID == system.DomainID) holds.
	// apps has no ecosystem_id column, so nothing else to cascade.
	if _, err := tx.Execute(
		`UPDATE apps SET domain_id = ?, updated_at = `+ds.queryBuilder.Now()+` WHERE system_id = ?`,
		newDomainID, systemID,
	); err != nil {
		return fmt.Errorf("failed to cascade domain_id to child apps: %w", err)
	}

	return tx.Commit()
}

// MoveApp reparents an app to a new (domain, system) pair in a single atomic
// transaction.
//
// Both newDomainID and newSystemID are written together because App carries
// both as denormalized FKs and they must remain consistent
// (app.DomainID == app.System.DomainID when SystemID is set).
//
// Special cases:
//   - newSystemID.Valid == false, newDomainID.Valid == true:
//     reparent app to a domain with no system.
//   - newSystemID.Valid == false AND newDomainID.Valid == false:
//     fully detach the app to ecosystem level (used by `dvm app detach`).
//
// On any failure the transaction is rolled back and no rows are modified.
// Returns an error if the app, target domain, or target system does not exist.
func (ds *SQLDataStore) MoveApp(appID int, newDomainID, newSystemID sql.NullInt64) error {
	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Validate the app exists.
	var existsCheck int
	if err := tx.QueryRow(`SELECT id FROM apps WHERE id = ?`, appID).Scan(&existsCheck); err != nil {
		if err == sql.ErrNoRows {
			return NewErrNotFound("app", appID)
		}
		return fmt.Errorf("failed to look up app: %w", err)
	}

	// Validate target Domain exists if provided.
	if newDomainID.Valid {
		if err := tx.QueryRow(
			`SELECT id FROM domains WHERE id = ?`, newDomainID.Int64,
		).Scan(&existsCheck); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("domain id %d not found", newDomainID.Int64)
			}
			return fmt.Errorf("failed to look up target domain: %w", err)
		}
	}

	// Validate target System exists if provided.
	if newSystemID.Valid {
		if err := tx.QueryRow(
			`SELECT id FROM systems WHERE id = ?`, newSystemID.Int64,
		).Scan(&existsCheck); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("system id %d not found", newSystemID.Int64)
			}
			return fmt.Errorf("failed to look up target system: %w", err)
		}
	}

	// Rewrite the app's denormalized parent FKs.
	if _, err := tx.Execute(
		`UPDATE apps SET domain_id = ?, system_id = ?, updated_at = `+ds.queryBuilder.Now()+` WHERE id = ?`,
		newDomainID, newSystemID, appID,
	); err != nil {
		return fmt.Errorf("failed to update app parent FKs: %w", err)
	}

	return tx.Commit()
}
