package gormdb

import (
	"github.com/go-gormigrate/gormigrate/v2"
)

var (
	// Migrations define all database migrations.
	Migrations = []*gormigrate.Migration{
		// {
		// 	ID: "20240306_create_users_table",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		type User struct {
		// 			ID        string `gorm:"primaryKey;length:36"`
		// 			Slug      string `gorm:"unique;length:255"`
		// 			Username  string `gorm:"unique;length:255"`
		// 			Hashword  string `gorm:"length:255"`
		// 			Email     string `gorm:"unique;length:255"`
		// 			Fullname  string `gorm:"length:255"`
		// 			Active    bool   `gorm:"default:false"`
		// 			Admin     bool   `gorm:"default:false"`
		// 			CreatedAt time.Time
		// 			UpdatedAt time.Time
		// 		}

		// 		return tx.Migrator().CreateTable(&User{})
		// 	},
		// 	Rollback: func(tx *gorm.DB) error {
		// 		return tx.Migrator().DropTable("users")
		// 	},
		// },
		// {
		// 	ID: "20240306_create_teams_Table",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		type Team struct {
		// 			ID        string `gorm:"primaryKey;length:36"`
		// 			Slug      string `gorm:"unique;length:255"`
		// 			Name      string `gorm:"unique;length:255"`
		// 			CreatedAt time.Time
		// 			UpdatedAt time.Time
		// 		}

		// 		return tx.Migrator().CreateTable(&Team{})
		// 	},
		// 	Rollback: func(tx *gorm.DB) error {
		// 		return tx.Migrator().DropTable("teams")
		// 	},
		// },
		// {
		// 	ID: "20240306_create_members_table",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		type Member struct {
		// 			TeamID    string `gorm:"index:idx_id,unique;length:36"`
		// 			UserID    string `gorm:"index:idx_id,unique;length:36"`
		// 			Perm      string `gorm:"length:255"`
		// 			CreatedAt time.Time
		// 			UpdatedAt time.Time
		// 		}

		// 		return tx.Migrator().CreateTable(&Member{})
		// 	},
		// 	Rollback: func(tx *gorm.DB) error {
		// 		return tx.Migrator().DropTable("members")
		// 	},
		// },
		// {
		// 	ID: "20240306_create_members_teams_constraint",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		type Member struct {
		// 			TeamID string `gorm:"index:idx_id,unique;length:36"`
		// 			UserID string `gorm:"index:idx_id,unique;length:36"`
		// 		}

		// 		type Team struct {
		// 			ID    string    `gorm:"primaryKey"`
		// 			Users []*Member `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		// 		}

		// 		return tx.Migrator().CreateConstraint(&Team{}, "Users")
		// 	},
		// 	Rollback: func(tx *gorm.DB) error {
		// 		type Member struct {
		// 			TeamID string `gorm:"index:idx_id,unique;length:36"`
		// 			UserID string `gorm:"index:idx_id,unique;length:36"`
		// 		}

		// 		type Team struct {
		// 			ID    string    `gorm:"primaryKey"`
		// 			Users []*Member `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		// 		}

		// 		return tx.Migrator().DropConstraint(&Team{}, "Users")
		// 	},
		// },
		// {
		// 	ID: "20240306_create_members_users_constraint",
		// 	Migrate: func(tx *gorm.DB) error {
		// 		type Member struct {
		// 			TeamID string `gorm:"index:idx_id,unique;length:36"`
		// 			UserID string `gorm:"index:idx_id,unique;length:36"`
		// 		}

		// 		type User struct {
		// 			ID    string    `gorm:"primaryKey"`
		// 			Teams []*Member `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		// 		}

		// 		return tx.Migrator().CreateConstraint(&User{}, "Teams")
		// 	},
		// 	Rollback: func(tx *gorm.DB) error {
		// 		type Member struct {
		// 			TeamID string `gorm:"index:idx_id,unique;length:36"`
		// 			UserID string `gorm:"index:idx_id,unique;length:36"`
		// 		}

		// 		type User struct {
		// 			ID    string    `gorm:"primaryKey"`
		// 			Teams []*Member `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		// 		}

		// 		return tx.Migrator().DropConstraint(&User{}, "Teams")
		// 	},
		// },
	}
)
