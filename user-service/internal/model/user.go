// Package model holds the database tables as Go structs (GORM models).
package model

import "gorm.io/gorm"

// User levels, as in the ERD. The JWT carries them as role names
// ("admin" / "member"), see handler.roleOf.
const (
	LevelAdmin  = 1
	LevelMember = 2
)

// User is one row of the "user" table.
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);column:username;not null"`
	// uniqueIndex makes the database reject a second account with the same
	// email, even if two sign-ups arrive at the same moment.
	Email string `gorm:"type:varchar(255);column:email;not null;uniqueIndex"`
	// Password holds a bcrypt hash, never the plain password.
	Password string `gorm:"type:varchar(255);column:password;not null"`
	Level    int16  `gorm:"type:smallint;column:level;not null;default:2"`
}
