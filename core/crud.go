package crud

import (
	"context"
	"fmt"
	"mpc-backend/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CRUD struct {
	Connection *pgxpool.Pool
}

func NewCRUD(conn *pgxpool.Pool) *CRUD {
	return &CRUD{conn}
}

func (c *CRUD) CreateOrganization(name string, threshold int, participants []types.Participant) (int, error) {
	tx, err := c.Connection.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	var orgID int
	err = tx.QueryRow(
		context.Background(),
		"INSERT INTO organizations (name, threshold) VALUES ($1, $2) RETURNING id",
		name, threshold,
	).Scan(&orgID)
	if err != nil {
		return 0, err
	}

	for _, p := range participants {
		_, err := tx.Exec(
			context.Background(),
			"INSERT INTO participants (organization_id, address) VALUES ($1, $2)",
			orgID, p.Address,
		)
		if err != nil {
			return 0, err
		}
	}

	return orgID, tx.Commit(context.Background())
}

func (c *CRUD) GetOrganizationsByAddress(address string) ([]types.Organization, error) {
	rows, err := c.Connection.Query(
		context.Background(),
		`SELECT o.id, o.name, o.threshold
		 FROM organizations o
		 JOIN participants p ON o.id = p.organization_id
		 WHERE p.address = $1`, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []types.Organization
	for rows.Next() {
		var org types.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Threshold); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrganizationByName fetches a single organization by its name, including its participants.
func (c *CRUD) GetOrganizationByName(name string) (types.Organization, error) {
	var org types.Organization
	err := c.Connection.QueryRow(context.Background(),
		`
        SELECT id, name, threshold
        FROM organizations
        WHERE name = $1
        `, name,
	).Scan(&org.ID, &org.Name, &org.Threshold)
	if err != nil {
		return org, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Fetch participants for the organization (select only the address)
	rows, err := c.Connection.Query(context.Background(),
		`
        SELECT address
        FROM participants
        WHERE organization_id = $1
        `, org.ID,
	)
	if err != nil {
		return org, fmt.Errorf("failed to fetch participants: %w", err)
	}
	defer rows.Close()

	var participants []types.Participant
	for rows.Next() {
		var p types.Participant
		if err := rows.Scan(&p.Address); err != nil {
			return org, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, p)
	}
	if err := rows.Err(); err != nil {
		return org, fmt.Errorf("error iterating participants: %w", err)
	}
	org.Participants = participants

	return org, nil
}
