package models

import "time"

// EmailVerificationToken is one outstanding (or spent) email verification
// link for a user. The raw token exists only in the email that was sent -
// the database stores just its SHA-256, so a database leak cannot be turned
// into working verification links.
//
// A user has at most one unused token at a time: issuing a new one (resend,
// re-registration) deletes the previous unused token first. A verified
// token is kept, stamped with UsedAt, so that re-following an old link can
// be answered "already verified" instead of "invalid".
type EmailVerificationToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index:idx_email_verification_tokens_user_id"`
	User      *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TokenHash string    `gorm:"size:64;not null;uniqueIndex:idx_email_verification_tokens_hash"`
	ExpiresAt time.Time `gorm:"not null;index:idx_email_verification_tokens_expires_at"`
	UsedAt    *time.Time
	CreatedAt time.Time
}
