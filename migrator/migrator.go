package migrator

import (
	"fmt"
	"log"

	"remnawave-migrate/remnawave"
	"remnawave-migrate/source"
)

type Migrator struct {
	source            source.SourcePanel
	destination       *remnawave.Panel
	PreferredStrategy string
	PreserveInbounds  bool
	PreserveStatus    bool
	PreserveSubHash   bool
}

func New(source source.SourcePanel, destination *remnawave.Panel, preferredStrategy string, preserveInbounds bool, preserveStatus bool, preserveSubHash bool) *Migrator {
	return &Migrator{
		source:            source,
		destination:       destination,
		PreferredStrategy: preferredStrategy,
		PreserveInbounds:  preserveInbounds,
		PreserveStatus:    preserveStatus,
		PreserveSubHash:   preserveSubHash,
	}
}

func (m *Migrator) MigrateUsers(batchSize int, lastUsers int) error {
	if lastUsers > 0 {
		users, err := m.source.GetUsers(0, 1)
		if err != nil {
			return fmt.Errorf("failed to get total users count: %w", err)
		}

		totalUsers := users.Total
		if lastUsers > totalUsers {
			lastUsers = totalUsers
		}

		startOffset := totalUsers - lastUsers
		if startOffset < 0 {
			startOffset = 0
		}

		return m.migrateUsersRange(startOffset, lastUsers, batchSize)
	}

	return m.migrateUsersRange(0, 0, batchSize)
}

func (m *Migrator) migrateUsersRange(startOffset, limit, batchSize int) error {
	offset := startOffset
	processedUsers := 0

	remnawaveInbounds, err := m.destination.GetInbounds()
	if err != nil {
		return fmt.Errorf("failed to get inbounds: %w", err)
	}

	for {
		users, err := m.source.GetUsers(offset, batchSize)
		if err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}

		log.Printf("Processing users %d-%d", offset+1, offset+len(users.Users))

		for i, user := range users.Users {
			if limit > 0 && processedUsers >= limit {
				log.Printf("Reached limit of %d users", limit)
				return nil
			}

			processed := user.Process()
			originalUsername := processed.Username
			createReq := processed.ToCreateUserRequest(m.PreferredStrategy, m.PreserveStatus, m.PreserveSubHash, m.PreserveInbounds, remnawaveInbounds)

			if originalUsername != createReq.Username {
				log.Printf("Username %s was sanitized to %s",
					originalUsername, createReq.Username)
			}

			log.Printf("Processing user %d: %s", offset+i+1, processed.Username)

			err := m.destination.CreateUser(createReq)
			if err != nil {
				if remnawave.IsUserExistsError(err) {
					log.Printf("Skipping user %s: already exists", processed.Username)
					processedUsers++
					continue
				}
				log.Printf("Failed to create user %s: %v", processed.Username, err)
				continue
			}
			log.Printf("Successfully created user: %s", processed.Username)
			processedUsers++
		}

		if len(users.Users) < batchSize || (limit > 0 && processedUsers >= limit) {
			break
		}

		offset += batchSize
	}

	return nil
}
