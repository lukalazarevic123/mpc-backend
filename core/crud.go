package crud

import (
	"context"
	"mpc-backend/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CRUD struct {
	Connection *pgxpool.Pool
}

func NewCRUD(conn *pgxpool.Pool) *CRUD {
	return &CRUD{conn}
}

func (c *CRUD) CreateOrganization(name string, threshold int, participants []types.Participant) error {
	tx, err := c.Connection.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var orgID int
	err = tx.QueryRow(
		context.Background(),
		"INSERT INTO organizations (name, threshold) VALUES ($1, $2) RETURNING id",
		name, threshold,
	).Scan(&orgID)
	if err != nil {
		return err
	}

	for _, p := range participants {
		_, err := tx.Exec(
			context.Background(),
			"INSERT INTO participants (organization_id, address) VALUES ($1, $2)",
			orgID, p.Address,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
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
